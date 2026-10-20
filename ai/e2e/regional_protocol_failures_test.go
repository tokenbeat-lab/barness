package e2e

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func regionalPrefix(t *testing.T, api ai.API) provider.Reply {
	t.Helper()
	switch api {
	case ai.APIOpenAICompletions:
		return provider.SSE([][]byte{[]byte("data: {\"choices\":[{\"delta\":{\"content\":\"known-content\"}}]}\n\n")})
	case ai.APIOpenAIResponses:
		return minimaxResponsesReply(t, `{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"regional-message","role":"assistant","content":[]}}`, `{"type":"response.output_text.delta","output_index":0,"delta":"known-content"}`)
	default:
		return provider.SSE([][]byte{[]byte("event: message_start\ndata: {\"type\":\"message_start\",\"message\":{\"id\":\"regional\",\"usage\":{\"input_tokens\":3}}}\n\nevent: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"known-content\"}}\n\n")})
	}
}

func TestRegionalProviderFailuresAndCancellation(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range regionalAPIs {
			for _, mode := range []string{"missing-terminal", "malformed", "http", "cancel-before-send", "cancel-before-headers", "cancel-after-content", "queue-limit"} {
				t.Run(string(p)+"/"+string(api)+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "RP14-failure-"+t.Name())
					ready := make(chan struct{})
					reply := regionalPrefix(t, api)
					if mode == "malformed" {
						reply = provider.SSE([][]byte{[]byte("event: message_start\ndata: {\n\n")})
					}
					if mode == "http" {
						reply = provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "X-Request-Id": "regional-failure", "Request-Id": "regional-failure"}, Chunks: [][]byte{[]byte(`{"error":{"type":"authentication_error","code":"invalid_api_key","message":"private-regional-error"}}`)}}
					}
					if mode == "cancel-after-content" {
						reply.End = provider.EndHold
					}
					if mode == "cancel-before-headers" {
						reply = provider.Reply{End: provider.EndSilent, OnReceive: func() { close(ready) }}
					}
					_, bytes := largestAndTotal(reply.Chunks)
					gauges, adm := probe.New(), host.NewAdmission()
					bodies := provider.TrackBodies(provider.LoopbackTransport())
					w, target := regionalWorld(t, p, api, func(c *ai.Config) {
						c.Probe, c.Admission, c.Transport = gauges, adm, bodies
						if mode == "cancel-after-content" {
							c.Transport = &glmReadBarrier{next: bodies, prefix: "/", bytes: bytes, ready: ready}
						}
						if mode == "queue-limit" {
							c.Policy.MaxQueuedEvents = 1
						}
					})
					enqueue(ev, w, reply)
					ctx, cancel := context.WithCancel(ctxFor(t))
					defer cancel()
					if mode == "cancel-before-send" {
						cancel()
					}
					s := w.client.Stream(ctx, textScope("regional-failure-"+mode), target, miniMaxMessagesRequest(), nil)
					defer s.Close()
					if mode == "cancel-before-headers" || mode == "cancel-after-content" {
						waitFor(ev, "controlled cancellation boundary", ready)
						cancel()
					}
					if mode == "queue-limit" {
						_ = resultWithin(ev, "queue pressure settles without a reader", s)
					}
					o := drain(s)
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					miniMaxSharedOneTerminal(ev, o)
					ev.Check("failure settles once", o.err != nil, "%v", o.err)
					want := 1
					if mode == "cancel-before-send" {
						want = 0
					}
					ev.Check("no implicit replay or routing", len(w.provider.Requests()) == want && len(o.result.Metadata.Attempts) == want, "%+v", o.result.Metadata)
					if strings.HasPrefix(mode, "cancel-") {
						ev.Check("cancel classification preserved", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}), "%v", o.err)
					}
					if mode == "cancel-after-content" {
						ev.Check("partial text retained", textOf(o.result.Message) == "known-content", "%+v", o.result.Message.Content)
					}
					if mode == "http" {
						ev.Check("private provider detail withheld", !strings.Contains(o.err.Error(), "private-regional-error") && o.result.Metadata.Attempts[0].ProviderRequestID == "regional-failure", "%v %+v", o.err, o.result.Metadata.Attempts)
					}
					eventually(ev, "calls release", func() bool { return gauges.ActiveCalls() == 0 })
					resources := map[string]int64{"calls": gauges.ActiveCalls(), "events": gauges.QueuedEvents(), "permits": gauges.Permits(), "waiters": gauges.AdmissionWaiters(), "bodies": bodies.Open(), "hostPermits": int64(adm.Held())}
					ev.Record("resources", resources)
					for name, n := range resources {
						ev.Check("released "+name, n == 0, "%d", n)
					}
				})
			}
		}
	}
}

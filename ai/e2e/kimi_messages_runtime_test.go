package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiMessagesHTTPFailures(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, c := range []struct {
			name, kind   string
			status, want int
		}{
			{"quota", "exceeded_current_quota_error", 429, 1}, {"throttle", "rate_limit_reached_error", 429, 2}, {"unknown429", "untrusted-type", 429, 1}, {"auth", "incorrect_api_key_error", 401, 1}, {"non-json-gateway", "", 504, 2},
		} {
			t.Run(string(market)+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K12-http-"+c.name)
				adm := host.NewAdmission()
				w := newKimiMessagesConfigured(t, market, nil, func(c *ai.Config) { c.Admission = adm })
				b := w.host.Binding(tenantA.tenant, kimiMessagesTarget.BindingID)
				b.Retry.MaxRetries = 1
				w.host.PutBinding(b)
				body := fmt.Sprintf(`{"error":{"type":%q,"message":"synthetic-private-upstream-value"}}`, c.kind)
				if c.kind == "" {
					body = "<html>synthetic-private-upstream-value</html>"
				}
				bad := provider.Reply{Status: c.status, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1", "request-id": "unverified-header"}, Chunks: [][]byte{[]byte(body)}}
				enqueue(ev, w, bad, kimiMessagesReply())
				o := drain(w.client.Stream(ctxFor(t), textScope("kimi-http"), kimiMessagesTarget, kimiMessagesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("bounded recognized retry preserves route", len(w.provider.Requests()) == c.want && len(o.result.Metadata.Attempts) == c.want && (o.err == nil) == (c.want == 2), "%v %+v", o.err, o.result.Metadata)
				if o.err != nil {
					ev.Check("upstream free text and unverified request id absent", !strings.Contains(o.err.Error(), "synthetic-private-upstream-value") && o.result.Metadata.Attempts[0].ProviderRequestID == "", "%v %+v", o.err, o.result.Metadata.Attempts)
				}
				ev.Check("all HTTP permits released", adm.Held() == 0, "%d", adm.Held())
			})
		}
	}
}

func TestKimiMessagesResources(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, mode := range []string{"request", "frame", "output", "tool-json", "queue-count", "queue-bytes", "header-timeout", "read-timeout"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K13-"+mode)
				adm := host.NewAdmission()
				w := newKimiMessagesConfigured(t, market, nil, func(c *ai.Config) {
					c.Admission = adm
					switch mode {
					case "request":
						setRequestBytes(c.Policy, 128)
					case "frame":
						setFrameBytes(c.Policy, 128)
					case "output":
						setOutputBytes(c.Policy, 128)
					case "tool-json":
						setToolJSONBytes(c.Policy, 4)
					case "queue-count":
						c.Policy.MaxQueuedEvents = 2
					case "queue-bytes":
						c.Policy.MaxQueuedEventBytes = 4
					case "header-timeout":
						c.Policy.ResponseHeaderTimeout = short
					case "read-timeout":
						c.Policy.ReadIdleTimeout = short
					}
				})
				reply := kimiMessagesReply()
				if mode == "tool-json" {
					reply = provider.SSE([][]byte{miniMaxMessagesFixture(t, "tools.sse", "kimi-k3")})
				}
				if mode == "header-timeout" {
					reply = provider.Reply{End: provider.EndSilent}
				}
				if mode == "read-timeout" {
					reply.End = provider.EndHold
					reply.Chunks = [][]byte{[]byte(`event: message_start
data: {"type":"message_start","message":{"model":"kimi-k3","usage":{"input_tokens":1}}}

`)}
				}
				enqueue(ev, w, reply)
				stream := w.client.Stream(ctxFor(t), textScope("kimi-resource"), kimiMessagesTarget, kimiMessagesRequest(), nil)
				defer stream.Close()
				settled := resultWithin(ev, "Result settles with unread event queue", stream)
				o := drain(stream)
				o.record(ev)
				code, phase := ai.CodeResourceLimit, ai.PhaseStream
				switch mode {
				case "request":
					phase = ai.PhaseRequest
				case "queue-count", "queue-bytes":
					phase = ai.PhaseEventQueue
				case "header-timeout":
					code, phase = ai.CodeDeadlineExceeded, ai.PhaseRequest
				case "read-timeout":
					code = ai.CodeDeadlineExceeded
				}
				ev.Check("finite bound settles once", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}) && settled.err == o.err, "%v", o.err)
				ev.Check("all permits released", adm.Held() == 0, "%d", adm.Held())
			})
		}
	}
}

func TestKimiMessagesCancellation(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K12-cancel-prefix")
			prefix := strings.Split(kimiMessagesScript(`{"input_tokens":100,"cache_read_input_tokens":20,"cache_creation_input_tokens":80}`), "event: message_delta")[0]
			ready := make(chan struct{})
			adm := host.NewAdmission()
			w := newKimiMessagesConfigured(t, market, nil, func(c *ai.Config) {
				c.Admission = adm
				c.Transport = &glmReadBarrier{next: c.Transport, prefix: "/anthropic", bytes: len(prefix), ready: ready}
			})
			reply := provider.SSE([][]byte{[]byte(prefix)})
			reply.End = provider.EndHold
			enqueue(ev, w, reply)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			stream := w.client.Stream(ctx, textScope("kimi-cancel"), kimiMessagesTarget, kimiMessagesRequest(), nil)
			defer stream.Close()
			waitFor(ev, "parser consumed exact lawful prefix", ready)
			cancel()
			settled := resultWithin(ev, "canceled Result settles unread", stream)
			o := drain(stream)
			o.record(ev)
			ev.Check("cancellation preserves last legal counters", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}) && settled.err == o.err && o.result.Message.Usage.Input == 100 && o.result.Metadata.Attempts[0].CostEstimate.Status == ai.CostPartial, "%v %+v", o.err, o.result.Metadata)
			ev.Check("cancellation releases permit without replay", adm.Held() == 0 && len(w.provider.Requests()) == 1, "held=%d", adm.Held())
		})
	}
}

func TestKimiMessagesHookGuards(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, mode := range []string{"auth", "x-api-key", "beta", "version", "panic"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K14-header-"+mode)
				w := newKimiMessagesWorld(t, market, nil)
				hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
					switch mode {
					case "auth":
						h.Set("Authorization", "changed")
					case "x-api-key":
						h.Set("X-Api-Key", "another-key")
					case "beta":
						h.Set("Anthropic-Beta", "unverified")
					case "version":
						h.Set("Anthropic-Version", "unverified")
					case "panic":
						panic(errors.New("synthetic-private-panic"))
					}
					return nil
				}}
				var o outcome
				escaped := false
				func() {
					defer func() { escaped = recover() != nil }()
					o.result, o.err = w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("kimi-hooks"), kimiMessagesTarget, kimiMessagesRequest(), nil)
				}()
				ev.Check("hook cannot escape or reach HTTP", !escaped && o.err != nil && len(w.provider.Requests()) == 0, "escaped=%t error=%v", escaped, o.err)
				if !escaped {
					o.record(ev)
				}
			})
		}
	}
}

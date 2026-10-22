package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"strings"
	"testing"
)

func TestKimiResponsesRetries(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, sc := range []struct {
			name, kind string
			status     int
			retry      bool
		}{
			{"quota", "exceeded_current_quota_error", 429, false}, {"unknown", "private-unknown-type", 429, false}, {"rate", "rate_limit_reached_error", 429, true}, {"overload", "engine_overloaded_error", 429, true}, {"html", "", 504, true}, {"auth", "invalid_authentication_error", 401, false},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K12-retry-")+sc.name)
				w := newKimiResponsesWorld(t, market, nil)
				b := w.host.Binding(tenantA.tenant, "kimi-responses")
				b.Retry.MaxRetries = 1
				w.host.PutBinding(b)
				raw := `{"error":{"type":` + string(mustMarshal(t, sc.kind)) + `,"message":"private provider text"}}`
				if sc.status == 504 {
					raw = "<html>private provider text</html>"
				}
				enqueue(ev, w, provider.Reply{Status: sc.status, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "1", "x-should-retry": "true"}, Chunks: [][]byte{[]byte(raw)}})
				enqueue(ev, w, kimiResponsesReply(t))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				want := 1
				if sc.retry {
					want = 2
				}
				ev.Check("bounded recognized transient retry only", len(o.result.Metadata.Attempts) == want && len(w.provider.Requests()) == want && (o.err == nil) == sc.retry, "%v %+v", o.err, o.result.Metadata)
				if o.err != nil {
					ev.Check("vendor text absent", !strings.Contains(o.err.Error(), "private"), "%v", o.err)
				}
			})
		}
	}
}
func TestKimiResponsesResources(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, mode := range []string{"request", "frame", "tool-json", "output", "queue-count", "queue-bytes", "read-timeout", "header-timeout"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K13-")+mode)
				gauges := probe.New()
				adm := host.NewAdmission()
				var bodies *provider.BodyTracker
				w := newKimiResponsesWorld(t, market, nil, func(c *ai.Config) {
					bodies = provider.TrackBodies(c.Transport)
					c.Transport = bodies
					c.Probe = gauges
					c.Admission = adm
					switch mode {
					case "request":
						setRequestBytes(c.Policy, 128)
					case "frame":
						setFrameBytes(c.Policy, 128)
					case "tool-json":
						setToolJSONBytes(c.Policy, 4)
					case "output":
						setOutputBytes(c.Policy, 128)
					case "queue-count":
						c.Policy.MaxQueuedEvents = 2
					case "queue-bytes":
						c.Policy.MaxQueuedEventBytes = 4
					case "read-timeout":
						c.Policy.ReadIdleTimeout = short
					case "header-timeout":
						c.Policy.ResponseHeaderTimeout = short
					}
				})
				frames := kimiResponsesFrames()
				if mode == "output" {
					frames = []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"` + strings.Repeat("x", 200) + `"}]}]}}`)}
				}
				if mode == "tool-json" {
					frames = kimiResponsesToolFrames()
				}
				reply := kimiResponsesReply(t, frames...)
				if mode == "read-timeout" {
					reply = kimiResponsesReply(t, frames[:3]...)
					reply.End = provider.EndHold
				}
				if mode == "header-timeout" {
					reply = provider.Reply{End: provider.EndSilent}
				}
				enqueue(ev, w, reply)
				ev.Fixture("frames.json", mustMarshal(t, frames))
				s := w.client.Stream(ctxFor(t), textScope("kimi-r-resource-"+string(market)+mode), kimiResponsesTarget(), kimiResponsesToolRequest(), nil)
				defer s.Close()
				settled := resultWithin(ev, "settles without consumer", s)
				o := drain(s)
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				code, phase := ai.CodeResourceLimit, ai.PhaseStream
				if mode == "request" {
					phase = ai.PhaseRequest
				}
				if mode == "queue-count" || mode == "queue-bytes" {
					phase = ai.PhaseEventQueue
				}
				if mode == "read-timeout" || mode == "header-timeout" {
					code = ai.CodeDeadlineExceeded
				}
				if mode == "header-timeout" {
					phase = ai.PhaseRequest
				}
				ev.Check("Host bounds settle once", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}) && settled.err == o.err, "%v", o.err)
				eventually(ev, "calls released", func() bool { return gauges.ActiveCalls() == 0 })
				state := map[string]int64{"calls": gauges.ActiveCalls(), "permits": gauges.Permits(), "events": gauges.QueuedEvents(), "waiters": gauges.AdmissionWaiters(), "bodies": bodies.Open(), "host-permits": int64(adm.Held())}
				ev.Record("resources", state)
				for k, n := range state {
					ev.Check("released "+k, n == 0, "%v", state)
				}
			})
		}
	}
}
func TestKimiResponsesCancellation(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, kimiResponsesID(market, "K12-K13-cancel"))
			gauges := probe.New()
			adm := host.NewAdmission()
			var bodies *provider.BodyTracker
			w := newKimiResponsesWorld(t, market, nil, func(c *ai.Config) {
				bodies = provider.TrackBodies(c.Transport)
				c.Transport = bodies
				c.Probe = gauges
				c.Admission = adm
			})
			b := w.host.Binding(tenantA.tenant, "kimi-responses")
			b.Retry.MaxRetries = 1
			w.host.PutBinding(b)
			reply := kimiResponsesReply(t, kimiResponsesFrames()[:6]...)
			reply.End = provider.EndHold
			enqueue(ev, w, reply)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			s := w.client.Stream(ctx, textScope("kimi-r-cancel-"+string(market)), kimiResponsesTarget(), kimiResponsesRequest(), nil)
			defer s.Close()
			var o outcome
			o.streamed = true
			observed := false
			for s.Next() {
				e := s.Event()
				o.events = append(o.events, e)
				if _, ok := e.(ai.TextDeltaEvent); ok {
					observed = true
					cancel()
					break
				}
			}
			for s.Next() {
				o.events = append(o.events, s.Event())
			}
			o.result, o.err = s.Result()
			o.record(ev)
			ev.Check("cancel preserves actual text prefix", observed && errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled, Phase: ai.PhaseStream}) && textOf(o.result.Message) == "Hello Kimi.", "%v %+v", o.err, o.result.Message)
			ev.Check("stream never replayed", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata)
			enqueue(ev, w, kimiResponsesReply(t))
			res, err := w.client.Complete(ctxFor(t), textScope("kimi-r-recover-"+string(market)), kimiResponsesTarget(), kimiResponsesRequest(), nil)
			ev.Record("recovery", res)
			ev.Check("next explicit call succeeds", err == nil, "%v", err)
			eventually(ev, "resources zero", func() bool {
				return gauges.ActiveCalls() == 0 && gauges.Permits() == 0 && gauges.QueuedEvents() == 0 && bodies.Open() == 0 && adm.Held() == 0
			})
			ev.Record("requests", w.provider.Requests())
		})
	}
}

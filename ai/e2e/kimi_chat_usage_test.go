package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"math"
	"testing"
)

func TestKimiChatUsage(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, sc := range []struct {
			name, usage, status string
			report              ai.UsageReporting
			invalid             bool
		}{
			{"bounded", `{"prompt_tokens":200,"completion_tokens":50,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"completion_tokens_details":{"reasoning_tokens":10}}`, "bounded", ai.UsageComplete, false},
			{"missing-write", `{"prompt_tokens":120,"completion_tokens":50,"prompt_tokens_details":{"cached_tokens":20},"completion_tokens_details":{"reasoning_tokens":10}}`, "bounded", ai.UsagePartial, false},
			{"zero-write", `{"prompt_tokens":120,"completion_tokens":50,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":10}}`, "resolved", ai.UsageComplete, false},
			{"negative", `{"prompt_tokens":-1}`, "unreported", ai.UsageUnreported, true},
			{"fraction", `{"prompt_tokens":1.00000000000000001}`, "unreported", ai.UsageUnreported, true},
			{"overflow", `{"prompt_tokens":9223372036854775808}`, "unreported", ai.UsageUnreported, true},
			{"read-write-conflict", `{"prompt_tokens":10,"prompt_tokens_details":{"cached_tokens":6,"cache_write_tokens":5}}`, "unreported", ai.UsageUnreported, true},
			{"reasoning-conflict", `{"completion_tokens":1,"completion_tokens_details":{"reasoning_tokens":2}}`, "unreported", ai.UsageUnreported, true},
			{"unreported", "", "unreported", ai.UsageUnreported, false},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K10-")+sc.name)
				rec := newRecorder()
				w := newKimiChatWorld(t, market, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
				frames := kimiChatFrames()[:3]
				if sc.usage != "" {
					frames = append(frames, json.RawMessage(`{"choices":[],"usage":`+sc.usage+`}`))
				}
				ev.Fixture("frames.json", mustMarshal(t, frames))
				enqueue(ev, w, chatEvents(t, frames, ""))
				scope := textScope("kimi-usage-" + string(market) + "-" + sc.name)
				o := drain(w.client.Stream(ctxFor(t), scope, kimiChatTarget(), kimiChatRequest(), nil))
				o.record(ev)
				obs := rec.awaitCall(ev, scope.RequestID)
				ev.Record("observations", obs)
				ev.Record("requests", w.provider.Requests())
				if sc.invalid {
					var e *ai.Error
					ev.Check("invalid counts fail as protocol", errors.As(o.err, &e) && e.Code == ai.CodeProtocol && e.Phase == ai.PhaseStream, "%v", o.err)
				} else {
					ev.Check("valid usage succeeds", o.err == nil, "%v", o.err)
				}
				if !ev.Check("one attempt", len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata) {
					return
				}
				a := o.result.Metadata.Attempts[0]
				if len(obs) == 4 {
					ev.Check("observer snapshots agree before delivery", obs[1].Attempt.CostEstimate.Status == ai.CostUnreported && obs[2].Attempt.CostEstimate == a.CostEstimate && obs[3].Call.Attempts[0].CostEstimate == a.CostEstimate && obs[3].Usage == o.result.Message.Usage, "%+v", obs)
				} else {
					ev.Check("four lifecycle observations", false, "%d", len(obs))
				}
				ev.Check("presence remains distinct", a.UsageReporting == sc.report, "%+v", a)
				var v map[string]any
				mustUnmarshal(t, mustMarshal(t, a), &v)
				estimate, _ := v["costEstimate"].(map[string]any)
				ev.Check("cost certainty independent", estimate != nil && estimate["status"] == sc.status, "%v", estimate)
				if sc.name == "bounded" && market == "moonshot" {
					min, _ := estimate["minimum"].(map[string]any)
					max, _ := estimate["maximum"].(map[string]any)
					ev.Check("independent worked USD example", min != nil && max != nil && math.Abs(min["total"].(float64)-0.001296) < 1e-12 && math.Abs(max["total"].(float64)-0.001536) < 1e-12 && math.Abs(a.Usage.Cost.Total-0.001296) < 1e-12, "%v %v", min, max)
				}
			})
		}
	}
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		t.Run(string(market)+"/http-unreported", func(t *testing.T) {
			ev := run.Case(t, kimiChatID(market, "K11-http-unreported"))
			w := newKimiChatWorld(t, market, nil)
			enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{"error":{"type":"invalid_authentication","message":"synthetic"}}`)}})
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), nil))
			o.record(ev)
			if !ev.Check("failure attempt", o.err != nil && len(o.result.Metadata.Attempts) == 1, "%v", o.err) {
				return
			}
			var v map[string]any
			mustUnmarshal(t, mustMarshal(t, o.result.Metadata.Attempts[0]), &v)
			estimate, _ := v["costEstimate"].(map[string]any)
			ev.Check("HTTP failure does not imply free usage", estimate != nil && estimate["status"] == "unreported", "%v", estimate)
		})
	}
}

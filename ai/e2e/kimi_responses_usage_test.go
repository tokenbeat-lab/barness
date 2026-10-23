package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"math"
	"strings"
	"testing"
)

func TestKimiResponsesUsage(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, sc := range []struct {
			name, usage, actual string
			status              ai.CostEstimateStatus
			report              ai.UsageReporting
			invalid             bool
		}{
			{"bounded", kimiResponsesUsage, "", ai.CostBounded, ai.UsageComplete, false},
			{"actual-short", kimiResponsesUsage, `{"mode":"implicit","ttl":"5m"}`, ai.CostResolved, ai.UsageComplete, false},
			{"actual-long", kimiResponsesUsage, `{"mode":"implicit","ttl":"1h"}`, ai.CostResolved, ai.UsageComplete, false},
			{"actual-unknown", kimiResponsesUsage, `{"mode":"implicit","ttl":"other"}`, ai.CostBounded, ai.UsageComplete, false},
			{"missing-write", `{"input_tokens":120,"output_tokens":50,"total_tokens":170,"input_tokens_details":{"cached_tokens":20},"output_tokens_details":{"reasoning_tokens":10}}`, "", ai.CostBounded, ai.UsagePartial, false},
			{"zero", `{"input_tokens":0,"output_tokens":0,"total_tokens":0,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}`, "", ai.CostResolved, ai.UsageComplete, false},
			{"missing-reasoning", `{"input_tokens":120,"output_tokens":50,"total_tokens":170,"input_tokens_details":{"cached_tokens":20,"cache_write_tokens":0}}`, "", ai.CostResolved, ai.UsagePartial, false},
			{"unreported", "", "", ai.CostUnreported, ai.UsageUnreported, false},
			{"fraction", `{"input_tokens":1.00000000000000001}`, "", ai.CostUnreported, ai.UsageUnreported, true},
			{"negative", `{"input_tokens":-1}`, "", ai.CostUnreported, ai.UsageUnreported, true},
			{"overflow", `{"input_tokens":9223372036854775808}`, "", ai.CostUnreported, ai.UsageUnreported, true},
			{"cache-conflict", `{"input_tokens":10,"input_tokens_details":{"cached_tokens":6,"cache_write_tokens":5}}`, "", ai.CostUnreported, ai.UsageUnreported, true},
			{"reasoning-conflict", `{"output_tokens":1,"output_tokens_details":{"reasoning_tokens":2}}`, "", ai.CostUnreported, ai.UsageUnreported, true},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiResponsesID(market, "K10-K11-")+sc.name)
				rec := newRecorder()
				w := newKimiResponsesWorld(t, market, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
				response := map[string]any{"status": "completed", "output": []any{json.RawMessage(kimiResponsesMessage)}}
				if sc.usage != "" {
					response["usage"] = json.RawMessage(sc.usage)
				}
				if sc.actual != "" {
					response["prompt_cache_options"] = json.RawMessage(sc.actual)
				}
				frame := mustMarshal(t, map[string]any{"type": "response.completed", "response": response})
				enqueue(ev, w, kimiResponsesReply(t, frame))
				scope := textScope("kimi-r-usage-" + string(market) + sc.name)
				o := drain(w.client.Stream(ctxFor(t), scope, kimiResponsesTarget(), kimiResponsesRequest(), ai.ResponsesOptions{CacheRetention: ai.CacheRetentionLong}))
				o.record(ev)
				obs := rec.awaitCall(ev, scope.RequestID)
				ev.Record("observations", obs)
				ev.Record("requests", w.provider.Requests())
				if sc.invalid {
					ev.Check("invalid counts fail", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
				} else {
					ev.Check("valid usage accepted", o.err == nil, "%v", o.err)
				}
				if !ev.Check("one attempt", len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata) {
					return
				}
				a := o.result.Metadata.Attempts[0]
				ev.Check("presence and certainty independent", a.UsageReporting == sc.report && a.CostEstimate.Status == sc.status, "%+v", a)
				if len(obs) == 4 {
					ev.Check("observer snapshots agree", obs[1].Attempt.CostEstimate.Status == ai.CostUnreported && obs[2].Attempt.CostEstimate == a.CostEstimate && obs[3].Call.Attempts[0].CostEstimate == a.CostEstimate, "%+v", obs)
				}
				if market == ai.ProviderMoonshot && (sc.name == "bounded" || sc.name == "actual-short" || sc.name == "actual-long") {
					lo, lok := a.CostEstimate.Minimum.Get()
					hi, hok := a.CostEstimate.Maximum.Get()
					min, max := 0.001296, 0.001536
					if sc.name == "actual-short" {
						max = min
					}
					if sc.name == "actual-long" {
						min = max
					}
					ev.Check("worked USD bounds use observed tier", lok && hok && math.Abs(lo.Total-min) < 1e-12 && math.Abs(hi.Total-max) < 1e-12 && a.Usage.Cost == lo, "%+v", a)
				}
			})
		}
	}
}
func TestKimiResponsesFailedUsage(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(market), func(t *testing.T) {
			ev := run.Case(t, kimiResponsesID(market, "K10-failed-usage"))
			w := newKimiResponsesWorld(t, market, nil)
			frame := json.RawMessage(`{"type":"response.failed","response":{"status":"failed","error":{"code":"private-code","message":"private provider text"},"output":[],"usage":` + kimiResponsesUsage + `}}`)
			enqueue(ev, w, kimiResponsesReply(t, frame))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), nil))
			o.record(ev)
			if !ev.Check("failed response classified safely", errors.Is(o.err, &ai.Error{Code: ai.CodeUpstreamError}) && !strings.Contains(o.err.Error(), "private"), "%v", o.err) {
				return
			}
			a := o.result.Metadata.Attempts[0]
			ev.Check("ADR0010 failed usage not taken", a.UsageReporting == ai.UsageUnreported && a.Usage.TotalTokens == 0 && a.CostEstimate.Status == ai.CostUnreported, "%+v", a)
		})
	}
}

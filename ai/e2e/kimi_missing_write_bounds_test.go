package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"math"
	"testing"
)

func TestKimiMissingWriteCostBounds(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses} {
			for _, sc := range []struct {
				name, actual               string
				input, output, read, write any
				rate, minimum, maximum     float64
				status                     ai.CostEstimateStatus
				invalid                    bool
			}{
				{"unknown-ttl", "", 120, 50, 20, nil, 3, .001056, .001356, ai.CostBounded, false},
				{"default-read", "", 120, 50, nil, nil, 3, .00111, .00147, ai.CostBounded, false},
				{"actual-short", "5m", 120, 50, 20, nil, 3, .001056, .001056, ai.CostResolved, false},
				{"actual-long", "1h", 120, 50, 20, nil, 3, .001056, .001356, ai.CostBounded, false},
				{"equal-known-write", "", 120, 50, 20, 80, 6, .001296, .001296, ai.CostResolved, false},
				{"reversed-write", "", 120, 50, 20, nil, 10, .001056, .001756, ai.CostBounded, false},
				{"cheaper-write", "", 120, 50, 20, nil, 1, .000856, .001356, ai.CostBounded, false},
				{"zero-remaining", "", 20, 50, 20, nil, 3, .000756, .000756, ai.CostResolved, false},
				{"missing-input", "", nil, 50, 20, nil, 3, 0, 0, ai.CostPartial, false},
				{"missing-output", "", 120, nil, 20, nil, 3, 0, 0, ai.CostPartial, false},
				{"total-conflict", "", 120, 50, 20, nil, 3, 0, 0, ai.CostUnreported, true},
				{"overflow", "", 120, 50, 20, nil, math.MaxFloat64, 0, 0, ai.CostUnreported, true},
			} {
				if sc.actual != "" && api != ai.APIOpenAIResponses {
					continue
				}
				t.Run(string(market)+"/"+string(api)+"/"+sc.name, func(t *testing.T) {
					ev := run.Case(t, "KIMI-"+string(market)+"-"+string(api)+"-K11-write-bounds-"+sc.name)
					// Explicit trusted Host rates give the same independently worked USD
					// amounts across both markets; they are not official price assertions.
					mutate := func(c *ai.Catalog) {
						for i := range c.Models {
							c.Models[i].Cost = ai.ModelCost{CostRates: ai.CostRates{Input: 3, Output: 15, CacheRead: .3, CacheWrite: sc.rate}}
						}
					}
					rec := newRecorder()
					cfg := func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 }
					w := newKimiChatWorld(t, market, mutate, cfg)
					target := kimiChatTarget()
					input, output, details, outdetails := "prompt_tokens", "completion_tokens", "prompt_tokens_details", "completion_tokens_details"
					if api == ai.APIOpenAIResponses {
						w = newKimiResponsesWorld(t, market, mutate, cfg)
						target = kimiResponsesTarget()
						input, output, details, outdetails = "input_tokens", "output_tokens", "input_tokens_details", "output_tokens_details"
					}
					usage := map[string]any{input: sc.input, output: sc.output, details: map[string]any{}, outdetails: map[string]any{"reasoning_tokens": 10}}
					if sc.input != nil && sc.output != nil {
						usage["total_tokens"] = sc.input.(int) + sc.output.(int)
					}
					if sc.name == "total-conflict" {
						usage["total_tokens"] = 171
					}
					if sc.read != nil {
						usage[details].(map[string]any)["cached_tokens"] = sc.read
					}
					if sc.write != nil {
						usage[details].(map[string]any)["cache_write_tokens"] = sc.write
					}
					raw := mustMarshal(t, usage)
					ev.Record("reported-usage", json.RawMessage(raw))
					if api == ai.APIOpenAICompletions {
						enqueue(ev, w, chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}],"usage":` + string(raw) + `}`)}, ""))
					} else {
						response := map[string]any{"status": "completed", "output": []any{json.RawMessage(kimiResponsesMessage)}, "usage": json.RawMessage(raw)}
						if sc.actual != "" {
							response["prompt_cache_options"] = map[string]any{"mode": "implicit", "ttl": sc.actual}
						}
						enqueue(ev, w, kimiResponsesReply(t, mustMarshal(t, map[string]any{"type": "response.completed", "response": response})))
					}
					scope := textScope("write-bounds-" + string(market) + "-" + string(api) + "-" + sc.name)
					var opts ai.Options = ai.ChatOptions{CacheRetention: ai.CacheRetentionLong}
					if api == ai.APIOpenAIResponses {
						opts = ai.ResponsesOptions{CacheRetention: ai.CacheRetentionLong}
					}
					result, err := w.client.Complete(ctxFor(t), scope, target, kimiChatRequest(), opts)
					ev.Record("result", result)
					obs := rec.awaitCall(ev, scope.RequestID)
					ev.Record("observations", obs)
					if sc.invalid {
						ev.Check("nonfinite bound cannot succeed", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", err)
						return
					}
					if !ev.Check("finite or explicitly partial report succeeds", err == nil && len(result.Metadata.Attempts) == 1, "%v", err) {
						return
					}
					a := result.Metadata.Attempts[0]
					u := result.Message.Usage
					lo, lok := a.CostEstimate.Minimum.Get()
					hi, hok := a.CostEstimate.Maximum.Get()
					ev.Check("cost certainty independent from missing write fact", a.CostEstimate.Status == sc.status && (sc.write != nil || a.UsageReporting == ai.UsagePartial), "%+v", a)
					if sc.status == ai.CostPartial {
						ev.Check("principal facts cannot be invented", !lok && !hok, "%+v", a)
					} else {
						ev.Check("worked feasible allocation bounds", lok && hok && math.Abs(lo.Total-sc.minimum) < 1e-12 && math.Abs(hi.Total-sc.maximum) < 1e-12 && u.Cost == lo, "%+v", a)
					}
					if sc.write == nil {
						ev.Check("cost scenarios do not become observed writes", u.CacheWrite == 0 && u.CacheWrite1h.IsZero(), "%+v", u)
					}
					ev.Check("observer result and attempt snapshots agree", len(obs) == 4 && obs[2].Attempt.Usage == u && obs[2].Attempt.CostEstimate == a.CostEstimate && obs[3].Usage == u && a.Usage == u, "%+v", obs)
				})
			}
		}
	}
}

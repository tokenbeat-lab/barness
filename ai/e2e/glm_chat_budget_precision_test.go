package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Final callback JSON numbers must be checked without float64 rounding.
// Integer-valued decimal and exponent spellings remain valid JSON-schema integers.
func TestGLMChatOutputBudgetPrecision(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"model", "profile", "call", "hook"} {
			for _, sc := range []struct {
				raw   string
				valid bool
			}{
				{"1.00000000000000001", false}, {"131072.00000000001", false},
				{"1.0", true}, {"1e0", true}, {"131072e0", true}, {"1e1000000000", false},
			} {
				t.Run(model+"/"+source+"/"+sc.raw, func(t *testing.T) {
					ev := run.Case(t, "GLM03-output-budget-precision-"+model+"-"+source+"-"+sc.raw)
					params := map[string]json.RawMessage{"max_tokens": json.RawMessage(sc.raw)}
					w := newGLMChatWorld(t, func(cat *ai.Catalog) {
						for i := range cat.Models {
							if cat.Models[i].ID == model {
								if source == "model" {
									cat.Models[i].SamplingParams = params
								}
								if source == "profile" {
									cat.Models[i].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingHigh: params}
								}
							}
						}
					})
					reply, raw := glmChatReply(t, model)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, reply)
					opts, hooks := ai.ChatOptions{ReasoningEffort: ai.ThinkingHigh}, ai.Hooks{}
					if source == "call" {
						opts.SamplingParams = params
					}
					if source == "hook" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							p.Body["max_tokens"] = json.Number(sc.raw)
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-budget-precision"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmChatRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if sc.valid {
						ev.Check("integer numeric wire accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
					} else {
						checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
						ev.Check("fractional or unbounded budget rejected before send", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
					}
				})
			}
		}
	}
}

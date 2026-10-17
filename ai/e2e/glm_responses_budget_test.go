package e2e

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// China Responses schema declares integer max_output_tokens 1..131072. Named null/0
// retain their existing omission semantics; final wire null/0 is invalid.
func TestGLMResponsesOutputBudget(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"named", "model", "profile", "call", "hook"} {
			for _, value := range []string{"-1", "0", "1", "131072", "131073", "1.5", "1.00000000000000001", "1e999999999", "1e0", "null", `"1024"`, "true"} {
				if source == "named" && (value == "1.5" || value == "1.00000000000000001" || value == "1e999999999" || value == "1e0" || value == `"1024"` || value == "true") {
					continue
				}
				name := model + "/" + source + "/" + value
				t.Run(name, func(t *testing.T) {
					ev := run.Case(t, "GLM04-output-budget-"+name)
					params := map[string]json.RawMessage{"max_output_tokens": json.RawMessage(value)}
					w := newGLMResponsesWorld(t, func(cat *ai.Catalog) {
						for i := range cat.Models {
							if cat.Models[i].ID != model {
								continue
							}
							if source == "model" {
								cat.Models[i].SamplingParams = params
							}
							if source == "profile" {
								level := ai.ThinkingLow
								if model == "glm-5.2" {
									level = ai.ThinkingLevel("off")
								}
								cat.Models[i].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{level: params}
							}
						}
					})
					scenarios, raw := glmResponsesFixtures(t)
					reply := glmResponsesReply(t, scenarios[0].Events)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, reply)
					opts, hooks := ai.ResponsesOptions{}, ai.Hooks{}
					if source == "named" {
						if value == "null" {
							opts.MaxTokens = ai.Null[int]()
						} else {
							n, _ := strconv.Atoi(value)
							opts.MaxTokens = ai.Value(n)
						}
					}
					if source == "call" {
						opts.SamplingParams = params
					}
					if source == "hook" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							p.Body["max_output_tokens"] = json.RawMessage(value)
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-budget-"+name), ai.Target{BindingID: "glm-responses", ModelID: model}, glmResponsesRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Record("candidate-catalog", w.client.Catalog())
					valid := value == "1" || value == "131072" || value == "1e0" || (source == "named" && (value == "0" || value == "null"))
					if valid {
						ev.Check("valid budget succeeds", o.err == nil && len(w.provider.Requests()) == 1, "err=%v", o.err)
						if len(w.provider.Requests()) == 1 {
							fields := bodyFields(t, w.provider.Requests()[0].Body)
							if source == "named" && (value == "0" || value == "null") {
								_, present := fields["max_output_tokens"]
								ev.Check("named null/zero omitted", !present, "%s", w.provider.Requests()[0].Body)
							} else {
								ev.Check("valid budget unchanged", jsonEqual(fields["max_output_tokens"], []byte(value)), "%s", w.provider.Requests()[0].Body)
							}
						}
						return
					}
					phase := ai.PhaseRequest
					if source == "named" && value == "-1" {
						phase = ai.PhaseCapability
					}
					checkRejected(ev, o, ai.CodeInvalidRequest, phase)
					ev.Check("zero provider requests and attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
				})
			}
		}
	}
}

func TestGLMResponsesProviderOutputLimit(t *testing.T) {
	for _, c := range []struct {
		name                  string
		modelLimit, requested int
		valid                 bool
	}{
		{"model-narrows-valid", 32, 32, true},
		{"model-narrows-invalid", 32, 33, false},
		{"provider-cap-valid", 200000, 131072, true},
		{"catalog-cannot-raise-provider-cap", 200000, 131073, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM04-provider-output-limit-"+c.name)
			w := newGLMResponsesWorld(t, func(cat *ai.Catalog) { cat.Models[0].MaxTokens = c.modelLimit })
			scenarios, raw := glmResponsesFixtures(t)
			reply := glmResponsesReply(t, scenarios[0].Events)
			ev.Fixture("text.json", raw)
			enqueue(ev, w, reply)
			o := drain(w.client.Stream(ctxFor(t), textScope("glm-provider-limit-"+c.name), ai.Target{BindingID: "glm-responses", ModelID: "glm-5.3"}, glmResponsesRequest(), ai.ResponsesOptions{MaxTokens: ai.Value(c.requested)}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Record("candidate-catalog", w.client.Catalog())
			if c.valid {
				ev.Check("within provider and model bounds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
				return
			}
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("catalog cannot exceed provider or model capacity", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

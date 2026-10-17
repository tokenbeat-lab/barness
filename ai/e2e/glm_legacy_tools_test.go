package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMChatLegacyToolFields(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"hook", "sampling"} {
			for _, field := range []string{"functions", "function_call", "both"} {
				t.Run(model+"/"+source+"/"+field, func(t *testing.T) {
					ev := run.Case(t, "P0-GLM13-legacy-tools-"+model+"-"+source+"-"+field)
					w := newGLMChatWorld(t, nil)
					reply, raw := glmChatReply(t, model)
					ev.Fixture("own-wire.json", raw)
					enqueue(ev, w, reply)
					values := map[string]any{}
					if field == "functions" || field == "both" {
						values["functions"] = []any{map[string]any{"name": "lookup", "parameters": map[string]any{"type": "object"}}}
					}
					if field == "function_call" || field == "both" {
						values["function_call"] = map[string]any{"name": "lookup"}
					}
					opts := ai.ChatOptions{}
					hooks := ai.Hooks{}
					if source == "hook" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							for k, v := range values {
								p.Body[k] = v
							}
							return ai.KeepPayload(), nil
						}
					} else {
						opts.SamplingParams = map[string]json.RawMessage{}
						for k, v := range values {
							b, _ := json.Marshal(v)
							opts.SamplingParams[k] = b
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm13-legacy-tool"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmChatRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					phase := ai.PhaseRequest
					if source == "sampling" && field != "function_call" {
						phase = ai.PhaseCapability
					}
					ev.Check("legacy schema cannot force an unintegrated choice", errors.Is(o.err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: phase}) && len(o.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "err=%v attempts=%d requests=%d", o.err, len(o.result.Metadata.Attempts), len(w.provider.Requests()))
				})
			}
		}
	}
}

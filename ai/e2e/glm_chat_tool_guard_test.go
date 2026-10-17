package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMChatToolChoiceNamed(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, choice := range []ai.ChatToolChoice{
			{Mode: "required"}, {Mode: "none"}, {Function: "lookup"},
			{AllowedTools: &ai.ResponsesAllowedTools{Mode: "auto", Functions: []string{"lookup"}}},
			{AllowedTools: &ai.ResponsesAllowedTools{Mode: "required", Functions: []string{"lookup"}}},
		} {
			raw := string(mustMarshal(t, choice))
			t.Run(model+"/"+raw, func(t *testing.T) {
				ev := run.Case(t, "GLM03-named-tool-choice-"+t.Name())
				w := newGLMChatWorld(t, nil)
				reply, fixture := glmChatReply(t, model)
				ev.Fixture("text.json", fixture)
				enqueue(ev, w, reply)
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-choice-named"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmToolRequest(model), ai.ChatOptions{ToolChoice: ai.Value(choice)}))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("no send or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
		for _, entry := range []string{"stream-simple", "complete-simple"} {
			t.Run(model+"/none/"+entry, func(t *testing.T) {
				ev := run.Case(t, "GLM03-simple-tool-none-"+model+entry)
				w := newGLMChatWorld(t, nil)
				reply, raw := glmChatReply(t, model)
				ev.Fixture("text.json", raw)
				enqueue(ev, w, reply)
				opts := ai.SimpleOptions{ToolChoice: ai.Value(ai.ToolChoiceNone)}
				var o outcome
				if entry == "stream-simple" {
					o = drain(w.client.StreamSimple(ctxFor(t), textScope("glm-choice-simple"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmToolRequest(model), opts))
				} else {
					o.result, o.err = w.client.CompleteSimple(ctxFor(t), textScope("glm-choice-simple"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmToolRequest(model), opts)
				}
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if entry == "stream-simple" {
					checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				} else {
					var failure *ai.Error
					ev.Check("simple none explicitly rejected", errors.As(o.err, &failure) && failure.Code == ai.CodeInvalidRequest && failure.Phase == ai.PhaseCapability, "%v", o.err)
				}
				ev.Check("no send or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

// Model defaults, canonical profiles, call sampling and final callback must
// all obey the actual final body. A hook may repair intermediate sampling.
func TestGLMChatFinalToolControls(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorize-first"} {
			for _, sc := range []struct{ id, field, raw string }{
				{"required", "tool_choice", `"required"`}, {"none", "tool_choice", `"none"`},
				{"function", "tool_choice", `{"type":"function","function":{"name":"lookup"}}`},
				{"allowed", "tool_choice", `{"type":"allowed_tools","allowed_tools":{"mode":"auto","tools":[{"type":"function","function":{"name":"lookup"}}]}}`},
				{"null", "tool_choice", `null`}, {"bad-tool-stream", "tool_stream", `false`},
				{"preserved-thinking", "thinking", `{"type":"enabled","clear_thinking":false}`},
				{"bad-clear-thinking", "thinking", `{"type":"enabled","clear_thinking":"true"}`},
			} {
				t.Run(model+"/"+source+"/"+sc.id, func(t *testing.T) {
					ev := run.Case(t, "GLM03-final-tool-controls-"+t.Name())
					params := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
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
					if source == "call" || source == "repair" {
						opts.SamplingParams = params
					}
					if source == "hook" || source == "repair" || source == "authorize-first" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var value any
							mustUnmarshal(t, []byte(sc.raw), &value)
							p.Body[sc.field] = value
							if source == "repair" {
								p.Body["tool_choice"] = "auto"
								p.Body["tool_stream"] = true
								p.Body["thinking"] = map[string]any{"type": "enabled", "clear_thinking": true}
							}
							if source == "authorize-first" {
								p.Body["model"] = "outside-allowlist"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-final-tools"), ai.Target{BindingID: "glm-chat", ModelID: model}, glmToolRequest(model), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if source == "repair" {
						ev.Check("legal repaired final controls accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						return
					}
					code := ai.CodeInvalidRequest
					if source == "authorize-first" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					ev.Check("zero network and Attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
				})
			}
		}
	}
}

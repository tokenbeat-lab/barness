package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/glm"
)

// Every source reaches the actual serialized-body boundary, including the
// no-hook path. A repairing hook may make intermediate sampling valid.
func TestGLMChatFinalThinking(t *testing.T) {
	for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorize-first"} {
		for _, field := range []string{"disabled", "unsupported-effort", "null-thinking", "malformed-effort"} {
			t.Run(source+"/"+field, func(t *testing.T) {
				ev := run.Case(t, "GLM02-final-thinking-"+source+"-"+field)
				param, value := "thinking", json.RawMessage(`{"type":"disabled"}`)
				switch field {
				case "unsupported-effort":
					param, value = "reasoning_effort", json.RawMessage(`"medium"`)
				case "null-thinking":
					value = json.RawMessage(`null`)
				case "malformed-effort":
					param, value = "reasoning_effort", json.RawMessage(`false`)
				}
				params := map[string]json.RawMessage{param: value}
				w := newGLMChatWorld(t, func(cat *ai.Catalog) {
					if source == "model" {
						cat.Models[0].SamplingParams = params
					}
					if source == "profile" {
						cat.Models[0].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingLow: params}
					}
				})
				reply, raw := glmChatReply(t, "glm-5.3")
				ev.Fixture("text.json", raw)
				enqueue(ev, w, reply)
				opts := ai.ChatOptions{}
				if source == "call" || source == "repair" {
					opts.SamplingParams = params
				}
				hooks := ai.Hooks{}
				if source == "hook" || source == "repair" || source == "authorize-first" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						var v any
						mustUnmarshal(t, value, &v)
						p.Body[param] = v
						if source == "repair" {
							p.Body["thinking"] = map[string]any{"type": "enabled"}
							p.Body["reasoning_effort"] = "high"
						}
						if source == "authorize-first" {
							p.Body["model"] = "outside-tenant-allowlist"
						}
						return ai.KeepPayload(), nil
					}
				}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-final-"+source+"-"+field), ai.Target{BindingID: "glm-chat", ModelID: "glm-5.3"}, glmChatRequest(), opts))
				o.record(ev)
				ev.Record("candidate-catalog", w.client.Catalog())
				ev.Record("requests", w.provider.Requests())
				if source == "repair" {
					ev.Check("valid repaired final request succeeds", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
					return
				}
				code := ai.CodeInvalidRequest
				if source == "authorize-first" {
					code = ai.CodeTenantDenied
				}
				checkRejected(ev, o, code, ai.PhaseRequest)
				ev.Check("no send or attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMChatSampling(t *testing.T) {
	for _, repairEffort := range []bool{false, true} {
		t.Run(map[bool]string{false: "canonical", true: "wire-rewrite"}[repairEffort], func(t *testing.T) {
			ev := run.Case(t, "GLM02-sampling-"+t.Name())
			w := newGLMChatWorld(t, func(cat *ai.Catalog) {
				m := &cat.Models[0]
				m.SamplingParams = map[string]json.RawMessage{"temperature": json.RawMessage(`0.1`), "top_p": json.RawMessage(`0.9`)}
				m.SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{
					ai.ThinkingHigh: {"temperature": json.RawMessage(`0.2`), "top_p": json.RawMessage(`0.8`)},
					ai.ThinkingLow:  {"temperature": json.RawMessage(`0.7`), "top_p": json.RawMessage(`0.1`)},
				}
			})
			reply, raw := glmChatReply(t, "glm-5.3")
			ev.Fixture("text.json", raw)
			enqueue(ev, w, reply)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				ev.Check("hook sees general-profile-call layers", p.Body["temperature"] == json.Number("0.3") && p.Body["top_p"] == json.Number("0.8"), "%v", p.Body)
				p.Body["temperature"] = json.Number("0.4")
				if repairEffort {
					p.Body["reasoning_effort"] = "low"
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(hooks).CompleteSimple(ctxFor(t), textScope("glm-sampling"), ai.Target{BindingID: "glm-chat", ModelID: "glm-5.3"}, glmChatRequest(), ai.SimpleOptions{Reasoning: ai.ThinkingMedium, SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage(`0.3`)}})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("one successful send", err == nil && len(w.provider.Requests()) == 1, "%v", err) {
				return
			}
			fields := bodyFields(t, w.provider.Requests()[0].Body)
			effort := `"high"`
			if repairEffort {
				effort = `"low"`
			}
			ev.Check("final hook wins without reselecting profile", string(fields["temperature"]) == "0.4" && string(fields["top_p"]) == "0.8" && string(fields["reasoning_effort"]) == effort, "%s", w.provider.Requests()[0].Body)
		})
	}
}

func TestGLMChatMandatoryThinkingMetadata(t *testing.T) {
	ev := run.Case(t, "GLM02-mandatory-thinking-metadata")
	cat := glm.ChatCatalog()
	hash := catalogHash(t, cat)
	// This public JSON field must survive NewClient snapshots and enter Hash.
	cat.Models[0].Compat.RequiresThinking = false
	ev.Check("semantic model constraint enters hash", hash != catalogHash(t, cat), "hash=%s", hash)
	w := newGLMChatWorld(t, nil)
	snapshot := w.client.Catalog()
	var fields map[string]any
	mustUnmarshal(t, mustMarshal(t, snapshot.Models[0].Compat), &fields)
	ev.Record("snapshot", snapshot)
	ev.Check("snapshot retains mandatory thinking", fields["requiresThinking"] == true, "%v", fields)
	// Altering a returned model cannot turn off the client's hard constraint.
	snapshot.Models[0].Compat.RequiresThinking = false
	ev.Check("returned metadata owns its value", catalogHash(t, w.client.Catalog()) == hash, "hash differs")
}

package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesToolChoices(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, choice := range []ai.ResponsesToolChoice{{Mode: "auto"}, {Mode: "none"}, {Mode: "required"}, {Function: "lookup_probe_code"}, {AllowedTools: &ai.ResponsesAllowedTools{Mode: "auto", Functions: []string{"lookup_probe_code"}}}} {
			t.Run(model+"/"+string(mustMarshal(t, choice)), func(t *testing.T) {
				ev := run.Case(t, "GLM05-named-choice-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				scenarios, raw := glmResponsesFixtures(t)
				ev.Fixture("text.json", raw)
				enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
				scope := textScope("glm-responses-choice")
				o := drain(w.client.Stream(ctxFor(t), scope, glmResponsesTarget(model), glmResponsesToolRequest(), ai.ResponsesOptions{ToolChoice: ai.Value(choice)}))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
				if choice.Mode == "auto" || choice.Mode == "none" {
					ev.Check("documented choice accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
					return
				}
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("no credential read, request or Attempt", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMResponsesNamedCapabilities(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, sc := range []struct {
			name string
			opts ai.ResponsesOptions
		}{
			{"summary", ai.ResponsesOptions{ReasoningSummary: "auto"}},
			{"tier", ai.ResponsesOptions{ServiceTier: ai.Value("auto")}},
			{"null-tier", ai.ResponsesOptions{ServiceTier: ai.Null[string]()}},
			{"null-choice", ai.ResponsesOptions{ToolChoice: ai.Null[ai.ResponsesToolChoice]()}},
		} {
			t.Run(model+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, "GLM05-named-capabilities-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				scope := textScope("glm-named-capabilities")
				called := false
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
					called = true
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, glmResponsesTarget(model), glmResponsesRequest(), sc.opts))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("explicit unsupported option fails before credentials or hooks", w.host.ReadsFor(scope.RequestID).Credentials == 0 && !called && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMResponsesFinalControls(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"model", "profile", "call", "hook", "repair", "authorize-first"} {
			for _, sc := range []struct{ id, field, raw string }{
				{"required", "tool_choice", `"required"`}, {"named", "tool_choice", `{"type":"function","name":"lookup_probe_code"}`},
				{"allowed", "tool_choice", `{"type":"allowed_tools","mode":"auto","tools":[{"type":"function","name":"lookup_probe_code"}]}`}, {"null-choice", "tool_choice", `null`},
				{"summary", "reasoning", `{"effort":"high","summary":"auto"}`}, {"include", "include", `["reasoning.encrypted_content"]`},
				{"tier", "service_tier", `"auto"`}, {"retention", "prompt_cache_retention", `"24h"`}, {"ttl", "prompt_cache_options", `{"ttl":"30m"}`},
				{"parallel", "parallel_tool_calls", `true`}, {"managed", "context_management", `[{"type":"compaction","compact_threshold":100}]`},
			} {
				t.Run(model+"/"+source+"/"+sc.id, func(t *testing.T) {
					ev := run.Case(t, "GLM05-final-"+t.Name())
					params := map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
					w := newGLMResponsesWorld(t, func(cat *ai.Catalog) {
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
					scenarios, raw := glmResponsesFixtures(t)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
					opts, hooks := ai.ResponsesOptions{ReasoningEffort: ai.ThinkingHigh}, ai.Hooks{}
					if source == "call" || source == "repair" {
						opts.SamplingParams = params
					}
					if source == "hook" || source == "repair" || source == "authorize-first" {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var value any
							_ = json.Unmarshal([]byte(sc.raw), &value)
							p.Body[sc.field] = value
							if source == "repair" {
								delete(p.Body, sc.field)
							}
							if source == "authorize-first" {
								p.Body["model"] = "unauthorized"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-responses-final"), glmResponsesTarget(model), glmResponsesToolRequest(), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if source == "repair" {
						ev.Check("legal final hook repair accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
						return
					}
					code := ai.CodeInvalidRequest
					if source == "authorize-first" {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, ai.PhaseRequest)
					ev.Check("zero requests and Attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
				})
			}
		}
	}
}

func TestGLMResponsesFinalToolsAndModality(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, sc := range []struct{ id, field, raw string }{
			{"strict-false", "tools", `[{"type":"function","name":"lookup_probe_code","parameters":{"type":"object"},"strict":false}]`},
			{"strict-true", "tools", `[{"type":"function","name":"lookup_probe_code","parameters":{"type":"object"},"strict":true}]`},
			{"function-namespace", "tools", `[{"type":"function","name":"lookup_probe_code","parameters":{"type":"object"},"namespace":"hidden"}]`},
			{"missing-function-name", "tools", `[{"type":"function","parameters":{"type":"object"}}]`},
			{"malformed-function-schema", "tools", `[{"type":"function","name":"lookup_probe_code","parameters":"not-a-schema"}]`},
			{"hosted", "tools", `[{"type":"web_search"}]`}, {"namespace", "tools", `[{"type":"namespace","name":"n","tools":[]}]`}, {"custom", "tools", `[{"type":"custom","name":"x"}]`},
			{"image-input", "input", `[{"role":"user","content":[{"type":"input_image","image_url":"data:image/png;base64,AA=="}]}]`},
			{"inline-file-input", "input", `[{"type":"message","role":"user","content":[{"type":"input_file","file_data":"synthetic-file"}]}]`},
			{"audio-input", "input", `[{"type":"message","role":"user","content":[{"type":"input_audio","input_audio":{"data":"AA==","format":"wav"}}]}]`},
		} {
			t.Run(model+"/"+sc.id, func(t *testing.T) {
				ev := run.Case(t, "GLM05-final-tools-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				w.updateBindingOf(tenantA, "glm-responses", func(b *ai.Binding) { b.AllowedHostedTools = []string{"web_search", "namespace", "custom"} })
				scenarios, raw := glmResponsesFixtures(t)
				ev.Fixture("text.json", raw)
				enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					var value any
					_ = json.Unmarshal([]byte(sc.raw), &value)
					p.Body[sc.field] = value
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-responses-final-tools"), glmResponsesTarget(model), glmResponsesToolRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("zero requests and Attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMResponsesCatalogCannotEnableStrict(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM05-catalog-strict-"+model)
			w := newGLMResponsesWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					cat.Models[i].Compat.SupportsStrictMode = true
					cat.Models[i].Compat.SupportsLongCacheRetention = ai.Value(true)
					cat.Models[i].Compat.SupportsExplicitPromptCacheMode = true
				}
			})
			scenarios, raw := glmResponsesFixtures(t)
			ev.Fixture("text.json", raw)
			enqueue(ev, w, glmResponsesReply(t, scenarios[0].Events))
			o := drain(w.client.Stream(ctxFor(t), textScope("glm-responses-catalog-strict"), glmResponsesTarget(model), glmResponsesToolRequest(), ai.ResponsesOptions{SessionID: "synthetic-session", CacheRetention: ai.CacheRetentionLong}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("narrow provider capabilities remain usable", o.err == nil, "%v", o.err)
			if len(w.provider.Requests()) == 1 {
				body := bodyFields(t, w.provider.Requests()[0].Body)
				var tools []map[string]any
				mustUnmarshal(t, body["tools"], &tools)
				_, strict := tools[0]["strict"]
				ev.Check("catalog cannot enable strict", !strict, "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}

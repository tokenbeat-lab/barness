package e2e

import (
	"context"
	"encoding/json"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMSharedChatFinalCapabilities(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, source := range []string{"sampling", "hook", "authorize-first"} {
			for _, sc := range []struct{ id, field, raw string }{
				{"strict-false", "tools", `[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"},"strict":false}}]`},
				{"strict-true", "tools", `[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"},"strict":true}}]`},
				{"hosted-field", "web_search_options", `{"search_context_size":"low"}`},
				{"hosted", "tools", `[{"type":"web_search","web_search":{"enable":true}}]`},
				{"malformed-tools", "tools", `{"type":"function"}`},
				{"tool-namespace", "tools", `[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"}},"namespace":"hidden"}]`},
				{"native-object", "messages", `[{"role":"user","content":{"type":"image_url","image_url":{"url":"data:image/png;base64,AA=="}}}]`},
				{"bad-schema", "tools", `[{"type":"function","function":{"name":"lookup","parameters":"schema"}}]`},
				{"image", "messages", `[{"role":"user","content":[{"type":"image_url","image_url":{"url":"data:image/png;base64,AA=="}}]}]`},
				{"audio", "messages", `[{"role":"user","content":[{"type":"input_audio","input_audio":{"data":"AA==","format":"wav"}}]}]`},
				{"file", "messages", `[{"role":"user","content":[{"type":"file","file":{"file_data":"synthetic"}}]}]`},
				{"developer", "messages", `[{"role":"developer","content":"Synthetic instruction"}]`},
				{"stream", "stream", `false`}, {"store", "store", `false`},
				{"cache-key", "prompt_cache_key", `"synthetic-session"`},
				{"cache-ttl", "prompt_cache_retention", `"24h"`}, {"tier", "service_tier", `"auto"`},
				{"parallel", "parallel_tool_calls", `true`}, {"format", "response_format", `{"type":"json_schema","json_schema":{"name":"strict"}}`},
			} {
				t.Run(model+"/"+source+"/"+sc.id, func(t *testing.T) {
					ev := run.Case(t, "GLM09-chat-final-"+t.Name())
					w := newGLMChatWorld(t, nil)
					w.updateBindingOf(tenantA, "glm-chat", func(b *ai.Binding) { b.AllowedHostedTools = []string{"web_search", "web_search_options"} })
					reply, raw := glmChatReply(t, model)
					ev.Fixture("text.json", raw)
					enqueue(ev, w, reply)
					opts := ai.ChatOptions{}
					hooks := ai.Hooks{}
					if source == "sampling" {
						opts.SamplingParams = map[string]json.RawMessage{sc.field: json.RawMessage(sc.raw)}
					} else {
						hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
							var v any
							mustUnmarshal(t, []byte(sc.raw), &v)
							p.Body[sc.field] = v
							if source == "authorize-first" {
								p.Body["model"] = "unauthorized-model"
							}
							return ai.KeepPayload(), nil
						}
					}
					o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm09-boundary-"+model+"-"+source+"-"+sc.id), ai.Target{BindingID: "glm-chat", ModelID: model}, glmToolRequest(model), opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					code := ai.CodeInvalidRequest
					if source == "authorize-first" {
						code = ai.CodeTenantDenied
					}
					phase := ai.PhaseRequest
					if source == "sampling" && slices.Contains([]string{"tools", "messages", "stream", "store", "prompt_cache_key", "web_search_options"}, sc.field) {
						phase = ai.PhaseCapability
					}
					if source == "hook" && (sc.id == "malformed-tools" || sc.id == "cache-key") {
						code = ai.CodeTenantDenied
					}
					checkRejected(ev, o, code, phase)
					ev.Check("preflight never creates a request or Attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
				})
			}
		}
	}
}

func TestGLMSharedChatCatalogAndData(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM09-chat-catalog-data-"+model)
			w := newGLMChatWorld(t, func(c *ai.Catalog) {
				for i := range c.Models {
					c.Models[i].Compat.SupportsStrictMode = true
				}
			})
			reply, raw := glmChatReply(t, model)
			ev.Fixture("text.json", raw)
			enqueue(ev, w, reply)
			req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}, Tools: []ai.Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object","properties":{"strict":{"type":"boolean"},"image_url":{"type":"string"},"service_tier":{"type":"string"}},"required":[]}`)}}}
			o := drain(w.client.Stream(ctxFor(t), textScope("glm09-catalog-data-"+model), ai.Target{BindingID: "glm-chat", ModelID: model}, req, ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"response_format": json.RawMessage(`{"type":"json_object"}`), "do_sample": json.RawMessage(`false`)}}))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("documented controls and schema property names accepted", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err)
			if len(w.provider.Requests()) == 1 {
				var tools []struct{ Function map[string]any }
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["tools"], &tools)
				_, strict := tools[0].Function["strict"]
				ev.Check("catalog cannot enable endpoint strict", !strict, "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}

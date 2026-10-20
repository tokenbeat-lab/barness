package e2e

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestGLMResponsesOriginIDCollision(t *testing.T) {
	idsCases := []glmOpaqueIDCase{
		{"punctuation", []string{"same/a", "same?a"}, []string{"fc/a", "fc?a"}},
		{"long-prefix", []string{strings.Repeat("long", 20) + "a", strings.Repeat("long", 20) + "b"}, []string{"fc_a", "fc_b"}},
	}
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, ids := range idsCases {
			for _, destination := range []string{"glm-chat", "glm-messages", "openai-responses", "deepseek-responses", "openai-chat", "deepseek-chat", "anthropic-messages"} {
				t.Run(model+"/"+ids.name+"/"+destination, func(t *testing.T) {
					ev := run.Case(t, "GLM05-origin-id-collision-"+t.Name())
					// The formal catalog now owns the six GLM identities.
					w := newGLMSharedWorld(t, func(cfg *ai.Config) { cfg.Catalog = nil })
					w.host.PutBinding(primaryBinding(tenantA, w.provider.URL()))
					w.host.PutCredential(primaryCredential(tenantA, "v1"))
					w.host.PutBinding(chatBinding(tenantA, w.provider.URL()))
					w.host.PutBinding(anthropicBinding(tenantA, w.provider.URL()))
					w.host.PutCredential(anthropicCredential(tenantA))
					w.host.PutBinding(deepseekBinding(tenantA, w.provider.URL()))
					w.host.PutBinding(deepseekChatBinding(tenantA, w.provider.URL()))
					w.host.PutCredential(deepseekCredential(tenantA))
					source := glmSharedRoute{"responses", ai.APIOpenAIResponses}
					events := glmOpaqueIDEvents(t, ids)
					w.provider.EnqueueAt(source.prefix(glmA), glmResponsesReply(t, events))
					ev.Fixture("native-ids.json", mustMarshal(t, map[string]any{"callIDs": ids.calls, "itemIDs": ids.items, "events": events}))
					req := glmResponsesToolRequest()
					first, err := w.client.Complete(ctxFor(t), textScope("glm-origin-ids-first"), source.target(model), req, nil)
					ev.Record("first-result", first)
					if !ev.Check("GLM source exposes two native calls", err == nil && len(first.Message.Content) == 2, "%v %+v", err, first.Message) {
						return
					}
					req.Messages = append(req.Messages, first.Message)
					for i, b := range first.Message.Content {
						out := "first-result"
						if i == 1 {
							out = "second-result"
						}
						req.Messages = append(req.Messages, ai.ToolResultText(b.(ai.ToolCall), out, false))
					}
					route := glmSharedRoute{"responses", ai.APIOpenAIResponses}
					target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
					var opts ai.Options
					reply := glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)})
					switch destination {
					case "glm-chat", "glm-messages":
						name := strings.TrimPrefix(destination, "glm-")
						api := ai.APIOpenAICompletions
						if name == "messages" {
							api = ai.APIAnthropicMessages
						}
						route = glmSharedRoute{name, api}
						target = route.target(model)
						opts = route.options()
						_, _, reply, _ = route.script(t, model)
						w.provider.EnqueueAt(route.prefix(glmA), reply)
					case "deepseek-responses":
						target = ai.Target{BindingID: "deepseek", ModelID: "deepseek-flash"}
						w.provider.Enqueue(reply)
					case "openai-chat", "deepseek-chat":
						route = glmSharedRoute{"chat", ai.APIOpenAICompletions}
						target = ai.Target{BindingID: "chat", ModelID: "gpt-4.1-mini"}
						if destination == "deepseek-chat" {
							target = ai.Target{BindingID: "deepseek-chat", ModelID: "deepseek-flash"}
						}
						reply = chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"Done."},"finish_reason":"stop"}]}`), json.RawMessage(`"[DONE]"`)}, provider.FramingLF)
						w.provider.Enqueue(reply)
					case "anthropic-messages":
						route = glmSharedRoute{"messages", ai.APIAnthropicMessages}
						for _, m := range ai.BuiltinCatalog().Models {
							if m.Provider == ai.ProviderAnthropic && m.API == ai.APIAnthropicMessages {
								target = ai.Target{BindingID: "claude", ModelID: m.ID}
								break
							}
						}
						opts = ai.AnthropicOptions{MaxTokens: ai.Value(1024)}
						reply = sseEvents(t, []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"m","type":"message","role":"assistant","content":[],"usage":{"input_tokens":1,"output_tokens":0}}}`), json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":0}}`), json.RawMessage(`{"type":"message_stop"}`)}, provider.FramingLF)
						w.provider.Enqueue(reply)
					default:
						w.provider.Enqueue(reply)
					}
					ev.Record("target-response", reply)
					o := drain(w.client.Stream(ctxFor(t), textScope("glm-origin-ids-second"), target, req, opts))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("cross-model native policy retained", o.err == nil && o.result.Metadata.NativeStateDowngrades.CrossModel == 1, "%v %+v", o.err, o.result.Metadata)
					qs := w.provider.Requests()
					if len(qs) != 2 {
						return
					}
					native, calls, results := protocolWirePairs(t, route.name, qs[1].Body)
					ev.Check("two different GLM calls never collapse on target wire", len(calls) == 2 && calls[0] != calls[1], "calls=%v", calls)
					ev.Check("all downgraded tool results retain pairing", native == 0 && slices.Equal(calls, results) && len(results) == 2, "native=%d calls=%v results=%v", native, calls, results)
					ev.Check("both distinct Host results remain visible", strings.Contains(string(qs[1].Body), "first-result") && strings.Contains(string(qs[1].Body), "second-result"), "%s", qs[1].Body)
					w.released(ev)
				})
			}
		}
	}
}

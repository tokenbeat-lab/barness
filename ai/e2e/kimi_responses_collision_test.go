package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"slices"
	"strings"
	"testing"
)

func TestKimiResponsesCrossTargetIDs(t *testing.T) {
	for _, source := range []string{"chat", "responses", "messages"} {
		for _, destination := range []string{"openai-chat", "openai-responses", "deepseek-responses"} {
			for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
				t.Run(string(market)+"/"+source+"/"+destination, func(t *testing.T) {
					ev := run.Case(t, kimiResponsesID(market, "K06-collision-")+t.Name())
					// The activated builtin already contains Kimi and the destination models.
					w := newWorld(t, tenantA)
					api := ai.APIOpenAICompletions
					prefix := "/v1"
					ids := []string{"same/a", "same?a"}
					var frames []json.RawMessage
					switch source {
					case "chat":
						frames = []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"same/a","function":{"name":"lookup","arguments":"{\"q\":\"a\"}"}},{"index":1,"id":"same?a","function":{"name":"lookup","arguments":"{\"q\":\"b\"}"}}]},"finish_reason":"tool_calls"}]}`)}
					case "responses":
						api = ai.APIOpenAIResponses
						ids = []string{strings.Repeat("opaque", 20) + "a", strings.Repeat("opaque", 20) + "b"}
						items := []any{}
						for i, id := range ids {
							items = append(items, map[string]any{"type": "function_call", "id": "item|" + id, "call_id": id, "name": "lookup", "arguments": `{"q":"x"}`})
							frames = append(frames, mustMarshal(t, map[string]any{"type": "response.output_item.done", "output_index": i, "item": items[i]}))
						}
						frames = append(frames, mustMarshal(t, map[string]any{"type": "response.completed", "response": map[string]any{"status": "completed", "output": items}}))
					case "messages":
						api = ai.APIAnthropicMessages
						prefix = "/anthropic"
						frames = []json.RawMessage{json.RawMessage(`{"type":"message_start","message":{"id":"m","model":"kimi-k3","role":"assistant","content":[],"usage":{"input_tokens":1,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}`)}
						for i, id := range ids {
							frames = append(frames, mustMarshal(t, map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "tool_use", "id": id, "name": "lookup", "input": map[string]any{"q": "x"}}}), mustMarshal(t, map[string]any{"type": "content_block_stop", "index": i}))
						}
						frames = append(frames, json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"tool_use"},"usage":{"output_tokens":1}}`), json.RawMessage(`{"type":"message_stop"}`))
					}
					b := ai.Binding{TenantID: tenantA.tenant, BindingID: "kimi-source", Version: "v1", Enabled: true, ProviderID: market, API: api, Endpoint: w.provider.URL() + prefix, AuthKind: ai.AuthAPIKey, AccountScopeID: "acct-source", CredentialRef: "cred-source", AllowedModels: []string{"kimi-k3"}}
					w.host.PutBinding(b)
					w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(kimiResponsesKey.secret)})
					var reply provider.Reply
					if source == "chat" {
						reply = chatEvents(t, frames, "")
					} else {
						reply = sseEvents(t, frames, provider.FramingLF)
					}
					enqueue(ev, w, reply)
					req := kimiResponsesToolRequest()
					first, err := w.client.Complete(ctxFor(t), textScope("collision-source-"+string(market)+source+destination), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, req, nil)
					ev.Record("first", first)
					if !ev.Check("source returns actual distinct calls", err == nil && len(first.Message.Content) == 2, "%v %+v", err, first.Message) {
						return
					}
					req.Messages = append(req.Messages, first.Message)
					for i, c := range first.Message.Content {
						req.Messages = append(req.Messages, ai.ToolResultText(c.(ai.ToolCall), []string{"first-result", "second-result"}[i], false))
					}
					target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
					route := "responses"
					response := sseEvents(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}, provider.FramingLF)
					if destination == "openai-chat" {
						route = "chat"
						target.BindingID = "chat"
						response = chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"Done."},"finish_reason":"stop"}]}`)}, "")
					}
					if destination == "deepseek-responses" {
						target = ai.Target{BindingID: "deepseek", ModelID: "deepseek-flash"}
					}
					enqueue(ev, w, response)
					o := drain(w.client.Stream(ctxFor(t), textScope("collision-target-"+string(market)+source+destination), target, req, nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					if !ev.Check("target accepts downgraded history", o.err == nil && len(w.provider.Requests()) == 2, "%v", o.err) {
						return
					}
					_, calls, results := protocolWirePairs(t, route, w.provider.Requests()[1].Body)
					ev.Check("distinct calls remain distinct and paired", len(calls) == 2 && calls[0] != "" && calls[0] != calls[1] && slices.Equal(calls, results), "%v %v", calls, results)
				})
			}
		}
	}
}

func TestKimiResponsesSequentialProvenanceIDs(t *testing.T) {
	for _, mode := range []string{"distinct-source-same-id", "source-then-native", "same-source-same-id", "different-trusted-accounts"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-INTL-RESPONSES-K06-sequential-"+mode)
			w := newWorld(t, tenantA)
			call := ai.ToolCall{ID: "same/a", Name: "lookup", Arguments: ai.UncheckedArguments(`{"q":"x"}`)}
			first := ai.AssistantMessage{Provider: ai.ProviderMoonshot, API: ai.APIOpenAICompletions, Model: "kimi-k3", StopReason: ai.StopReasonToolUse, Content: []ai.AssistantContent{call}}
			second := first
			second.Provider = ai.ProviderMoonshotCN
			if mode == "same-source-same-id" || mode == "different-trusted-accounts" {
				second.Provider = first.Provider
			}
			if mode == "different-trusted-accounts" {
				for i, m := range []*ai.AssistantMessage{&first, &second} {
					state, err := ai.TrustNativeState(textScope("trusted-account"), ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: []string{"source-account-a", "source-account-b"}[i], ProviderID: m.Provider, API: m.API, ModelID: m.Model})
					if err != nil {
						t.Fatal(err)
					}
					m.NativeState = state
				}
			}
			if mode == "source-then-native" {
				second.Provider = ai.ProviderOpenAI
				second.Model = "gpt-4.1-mini"
			}
			req := kimiResponsesToolRequest()
			req.Messages = append(req.Messages, first, ai.ToolResultText(call, "first-result", false), second, ai.ToolResultText(call, "second-result", false))
			enqueue(ev, w, chatEvents(t, []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"Done."},"finish_reason":"stop"}]}`)}, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope("sequential-"+mode), ai.Target{BindingID: "chat", ModelID: "gpt-4.1-mini"}, req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("history sends", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
				return
			}
			_, calls, results := protocolWirePairs(t, "chat", w.provider.Requests()[0].Body)
			ev.Check("each source turn keeps its own unambiguous result", len(calls) == 2 && len(results) == 2 && calls[0] != calls[1] && slices.Equal(calls, results), "%v %v", calls, results)
		})
	}
}

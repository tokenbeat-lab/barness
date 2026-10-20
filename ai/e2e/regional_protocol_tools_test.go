package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func regionalToolReply(t *testing.T, p ai.ProviderID, api ai.API, model string) provider.Reply {
	t.Helper()
	switch api {
	case ai.APIOpenAICompletions:
		reasoning := `"reasoning_content":"Regional reasoning."`
		if p == ai.ProviderMiniMaxCN {
			reasoning += `,"reasoning_details":[{"type":"reasoning.text","text":"Regional reasoning."}]`
		}
		return chatEvents(t, []json.RawMessage{
			json.RawMessage(`{"choices":[{"delta":{` + reasoning + `,"tool_calls":[{"index":0,"id":"regional-call","function":{"name":"inspect","arguments":"{\"label\":"}}]}}]}`),
			json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"alpha\"}"}}]},"finish_reason":"tool_calls"}]}`),
			json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":7,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"reasoning_tokens":3}}}`),
		}, "")
	case ai.APIOpenAIResponses:
		return minimaxResponsesReply(t,
			`{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"regional-reasoning","summary":[],"content":[{"type":"reasoning_text","text":"Regional reasoning."}]}}`,
			`{"type":"response.output_item.added","output_index":1,"item":{"type":"function_call","id":"opaque-item","call_id":"regional-call","name":"inspect","arguments":""}}`,
			`{"type":"response.function_call_arguments.delta","output_index":1,"delta":"{\"label\":"}`,
			`{"type":"response.function_call_arguments.delta","output_index":1,"delta":"\"alpha\"}"}`,
			`{"type":"response.output_item.done","output_index":1,"item":{"type":"function_call","id":"opaque-item","call_id":"regional-call","name":"inspect","arguments":"{\"label\":\"alpha\"}"}}`,
			strings.ReplaceAll(minimaxResponsesCompleted, "MiniMax-M3", model))
	default:
		return provider.SSE([][]byte{miniMaxMessagesFixture(t, "tools.sse", model)})
	}
}

func TestRegionalProviderToolsAndNativeHistory(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range regionalAPIs {
			for _, model := range regionalModels(p) {
				for _, origin := range []string{"same", "sibling", "untrusted", "rotated", "unsigned-same", "unsigned-sibling", "unsigned-untrusted"} {
					if strings.HasPrefix(origin, "unsigned-") && (api != ai.APIAnthropicMessages || p != ai.ProviderZAI) {
						continue
					}
					t.Run(string(p)+"/"+string(api)+"/"+model+"/"+origin, func(t *testing.T) {
						mode := strings.TrimPrefix(origin, "unsigned-")
						ev := run.Case(t, "RP07-native-"+t.Name())
						w, target := regionalWorld(t, p, api)
						target.ModelID = model
						enqueue(ev, w, regionalToolReply(t, p, api, target.ModelID), regionalTextReply(t, p, api, target.ModelID))
						req := miniMaxToolRequest()
						first := drain(w.client.Stream(ctxFor(t), textScope("regional-tools-first"), target, req, nil))
						first.record(ev)
						if !ev.Check("fragmented tool stream completes", first.err == nil && first.result.Message.StopReason == ai.StopReasonToolUse, "%v %+v", first.err, first.result.Message) {
							return
						}
						m := first.result.Message
						if strings.HasPrefix(origin, "unsigned-") {
							b := m.Content[0].(ai.Thinking)
							b.Signature = ""
							m.Content[0] = b
						}
						if mode == "rotated" {
							b := w.host.Binding(tenantA.tenant, target.BindingID)
							w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, Version: "v2", BindingVersion: b.Version, AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantA2.secret)})
						}
						if mode == "untrusted" {
							m, _, _ = restore(t, persist(t, m))
						}
						if mode == "sibling" {
							source := ai.ProviderZhipu
							if p == ai.ProviderMiniMaxCN {
								source = ai.ProviderMiniMax
							}
							m.Provider = source
							env := ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "shared-alias", ProviderID: source, API: api, ModelID: target.ModelID}
							var err error
							m.NativeState, err = ai.TrustNativeState(textScope("regional-trust"), env)
							ev.Check("Host trusts original service identity", err == nil, "%v", err)
						}
						req.Messages = append(req.Messages, m)
						calls := 0
						for i, block := range m.Content {
							if call, ok := block.(ai.ToolCall); ok {
								valid, err := ai.ValidateToolCall(m, i, req.Tools)
								var args struct{ Label string }
								if !ev.Check("Host can execute complete tool only", err == nil && valid.Decode(&args) == nil && (args.Label == "alpha" || args.Label == "beta"), "%v %+v", err, args) {
									return
								}
								req.Messages = append(req.Messages, ai.ToolResultText(call, "regional-result", false))
								calls++
							}
						}
						second := drain(w.client.Stream(ctxFor(t), textScope("regional-tools-second"), target, req, nil))
						ev.Record("continuation", second.result)
						ev.Record("requests", w.provider.Requests())
						if !ev.Check("explicit Host continuation succeeds", second.err == nil && calls > 0 && len(w.provider.Requests()) == 2, "%v", second.err) {
							return
						}
						want := ai.NativeStateDowngrades{}
						if mode == "untrusted" {
							want.NoEnvelope = 1
						}
						if mode == "sibling" {
							want.CrossModel = 1
						}
						ev.Check("exact service provenance enforced", second.result.Metadata.NativeStateDowngrades == want, "%+v", second.result.Metadata.NativeStateDowngrades)
						body := string(w.provider.Requests()[1].Body)
						if mode == "same" || mode == "rotated" {
							switch api {
							case ai.APIOpenAICompletions:
								ev.Check("native reasoning replayed", strings.Contains(body, `"reasoning_content":"Regional reasoning."`), "%s", body)
							case ai.APIOpenAIResponses:
								ev.Check("native reasoning projected", strings.Contains(body, `"type":"reasoning"`) && !strings.Contains(body, `opaque-item`), "%s", body)
							default:
								ev.Check("native signature retained", (strings.HasPrefix(origin, "unsigned-") && strings.Contains(body, `"type":"thinking"`) && !strings.Contains(body, `"signature"`) || !strings.HasPrefix(origin, "unsigned-") && strings.Contains(body, `"signature":"sig-part1-part2"`)), "%s", body)
							}
						} else {
							ev.Check("opaque native state removed", !strings.Contains(body, `"type":"reasoning"`) && !strings.Contains(body, `"signature"`) && !strings.Contains(body, `"reasoning_details"`), "%s", body)
						}
						ev.Check("tool results retain pairing", strings.Count(body, "regional-result") == calls, "%s", body)
						ev.Check("regional result stays attributed", second.result.Message.Provider == p && second.result.Metadata.ProviderID == p, "%+v", second.result.Metadata)
					})
				}
			}
		}
	}
}

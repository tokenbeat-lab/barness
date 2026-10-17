package e2e

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type glmOpaqueIDCase struct {
	name         string
	calls, items []string
}

var glmOpaqueIDCases = []glmOpaqueIDCase{
	{"ordinary", []string{"call_alpha", "call_beta"}, []string{"fc_alpha", "fc_beta"}},
	{"delimiter", []string{"same|a", "same|b"}, []string{"fc|a", "fc|b"}},
	{"ambiguous-composite", []string{"same|a", "same"}, []string{"item", "a|item"}},
	{"percent", []string{"same%7Ca", "same|a"}, []string{"fc%25|a", "fc%7C|b"}},
	{"unicode", []string{"同名|甲", "同名|乙"}, []string{"条目|甲%", "条目|乙%7C"}},
}

func glmOpaqueIDEvents(t *testing.T, ids glmOpaqueIDCase) []json.RawMessage {
	t.Helper()
	var events []json.RawMessage
	for i, id := range ids.calls {
		events = append(events, mustMarshal(t, map[string]any{"type": "response.output_item.done", "output_index": i, "item": map[string]any{"type": "function_call", "id": ids.items[i], "call_id": id, "name": "lookup_probe_code", "arguments": fmt.Sprintf(`{"label":"%d"}`, i)}}))
	}
	return append(events, json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`))
}

func TestGLMResponsesOpaqueNativeIDs(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, ids := range glmOpaqueIDCases {
			for _, mode := range []string{"trusted", "restored", "untrusted", "cross-model", "cross-account"} {
				t.Run(model+"/"+ids.name+"/"+mode, func(t *testing.T) {
					ev := run.Case(t, "GLM05-opaque-ids-"+t.Name())
					w := newGLMResponsesWorld(t, nil)
					events := glmOpaqueIDEvents(t, ids)
					ev.Fixture("native-ids.json", mustMarshal(t, map[string]any{"source": "official Responses schema, plain strings; synthetic edge IDs", "callIDs": ids.calls, "itemIDs": ids.items, "events": events}))
					enqueue(ev, w, glmResponsesReply(t, events))
					enqueue(ev, w, glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}))
					req, target := glmResponsesToolRequest(), glmResponsesTarget(model)
					first, err := w.client.Complete(ctxFor(t), textScope("glm-opaque-first"), target, req, nil)
					ev.Record("first-result", first)
					if !ev.Check("both opaque native calls returned", err == nil && len(first.Message.Content) == 2, "%v %+v", err, first.Message) {
						return
					}
					message := first.Message
					want := ai.NativeStateDowngrades{}
					if mode != "trusted" {
						stored := persist(t, message)
						ev.Record("stored-record", json.RawMessage(stored))
						var env *ai.NativeStateEnvelope
						message, env, _ = restore(t, stored)
						if env == nil {
							t.Fatal("missing envelope")
						}
						if mode == "untrusted" {
							want.NoEnvelope = 1
						} else {
							if mode == "cross-account" {
								env.AccountScopeID = "another-history-account"
								want.AccountMismatch = 1
							}
							if mode == "cross-model" {
								target.ModelID = map[string]string{"glm-5.3": "glm-5.2", "glm-5.2": "glm-5.3"}[model]
								want.CrossModel = 1
							}
							state, err := ai.TrustNativeState(textScope("glm-opaque-restore"), *env)
							ev.Check("Host trust established", err == nil, "%v", err)
							message.NativeState = state
						}
					}
					before := persist(t, message)
					req.Messages = append(req.Messages, message)
					publicIDs := []string{}
					for i, block := range message.Content {
						call := block.(ai.ToolCall)
						validated, err := ai.ValidateToolCall(message, i, req.Tools)
						ev.Check("Host validates opaque public ID", err == nil && validated.ID() == call.ID, "%v %+v", err, call)
						publicIDs = append(publicIDs, call.ID)
						req.Messages = append(req.Messages, ai.ToolResultText(call, fmt.Sprintf("result-%d", i), false))
					}
					ev.Check("public function IDs are distinct", publicIDs[0] != publicIDs[1], "%v", publicIDs)
					if ids.name == "ordinary" {
						ev.Check("ordinary captured public spelling unchanged", slices.Equal(publicIDs, []string{"call_alpha|fc_alpha", "call_beta|fc_beta"}), "%v", publicIDs)
					}
					second := drain(w.client.Stream(ctxFor(t), textScope("glm-opaque-second"), target, req, nil))
					second.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("history continues with expected native-state policy", second.err == nil && second.result.Metadata.NativeStateDowngrades == want, "%v %+v", second.err, second.result.Metadata)
					ev.Check("original history state immutable", string(before) == string(persist(t, message)), "history changed")
					if len(w.provider.Requests()) != 2 {
						return
					}
					var input []map[string]any
					mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
					var calls, results, outputs []string
					for _, item := range input {
						switch item["type"] {
						case "function_call":
							calls = append(calls, item["call_id"].(string))
							_, has := item["id"]
							ev.Check("output item IDs are not replayed", !has, "%+v", item)
						case "function_call_output":
							results = append(results, item["call_id"].(string))
							outputs = append(outputs, item["output"].(string))
						}
					}
					ev.Check("native call IDs replay exactly without collision", slices.Equal(calls, ids.calls), "got %v want %v", calls, ids.calls)
					ev.Check("each Host result answers its exact call", slices.Equal(results, ids.calls) && slices.Equal(outputs, []string{"result-0", "result-1"}), "calls=%v results=%v outputs=%v", calls, results, outputs)
				})
			}
		}
	}
}

func TestGLMResponsesCrossOriginOpaqueIDs(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, ids := range glmOpaqueIDCases[1:] {
			t.Run(model+"/"+ids.name, func(t *testing.T) {
				ev := run.Case(t, "GLM05-cross-origin-ids-"+t.Name())
				w := newGLMResponsesWorld(t, nil)
				enqueue(ev, w, glmResponsesReply(t, []json.RawMessage{json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[]}}`)}))
				req := glmResponsesToolRequest()
				message := ai.AssistantMessage{Provider: ai.ProviderOpenAI, API: ai.APIOpenAICompletions, Model: "synthetic-chat-history", StopReason: ai.StopReasonToolUse}
				for _, id := range ids.calls {
					message.Content = append(message.Content, ai.ToolCall{ID: id, Name: "lookup_probe_code", Arguments: ai.UncheckedArguments(`{"label":"source"}`)})
				}
				req.Messages = append(req.Messages, message)
				for i, block := range message.Content {
					req.Messages = append(req.Messages, ai.ToolResultText(block.(ai.ToolCall), fmt.Sprintf("source-%d", i), false))
				}
				o := drain(w.client.Stream(ctxFor(t), textScope("glm-cross-origin-ids"), glmResponsesTarget(model), req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				ev.Check("authorized cross-origin history continues", o.err == nil, "%v", o.err)
				if len(w.provider.Requests()) != 1 {
					return
				}
				var input []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[0].Body)["input"], &input)
				var calls, results, outputs []string
				for _, item := range input {
					switch item["type"] {
					case "function_call":
						calls = append(calls, item["call_id"].(string))
					case "function_call_output":
						results = append(results, item["call_id"].(string))
						outputs = append(outputs, item["output"].(string))
					}
				}
				ev.Check("GLM target retains both opaque source IDs without collision", slices.Equal(calls, ids.calls), "got %v want %v", calls, ids.calls)
				ev.Check("cross-origin result pairing remains valid", slices.Equal(calls, results) && slices.Equal(outputs, []string{"source-0", "source-1"}), "%v %v %v", calls, results, outputs)
			})
		}
	}
}

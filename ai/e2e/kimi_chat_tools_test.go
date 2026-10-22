package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func kimiChatToolRequest() ai.Request {
	r := kimiChatRequest()
	r.Tools = []ai.Tool{{Name: "lookup", Description: "Look up a value.", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`)}}
	return r
}
func TestKimiChatToolsAndHistory(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, mode := range []string{"native", "restored", "untrusted", "cross-account", "cross-market"} {
			t.Run(string(market)+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K06-K07-K08-")+mode)
				w := newKimiChatWorld(t, market, nil)
				frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"reasoning_content":"Check tools."}}]}`), json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"opaque|α/tool-a","function":{"name":"lookup","arguments":"{\"q\":"}},{"index":1,"id":"opaque|α/tool-b","function":{"name":"lookup","arguments":"{\"q\":\"b\"}"}}]}}]}`), json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"a\"}"}}]},"finish_reason":"tool_calls"}]}`)}
				ev.Fixture("tools.json", mustMarshal(t, frames))
				enqueue(ev, w, chatEvents(t, frames, ""))
				enqueue(ev, w, kimiChatReply(t))
				req := kimiChatToolRequest()
				first, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), kimiChatTarget(), req, ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{Mode: "required"})})
				ev.Record("first", first)
				if !ev.Check("required tools complete with exact opaque ids", err == nil && first.Message.StopReason == ai.StopReasonToolUse && len(first.Message.Content) == 3, "%v %+v", err, first.Message) {
					return
				}
				m := first.Message
				if mode != "native" {
					stored := persist(t, m)
					ev.Record("stored", json.RawMessage(stored))
					var env *ai.NativeStateEnvelope
					m, env, _ = restore(t, stored)
					if mode != "untrusted" && env != nil {
						if mode == "cross-account" {
							env.AccountScopeID = "different-account"
						}
						if mode == "cross-market" {
							if market == "moonshot" {
								env.ProviderID = "moonshot-cn"
								m.Provider = "moonshot-cn"
							} else {
								env.ProviderID = "moonshot"
								m.Provider = "moonshot"
							}
						}
						state, e := ai.TrustNativeState(textScope(t.Name()+"-restore"), *env)
						if e != nil {
							t.Fatal(e)
						}
						m.NativeState = state
					}
				}
				req.Messages = append(req.Messages, m)
				for _, block := range m.Content {
					if call, ok := block.(ai.ToolCall); ok {
						req.Messages = append(req.Messages, ai.ToolResultText(call, "found", false))
					}
				}
				second := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-second"), kimiChatTarget(), req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("history continues", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
					return
				}
				var msgs []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["messages"], &msgs)
				a := msgs[2]
				native := mode == "native" || mode == "restored"
				ev.Check("reasoning only replayed with matching provenance", (a["reasoning_content"] == "Check tools.") == native, "%v", a)
				calls := a["tool_calls"].([]any)
				ev.Check("opaque ids and matching results remain paired", len(calls) == 2 && calls[0].(map[string]any)["id"] == msgs[3]["tool_call_id"] && calls[1].(map[string]any)["id"] == msgs[4]["tool_call_id"], "%v", msgs)
				if native {
					ev.Check("opaque original bytes retained", calls[0].(map[string]any)["id"] == "opaque|α/tool-a" && calls[1].(map[string]any)["id"] == "opaque|α/tool-b", "%v", calls)
				}
			})
		}
	}
}

func TestKimiChatStreamTerminals(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, sc := range []struct {
			name, frame string
			required    bool
			code        ai.Code
			stop        ai.StopReason
		}{
			{"plain-stop", `{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}]}`, false, "", ai.StopReasonStop},
			{"required-no-tools", `{"choices":[{"delta":{"content":"ok"},"finish_reason":"stop"}]}`, true, ai.CodeProtocol, ""},
			{"required-length", `{"choices":[{"delta":{"content":"ok"},"finish_reason":"length"}]}`, true, "", ai.StopReasonLength},
			{"missing-finish", `{"choices":[{"delta":{"content":"ok"}}]}`, false, ai.CodeProtocol, ""},
			{"empty-tool-finish", `{"choices":[{"delta":{},"finish_reason":"tool_calls"}]}`, false, ai.CodeProtocol, ""},
			{"malformed-arguments", `{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"lookup","arguments":"{\"q\":"}}]},"finish_reason":"tool_calls"}]}`, false, ai.CodeProtocol, ""},
			{"nonobject-arguments", `{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"lookup","arguments":"[]"}}]},"finish_reason":"tool_calls"}]}`, false, ai.CodeProtocol, ""},
			{"tool-stop-conflict", `{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"lookup","arguments":"{}"}}]},"finish_reason":"stop"}]}`, false, ai.CodeProtocol, ""},
			{"duplicate-id", `{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"lookup","arguments":"{}"}},{"index":1,"id":"a","function":{"name":"lookup","arguments":"{}"}}]},"finish_reason":"tool_calls"}]}`, false, ai.CodeProtocol, ""},
			{"conflicting-id", `{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"a","function":{"name":"lookup","arguments":"{}"}},{"index":0,"id":"b","function":{"name":"lookup"}}]},"finish_reason":"tool_calls"}]}`, false, ai.CodeProtocol, ""},
			{"filtered", `{"choices":[{"delta":{},"finish_reason":"content_filter"}]}`, true, ai.CodeUpstreamError, ""},
			{"error", `{"error":{"type":"engine_overloaded","message":"private prompt text"}}`, false, ai.CodeUpstreamError, ""},
			{"malformed-json", `{"choices":`, false, ai.CodeProtocol, ""},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K05-K12-")+sc.name)
				w := newKimiChatWorld(t, market, nil)
				reply := chatEvents(t, []json.RawMessage{}, "")
				reply.Chunks = [][]byte{[]byte("data: " + sc.frame + "\n\n")}
				enqueue(ev, w, reply)
				opts := ai.ChatOptions{}
				if sc.required {
					opts.ToolChoice = ai.Value(ai.ChatToolChoice{Mode: "required"})
				}
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatToolRequest(), opts))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if sc.code == "" {
					ev.Check("successful terminal preserved", o.err == nil && o.result.Message.StopReason == sc.stop, "%v", o.err)
				} else {
					ev.Check("classified single failure", errors.Is(o.err, &ai.Error{Code: sc.code, Phase: ai.PhaseStream}), "%v", o.err)
				}
				ev.Check("stream never replayed", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 1, "%+v", o.result.Metadata.Attempts)
			})
		}
	}
}

func TestKimiChatMalformedFrames(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, sc := range []struct {
			name   string
			frames []string
		}{
			{"invalid-choices", []string{`{"choices":{}}`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`}},
			{"nonobject-event", []string{`[]`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`}},
			{"malformed-tools", []string{`{"choices":[{"delta":{"tool_calls":{}}}]}`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`}},
			{"malformed-tool-entry", []string{`{"choices":[{"delta":{"tool_calls":[null]}}]}`, `{"choices":[{"delta":{},"finish_reason":"stop"}]}`}},
			{"contradictory-terminals", []string{`{"choices":[{"delta":{},"finish_reason":"stop"}]}`, `{"choices":[{"delta":{},"finish_reason":"tool_calls"}]}`}},
			{"content-after-terminal", []string{`{"choices":[{"delta":{},"finish_reason":"stop"}]}`, `{"choices":[{"delta":{"content":"late"}}]}`}},
		} {
			t.Run(string(market)+"/"+sc.name, func(t *testing.T) {
				ev := run.Case(t, kimiChatID(market, "K12-malformed-")+sc.name)
				w := newKimiChatWorld(t, market, nil)
				var frames []json.RawMessage
				for _, f := range sc.frames {
					frames = append(frames, json.RawMessage(f))
				}
				enqueue(ev, w, chatEvents(t, frames, ""))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiChatTarget(), kimiChatRequest(), nil))
				o.record(ev)
				ev.Check("malformed stream cannot succeed", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
			})
		}
	}
}

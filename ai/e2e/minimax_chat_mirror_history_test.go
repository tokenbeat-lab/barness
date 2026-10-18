package e2e

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxChatObservedMirrorHistory(t *testing.T) {
	for _, mode := range []string{"in-process", "trusted-restored", "untrusted", "cross-account", "cross-model", "cross-region"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "MM10-actual-chat-mirror-history-"+mode)
			first, err := os.ReadFile("testdata/minimax-observed/chat-mirror-tool-first-frame.sse")
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("actual-first-frame.sse", first)
			tail := []json.RawMessage{
				json.RawMessage(`{"choices":[{"delta":{"reasoning_content":" inspect.","reasoning_details":[{"type":"reasoning.text","id":"reasoning-text-1","format":"MiniMax-response-v1","index":0,"text":" inspect."}],"tool_calls":[{"index":0,"id":"mirror-call","function":{"name":"lookup","arguments":"{\"q\":"}}]}}]}`),
				json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"a\"}"}}]},"finish_reason":"tool_calls"}]}`),
				json.RawMessage(`{"choices":[],"usage":{"prompt_tokens":12,"completion_tokens":8,"prompt_tokens_details":{"cached_tokens":2},"completion_tokens_details":{"reasoning_tokens":3}}}`),
			}
			ev.Fixture("synthetic-completion-frames.json", mustMarshal(t, tail))
			w := newMiniMaxChatWorld(t, nil)
			reply := chatEvents(t, tail, provider.FramingBytewise)
			enqueue(ev, w, provider.SSE(append([][]byte{first}, reply.Chunks...)))
			enqueue(ev, w, miniMaxChatReply(t, "MiniMax-M3"))
			req := miniMaxChatRequest()
			req.Tools = []ai.Tool{{Name: "lookup", Description: "Look up q.", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"}},"required":["q"]}`)}}
			initial, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), miniMaxChatTarget("MiniMax-M3"), req, nil)
			ev.Record("first-result", initial)
			if !ev.Check("mirrored thinking and complete tool finish", err == nil && initial.Message.StopReason == ai.StopReasonToolUse && len(initial.Message.Content) == 2, "%v %+v", err, initial.Message) {
				return
			}
			call, err := ai.ValidateToolCall(initial.Message, 1, req.Tools)
			ev.Check("Host gets complete executable tool", err == nil && string(call.Arguments()) == `{"q":"a"}`, "%v", err)
			message := initial.Message
			if mode != "in-process" {
				stored := persist(t, message)
				ev.Record("stored-record", json.RawMessage(stored))
				var env *ai.NativeStateEnvelope
				message, env, _ = restore(t, stored)
				if env == nil {
					t.Fatal("missing native envelope")
				}
				if mode == "cross-account" {
					env.AccountScopeID = "other-account"
				}
				if mode == "cross-region" {
					env.ProviderID = "minimax-cn"
				}
				if mode != "untrusted" {
					message.NativeState, err = ai.TrustNativeState(textScope(t.Name()+"-trust"), *env)
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			target := miniMaxChatTarget("MiniMax-M3")
			if mode == "cross-model" {
				target.ModelID = "MiniMax-M3.1-Flash-Preview"
			}
			req.Messages = append(req.Messages, message, ai.ToolResultText(initial.Message.Content[1].(ai.ToolCall), "A", false))
			second := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-second"), target, req, nil))
			second.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("native continuation sends once", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
				return
			}
			var messages []map[string]json.RawMessage
			mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["messages"], &messages)
			var assistant map[string]json.RawMessage
			for _, m := range messages {
				if string(m["role"]) == `"assistant"` {
					assistant = m
				}
			}
			trusted := mode == "in-process" || mode == "trusted-restored"
			if trusted {
				details := json.RawMessage(`[{"type":"reasoning.text","id":"reasoning-text-1","format":"MiniMax-response-v1","index":0,"text":"The user wants me to inspect."}]`)
				ev.Check("original native details and mirror both replay", jsonEqual(assistant["reasoning_details"], details) && string(assistant["reasoning_content"]) == `"The user wants me to inspect."`, "%v", assistant)
			} else {
				ev.Check("unqualified native mirror is stripped", len(assistant["reasoning_details"]) == 0 && len(assistant["reasoning_content"]) == 0, "%v", assistant)
			}
		})
	}
}

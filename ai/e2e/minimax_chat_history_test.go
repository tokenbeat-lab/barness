package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxChatToolHistory(t *testing.T) {
	for _, format := range []string{"reasoning_content", "reasoning_details"} {
		for _, mode := range []string{"in-process", "trusted-restored", "untrusted", "cross-account", "cross-model"} {
			t.Run(format+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM02-history-"+format+"-"+mode)
				w := newMiniMaxChatWorld(t, nil)
				reasoning := `"reasoning_content":"Check both."`
				details := json.RawMessage(`[{"type":"reasoning.text","id":"r1","text":"Check.","signature":"s1"},{"type":"reasoning.text","id":"r2","text":"Next.","signature":"s2"}]`)
				if format == "reasoning_details" {
					reasoning = `"reasoning_details":` + string(details)
				}
				frames := []json.RawMessage{
					json.RawMessage(`{"id":"minimax-tools","choices":[{"delta":{` + reasoning + `,"content":"Looking.","tool_calls":[{"index":1,"id":"call-b","function":{"name":"lookup","arguments":"{\"q\":"}},{"index":0,"id":"call-a","function":{"name":"lookup","arguments":"{\"q\":"}}]}}]}`),
					json.RawMessage(`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\"a\"}"}},{"index":1,"function":{"arguments":"\"b\"}"}}]},"finish_reason":"tool_calls"}]}`),
				}
				req := miniMaxChatRequest()
				req.Tools = []ai.Tool{{Name: "lookup", Description: "Look up q.", Parameters: json.RawMessage(`{"type":"object","properties":{"q":{"type":"string"},"thinking":{"type":"string"},"service_tier":{"type":"string"},"strict":{"type":"boolean"}},"required":["q"]}`)}}
				ev.Fixture("synthetic-tool-frames.json", mustMarshal(t, frames))
				enqueue(ev, w, chatEvents(t, frames, ""))
				enqueue(ev, w, miniMaxChatReply(t, "MiniMax-M3"))
				first, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), miniMaxChatTarget("MiniMax-M3"), req, nil)
				ev.Record("first-result", first)
				if !ev.Check("parallel tools finish", err == nil && first.Message.StopReason == ai.StopReasonToolUse && len(first.Message.Content) == 4, "%v %+v", err, first.Message) {
					return
				}
				var calls []ai.ToolCall
				for i, block := range first.Message.Content {
					c, ok := block.(ai.ToolCall)
					if !ok {
						continue
					}
					want := `{"q":"b"}`
					if c.ID == "call-a" {
						want = `{"q":"a"}`
					}
					call, err := ai.ValidateToolCall(first.Message, i, req.Tools)
					ev.Check("Host receives complete independent tool", err == nil && string(call.Arguments()) == want, "%v %+v", err, call)
					calls = append(calls, c)
				}
				message := first.Message
				if mode != "in-process" {
					stored := persist(t, message)
					ev.Record("stored-record", json.RawMessage(stored))
					var env *ai.NativeStateEnvelope
					message, env, _ = restore(t, stored)
					if env == nil {
						t.Fatal("missing envelope")
					}
					if mode == "cross-account" {
						env.AccountScopeID = "other-account"
					}
					if mode != "untrusted" {
						state, err := ai.TrustNativeState(textScope(t.Name()+"-trust"), *env)
						if err != nil {
							t.Fatal(err)
						}
						message.NativeState = state
					}
				}
				target := miniMaxChatTarget("MiniMax-M3")
				if mode == "cross-model" {
					target.ModelID = "MiniMax-M3.1-Flash-Preview"
				}
				req.Messages = append(req.Messages, message, ai.ToolResultText(calls[0], "B", false), ai.ToolResultText(calls[1], "A", false))
				second := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-second"), target, req, nil))
				second.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("history continues", second.err == nil && len(w.provider.Requests()) == 2, "%v", second.err) {
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
				wantText := `"Looking."`
				if !trusted && format == "reasoning_content" {
					wantText = `"Looking.Check both."`
				}
				ev.Check("complete text and paired calls replayed", string(assistant["content"]) == wantText && len(assistant["tool_calls"]) > 0, "%v", assistant)
				if trusted && format == "reasoning_details" {
					ev.Check("native reasoning entries preserved separately", jsonEqual(assistant[format], details), "%s", assistant[format])
				}
				if trusted && format == "reasoning_content" {
					ev.Check("visible thinking replayed under original field", string(assistant[format]) == `"Check both."`, "%s", assistant[format])
				}
				if !trusted {
					ev.Check("native reasoning stripped at boundary", len(assistant["reasoning_details"]) == 0 && len(assistant["reasoning_content"]) == 0, "%v", assistant)
				}
				ev.Check("no DeepSeek empty field", format != "reasoning_details" || len(assistant["reasoning_content"]) == 0, "%v", assistant)
			})
		}
	}
}

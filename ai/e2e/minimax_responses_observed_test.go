package e2e

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxResponsesObservedNative(t *testing.T) {
	for _, mode := range []string{"in-process", "restored", "untrusted", "cross-account", "cross-model", "cross-region"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "MM10-actual-responses-native-"+mode)
			raw, err := os.ReadFile("testdata/minimax-observed/responses-tool-complete.sse")
			if err != nil {
				t.Fatal(err)
			}
			native, err := os.ReadFile("testdata/minimax-observed/responses-reasoning-item.json")
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("actual-tool-stream.sse", raw)
			ev.Fixture("actual-native-item.json", native)
			w := newMiniMaxResponsesWorld(t, nil)
			enqueue(ev, w, provider.SSE([][]byte{raw}))
			enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
			req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic arithmetic task. You MUST call the add function with a=17 and b=25 before answering. Think through the calculation first.")}, Tools: []ai.Tool{{Name: "add", Description: "Add the two given integers and return their sum.", Parameters: json.RawMessage(`{"type":"object","properties":{"a":{"type":"integer"},"b":{"type":"integer"}},"required":["a","b"],"additionalProperties":false}`)}}}
			target := minimaxResponsesTarget("MiniMax-M3")
			first, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), target, req, nil)
			ev.Record("first-result", first)
			if !ev.Check("complete actual stream retains reasoning text and tool", err == nil && len(first.Message.Content) == 3 && first.Message.StopReason == ai.StopReasonToolUse, "%v", err) {
				return
			}
			thinking := first.Message.Content[0].(ai.Thinking)
			ev.Check("entire actual native item retained", jsonEqual([]byte(thinking.Signature), native), "native item changed")
			ev.Check("observed known usage retained", first.Message.Usage.Input == 246 && first.Message.Usage.CacheRead == 203 && first.Message.Usage.Output == 109, "%+v", first.Message.Usage)
			call, err := ai.ValidateToolCall(first.Message, 2, req.Tools)
			ev.Check("Host validates actual tool", err == nil && string(call.Arguments()) == `{"a":17,"b":25}`, "%v", err)
			message := first.Message
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
			if mode == "cross-model" {
				target.ModelID = "MiniMax-M3.1-Flash-Preview"
			}
			req.Messages = append(req.Messages, message, ai.ToolResultText(first.Message.Content[2].(ai.ToolCall), "42", false))
			second := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-second"), target, req, nil))
			second.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("documented text projection reaches transport", second.err == nil && len(w.provider.Requests()) == 2 && len(second.result.Metadata.Attempts) == 1, "%v", second.err) {
				return
			}
			var input []map[string]json.RawMessage
			mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
			found := false
			calls, results := 0, 0
			for _, item := range input {
				ev.Check("output ciphertext is never an input field", len(item["encrypted_content"]) == 0, "unexpected output-only state")
				switch string(item["type"]) {
				case `"reasoning"`:
					want := map[string]any{"type": "reasoning", "summary": []map[string]string{{"type": "summary_text", "text": thinking.Thinking}}}
					found = jsonEqual(mustMarshal(t, item), mustMarshal(t, want))
				case `"function_call"`:
					if string(item["call_id"]) == `"call_01a14cb4d5b67721b0350ba2"` {
						calls++
					}
				case `"function_call_output"`:
					if string(item["call_id"]) == `"call_01a14cb4d5b67721b0350ba2"` {
						results++
					}
				}
			}
			trusted := mode == "in-process" || mode == "restored"
			ev.Check("native projection qualification is exact", found == trusted, "found=%v trusted=%v", found, trusted)
			ev.Check("actual call IDs pair exactly", calls == 1 && results == 1, "calls=%d results=%d", calls, results)
		})
	}
}

func TestMiniMaxResponsesObservedEncryptedType(t *testing.T) {
	for _, bad := range []string{"null", "42", "true", "[]", "{}"} {
		t.Run(bad, func(t *testing.T) {
			ev := run.Case(t, "MM10-actual-responses-encrypted-type-"+bad)
			native, err := os.ReadFile("testdata/minimax-observed/responses-reasoning-item.json")
			if err != nil {
				t.Fatal(err)
			}
			var item map[string]json.RawMessage
			mustUnmarshal(t, native, &item)
			item["encrypted_content"] = json.RawMessage(bad)
			events := []string{`{"type":"response.output_item.done","output_index":0,"item":` + string(mustMarshal(t, item)) + `}`, minimaxResponsesCompleted}
			ev.Fixture("synthetic-invalid-events.json", mustMarshal(t, events))
			w := newMiniMaxResponsesWorld(t, nil)
			enqueue(ev, w, minimaxResponsesReply(t, events...))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
			o.record(ev)
			ev.Check("unobserved encrypted type fails protocol", o.err != nil && o.result.Message.StopReason == ai.StopReasonError, "%v", o.err)
		})
	}
}

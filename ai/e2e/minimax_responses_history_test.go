package e2e

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxResponsesNativeRoundTrip(t *testing.T) {
	for _, mode := range []string{"trusted", "restored", "untrusted", "cross-model", "cross-account"} {
		for _, ids := range []struct {
			name         string
			calls, items []string
		}{
			{"collision", []string{"same|call%7C", "same"}, []string{"item%", "call%7C|item%"}},
			{"unicode-long", []string{"同名|甲%", strings.Repeat("opaque%|长", 100)}, []string{"条目|甲%", "not-fc-" + strings.Repeat("条目%|", 100)}},
		} {
			t.Run(mode+"/"+ids.name, func(t *testing.T) {
				ev := run.Case(t, "MM03-native-"+t.Name())
				w := newMiniMaxResponsesWorld(t, nil)
				native := json.RawMessage(`{"type":"reasoning","id":"raw|推理%","summary":[],"content":[{"type":"reasoning_text","text":"first exact part"},{"type":"reasoning_text","text":"second exact part"}]}`)
				events := []string{`{"type":"response.output_item.done","output_index":0,"item":` + string(native) + `}`}
				for i := range ids.calls {
					events = append(events, string(mustMarshal(t, map[string]any{"type": "response.output_item.done", "output_index": i + 1, "item": map[string]any{"type": "function_call", "id": ids.items[i], "call_id": ids.calls[i], "name": "lookup_probe_code", "arguments": fmt.Sprintf(`{"label":"%d"}`, i)}})))
				}
				events = append(events, minimaxResponsesCompleted)
				ev.Fixture("synthetic-raw-events.json", mustMarshal(t, map[string]any{"source": "official MiniMax output/input schema, exact part projection is provisional until M3 live ticket07", "events": events}))
				enqueue(ev, w, minimaxResponsesReply(t, events...))
				enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
				req := glmResponsesToolRequest()
				target := minimaxResponsesTarget("MiniMax-M3")
				first, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), target, req, nil)
				ev.Record("first", first)
				if !ev.Check("raw native reasoning and tools returned", err == nil && len(first.Message.Content) == 3, "%v %+v", err, first.Message) {
					return
				}
				b := first.Message.Content[0].(ai.Thinking)
				ev.Check("full output reasoning is retained unchanged", jsonEqual([]byte(b.Signature), native), "%s", b.Signature)
				message := first.Message
				if mode != "trusted" {
					stored := persist(t, message)
					var envelope *ai.NativeStateEnvelope
					message, envelope, _ = restore(t, stored)
					if mode != "untrusted" {
						if mode == "cross-account" {
							envelope.AccountScopeID = "another-account"
						}
						if mode == "cross-model" {
							target.ModelID = "MiniMax-M3.1-Flash-Preview"
						}
						state, err := ai.TrustNativeState(textScope(t.Name()+"-restore"), *envelope)
						if err != nil {
							t.Fatal(err)
						}
						message.NativeState = state
					}
				}
				req.Messages = append(req.Messages, message)
				public := []string{}
				for i, block := range message.Content {
					if call, ok := block.(ai.ToolCall); ok {
						valid, err := ai.ValidateToolCall(message, i, req.Tools)
						ev.Check("Host validates opaque ID", err == nil && valid.ID() == call.ID, "%v", err)
						public = append(public, call.ID)
						req.Messages = append(req.Messages, ai.ToolResultText(call, fmt.Sprintf("result-%d", len(public)-1), false))
					}
				}
				ev.Check("composite IDs do not collide", len(public) == 2 && public[0] != public[1], "%v", public)
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-second"), target, req, nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				if !ev.Check("synthetic schema history continues", o.err == nil && len(w.provider.Requests()) == 2, "%v", o.err) {
					return
				}
				var input []map[string]any
				mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
				calls, results := []string{}, []string{}
				var reasoning map[string]any
				for _, item := range input {
					switch item["type"] {
					case "function_call":
						calls = append(calls, item["call_id"].(string))
						_, has := item["id"]
						ev.Check("output item ID is not an input field", !has, "%v", item)
					case "function_call_output":
						results = append(results, item["call_id"].(string))
					case "reasoning":
						reasoning = item
					}
				}
				ev.Check("exact call IDs pair without collision", slices.Equal(calls, ids.calls) && slices.Equal(results, ids.calls), "calls=%v results=%v", calls, results)
				if mode == "trusted" || mode == "restored" {
					ev.Check("lossless documented input projection", jsonEqual(mustMarshal(t, reasoning), json.RawMessage(`{"type":"reasoning","summary":[{"type":"summary_text","text":"first exact part"},{"type":"summary_text","text":"second exact part"}]}`)), "%v", reasoning)
				} else {
					ev.Check("untrusted native reasoning stripped", reasoning == nil, "%v", reasoning)
				}
			})
		}
	}
}

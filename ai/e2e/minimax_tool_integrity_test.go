package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxChatFinalToolIntegrity(t *testing.T) {
	for _, c := range []struct {
		name, id, nameFirst, nameLate, args string
		valid                               bool
	}{
		{"incomplete", "call-a", "lookup", "", `{"text":"hel`, false},
		{"malformed", "call-a", "lookup", "", `{oops}`, false},
		{"array", "call-a", "lookup", "", `[]`, false},
		{"missing-id", "", "lookup", "", `{}`, false},
		{"missing-name", "call-a", "", "", `{}`, false},
		{"changed-name", "call-a", "lookup", "other", `{}`, false},
		{"empty-arguments", "call-a", "lookup", "", ``, true},
		{"late-id", "", "lookup", "", `{}`, true},
		{"late-name", "call-a", "", "lookup", `{}`, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM10-chat-tool-integrity-"+c.name)
			w := newMiniMaxChatWorld(t, nil)
			call := map[string]any{"index": 0, "id": c.id, "function": map[string]any{"name": c.nameFirst, "arguments": c.args}}
			first, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]any{"content": "known progress", "tool_calls": []any{call}}}}})
			frames := []json.RawMessage{first}
			if c.nameLate != "" || c.name == "late-id" {
				lateID := ""
				if c.name == "late-id" {
					lateID = "call-a"
				}
				later, _ := json.Marshal(map[string]any{"choices": []any{map[string]any{"delta": map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": lateID, "function": map[string]any{"name": c.nameLate}}}}}}})
				frames = append(frames, later)
			}
			frames = append(frames, json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":5,"completion_tokens":2}}`))
			enqueue(ev, w, chatEvents(t, frames, ""))
			req := miniMaxChatRequest()
			req.Tools = []ai.Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`)}}
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Fixture("provider-frames.json", mustMarshal(t, frames))
			ev.Check("only complete pairable tool calls succeed", (o.err == nil) == c.valid, "%v %+v", o.err, o.result.Message)
			found := false
			for i, p := range o.result.Message.Content {
				if _, ok := p.(ai.ToolCall); ok {
					found = true
					_, err := ai.ValidateToolCall(o.result.Message, i, req.Tools)
					ev.Check("Host execution follows terminal integrity", (err == nil) == c.valid, "%v", err)
				}
			}
			ev.Check("partial tool and visible progress retained", found && textOf(o.result.Message) == "known progress", "%+v", o.result.Message)
			if c.name != "changed-name" {
				ev.Check("known final usage retained", o.result.Message.Usage.Input == 5 && o.result.Message.Usage.Output == 2, "%+v", o.result.Message.Usage)
			}
			ev.Check("one actual attempt without stream replay", len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%+v", o.result.Metadata)
		})
	}
}

func TestMiniMaxNativeCallIDsAreDistinct(t *testing.T) {
	for _, route := range []string{"responses", "messages"} {
		for _, duplicate := range []bool{false, true} {
			name := fmt.Sprintf("%s-%t", route, duplicate)
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "MM10-native-distinct-call-"+name)
				id2 := "call-b"
				if duplicate {
					id2 = "call-a"
				}
				var frames []json.RawMessage
				if route == "responses" {
					output := []any{}
					for i, id := range []string{"call-a", id2} {
						item := map[string]any{"type": "function_call", "id": fmt.Sprintf("item-%d", i), "call_id": id, "name": "lookup", "arguments": "{}"}
						raw, _ := json.Marshal(map[string]any{"type": "response.output_item.done", "output_index": i, "item": item})
						frames = append(frames, raw)
						output = append(output, item)
					}
					raw, _ := json.Marshal(map[string]any{"type": "response.completed", "response": map[string]any{"model": "MiniMax-M3", "status": "completed", "output": output}})
					frames = append(frames, raw)
					w := newMiniMaxResponsesWorld(t, nil)
					chunks := []string{}
					for _, f := range frames {
						chunks = append(chunks, string(f))
					}
					enqueue(ev, w, minimaxResponsesReply(t, chunks...))
					req := minimaxResponsesRequest()
					req.Tools = []ai.Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object"}`)}}
					o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), req, nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("distinct native call ids required", (o.err == nil) == !duplicate, "%v", o.err)
					for i, p := range o.result.Message.Content {
						if _, ok := p.(ai.ToolCall); ok {
							_, e := ai.ValidateToolCall(o.result.Message, i, req.Tools)
							ev.Check("ambiguous result never executable", (e == nil) == !duplicate, "%v", e)
						}
					}
				} else {
					frames = append(frames, json.RawMessage(`{"type":"message_start","message":{"id":"native-tools","model":"MiniMax-M3","role":"assistant","content":[]}}`))
					for i, id := range []string{"call-a", id2} {
						raw, _ := json.Marshal(map[string]any{"type": "content_block_start", "index": i, "content_block": map[string]any{"type": "tool_use", "id": id, "name": "lookup", "input": map[string]any{}}})
						frames = append(frames, raw)
						frames = append(frames, json.RawMessage(fmt.Sprintf(`{"type":"content_block_stop","index":%d}`, i)))
					}
					frames = append(frames, json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"tool_use"}}`), json.RawMessage(`{"type":"message_stop"}`))
					w := newMiniMaxMessagesWorld(t, nil)
					enqueue(ev, w, sseEvents(t, frames, provider.FramingLF))
					req := miniMaxMessagesRequest()
					req.Tools = []ai.Tool{{Name: "lookup", Parameters: json.RawMessage(`{"type":"object"}`)}}
					o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxMessagesTarget("MiniMax-M3"), req, nil))
					o.record(ev)
					ev.Record("requests", w.provider.Requests())
					ev.Check("distinct native call ids required", (o.err == nil) == !duplicate, "%v", o.err)
					for i, p := range o.result.Message.Content {
						if _, ok := p.(ai.ToolCall); ok {
							_, e := ai.ValidateToolCall(o.result.Message, i, req.Tools)
							ev.Check("ambiguous result never executable", (e == nil) == !duplicate, "%v", e)
						}
					}
				}
				ev.Fixture("provider-frames.json", mustMarshal(t, frames))
			})
		}
	}
}

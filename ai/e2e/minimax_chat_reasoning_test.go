package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"strings"
	"testing"
)

func TestMiniMaxChatUnknownReasoningFormat(t *testing.T) {
	for _, raw := range []string{
		`"reasoning_details":{"text":"unknown"}`,
		`"reasoning_details":[{"type":"unknown","text":"opaque"}]`,
		`"reasoning_content":"visible","reasoning_details":[{"type":"reasoning.text","text":"other"}]`,
		`"reasoning_content":true`,
		`"reasoning":"unknown field"`,
	} {
		t.Run(raw, func(t *testing.T) {
			ev := run.Case(t, "MM02-unknown-reasoning-"+t.Name())
			w := newMiniMaxChatWorld(t, nil)
			frames := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{` + raw + `},"finish_reason":"stop"}]}`)}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("unknown native format is never silently lost", o.err != nil && o.result.Message.StopReason == ai.StopReasonError && len(w.provider.Requests()) == 1, "%v %+v", o.err, o.result.Message)
		})
	}
}

func TestMiniMaxChatMixedReasoningFrames(t *testing.T) {
	for _, first := range []string{"content", "details"} {
		t.Run(first, func(t *testing.T) {
			ev := run.Case(t, "MM10-mixed-reasoning-"+first)
			w := newMiniMaxChatWorld(t, nil)
			content := json.RawMessage(`{"choices":[{"delta":{"reasoning_content":"first supported reasoning"}}]}`)
			details := json.RawMessage(`{"choices":[{"delta":{"reasoning_details":[{"type":"reasoning.text","text":"first supported reasoning"}]}}]}`)
			frames := []json.RawMessage{content, details, json.RawMessage(`{"choices":[{"delta":{"content":"answer"},"finish_reason":"stop"}]}`)}
			if first == "details" {
				frames[0], frames[1] = details, content
			}
			ev.Fixture("synthetic-frames.json", mustMarshal(t, frames))
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("cross-frame conflicting format fails", o.err != nil && o.result.Message.StopReason == ai.StopReasonError, "%v %+v", o.err, o.result.Message)
			text := ""
			signature := ""
			for _, part := range o.result.Message.Content {
				if v, ok := part.(ai.Thinking); ok {
					text += v.Thinking
					signature += v.Signature
				}
			}
			ev.Check("previous accepted reasoning remains available", text == "first supported reasoning" || strings.Contains(signature, "first supported reasoning"), "%q", text)
			ev.Check("one attempted stream, no replay", len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%+v", o.result.Metadata)
		})
	}
}

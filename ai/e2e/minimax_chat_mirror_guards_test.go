package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxChatMirrorConflicts(t *testing.T) {
	mirror := json.RawMessage(`{"choices":[{"delta":{"reasoning_content":"kept","reasoning_details":[{"type":"reasoning.text","id":"r","text":"kept"}]}}]}`)
	for _, bad := range []string{
		`{"reasoning_content":"left","reasoning_details":[{"type":"reasoning.text","text":"right"}]}`,
		`{"reasoning_content":"","reasoning_details":[{"type":"reasoning.text","text":"not empty"}]}`,
		`{"reasoning_content":"next"}`,
		`{"reasoning_details":[{"type":"reasoning.text","text":"next"}]}`,
		`{"reasoning_content":"next","reasoning_details":[{"type":"reasoning.summary","summary":"next"}]}`,
	} {
		t.Run(bad, func(t *testing.T) {
			ev := run.Case(t, "MM10-chat-mirror-conflict-"+t.Name())
			frames := []json.RawMessage{mirror, json.RawMessage(`{"choices":[{"delta":` + bad + `}]}`), json.RawMessage(`{"choices":[{"delta":{},"finish_reason":"stop"}]}`)}
			ev.Fixture("synthetic-conflicting-frames.json", mustMarshal(t, frames))
			w := newMiniMaxChatWorld(t, nil)
			enqueue(ev, w, chatEvents(t, frames, ""))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			visible := ""
			for _, block := range o.result.Message.Content {
				if v, ok := block.(ai.Thinking); ok {
					visible += v.Thinking
				}
			}
			ev.Check("conflict fails without losing accepted mirror", o.err != nil && visible == "kept" && len(o.result.Metadata.Attempts) == 1, "%v %q", o.err, visible)
		})
	}
}

func TestMiniMaxChatMirrorFinalPayload(t *testing.T) {
	for _, bad := range []string{"other", ""} {
		t.Run(bad, func(t *testing.T) {
			ev := run.Case(t, "MM10-chat-mirror-final-"+t.Name())
			w := newMiniMaxChatWorld(t, nil)
			hooked := w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["messages"] = []any{map[string]any{"role": "assistant", "content": "answer", "reasoning_content": bad, "reasoning_details": []any{map[string]any{"type": "reasoning.text", "text": "known"}}}}
				return ai.KeepPayload(), nil
			}})
			o := drain(hooked.Stream(ctxFor(t), textScope(t.Name()), miniMaxChatTarget("MiniMax-M3"), miniMaxChatRequest(), nil))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("mismatched payload mirror consumes no attempt", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

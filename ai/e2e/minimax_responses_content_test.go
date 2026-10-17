package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxResponsesAuthoritativeContent(t *testing.T) {
	for _, sc := range []struct {
		name           string
		events         []string
		thinking, text string
		length         bool
		failure        bool
	}{
		{name: "terminal-only", events: []string{`{"type":"response.completed","response":{"status":"completed","output":[{"id":"rs|原%","type":"reasoning","summary":[],"content":[{"type":"reasoning_text","text":"exact reasoning"}]},{"id":"msg|原%","type":"message","role":"assistant","content":[{"type":"output_text","text":"hello"}]}]}}`}, thinking: "exact reasoning", text: "hello"},
		{name: "done-only", events: []string{`{"type":"response.reasoning_text.done","output_index":0,"text":"reasoned"}`, `{"type":"response.output_text.done","output_index":1,"text":"hello"}`, minimaxResponsesCompleted}, thinking: "reasoned", text: "hello"},
		{name: "done-corrects-delta", events: []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"message","id":"m","content":[]}}`, `{"type":"response.output_text.delta","output_index":0,"delta":"wrong"}`, `{"type":"response.output_text.done","output_index":0,"text":"right"}`, `{"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"m","content":[]}}`, minimaxResponsesCompleted}, text: "right"},
		{name: "reasoning-length", events: []string{`{"type":"response.incomplete","response":{"status":"incomplete","incomplete_details":{"reason":"max_output_tokens"},"output":[{"type":"reasoning","id":"rs","summary":[],"content":[{"type":"reasoning_text","text":"only reasoning"}]}]}}`}, thinking: "only reasoning", length: true},
		{name: "object-content", events: []string{`{"type":"response.completed","response":{"status":"completed","output":[{"type":"reasoning","id":"rs","content":{"type":"reasoning_text","text":"wrong shape"}}]}}`}, failure: true},
		{name: "duplicate-terminal", events: []string{minimaxResponsesCompleted, minimaxResponsesCompleted}, failure: true},
		{name: "missing-slot-delta", events: []string{`{"type":"response.output_text.delta","output_index":0,"delta":"lost"}`, minimaxResponsesCompleted}, failure: true},
		{name: "wrong-delta-kind", events: []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"reasoning","id":"r","content":[]}}`, `{"type":"response.output_text.delta","output_index":0,"delta":"wrong"}`, minimaxResponsesCompleted}, failure: true},
		{name: "missing-terminal", events: []string{`{"type":"response.output_text.done","output_index":0,"text":"partial"}`}, failure: true},
		{name: "invalid-terminal-status", events: []string{`{"type":"response.completed","response":{"status":"in_progress","output":[]}}`}, failure: true},
		{name: "after-terminal", events: []string{minimaxResponsesCompleted, `{"type":"response.output_text.done","output_index":0,"text":"late"}`}, failure: true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM03-content-"+sc.name)
			ev.Fixture("synthetic-events.json", mustMarshal(t, sc.events))
			w := newMiniMaxResponsesWorld(t, nil)
			enqueue(ev, w, minimaxResponsesReply(t, sc.events...))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.failure {
				checkFailed(ev, o.result, o.err)
				return
			}
			ev.Check("valid stream succeeds", o.err == nil, "%v", o.err)
			var thinking, text string
			for _, b := range o.result.Message.Content {
				switch b := b.(type) {
				case ai.Thinking:
					thinking += b.Thinking
				case ai.Text:
					text += b.Text
				}
			}
			ev.Check("authoritative content retained once", thinking == sc.thinking && text == sc.text, "thinking=%q text=%q", thinking, text)
			if sc.length {
				ev.Check("reasoning-only budget is length", o.result.Message.StopReason == ai.StopReasonLength, "%v", o.result.Message.StopReason)
			}
		})
	}
}

func TestMiniMaxResponsesRetainsOpaqueReasoningOutput(t *testing.T) {
	ev := run.Case(t, "MM03-raw-reasoning-output")
	w := newMiniMaxResponsesWorld(t, nil)
	raw := `{"type":"reasoning","id":"rs","content":[{"type":"reasoning_text","text":"exact"}],"capture_marker":1.00000000000000001}`
	events := []string{`{"type":"response.output_item.done","output_index":0,"item":` + raw + `}`, minimaxResponsesCompleted}
	ev.Fixture("synthetic-events.json", mustMarshal(t, events))
	enqueue(ev, w, minimaxResponsesReply(t, events...))
	first, err := w.client.Complete(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil)
	ev.Record("result", first)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("original reasoning returned", err == nil && len(first.Message.Content) == 1, "%v", err) {
		return
	}
	b := first.Message.Content[0].(ai.Thinking)
	ev.Check("original raw bytes retained before input projection", b.Signature == raw, "%s", b.Signature)
	req := minimaxResponsesRequest()
	req.Messages = append(req.Messages, first.Message, ai.UserText("Continue."))
	o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-continue"), minimaxResponsesTarget("MiniMax-M3"), req, nil))
	o.record(ev)
	checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
	ev.Check("unknown native field is not silently removed on replay", len(w.provider.Requests()) == 1 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
}

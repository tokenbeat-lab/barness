package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxResponsesPartsAndOrder(t *testing.T) {
	for _, sc := range []struct {
		name    string
		events  []string
		want    string
		failure bool
	}{
		{"multiple-done-parts", []string{`{"type":"response.output_text.done","output_index":0,"content_index":1,"text":"world"}`, `{"type":"response.output_text.done","output_index":0,"content_index":0,"text":"hello "}`, minimaxResponsesCompleted}, "hello world", false},
		{"delta-after-other-part-done", []string{`{"type":"response.output_text.done","output_index":0,"content_index":0,"text":"hello "}`, `{"type":"response.output_text.delta","output_index":0,"content_index":1,"delta":"world"}`, `{"type":"response.output_text.done","output_index":0,"content_index":1,"text":"world"}`, minimaxResponsesCompleted}, "hello world", false},
		{"duplicate-part-done", []string{`{"type":"response.output_text.done","output_index":0,"text":"hello"}`, `{"type":"response.output_text.done","output_index":0,"text":"hello"}`, minimaxResponsesCompleted}, "", true},
		{"arguments-done-unassociated", []string{`{"type":"response.function_call_arguments.done","output_index":0,"arguments":"{}"}`, minimaxResponsesCompleted}, "", true},
		{"terminal-switches-finished-kind", []string{`{"type":"response.output_item.done","output_index":0,"item":{"type":"message","id":"same","content":[{"type":"output_text","text":"hello"}]}}`, `{"type":"response.completed","response":{"status":"completed","output":[{"type":"reasoning","id":"same","content":[{"type":"reasoning_text","text":"wrong kind"}]}]}}`}, "", true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "MM03-order-"+sc.name)
			w := newMiniMaxResponsesWorld(t, nil)
			ev.Fixture("synthetic-events.json", mustMarshal(t, sc.events))
			enqueue(ev, w, minimaxResponsesReply(t, sc.events...))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if sc.failure {
				checkFailed(ev, o.result, o.err)
			} else {
				ev.Check("associated parts complete exactly", o.err == nil && miniMaxResultText(o.result.Message) == sc.want, "%v %+v", o.err, o.result.Message)
			}
		})
	}
}

func miniMaxResultText(m ai.AssistantMessage) string {
	var text string
	for _, b := range m.Content {
		if b, ok := b.(ai.Text); ok {
			text += b.Text
		}
	}
	return text
}

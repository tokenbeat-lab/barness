package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxResponsesReasoningHistoryWithoutContent(t *testing.T) {
	for _, mode := range []string{"item-done", "text-done"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "MM03-reasoning-history-"+mode)
			w := newMiniMaxResponsesWorld(t, nil)
			events := []string{`{"type":"response.output_item.added","output_index":0,"item":{"type":"reasoning","id":"rs","content":[]}}`, `{"type":"response.reasoning_text.delta","output_index":0,"delta":"complete exact thinking"}`}
			if mode == "item-done" {
				events = append(events, `{"type":"response.output_item.done","output_index":0,"item":{"type":"reasoning","id":"rs","content":[]}}`)
			} else {
				events = append(events, `{"type":"response.reasoning_text.done","output_index":0,"text":"complete exact thinking"}`)
			}
			events = append(events, minimaxResponsesCompleted)
			ev.Fixture("synthetic-events.json", mustMarshal(t, events))
			enqueue(ev, w, minimaxResponsesReply(t, events...))
			enqueue(ev, w, minimaxResponsesReply(t, minimaxResponsesCompleted))
			target, req := minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest()
			first, err := w.client.Complete(ctxFor(t), textScope(t.Name()+"-first"), target, req, nil)
			ev.Record("first", first)
			if !ev.Check("first thinking succeeds", err == nil, "%v", err) {
				return
			}
			req.Messages = append(req.Messages, first.Message, ai.UserText("Continue."))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()+"-continue"), target, req, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			if !ev.Check("history preserves available thinking", o.err == nil && len(w.provider.Requests()) == 2, "%v", o.err) {
				return
			}
			var input []map[string]any
			mustUnmarshal(t, bodyFields(t, w.provider.Requests()[1].Body)["input"], &input)
			found := false
			for _, item := range input {
				if item["type"] == "reasoning" {
					found = jsonEqual(mustMarshal(t, item), json.RawMessage(`{"type":"reasoning","summary":[{"type":"summary_text","text":"complete exact thinking"}]}`))
				} else if item["role"] == "assistant" {
					found = jsonEqual(mustMarshal(t, item["content"]), json.RawMessage(`[{"type":"output_text","text":"complete exact thinking"}]`))
				}
			}
			ev.Check("no available reasoning is silently erased", found, "%v", input)
		})
	}
}

package e2e

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestXAIReasoningAuthority(t *testing.T) {
	for _, sc := range []struct {
		name, delta, other, added, done, final string
		invalid                                bool
	}{
		{name: "alternate-channel-authority", delta: "Check.", final: xaiReasoning},
		{name: "alternate-conflict", delta: "Wrong.", final: xaiReasoning, invalid: true},
		{name: "dual-channel-conflict", delta: "Wrong.", other: "Check.", final: xaiReasoning, invalid: true},
		{name: "explicit-channel-structure", delta: "Detail.", other: "Summary.", final: `{"type":"reasoning","id":"r","summary":[{"type":"summary_text","text":"Summary."}],"content":[{"type":"reasoning_text","text":"Detail."}],"encrypted_content":"cipher"}`},
		{name: "ciphertext-backfill", done: `{"id":"r","type":"reasoning","summary":[{"type":"summary_text","text":"Check."}]}`, final: `{"id":"r","type":"reasoning","summary":[{"type":"summary_text","text":"Check."}],"encrypted_content":"cipher","vendor_extra":{"x":1}}`},
		{name: "ciphertext-conflict-after-added", added: `{"id":"r","type":"reasoning","summary":[],"encrypted_content":"first"}`, final: `{"id":"r","type":"reasoning","summary":[],"encrypted_content":"changed"}`, invalid: true},
		{name: "ciphertext-conflict-after-done", done: `{"id":"r","type":"reasoning","summary":[],"encrypted_content":"first"}`, final: `{"id":"r","type":"reasoning","summary":[],"encrypted_content":"changed"}`, invalid: true},
		{name: "ciphertext-malformed", final: `{"id":"r","type":"reasoning","summary":[],"encrypted_content":{}}`, invalid: true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E05-authority-"+sc.name)
			w := newXAIWorld(t, nil)
			id := "r"
			if strings.Contains(sc.final, "reason|α") {
				id = "reason|α"
			}
			var frames []json.RawMessage
			added := sc.added
			if added == "" {
				added = `{"id":"` + id + `","type":"reasoning","summary":[]}`
			}
			if sc.delta != "" || sc.added != "" {
				frames = append(frames, json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":`+added+`}`))
			}
			if sc.delta != "" {
				frames = append(frames, mustMarshal(t, map[string]any{"type": "response.reasoning_text.delta", "output_index": 0, "item_id": id, "content_index": 0, "delta": sc.delta}))
			}
			if sc.other != "" {
				frames = append(frames, mustMarshal(t, map[string]any{"type": "response.reasoning_summary_text.delta", "output_index": 0, "item_id": id, "summary_index": 0, "delta": sc.other}))
			}
			if sc.done != "" {
				frames = append(frames, json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":`+sc.done+`}`))
			}
			frames = append(frames, xaiTerminal(sc.final, xaiUsage))
			enqueue(ev, w, xaiReply(t, frames...))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("frames", frames)
			if sc.invalid {
				ev.Check("contradictory native state fails", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v %+v", o.err, o.result.Message)
				return
			}
			ev.Check("native authority reconciles", o.err == nil && len(o.result.Message.Content) == 1, "%v %+v", o.err, o.result.Message)
			if len(o.result.Message.Content) == 1 {
				thinking, ok := o.result.Message.Content[0].(ai.Thinking)
				ev.Check("complete raw reasoning is final authority", ok && jsonEqual(json.RawMessage(thinking.Signature), json.RawMessage(sc.final)), "%+v", thinking)
			}
		})
	}
}

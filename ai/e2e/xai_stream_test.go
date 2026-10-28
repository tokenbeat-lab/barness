package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const xaiReasoning = `{"id":"reason|α","type":"reasoning","summary":[{"type":"summary_text","text":"Check."}],"encrypted_content":"synthetic-cipher","extra":{"preserve":true}}`

func xaiTerminal(output, usage string) json.RawMessage {
	r := `{"type":"response.completed","response":{"id":"xai-final","model":"grok-4.7","status":"completed","output":[` + output + `]`
	if usage != "" {
		r += `,"usage":` + usage
	}
	return json.RawMessage(r + `}}`)
}

func TestXAIStreamSnapshots(t *testing.T) {
	for _, entry := range []string{"stream", "complete"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "XAI-E05-final-only-"+entry)
			w := newXAIWorld(t, nil)
			frame := xaiTerminal(xaiReasoning+`,`+xaiMessage, xaiUsage)
			enqueue(ev, w, xaiReply(t, frame))
			var o outcome
			if entry == "stream" {
				o = drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			} else {
				o.result, o.err = w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
			}
			o.record(ev)
			ev.Record("frames", []json.RawMessage{frame})
			ev.Check("final-only produces all content", o.err == nil && len(o.result.Message.Content) == 2, "err=%v result=%+v", o.err, o.result)
			if len(o.result.Message.Content) == 2 {
				thinking, ok := o.result.Message.Content[0].(ai.Thinking)
				text, tok := o.result.Message.Content[1].(ai.Text)
				ev.Check("raw native reasoning survives with visible text", ok && thinking.Thinking == "Check." && jsonEqual(json.RawMessage(thinking.Signature), json.RawMessage(xaiReasoning)), "%+v", thinking)
				ev.Check("final text retained", tok && text.Text == "Hello Grok.", "%+v", text)
			}
			ev.Check("actual and requested model retained", o.result.Message.Model == "grok-4.7" && o.result.Message.ResponseModel == "grok-4.7", "%+v", o.result.Message)
		})
	}
}

func TestXAIStreamAuthoritativeSnapshots(t *testing.T) {
	for _, sc := range []struct {
		name           string
		frames         []json.RawMessage
		text, thinking string
	}{
		{"done-only", []json.RawMessage{json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + xaiMessage + `}`), xaiTerminal(xaiMessage, xaiUsage)}, "Hello Grok.", ""},
		{"duplicate-final", []json.RawMessage{xaiTerminal(xaiReasoning+`,`+xaiMessage, xaiUsage), xaiTerminal(xaiReasoning+`,`+xaiMessage, xaiUsage)}, "Hello Grok.", "Check."},
		{"reasoning-dual-channel", []json.RawMessage{
			json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"reason|α","type":"reasoning","summary":[]}}`),
			json.RawMessage(`{"type":"response.reasoning_text.delta","output_index":0,"item_id":"reason|α","content_index":0,"delta":"Check."}`),
			json.RawMessage(`{"type":"response.reasoning_summary_text.delta","output_index":0,"item_id":"reason|α","summary_index":0,"delta":"Check."}`),
			json.RawMessage(`{"type":"response.reasoning_text.done","output_index":0,"item_id":"reason|α","content_index":0,"text":"Check."}`),
			json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":` + xaiReasoning + `}`), xaiTerminal(xaiReasoning+`,`+xaiMessage, xaiUsage)}, "Hello Grok.", "Check."},
		{"ciphertext-only", []json.RawMessage{xaiTerminal(`{"id":"r","type":"reasoning","encrypted_content":"cipher","summary":[]}`, xaiUsage)}, "", ""},
		{"indexed-parts", []json.RawMessage{
			json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"m","type":"message","role":"assistant","content":[]}}`),
			json.RawMessage(`{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":1,"delta":"B"}`),
			json.RawMessage(`{"type":"response.output_text.done","output_index":0,"item_id":"m","content_index":0,"text":"A"}`),
			json.RawMessage(`{"type":"response.output_text.done","output_index":0,"item_id":"m","content_index":1,"text":"B"}`),
			xaiTerminal(`{"id":"m","type":"message","role":"assistant","content":[{"type":"output_text","text":"A"},{"type":"output_text","text":"B"}]}`, xaiUsage)}, "AB", ""},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E05-snapshots-"+sc.name)
			w := newXAIWorld(t, nil)
			enqueue(ev, w, xaiReply(t, sc.frames...))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("frames", sc.frames)
			ev.Check("authoritative snapshots reconcile", o.err == nil && textOf(o.result.Message) == sc.text, "err=%v result=%+v", o.err, o.result.Message)
			var got string
			starts, ends := 0, 0
			for _, b := range o.result.Message.Content {
				if v, ok := b.(ai.Thinking); ok {
					got += v.Thinking
				}
			}
			for _, e := range o.events {
				switch e.(type) {
				case ai.ThinkingStartEvent, ai.TextStartEvent:
					starts++
				case ai.ThinkingEndEvent, ai.TextEndEvent:
					ends++
				}
			}
			ev.Check("reasoning is not duplicated", got == sc.thinking, "%q", got)
			ev.Check("blocks complete once", starts == len(o.result.Message.Content) && ends == starts, "start=%d end=%d content=%d", starts, ends, len(o.result.Message.Content))
		})
	}
}

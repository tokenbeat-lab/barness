package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestXAIStreamStructure(t *testing.T) {
	added := `{"type":"response.output_item.added","output_index":0,"item":{"id":"m","type":"message","role":"assistant","content":[]}}`
	textDone := `{"type":"response.output_text.done","output_index":0,"item_id":"m","content_index":0,"text":"Same"}`
	message := `{"id":"m","type":"message","role":"assistant","content":[{"type":"output_text","text":"Same"}]}`
	itemDone := `{"type":"response.output_item.done","output_index":0,"item":` + message + `}`
	final := string(xaiTerminal(message, xaiUsage))
	for _, sc := range []struct {
		name    string
		frames  []string
		invalid bool
	}{
		{"repeat-completed-part", []string{added, textDone, itemDone, textDone, final}, false},
		{"changed-completed-part-kind", []string{added, textDone, strings.Replace(itemDone, `"output_text","text"`, `"refusal","refusal"`, 1), final}, true},
		{"delta-after-done", []string{added, textDone, `{"type":"response.output_text.delta","output_index":0,"item_id":"m","content_index":0,"delta":"Changed"}`}, true},
		{"null-item-status", []string{strings.Replace(itemDone, `"role":"assistant"`, `"role":"assistant","status":null`, 1), final}, true},
		{"unfinished-item-status", []string{strings.Replace(itemDone, `"role":"assistant"`, `"role":"assistant","status":"in_progress"`, 1), final}, true},
		{"reasoning-resource", []string{string(xaiTerminal(`{"type":"reasoning","id":"r","summary":[],"extra":{"file_id":"untrusted"}}`, xaiUsage))}, true},
		{"reasoning-large-unknown-number", []string{string(xaiTerminal(`{"type":"reasoning","id":"r","summary":[],"opaque":1e400}`, xaiUsage))}, false},
		{"reasoning-null-status", []string{string(xaiTerminal(`{"type":"reasoning","id":"r","status":null,"summary":[]}`, xaiUsage))}, true},
		{"bad-done-marker", []string{final, `[DONE]junk`}, true},
		{"known-lower-bound-overflow", []string{string(xaiTerminal(message, `{"input_tokens":9223372036854775807,"output_tokens":1}`))}, true},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E10-structure-"+sc.name)
			w := newXAIWorld(t, nil)
			var chunks [][]byte
			for _, f := range sc.frames {
				chunks = append(chunks, []byte("data: "+f+"\n\n"))
			}
			enqueue(ev, w, provider.SSE(chunks))
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("frames", sc.frames)
			if sc.invalid {
				ev.Check("invalid structure fails protocol", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v %+v", o.err, o.result.Message)
			} else {
				ev.Check("valid exact native structure retained", o.err == nil, "%v", o.err)
				if sc.name == "reasoning-large-unknown-number" && len(o.result.Message.Content) == 1 {
					b, ok := o.result.Message.Content[0].(ai.Thinking)
					ev.Check("unknown native number stays exact", ok && strings.Contains(b.Signature, "1e400"), "%+v", b)
				}
			}
		})
	}
}

func TestXAINamedStreamEvent(t *testing.T) {
	ev := run.Case(t, "XAI-E10-named-event-mismatch")
	w := newXAIWorld(t, nil)
	enqueue(ev, w, provider.SSE([][]byte{[]byte("event: response.completed\ndata: " + string(xaiTerminal(xaiMessage, xaiUsage)) + "\n\n"), []byte("event: response.created\ndata: " + string(xaiTerminal(xaiMessage, xaiUsage)) + "\n\n")}))
	res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Check("named event cannot contradict payload type", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", err)
}

func TestXAICompletedStatusShape(t *testing.T) {
	for _, sc := range []struct{ name, output, extra string }{
		{"message-role", `{"type":"message","id":"m","role":"user","content":[]}`, ""},
		{"message-content-missing", `{"type":"message","id":"m","role":"assistant"}`, ""},
		{"completed-with-incomplete-reason", xaiMessage, `,"incomplete_details":{"reason":"max_output_tokens"}`},
		{"completed-with-error", xaiMessage, `,"error":{"code":"server_error","message":"private"}`},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E10-status-shape-"+sc.name)
			w := newXAIWorld(t, nil)
			frame := []byte(`{"type":"response.completed","response":{"status":"completed","model":"grok-4.7","output":[` + sc.output + `]` + sc.extra + `}}`)
			enqueue(ev, w, provider.SSE([][]byte{append(append([]byte("data: "), frame...), []byte("\n\n")...)}))
			res, err := w.client.Complete(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("inconsistent completed shape fails", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", err)
		})
	}
}

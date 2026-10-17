package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMResponsesEmptyReasoning(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, shape := range []string{"empty-object", "empty-part-array", "no-parts", "absent", "absent-empty-done"} {
			t.Run(model+"/"+shape, func(t *testing.T) {
				ev := run.Case(t, "P0-GLM13-authoritative-reasoning-"+model+"-"+shape)
				w := newGLMResponsesWorld(t, nil)
				item := `{"type":"reasoning","id":"r1"}`
				switch shape {
				case "empty-object":
					item = `{"type":"reasoning","id":"r1","content":{"type":"reasoning_text","text":""}}`
				case "empty-part-array":
					item = `{"type":"reasoning","id":"r1","content":[{"type":"reasoning_text","text":""}]}`
				case "no-parts":
					item = `{"type":"reasoning","id":"r1","content":[]}`
				}
				events := []json.RawMessage{json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"reasoning","id":"r1","content":[]}}`), json.RawMessage(`{"type":"response.reasoning_text.delta","output_index":0,"delta":"stale draft"}`)}
				if shape == "absent-empty-done" {
					events = append(events, json.RawMessage(`{"type":"response.reasoning_text.done","output_index":0,"text":""}`))
				}
				events = append(events, json.RawMessage(fmt.Sprintf(`{"type":"response.output_item.done","output_index":0,"item":%s}`, item)), json.RawMessage(fmt.Sprintf(`{"type":"response.completed","response":{"id":"resp1","status":"completed","output":[%s]}}`, item)))
				ev.Fixture("wire.json", mustMarshal(t, events))
				enqueue(ev, w, glmResponsesReply(t, events))
				o := drain(w.client.Stream(ctxFor(t), textScope("glm13-empty-reasoning"), glmResponsesTarget(model), glmResponsesRequest(), nil))
				o.record(ev)
				want := ""
				if shape == "absent" || shape == "no-parts" {
					want = "stale draft"
				}
				ev.Check("valid final content succeeds", o.err == nil, "err=%v", o.err)
				found := false
				for _, b := range o.result.Message.Content {
					if thought, ok := b.(ai.Thinking); ok {
						found = true
						ev.Check("presence governs authoritative thinking, including empty", thought.Thinking == want, "got %q want %q", thought.Thinking, want)
						ev.Check("native original shape unchanged", jsonEqual([]byte(thought.Signature), []byte(item)), "signature=%s", thought.Signature)
					}
				}
				ev.Check("one reasoning block retained", found, "reasoning missing")
			})
		}
	}
}

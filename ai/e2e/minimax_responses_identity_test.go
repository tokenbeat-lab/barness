package e2e

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMiniMaxResponsesFunctionIdentityFrozen(t *testing.T) {
	for _, phase := range []string{"done", "terminal"} {
		for _, field := range []string{"call_id", "name"} {
			t.Run(phase+"/"+field, func(t *testing.T) {
				ev := run.Case(t, "MM10-function-identity-"+phase+"-"+field)
				w := newMiniMaxResponsesWorld(t, nil)
				item := map[string]any{"type": "function_call", "id": "same-item", "call_id": "original-call", "name": "original-function", "arguments": ""}
				raw, _ := json.Marshal(item)
				events := []string{fmt.Sprintf(`{"type":"response.output_item.added","output_index":0,"item":%s}`, raw)}
				item[field] = "changed-identity"
				item["arguments"] = `{"value":1}`
				raw, _ = json.Marshal(item)
				if phase == "done" {
					events = append(events, fmt.Sprintf(`{"type":"response.output_item.done","output_index":0,"item":%s}`, raw), minimaxResponsesCompleted)
				} else {
					events = append(events, fmt.Sprintf(`{"type":"response.completed","response":{"status":"completed","output":[%s]}}`, raw))
				}
				ev.Fixture("synthetic-events.json", mustMarshal(t, events))
				enqueue(ev, w, minimaxResponsesReply(t, events...))
				o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), minimaxResponsesTarget("MiniMax-M3"), minimaxResponsesRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkFailed(ev, o.result, o.err)
				ev.Check("one original call identity retained", len(o.result.Metadata.Attempts) == 1 && len(w.provider.Requests()) == 1, "%+v", o.result)
			})
		}
	}
}

package e2e

import (
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestKimiResponsesCompletedTextConsistency(t *testing.T) {
	for _, market := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, kind := range []string{"message", "reasoning"} {
			for _, boundary := range []string{"item", "terminal", "part"} {
				for _, mode := range []string{"consistent", "contradictory", "missing-part"} {
					t.Run(string(market)+"/"+kind+"/"+boundary+"/"+mode, func(t *testing.T) {
						ev := run.Case(t, kimiResponsesID(market, "K12-completed-text-")+kind+"-"+boundary+"-"+mode)
						w := newKimiResponsesWorld(t, market, nil)
						field, event, part, index := "content", "output_text", "output_text", "content_index"
						if kind == "reasoning" {
							field, event, part, index = "summary", "reasoning_summary_text", "summary_text", "summary_index"
						}
						item := func(text string) string {
							return `{"type":"` + kind + `","id":"completed","` + field + `":[{"type":"` + part + `","text":"` + text + `"}]}`
						}
						finished := item("A")
						if mode == "contradictory" {
							finished = item("B")
						}
						if mode == "missing-part" {
							finished = `{"type":"` + kind + `","id":"completed","` + field + `":[]}`
						}
						frames := []json.RawMessage{json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"type":"` + kind + `","id":"completed","` + field + `":[]}}`), json.RawMessage(`{"type":"response.` + event + `.done","output_index":0,"` + index + `":0,"text":"A"}`)}
						if boundary == "item" {
							frames = append(frames, json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":`+finished+`}`))
						}
						if boundary == "part" {
							partEvent := "content_part"
							if kind == "reasoning" {
								partEvent = "reasoning_summary_part"
							}
							text := "A"
							if mode != "consistent" {
								text = "B"
							}
							frames = append(frames, json.RawMessage(`{"type":"response.`+partEvent+`.done","output_index":0,"`+index+`":0,"part":{"type":"`+part+`","text":"`+text+`"}}`))
						}
						frames = append(frames, json.RawMessage(`{"type":"response.completed","response":{"status":"completed","output":[`+finished+`]}}`))
						ev.Fixture("synthetic.json", mustMarshal(t, frames))
						enqueue(ev, w, kimiResponsesReply(t, frames...))
						o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), kimiResponsesTarget(), kimiResponsesRequest(), nil))
						o.record(ev)
						if mode == "consistent" {
							ev.Check("consistent completed part succeeds", o.err == nil, "%v", o.err)
						} else {
							ev.Check("completed text cannot be replaced or dropped", errors.Is(o.err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseStream}), "%v", o.err)
						}
					})
				}
			}
		}
	}
}

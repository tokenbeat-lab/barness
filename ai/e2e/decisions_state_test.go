package e2e

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsStateAndOrder(t *testing.T) {
	for _, sc := range []struct {
		name        string
		state, text string
	}{
		{"string", `"raw <>& \"text\""`, `raw <>& "text"`},
		{"empty-string", `""`, ""},
		{"empty-object", ` { } `, `{}`},
		{"empty-array", ` [ ] `, `[]`},
		{"object", ` { "large": 9007199254740993, "decimal": 1.2300e+90 } `, `{"large":9007199254740993,"decimal":1.2300e+90}`},
		{"array", ` [12345678901234567890, { "word": "<>&" }] `, `[12345678901234567890,{"word":"<>&"}]`},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD11-state-order-"+sc.name+"-plain")
			u := newDecisionsWorld(t, nil)
			req := ai.ClassifierRequest{State: json.RawMessage(sc.state), Questions: map[string]ai.ClassifierQuestion{}}
			criteria := map[string]json.RawMessage{"true": json.RawMessage(`null`), "false": json.RawMessage(`""`), "10": json.RawMessage(`"Ten"`), "2": json.RawMessage(`"Two"`), "01": json.RawMessage(`null`)}
			for _, name := range []string{"a", "10", "01", "2"} {
				req.Questions[name] = ai.ChoiceQuestion{Instructions: json.RawMessage(`"Question"`), Criteria: criteria}
			}
			// Response intentionally reverses question and option order.
			answers := []any{}
			for _, name := range []string{"a", "01", "10", "2"} {
				answers = append(answers, map[string]any{"type": "choice", "name": name, "choice": "true", "confidence": .6, "probabilities": []any{map[string]any{"value": "true", "probability": 1}, map[string]any{"value": "false", "probability": 0}, map[string]any{"value": "01", "probability": 0}, map[string]any{"value": "10", "probability": 0}, map[string]any{"value": "2", "probability": 0}}})
			}
			reply := mustMarshal(t, map[string]any{"answers": answers})
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: string(reply)}.script(t, decisionsProtocol, ""))
			res, err := u.client.Classify(ctxFor(t), textScope("req-state-"+sc.name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			ev.Record("input", req)
			ev.Record("result", res)
			ev.Check("successful complete name matching", err == nil && len(res.Answers) == 4, "got %v", err)
			for name := range req.Questions {
				a, ok := res.Answers[name].(ai.ChoiceAnswer)
				ev.Check("string choice "+name, ok && a.Choice == "true" && a.Confidence == .6, "got %+v", res.Answers[name])
			}
			requests := u.provider.Requests()
			if ev.Check("one request", len(requests) == 1, "got %d", len(requests)) {
				var body struct {
					Input     string
					Questions []struct {
						Name    string
						Choices []map[string]any
					}
				}
				mustUnmarshal(t, requests[0].Body, &body)
				ev.Check("State exact original text or compact numbers", body.Input == sc.text, "got %q", body.Input)
				names := []string{}
				for _, q := range body.Questions {
					names = append(names, q.Name)
					values := []any{}
					for _, c := range q.Choices {
						values = append(values, c["value"])
						if c["value"] == "true" || c["value"] == "false" {
							_, present := c["description"]
							ev.Check("null and empty descriptions omitted", !present, "got %+v", c)
						}
					}
					ev.Check("canonical option order", reflect.DeepEqual(values, []any{"2", "10", "01", "false", "true"}), "got %+v", values)
				}
				ev.Check("canonical question order", reflect.DeepEqual(names, []string{"2", "10", "01", "a"}), "got %v", names)
			}
			ev.Check("missing usage remains unreported", len(res.Metadata.Attempts) == 1 && res.Metadata.Attempts[0].UsageReporting == ai.UsageUnreported && res.Usage.Input == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

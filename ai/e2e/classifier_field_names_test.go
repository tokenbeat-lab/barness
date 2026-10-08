package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestClassifierExactAnswerFields(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "mixed.json")
	sc := f.Scenarios[0]
	for name, pair := range map[string][2]string{
		"bool-case-collision":       {`"noul":0.95`, `"noul":-1,"Noul":0.95`},
		"bool-case-alias":           {`"noul":0.95`, `"Noul":0.95`},
		"score-case-collision":      {`"score":1.05`, `"score":-1,"Score":1.05`},
		"confidence-case-collision": {`"confidence":0.7`, `"confidence":-1,"Confidence":0.7`},
		"choice-case-collision":     {`"choice":"track"`, `"choice":"invalid","Choice":"track"`},
		"type-case-collision":       {`"type":"noul"`, `"type":"choice","Type":"noul"`},
		"usage-case-alias":          {`"input_tokens":296`, `"Input_tokens":296`},
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E02-exact-fields-"+name)
			w := scenarioWorld(t, sc)
			body := strings.Replace(sc.Replies[0].Body, pair[0], pair[1], 1)
			enqueue(ev, w, fixtureReply{Status: 200, Body: body}.script(t, classifierProtocol, ""))
			res, err := w.client.Classify(ctxFor(t), textScope("req-fields"), sc.target(), classifierInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "usage-case-alias" { // token field names are map keys; aliases are never counts.
				ev.Check("usage alias never treated as canonical count", res.Usage.Input == 0, "got %+v", res.Usage)
			} else {
				ev.Check("case aliases rejected atomically", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0 && res.Usage.Input == 296, "got %v / %+v", err, res)
			}
		})
	}
}

func TestClassifierExactRequestFields(t *testing.T) {
	for name, question := range map[string]string{
		"bool-null-criteria":       `{"type":"bool","instructions":"Ask","criteria":null}`,
		"bool-case-criteria":       `{"type":"bool","instructions":"Ask","criteria":{"True":"yes","false":"no"}}`,
		"score-case-type":          `{"Type":"score","instructions":"Ask","criteria":["a","b"]}`,
		"choice-case-instructions": `{"type":"choice","Instructions":"Ask","criteria":{"a":"a"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E07-exact-request-"+name)
			var req ai.ClassifierRequest
			err := json.Unmarshal([]byte(`{"state":{},"questions":{"q":`+question+`}}`), &req)
			ev.Record("error", errString(err))
			ev.Check("explicit invalid record fields refused", err != nil, "got %+v", req)
		})
	}
	for _, name := range []string{"null-criteria", "case-criteria", "case-score"} {
		t.Run("callback-"+name, func(t *testing.T) {
			ev := run.Case(t, "P07-E04-exact-request-"+name)
			f, _ := loadFixture(t, classifierProtocol, "mixed.json")
			sc := f.Scenarios[0]
			w := scenarioWorld(t, sc)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				questions := p.Body["questions"].(map[string]any)
				if name == "case-score" {
					q := questions["anger"].(map[string]any)
					q["Instructions"] = q["instructions"]
					delete(q, "instructions")
				} else {
					q := questions["urgent"].(map[string]any)
					if name == "null-criteria" {
						q["criteria"] = nil
					} else {
						q["criteria"] = map[string]any{"True": "yes", "false": "no"}
					}
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-exact-callback"), sc.target(), classifierInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("invalid final record blocked before HTTP", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0, "got %v", err)
		})
	}
}

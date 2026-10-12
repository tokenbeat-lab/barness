package e2e

import (
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsAnswerValidation(t *testing.T) {
	for _, name := range []string{"valid", "tie", "sum-within", "sum-outside", "name-missing", "name-null", "name-empty", "name-foreign", "name-duplicate-field", "name-case", "extra-answer", "duplicate-answer", "missing-answer", "answers-null", "type-missing", "type-unknown", "type-score", "type-bool", "type-null", "choice-missing", "choice-null", "choice-boolean", "choice-foreign", "choice-case", "choice-lowest", "probabilities-missing", "probabilities-null", "probabilities-object", "probabilities-short", "probabilities-extra", "probabilities-duplicate", "value-null", "value-boolean", "value-case", "probability-null", "probability-string", "probability-overflow", "probability-negative", "probability-above-one", "probability-case", "probability-duplicate-field", "confidence-missing", "confidence-null", "confidence-string", "confidence-overflow", "confidence-negative", "confidence-above-one", "confidence-case", "answer-unknown-field", "entry-unknown-field", "envelope-answers-case", "envelope-usage-case", "envelope-model-case", "envelope-duplicate-answers", "envelope-duplicate-model"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD31-answer-"+name+"-hooked")
			u := newDecisionsWorld(t, nil)
			answer := decisionsAnswer
			valid := false
			switch name {
			case "valid":
				valid = true
			case "tie":
				answer = strings.ReplaceAll(strings.ReplaceAll(answer, `0.9`, `0.5`), `0.1`, `0.5`)
				valid = true
			case "sum-within":
				answer = strings.ReplaceAll(strings.ReplaceAll(answer, `0.9`, `0.6`), `0.1`, `0.4000005`)
				valid = true
			case "sum-outside":
				answer = strings.ReplaceAll(strings.ReplaceAll(answer, `0.9`, `0.6`), `0.1`, `0.400002`)
			case "name-missing":
				answer = strings.Replace(answer, `,"name":"intent"`, "", 1)
			case "name-null":
				answer = strings.Replace(answer, `"name":"intent"`, `"name":null`, 1)
			case "name-empty":
				answer = strings.Replace(answer, `"name":"intent"`, `"name":""`, 1)
			case "name-foreign":
				answer = strings.Replace(answer, `"name":"intent"`, `"name":"foreign"`, 1)
			case "name-duplicate-field":
				answer = strings.Replace(answer, `"name":"intent"`, `"name":"intent","name":"intent"`, 1)
			case "name-case":
				answer = strings.Replace(answer, `"name"`, `"Name"`, 1)
			case "type-missing":
				answer = strings.Replace(answer, `"type":"choice",`, "", 1)
			case "type-unknown":
				answer = strings.Replace(answer, `"type":"choice"`, `"type":"novel"`, 1)
			case "type-score":
				answer = strings.Replace(answer, `"type":"choice"`, `"type":"score"`, 1)
			case "type-bool":
				answer = strings.Replace(answer, `"type":"choice"`, `"type":"predicate"`, 1)
			case "type-null":
				answer = strings.Replace(answer, `"type":"choice"`, `"type":null`, 1)
			case "choice-missing":
				answer = strings.Replace(answer, `,"choice":"track"`, "", 1)
			case "choice-null":
				answer = strings.Replace(answer, `"choice":"track"`, `"choice":null`, 1)
			case "choice-boolean":
				answer = strings.Replace(answer, `"choice":"track"`, `"choice":true`, 1)
			case "choice-foreign":
				answer = strings.Replace(answer, `"choice":"track"`, `"choice":"foreign"`, 1)
			case "choice-case":
				answer = strings.Replace(answer, `"choice":"track"`, `"Choice":"track"`, 1)
			case "choice-lowest":
				answer = strings.Replace(answer, `"choice":"track"`, `"choice":"other"`, 1)
			case "probabilities-missing":
				answer = `{"type":"choice","name":"intent","choice":"track","confidence":0.7}`
			case "probabilities-null":
				answer = `{"type":"choice","name":"intent","choice":"track","probabilities":null,"confidence":0.7}`
			case "probabilities-object":
				answer = `{"type":"choice","name":"intent","choice":"track","probabilities":{"track":0.9,"other":0.1},"confidence":0.7}`
			case "probabilities-short":
				answer = strings.Replace(answer, `,{"value":"other","probability":0.1}`, "", 1)
			case "probabilities-extra":
				answer = strings.Replace(answer, `],"confidence"`, `,{"value":"third","probability":0}],"confidence"`, 1)
			case "probabilities-duplicate":
				answer = strings.Replace(answer, `"value":"other"`, `"value":"track"`, 1)
			case "value-null":
				answer = strings.Replace(answer, `"value":"track"`, `"value":null`, 1)
			case "value-boolean":
				answer = strings.Replace(answer, `"value":"track"`, `"value":true`, 1)
			case "value-case":
				answer = strings.Replace(answer, `"value":"track"`, `"Value":"track"`, 1)
			case "probability-null":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":null`, 1)
			case "probability-string":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":"0.9"`, 1)
			case "probability-overflow":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":1e999`, 1)
			case "probability-negative":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":-0.9`, 1)
			case "probability-above-one":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":1.1`, 1)
			case "probability-case":
				answer = strings.Replace(answer, `"probability":0.9`, `"Probability":0.9`, 1)
			case "probability-duplicate-field":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":0.9,"probability":0.9`, 1)
			case "confidence-missing":
				answer = strings.Replace(answer, `,"confidence":0.7`, "", 1)
			case "confidence-null":
				answer = strings.Replace(answer, `"confidence":0.7`, `"confidence":null`, 1)
			case "confidence-string":
				answer = strings.Replace(answer, `"confidence":0.7`, `"confidence":"0.7"`, 1)
			case "confidence-overflow":
				answer = strings.Replace(answer, `"confidence":0.7`, `"confidence":1e999`, 1)
			case "confidence-negative":
				answer = strings.Replace(answer, `"confidence":0.7`, `"confidence":-1`, 1)
			case "confidence-above-one":
				answer = strings.Replace(answer, `"confidence":0.7`, `"confidence":1.1`, 1)
			case "confidence-case":
				answer = strings.Replace(answer, `"confidence":0.7`, `"Confidence":0.7`, 1)
			case "answer-unknown-field":
				answer = strings.TrimSuffix(answer, `}`) + `,"probability":0.9}`
			case "entry-unknown-field":
				answer = strings.Replace(answer, `"probability":0.9`, `"probability":0.9,"label":"foreign"`, 1)
			}
			answers := answer
			switch name {
			case "extra-answer":
				answers += `,{"type":"refusal","name":"extra"}`
			case "duplicate-answer":
				answers += `,` + answer
			case "missing-answer":
				answers = ""
			}
			body := `{"model":"response-version","answers":[` + answers + `],` + decisionsUsage + `}`
			switch name {
			case "answers-null":
				body = `{"answers":null,` + decisionsUsage + `}`
			case "envelope-answers-case":
				body = strings.TrimSuffix(body, `}`) + `,"Answers":[]}`
			case "envelope-usage-case":
				body = strings.TrimSuffix(body, `}`) + `,"Usage":{}}`
			case "envelope-model-case":
				body = strings.TrimSuffix(body, `}`) + `,"Model":"alias"}`
			case "envelope-duplicate-answers":
				body = strings.TrimSuffix(body, `}`) + `,"answers":[]}`
			case "envelope-duplicate-model":
				body = strings.TrimSuffix(body, `}`) + `,"model":"ambiguous"}`
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: body}.script(t, decisionsProtocol, ""))
			res, err := u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-answer-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			if valid {
				ev.Record("result", res)
				ev.Check("strict legal distribution", err == nil && len(res.Answers) == 1, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
			}
			ev.Check("unique basic usage survives invalid answers", res.Usage.Input == 296 && res.Usage.Output == 20 && res.Usage.Cost.Total == .000296, "got %+v", res.Usage)
			ev.Check("2xx never replayed", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsAnswersAtomic(t *testing.T) {
	ev := run.Case(t, "P10-OD30-atomic-name-set-hooked")
	u := newDecisionsWorld(t, nil)
	req := decisionsInput()
	req.Questions["broken"] = req.Questions["intent"]
	broken := strings.ReplaceAll(strings.Replace(decisionsAnswer, `"name":"intent"`, `"name":"broken"`, 1), `"choice":"track"`, `"choice":"foreign"`)
	enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + decisionsAnswer + `,` + broken + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
	res, err := u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-answer-atomic"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
	ev.Check("valid part never published", res.Answers == nil, "got %+v", res.Answers)
	u.records(ev, res, err)
	u.released(ev)
}

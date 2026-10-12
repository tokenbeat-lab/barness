package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type decisionsAnswerCase struct {
	name, answer string
	valid        bool
}

func decisionsScoreCases() []decisionsAnswerCase {
	cases := []decisionsAnswerCase{{"valid", decisionsScore, true}}
	for _, sc := range []struct {
		name, from, to string
		valid          bool
	}{
		{"score-within", `"score":1.1`, `"score":1.1000005`, true},
		{"score-outside", `"score":1.1`, `"score":1.100002`, false},
		{"sum-within", `"probability":0.2`, `"probability":0.2000005`, true},
		{"sum-outside", `"probability":0.2`, `"probability":0.200002`, false},
		{"high-precision-inconsistent", `"score":1.1`, `"score":1.10001`, false},
		{"missing-level", `,{"value":1,"label":"低","probability":0.5}`, ``, false},
		{"extra-level", `"probability":0.5}]`, `"probability":0.5},{"value":3,"label":"extra","probability":0}]`, false},
		{"duplicate-index", `"value":1`, `"value":0`, false},
		{"wrong-label", `"label":"高"`, `"label":"wrong"`, false},
		{"label-null", `"label":"高"`, `"label":null`, false},
		{"label-numeric", `"label":"高"`, `"label":2`, false},
		{"label-missing", `"label":"高",`, ``, false},
		{"label-alias", `"label":"高"`, `"Label":"高"`, false},
		{"label-duplicate", `"label":"高"`, `"label":"高","label":"高"`, false},
		{"index-fractional", `"value":2`, `"value":1.5`, false},
		{"index-string", `"value":2`, `"value":"2"`, false},
		{"index-negative", `"value":2`, `"value":-1`, false},
		{"index-extra", `"value":2`, `"value":3`, false},
		{"index-overflow", `"value":2`, `"value":1e999`, false},
		{"index-null", `"value":2`, `"value":null`, false},
		{"index-missing", `"value":2,`, ``, false},
		{"index-alias", `"value":2`, `"Value":2`, false},
		{"index-duplicate-field", `"value":2`, `"value":2,"value":2`, false},
		{"wrong-kind", `"type":"score"`, `"type":"predicate"`, false},
		{"native-bool", `"type":"score"`, `"type":"bool"`, false},
		{"null-type", `"type":"score"`, `"type":null`, false},
		{"unknown-type", `"type":"score"`, `"type":"unknown"`, false},
		{"name-missing", `,"name":"urgency"`, ``, false},
		{"name-null", `"name":"urgency"`, `"name":null`, false},
		{"name-empty", `"name":"urgency"`, `"name":""`, false},
		{"name-foreign", `"name":"urgency"`, `"name":"foreign"`, false},
		{"name-duplicate", `"name":"urgency"`, `"name":"urgency","name":"urgency"`, false},
		{"name-alias", `"name":"urgency"`, `"Name":"urgency"`, false},
		{"type-alias", `"type":"score"`, `"Type":"score"`, false},
		{"probabilities-null", `"probabilities":[{"value":2,"label":"高","probability":0.3},{"value":0,"label":"低","probability":0.2},{"value":1,"label":"低","probability":0.5}]`, `"probabilities":null`, false},
		{"probabilities-object", `"probabilities":[{"value":2,"label":"高","probability":0.3},{"value":0,"label":"低","probability":0.2},{"value":1,"label":"低","probability":0.5}]`, `"probabilities":{"0":0.2,"1":0.5,"2":0.3}`, false},
	} {
		cases = append(cases, decisionsAnswerCase{sc.name, strings.Replace(decisionsScore, sc.from, sc.to, 1), sc.valid})
	}
	for _, field := range []struct{ name, value string }{{"score", "1.1"}, {"confidence", "0.6"}, {"probability", "0.3"}} {
		for _, value := range []string{"null", `"0.5"`, "true", "-0.1", "1e999"} {
			cases = append(cases, decisionsAnswerCase{field.name + "-" + fmt.Sprint(len(cases)), strings.Replace(decisionsScore, `"`+field.name+`":`+field.value, `"`+field.name+`":`+value, 1), false})
		}
		above := "1.1"
		if field.name == "score" {
			above = "2.1"
		}
		cases = append(cases, decisionsAnswerCase{field.name + "-above", strings.Replace(decisionsScore, `"`+field.name+`":`+field.value, `"`+field.name+`":`+above, 1), false})
		cases = append(cases, decisionsAnswerCase{field.name + "-missing", strings.Replace(decisionsScore, `,"`+field.name+`":`+field.value, ``, 1), false})
		cases = append(cases, decisionsAnswerCase{field.name + "-alias", strings.Replace(decisionsScore, `"`+field.name+`":`, `"`+strings.ToUpper(field.name)+`":`, 1), false})
		cases = append(cases, decisionsAnswerCase{field.name + "-duplicate-field", strings.Replace(decisionsScore, `"`+field.name+`":`+field.value, `"`+field.name+`":`+field.value+`,"`+field.name+`":`+field.value, 1), false})
	}
	return cases
}

func TestDecisionsMixedAnswerValidation(t *testing.T) {
	cases := decisionsScoreCases()
	for _, value := range []string{"null", `"0.5"`, "true", "-0.1", "1.1", "1e999", "NaN", "0", "1"} {
		cases = append(cases, decisionsAnswerCase{"predicate-value-" + fmt.Sprint(len(cases)), strings.Replace(decisionsPredicate, `0.37`, value, 1), value == "0" || value == "1"})
	}
	for _, sc := range []struct{ name, from, to string }{
		{"predicate-missing", `,"probability":0.37`, ``},
		{"predicate-alias", `"probability"`, `"Probability"`},
		{"predicate-duplicate", `"probability":0.37`, `"probability":0.37,"probability":0.37`},
		{"predicate-wrong-kind", `"type":"predicate"`, `"type":"choice"`},
	} {
		cases = append(cases, decisionsAnswerCase{sc.name, strings.Replace(decisionsPredicate, sc.from, sc.to, 1), false})
	}
	for _, kind := range []string{"score", "predicate", "refusal", "choice"} {
		answer := decisionsScore
		if kind == "predicate" {
			answer = decisionsPredicate
		}
		if kind == "choice" {
			answer = decisionsAnswer
		}
		if kind == "refusal" {
			answer = `{"type":"refusal","name":"urgency"}`
		}
		for _, field := range []string{`"reason":"fabricated"`, `"noul":0.5`, `"answer":true`} {
			cases = append(cases, decisionsAnswerCase{kind + "-semantic-" + fmt.Sprint(len(cases)), strings.TrimSuffix(answer, `}`) + `,` + field + `}`, false})
		}
		for _, field := range []string{`"probability":0.5`, `"choice":"track"`, `"score":1`, `"confidence":0.5`, `"probabilities":[]`, `"label":"low"`, `"value":0`} {
			// Existing kind fields already have dedicated duplicate-field cases.
			key := strings.Split(field, ":")[0]
			if strings.Contains(answer, key+":") {
				continue
			}
			cases = append(cases, decisionsAnswerCase{kind + "-contradictory-" + fmt.Sprint(len(cases)), strings.TrimSuffix(answer, `}`) + `,` + field + `}`, false})
		}
	}
	for _, sc := range cases {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD33-mixed-validation-"+sc.name+"-"+entry)
				u := newDecisionsMixedWorld(t, nil)
				score, predicate, choice := decisionsScore, decisionsPredicate, decisionsAnswer
				if strings.HasPrefix(sc.name, "predicate") {
					predicate = sc.answer
				} else if strings.HasPrefix(sc.name, "choice") {
					choice = sc.answer
				} else {
					score = sc.answer
				}
				// The normal choice precedes damage: no partial publication.
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + choice + `,` + score + `,` + predicate + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
				res, err := decisionsPublicCall(t, u, entry, "req-mixed-validation-"+sc.name+"-"+entry, decisionsMixedInput(), decisionsNoopHooks())
				if sc.valid {
					ev.Record("result", res)
					ev.Record("error", errString(err))
					ev.Check("legal values retained", err == nil && len(res.Answers) == 3, "got %+v / %v", res.Answers, err)
					if a, ok := res.Answers["urgency"].(ai.ScoreAnswer); ok && sc.name == "score-within" {
						ev.Check("score not recomputed", a.Score == 1.1000005, "got %+v", a)
					}
					if a, ok := res.Answers["urgency"].(ai.ScoreAnswer); ok && sc.name == "sum-within" {
						ev.Check("probabilities not normalized", a.Probabilities[0] == .2000005, "got %+v", a)
					}
				} else {
					checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
					ev.Check("valid part never published", res.Answers == nil, "got %+v", res.Answers)
				}
				// Lexically invalid JSON cannot make its trailing usage authoritative.
				if !strings.Contains(sc.answer, "NaN") {
					ev.Check("unique usage survives", res.Usage.Input == 296 && res.Usage.Cost.Total == .000296, "got %+v", res.Usage)
				}
				ev.Check("2xx not replayed", len(u.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsMixedNames(t *testing.T) {
	for _, name := range []string{"missing", "extra", "duplicate", "wrong-kind", "missing-probabilities", "metadata-duplicate", "metadata-alias"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD30-mixed-names-"+name+"-"+entry)
				u := newDecisionsMixedWorld(t, nil)
				answers := decisionsAnswer + `,` + decisionsScore + `,` + decisionsPredicate
				switch name {
				case "missing":
					answers = decisionsAnswer + `,` + decisionsScore
				case "extra":
					answers += `,{"type":"refusal","name":"extra"}`
				case "duplicate":
					answers = decisionsAnswer + `,` + decisionsScore + `,` + decisionsScore
				case "wrong-kind":
					answers = decisionsAnswer + `,` + decisionsScore + `,` + strings.Replace(decisionsAnswer, `"intent"`, `"approved"`, 1)
				case "missing-probabilities":
					answers = decisionsAnswer + `,{"type":"score","name":"urgency","score":1.1,"confidence":0.6},` + decisionsPredicate
				case "metadata-duplicate":
					answers = strings.Replace(answers, `"score":1.1`, `"vendor":1,"vendor":1,"score":1.1`, 1)
				case "metadata-alias":
					answers = strings.Replace(answers, `"score":1.1`, `"Score":1.1,"score":1.1`, 1)
				}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + answers + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
				res, err := decisionsPublicCall(t, u, entry, "req-mixed-names-"+name+"-"+entry, decisionsMixedInput(), decisionsNoopHooks())
				checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
				ev.Check("unique usage retained", res.Usage.Input == 296, "got %+v", res.Usage)
				ev.Check("one attempt", len(res.Metadata.Attempts) == 1, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsRejectsJevRounding(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "P10-OD34-jev-rounding-"+entry)
			u := newDecisionsMixedWorld(t, nil)
			req := decisionsMixedInput()
			delete(req.Questions, "intent")
			delete(req.Questions, "approved")
			score := `{"type":"score","name":"urgency","score":1.32,"confidence":0.7,"probabilities":[{"value":0,"label":"低","probability":0.01},{"value":1,"label":"低","probability":0.67},{"value":2,"label":"高","probability":0.32}]}`
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + score + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
			res, err := decisionsPublicCall(t, u, entry, "req-rounding-"+entry, req, decisionsNoopHooks())
			// Same reported numeric values as Jev's fixed rounded-1.32 replay:
			// exact expectation 1.31; Decisions has no rounding exception.
			checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
			ev.Check("known usage kept", res.Usage.Input == 296, "got %+v", res.Usage)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

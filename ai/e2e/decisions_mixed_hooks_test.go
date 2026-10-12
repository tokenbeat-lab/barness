package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const decisionsChangedQuestions = `[{"type":"score","name":"from-choice","instructions":"Changed score","levels":[{"label":"高","description":"first"},{"label":"低"},{"label":"低","description":""}]},{"type":"choice","name":"from-bool","instructions":"Changed choice","choices":[{"value":"track"},{"value":"other"}]},{"type":"predicate","name":"from-score","instructions":"Changed predicate"}]`
const decisionsChangedAnswers = `{"type":"predicate","name":"from-score","probability":0.23},{"type":"choice","name":"from-bool","choice":"other","confidence":0.8,"probabilities":[{"value":"other","probability":0.5},{"value":"track","probability":0.5}]},{"type":"score","name":"from-choice","score":0.75,"confidence":0.6,"probabilities":[{"value":0,"label":"高","probability":0.5},{"value":1,"label":"低","probability":0.25},{"value":2,"label":"低","probability":0.25}]}`

func TestDecisionsMixedFinalQuestions(t *testing.T) {
	for _, name := range []string{"keep", "replace", "old-names", "old-labels", "all-refusal"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD22-mixed-final-"+name+"-hooked")
			u := newDecisionsMixedWorld(t, nil)
			answers := decisionsChangedAnswers
			if name == "old-names" {
				answers = decisionsAnswer + `,` + decisionsScore + `,` + decisionsPredicate
			}
			if name == "old-labels" {
				answers = strings.Replace(answers, `"label":"高"`, `"label":"低"`, 1)
			}
			if name == "all-refusal" {
				answers = `{"type":"refusal","name":"from-score"},{"type":"refusal","name":"from-choice"},{"type":"refusal","name":"from-bool"}`
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + answers + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
			req := decisionsMixedInput()
			original := mustMarshal(t, req)
			var borrowed map[string]any
			calls := 0
			hooks := ai.Hooks{
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					calls++
					var questions []any
					mustUnmarshal(t, []byte(decisionsChangedQuestions), &questions)
					p.Body["questions"] = questions
					borrowed = p.Body
					if name == "replace" {
						return ai.ReplacePayload(borrowed), nil
					}
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					// After send, mutate both names and deep level/choice references.
					questions := borrowed["questions"].([]any)
					q := questions[0].(map[string]any)
					q["name"] = "late"
					q["levels"].([]any)[0].(map[string]any)["label"] = "late"
					questions[1].(map[string]any)["choices"].([]any)[0].(map[string]any)["value"] = "late"
					return nil
				},
			}
			res, err := decisionsPublicCall(t, u, "hooked", "req-mixed-final-"+name, req, hooks)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "old-names" || name == "old-labels" {
				checkUnaryFailure(ev, res, err, ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError)
			} else {
				ev.Check("final complete collection", err == nil && len(res.Answers) == 3 && res.Answers["from-score"] != nil && res.Answers["from-bool"] != nil && res.Answers["from-choice"] != nil, "got %+v / %v", res.Answers, err)
				if name == "all-refusal" {
					ev.Check("final names all refusal", jsonEqual(mustMarshal(t, res.Answers), []byte(`{"from-score":{"type":"refusal"},"from-choice":{"type":"refusal"},"from-bool":{"type":"refusal"}}`)), "got %+v", res.Answers)
				} else {
					a, ok := res.Answers["from-choice"].(ai.ScoreAnswer)
					ev.Check("final rubric and original values", ok && a.Score == .75 && string(a.Legend[0]) == `"高"` && string(a.Legend[2]) == `"低"`, "got %+v", a)
				}
			}
			wire := `{"model":"host-decisions","input":"track order","questions":` + decisionsChangedQuestions + `}`
			ev.Fixture("expected-wire.json", []byte(wire))
			requests := u.provider.Requests()
			if ev.Check("one frozen attempt", len(requests) == 1 && len(res.Metadata.Attempts) == 1 && calls == 1, "got %+v", res.Metadata) {
				ev.Check("late mutation cannot change bytes", jsonEqual(requests[0].Body, []byte(wire)), "got %s", requests[0].Body)
			}
			ev.Check("entry object untouched", jsonEqual(mustMarshal(t, req), original), "input changed")
			ev.Check("usage preserved", res.Usage.Input == 296, "got %+v", res.Usage)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsMixedHookGuard(t *testing.T) {
	for _, name := range []string{"score-one", "score-eleven", "score-no-levels", "score-null-levels", "score-null-label", "score-label-object", "score-description-null", "score-description-object", "score-label-alias", "score-label-duplicate", "score-choices", "predicate-levels", "predicate-choices", "predicate-instructions-null", "predicate-instructions-array", "model-score-min", "model-score-max", "model-score-kind", "model-bool-kind", "question-limit", "question-bytes"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD25-mixed-guard-"+name+"-hooked")
			u := newDecisionsMixedWorld(t, func(c *ai.Config) {
				m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
				if name == "model-score-min" {
					m.Capabilities.MinScoreLevels = 3
				}
				if name == "model-score-max" {
					m.Capabilities.MaxScoreLevels = 2
				}
				if name == "model-score-kind" {
					m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionBool}
					m.Capabilities.MinScoreLevels = 0
					m.Capabilities.MaxScoreLevels = 0
				}
				if name == "model-bool-kind" {
					m.Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore}
				}
				if name == "question-limit" {
					c.Policy.Classifier.MaxQuestions = 1
				}
				if name == "question-bytes" {
					c.Policy.Classifier.MaxQuestionBytes = 256
				}
			})
			q := `{"type":"score","name":"changed","instructions":"Rate","levels":[{"label":"low"},{"label":"high"}]}`
			switch name {
			case "score-one":
				q = strings.Replace(q, `,{"label":"high"}`, "", 1)
			case "score-eleven":
				q = strings.Replace(q, `{"label":"low"},{"label":"high"}`, strings.TrimSuffix(strings.Repeat(`{"label":"same"},`, 11), ","), 1)
			case "score-no-levels":
				q = `{"type":"score","name":"changed","instructions":"Rate"}`
			case "score-null-levels":
				q = `{"type":"score","name":"changed","instructions":"Rate","levels":null}`
			case "score-null-label":
				q = strings.Replace(q, `"label":"low"`, `"label":null`, 1)
			case "score-label-object":
				q = strings.Replace(q, `"label":"low"`, `"label":{}`, 1)
			case "score-description-null":
				q = strings.Replace(q, `"label":"low"`, `"label":"low","description":null`, 1)
			case "score-description-object":
				q = strings.Replace(q, `"label":"low"`, `"label":"low","description":{}`, 1)
			case "score-label-alias":
				q = strings.Replace(q, `"label":"low"`, `"Label":"low"`, 1)
			case "score-label-duplicate":
				q = strings.Replace(q, `"label":"low"`, `"label":"low","label":"low"`, 1)
			case "score-choices":
				q = strings.TrimSuffix(q, `}`) + `,"choices":[]}`
			case "predicate-levels", "predicate-choices", "predicate-instructions-null", "predicate-instructions-array", "model-bool-kind":
				q = `{"type":"predicate","name":"changed","instructions":"Judge"}`
				if name == "predicate-levels" {
					q = strings.TrimSuffix(q, `}`) + `,"levels":[]}`
				}
				if name == "predicate-choices" {
					q = strings.TrimSuffix(q, `}`) + `,"choices":[]}`
				}
				if name == "predicate-instructions-null" {
					q = strings.Replace(q, `"Judge"`, `null`, 1)
				}
				if name == "predicate-instructions-array" {
					q = strings.Replace(q, `"Judge"`, `[]`, 1)
				}
			case "model-score-max":
				q = strings.Replace(q, `{"label":"low"}`, `{"label":"low"},{"label":"middle"}`, 1)
			case "question-bytes":
				q = strings.Replace(q, `"Rate"`, `"`+strings.Repeat("x", 300)+`"`, 1)
			}
			res, err := decisionsPublicCall(t, u, "hooked", "req-mixed-guard-"+name, decisionsInput(), ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["questions"] = []any{decisionsRaw(q)}
				if name == "question-limit" {
					p.Body["questions"] = []any{decisionsRaw(q), decisionsRaw(q)}
				}
				return ai.KeepPayload(), nil
			}})
			code := ai.CodeCallbackFailed
			if name == "question-limit" || name == "question-bytes" {
				code = ai.CodeResourceLimit
			}
			checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
			ev.Check("zero send and attempt", len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsMixedPreflight(t *testing.T) {
	for _, name := range []string{"score-one", "score-eleven", "score-structured", "bool-true-structured", "bool-false-structured", "score-model-min", "score-model-max"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD13-mixed-preflight-"+name+"-"+entry)
				u := newDecisionsMixedWorld(t, func(c *ai.Config) {
					m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
					if name == "score-model-min" {
						m.Capabilities.MinScoreLevels = 4
					}
					if name == "score-model-max" {
						m.Capabilities.MaxScoreLevels = 2
					}
				})
				req := decisionsMixedInput()
				phase := ai.PhaseCapability
				switch name {
				case "score-one":
					req.Questions["urgency"] = baselineScore(1)
					phase = ai.PhaseScope
				case "score-eleven":
					req.Questions["urgency"] = baselineScore(11)
					phase = ai.PhaseScope
				case "score-structured":
					q := req.Questions["urgency"].(ai.ScoreQuestion)
					q.Criteria[1] = json.RawMessage(`{"label":"low"}`)
					req.Questions["urgency"] = q
				case "bool-true-structured", "bool-false-structured":
					q := req.Questions["approved"].(ai.BoolQuestion)
					if name == "bool-true-structured" {
						q.Criteria.True = json.RawMessage(`[]`)
					} else {
						q.Criteria.False = json.RawMessage(`{}`)
					}
					req.Questions["approved"] = q
				}
				id := "req-mixed-preflight-" + name + "-" + entry
				res, err := decisionsPublicCall(t, u, entry, id, req, decisionsNoopHooks())
				checkUnaryFailure(ev, res, err, ai.CodeInvalidRequest, phase, ai.StopReasonError)
				ev.Check("no credentials or attempts", u.host.ReadsFor(id).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

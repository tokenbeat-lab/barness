package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestClassifierMixedRequestRefusals(t *testing.T) {
	f, _ := loadFixture(t, classifierProtocol, "mixed.json")
	sc := f.Scenarios[0]
	changes := map[string]func(*ai.ClassifierRequest){
		"score-one-level": func(r *ai.ClassifierRequest) {
			r.Questions["anger"] = ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate"`), Criteria: []json.RawMessage{json.RawMessage(`"Only"`)}}
		},
		"score-eleven-levels": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			for range 8 {
				q.Criteria = append(q.Criteria, json.RawMessage(`"Extra"`))
			}
			r.Questions["anger"] = q
		},
		"score-null-level": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			q.Criteria[0] = json.RawMessage(`null`)
		},
		"score-number-level": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			q.Criteria[0] = json.RawMessage(`42`)
		},
		"score-large-level": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			q.Criteria[0] = mustMarshal(t, strings.Repeat("x", 5000))
		},
		"score-pointer":     func(r *ai.ClassifierRequest) { q := r.Questions["anger"].(ai.ScoreQuestion); r.Questions["anger"] = &q },
		"bool-missing-true": func(r *ai.ClassifierRequest) { r.Questions["urgent"].(ai.BoolQuestion).Criteria.True = nil },
		"bool-null-false": func(r *ai.ClassifierRequest) {
			r.Questions["urgent"].(ai.BoolQuestion).Criteria.False = json.RawMessage(`null`)
		},
		"bool-large-criterion": func(r *ai.ClassifierRequest) {
			r.Questions["urgent"].(ai.BoolQuestion).Criteria.True = mustMarshal(t, strings.Repeat("x", 5000))
		},
		"empty-object-instructions": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			q.Instructions = json.RawMessage(`{}`)
			r.Questions["anger"] = q
		},
		"empty-array-instructions": func(r *ai.ClassifierRequest) {
			q := r.Questions["urgent"].(ai.BoolQuestion)
			q.Instructions = json.RawMessage(`[]`)
			r.Questions["urgent"] = q
		},
		"scalar-instructions": func(r *ai.ClassifierRequest) {
			q := r.Questions["urgent"].(ai.BoolQuestion)
			q.Instructions = json.RawMessage(`true`)
			r.Questions["urgent"] = q
		},
		"duplicate-description": func(r *ai.ClassifierRequest) {
			q := r.Questions["anger"].(ai.ScoreQuestion)
			q.Instructions = json.RawMessage(`{"x":1,"x":2}`)
			r.Questions["anger"] = q
		},
		"duplicate-state": func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`{"x":1,"x":2}`) },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E07-mixed-"+name)
			w := scenarioWorld(t, sc)
			req := classifierInput(t, sc)
			change(&req)
			res, err := w.client.Classify(ctxFor(t), textScope("req-mixed-invalid"), sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			var ae *ai.Error
			ev.Check("invalid input before credentials or HTTP", errors.As(err, &ae) && ae.Phase == ai.PhaseScope && len(res.Answers) == 0 && len(w.provider.Requests()) == 0 && w.host.ReadsFor("req-mixed-invalid").Credentials == 0, "got %v", err)
		})
	}
}

func TestClassifierClosedQuestionJSON(t *testing.T) {
	for name, question := range map[string]string{
		"unknown":               `{"type":"other","instructions":"x"}`,
		"native-noul":           `{"type":"noul","instructions":"x"}`,
		"score-object-criteria": `{"type":"score","instructions":"x","criteria":{"0":"x"}}`,
		"choice-array-criteria": `{"type":"choice","instructions":"x","criteria":["x","y"]}`,
		"bool-array-criteria":   `{"type":"bool","instructions":"x","criteria":["x","y"]}`,
		"bool-extra-criteria":   `{"type":"bool","instructions":"x","criteria":{"true":"x","false":"y","extra":"z"}}`,
		"score-extra-field":     `{"type":"score","instructions":"x","criteria":["x","y"],"choice":"x"}`,
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E07-closed-json-"+name)
			var req ai.ClassifierRequest
			err := json.Unmarshal([]byte(`{"state":{},"questions":{"q":`+question+`}}`), &req)
			ev.Record("error", errString(err))
			ev.Check("closed input decoding rejects wrong fields", err != nil, "got %+v", req)
		})
	}
}

func TestClassifierMixedCallbacks(t *testing.T) {
	f, raw := loadFixture(t, classifierProtocol, "mixed.json")
	sc := f.Scenarios[0]
	for _, name := range []string{"final-questions", "final-kind-mismatch", "final-legend-mismatch", "model-widen", "count-limit", "state-limit", "question-limit", "wrong-shape", "unencodable", "score-capability", "bool-capability", "numeric-precision"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E04-mixed-"+name)
			ev.Fixture("fixture.json", raw)
			w := scenarioWorld(t, sc)
			req := classifierInput(t, sc)
			if name == "numeric-precision" {
				req.State = json.RawMessage(`{"id":9007199254740993,"decimal":0.12345678901234567890123456789}`)
			}
			original := mustMarshal(t, req)
			body := `{"model":"actual-version","answers":{"changed":{"type":"score","score":0.25,"confidence":0.7,"probabilities":{"0":0.75,"1":0.25},"legend":{"0":"High","1":"Low"}},"yes":{"type":"noul","noul":0.2}},"usage":{"input_tokens":296,"output_tokens":20}}`
			if name == "final-kind-mismatch" {
				body = strings.Replace(body, `"type":"noul"`, `"type":"choice"`, 1)
			}
			if name == "final-legend-mismatch" {
				body = strings.Replace(body, `"0":"High","1":"Low"`, `"0":"Low","1":"High"`, 1)
			}
			if name == "numeric-precision" {
				body = sc.Replies[0].Body
			}
			enqueue(ev, w, fixtureReply{Status: 200, Body: body}.script(t, classifierProtocol, ""))
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				if name == "numeric-precision" {
					state := p.Body["state"].(map[string]any)
					ev.Check("callback sees exact input numbers", state["id"] == json.Number("9007199254740993") && state["decimal"] == json.Number("0.12345678901234567890123456789"), "got %+v", state)
					p.Body["state"] = map[string]any{"id": json.Number("9007199254740993"), "decimal": json.Number("0.12345678901234567890123456789")}
					return ai.KeepPayload(), nil
				}
				p.Body["questions"] = map[string]any{
					"changed": map[string]any{"type": "score", "instructions": "New score", "criteria": []any{"High", "Low"}},
					"yes":     map[string]any{"type": "noul", "instructions": "New bool"},
				}
				switch name {
				case "model-widen":
					p.Body["model"] = "other"
				case "count-limit":
					for i := range 9 {
						p.Body["questions"].(map[string]any)[string(rune('a'+i))] = map[string]any{"type": "noul", "instructions": "Extra"}
					}
				case "state-limit":
					p.Body["state"] = strings.Repeat("x", 5000)
				case "question-limit":
					p.Body["questions"].(map[string]any)["yes"].(map[string]any)["instructions"] = strings.Repeat("x", 5000)
				case "wrong-shape":
					p.Body["questions"].(map[string]any)["changed"].(map[string]any)["criteria"] = map[string]any{"0": "Bad"}
				case "unencodable":
					p.Body["state"] = math.NaN()
				case "score-capability":
					p.Body["questions"].(map[string]any)["changed"].(map[string]any)["criteria"] = []any{"a", "b", "c", "d"}
				}
				return ai.ReplacePayload(p.Body), nil
			}}
			if name == "score-capability" || name == "bool-capability" {
				cat := w.client.Catalog()
				if name == "score-capability" {
					cat.ClassifierModels[0].Capabilities.MaxScoreLevels = 3
				} else {
					cat.ClassifierModels[0].Capabilities.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore}
					delete(req.Questions, "urgent")
					original = mustMarshal(t, req)
				}
				p := validPolicy()
				p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
				var err error
				w.client, err = ai.NewClient(ai.Config{Policy: p, Bindings: w.host, Credentials: w.host, Catalog: &cat, AllowLoopbackHTTP: true})
				if err != nil {
					t.Fatal(err)
				}
			}
			res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-mixed-callback"), sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("original request unchanged", string(mustMarshal(t, req)) == string(original), "input was mutated")
			switch name {
			case "final-questions":
				a, ok := res.Answers["changed"].(ai.ScoreAnswer)
				b, boolean := res.Answers["yes"].(ai.BoolAnswer)
				ev.Check("final questions and ordered levels authoritative", err == nil && len(res.Answers) == 2 && ok && a.Score == 0.25 && boolean && b.Probability == 0.2, "got %v / %+v", err, res.Answers)
				ev.Check("response identity separate", res.Metadata.ModelID == "host-jev" && res.ResponseModel == ai.Value("actual-version"), "got %+v", res)
			case "numeric-precision":
				reqs := w.provider.Requests()
				ev.Check("raw callback numbers survive", err == nil && len(reqs) == 1 && strings.Contains(string(reqs[0].Body), "9007199254740993") && strings.Contains(string(reqs[0].Body), "0.12345678901234567890123456789"), "got %v", err)
			case "final-kind-mismatch", "final-legend-mismatch":
				ev.Check("validate final answers atomically", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0 && res.Usage.Input == 296, "got %v / %+v", err, res)
			default:
				code := ai.CodeCallbackFailed
				if name == "model-widen" {
					code = ai.CodeTenantDenied
				}
				ev.Check("invalid callback before HTTP", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0 && len(res.Answers) == 0, "got %v", err)
			}
		})
	}
}

func TestClassifierMixedSnapshots(t *testing.T) {
	ev := run.Case(t, "P07-E07-mixed-snapshots")
	f, _ := loadFixture(t, classifierProtocol, "mixed.json")
	sc := f.Scenarios[0]
	w := scenarioWorld(t, sc)
	req := classifierInput(t, sc)
	var retained *ai.Payload
	hooks := ai.Hooks{
		TransformHeaders: func(context.Context, ai.CallScope, http.Header) error {
			req.Questions["anger"].(ai.ScoreQuestion).Criteria[0][1] = 'X'
			req.Questions["urgent"].(ai.BoolQuestion).Criteria.True[1] = 'X'
			return nil
		},
		OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			retained = p
			return ai.KeepPayload(), nil
		},
		OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
			retained.Body["questions"] = map[string]any{}
			return nil
		},
	}
	enqueue(ev, w, sc.replies(t)...)
	res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-mixed-copy"), sc.target(), req, nil)
	ev.Record("result", res)
	ev.Record("error", errString(err))
	ev.Record("requests", w.provider.Requests())
	ev.Check("copied input and frozen callback independent", err == nil && len(res.Answers) == 3, "got %v", err)
}

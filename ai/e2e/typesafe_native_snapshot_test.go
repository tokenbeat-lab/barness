package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// These public regressions are frozen before replacing native -> public
// decoding. They exercise owned final questions, not private snapshot types.
func TestTypeSafeNativeSnapshot(t *testing.T) {
	const questions = `{"renamed":{"type":"score","instructions":{"rate":9007199254740993},"criteria":[{"id":9007199254740993},["low",1]]},"single":{"type":"choice","instructions":["Choose"],"criteria":{"only":null}},"yes":{"type":"noul","instructions":"Ask","criteria":{"true":{},"false":[]}}}`
	const answers = `{"renamed":{"type":"score","score":0.25,"confidence":0.8,"probabilities":{"0":0.75,"1":0.25},"legend":{"0":{"id":9007199254740993},"1":["low",1]}},"single":{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1},"yes":{"type":"noul","noul":0.4}}`
	for _, name := range []string{"keep", "replace", "retry", "old-legend", "bad-kind"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-JN01-native-snapshot-"+name)
			u := newUnaryWorld(t, nil)
			u.world.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
			req := classifierInput(t, u.sc)
			original := mustMarshal(t, req)
			responseAnswers := answers
			if name == "old-legend" {
				responseAnswers = strings.Replace(answers, `9007199254740993`, `9007199254740992`, 1)
			}
			if name == "bad-kind" {
				responseAnswers = strings.Replace(answers, `"type":"noul","noul":0.4`, `"type":"refusal"`, 1)
			}
			if name == "retry" {
				enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"final-native-model","answers":` + responseAnswers + `,"usage":{"input_tokens":296,"output_tokens":20}}`}.script(t, classifierProtocol, ""))
			frozenQuestions := json.RawMessage(questions)
			var retained map[string]any
			calls := [3]int{}
			hooks := ai.Hooks{
				TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { calls[0]++; return nil },
				OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					calls[1]++
					retained = map[string]any{"model": "host-jev", "state": json.RawMessage(`{"id":9007199254740993}`), "questions": frozenQuestions}
					if name == "replace" {
						return ai.ReplacePayload(retained), nil
					}
					p.Body = retained
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					calls[2]++
					// Mutate both retained containers and the raw criteria bytes.
					// A borrowed snapshot would now reject the valid final legend.
					for i := range frozenQuestions {
						frozenQuestions[i] = ' '
					}
					retained["questions"] = nil
					retained["model"] = "foreign"
					return nil
				},
			}
			ev.Record("input", req)
			ev.Fixture("final-questions.json", []byte(questions))
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-native-"+name), u.sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("hook-counts", calls)
			ev.Check("hooks once per logical call", calls == [3]int{1, 1, 1}, "got %v", calls)
			ev.Check("public request unchanged", string(mustMarshal(t, req)) == string(original), "input changed")
			wantAttempts := 1
			if name == "retry" {
				wantAttempts = 2
			}
			requests := u.provider.Requests()
			ev.Check("only initial failures retry", len(requests) == wantAttempts && len(res.Metadata.Attempts) == wantAttempts, "requests=%d attempts=%d", len(requests), len(res.Metadata.Attempts))
			wire := []byte(`{"model":"host-jev","state":{"id":9007199254740993},"questions":` + questions + `}`)
			for _, request := range requests {
				ev.Check("final native wire", jsonEqual(request.Body, wire) && strings.Contains(string(request.Body), "9007199254740993"), "got %s", request.Body)
				ev.Check("retry bytes and credential fixed", string(request.Body) == string(requests[0].Body) && request.KeyAlias == tenantA.alias, "request changed")
			}
			if name == "old-legend" || name == "bad-kind" {
				ev.Check("one bad answer clears all", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0, "got %v", err)
			} else {
				a, ok := res.Answers["renamed"].(ai.ScoreAnswer)
				b, boolean := res.Answers["yes"].(ai.BoolAnswer)
				c, choice := res.Answers["single"].(ai.ChoiceAnswer)
				ev.Check("final native kinds and reordered criteria", err == nil && len(res.Answers) == 3 && ok && a.Score == .25 && string(a.Legend[0]) == `{"id":9007199254740993}` && boolean && b.Probability == .4 && choice && c.Choice == "only", "got %v / %+v", err, res.Answers)
			}
			ev.Check("usage and response identity retained", res.Usage.Input == 296 && res.Usage.Output == 20 && res.ResponseModel == ai.Value("final-native-model") && res.Metadata.ModelID == "host-jev", "got %+v", res)
			reads := u.host.ReadsFor(res.Metadata.RequestID)
			ev.Record("host-reads", reads)
			ev.Check("one pinned configuration", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestTypeSafeNativeGuard(t *testing.T) {
	for _, tc := range []struct {
		name     string
		question string
	}{
		{"public-bool", `{"type":"bool","instructions":"Ask"}`},
		{"box", `{"type":"box","instructions":"Ask"}`},
		{"case-tag", `{"Type":"noul","instructions":"Ask"}`},
		{"duplicate-tag", `{"type":"noul","type":"noul","instructions":"Ask"}`},
		{"duplicate-description", `{"type":"noul","instructions":{"a":1,"a":2}}`},
		{"duplicate-choice", `{"type":"choice","instructions":"Ask","criteria":{"a":null,"a":null}}`},
		{"cross-kind", `{"type":"noul","instructions":"Ask","choices":[]}`},
		{"null-bool", `{"type":"noul","instructions":"Ask","criteria":null}`},
		{"case-bool", `{"type":"noul","instructions":"Ask","criteria":{"True":{},"false":[]}}`},
		{"missing-bool", `{"type":"noul","instructions":"Ask","criteria":{"true":{}}}`},
		{"empty-instructions", `{"type":"noul","instructions":{}}`},
		{"scalar-instructions", `{"type":"noul","instructions":true}`},
		{"null-choice", `{"type":"choice","instructions":"Ask","criteria":null}`},
		{"array-choice", `{"type":"choice","instructions":"Ask","criteria":["a"]}`},
		{"scalar-choice", `{"type":"choice","instructions":"Ask","criteria":{"a":42}}`},
		{"short-score", `{"type":"score","instructions":"Ask","criteria":["a"]}`},
		{"map-score", `{"type":"score","instructions":"Ask","criteria":{"0":"a","1":"b"}}`},
		{"null-score", `{"type":"score","instructions":"Ask","criteria":["a",null]}`},
		{"kind-capability", `{"type":"noul","instructions":"Ask"}`},
		{"format-capability", `{"type":"choice","instructions":{"task":"Ask"},"criteria":{"a":null}}`},
		{"min-choice-capability", `{"type":"choice","instructions":"Ask","criteria":{"a":null}}`},
		{"score-capability", `{"type":"score","instructions":"Ask","criteria":["a","b","c"]}`},
		{"encoded-exact", `{"type":"noul","instructions":"<<<<<"}`},
		{"encoded-over", `{"type":"noul","instructions":"<<<<<<"}`},
		{"raw-question-map", `{"type":"noul","instructions":"Ask"}`},
		{"model-before-size", `{"type":"noul","instructions":"Ask"}`},
		{"authority-before-shape", `{"type":"noul","instructions":"Ask"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "P07-JN10-native-guard-"+tc.name)
			u := newUnaryWorld(t, func(c *ai.Config) {
				caps := &c.Catalog.ClassifierModels[0].Capabilities
				switch tc.name {
				case "kind-capability":
					caps.Kinds = []ai.ClassifierQuestionKind{ai.ClassifierQuestionChoice, ai.ClassifierQuestionScore}
				case "format-capability":
					caps.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
				case "min-choice-capability":
					caps.MinChoices = 2
				case "score-capability":
					caps.MaxScoreLevels = 2
				case "encoded-exact", "encoded-over":
					// Canonical bool question: 33 framing bytes plus five
					// six-byte HTML escapes = 63; native noul is also 4 bytes.
					c.Policy.Classifier.MaxQuestionBytes = 63
				}
			})
			req := classifierInput(t, u.sc)
			if strings.HasPrefix(tc.name, "encoded-") {
				req.Questions = map[string]ai.ClassifierQuestion{"q": ai.BoolQuestion{Instructions: json.RawMessage(`"Ask"`)}}
			}
			ev.Record("input", req)
			ev.Fixture("hook-question.json", []byte(tc.question))
			if tc.name == "encoded-exact" {
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["questions"] = map[string]any{"q": json.RawMessage(tc.question)}
				switch tc.name {
				case "raw-question-map":
					// A raw map bypasses borrowed map inspection; its exact
					// schema must still be checked on the frozen bytes.
					p.Body["questions"] = json.RawMessage(`{"q":{"Type":"noul","instructions":"Ask"}}`)
				case "model-before-size":
					p.Body["model"], p.Body["state"] = "foreign", strings.Repeat("x", 5000)
				case "authority-before-shape":
					p.Body["endpoint"], p.Body["state"] = "https://other.example", nil
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-native-guard-"+tc.name), u.sc.target(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if tc.name == "encoded-exact" {
				ev.Check("canonical encoded limit inclusive", err == nil && len(res.Answers) == 1 && len(res.Metadata.Attempts) == 1, "got %v", err)
			} else {
				code := ai.CodeCallbackFailed
				if tc.name == "model-before-size" || tc.name == "authority-before-shape" {
					code = ai.CodeTenantDenied
				}
				ev.Check("native rejection before attempt", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && len(res.Answers) == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0 && len(u.adm.Requests()) == 0, "got %v", err)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

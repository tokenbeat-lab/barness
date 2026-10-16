package e2e

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestDecisionsImageInputsFailures(t *testing.T) {
	for _, name := range []string{"bad-answer", "one-bad-answer", "duplicate-answer", "bad-usage", "duplicate-usage", "bad-model", "malformed", "read", "http", "response-hook", "cancel", "refusal", "usage-missing", "usage-null"} {
		for _, entry := range []string{"plain", "hooked"} {
			if entry == "plain" && (name == "response-hook" || name == "cancel") {
				continue
			}
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC03-atomic-"+name+"-"+entry)
				u := newReferenceWorld(t, ev, nil)
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
				req := decisionsInput()
				req.Images = classifierAssets(t)[:2]
				recordClassifierAssets(ev, classifierAssets(t))
				ev.Fixture("public-request.json", mustMarshal(t, req))
				answer, usage, model := decisionsAnswer, decisionsUsage, `"model":"response-version",`
				code, phase, stop := ai.CodeProtocol, ai.PhaseResponse, ai.StopReasonError
				want := ai.Usage{Input: 296, Output: 20, TotalTokens: 316, Cost: ai.UsageCost{Input: .000296, Total: .000296}}
				wantModel := true
				switch name {
				case "bad-answer":
					answer = strings.Replace(answer, `"choice":"track"`, `"choice":"foreign"`, 1)
				case "one-bad-answer":
					req.Questions["broken"] = req.Questions["intent"]
					answer += "," + strings.ReplaceAll(strings.Replace(decisionsAnswer, `"choice":"track"`, `"choice":"foreign"`, 1), `"intent"`, `"broken"`)
				case "duplicate-answer":
					req.Questions["broken"] = req.Questions["intent"]
					answer += "," + decisionsAnswer
				case "bad-usage":
					usage, want, wantModel = `"usage":{"input_tokens":-1}`, ai.Usage{}, false
				case "duplicate-usage":
					usage, want = decisionsUsage+","+decisionsUsage, ai.Usage{}
				case "bad-model":
					model, wantModel = `"model":null,`, false
				case "malformed", "read":
					want, wantModel = ai.Usage{}, false
					if name == "read" {
						code = ai.CodeTransport
					}
				case "http":
					code, phase, want, wantModel = ai.CodeUpstreamError, ai.PhaseRequest, ai.Usage{}, false
				case "response-hook":
					code, phase, want, wantModel = ai.CodeCallbackFailed, ai.PhaseRequest, ai.Usage{}, false
				case "cancel":
					code, phase, stop, want, wantModel = ai.CodeCanceled, ai.PhaseRequest, ai.StopReasonAborted, ai.Usage{}, false
				case "refusal":
					answer, code = `{"type":"refusal","name":"intent"}`, ""
				case "usage-missing":
					usage, code, want = `"vendor":0`, "", ai.Usage{}
				case "usage-null":
					usage, code, want = `"usage":null`, "", ai.Usage{}
				}
				// Save after adding questions so the complete input is repeatable.
				ev.Record("public-input", req)
				body := `{` + model + `"answers":[` + answer + `],` + usage + `}`
				if name == "malformed" {
					body = "{"
				}
				reply := fixtureReply{Status: 200, Body: body}.script(t, decisionsProtocol, "")
				if name == "read" {
					reply.End = provider.EndAbort
				}
				if name == "http" {
					reply = fixtureReply{Status: 503, Header: map[string]string{"x-should-retry": "false"}, Body: `{"error":{"message":"synthetic failure"}}`}.script(t, decisionsProtocol, "")
				}
				enqueue(ev, u.world, reply)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					if name == "cancel" {
						cancel()
					}
					if name == "response-hook" {
						return errors.New("synthetic private callback error")
					}
					return nil
				}}
				id := textScope("req-image-atomic-" + name + "-" + entry)
				target := ai.Target{BindingID: "classifier", ModelID: "host-decisions"}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctx, id, target, req, nil)
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctx, id, target, req, nil)
				}
				if code != "" {
					checkUnaryFailure(ev, res, err, code, phase, stop)
					checkImageInputs(ev, res, nil)
				} else {
					ev.Check("valid answer/refusal succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				}
				_, hasModel := res.ResponseModel.Get()
				ev.Check("only validated usage and original ResponseModel contract survive", reflect.DeepEqual(res.Usage, want) && hasModel == wantModel, "usage/model mismatch")
				ev.Check("2xx failure cannot replay; HTTP opt-out honored", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "replayed request")
				recordImageInputsCall(ev, u, res, err)
			})
		}
	}
}

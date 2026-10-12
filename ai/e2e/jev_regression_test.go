package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Issue 01 is a public behavior baseline, not a production fix. Failures to
// catch before any adapter work: narrowed single-choice/JSON/empty-state/bool
// inputs, changed score bounds or rounding, case aliases/duplicates/refusal
// accepted, float64 precision loss, validation of pre-hook questions, changed
// usage/model/attribution, extra callbacks/retries, content in Observer, and
// leaked HTTP bodies or host permits. Existing assertions remain independent.
func TestJevRegression(t *testing.T) {
	raw, err := os.ReadFile("testdata/typesafe/jev-regression.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			ID       string
			Request  ai.ClassifierRequest
			Wire     json.RawMessage
			Response string
			Answers  json.RawMessage
			Code     ai.Code
			Phase    ai.Phase
		}
	}
	mustUnmarshal(t, raw, &fixture)
	for _, sc := range fixture.Cases {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(sc.ID+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P07-OD01-jev-"+sc.ID+"-"+entry)
				ev.Fixture("fixture.json", raw)
				u := newUnaryWorld(t, nil)
				// A successful HTTP response may never be replayed, even if its
				// answers are malformed. A second request has no scripted reply.
				u.world.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				enqueue(ev, u.world, fixtureReply{Status: 200, Header: map[string]string{"x-typesafe-request-id": "req-jev-baseline"}, Body: sc.Response}.script(t, classifierProtocol, ""))
				order := []string{}
				hooks := ai.Hooks{
					TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { order = append(order, "headers"); return nil },
					OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						order = append(order, "payload")
						ev.Check("hook receives Jev protocol", p.API == ai.APITypeSafeSystemOne && p.Operation == ai.OperationClassifier, "got %+v", p)
						return ai.KeepPayload(), nil
					},
					OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
						order = append(order, "response")
						return nil
					},
				}
				scope := textScope("req-jev-" + sc.ID + "-" + entry)
				scope.ActorID, scope.JobID, scope.Operation = "baseline-actor", "baseline-job", ai.OperationImage
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, u.sc.target(), sc.Request, nil)
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), scope, u.sc.target(), sc.Request, ai.TypeSafeOptions{})
				}
				ev.Record("result", res)
				ev.Record("error", errString(err))
				ev.Record("attempts", res.Metadata.Attempts)
				ev.Record("hook-order", order)
				if sc.Code == "" {
					ev.Check("success", err == nil && res.StopReason == ai.StopReasonStop, "got %v", err)
					ev.Check("complete typed answers", jsonEqual(mustMarshal(t, res.Answers), sc.Answers), "got %+v", res.Answers)
				} else {
					ev.Check("exact atomic rejection", errors.Is(err, &ai.Error{Code: sc.Code, Phase: sc.Phase}) && res.StopReason == ai.StopReasonError && len(res.Answers) == 0, "got %v", err)
				}
				expectedOrder := []string{}
				if entry == "hooked" {
					expectedOrder = []string{"headers", "payload", "response"}
				}
				ev.Check("hook order and exact counts", reflect.DeepEqual(order, expectedOrder), "got %v", order)
				requests := u.provider.Requests()
				if ev.Check("one HTTP attempt and permit", len(requests) == 1 && len(res.Metadata.Attempts) == 1 && len(u.adm.Requests()) == 1, "requests=%d attempts=%d", len(requests), len(res.Metadata.Attempts)) {
					r := requests[0]
					ev.Check("exact native wire", r.Method == "POST" && r.Path == "/v1/systemone" && r.KeyAlias == tenantA.alias && jsonEqual(r.Body, sc.Wire), "got %+v", r)
					for _, n := range []string{"9007199254740993", "12345678901234567890"} {
						ev.Check("unrounded JSON number "+n, strings.Count(string(r.Body), n) == strings.Count(string(sc.Wire), n), "number changed")
					}
					a := res.Metadata.Attempts[0]
					ev.Check("attempt retains usage and vendor ID", a.Usage == res.Usage && a.UsageReporting == ai.UsageComplete && a.HTTPStatus == 200 && a.ProviderRequestID == "req-jev-baseline", "got %+v", a)
				}
				wantUsage := ai.Usage{Input: 296, Output: 20, TotalTokens: 316, Cost: ai.UsageCost{Input: .000296, Total: .000296}}
				ev.Check("known usage and input-only cost", res.Usage == wantUsage, "got %+v", res.Usage)
				model, ok := res.ResponseModel.Get()
				ev.Check("response model survives invalid answers", ok && model == "jev-baseline-version", "got %s", model)
				hash, hashErr := u.client.Catalog().Hash()
				m := res.Metadata
				ev.Check("trusted attribution and snapshots", hashErr == nil && m.TenantID == tenantA.tenant && m.RequestID == scope.RequestID && m.ActorID == scope.ActorID && m.JobID == scope.JobID && m.Operation == ai.OperationClassifier && m.Resolved && m.BindingID == "classifier" && m.ProviderID == ai.ProviderTypeSafe && m.API == ai.APITypeSafeSystemOne && m.ModelID == "host-jev" && m.AccountScopeID == "acct-tenant-a" && m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.CatalogHash == hash, "got %+v", m)
				reads := u.host.ReadsFor(scope.RequestID)
				ev.Record("host-reads", reads)
				ev.Check("single binding and credential snapshot", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)
				u.records(ev, res, err)
				obs := u.rec.forRequest(scope.RequestID)
				ev.Check("ordered Observer lifecycle", reflect.DeepEqual(kinds(obs), lifecycle(scope.RequestID, 1)), "got %v", kinds(obs))
				u.released(ev)
			})
		}
	}
}

func TestJevRegressionFinalQuestions(t *testing.T) {
	for _, name := range []string{"final-answer", "old-answer"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-OD16-jev-final-"+name)
			u := newUnaryWorld(t, nil)
			reply := u.sc.replies(t)[0]
			wire := []byte(`{"model":"host-jev","state":{"ticket":"track order"},"questions":{"changed":{"type":"choice","instructions":["New question"],"criteria":{"only":{"meaning":"One"}}}}}`)
			if name == "final-answer" {
				reply = fixtureReply{Status: 200, Body: `{"model":"version","answers":{"changed":{"type":"choice","choice":"only","probabilities":{"only":1},"confidence":1}},"usage":{"input_tokens":296,"output_tokens":20}}`}.script(t, classifierProtocol, "")
			}
			enqueue(ev, u.world, reply)
			payloads := 0
			res, err := u.call(t, ctxFor(t), "req-jev-final-"+name, ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				payloads++
				p.Body["questions"] = map[string]any{"changed": map[string]any{"type": "choice", "instructions": []any{"New question"}, "criteria": map[string]any{"only": map[string]any{"meaning": "One"}}}}
				return ai.KeepPayload(), nil
			}})
			ev.Fixture("expected-wire.json", wire)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("hook-count", payloads)
			ev.Check("payload once", payloads == 1, "got %d", payloads)
			requests := u.provider.Requests()
			if ev.Check("one final native request", len(requests) == 1, "got %d", len(requests)) {
				ev.Check("final wire frozen", jsonEqual(requests[0].Body, wire), "got %s", requests[0].Body)
			}
			if name == "final-answer" {
				a, ok := res.Answers["changed"].(ai.ChoiceAnswer)
				ev.Check("new question accepted", err == nil && ok && a.Choice == "only" && len(res.Answers) == 1, "got %v", err)
			} else {
				ev.Check("old question rejected", errors.Is(err, &ai.Error{Code: ai.CodeProtocol, Phase: ai.PhaseResponse}) && len(res.Answers) == 0, "got %v", err)
			}
			ev.Check("known usage retained", res.Usage.Input == 296 && res.Usage.Output == 20, "got %+v", res.Usage)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

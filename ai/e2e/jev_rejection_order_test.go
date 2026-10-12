package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Before implementation, enumerate ordering failures: a resolver/credential
// read, hook, admission, Attempt or HTTP request on a scope rejection; foreign
// ownership, disabled binding, operation/provider/API/model denial reaching
// credentials; and unresolved attribution echoing the requested model.
func TestJevRegressionRejectionOrder(t *testing.T) {
	changes := []struct {
		name   string
		change func(*ai.ClassifierRequest)
	}{
		{"no-questions", func(r *ai.ClassifierRequest) { r.Questions = nil }},
		{"empty-name", func(r *ai.ClassifierRequest) { r.Questions[""] = r.Questions["intent"] }},
		{"nil-question", func(r *ai.ClassifierRequest) { r.Questions["intent"] = nil }},
		{"state-null", func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`null`) }},
		{"state-number", func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`42`) }},
		{"state-bool", func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`false`) }},
		{"state-malformed", func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`{`) }},
		{"state-duplicate", func(r *ai.ClassifierRequest) { r.State = json.RawMessage(`{"x":1,"x":2}`) }},
		{"choice-256", func(r *ai.ClassifierRequest) {
			q := r.Questions["intent"].(ai.ChoiceQuestion)
			q.Criteria = map[string]json.RawMessage{}
			for i := range 256 {
				q.Criteria[fmt.Sprint(i)] = json.RawMessage(`"Candidate"`)
			}
			r.Questions["intent"] = q
		}},
		{"score-1", func(r *ai.ClassifierRequest) { r.Questions["intent"] = baselineScore(1) }},
		{"score-11", func(r *ai.ClassifierRequest) { r.Questions["intent"] = baselineScore(11) }},
	}
	for _, tc := range changes {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(tc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P07-OD09-jev-"+tc.name+"-"+entry)
				// Legacy catalog construction permits wider declared maxima;
				// public request bounds must still win before Binding resolution.
				u := newUnaryWorld(t, func(c *ai.Config) {
					c.Catalog.ClassifierModels[0].Capabilities.MaxChoices = 300
					c.Catalog.ClassifierModels[0].Capabilities.MaxScoreLevels = 20
				})
				req := classifierInput(t, u.sc)
				tc.change(&req)
				ev.Record("input", map[string]any{"stateJSON": string(req.State), "questions": req.Questions})
				scope := textScope("req-jev-order-" + tc.name + "-" + entry)
				target := ai.Target{BindingID: "missing-binding", ModelID: "unauthorized-model"}
				hookCalls := 0
				hooks := ai.Hooks{
					TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { hookCalls++; return nil },
					OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
						hookCalls++
						return ai.KeepPayload(), nil
					},
					OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { hookCalls++; return nil },
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, target, req, nil)
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), scope, target, req, nil)
				}
				ev.Record("result", res)
				ev.Record("error", errString(err))
				ev.Check("public shape precedes missing binding", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseScope}), "got %v", err)
				reads := u.host.ReadsFor(scope.RequestID)
				ev.Record("host-reads", reads)
				ev.Check("zero resolver and credential reads", reads.Bindings == 0 && reads.Credentials == 0, "got %+v", reads)
				ev.Check("zero hooks permits attempts and requests", hookCalls == 0 && len(u.adm.Requests()) == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "hooks=%d metadata=%+v", hookCalls, res.Metadata)
				m := res.Metadata
				ev.Check("unresolved entry attribution", !m.Resolved && m.Operation == ai.OperationClassifier && m.BindingID == target.BindingID && m.ModelID == "" && m.API == "" && m.ProviderID == "", "got %+v", m)
				obs := u.rec.awaitCall(ev, scope.RequestID)
				ev.Check("only call observations", reflect.DeepEqual(kinds(obs), lifecycle(scope.RequestID, 0)), "got %v", kinds(obs))
				u.released(ev)
			})
		}
	}
}

func baselineScore(n int) ai.ScoreQuestion {
	criteria := make([]json.RawMessage, n)
	for i := range criteria {
		criteria[i] = json.RawMessage(`"Level"`)
	}
	return ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate"`), Criteria: criteria}
}

func TestJevRegressionAuthorizationOrder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		code  ai.Code
		phase ai.Phase
	}{
		{"foreign-tenant", ai.CodeBindingNotFound, ai.PhaseBinding},
		{"foreign-owner", ai.CodeBindingNotFound, ai.PhaseBinding},
		{"disabled", ai.CodeTenantDenied, ai.PhaseBinding},
		{"operation", ai.CodeTenantDenied, ai.PhaseCapability},
		{"provider", ai.CodeInvalidRequest, ai.PhaseCapability},
		{"protocol", ai.CodeInvalidRequest, ai.PhaseCapability},
		{"not-allowed", ai.CodeTenantDenied, ai.PhaseCapability},
		{"not-in-catalog", ai.CodeInvalidRequest, ai.PhaseCapability},
	} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(tc.name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P07-OD03-jev-"+tc.name+"-"+entry)
				u := newUnaryWorld(t, nil)
				scope := textScope("req-jev-auth-" + tc.name + "-" + entry)
				target := u.sc.target()
				switch tc.name {
				case "foreign-tenant":
					scope.TenantID = "foreign"
				case "foreign-owner":
					b := primaryBinding(tenantB, u.provider.URL())
					b.BindingID = "classifier"
					catalog := u.client.Catalog()
					client, err := ai.NewClient(ai.Config{Policy: func() *ai.ResourcePolicy {
						p := validPolicy()
						p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
						p.MaxQueuedObservations = 64
						return p
					}(), Catalog: &catalog, Bindings: bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
						_, _ = u.host.ResolveBinding(ctx, s, id)
						return b, nil
					}), Credentials: u.host, Transport: u.bodies, Admission: u.adm, Observer: u.rec, Probe: u.gauges, AllowLoopbackHTTP: true})
					if err != nil {
						t.Fatal(err)
					}
					u.client = client
				default:
					u.world.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) {
						switch tc.name {
						case "disabled":
							b.Enabled = false
						case "operation":
							b.Operation = ai.OperationChat
						case "provider":
							b.ProviderID = ai.ProviderOpenAI
						case "protocol":
							b.API = ai.APIOpenAIResponses
						case "not-allowed":
							b.AllowedModels = nil
						case "not-in-catalog":
							b.AllowedModels = []string{"unknown"}
							target.ModelID = "unknown"
						}
					})
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, target, classifierInput(t, u.sc), nil)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), scope, target, classifierInput(t, u.sc), nil)
				}
				ev.Record("result", res)
				ev.Record("error", errString(err))
				ev.Check("exact authorization rejection", errors.Is(err, &ai.Error{Code: tc.code, Phase: tc.phase}), "got %v", err)
				reads := u.host.ReadsFor(scope.RequestID)
				ev.Record("host-reads", reads)
				ev.Check("one binding zero credentials attempts requests permits", reads.Bindings == 1 && reads.Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0 && len(u.adm.Requests()) == 0, "got %+v", reads)
				ev.Check("unresolved model omitted", !res.Metadata.Resolved && res.Metadata.ModelID == "" && res.Metadata.ProviderID == "" && len(res.Answers) == 0, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

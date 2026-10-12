package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsPreflight(t *testing.T) {
	for _, name := range []string{"scope-empty", "scope-name", "scope-nil", "scope-state", "scope-duplicate", "scope-choice-256", "scope-score-11", "single", "structured-instructions", "structured-criterion", "score", "bool", "foreign-tenant", "disabled", "operation", "provider", "protocol", "not-allowed", "catalog", "chat-only", "foreign-options", "typed-nil", "typed-nil-foreign", "model-min", "model-max"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD03-preflight-"+name+"-"+entry)
				u := newDecisionsWorld(t, func(c *ai.Config) {
					m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
					if name == "model-min" {
						m.Capabilities.MinChoices = 3
					}
					if name == "model-max" {
						m.Capabilities.MaxChoices = 2
					}
					if name == "chat-only" {
						c.Catalog.Models = append(c.Catalog.Models, ai.Model{Provider: ai.ProviderOpenAI, API: ai.APIOpenAIDecisions, ID: "chat-only", Input: []ai.Modality{ai.ModalityText}})
					}
				})
				scope := textScope("req-preflight-" + name + "-" + entry)
				target := ai.Target{BindingID: "classifier", ModelID: "host-decisions"}
				req := decisionsInput()
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				var opts ai.Options
				switch name {
				case "scope-empty":
					req.Questions = nil
				case "scope-name":
					req.Questions[""] = req.Questions["intent"]
				case "scope-nil":
					req.Questions["intent"] = nil
				case "scope-state":
					req.State = json.RawMessage(`null`)
				case "scope-duplicate":
					req.State = json.RawMessage(`{"x":1,"x":2}`)
				case "scope-choice-256":
					q := req.Questions["intent"].(ai.ChoiceQuestion)
					q.Criteria = map[string]json.RawMessage{}
					for i := range 256 {
						q.Criteria[fmt.Sprint(i)] = json.RawMessage(`null`)
					}
					req.Questions["intent"] = q
				case "scope-score-11":
					req.Questions["intent"] = baselineScore(11)
				case "single":
					q := req.Questions["intent"].(ai.ChoiceQuestion)
					delete(q.Criteria, "other")
					req.Questions["intent"] = q
				case "structured-instructions":
					q := req.Questions["intent"].(ai.ChoiceQuestion)
					q.Instructions = json.RawMessage(`["Classify"]`)
					req.Questions["intent"] = q
				case "structured-criterion":
					q := req.Questions["intent"].(ai.ChoiceQuestion)
					q.Criteria["track"] = json.RawMessage(`{"description":"Tracking"}`)
					req.Questions["intent"] = q
				case "score":
					req.Questions["intent"] = baselineScore(2)
				case "bool":
					req.Questions["intent"] = ai.BoolQuestion{Instructions: json.RawMessage(`"Judge"`)}
				case "foreign-tenant":
					scope.TenantID = "foreign"
					code, phase = ai.CodeBindingNotFound, ai.PhaseBinding
				case "disabled":
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Enabled = false })
					code, phase = ai.CodeTenantDenied, ai.PhaseBinding
				case "operation":
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Operation = ai.OperationChat })
					code = ai.CodeTenantDenied
				case "provider":
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.ProviderID = ai.ProviderGoogle })
				case "protocol":
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.API = ai.APIOpenAIResponses })
				case "not-allowed":
					target.ModelID = "forbidden"
					code = ai.CodeTenantDenied
				case "catalog", "chat-only":
					target.ModelID = name
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowedModels = append(b.AllowedModels, name) })
				case "foreign-options":
					opts = ai.TypeSafeOptions{}
				case "typed-nil":
					opts = (*ai.OpenAIDecisionsOptions)(nil)
				case "typed-nil-foreign":
					opts = (*ai.TypeSafeOptions)(nil)
				case "model-max":
					q := req.Questions["intent"].(ai.ChoiceQuestion)
					q.Criteria["third"] = json.RawMessage(`null`)
					req.Questions["intent"] = q
				}
				if len(name) > 6 && name[:6] == "scope-" {
					phase = ai.PhaseScope
				}
				hookCalls := 0
				hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
					hookCalls++
					return ai.KeepPayload(), nil
				}}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, target, req, opts)
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), scope, target, req, opts)
				}
				ev.Record("input", req)
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				reads := u.host.ReadsFor(scope.RequestID)
				ev.Record("host-reads", reads)
				wantBindings := 1
				if phase == ai.PhaseScope {
					wantBindings = 0
				}
				ev.Check("authorize before credentials", reads.Bindings == wantBindings && reads.Credentials == 0, "got %+v", reads)
				ev.Check("no hook attempt admission or request", hookCalls == 0 && len(res.Metadata.Attempts) == 0 && len(u.adm.Requests()) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
				ev.Check("failed entry does not leak requested identity", !res.Metadata.Resolved && res.Metadata.ModelID == "" && res.Metadata.Operation == ai.OperationClassifier, "got %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsOptionsDefaults(t *testing.T) {
	for i, opts := range []ai.Options{nil, ai.OpenAIDecisionsOptions{}, &ai.OpenAIDecisionsOptions{}} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("P10-OD04-options-%d-plain", i))
			u := newDecisionsWorld(t, nil)
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + decisionsAnswer + `]}`}.script(t, decisionsProtocol, ""))
			res, err := u.client.Classify(ctxFor(t), textScope(fmt.Sprintf("req-options-%d", i)), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), opts)
			ev.Check("defaults all work", err == nil && len(res.Answers) == 1, "got %v", err)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
	for _, opts := range []ai.Options{ai.OpenAIDecisionsOptions{}, &ai.OpenAIDecisionsOptions{}, (*ai.TypeSafeOptions)(nil)} {
		t.Run(fmt.Sprintf("jev-%T", opts), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("P10-OD04-jev-options-%T", opts))
			u := newUnaryWorld(t, nil)
			res, err := u.client.Classify(ctxFor(t), textScope(fmt.Sprintf("req-jev-options-%T", opts)), u.sc.target(), classifierInput(t, u.sc), opts)
			ev.Check("Jev refuses foreign or typed nil", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseCapability}) && len(res.Metadata.Attempts) == 0, "got %v", err)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

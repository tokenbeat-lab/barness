package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"reflect"
	"testing"
)

func TestJevCriteriaLifecycle(t *testing.T) {
	for _, name := range []string{"retry", "cancel", "snapshot", "actor", "tenant"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC07-lifecycle-"+name+"-"+entry)
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`"judge"`), Criteria: &ai.BoolCriteria{True: json.RawMessage(`["example",null]`), False: json.RawMessage(`"other"`)}}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/criteria/true/1"}}
				var original ai.ClassifierRequest
				mustUnmarshal(t, mustMarshal(t, req), &original)
				opts := jevContextOptions()
				var mutate func()
				u := newJevCriteriaWorld(t, ev, func(c *ai.Config) {
					if name == "snapshot" {
						previous := c.Bindings
						c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
							mutate()
							return previous.ResolveBinding(ctx, s, id)
						})
					}
				})
				mutate = func() {
					q := req.Questions["q"].(ai.BoolQuestion)
					q.Criteria.True[1] = 'X'
					req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`"changed"`)}
					req.Images[0] = jevContextImages(t)[1]
					req.ImageReferences[0].Pointer = "/state/missing"
					opts.ExperimentalImages = false
				}
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				if name == "cancel" {
					cancel()
				}
				scope := textScope("req-criteria-life-" + name + entry)
				if name == "actor" {
					u.host.RestrictActors(tenantA.tenant, "classifier", "different-actor")
				}
				if name == "tenant" {
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.TenantID = tenantB.tenant })
				}
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 1 })
				if name == "retry" {
					enqueue(ev, u.world, fixtureReply{Status: 429, Header: map[string]string{"retry-after-ms": "0"}, Body: `{"error":"limited"}`}.script(t, classifierProtocol, ""))
				}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctx, scope, jevContextTarget(), req, &opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctx, scope, jevContextTarget(), req, &opts)
				}
				if name == "cancel" || name == "actor" || name == "tenant" {
					code := ai.CodeTenantDenied
					if name == "cancel" {
						code = ai.CodeCanceled
					}
					if name == "tenant" {
						code = ai.CodeBindingNotFound
					}
					ev.Check("identity/cancellation denies before pin and send", errors.Is(err, &ai.Error{Code: code}) && u.host.ReadsFor(scope.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
				} else {
					ev.Check("frozen local criteria succeed", err == nil, "got %v", err)
					checkImageInputs(ev, res, original.Images)
					if name == "retry" {
						requests := u.provider.Requests()
						if ev.Check("retry has two Attempts", len(requests) == 2 && len(res.Metadata.Attempts) == 2, "unexpected attempts") {
							ev.Check("frozen final bytes and headers reused", string(requests[0].Body) == string(requests[1].Body) && reflect.DeepEqual(requests[0].Header, requests[1].Header), "replay changed")
						}
					}
				}
				recordJevContext(ev, u, original, res, err)
			})
		}
	}
}

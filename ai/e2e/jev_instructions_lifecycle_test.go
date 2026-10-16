package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevInstructionsOrdering(t *testing.T) {
	for _, name := range []string{"local", "state", "context", "same-bytes", "literal"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC06-order-"+name+"-"+entry)
				u := newJevInstructionsWorld(t, ev, nil)
				req := jevContextInput(t)
				req.Questions = map[string]ai.ClassifierQuestion{"a/~": ai.BoolQuestion{Instructions: json.RawMessage(`[null,"<image:ordinary>",null]`)}, "z": ai.BoolQuestion{Instructions: json.RawMessage(`null`)}}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/z/instructions"}, {ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/questions/a~1~0/instructions/0"}, {ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/a~1~0/instructions/2"}}
				want := []ai.Image{req.Images[1], req.Images[0]}
				switch name {
				case "state":
					req.State = json.RawMessage(`{"photo":null}`)
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"})
					want = req.Images
				case "context":
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
					want = req.Images
				case "same-bytes":
					req.Images[1] = req.Images[0]
					want = []ai.Image{req.Images[0], req.Images[0]}
				case "literal":
					req.Questions["a/~"] = ai.BoolQuestion{Instructions: json.RawMessage(`{"":null,"01":null,"图":null,"plain":null}`)}
					req.ImageReferences[1].Pointer = "/questions/a~1~0/instructions/"
					req.ImageReferences[2].Pointer = "/questions/a~1~0/instructions/01"
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/a~1~0/instructions/图"})
				}
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"a/~":{"type":"noul","noul":1},"z":{"type":"noul","noul":0}}}`}.script(t, classifierProtocol, ""))
				res, err := jevStateCall(t, u, entry, "req-instructions-order-"+name+entry, req, ai.Hooks{})
				ev.Check("public order is final traversal, independent of reference order", err == nil, "got %v", err)
				checkImageInputs(ev, res, want)
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevInstructionsLifecycle(t *testing.T) {
	for _, name := range []string{"retry", "cancel", "snapshot", "actor", "tenant"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC06-lifecycle-"+name+"-"+entry)
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`["ask",null]`)}
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/instructions/1"}}
				var original ai.ClassifierRequest
				mustUnmarshal(t, mustMarshal(t, req), &original)
				opts := jevContextOptions()
				var mutate func()
				u := newJevInstructionsWorld(t, ev, func(c *ai.Config) {
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
					q.Instructions[1] = 'X'
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
				scope := textScope("req-instructions-life-" + name + entry)
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
					ev.Check("frozen local instructions succeed", err == nil, "got %v", err)
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

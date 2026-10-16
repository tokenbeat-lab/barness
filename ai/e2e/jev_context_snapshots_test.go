package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextSelectionAuthority(t *testing.T) {
	for _, name := range []string{"policy", "binding", "model"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-selection-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, func(c *ai.Config) {
					if name == "policy" {
						c.Policy.Classifier.Images = nil
					}
					if name == "model" {
						m := &c.Catalog.ClassifierModels[0]
						m.Input = nil
						m.Capabilities.ImageLocations = nil
						m.Capabilities.ImageMIMETypes = nil
						m.Capabilities.MaxInputImages = 0
					}
				})
				if name == "binding" {
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				}
				req := jevContextInput(t)
				req.Images = nil
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-jev-select-"+name+entry), jevContextTarget(), req, jevContextOptions())
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-jev-select-"+name+entry), jevContextTarget(), req, jevContextOptions())
				}
				code := ai.CodeInvalidRequest
				if name == "binding" {
					code = ai.CodeTenantDenied
				}
				ev.Check("selection cannot emit unauthorized experiment", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseCapability}) && u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0, "got %v", err)
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevContextSnapshots(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC04-snapshot-"+entry)
			req := jevContextInput(t)
			original := req.Images
			opts := jevContextOptions()
			var mutate func()
			u := newJevContextWorld(t, ev, func(c *ai.Config) {
				previous := c.Bindings
				c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
					if mutate != nil {
						mutate()
					}
					return previous.ResolveBinding(ctx, s, id)
				})
			})
			ev.Record("original-input", req)
			mutate = func() {
				req.Images[0] = req.Images[1]
				req.State[1] = 'X'
				req.Questions["q"] = nil
				opts.ExperimentalImages = false
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			var res ai.ClassifierResult
			var err error
			if entry == "plain" {
				res, err = u.client.Classify(ctxFor(t), textScope("req-jev-snapshot-"+entry), jevContextTarget(), req, &opts)
			} else {
				res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-jev-snapshot-"+entry), jevContextTarget(), req, &opts)
			}
			ev.Check("input and option copied before external resolver", err == nil, "got %v", err)
			// Independent original file fixtures; the resolver changed the public slice.
			original = jevContextImages(t)
			checkImageInputs(ev, res, original)
			saved := res.ImageInputs
			if len(saved) > 0 {
				res.ImageInputs = append([]ai.ClassifierImageInput(nil), saved...)
			}
			req = jevContextInput(t)
			recordJevContext(ev, u, req, res, err)
		})
	}
}

func TestJevContextConcurrentTenants(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC04-tenants-"+entry)
			u := newJevContextWorld(t, ev, nil)
			for range 2 {
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":0.5}}}`}.script(t, classifierProtocol, ""))
			}
			reqs := []ai.ClassifierRequest{jevContextInput(t), jevContextInput(t)}
			reqs[0].Images = reqs[0].Images[:1]
			reqs[1].Images = reqs[1].Images[1:]
			var wg sync.WaitGroup
			var results [2]ai.ClassifierResult
			var errs [2]error
			for i := range 2 {
				wg.Go(func() {
					scope := textScope([]string{"req-jev-tenant-a", "req-jev-tenant-b"}[i])
					scope.TenantID = []string{tenantA.tenant, tenantB.tenant}[i]
					if entry == "plain" {
						results[i], errs[i] = u.client.Classify(ctxFor(t), scope, jevContextTarget(), reqs[i], jevContextOptions())
					} else {
						results[i], errs[i] = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), scope, jevContextTarget(), reqs[i], jevContextOptions())
					}
				})
			}
			wg.Wait()
			for i, res := range results {
				ev.Check("tenant call succeeds", errs[i] == nil, "got %v", errs[i])
				checkImageInputs(ev, res, reqs[i].Images)
				u.records(ev, res, errs[i])
				ev.Record([]string{"result-a", "result-b"}[i], res)
			}
			ev.Record("inputs", reqs)
			ev.Record("observer-records", u.rec.all())
			recordClassifierAssets(ev, jevContextImages(t))
			u.released(ev)
			captures := u.provider.Requests()
			ev.Check("two requests", len(captures) == 2, "got %d", len(captures))
			for _, capture := range captures {
				var body struct {
					Assets map[string]struct {
						Data string `json:"data_base64"`
					} `json:"assets"`
				}
				mustUnmarshal(t, capture.Body, &body)
				index := 0
				if capture.KeyAlias == tenantB.alias {
					index = 1
				}
				for _, asset := range body.Assets {
					ev.Check("credential and file stay in same tenant", asset.Data == reqs[index].Images[0].Data, "cross-tenant file")
				}
			}
		})
	}
}

package e2e

import (
	"context"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateSnapshot(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC05-snapshot-"+entry)
			req := jevStateInput(t)
			opts := jevContextOptions()
			var mutate func()
			u := newJevStateWorld(t, ev, func(c *ai.Config) {
				previous := c.Bindings
				c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
					mutate()
					return previous.ResolveBinding(ctx, s, id)
				})
			})
			mutate = func() {
				req.ImageReferences[0].Pointer = "/state/missing"
				req.ImageReferences[1].ImageIndex = 100
				req.State[1] = 'X'
				req.Images[0] = req.Images[1]
				opts.ExperimentalImages = false
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			var res ai.ClassifierResult
			var err error
			if entry == "plain" {
				res, err = u.client.Classify(ctxFor(t), textScope("req-state-snapshot-"+entry), jevContextTarget(), req, &opts)
			} else {
				res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-state-snapshot-"+entry), jevContextTarget(), req, &opts)
			}
			ev.Check("State, sources, references and options frozen before resolver", err == nil, "got %v", err)
			req = jevStateInput(t)
			checkImageInputs(ev, res, req.Images)
			recordJevContext(ev, u, req, res, err)
		})
	}
}

func TestJevStateCatalogBoundary(t *testing.T) {
	u := newUnaryWorld(t, nil)
	for _, location := range []ai.ClassifierImageLocation{ai.ClassifierImageLocationState, ai.ClassifierImageLocationInstructions, ai.ClassifierImageLocationCriteria, ai.ClassifierImageLocation("unknown")} {
		t.Run(string(location), func(t *testing.T) {
			ev := run.Case(t, "UC05-catalog-"+string(location))
			catalog := u.client.Catalog()
			model := &catalog.ClassifierModels[0]
			model.ID = "jev-1.14.0"
			model.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
			model.Capabilities.ImageLocations = []ai.ClassifierImageLocation{location}
			model.Capabilities.ImageMIMETypes = []string{"image/png"}
			model.Capabilities.MaxInputImages = 1
			_, hashErr := catalog.Hash()
			_, constructErr := ai.NewClient(ai.Config{Policy: validPolicy(), Catalog: &catalog, Bindings: u.host, Credentials: u.host})
			want := location == ai.ClassifierImageLocationState || location == ai.ClassifierImageLocationInstructions || location == ai.ClassifierImageLocationCriteria
			ev.Check("metadata matches implemented locations", (hashErr == nil) == want && (constructErr == nil) == want, "hash=%v constructor=%v", hashErr, constructErr)
			ev.Record("catalog", catalog)
		})
	}
}

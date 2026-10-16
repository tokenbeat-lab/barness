package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Regression slices for findings from the independent Spec review; written
// before changing the implementation. Source limits and occurrences differ.
func TestJevContextReviewedBudgets(t *testing.T) {
	for _, name := range []string{"source-count", "hook-source-count", "initial-state", "hook-state", "raw-state"} {
		for _, entry := range []string{"plain", "hooked"} {
			if strings.HasPrefix(name, "hook-") || name == "raw-state" {
				if entry == "plain" {
					continue
				}
			}
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-reviewed-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, func(c *ai.Config) {
					if strings.Contains(name, "source-count") {
						c.Catalog.ClassifierModels[0].Capabilities.MaxInputImages = 1
					}
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 1 << 19}
				})
				req := jevContextInput(t)
				req.Images = req.Images[:1]
				req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}}
				var hooks ai.Hooks
				if name == "initial-state" {
					req.State = json.RawMessage(`"` + strings.Repeat("x", 4000) + `"`)
				}
				if name == "hook-source-count" || name == "hook-state" || name == "raw-state" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						state := p.Body["state"].(map[string]any)
						if name == "hook-source-count" {
							state["images"] = append(state["images"].([]any), state["images"].([]any)[0])
						} else {
							state["state"] = strings.Repeat("x", 4094)
							if name == "raw-state" {
								p.Body["state"] = json.RawMessage(mustMarshal(t, state))
							}
						}
						return ai.KeepPayload(), nil
					}
				}
				success := name == "source-count" || name == "hook-source-count"
				if success {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-jev-reviewed-"+name+entry), jevContextTarget(), req, jevContextOptions())
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-jev-reviewed-"+name+entry), jevContextTarget(), req, jevContextOptions())
				}
				if success {
					ev.Check("one model source supports repeated context", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					phase := ai.PhaseRequest
					if name == "initial-state" {
						phase = ai.PhaseScope
					}
					ev.Check("complete native State budget", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}) && len(res.Metadata.Attempts) == 0, "got %v", err)
					if phase == ai.PhaseScope {
						ev.Check("projected budget before credentials", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "read credentials")
					}
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevContextProjectedStateBeforeCopy(t *testing.T) {
	ev := run.Case(t, "UC04-projected-state-before-copy")
	var before, after runtime.MemStats
	u := newJevContextWorld(t, ev, func(c *ai.Config) {
		c.Policy.Classifier.MaxStateBytes = 128
		c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 10000, MaxExpandedImageBytes: 32 << 20}
		previous := c.Bindings
		c.Bindings = bindingResolverFunc(func(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
			b, err := previous.ResolveBinding(ctx, s, id)
			runtime.ReadMemStats(&before)
			return b, err
		})
	})
	req := jevContextInput(t)
	req.Images = req.Images[:1]
	for range 10000 {
		req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
	}
	res, err := u.client.Classify(ctxFor(t), textScope("req-jev-projected-before-copy"), jevContextTarget(), req, jevContextOptions())
	runtime.ReadMemStats(&after)
	ev.Check("projected State rejected before expanding or encoding references", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseScope}) && after.TotalAlloc-before.TotalAlloc < 128<<10, "got %v allocated %d", err, after.TotalAlloc-before.TotalAlloc)
	ev.Record("allocated-after-resolve", after.TotalAlloc-before.TotalAlloc)
	recordJevContext(ev, u, req, res, err)
}

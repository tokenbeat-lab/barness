package e2e

import (
	"context"
	"encoding/json"
	"errors"
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

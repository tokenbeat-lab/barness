package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func newJevStateWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	t.Helper()
	return newJevContextWorld(t, ev, func(c *ai.Config) {
		c.Catalog.Version = "host-jev-1.14-state-candidate"
		c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext, ai.ClassifierImageLocationState}
		if change != nil {
			change(c)
		}
	})
}

func jevStateCall(t *testing.T, u *unaryWorld, entry, id string, req ai.ClassifierRequest, hooks ai.Hooks) (ai.ClassifierResult, error) {
	t.Helper()
	if entry == "plain" {
		return u.client.Classify(ctxFor(t), textScope(id), jevContextTarget(), req, jevContextOptions())
	}
	return u.client.WithHooks(hooks).Classify(ctxFor(t), textScope(id), jevContextTarget(), req, jevContextOptions())
}

func jevStateInput(t *testing.T) ai.ClassifierRequest {
	req := jevContextInput(t)
	req.State = json.RawMessage(`{"state":"kept","images":["business"],"a/b":{"~":[null,null,"<image:ordinary>",{"imageIndex":0}]},"0":null,"~1":null,"":null,"large":9007199254740993,"decimal":1.2300e+04}`)
	req.ImageReferences = []ai.ImageReference{
		{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/state/a~1b/~0/0"},
		{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/0"},
		{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/~01"},
		{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/"},
	}
	return req
}

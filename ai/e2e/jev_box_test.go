package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Start at the public JSON seam so the first red is a behavioral rejection,
// rather than a compiler failure for the new exported variants.
func TestJevBoxPublicJSON(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC08-public-"+entry)
			var req ai.ClassifierRequest
			err := json.Unmarshal([]byte(`{"state":"photo","questions":{"location":{"type":"box","instructions":"locate the mug"},"exists":{"type":"bool","instructions":"is the mug present?"}}}`), &req)
			ev.Check("public Box JSON decodes", err == nil, "got %v", err)
			if err != nil {
				return
			}
			u := newJevInstructionsWorld(t, ev, func(c *ai.Config) {
				c.Catalog.ClassifierModels[0].Capabilities.Kinds = append(c.Catalog.ClassifierModels[0].Capabilities.Kinds, ai.ClassifierQuestionKind("box"))
			})
			req.Images = jevContextImages(t)[:1]
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"location":{"type":"box","box":{"top":0.25,"bottom":0.82,"left":0.44,"right":0.84},"confidence":null},"exists":{"type":"noul","noul":0}}}`}.script(t, classifierProtocol, ""))
			res, err := jevStateCall(t, u, entry, "req-box-public-"+entry, req, ai.Hooks{})
			ev.Check("Box and negative existence return independently", err == nil && len(res.Answers) == 2, "got %v", err)
			ev.Check("normalized Box and explicit null retained", jsonEqual(mustMarshal(t, res.Answers["location"]), []byte(`{"type":"box","box":{"top":0.25,"bottom":0.82,"left":0.44,"right":0.84},"confidence":null}`)), "box changed")
			checkImageInputs(ev, res, req.Images)
			recordJevContext(ev, u, req, res, err)
		})
	}
}

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevInstructionsRemoveContext(t *testing.T) {
	for _, name := range []string{"empty", "missing", "null"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC06-hook-context-removal-"+name)
			u := newJevInstructionsWorld(t, ev, nil)
			req := jevContextInput(t)
			req.Images = req.Images[:1]
			req.Questions["q"] = ai.BoolQuestion{Instructions: json.RawMessage(`["local",null]`)}
			req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/instructions/1"}}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				state := p.Body["state"].(map[string]any)
				switch name {
				case "empty":
					state["images"] = []any{}
				case "missing":
					delete(state, "images")
				case "null":
					state["images"] = nil
				}
				ev.Record("hook-final", p.Body)
				return ai.KeepPayload(), nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-instructions-context-removal-"+name), jevContextTarget(), req, jevContextOptions())
			if name == "empty" {
				ev.Check("removing global occurrences keeps valid local image", err == nil && len(res.Answers) == 1, "got %v", err)
				checkImageInputs(ev, res, req.Images)
			} else {
				ev.Check("builder envelope topology stays explicit", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

package e2e

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextProjection(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC04-projection-"+entry)
			u := newJevContextWorld(t, ev, nil)
			req := jevContextInput(t)
			req.ImageReferences = []ai.ImageReference{{ImageIndex: 1, Placement: ai.ImagePlacementContext}, {ImageIndex: 0, Placement: ai.ImagePlacementContext}, {ImageIndex: 1, Placement: ai.ImagePlacementContext}}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"model":"jev-1.14.0","answers":{"q":{"type":"noul","noul":0.75}},"usage":{"input_tokens":100,"output_tokens":2}}`}.script(t, classifierProtocol, ""))
			var res ai.ClassifierResult
			var err error
			if entry == "plain" {
				res, err = u.client.Classify(ctxFor(t), textScope("req-context-"+entry), jevContextTarget(), req, jevContextOptions())
			} else {
				res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-context-"+entry), jevContextTarget(), req, jevContextOptions())
			}
			ev.Check("classification succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
			checkImageInputs(ev, res, []ai.Image{req.Images[1], req.Images[0]})
			requests := u.provider.Requests()
			if ev.Check("one request and Attempt", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "unexpected requests") {
				var body struct {
					State struct {
						State  json.RawMessage `json:"state"`
						Images []string        `json:"images"`
					} `json:"state"`
					Assets map[string]struct {
						Type, Format string
						Data         string `json:"data_base64"`
					} `json:"assets"`
				}
				mustUnmarshal(t, requests[0].Body, &body)
				ev.Check("original State subtree preserved", jsonEqual(body.State.State, req.State), "state changed")
				if ev.Check("ordered references and source reuse", len(body.State.Images) == 3 && len(body.Assets) == 2, "projection shape") {
					keys := body.State.Images
					ev.Check("repeat uses the same asset", keys[0] == keys[2] && keys[0] != keys[1], "reference reuse")
					ev.Check("native JSON assets", body.Assets[keys[0]].Format == "jpeg" && body.Assets[keys[0]].Data == req.Images[1].Data && body.Assets[keys[1]].Format == "png" && body.Assets[keys[1]].Type == "image", "asset differs")
				}
			}
			ev.Check("usage priced once", res.Usage.Input == 100 && math.Abs(res.Usage.Cost.Input-0.0001) < 1e-12, "got %+v", res.Usage)
			recordJevContext(ev, u, req, res, err)
		})
	}
}

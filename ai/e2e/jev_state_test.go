package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateProjection(t *testing.T) {
	for _, name := range []string{"nested", "root", "mixed", "root-mixed", "state-only"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC05-projection-"+name+"-"+entry)
				u := newJevStateWorld(t, ev, func(c *ai.Config) {
					if name == "state-only" {
						c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
					}
				})
				req := jevStateInput(t)
				root, mixed := strings.HasPrefix(name, "root"), strings.Contains(name, "mixed")
				if root {
					req.State = json.RawMessage(`null`)
					req.Images = req.Images[:1]
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state"}}
				}
				if mixed {
					req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext}, ai.ImageReference{ImageIndex: len(req.Images) - 1, Placement: ai.ImagePlacementContext}, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
				}
				original := string(mustMarshal(t, req))
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":0.75}},"usage":{"input_tokens":100,"output_tokens":2}}`}.script(t, classifierProtocol, ""))
				res, err := jevStateCall(t, u, entry, "req-state-"+name+entry, req, ai.Hooks{})
				ev.Check("classification succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
				ev.Check("borrowed input unchanged", original == string(mustMarshal(t, req)), "input mutated")
				checkImageInputs(ev, res, req.Images)
				requests := u.provider.Requests()
				if ev.Check("single request and Attempt", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "unexpected sends") {
					var body struct {
						State  json.RawMessage `json:"state"`
						Assets map[string]struct {
							Data string `json:"data_base64"`
						} `json:"assets"`
						Questions map[string]json.RawMessage `json:"questions"`
					}
					mustUnmarshal(t, requests[0].Body, &body)
					state := body.State
					if mixed {
						var envelope struct {
							State  json.RawMessage `json:"state"`
							Images []string        `json:"images"`
						}
						mustUnmarshal(t, state, &envelope)
						if ev.Check("ordered context references", len(envelope.Images) == 3, "context lost") {
							ev.Check("context reuse", envelope.Images[0] == envelope.Images[2] && body.Assets[envelope.Images[0]].Data == req.Images[0].Data, "context differs")
						}
						state = envelope.State
					}
					if root {
						var key string
						mustUnmarshal(t, state, &key)
						ev.Check("explicit root image", body.Assets[key].Data == req.Images[0].Data && len(body.Assets) == 1, "root projection")
					} else {
						var obj map[string]json.RawMessage
						mustUnmarshal(t, state, &obj)
						var escaped map[string][]json.RawMessage
						mustUnmarshal(t, obj["a/b"], &escaped)
						var k0, k1, k2, k3 string
						mustUnmarshal(t, obj["0"], &k0)
						mustUnmarshal(t, obj["~1"], &k2)
						mustUnmarshal(t, obj[""], &k3)
						mustUnmarshal(t, escaped["~"][0], &k1)
						ev.Check("only declared slots receive corresponding assets", body.Assets[k0].Data == req.Images[0].Data && body.Assets[k1].Data == req.Images[1].Data && k0 == k2 && k0 == k3 && len(body.Assets) == 2, "incorrect image assignment")
						ev.Check("ordinary values and business fields preserved", string(escaped["~"][1]) == "null" && string(escaped["~"][2]) == `"\u003cimage:ordinary\u003e"` && jsonEqual(escaped["~"][3], []byte(`{"imageIndex":0}`)) && string(obj["state"]) == `"kept"` && string(obj["images"]) == `["business"]`, "business values changed: %s", state)
						ev.Check("exact number spellings", string(obj["large"]) == "9007199254740993" && string(obj["decimal"]) == "1.2300e+04", "numbers changed")
					}
					ev.Check("question names and instructions preserved", len(body.Questions) == 1 && strings.Contains(string(body.Questions["q"]), "图片中有红色吗？"), "question changed")
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevInstructionsProjection(t *testing.T) {
	for _, kind := range []string{"choice", "score", "bool"} {
		for _, shape := range []string{"root", "nested", "ordered", "mixed"} {
			for _, entry := range []string{"plain", "hooked"} {
				t.Run(kind+"/"+shape+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "UC06-projection-"+kind+"-"+shape+"-"+entry)
					u := newJevInstructionsWorld(t, ev, nil)
					req := jevContextInput(t)
					req.Images = req.Images[:1]
					instructions, tail := `null`, ""
					if shape == "nested" {
						instructions, tail = `{"a/b":{"~":[null,null]},"n":9007199254740993,"d":1.2300e+04,"text":"<image:ordinary>"}`, "/a~1b/~0/0"
					}
					if shape == "ordered" || shape == "mixed" {
						instructions, tail = `["before",null,"after",null]`, "/1"
					}
					req.Questions["local/~"] = jevInstructionsQuestion(kind, json.RawMessage(instructions))
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/local~1~0/instructions" + tail}}
					if shape == "mixed" {
						req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
					}
					original := string(mustMarshal(t, req))
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1},"local/~":` + jevInstructionsAnswer(kind) + `}}`}.script(t, classifierProtocol, ""))
					res, err := jevStateCall(t, u, entry, "req-instructions-"+kind+shape+entry, req, ai.Hooks{})
					ev.Check("local image classification succeeds", err == nil && len(res.Answers) == 2, "got %v", err)
					ev.Check("borrowed request unchanged", original == string(mustMarshal(t, req)), "input mutated")
					checkImageInputs(ev, res, req.Images)
					if requests := u.provider.Requests(); ev.Check("one request and Attempt", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "unexpected sends") {
						var body struct {
							State     json.RawMessage `json:"state"`
							Questions map[string]struct {
								Instructions json.RawMessage `json:"instructions"`
							} `json:"questions"`
							Assets map[string]struct {
								Data string `json:"data_base64"`
							} `json:"assets"`
						}
						mustUnmarshal(t, requests[0].Body, &body)
						var key string
						for k, asset := range body.Assets {
							key = k
							ev.Check("source uploaded once", len(body.Assets) == 1 && asset.Data == req.Images[0].Data, "assets changed")
						}
						var expected any
						if shape == "root" {
							expected = key
						} else if shape == "nested" {
							expected = map[string]any{"a/b": map[string]any{"~": []any{key, nil}}, "n": json.Number("9007199254740993"), "d": json.Number("1.2300e+04"), "text": "<image:ordinary>"}
						} else {
							expected = []any{"before", key, "after", nil}
						}
						ev.Check("only explicit instruction slot projected", jsonEqual(body.Questions["local/~"].Instructions, mustMarshal(t, expected)), "instructions changed")
						ev.Check("unrelated question unchanged", jsonEqual(body.Questions["q"].Instructions, []byte(`"图片中有红色吗？"`)), "local image leaked")
						if shape != "mixed" {
							ev.Check("State stays at original position", jsonEqual(body.State, req.State), "image promoted to State")
						}
					}
					recordJevContext(ev, u, req, res, err)
				})
			}
		}
	}
}

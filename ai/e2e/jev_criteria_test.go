package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevCriteriaProjection(t *testing.T) {
	for _, kind := range []string{"choice", "score", "bool"} {
		for _, shape := range []string{"root", "nested", "ordered", "mixed"} {
			for _, entry := range []string{"plain", "hooked"} {
				t.Run(kind+"/"+shape+"/"+entry, func(t *testing.T) {
					ev := run.Case(t, "UC07-projection-"+kind+"-"+shape+"-"+entry)
					u := newJevCriteriaWorld(t, ev, nil)
					req := jevContextInput(t)
					req.Images = req.Images[:1]
					raw, tail := `null`, ""
					if shape == "nested" {
						raw, tail = `{"a/b":{"~":[null,null]},"n":9007199254740993,"d":1.2300e+04,"text":"<image:ordinary>"}`, "/a~1b/~0/0"
					}
					if shape == "ordered" || shape == "mixed" {
						raw, tail = `["before",null,"after",null]`, "/1"
					}
					q, paths, answer := jevCriteriaQuestion(kind, json.RawMessage(raw))
					req.Questions["local/~"] = q
					for _, path := range paths {
						req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/local~1~0/criteria" + path + tail})
					}
					if shape == "mixed" {
						req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
					}
					original := string(mustMarshal(t, req))
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1},"local/~":` + answer + `}}`}.script(t, classifierProtocol, ""))
					res, err := jevStateCall(t, u, entry, "req-criteria-"+kind+shape+entry, req, ai.Hooks{})
					ev.Check("visual criteria succeeds", err == nil && len(res.Answers) == 2, "got %v", err)
					ev.Check("borrowed input unchanged", original == string(mustMarshal(t, req)), "mutated")
					checkImageInputs(ev, res, req.Images)
					if requests := u.provider.Requests(); ev.Check("one request", len(requests) == 1 && len(res.Metadata.Attempts) == 1, "unexpected sends") {
						var body struct {
							State     json.RawMessage
							Questions map[string]struct {
								Instructions json.RawMessage
								Criteria     json.RawMessage
							}
							Assets map[string]struct {
								Data string `json:"data_base64"`
							}
						}
						mustUnmarshal(t, requests[0].Body, &body)
						var key string
						for k, asset := range body.Assets {
							key = k
							ev.Check("one source uploaded", len(body.Assets) == 1 && asset.Data == req.Images[0].Data, "assets differ")
						}
						encoded := string(mustMarshal(t, key))
						expected := strings.Replace(raw, "null", encoded, 1)
						var criterion any
						switch kind {
						case "choice":
							criterion = map[string]json.RawMessage{"only/~": json.RawMessage(expected)}
						case "score":
							criterion = []json.RawMessage{json.RawMessage(expected), json.RawMessage(expected)}
						default:
							criterion = map[string]json.RawMessage{"true": json.RawMessage(expected), "false": json.RawMessage(expected)}
						}
						ev.Check("only declared criterion values change", jsonEqual(body.Questions["local/~"].Criteria, mustMarshal(t, criterion)), "wrong criteria")
						ev.Check("instructions and unrelated question intact", jsonEqual(body.Questions["local/~"].Instructions, []byte(`"judge"`)) && !strings.Contains(string(body.Questions["q"].Instructions), key), "scope leaked")
						if shape != "mixed" {
							ev.Check("State stays in place", jsonEqual(body.State, req.State), "promoted")
						}
					}
					recordJevContext(ev, u, req, res, err)
				})
			}
		}
	}
}

package e2e

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateBudgets(t *testing.T) {
	for _, name := range []string{"shared", "inherited-references", "references", "expanded", "source", "single", "state", "body"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC05-budget-"+name+"-"+entry)
				req := jevStateInput(t)
				req.Questions["second/~"] = ai.ChoiceQuestion{Instructions: json.RawMessage(`"compare"`), Criteria: map[string]json.RawMessage{"candidate/~": json.RawMessage(`null`), "other": json.RawMessage(`"other"`)}}
				u := newJevStateWorld(t, ev, func(c *ai.Config) {
					c.Policy.Classifier.Images.MaxImages = 2
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 4, MaxExpandedImageBytes: 1 << 19}
					switch name {
					case "inherited-references":
						c.Policy.Classifier.Images.References = nil
					case "references":
						c.Policy.Classifier.Images.References.MaxReferences = 3
					case "expanded":
						var total int64
						for _, img := range req.Images {
							data, _ := base64.StdEncoding.DecodeString(img.Data)
							total += int64(len(data))
						}
						c.Policy.Classifier.Images.References.MaxExpandedImageBytes = total
					case "source":
						c.Policy.Classifier.Images.MaxTotalImageBytes = 200
					case "single":
						c.Policy.MaxImageBytes = 200
					case "state":
						c.Policy.Classifier.MaxStateBytes = int64(len(req.State)) + 20
					case "body":
						c.Policy.MaxRequestBytes = 4400
						c.Policy.MaxImageBytes = 4000
						c.Policy.Classifier.Images.MaxTotalImageBytes = 4000
					}
				})
				if name == "shared" {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1},"second/~":{"type":"choice","choice":"candidate/~","probabilities":{"candidate/~":1,"other":0},"confidence":1}}}`}.script(t, classifierProtocol, ""))
				}
				res, err := jevStateCall(t, u, entry, "req-state-budget-"+name+entry, req, ai.Hooks{})
				if name == "shared" {
					ev.Check("shared State counted once across questions", err == nil && len(res.Answers) == 2, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					ev.Check("independent finite budget enforced", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseScope}) && len(res.Metadata.Attempts) == 0 && u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "got %v", err)
					if name == "state" {
						ev.Check("projected State including escaped keys", err != nil && strings.Contains(err.Error(), "MaxStateBytes"), "got %v", err)
					}
				}
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevStateLiteralPaths(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC05-literal-paths-"+entry)
			u := newJevStateWorld(t, ev, nil)
			req := jevContextInput(t)
			req.State = json.RawMessage(`[{"01":null,"%2F":null,"图":null},null,"<image:ordinary>",{"type":"image","data_base64":"ordinary"}]`)
			req.ImageReferences = []ai.ImageReference{
				{ImageIndex: 1, Placement: ai.ImagePlacementValue, Pointer: "/state/0/01"},
				{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/0/%2F"},
				{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/0/图"},
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			res, err := jevStateCall(t, u, entry, "req-state-literals-"+entry, req, ai.Hooks{})
			ev.Check("array root and literal object path segments", err == nil, "got %v", err)
			checkImageInputs(ev, res, []ai.Image{req.Images[0], req.Images[1]}) // % sorts before 0
			if requests := u.provider.Requests(); len(requests) == 1 {
				var body struct {
					State []json.RawMessage `json:"state"`
				}
				mustUnmarshal(t, requests[0].Body, &body)
				ev.Check("template objects and unreferenced null remain ordinary", len(body.State) == 4 && string(body.State[1]) == "null" && jsonEqual(body.State[3], []byte(`{"type":"image","data_base64":"ordinary"}`)), "content changed")
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

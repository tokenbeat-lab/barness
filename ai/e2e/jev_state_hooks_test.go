package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevStateFinalHooks(t *testing.T) {
	for _, name := range []string{"replace", "move", "root", "array", "raw", "mixed", "remove-all", "from-text", "business-envelope", "business-null-state", "unused", "substring", "object-key", "question-key", "instruction", "criteria", "location", "references", "expanded", "state-bytes", "body-bytes"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC05-hook-"+name)
			u := newJevStateWorld(t, ev, func(c *ai.Config) {
				c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 1 << 19}
				if name == "from-text" {
					c.Catalog.ClassifierModels[0].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
				}
				if name == "references" {
					c.Policy.Classifier.Images.References.MaxReferences = 4
				}
				if name == "expanded" {
					c.Policy.Classifier.Images.References.MaxExpandedImageBytes = 6000
				}
				if name == "body-bytes" {
					c.Policy.MaxRequestBytes = 6000
					c.Policy.MaxImageBytes = 6000
					c.Policy.Classifier.Images.MaxTotalImageBytes = 6000
				}
			})
			req := jevStateInput(t)
			if name == "mixed" {
				req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
			}
			if name == "from-text" {
				req.Images, req.ImageReferences = nil, nil
			}
			images := jevContextImages(t)
			want := []ai.Image{images[1], images[0]}
			success := name == "replace" || name == "move" || name == "root" || name == "array" || name == "raw" || name == "mixed" || name == "remove-all" || name == "from-text" || name == "business-envelope" || name == "business-null-state"
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":0.5}}}`}.script(t, classifierProtocol, ""))
			var borrowed map[string]any
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				assets := map[string]any{"image-A": map[string]any{"type": "image", "format": "png", "data_base64": images[0].Data}, "image-B": map[string]any{"format": "jpg", "data_base64": images[1].Data}}
				state := any(map[string]any{"z": "image-A", "a": []any{"image-B", "image-A"}, "ordinary": nil})
				p.Body["assets"] = assets
				switch name {
				case "replace":
					assets["image-B"] = assets["image-A"]
					want = []ai.Image{images[0], images[0]}
				case "move":
					state = map[string]any{"z": "image-B", "a": "image-A"}
					want = images
				case "root", "from-text":
					state = "image-B"
					delete(assets, "image-A")
					want = images[1:]
				case "array":
					state = []any{"image-B", nil, "image-A"}
				case "mixed":
					state = map[string]any{"state": state, "images": []any{"image-B", "image-A", "image-B"}}
				case "remove-all":
					state = map[string]any{"ordinary": nil}
					delete(p.Body, "assets")
					want = nil
				case "business-envelope":
					state = map[string]any{"state": "image-B", "images": []any{"image-A"}}
					want = images
				case "business-null-state":
					state = map[string]any{"state": nil, "images": []any{"image-A", "image-B"}}
					want = images
				case "unused":
					assets["unused-image"] = assets["image-A"]
				case "substring":
					state = []any{"before image-A after", "image-B"}
				case "object-key":
					state = map[string]any{"nested-image-A": "image-B", "photo": "image-A"}
				case "question-key":
					p.Body["questions"] = map[string]any{"image-A": map[string]any{"type": "noul", "instructions": "ask"}}
				case "instruction":
					p.Body["questions"] = map[string]any{"q": map[string]any{"type": "noul", "instructions": "image-A"}}
				case "criteria":
					p.Body["questions"] = map[string]any{"q": map[string]any{"type": "score", "instructions": "ask", "criteria": []any{"image-A", "other"}}}
				case "location":
					state = map[string]any{"state": "image-B", "images": []any{"image-A"}}
					p.Body["questions"] = map[string]any{"q": map[string]any{"type": "noul", "instructions": []any{"ask", "image-A"}}}
				case "references", "expanded":
					state = []any{"image-A", "image-B", "image-B", "image-B", "image-B"}
				case "state-bytes":
					state = []any{"image-A", "image-B", strings.Repeat("x", 4096)}
				case "body-bytes":
					assets["image-C"] = assets["image-B"]
					state = []any{"image-A", "image-B", "image-C"}
				}
				p.Body["state"] = state
				if name == "raw" {
					p.Body["state"] = json.RawMessage(mustMarshal(t, state))
					p.Body["assets"] = json.RawMessage(mustMarshal(t, assets))
				}
				ev.Record("hook-input", p.Body)
				borrowed = p.Body
				return ai.KeepPayload(), nil
			}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				borrowed["state"], borrowed["assets"] = "changed after freeze", nil
				return nil
			}}
			res, err := jevStateCall(t, u, "hooked", "req-state-hook-"+name, req, hooks)
			if success {
				ev.Check("final native change succeeds", err == nil, "got %v", err)
				checkImageInputs(ev, res, want)
			} else {
				code := ai.CodeCallbackFailed
				if name == "references" || name == "expanded" || name == "state-bytes" || name == "body-bytes" {
					code = ai.CodeResourceLimit
				}
				ev.Check("invalid final usage rejected before Attempt", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
				ev.Check("no partial image result", res.Answers == nil && res.ImageInputs == nil, "partial result")
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

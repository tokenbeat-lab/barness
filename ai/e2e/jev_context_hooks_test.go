package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextFinalHooks(t *testing.T) {
	for _, name := range []string{"replace", "reorder", "add", "remove", "same-file", "raw", "unused", "part", "url", "data-url", "substring", "overlap", "object-key", "question-key", "instruction", "criteria", "nested-state", "empty-key", "long-key", "null-assets", "empty-assets", "asset-case", "asset-duplicate", "asset-missing-type", "asset-unknown", "asset-format", "apng", "small", "over-count", "over-references", "over-expanded", "unauthorized", "unselected", "model", "authority"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC04-hook-"+name)
			u := newJevContextWorld(t, ev, func(c *ai.Config) {
				switch name {
				case "over-count":
					c.Policy.Classifier.Images.MaxImages = 2
				case "over-references":
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 2, MaxExpandedImageBytes: 1 << 19}
				case "over-expanded":
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 4000}
				}
			})
			req := jevContextInput(t)
			opts := jevContextOptions()
			want := []ai.Image{req.Images[0], req.Images[1]}
			if name == "unauthorized" || name == "unselected" {
				req.Images = nil
				if name == "unauthorized" {
					opts = ai.TypeSafeOptions{}
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				} else {
					opts = ai.TypeSafeOptions{}
				}
			}
			success := name == "replace" || name == "reorder" || name == "add" || name == "remove" || name == "same-file" || name == "raw"
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":0.5}}}`}.script(t, classifierProtocol, ""))
			hook := func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				// Fix understandable Hook keys independently of random adapter keys.
				images := jevContextImages(t)
				assets := map[string]any{"image-A": map[string]any{"type": "image", "format": "png", "data_base64": images[0].Data}, "image-B": map[string]any{"type": "image", "format": "jpeg", "data_base64": images[1].Data}}
				order := []any{"image-A", "image-B"}
				state := map[string]any{"state": json.RawMessage(`{"original":9007199254740993}`), "images": order}
				p.Body["assets"], p.Body["state"] = assets, state
				switch name {
				case "replace":
					assets["image-A"] = assets["image-B"]
					want = []ai.Image{images[1], images[1]}
				case "reorder":
					order[0], order[1] = order[1], order[0]
					want = []ai.Image{images[1], images[0]}
				case "add", "over-count":
					assets["image-C"] = assets["image-A"]
					state["images"] = append(order, "image-C")
					want = append(want, images[0])
				case "remove":
					delete(assets, "image-A")
					state["images"] = []any{"image-B"}
					want = []ai.Image{images[1]}
				case "same-file":
					assets["image-B"] = assets["image-A"]
					want = []ai.Image{images[0], images[0]}
				case "raw":
					p.Body["assets"] = json.RawMessage(mustMarshal(t, assets))
					p.Body["state"] = json.RawMessage(mustMarshal(t, state))
				case "unused":
					assets["unused-image"] = assets["image-A"]
				case "part":
					assets["image-A"].(map[string]any)["part"] = "image-file"
				case "url":
					assets["image-A"] = map[string]any{"type": "image", "url": "https://example.invalid/image.png"}
				case "data-url":
					assets["image-A"].(map[string]any)["data_base64"] = "data:image/png;base64," + images[0].Data
				case "substring":
					order[0] = "before image-A after"
				case "overlap":
					assets["A"] = assets["image-A"]
					state["images"] = append(order, "A")
				case "object-key":
					state["state"] = map[string]any{"nested-image-A": true}
				case "question-key":
					p.Body["questions"] = map[string]any{"image-A": map[string]any{"type": "noul", "instructions": "Ask"}}
				case "instruction":
					p.Body["questions"] = map[string]any{"q": map[string]any{"type": "noul", "instructions": "image-A"}}
				case "criteria":
					p.Body["questions"] = map[string]any{"q": map[string]any{"type": "score", "instructions": "Ask", "criteria": []any{"image-A", "text"}}}
				case "nested-state":
					state["state"] = map[string]any{"nested": "image-A"}
				case "empty-key":
					assets[" "] = assets["image-A"]
					delete(assets, "image-A")
					order[0] = " "
				case "long-key":
					key := strings.Repeat("图", 342)
					assets[key] = assets["image-A"]
					delete(assets, "image-A")
					order[0] = key
				case "null-assets":
					p.Body["assets"] = nil
				case "empty-assets":
					p.Body["assets"] = map[string]any{}
				case "asset-case":
					a := assets["image-A"].(map[string]any)
					a["Type"] = a["type"]
					delete(a, "type")
				case "asset-duplicate":
					assets["image-A"] = json.RawMessage(`{"type":"image","type":"image","format":"png","data_base64":"` + images[0].Data + `"}`)
				case "asset-missing-type":
					delete(assets["image-A"].(map[string]any), "type")
				case "asset-unknown":
					assets["image-A"].(map[string]any)["mime"] = "image/png"
				case "asset-format":
					assets["image-A"].(map[string]any)["format"] = "PNG"
				case "apng":
					assets["image-A"].(map[string]any)["data_base64"] = jevAPNG(t).Data
				case "small":
					assets["image-A"].(map[string]any)["data_base64"] = jevPNG(t, 256, 256).Data
				case "over-references", "over-expanded":
					state["images"] = []any{"image-A", "image-B", "image-B"}
				case "model":
					p.Body["model"] = "jev-latest"
				case "authority":
					p.Body["endpoint"] = "https://example.invalid"
				}
				ev.Record("hook-input", p.Body)
				return ai.KeepPayload(), nil
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: hook}).Classify(ctxFor(t), textScope("req-jev-hook-"+name), jevContextTarget(), req, opts)
			if success {
				ev.Check("legal final native mutation", err == nil, "got %v", err)
				checkImageInputs(ev, res, want)
			} else {
				code := ai.CodeCallbackFailed
				switch name {
				case "unauthorized", "unselected", "model", "authority":
					code = ai.CodeTenantDenied
				case "over-count", "over-references", "over-expanded", "long-key":
					code = ai.CodeResourceLimit
				}
				ev.Check("final native rejection before Attempt", errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %v", err)
				ev.Check("atomic rejection", res.Answers == nil && res.ImageInputs == nil, "partial result")
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesDecodedBoundary(t *testing.T) {
	img := classifierAssets(t)[0]
	raw, _ := base64.StdEncoding.DecodeString(img.Data)
	for _, name := range []string{"single-at", "single-over", "total-at", "total-over"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD19-images-decoded-"+name)
			limit := int64(len(raw))
			count := 1
			if strings.HasPrefix(name, "total") {
				count = 2
				limit *= 2
			}
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				if strings.HasPrefix(name, "single") {
					c.Policy.MaxImageBytes = limit
					if strings.HasSuffix(name, "over") {
						c.Policy.MaxImageBytes--
					}
				} else {
					c.Policy.Classifier.Images.MaxTotalImageBytes = limit
					if strings.HasSuffix(name, "over") {
						c.Policy.Classifier.Images.MaxTotalImageBytes--
					}
				}
			})
			req := decisionsInput()
			for i := 0; i < count; i++ {
				req.Images = append(req.Images, img)
			}
			if strings.HasSuffix(name, "at") {
				enqueueClassifierImages(t, ev, u)
			}
			res, err := u.client.Classify(ctxFor(t), textScope("req-image-bound-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
			ev.Record("bytes", map[string]any{"singleDecoded": len(raw), "totalDecoded": len(raw) * count, "base64": len(img.Data) * count, "budget": limit})
			if strings.HasSuffix(name, "at") {
				ev.Check("exact decoded boundary accepted", err == nil, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseScope, ai.StopReasonError)
				ev.Check("before credentials", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "credential read")
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func TestDecisionsImagesHookCapacity(t *testing.T) {
	img := classifierAssets(t)[0]
	for _, name := range []string{"128", "129", "model", "host", "escaped-raw", "deep"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD26-images-hook-capacity-"+name)
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				if name == "model" {
					c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.MaxInputImages = 1
				}
				if name == "host" {
					c.Policy.Classifier.Images.MaxImages = 1
				}
			})
			count := 129
			if name == "128" {
				count = 128
			}
			if name == "model" || name == "host" {
				count = 2
			}
			images := make([]ai.Image, count)
			for i := range images {
				images[i] = img
			}
			if name == "128" {
				enqueueClassifierImages(t, ev, u)
			}
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["input"] = classifierImageMessages(images)
				if name == "escaped-raw" {
					// JSON escaping must not conceal a large decoded image.
					p.Body["input"] = json.RawMessage(`[{"role":"user","content":[{"type":"input_text","text":"ok"},{"type":"input_image","image_url":"data:image/png;base64,` + strings.Repeat(`\u0041`, 800000) + `"}]}]`)
				}
				if name == "deep" {
					var v any = "x"
					for i := 0; i < 12000; i++ {
						v = []any{v}
					}
					p.Body["input"] = v
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-image-capacity-"+name), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			if name == "128" {
				ev.Check("hook may send bounded 128 images", err == nil, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseRequest, ai.StopReasonError)
				ev.Check("capacity never attempts", len(res.Metadata.Attempts) == 0, "sent request")
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

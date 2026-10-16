package e2e

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestJevContextNativeBudgets(t *testing.T) {
	for _, name := range []string{"shared-upload", "file-limit", "body-limit", "key-total"} {
		for _, entry := range []string{"plain", "hooked"} {
			if name == "key-total" && entry == "plain" {
				continue
			}
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-native-budget-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, func(c *ai.Config) {
					c.Policy.MaxRequestBytes = 40 << 20
					c.Policy.MaxImageBytes = 32 << 20
					c.Policy.Classifier.Images.MaxTotalImageBytes = 32 << 20
					c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 128, MaxExpandedImageBytes: 40 << 20}
					if name == "key-total" {
						c.Policy.Classifier.Images.MaxImages = 128
						c.Catalog.ClassifierModels[0].Capabilities.MaxInputImages = 128
					}
					if name == "shared-upload" {
						c.Policy.MaxRequestBytes = 6000
						c.Policy.MaxImageBytes = 6000
						c.Policy.Classifier.Images.MaxTotalImageBytes = 6000
					}
				})
				req := jevContextInput(t)
				var hooks ai.Hooks
				switch name {
				case "shared-upload":
					req.Images = req.Images[1:]
					for range 8 {
						req.ImageReferences = append(req.ImageReferences, ai.ImageReference{ImageIndex: 0, Placement: ai.ImagePlacementContext})
					}
				case "file-limit", "body-limit":
					bytes := 16<<20 + 1
					if name == "body-limit" {
						bytes = 13 << 20
					}
					raw, _ := base64.StdEncoding.DecodeString(req.Images[1].Data)
					raw = append(raw, make([]byte, bytes-len(raw))...)
					req.Images = []ai.Image{{Data: base64.StdEncoding.EncodeToString(raw), MimeType: "image/jpeg"}}
					if name == "body-limit" {
						req.Images = append(req.Images, req.Images[0])
					}
				case "key-total":
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						assets := map[string]any{}
						keys := []string{}
						for i := range 66 {
							key := fmt.Sprintf("%03d", i) + strings.Repeat("x", 997)
							keys = append(keys, key)
							assets[key] = map[string]any{"type": "image", "format": "png", "data_base64": req.Images[0].Data}
						}
						p.Body["assets"], p.Body["state"] = assets, map[string]any{"state": "ok", "images": keys}
						return ai.KeepPayload(), nil
					}
				}
				if name == "shared-upload" {
					enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
				}
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), textScope("req-jev-native-budget-"+name+entry), jevContextTarget(), req, jevContextOptions())
				} else {
					res, err = u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-jev-native-budget-"+name+entry), jevContextTarget(), req, jevContextOptions())
				}
				if name == "shared-upload" {
					ev.Check("one uploaded asset fits independent expansion budget", err == nil, "got %v", err)
					checkImageInputs(ev, res, req.Images)
				} else {
					phase := ai.PhaseCapability
					if name == "key-total" {
						phase = ai.PhaseRequest
					}
					ev.Check("native capacity rejected before Attempt", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}) && len(res.Metadata.Attempts) == 0, "got %v", err)
					if name == "key-total" {
						ev.Check("total key byte bound exercised", strings.Contains(err.Error(), "MaxTotalAssetKeyBytes"), "got %v", err)
					}
					if phase == ai.PhaseCapability {
						ev.Check("capacity before credentials", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "credential read")
					}
				}
				// Large fixtures are deterministic synthetic JPEGs with zero padding;
				// record their exact bytes/hash through the common evidence recorder.
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

package e2e

import (
	"encoding/base64"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesInput(t *testing.T) {
	assets := classifierAssets(t)
	for _, name := range []string{"three-formats", "128", "129", "host-count", "nil-policy", "nil-classifier", "generate-policy", "text-only", "model-count", "model-mime", "binding-denied", "jev", "bad-mime", "magic", "alphabet", "padding", "tail-bits", "crlf", "url", "path", "state-missing", "state-null", "per-image", "total-image", "total-boundary", "known-request", "known-size", "state-order"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD19-images-input-"+name)
			req := decisionsInput()
			req.Images = append([]ai.Image(nil), assets...)
			code, phase, success := ai.CodeInvalidRequest, ai.PhaseScope, false
			u := newClassifierImagesWorld(t, func(c *ai.Config) {
				m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
				switch name {
				case "three-formats":
					success = true
				case "128", "129", "host-count":
					count := 128
					if name != "128" {
						count = 129
					}
					req.Images = make([]ai.Image, count)
					for i := range req.Images {
						req.Images[i] = assets[0]
					}
					success = name == "128"
					phase = ai.PhaseCapability
					if name == "host-count" {
						c.Policy.Classifier.Images.MaxImages = 128
						code, phase = ai.CodeResourceLimit, ai.PhaseScope
					}
				case "nil-policy", "generate-policy":
					c.Policy.Classifier.Images = nil
					if name == "generate-policy" {
						c.Policy.Image = &ai.ImagePolicy{MaxInputImages: 16, MaxOutputImages: 1, MaxOutputImageBytes: 1024, MaxTotalOutputImageBytes: 1024}
					}
				case "nil-classifier":
					c.Policy.Classifier = nil
				case "text-only":
					m.Input = []ai.Modality{ai.ModalityText}
					m.Capabilities.MaxInputImages = 0
					m.Capabilities.ImageMIMETypes = nil
					phase = ai.PhaseCapability
				case "model-count":
					m.Capabilities.MaxInputImages = 2
					phase = ai.PhaseCapability
				case "model-mime":
					m.Capabilities.ImageMIMETypes = []string{"image/png"}
					phase = ai.PhaseCapability
				case "binding-denied":
					code, phase = ai.CodeTenantDenied, ai.PhaseCapability
				case "jev":
					phase = ai.PhaseCapability
				case "bad-mime":
					req.Images[0].MimeType = "image/gif"
				case "magic":
					req.Images[0].MimeType = "image/jpeg"
				case "alphabet":
					req.Images[0].Data += "!AAA"
				case "padding":
					req.Images[0].Data += "="
				case "tail-bits":
					req.Images[0].Data = base64.StdEncoding.EncodeToString([]byte("\x89PNG\r\n\x1a\n"))
					req.Images[0].Data = req.Images[0].Data[:len(req.Images[0].Data)-2] + "p="
				case "crlf":
					req.Images[0].Data += "\r\n"
				case "url":
					req.Images[0].Data = "https://foreign.invalid/x.png"
				case "path":
					req.Images[0].Data = "/tmp/image.png"
				case "state-missing":
					req.State = nil
				case "state-null":
					req.State = json.RawMessage(`null`)
				case "per-image":
					c.Policy.MaxImageBytes = 32
					code = ai.CodeResourceLimit
				case "total-image", "total-boundary":
					total := int64(0)
					for _, img := range req.Images {
						raw, _ := base64.StdEncoding.DecodeString(img.Data)
						total += int64(len(raw))
					}
					c.Policy.Classifier.Images.MaxTotalImageBytes = total
					if name == "total-image" {
						c.Policy.Classifier.Images.MaxTotalImageBytes--
						code = ai.CodeResourceLimit
					} else {
						success = true
					}
				case "known-request":
					c.Policy.MaxRequestBytes = 600
					c.Policy.MaxImageBytes = 600
					c.Policy.Classifier.Images.MaxTotalImageBytes = 600
					req.Images = []ai.Image{assets[1]}
					code = ai.CodeResourceLimit
				case "known-size":
					req.Images = []ai.Image{{Data: strings.Repeat("A", 16<<20), MimeType: "image/png"}}
					code = ai.CodeResourceLimit
				case "state-order":
					req.State = json.RawMessage(`null`)
					req.Images = []ai.Image{{Data: strings.Repeat("A", 16<<20), MimeType: "image/png"}}
					code = ai.CodeResourceLimit
				}
			})
			target := ai.Target{BindingID: "classifier", ModelID: "host-decisions"}
			if name == "binding-denied" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
			}
			if name == "jev" {
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) {
					b.ProviderID = ai.ProviderTypeSafe
					b.API = ai.APITypeSafeSystemOne
					b.AllowedModels = []string{"jev-1.13.0"}
				})
				target.ModelID = "jev-1.13.0"
			}
			if success {
				enqueueClassifierImages(t, ev, u)
			}
			if name != "known-size" && name != "state-order" {
				ev.Fixture("public-request.json", mustMarshal(t, req))
			}
			encoded, decoded, known := int64(0), int64(0), int64(len(req.State))
			for _, img := range req.Images {
				encoded += int64(len(img.Data))
				known += int64(len(img.Data) + len(img.MimeType) + len("data:;base64,"))
				if name == "known-size" || name == "state-order" {
					continue
				}
				if raw, err := base64.StdEncoding.Strict().DecodeString(img.Data); err == nil {
					decoded += int64(len(raw))
				}
			}
			ev.Record("catalog", u.client.Catalog())
			ev.Record("binding", u.host.Binding(tenantA.tenant, "classifier"))
			ev.Record("input-measurements", map[string]any{"count": len(req.Images), "base64Bytes": encoded, "decodedBytes": decoded, "stateAndImageKnownBytes": known})
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := u.client.Classify(ctxFor(t), textScope("req-images-"+name), target, req, nil)
			runtime.ReadMemStats(&after)
			if success {
				ev.Check("images single-choice succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
				requests := u.provider.Requests()
				if ev.Check("one request", len(requests) == 1, "got %d", len(requests)) {
					var body map[string]any
					mustUnmarshal(t, requests[0].Body, &body)
					ev.Check("text-first ordered inline images", jsonEqual(mustMarshal(t, body["input"]), mustMarshal(t, classifierImageMessages(req.Images))), "incorrect image wire")
					ev.Record("whole-json-bytes", len(requests[0].Body))
				}
				if name == "three-formats" {
					recordClassifierAssets(ev, assets)
				}
			} else {
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				ev.Check("pre-network denial", u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0, "got %+v", res.Metadata)
			}
			if name == "known-size" || name == "state-order" {
				ev.Record("allocation-bytes", after.TotalAlloc-before.TotalAlloc)
				ev.Check("known oversize before proportional allocation", after.TotalAlloc-before.TotalAlloc < 1<<20, "allocated %d", after.TotalAlloc-before.TotalAlloc)
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

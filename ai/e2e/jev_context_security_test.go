package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func jevPNG(t *testing.T, w, h int) ai.Image {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return ai.Image{Data: base64.StdEncoding.EncodeToString(buf.Bytes()), MimeType: "image/png"}
}

func jevAPNG(t *testing.T) ai.Image {
	img := jevPNG(t, 280, 280)
	raw, _ := base64.StdEncoding.DecodeString(img.Data)
	chunk := make([]byte, 20)
	binary.BigEndian.PutUint32(chunk, 8)
	copy(chunk[4:], "acTL")
	binary.BigEndian.PutUint32(chunk[8:], 1)
	binary.BigEndian.PutUint32(chunk[16:], crc32.ChecksumIEEE(chunk[4:16]))
	raw = append(append(append([]byte(nil), raw[:33]...), chunk...), raw[33:]...)
	img.Data = base64.StdEncoding.EncodeToString(raw)
	return img
}

func TestJevContextPreflight(t *testing.T) {
	for _, name := range []string{"no-policy", "no-binding", "no-experiment", "no-capability", "too-many", "mime", "local", "apng", "small-area", "small-edge", "large-edge", "bad-base64", "mismatch", "host-image", "host-total", "host-references", "host-expanded", "host-body"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "UC04-preflight-"+name+"-"+entry)
				u := newJevContextWorld(t, ev, func(c *ai.Config) {
					switch name {
					case "no-policy":
						c.Policy.Classifier.Images = nil
					case "no-capability":
						m := &c.Catalog.ClassifierModels[0]
						m.Input = []ai.Modality{ai.ModalityText}
						m.Capabilities.ImageLocations = nil
						m.Capabilities.ImageMIMETypes = nil
						m.Capabilities.MaxInputImages = 0
					case "too-many":
						c.Catalog.ClassifierModels[0].Capabilities.MaxInputImages = 1
					case "host-image":
						c.Policy.MaxImageBytes = 32
					case "host-total":
						c.Policy.Classifier.Images.MaxTotalImageBytes = 32
					case "host-references":
						c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 1, MaxExpandedImageBytes: 1 << 19}
					case "host-expanded":
						c.Policy.Classifier.Images.References = &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 32}
					case "host-body":
						c.Policy.MaxRequestBytes = 5000
						c.Policy.MaxImageBytes = 4096
						c.Policy.Classifier.Images.MaxTotalImageBytes = 4096
					}
				})
				req := jevContextInput(t)
				opts := jevContextOptions()
				code, phase := ai.CodeInvalidRequest, ai.PhaseCapability
				switch name {
				case "no-policy":
					phase = ai.PhaseScope
				case "no-binding":
					code = ai.CodeTenantDenied
					u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = false })
				case "no-experiment":
					opts = ai.TypeSafeOptions{}
				case "mime":
					req.Images[0] = classifierAssets(t)[2]
				case "local":
					req.State = json.RawMessage(`{"photo":null}`)
					req.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state/photo"}, {ImageIndex: 1, Placement: ai.ImagePlacementContext}}
				case "apng":
					req.Images[0] = jevAPNG(t)
				case "small-area":
					req.Images[0] = jevPNG(t, 256, 256)
				case "small-edge":
					req.Images[0] = jevPNG(t, 255, 400)
				case "large-edge":
					req.Images[0] = jevPNG(t, 1569, 280)
					code = ai.CodeResourceLimit
				case "bad-base64":
					req.Images[0].Data = "aGVsbG8="
					phase = ai.PhaseScope
				case "mismatch":
					req.Images[0].MimeType = "image/jpeg"
					phase = ai.PhaseScope
				case "host-image", "host-total", "host-references", "host-expanded":
					code, phase = ai.CodeResourceLimit, ai.PhaseScope
				case "host-body":
					req.State = json.RawMessage(`"` + strings.Repeat("x", 4000) + `"`)
					code, phase = ai.CodeResourceLimit, ai.PhaseScope
				}
				scope := textScope("req-jev-preflight-" + name + entry)
				var res ai.ClassifierResult
				var err error
				if entry == "plain" {
					res, err = u.client.Classify(ctxFor(t), scope, jevContextTarget(), req, opts)
				} else {
					res, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), scope, jevContextTarget(), req, opts)
				}
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				reads := u.host.ReadsFor(scope.RequestID)
				ev.Check("rejected before credentials and Attempt", reads.Credentials == 0 && len(res.Metadata.Attempts) == 0 && len(u.provider.Requests()) == 0 && len(u.adm.Requests()) == 0, "got %v / %+v", err, reads)
				recordJevContext(ev, u, req, res, err)
			})
		}
	}
}

func TestJevContextHeaders(t *testing.T) {
	for _, name := range []string{"unchanged", "lowercase", "duplicate", "contradictory", "remove", "long-context", "unselected"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC04-headers-"+name)
			u := newJevContextWorld(t, ev, nil)
			req := jevContextInput(t)
			opts := jevContextOptions()
			if name == "unselected" {
				req.Images = nil
				opts = ai.TypeSafeOptions{}
			}
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":{"q":{"type":"noul","noul":1}}}`}.script(t, classifierProtocol, ""))
			hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
				switch name {
				case "lowercase":
					h["typesafe-danger-zone-images"] = h["Typesafe-Danger-Zone-Images"]
					delete(h, "Typesafe-Danger-Zone-Images")
				case "duplicate":
					h["typesafe-danger-zone-images"] = []string{"true"}
				case "contradictory":
					h.Set("Typesafe-Danger-Zone-Images", "false")
				case "remove":
					h.Del("Typesafe-Danger-Zone-Images")
				case "long-context":
					h.Set("TypeSafe-Danger-Zone-Long-Context", "true")
				case "unselected":
					h.Set("TypeSafe-Danger-Zone-Images", "true")
				}
				return nil
			}}
			res, err := u.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-jev-header-"+name), jevContextTarget(), req, opts)
			if name == "unchanged" || name == "lowercase" {
				ev.Check("selected header accepted", err == nil, "got %v", err)
			} else {
				ev.Check("experiment authority pinned", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseRequest}) && len(res.Metadata.Attempts) == 0, "got %v", err)
			}
			recordJevContext(ev, u, req, res, err)
		})
	}
}

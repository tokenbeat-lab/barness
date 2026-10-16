package e2e

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func jevContextOptions() ai.TypeSafeOptions {
	return ai.TypeSafeOptions{ExperimentalImages: true}
}

func jevContextImages(t *testing.T) []ai.Image {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 280, 280))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var p, j bytes.Buffer
	if err := png.Encode(&p, img); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&j, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return []ai.Image{{Data: base64.StdEncoding.EncodeToString(p.Bytes()), MimeType: "image/png"}, {Data: base64.StdEncoding.EncodeToString(j.Bytes()), MimeType: "image/jpeg"}}
}

func newJevContextWorld(t *testing.T, ev *evidence.Case, change func(*ai.Config)) *unaryWorld {
	t.Helper()
	u := newUnaryWorld(t, func(c *ai.Config) {
		c.Catalog.Version = "host-jev-1.14-context-candidate"
		m := &c.Catalog.ClassifierModels[0]
		m.ID = "jev-1.14.0"
		m.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
		m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
		m.Capabilities.ImageMIMETypes = []string{"image/png", "image/jpeg"}
		m.Capabilities.MaxInputImages = 8
		c.Policy.Classifier.Images = &ai.ClassifierImagePolicy{MaxImages: 8, MaxTotalImageBytes: 1 << 19}
		if change != nil {
			change(c)
		}
		ev.Record("config", map[string]any{"catalog": c.Catalog, "policy": c.Policy, "realService": "NOT_RUN"})
	})
	for _, tenant := range []tenantKey{tenantA, tenantB} {
		u.updateBindingOf(tenant, "classifier", func(b *ai.Binding) { b.AllowedModels = []string{"jev-1.14.0"}; b.AllowClassifierImages = true })
	}
	return u
}

func jevContextInput(t *testing.T) ai.ClassifierRequest {
	return ai.ClassifierRequest{State: json.RawMessage(`{"state":"kept","images":["business"],"id":9007199254740993}`), Questions: map[string]ai.ClassifierQuestion{"q": ai.BoolQuestion{Instructions: json.RawMessage(`"图片中有红色吗？"`)}}, Images: jevContextImages(t)}
}

func jevContextTarget() ai.Target { return ai.Target{BindingID: "classifier", ModelID: "jev-1.14.0"} }

func recordJevContext(ev *evidence.Case, u *unaryWorld, req ai.ClassifierRequest, res ai.ClassifierResult, err error) {
	ev.Record("input", req)
	ev.Record("binding", u.host.Binding(res.Metadata.TenantID, "classifier"))
	recordClassifierAssets(ev, req.Images)
	recordImageInputsCall(ev, u, res, err)
	ev.Record("host-reads", u.host.ReadsFor(res.Metadata.RequestID))
}

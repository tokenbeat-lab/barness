package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func classifierAssets(t *testing.T) []ai.Image {
	t.Helper()
	var pngBytes, jpegBytes bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&pngBytes, img); err != nil {
		t.Fatal(err)
	}
	img.Set(0, 0, color.RGBA{B: 255, A: 255})
	if err := jpeg.Encode(&jpegBytes, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	webp, err := os.ReadFile("testdata/decisions/images/gopher.webp")
	if err != nil {
		t.Fatal(err)
	}
	return []ai.Image{{Data: base64.StdEncoding.EncodeToString(pngBytes.Bytes()), MimeType: "image/png"}, {Data: base64.StdEncoding.EncodeToString(jpegBytes.Bytes()), MimeType: "image/jpeg"}, {Data: base64.StdEncoding.EncodeToString(webp), MimeType: "image/webp"}}
}

func recordClassifierAssets(ev *evidence.Case, assets []ai.Image) {
	for i, img := range assets {
		raw, err := base64.StdEncoding.Strict().DecodeString(img.Data)
		if err != nil {
			ev.T().Fatal(err)
		}
		sum := sha256.Sum256(raw)
		ev.Fixture([]string{"red.png", "blue.jpg", "gopher.webp"}[i], raw)
		ev.Record([]string{"png-size", "jpeg-size", "webp-size"}[i], map[string]any{"mime": img.MimeType, "sha256": hex.EncodeToString(sum[:]), "decoded": len(raw), "base64": len(img.Data), "dataURL": len(img.Data) + len(img.MimeType) + 13})
	}
}

func newClassifierImagesWorld(t *testing.T, configure func(*ai.Config)) *unaryWorld {
	t.Helper()
	u := newDecisionsWorld(t, func(c *ai.Config) {
		c.Policy.Classifier.Images = &ai.ClassifierImagePolicy{MaxImages: 256, MaxTotalImageBytes: 1 << 19}
		m := &c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1]
		m.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
		m.Capabilities.MaxInputImages = 128
		m.Capabilities.ImageMIMETypes = []string{"image/png", "image/jpeg", "image/webp"}
		if configure != nil {
			configure(c)
		}
	})
	u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.AllowClassifierImages = true })
	return u
}

func classifierImageMessages(images []ai.Image) []any {
	content := []any{map[string]any{"type": "input_text", "text": "track order"}}
	for _, img := range images {
		content = append(content, map[string]any{"type": "input_image", "image_url": "data:" + img.MimeType + ";base64," + img.Data})
	}
	return []any{map[string]any{"role": "user", "content": content}}
}

func enqueueClassifierImages(t *testing.T, ev *evidence.Case, u *unaryWorld) {
	t.Helper()
	enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + decisionsAnswer + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
}

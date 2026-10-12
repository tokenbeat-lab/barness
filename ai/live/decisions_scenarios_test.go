//go:build live

package live

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	_ "golang.org/x/image/webp"
)

func decisionsScenarios(t *testing.T) []decisionsScenario {
	// These literals are synthetic protocol tasks, not a business evaluation.
	choice := singleChoiceRequest()
	score := ai.ClassifierRequest{State: choice.State, Questions: map[string]ai.ClassifierQuestion{"severity": mixedQuestionsRequest().Questions["severity"]}}
	predicate := ai.ClassifierRequest{State: choice.State, Questions: map[string]ai.ClassifierQuestion{"urgent": mixedQuestionsRequest().Questions["urgent"]}}
	out := []decisionsScenario{{id: "text-choice", request: choice}, {id: "text-score", request: score}, {id: "text-bool", request: predicate}, {id: "mixed-questions", request: mixedQuestionsRequest()}}
	images := candidateImageAssets(t)
	for i, id := range []string{"image-png", "image-jpeg", "image-webp"} {
		req := singleChoiceRequest()
		req.State = json.RawMessage(`"Describe the subject of this synthetic test image."`)
		req.Questions = map[string]ai.ClassifierQuestion{"subject": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Select the image subject"`), Criteria: map[string]json.RawMessage{"shape": json.RawMessage(`"A colored geometric shape"`), "gopher": json.RawMessage(`"A cartoon gopher"`)}}}
		req.Images = []ai.Image{images[i]}
		out = append(out, decisionsScenario{id: id, request: req})
	}
	return append(out, decisionsScenario{id: "cancel-in-flight", request: singleChoiceRequest(), cancel: true}, decisionsScenario{id: "recovery", request: mixedQuestionsRequest()})
}
func candidateImageAssets(t *testing.T) []ai.Image {
	t.Helper()
	raster := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			raster.Set(x, y, color.RGBA{R: 230, G: 40, B: 30, A: 255})
		}
	}
	var p, j bytes.Buffer
	if err := png.Encode(&p, raster); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&j, raster, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	// Independent live assets use the same licensed WebP binary as public E2E;
	// it is copied into candidate sources so child filesystem tests are isolated.
	webp, err := os.ReadFile("testdata/decisions/gopher.webp")
	if err != nil {
		t.Fatal(err)
	}
	return []ai.Image{{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(p.Bytes())}, {MimeType: "image/jpeg", Data: base64.StdEncoding.EncodeToString(j.Bytes())}, {MimeType: "image/webp", Data: base64.StdEncoding.EncodeToString(webp)}}
}
func decodeCandidateImage(img ai.Image) ([]byte, error) {
	return base64.StdEncoding.Strict().DecodeString(img.Data)
}
func recordCandidateImages(cs *evidence.Case, images []ai.Image) {
	for i, img := range images {
		raw, err := decodeCandidateImage(img)
		if err != nil {
			cs.T().Fatal(err)
		}
		raster, format, err := image.Decode(bytes.NewReader(raw))
		if !cs.Check("real raster fully decodes", err == nil, "invalid %s asset", img.MimeType) {
			continue
		}
		ext := map[string]string{"image/png": "png", "image/jpeg": "jpg", "image/webp": "webp"}[img.MimeType]
		cs.Fixture(fmt.Sprintf("input-%d.%s", i, ext), raw)
		cs.Record(fmt.Sprintf("input-%d-image", i), map[string]any{"mime": img.MimeType, "codec": format, "width": raster.Bounds().Dx(), "height": raster.Bounds().Dy(), "decodedBytes": len(raw), "base64Bytes": len(img.Data), "dataURLBytes": len(img.Data) + len(img.MimeType) + 13, "sha256": sha256Text(raw)})
	}
}

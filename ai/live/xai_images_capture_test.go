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
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xaiimages"
)

func xaiImagesRequest(t *testing.T, scene xaiimages.Scene) ai.ImagesRequest {
	t.Helper()
	ref, _ := smokeImageInputs(t)
	req := ai.ImagesRequest{Prompt: "Draw a simple red circle on a plain white background."}
	if scene.References == 0 {
		return req
	}
	req.Prompt = "Combine the supplied images in their given order. Keep a simple white background."
	if scene.ID == "single-edit" {
		raster := image.NewRGBA(image.Rect(0, 0, 256, 256))
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, raster, nil); err != nil {
			t.Fatal(err)
		}
		ref = ai.Image{Data: base64.StdEncoding.EncodeToString(buf.Bytes()), MimeType: "image/jpeg"}
	}
	for index := 0; index < scene.References; index++ {
		req.ReferenceImages = append(req.ReferenceImages, ref)
	}
	if scene.ID == "ordered-edit" {
		root, err := minimaxlive.Root()
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(root, "ai/e2e/testdata/decisions/images/gopher.webp"))
		if err != nil {
			t.Fatal(err)
		}
		req.ReferenceImages[1] = ai.Image{Data: base64.StdEncoding.EncodeToString(raw), MimeType: "image/webp"}
	}
	return req
}

func validateXAIImagesCapture(ex exchange, req ai.ImagesRequest, scene xaiimages.Scene, res ai.ImagesResult) error {
	path := "/images/generations"
	if scene.References > 0 {
		path = "/images/edits"
	}
	var request struct {
		Model, Prompt, Quality, Resolution string
		N                                  int
		AspectRatio                        string `json:"aspect_ratio"`
		ResponseFormat                     string `json:"response_format"`
		Image                              *struct{ Type, URL string }
		Images                             []struct{ Type, URL string }
	}
	if json.Unmarshal([]byte(ex.RequestBody), &request) != nil || ex.Method != "POST" || ex.URL != xaiimages.Endpoint+path || ex.RequestHeaders["content-type"] != "application/json" || ex.Truncated || request.Model != xaiimages.Model || request.Prompt != req.Prompt || request.N != scene.Outputs || request.Quality != scene.Quality || request.Resolution != scene.Resolution || request.AspectRatio != scene.AspectRatio || request.ResponseFormat != "b64_json" {
		return fmt.Errorf("fixed native JSON request differs")
	}
	var sources []struct{ Type, URL string }
	if request.Image != nil {
		sources = append(sources, *request.Image)
	}
	sources = append(sources, request.Images...)
	if len(sources) != len(req.ReferenceImages) || scene.References == 1 && request.Image == nil || scene.References != 1 && request.Image != nil {
		return fmt.Errorf("native reference shape differs")
	}
	for index, source := range sources {
		img := req.ReferenceImages[index]
		if source.Type != "image_url" || source.URL != "data:"+img.MimeType+";base64,"+img.Data {
			return fmt.Errorf("ordered inline reference differs")
		}
	}
	var response struct {
		Data []struct {
			Base64 string `json:"b64_json"`
			MIME   string `json:"mime_type"`
			URL    string
		}
		Usage struct {
			Ticks json.RawMessage `json:"cost_in_usd_ticks"`
		}
	}
	if json.Unmarshal([]byte(ex.ResponseBody), &response) != nil || len(response.Data) != scene.Outputs || len(res.Content) != scene.Outputs {
		return fmt.Errorf("incomplete result capture")
	}
	var ticks int64
	if json.Unmarshal(response.Usage.Ticks, &ticks) != nil || ticks < 0 {
		return fmt.Errorf("integer reported charge absent")
	}
	usd, known := res.Metadata.Attempts[0].ReportedCostUSD.Get()
	if !known || math.Abs(usd-float64(ticks)/1e10) > 1e-12 {
		return fmt.Errorf("raw ticks and public USD differ")
	}
	for index, entry := range response.Data {
		block, ok := res.Content[index].(ai.ImageOutputImage)
		raw, err := base64.StdEncoding.Strict().DecodeString(entry.Base64)
		config, format, imageErr := image.DecodeConfig(bytes.NewReader(raw))
		if !ok || err != nil || imageErr != nil || entry.URL != "" || block.Data != entry.Base64 || block.MimeType != "image/"+format || config.Width < 1 || config.Height < 1 {
			return fmt.Errorf("raster capture differs")
		}
	}
	return nil
}

func xaiImagesFixture(req *http.Request, scene xaiimages.Scene) (*http.Response, error) {
	// An incompressible PNG keeps native response bytes above the old 512KiB cap.
	// Trailing integer ticks must survive intact for independent consumption.
	raster := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	var state uint32 = 23
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			state = state*1664525 + 1013904223
			raster.SetNRGBA(x, y, color.NRGBA{uint8(state), uint8(state >> 8), uint8(state >> 16), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, raster); err != nil {
		return nil, err
	}
	data := make([]map[string]any, scene.Outputs)
	for i := range data {
		data[i] = map[string]any{"b64_json": base64.StdEncoding.EncodeToString(buf.Bytes())}
	}
	raw, err := json.Marshal(map[string]any{"data": data, "usage": map[string]any{"cost_in_usd_ticks": int64(math.Round(scene.MaximumUSD * 1e10))}})
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}, "X-Request-Id": {"req-xai-images-controlled"}}, Body: io.NopCloser(strings.NewReader(string(raw))), Request: req}, nil
}

//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Exercise the evidence boundary through the public Client: even rejected,
// unreadable or oversized vendor responses must not persist signed material.
func TestGoogleImagesHarnessCapture(t *testing.T) {
	for _, name := range []string{"request-prefix", "response-prefix", "malformed", "read-interrupted"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "P09-live-capture-"+name)
			i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "google-interactions-image" })
			c := &combos[i]
			rec := newRecorder()
			defer rec.next.(*http.Transport).CloseIdleConnections()
			secrets := []string{"fixture-thought-signature", "fixture-output-signature", "fixture-continuation-token", "https://example.invalid/image?signature=fixture-signed-resource"}
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				// Sorted before padding, so even the captured response prefix
				// contains both native signature fields.
				response := map[string]any{
					"aaa":    map[string]any{"thought_signature": secrets[0], "signature": secrets[1]},
					"status": "completed", "model": c.model,
					"steps": []any{map[string]any{"type": "model_output", "content": []any{map[string]any{"type": "image", "mime_type": "image/png", "data": testPNG(t, 1024, 1024)}}}},
				}
				if name == "response-prefix" {
					response["padding"] = strings.Repeat("p", 16<<20)
				}
				wire, _ := json.Marshal(response)
				var body io.Reader = bytes.NewReader(wire)
				if name == "malformed" || name == "read-interrupted" {
					wire = []byte(`{"thought_signature":"` + secrets[0] + `","signature":"` + secrets[1] + `","continuation_token":"` + secrets[2] + `","uri":"` + secrets[3] + `",`)
					body = bytes.NewReader(wire)
					if name == "read-interrupted" {
						body = io.MultiReader(body, captureReadFailure{})
					}
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(body), Request: req}, nil
			})
			client, err := newClient(c, "key:google-image-test", rec, googleImageProbeCatalog(t))
			if err != nil {
				t.Fatal(err)
			}
			prompt := "Generate a square."
			request := ai.ImagesRequest{Prompt: prompt}
			if name == "request-prefix" {
				var data bytes.Buffer
				encoder := png.Encoder{CompressionLevel: png.NoCompression}
				if err := encoder.Encode(&data, image.NewNRGBA(image.Rect(0, 0, 256, 256))); err != nil {
					t.Fatal(err)
				}
				request.ReferenceImages = []ai.Image{{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(data.Bytes())}}
			}
			res, err := client.GenerateImages(context.Background(), newScope(), ai.Target{BindingID: c.name, ModelID: c.model}, request, googleImageSmokeOptions())
			cs.Check("client preserves response failure", (err == nil) == (name == "request-prefix"), "err=%v", err)
			exchanges := rec.take()
			cs.Record("exchanges", exchanges)
			cs.Record("result", map[string]any{"result": res, "error": errText(err)})
			cs.Check("one isolated capture", len(exchanges) == 1, "wrong transport count")
			if len(exchanges) != 1 {
				return
			}
			capture := exchanges[0].ResponseBody
			for _, secret := range secrets {
				cs.Check("sensitive output omitted", !strings.Contains(capture, secret), "native signed material persisted")
			}
			cs.Check("safe bounded failure evidence", len(capture) < 2048 && strings.Contains(capture, "ELIDED") && strings.Contains(capture, "sha256="), "missing bounded diagnostic marker")
			cs.Check("prefix flags remain accurate", exchanges[0].Truncated == (name == "request-prefix" || name == "response-prefix"), "wrong truncation flag")
			if name == "request-prefix" {
				cs.Check("complete response still preserves shape", strings.Contains(capture, `"status":"completed"`) && strings.Contains(capture, `"thought_signature"`), "request truncation lost response shape")
			}
		})
	}
}

type captureReadFailure struct{}

func (captureReadFailure) Read([]byte) (int, error) {
	return 0, errors.New("fixture interrupted response")
}

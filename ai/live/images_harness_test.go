//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestImagesHarnessRegistration(t *testing.T) {
	i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "openai-images" })
	if i < 0 {
		t.Fatal("P08 has no isolated image smoke")
	}
	c := &combos[i]
	if c.operation != ai.OperationImage || c.provider != ai.ProviderOpenAI || c.api != ai.APIOpenAIImages || c.model != "gpt-image-2.5-sunburst-2026-09-08" || c.endpoint != "https://api.openai.com/v1" {
		t.Fatal("P08 does not use fixed official route")
	}
	var ids []string
	for _, s := range scenariosOf(c) {
		ids = append(ids, s.id)
	}
	if !slices.Equal(ids, []string{"json-edit", "generation", "mask-edit"}) {
		t.Fatalf("edit must be first and mandatory: %v", ids)
	}
}

func testPNG(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 80, G: 120, B: 200, A: 255})
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b.Bytes())
}

func TestImagesHarnessWire(t *testing.T) {
	for _, name := range []string{"success", "usage-absent", "usage-partial", "json-refused", "mask-refused", "mask-format-refused", "mask-size-refused", "invalid-image", "wrong-size", "transparent-opaque"} {
		t.Run(name, func(t *testing.T) {
			before := budget
			defer func() { budget = before }()
			budget = imagesBudget()
			cs := run.Case(t, "P08-live-harness-"+name)
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				cs.Check("native JSON edit", r.URL.Path == "/v1/images/edits" && r.Header.Get("Content-Type") == "application/json" && body["model"] == "gpt-image-2.5-sunburst-2026-09-08" && body["size"] == "1024x1024" && body["n"] == float64(1), "wrong native request")
				refs, _ := body["images"].([]any)
				cs.Check("inline reference", len(refs) == 1 && strings.HasPrefix(refs[0].(map[string]any)["image_url"].(string), "data:image/png;base64,"), "missing inline JSON image")
				_, fidelity := body["input_fidelity"]
				cs.Check("undocumented fidelity omitted", !fidelity, "input_fidelity sent")
				w.Header().Set("x-request-id", "req-controlled-image")
				w.Header().Set("Content-Type", "application/json")
				if name == "json-refused" || strings.HasPrefix(name, "mask-") && strings.HasSuffix(name, "refused") {
					w.WriteHeader(400)
					message := "JSON edit input rejected"
					if name == "mask-refused" {
						message = "mask is not supported by this model"
					}
					if name == "mask-format-refused" {
						message = "unsupported mask image format"
					}
					if name == "mask-size-refused" {
						message = "mask size is not supported"
					}
					json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": message, "type": "invalid_request_error"}})
					return
				}
				width := 1024
				if name == "wrong-size" {
					width = 1536
				}
				data := testPNG(t, width, 1024)
				if name == "invalid-image" {
					data = "invalid"
				}
				response := map[string]any{"data": []any{map[string]any{"b64_json": data}}, "output_format": "png", "size": "1024x1024"}
				if name != "usage-absent" {
					usage := map[string]any{"input_tokens": 30, "output_tokens": 10, "total_tokens": 40, "input_tokens_details": map[string]int{"text_tokens": 10, "image_tokens": 20}}
					if name != "usage-partial" {
						usage["output_tokens_details"] = map[string]int{"text_tokens": 0, "image_tokens": 10}
					}
					response["usage"] = usage
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()
			i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "openai-images" })
			if i < 0 {
				t.Fatal("missing image combo")
			}
			c := &combos[i]
			rec := newRecorder()
			defer rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				copy := req.Clone(req.Context())
				u := *req.URL
				copy.URL = &u
				copy.URL.Host = strings.TrimPrefix(server.URL, "https://")
				copy.Host = copy.URL.Host
				return server.Client().Transport.RoundTrip(copy)
			})
			client, err := newClient(c, "key:image-test", rec, imageProbeCatalog(t))
			if err != nil {
				t.Fatal(err)
			}
			s := &session{t: t, ctx: context.Background(), cs: cs, env: &liveEnv{combo: c, client: client, rec: rec}, model: c.model, attempt: 1}
			req := ai.ImagesRequest{Prompt: "Change the blue square to green.", ReferenceImages: []ai.Image{{Data: testPNG(t, 256, 256), MimeType: "image/png"}}}
			opts := imageSmokeOptions("png", "low")
			if name == "transparent-opaque" {
				opts.Background = ai.Value("transparent")
			}
			if name == "mask-refused" {
				mask := req.ReferenceImages[0]
				opts.Mask = &mask
			}
			res, err := s.generateImage("controlled", req, opts)
			cs.Check("count sent and reserved", count == 1 && budget.HTTPAttempts == 1 && budget.ImagesUsed == 1 && budget.CallsUsed == 1 && s.images == 1, "lost request consumption")
			if name == "json-refused" || strings.HasPrefix(name, "mask-") && strings.HasSuffix(name, "refused") {
				cs.Check("only optional explicit mask refusal", explicitMaskRefusal(err, s.lastExchanges) == (name == "mask-refused"), "wrong optional classification")
				return
			}
			if name == "invalid-image" {
				cs.Check("invalid image clears output but retains usage", err != nil && len(res.Content) == 0 && res.Usage.Input == 30, "lost usage")
				return
			}
			cs.Check("call succeeds", err == nil, "err=%v", err)
			if name == "transparent-opaque" {
				cs.Check("opaque output cannot confirm transparency", s.imageCalls[0].Verified == 0, "opaque raster was verified as transparent")
				return
			}
			_, validation := verifyImages(res, imageSmokeOptions("png", "low"))
			cs.Check("dimension guard", (validation == nil) == (name != "wrong-size"), "wrong raster acceptance: %v", validation)
			reporting := ai.UsageComplete
			if name == "usage-absent" {
				reporting = ai.UsageUnreported
			}
			if name == "usage-partial" {
				reporting = ai.UsagePartial
			}
			cs.Check("usage shape preserved", res.Metadata.Attempts[0].UsageReporting == reporting, "wrong usage reporting")
			expectedCost := 0.00051
			if name == "usage-absent" {
				expectedCost = 0
			}
			if name == "usage-partial" {
				expectedCost = 0.00021
			}
			cs.Check("known-only official token pricing", math.Abs(res.Usage.Cost.Total-expectedCost) < 1e-12, "cost=%+v", res.Usage.Cost)
			cs.Record("image-check", s.imageCalls)
		})
	}
}

func TestImagesHarnessNotRun(t *testing.T) {
	for _, c := range combos {
		t.Setenv(keyVar(c.name), "")
	}
	t.Setenv(envCombo, "openai-images")
	t.Setenv(envLive, "1")
	config := loadConfig()
	if config.combo == nil || config.notRun == "" || config.refused != "" {
		t.Fatal("missing image credential must be NOT_RUN")
	}
	t.Setenv(keyVar("openai-images"), "key:image-test")
	t.Setenv(envAlias, "dotenv-openai@region-unconfirmed")
	if loadConfig().refused != "" || loadConfig().notRun != "" {
		t.Fatal("isolated key refused")
	}
	t.Setenv(keyVar("openai-chat"), "key:foreign-test")
	if loadConfig().refused == "" {
		t.Fatal("foreign combo key accepted")
	}
}

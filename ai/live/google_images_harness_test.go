//go:build live

package live

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGoogleImagesHarnessRegistration(t *testing.T) {
	i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "google-interactions-image" })
	if i < 0 {
		t.Fatal("P09 has no isolated Google image smoke")
	}
	c := &combos[i]
	if c.operation != ai.OperationImage || c.provider != ai.ProviderGoogle || c.api != ai.APIGoogleInteractions || c.model != "gemini-nano-banana-2.1" || c.endpoint != "https://generativelanguage.googleapis.com/v1beta" {
		t.Fatal("P09 must pin its own official v1beta target and bare model")
	}
	var ids []string
	for _, s := range scenariosOf(c) {
		ids = append(ids, s.id)
	}
	if !slices.Equal(ids, []string{"generation", "reference-edit"}) {
		t.Fatalf("mandatory generation and editing missing: %v", ids)
	}
}

func TestGoogleImagesHarnessWire(t *testing.T) {
	for _, name := range []string{"complete", "thought-included", "partial", "unreported", "identity-absent", "live-image-only", "wrong-size", "uri", "continuation", "pending", "refused"} {
		t.Run(name, func(t *testing.T) {
			old := budget
			defer func() { budget = old }()
			budget = imagesBudget()
			cs := run.Case(t, "P09-live-harness-"+name)
			i := slices.IndexFunc(combos, func(c combo) bool { return c.name == "google-interactions-image" })
			if i < 0 {
				t.Fatal("missing Google image combo")
			}
			c := &combos[i]
			rec := newRecorder()
			defer rec.next.(*http.Transport).CloseIdleConnections()
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				data, _ := io.ReadAll(req.Body)
				var body map[string]any
				json.Unmarshal(data, &body)
				cs.Check("fixed route and header authentication", req.URL.String() == "https://generativelanguage.googleapis.com/v1beta/interactions" && req.Header.Get("X-Goog-Api-Key") == "key:google-image-test" && body["store"] == false, "wrong target or auth")
				format, _ := body["response_format"].(map[string]any)
				cs.Check("fixed inline 1K output", format["type"] == "image" && format["delivery"] == nil && format["image_size"] == "1K" && format["aspect_ratio"] == "1:1", "wrong output format")
				width := 1024
				if name == "wrong-size" {
					width = 1536
				}
				block := map[string]any{"type": "image", "mime_type": "image/png", "data": testPNG(t, width, 1024)}
				if name == "uri" {
					delete(block, "data")
					block["uri"] = "https://example.invalid/image.png"
				}
				textTokens := 0
				if name == "thought-included" {
					textTokens = 22
				}
				usage := map[string]any{"total_input_tokens": 7, "total_output_tokens": 1120, "total_thought_tokens": 22, "total_tokens": 1149, "input_tokens_by_modality": []any{map[string]any{"modality": "text", "tokens": 7}, map[string]any{"modality": "image", "tokens": 0}}, "output_tokens_by_modality": []any{map[string]any{"modality": "text", "tokens": textTokens}, map[string]any{"modality": "image", "tokens": 1120}}}
				response := map[string]any{"id": "interaction-controlled", "model": c.model, "status": "completed", "steps": []any{map[string]any{"type": "model_output", "content": []any{map[string]any{"type": "text", "text": "before"}}}, map[string]any{"type": "model_output", "content": []any{block, map[string]any{"type": "text", "text": "after"}}}}, "usage": usage}
				if name == "live-image-only" {
					delete(response, "id")
					usage["total_output_tokens"] = 1260
					usage["total_thought_tokens"] = 121
					usage["total_tokens"] = 1388
					usage["output_tokens_by_modality"] = []any{map[string]any{"modality": "image", "tokens": 1120}}
				}
				if name == "identity-absent" {
					delete(response, "id")
					delete(response, "model")
				}
				if name == "partial" {
					delete(usage, "output_tokens_by_modality")
				}
				if name == "unreported" {
					delete(response, "usage")
				}
				if name == "continuation" {
					response["continuation_token"] = "continue"
				}
				if name == "pending" {
					response["status"] = "in_progress"
				}
				status := 200
				if name == "refused" {
					status = 400
					response = map[string]any{"error": map[string]any{"message": "fixed image model refused"}}
				}
				// Force usage beyond the chat recorder's prefix budget.
				response["padding"] = strings.Repeat("p", 600<<10)
				wire, _ := json.Marshal(response)
				return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(wire)), Request: req}, nil
			})
			client, err := newClient(c, "key:google-image-test", rec, googleImageProbeCatalog(t))
			if err != nil {
				t.Fatal(err)
			}
			s := &session{t: t, ctx: context.Background(), cs: cs, env: &liveEnv{combo: c, client: client, rec: rec}, model: c.model, attempt: 1}
			ref, _ := smokeImageInputs(t)
			res, err := s.generateGoogleImage("controlled", ai.ImagesRequest{Prompt: "Edit the square.", ReferenceImages: []ai.Image{ref}}, googleImageSmokeOptions())
			cs.Check("bounded reservation", budget.CallsUsed == 1 && budget.ImagesUsed == 1 && budget.HTTPAttempts == 1, "lost consumption")
			failed := slices.Contains([]string{"uri", "continuation", "pending", "refused"}, name)
			cs.Check("protocol errors remain failures", (err != nil) == failed, "err=%v", err)
			if name == "uri" || name == "continuation" {
				capture := s.lastExchanges[0].ResponseBody
				cs.Check("rejected resources stay redacted", !strings.Contains(capture, "https://example.invalid/image.png") && !strings.Contains(capture, `"continue"`), "rejected URI or continuation value retained")
			}
			if failed {
				return
			}
			_, validation := verifyGoogleImages(res)
			cs.Check("readable square raster", (validation == nil) == (name != "wrong-size"), "validation=%v", validation)
			expected := ai.UsageComplete
			if name == "partial" || name == "live-image-only" {
				expected = ai.UsagePartial
			}
			if name == "unreported" {
				expected = ai.UsageUnreported
			}
			cs.Check("usage reporting", res.Metadata.Attempts[0].UsageReporting == expected, "wrong reporting")
			if name != "unreported" {
				want := 0.0337755
				if name == "partial" {
					want = 0.0001755
				}
				output, total := int64(1142), int64(1149)
				if name == "live-image-only" {
					want = 0.034518
					output = 1381
					total = 1388
				}
				cs.Check("thinking counted once", res.Usage.Output == output && res.Usage.TotalTokens == total && math.Abs(res.Usage.Cost.Total-want) < 1e-12, "usage=%+v", res.Usage)
			}
			if name == "live-image-only" {
				g := s.imageCalls[0].Google
				cs.Check("image-only shape excludes separately reported text thought", g.ThoughtInModalities != nil && !*g.ThoughtInModalities, "thought inclusion not established")
			}
			cs.Check("all model output content retained", len(res.Content) == 3, "lost ordered blocks")
			cs.Check("bounded capture retains trailing usage", !s.lastExchanges[0].Truncated && (name == "unreported" || strings.Contains(s.lastExchanges[0].ResponseBody, "total_thought_tokens")) && len(s.lastExchanges[0].ResponseBody) < 8192, "shape evidence lost or unbounded")
		})
	}
}

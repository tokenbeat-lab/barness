package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGoogleImagesEditingCallbackBudgets(t *testing.T) {
	large := strings.Repeat("A", 8<<20)
	byteData := append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 8<<20)...)
	nearData := base64.StdEncoding.EncodeToString(byteData[:6<<20])
	raw := json.RawMessage(`[{"type":"text","text":"p"},{"type":"image","mime_type":"image/png","data":"` + large + `"}]`)
	escaped := json.RawMessage(`[{"type":"text","text":"p"},{"type":"image","mime_type":"image/png","data":"` + strings.Repeat(`\u0041`, 1<<20) + `"}]`)
	for _, name := range []string{"ordinary", "typed", "pointer", "struct", "raw", "escaped-raw", "raw-data", "host-count", "model-count", "protocol-count", "raw-count", "protocol-bytes", "struct-total", "promoted", "byte-data", "custom-key", "struct-overhead", "raw-struct-total"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E08-edit-allocation-"+name)
			sc := googleEditScenario(t)
			w := newWorldWith(t, func(c *ai.Config) {
				configureGoogleImages(c)
				c.Policy.MaxRequestBytes = 64 << 20
				switch name {
				case "host-count":
					c.Policy.Image.MaxInputImages = 1
				case "model-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 1
				case "protocol-count", "raw-count":
					c.Catalog.ImageModels[0].Capabilities.MaxReferenceImages = 16
				case "struct-overhead":
					c.Policy.MaxRequestBytes, c.Policy.MaxImageBytes = 16_800_000, 8<<20
				case "protocol-bytes", "struct-total", "raw-struct-total":
					c.Policy.MaxImageBytes = 8 << 20
				}
			}, tenantA)
			installGoogleImagesBinding(w, tenantA)
			refs := imagesInput(t, sc).ReferenceImages
			countRaw := json.RawMessage(`[{"type":"text","text":"p"},` + strings.Repeat(`{"type":"image","mime_type":"image/png","data":"`+refs[0].Data+`"},`, 14) + `{"type":"image","mime_type":"image/png","data":"` + refs[0].Data + `"}]`)
			input := []any{map[string]any{"type": "text", "text": "p"}, map[string]any{"type": "image", "mime_type": "image/png", "data": large}}
			valueRaw := json.RawMessage(`"` + large + `"`)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "ordinary":
					p.Body["input"] = input
				case "typed":
					p.Body["input"] = []map[string]string{{"type": "text", "text": "p"}, {"type": "image", "mime_type": "image/png", "data": large}}
				case "pointer":
					p.Body["input"] = &input
				case "struct":
					p.Body["input"] = []any{input[0], struct {
						Type string  `json:"type"`
						Mime string  `json:"mime_type"`
						Data *string `json:"data"`
					}{"image", "image/png", &large}}
				case "raw":
					p.Body["input"] = &raw
				case "escaped-raw":
					p.Body["input"] = escaped
				case "raw-data":
					p.Body["input"] = []any{input[0], map[string]any{"type": "image", "mime_type": "image/png", "data": valueRaw}}
				case "promoted":
					p.Body["input"] = []any{input[0], googleDataWrapper{googleDataFields{large}, "image", "image/png"}}
				case "byte-data":
					p.Body["input"] = []any{input[0], map[string]any{"type": "image", "mime_type": "image/png", "data": byteData}}
				case "custom-key":
					p.Body["input"] = []any{input[0], map[googleDataKey]string{0: "image", 1: "image/png", 2: large}}
				case "struct-overhead":
					p.Body["input"] = []any{struct {
						Type string `json:"type"`
						Text string `json:"text"`
					}{"text", strings.Repeat("\x00", 30000)}, googleDataWrapper{googleDataFields{nearData}, "image", "image/png"}, googleDataWrapper{googleDataFields{nearData}, "image", "image/png"}}
				case "raw-count":
					p.Body["input"] = countRaw
				case "host-count", "model-count", "protocol-count":
					parts := []any{input[0]}
					count := 2
					if name == "protocol-count" {
						count = 15
					}
					for i := 0; i < count; i++ {
						parts = append(parts, map[string]any{"type": "image", "mime_type": "image/png", "data": refs[0].Data})
					}
					p.Body["input"] = parts
				case "raw-struct-total":
					parts := []any{input[0]}
					for i := 0; i < 3; i++ {
						parts = append(parts, struct {
							Type string          `json:"type"`
							Mime string          `json:"mime_type"`
							Data json.RawMessage `json:"data"`
						}{"image", "image/png", valueRaw})
					}
					p.Body["input"] = parts
				case "struct-total":
					parts := []any{input[0]}
					for i := 0; i < 3; i++ {
						parts = append(parts, struct {
							Type string `json:"type"`
							Mime string `json:"mime_type"`
							Data string `json:"data"`
						}{"image", "image/png", large})
					}
					p.Body["input"] = parts
				case "protocol-bytes":
					p.Body["input"] = []any{input[0], input[1], input[1], input[1]}
				}
				return ai.KeepPayload(), nil
			}}
			req := imagesInput(t, sc)
			// Callback insertion is also legal from generation on this same route.
			req.ReferenceImages = nil
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-google-allocation-"+name), sc.target(), req, nil)
			runtime.ReadMemStats(&after)
			ev.Record("allocated-bytes", after.TotalAlloc-before.TotalAlloc)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("known callback budget before serialization", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && after.TotalAlloc-before.TotalAlloc < 4<<20 && len(w.provider.Requests()) == 0, "got %v allocated=%d", err, after.TotalAlloc-before.TotalAlloc)
		})
	}
}

type googleDataFields struct {
	Data string `json:"data"`
}
type googleDataWrapper struct {
	googleDataFields
	Type string `json:"type"`
	Mime string `json:"mime_type"`
}
type googleDataKey int

func (k googleDataKey) MarshalText() ([]byte, error) {
	return []byte([]string{"type", "mime_type", "data"}[k]), nil
}

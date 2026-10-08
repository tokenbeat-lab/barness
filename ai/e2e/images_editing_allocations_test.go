package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestOpenAIImagesEditingKnownSize(t *testing.T) {
	large := "data:image/png;base64," + strings.Repeat("A", 8<<20)
	raw := json.RawMessage(`[{"image_url":"` + large + `"}]`)
	for _, name := range []string{"images", "mask", "raw", "many", "typed-url", "raw-many", "pointer-url", "pointer-slice", "struct", "large-budget-raw", "large-budget-mask", "escaped-raw", "embedded", "shadowed", "text-key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E08-edit-allocation-"+name)
			sc := editScenario(t)
			w := scenarioWorld(t, sc)
			if name == "typed-url" || name == "pointer-url" || name == "struct" || name == "embedded" || name == "shadowed" || name == "text-key" || strings.HasPrefix(name, "large-budget-") || name == "escaped-raw" {
				w = newWorldWith(t, func(c *ai.Config) {
					configureImages(c)
					c.Policy.MaxRequestBytes = 64 << 20
					if name == "typed-url" {
						c.Policy.MaxImageBytes = 32 << 20
					}
				}, tenantA)
				installImagesBinding(w, tenantA)
			}
			typed := []map[string]string{{"image_url": "data:image/png;base64," + strings.Repeat("A", 20971520)}}
			smallURL := "data:image/png;base64," + imagesInput(t, sc).ReferenceImages[0].Data
			manyRaw := json.RawMessage(`[` + strings.Repeat(`{"image_url":"`+smallURL+`"},`, 1000) + `{"image_url":"` + smallURL + `"}]`)
			maskRaw := json.RawMessage(`{"image_url":"` + large + `"}`)
			escapedRaw := json.RawMessage(`[{"image_url":"data:image/png;base64,` + strings.Repeat(`\u0041`, 1<<20) + `"}]`)
			many := make([]any, 100000)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "images":
					p.Body["images"] = []any{map[string]any{"image_url": large}}
				case "mask":
					p.Body["mask"] = map[string]any{"image_url": large}
				case "raw":
					p.Body["images"] = raw
				case "many":
					p.Body["images"] = make([]any, 100000)
				case "typed-url":
					p.Body["images"] = typed
				case "raw-many":
					p.Body["images"] = manyRaw
				case "pointer-url":
					p.Body["images"] = []map[string]*string{{"image_url": &large}}
				case "pointer-slice":
					p.Body["images"] = &many
				case "struct":
					p.Body["images"] = []struct {
						URL *string `json:"image_url"`
					}{{&large}}
				case "text-key":
					p.Body["images"] = []map[imageURLKey]string{{{}: large}}
				case "embedded":
					p.Body["images"] = []wrappedInlineImage{{inlineImageFields{large}}}
				case "shadowed":
					p.Body["images"] = []shadowedInlineImage{{inlineImageFields: inlineImageFields{large}}}
				case "large-budget-raw":
					p.Body["images"] = &raw
				case "large-budget-mask":
					p.Body["mask"] = maskRaw
				case "escaped-raw":
					p.Body["images"] = escapedRaw
				}
				return ai.KeepPayload(), nil
			}}
			req := imagesInput(t, sc)
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-edit-allocation-"+name), sc.target(), req, nil)
			runtime.ReadMemStats(&after)
			allocated := after.TotalAlloc - before.TotalAlloc
			ev.Record("allocated-bytes", allocated)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("known edit input bounded before serialization", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && allocated < 4<<20 && len(w.provider.Requests()) == 0, "got %v allocated=%d", err, allocated)
		})
	}
}

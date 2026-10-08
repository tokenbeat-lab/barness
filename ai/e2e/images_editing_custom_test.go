package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

const callbackInlineImage = `{"image_url":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+jBMsAAAAASUVORK5CYII="}`

type customImageReferences []any

func (customImageReferences) MarshalJSON() ([]byte, error) {
	return []byte(`[` + callbackInlineImage + `]`), nil
}

type customImageReference map[string]any

func (customImageReference) MarshalJSON() ([]byte, error) { return []byte(callbackInlineImage), nil }

type pointerEncodedImageReference struct{ Backing string }

func (*pointerEncodedImageReference) MarshalJSON() ([]byte, error) {
	return []byte(callbackInlineImage), nil
}

type imageURLKey struct{}

func (imageURLKey) MarshalText() ([]byte, error) { return []byte("image_url"), nil }

type inlineImageFields struct {
	URL string `json:"image_url"`
}
type wrappedInlineImage struct{ inlineImageFields }
type shadowedInlineImage struct {
	inlineImageFields
	URL string `json:"unrelated,omitempty"`
}

func TestOpenAIImagesEditingCustomEncoding(t *testing.T) {
	for _, name := range []string{"slice", "map", "pointer-element", "text-key", "embedded", "shadowed", "raw-value", "raw-struct-value", "raw-text-key-value"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E04-edit-custom-"+name)
			sc := editScenario(t)
			w := scenarioWorld(t, sc)
			enqueue(ev, w, sc.replies(t)...)
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				switch name {
				case "slice":
					p.Body["images"] = make(customImageReferences, 17)
				case "map":
					p.Body["images"] = []any{customImageReference{"backing": true}}
				case "pointer-element":
					p.Body["images"] = []pointerEncodedImageReference{{"backing"}}
				case "text-key":
					p.Body["images"] = []map[imageURLKey]string{{{}: "data:image/png;base64," + imagesInput(t, sc).ReferenceImages[0].Data}}
				case "raw-value":
					p.Body["images"] = []map[string]json.RawMessage{{"image_url": json.RawMessage(`"data:image/png;base64,` + imagesInput(t, sc).ReferenceImages[0].Data + `"`)}}
				case "raw-struct-value":
					p.Body["images"] = []struct {
						URL json.RawMessage `json:"image_url"`
					}{{json.RawMessage(`"data:image/png;base64,` + imagesInput(t, sc).ReferenceImages[0].Data + `"`)}}
				case "raw-text-key-value":
					p.Body["images"] = []map[imageURLKey]json.RawMessage{{{}: json.RawMessage(`"data:image/png;base64,` + imagesInput(t, sc).ReferenceImages[0].Data + `"`)}}
				case "embedded":
					p.Body["images"] = []wrappedInlineImage{{inlineImageFields{"data:image/png;base64," + imagesInput(t, sc).ReferenceImages[0].Data}}}
				case "shadowed":
					p.Body["images"] = []shadowedInlineImage{{inlineImageFields: inlineImageFields{"data:image/png;base64," + imagesInput(t, sc).ReferenceImages[0].Data}}}
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-edit-custom-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("custom representation uses frozen JSON", err == nil && len(w.provider.Requests()) == 1, "got %v", err)
		})
	}
}

func TestOpenAIImagesEditingBoundedPointers(t *testing.T) {
	for _, name := range []string{"cycle", "deep", "mask-cycle", "mask-deep"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E04-edit-bounded-"+name)
			sc := editScenario(t)
			w := scenarioWorld(t, sc)
			var field any
			if name == "cycle" || name == "mask-cycle" {
				field = &field
			} else {
				field = "data:image/png;base64," + imagesInput(t, sc).ReferenceImages[0].Data
				for i := 0; i < 10001; i++ {
					previous := field
					field = &previous
				}
			}
			h := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				image := struct {
					URL any `json:"image_url"`
				}{field}
				if name == "mask-cycle" || name == "mask-deep" {
					p.Body["mask"] = image
				} else {
					p.Body["images"] = []any{image}
				}
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(h).GenerateImages(ctxFor(t), textScope("req-edit-bounded-"+name), sc.target(), imagesInput(t, sc), nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Check("pointer expansion is bounded before serialization", errors.Is(err, &ai.Error{Code: ai.CodeResourceLimit, Phase: ai.PhaseRequest}) && len(w.provider.Requests()) == 0, "got %v", err)
		})
	}
}

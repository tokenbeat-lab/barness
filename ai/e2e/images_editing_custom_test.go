package e2e

import (
	"context"
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

func TestOpenAIImagesEditingCustomEncoding(t *testing.T) {
	for _, name := range []string{"slice", "map", "pointer-element"} {
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

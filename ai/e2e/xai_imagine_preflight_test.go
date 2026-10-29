package e2e

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

type imagineCountedPrompt struct{ calls *int }

func (p imagineCountedPrompt) MarshalJSON() ([]byte, error) {
	*p.calls++
	return []byte(`"Draw."`), nil
}
func TestXAIImaginePreEncodingLimits(t *testing.T) {
	for _, name := range []string{"typed-url", "raw-url", "typed-count", "raw-count"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-pre-encoding-"+name)
			w := newImagineWorld(t, func(c *ai.Config) { c.Policy.MaxImageBytes = 32 })
			encoded := 0
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["prompt"] = imagineCountedPrompt{&encoded}
				url := "data:image/png;base64," + strings.Repeat("A", 1000)
				switch name {
				case "typed-url":
					type ref struct {
						Type string `json:"type"`
						URL  string `json:"url"`
					}
					p.Body["image"] = &ref{"image_url", url}
				case "raw-url":
					p.Body["image"] = json.RawMessage(`{"type":"image_url","url":"` + url + `"}`)
				case "typed-count":
					delete(p.Body, "image")
					p.Body["images"] = make([]map[string]string, 6)
				case "raw-count":
					delete(p.Body, "image")
					p.Body["images"] = json.RawMessage(`[{}, {}, {}, {}, {}, {}]`)
				}
				return ai.KeepPayload(), nil
			}}
			req := ai.ImagesRequest{Prompt: "Draw.", ReferenceImages: []ai.Image{{Data: imaginePNG, MimeType: "image/png"}}}
			res, err := w.client.WithHooks(hooks).GenerateImages(ctxFor(t), textScope("imagine-preflight-"+name), imagineTarget(), req, nil)
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("encodings", encoded)
			ev.Check("known capacity fails before encoding", err != nil && encoded == 0 && len(w.provider.Requests()) == 0, "encoded=%d err=%v", encoded, err)
		})
	}
}

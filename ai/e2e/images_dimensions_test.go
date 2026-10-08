package e2e

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestOpenAIImagesDimensionCapabilities(t *testing.T) {
	for _, name := range []string{"missing-source", "zero-multiple", "zero-ratio", "zero-edge", "zero-min", "max-less-than-min", "snapshot", "max-square", "max-landscape", "max-portrait", "model-stricter"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P08-E07-edit-dimensions-"+name)
			w := newWorld(t, tenantA)
			cfg := ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true}
			configureImages(&cfg)
			var dims map[string]any
			mustUnmarshal(t, []byte(dimensionEvidence), &dims)
			switch name {
			case "missing-source":
				delete(dims, "source")
			case "zero-multiple":
				dims["edgeMultiple"] = 0
			case "zero-ratio":
				dims["maxAspectRatio"] = 0
			case "zero-edge":
				dims["maxEdge"] = 0
			case "zero-min":
				dims["minPixels"] = 0
			case "max-less-than-min":
				dims["maxPixels"] = 1
			case "model-stricter":
				dims["maxEdge"] = 1024
			}
			patchImageModel(t, &cfg, mustMarshal(t, map[string]any{"capabilities": map[string]any{"maxReferenceImages": 16, "maxOutputImages": 10, "customSizes": dims}}))
			client, err := ai.NewClient(cfg)
			ev.Record("config-error", errString(err))
			if name == "missing-source" || name == "zero-multiple" || name == "zero-ratio" || name == "zero-edge" || name == "zero-min" || name == "max-less-than-min" {
				ev.Check("invalid constraints refused", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			w.client = client
			installImagesBinding(w, tenantA)
			if name == "snapshot" {
				patchImageModel(t, &cfg, json.RawMessage(`{"capabilities":{"maxReferenceImages":0,"maxOutputImages":1}}`))
				cat := client.Catalog()
				// JSON mutation goes through the public catalog rather than an internal clone.
				cat.ImageModels[0].Capabilities.CustomSizes.MaxEdge = 16
			}
			sc := firstImagesScenario(t)
			enqueue(ev, w, sc.replies(t)...)
			size := "1536x864"
			switch name {
			case "max-square":
				size = "2048x2048"
			case "max-landscape":
				size = "3840x2160"
			case "max-portrait":
				size = "2160x3840"
			}
			res, err := client.GenerateImages(ctxFor(t), textScope("req-edit-dimensions-"+name), sc.target(), imagesInput(t, sc), ai.OpenAIImagesOptions{Size: ai.Value(size)})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			if name == "model-stricter" {
				ev.Check("model bounds used", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest}) && len(w.provider.Requests()) == 0, "got %v", err)
			} else {
				ev.Check("source constraints independently frozen", err == nil, "got %v", err)
			}
		})
	}
}

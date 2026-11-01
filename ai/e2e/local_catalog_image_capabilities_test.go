package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/catalog"
)

// A declaration must be usable through the route's actual option/guard
// surface. Legacy explicit Host catalogs keep their historical Hash/Client
// acceptance; isolated local loading reports unavailable capabilities instead.
func TestLocalCatalogImageDefinitionCapabilities(t *testing.T) {
	var openai, google, xai ai.ImageModel
	for _, m := range ai.BuiltinCatalog().ImageModels {
		switch m.Provider {
		case ai.ProviderOpenAI:
			openai = m
		case ai.ProviderGoogle:
			google = m
		case ai.ProviderXAI:
			xai = m
		}
	}
	if openai.ID == "" || google.ID == "" || xai.ID == "" {
		t.Fatal("built-in image baseline identities unavailable")
	}
	modes := []string{"valid-openai", "valid-google", "valid-xai", "google-mask", "google-transparency", "google-custom-sizes", "google-fidelity", "google-sizes", "google-qualities", "google-reference-limit", "openai-aspect-ratios", "openai-image-sizes", "openai-quantity-limit", "openai-reference-limit", "openai-unknown-quality", "openai-unknown-fidelity", "invalid-output-limit"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CAT-T08-image-capability-"+mode)
			m := openai
			wantReason := catalog.UnsupportedDefinition
			wantError := ""
			switch mode {
			case "valid-openai":
				wantReason = ""
			case "valid-google":
				m, wantReason = google, ""
			case "valid-xai":
				m, wantReason = xai, ""
			case "google-mask":
				m = google
				m.Capabilities.Mask = true
			case "google-transparency":
				m = google
				m.Capabilities.TransparentBackground = true
			case "google-custom-sizes":
				m = google
				m.Capabilities.CustomSizes = &ai.ImageSizeConstraints{EdgeMultiple: 16, MaxAspectRatio: 3, MaxEdge: 4096, MinPixels: 65536, MaxPixels: 16777216, Source: "synthetic reviewed bounds"}
			case "google-fidelity":
				m = google
				m.Capabilities.InputFidelity = []string{"high"}
			case "google-sizes":
				m = google
				m.Capabilities.Sizes = []string{"1024x1024"}
			case "google-qualities":
				m = google
				m.Capabilities.Qualities = []string{"high"}
			case "google-reference-limit":
				m = google
				m.Capabilities.MaxReferenceImages = 15
			case "openai-aspect-ratios":
				m.Capabilities.AspectRatios = []string{"1:1"}
			case "openai-image-sizes":
				m.Capabilities.ImageSizes = []string{"1K"}
			case "openai-quantity-limit":
				m.Capabilities.MaxOutputImages = 11
			case "openai-reference-limit":
				m.Capabilities.MaxReferenceImages = 17
			case "openai-unknown-quality":
				m.Capabilities.Qualities = []string{"future-quality"}
			case "openai-unknown-fidelity":
				m.Capabilities.InputFidelity = []string{"future-fidelity"}
			case "invalid-output-limit":
				m.Capabilities.MaxOutputImages = 0
				wantReason = catalog.InvalidDefinition
				wantError = "ai: invalid config Catalog: image model " + m.ID + ": image capacity limits are invalid"
			}
			input := ai.Catalog{Version: "image-capability-taxonomy", Models: []ai.Model{}, ImageModels: []ai.ImageModel{m}}
			raw := localCatalogFile(string(mustMarshal(t, input)))
			effective, report, err := catalog.Load(raw)
			ev.Record("source", json.RawMessage(raw))
			ev.Record("report", report)
			ev.Check("image declaration failures are isolated", err == nil && len(report.Models) == 1, "err=%v report=%+v", err, report)
			if err != nil || len(report.Models) != 1 {
				t.FailNow()
			}
			ev.Check("only actual option and guard capabilities load", report.Models[0].Reason == wantReason, "got=%s want=%s", report.Models[0].Reason, wantReason)
			wantLoaded := 0
			if wantReason == "" {
				wantLoaded = 1
			}
			ev.Check("unsupported declarations never enter effective image index", report.Loaded == wantLoaded && report.Failed == 1-wantLoaded && len(effective.ImageModels) == wantLoaded, "report=%+v", report)
			_, hashErr := input.Hash()
			h := host.New()
			_, clientErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &input})
			ev.Check("legacy Hash contract unchanged", errString(hashErr) == wantError, "err=%v want=%s", hashErr, wantError)
			ev.Check("legacy Client acceptance unchanged", errString(clientErr) == wantError, "err=%v want=%s", clientErr, wantError)
		})
	}
}

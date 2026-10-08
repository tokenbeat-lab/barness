package e2e

import (
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGoogleImagesThinkingPriceRequired(t *testing.T) {
	for _, name := range []string{"unset", "null", "explicit-free"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P09-E11-thinking-price-"+name)
			w := newWorld(t, tenantA)
			cfg := ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host}
			configureGoogleImages(&cfg)
			// An image-only response format can still incur text thinking tokens.
			cfg.Catalog.ImageModels[0].Output = []ai.Modality{ai.ModalityImage}
			switch name {
			case "unset":
				cfg.Catalog.ImageModels[0].Pricing.OutputText = ai.Nullable[float64]{}
			case "null":
				cfg.Catalog.ImageModels[0].Pricing.OutputText = ai.Null[float64]()
			case "explicit-free":
				cfg.Catalog.ImageModels[0].Pricing.OutputText = ai.Value(0.0)
			}
			_, err := ai.NewClient(cfg)
			ev.Record("error", errString(err))
			if name == "explicit-free" {
				ev.Check("explicit zero price allowed", err == nil, "got %v", err)
			} else {
				ev.Check("thinking price cannot silently be missing", errors.Is(err, ai.ErrInvalidConfig), "got %v", err)
			}
		})
	}
}

package e2e

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// operationCatalog uses host-owned, synthetic models: no new vendor route is
// claimed as supported before its own implementation and live acceptance.
func operationCatalog() ai.Catalog {
	return ai.Catalog{
		Version: "host-operations-04",
		Models: []ai.Model{{
			Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ID: "gpt-4.1-mini", Name: "Host chat",
			Input: []ai.Modality{ai.ModalityText, ai.ModalityImage}, ContextWindow: 1047576, MaxTokens: 32768,
			SamplingParams: map[string]json.RawMessage{"top_p": json.RawMessage(`0.5`)},
			Cost: ai.ModelCost{CostRates: ai.CostRates{Input: 0.4, Output: 1.6, CacheRead: 0.1},
				Tiers: []ai.CostTier{{InputTokensAbove: 100, CostRates: ai.CostRates{Input: 2}}}},
		}},
		ImageModels: []ai.ImageModel{{
			Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ID: "gpt-4.1-mini", Name: "Host image",
			Input: []ai.Modality{ai.ModalityText, ai.ModalityImage}, Output: []ai.Modality{ai.ModalityText, ai.ModalityImage},
			Capabilities: ai.ImageCapabilities{MaxReferenceImages: 2, MaxOutputImages: 1, Mask: true,
				Sizes: []string{"auto"}, ImageSizes: []string{"1K"}, AspectRatios: []string{"1:1"}, InputFidelity: []string{"high"}},
			Pricing: ai.ImagePricing{InputText: ai.Value(1.0), InputImage: ai.Value(2.0),
				OutputText: ai.Value(3.0), OutputImage: ai.Value(4.0), Source: "synthetic fixture; 2026-10-08"},
		}},
		ClassifierModels: []ai.ClassifierModel{{
			Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ID: "gpt-4.1-mini", Name: "Host classifier", ContextWindow: 32768,
			Capabilities: ai.ClassifierCapabilities{Kinds: []string{"choice", "score", "bool"}, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10},
			Cost:         ai.ModelCost{CostRates: ai.CostRates{Input: 1}, Tiers: []ai.CostTier{{InputTokensAbove: 100, CostRates: ai.CostRates{Input: 2}}}},
		}},
	}
}

func TestOperationCatalogValidation(t *testing.T) {
	cases := map[string]func(*ai.Catalog){
		"duplicate-chat":                       func(c *ai.Catalog) { c.Models = append(c.Models, c.Models[0]) },
		"duplicate-image":                      func(c *ai.Catalog) { c.ImageModels = append(c.ImageModels, c.ImageModels[0]) },
		"duplicate-classifier":                 func(c *ai.Catalog) { c.ClassifierModels = append(c.ClassifierModels, c.ClassifierModels[0]) },
		"chat-unknown-modality":                func(c *ai.Catalog) { c.Models[0].Input = []ai.Modality{"audio"} },
		"chat-reserved-sampling":               func(c *ai.Catalog) { c.Models[0].SamplingParams["model"] = json.RawMessage(`"other"`) },
		"chat-invalid-sampling":                func(c *ai.Catalog) { c.Models[0].SamplingParams["top_p"] = json.RawMessage(`{`) },
		"image-unknown-input":                  func(c *ai.Catalog) { c.ImageModels[0].Input = []ai.Modality{"audio"} },
		"image-unknown-output":                 func(c *ai.Catalog) { c.ImageModels[0].Output = append(c.ImageModels[0].Output, "video") },
		"image-no-output-image":                func(c *ai.Catalog) { c.ImageModels[0].Output = []ai.Modality{ai.ModalityText} },
		"image-duplicate-modality":             func(c *ai.Catalog) { c.ImageModels[0].Input = append(c.ImageModels[0].Input, ai.ModalityImage) },
		"image-no-input":                       func(c *ai.Catalog) { c.ImageModels[0].Input = nil },
		"image-references-without-image-input": func(c *ai.Catalog) { c.ImageModels[0].Input = []ai.Modality{ai.ModalityText} },
		"image-negative-reference-limit":       func(c *ai.Catalog) { c.ImageModels[0].Capabilities.MaxReferenceImages = -1 },
		"image-zero-output-limit":              func(c *ai.Catalog) { c.ImageModels[0].Capabilities.MaxOutputImages = 0 },
		"image-mask-without-references":        func(c *ai.Catalog) { c.ImageModels[0].Capabilities.MaxReferenceImages = 0 },
		"classifier-no-kinds":                  func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.Kinds = nil },
		"classifier-unknown-kind":              func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.Kinds[0] = "noul" },
		"classifier-duplicate-kind": func(c *ai.Catalog) {
			c.ClassifierModels[0].Capabilities.Kinds = append(c.ClassifierModels[0].Capabilities.Kinds, "bool")
		},
		"classifier-choice-limit":       func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.MaxChoices = 1 },
		"classifier-score-min":          func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.MinScoreLevels = 1 },
		"classifier-score-range":        func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.MaxScoreLevels = 1 },
		"classifier-range-without-kind": func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.Kinds = []string{"bool"} },
		"classifier-zero-context":       func(c *ai.Catalog) { c.ClassifierModels[0].ContextWindow = 0 },
		"classifier-negative-price":     func(c *ai.Catalog) { c.ClassifierModels[0].Cost.Input = -1 },
		"classifier-nonfinite-tier":     func(c *ai.Catalog) { c.ClassifierModels[0].Cost.Tiers[0].Output = math.Inf(1) },
	}
	for _, field := range []string{"inputText", "inputImage", "outputText", "outputImage", "cacheReadText", "cacheReadImage"} {
		for _, value := range []struct {
			name string
			rate ai.Nullable[float64]
		}{
			{"negative", ai.Value(-1.0)}, {"nan", ai.Value(math.NaN())}, {"infinite", ai.Value(math.Inf(1))},
		} {
			cases["image-"+field+"-"+value.name] = func(c *ai.Catalog) { setImageRate(&c.ImageModels[0].Pricing, field, value.rate) }
		}
		if field == "cacheReadText" || field == "cacheReadImage" {
			continue
		}
		for _, value := range []struct {
			name string
			rate ai.Nullable[float64]
		}{
			{"missing", ai.Nullable[float64]{}}, {"null", ai.Null[float64]()},
		} {
			cases["image-"+field+"-"+value.name] = func(c *ai.Catalog) { setImageRate(&c.ImageModels[0].Pricing, field, value.rate) }
		}
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "E03-operation-catalog-refuse-"+name)
			catalog := operationCatalog()
			change(&catalog)
			h := host.New()
			client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &catalog})
			// Invalid non-finite JSON cannot be an evidence record; record only
			// the public construction error and the independent scenario name.
			ev.Record("construction-error", errString(err))
			var ce *ai.ConfigError
			ev.Check("invalid directory is refused at construction", client == nil && errors.As(err, &ce) && ce.Field == "Catalog" && errors.Is(err, ai.ErrInvalidConfig), "err=%v", err)
			_, hashErr := catalog.Hash()
			ev.Check("Hash refuses the same invalid catalog", errors.Is(hashErr, ai.ErrInvalidConfig), "err=%v", hashErr)
		})
	}
}

func setImageRate(p *ai.ImagePricing, field string, value ai.Nullable[float64]) {
	switch field {
	case "inputText":
		p.InputText = value
	case "inputImage":
		p.InputImage = value
	case "outputText":
		p.OutputText = value
	case "outputImage":
		p.OutputImage = value
	case "cacheReadText":
		p.CacheReadText = value
	case "cacheReadImage":
		p.CacheReadImage = value
	}
}

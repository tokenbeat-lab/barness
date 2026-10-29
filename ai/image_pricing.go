package ai

import (
	"math"
	"slices"
	"strings"
)

// ImagePerImagePricing expresses USD per input/output image. The output
// table uses explicit quality and resolution tiers; auto is a selection
// strategy, never a fixed price tier. Source is on the parent ImagePricing.
type ImagePerImagePricing struct {
	InputImage Nullable[float64]  `json:"inputImage,omitzero"`
	Output     []ImageOutputPrice `json:"output"`
}

// ImageOutputPrice is one quality/resolution's USD per generated image.
type ImageOutputPrice struct {
	Quality   string            `json:"quality"`
	ImageSize string            `json:"imageSize"`
	Price     Nullable[float64] `json:"price,omitzero"`
}

func xaiImageModelProblem(m ImageModel) string {
	if m.Provider != ProviderXAI {
		return ""
	}
	if m.API != APIXAIImages || m.ID != "grok-imagine-image-2.0" {
		return "xAI image identity is outside the implemented subset"
	}
	c := m.Capabilities
	if !slices.Contains(m.Input, ModalityText) || len(m.Output) != 1 || m.Output[0] != ModalityImage {
		return "xAI images require text input and only image output"
	}
	if c.MaxReferenceImages > 5 || c.MaxOutputImages > 10 {
		return "xAI image capacity exceeds the implemented subset"
	}
	if len(c.Sizes) > 0 || c.CustomSizes != nil || c.Mask || c.TransparentBackground || len(c.InputFidelity) > 0 {
		return "xAI image capabilities include unsupported options"
	}
	validList := func(values, allowed []string) bool {
		seen := map[string]bool{}
		for _, v := range values {
			if seen[v] || !slices.Contains(allowed, v) {
				return false
			}
			seen[v] = true
		}
		return true
	}
	if !validList(c.ImageSizes, []string{"1k", "1.5k", "2k"}) ||
		!validList(c.Qualities, []string{"low", "medium", "auto"}) ||
		!validList(c.AspectRatios, []string{"1:1", "3:4", "4:3", "9:16", "16:9", "2:3", "3:2", "9:19.5", "19.5:9", "9:20", "20:9", "1:2", "2:1", "21:9", "5:2", "auto"}) {
		return "xAI image options exceed the implemented subset or contain duplicates"
	}
	if m.Pricing.PerImage == nil {
		return "xAI images require per-image pricing"
	}
	return ""
}

func (p ImagePricing) perImageProblem(m ImageModel) string {
	// Other adapters still calculate token prices. Accepting a different price
	// unit there would silently price known consumption as zero.
	if m.Provider != ProviderXAI || m.API != APIXAIImages {
		return "the image protocol does not implement per-image pricing"
	}
	for _, rate := range []Nullable[float64]{p.InputText, p.InputImage, p.OutputText, p.OutputImage, p.CacheReadText, p.CacheReadImage} {
		if !rate.IsZero() {
			return "per-image and token image prices are mutually exclusive"
		}
	}
	if strings.TrimSpace(p.Source) == "" {
		return "per-image pricing requires a source"
	}
	validRate := func(rate Nullable[float64]) bool {
		value, ok := rate.Get()
		return ok && value >= 0 && !math.IsNaN(value) && !math.IsInf(value, 0)
	}
	if slices.Contains(m.Input, ModalityImage) || !p.PerImage.InputImage.IsZero() {
		if !validRate(p.PerImage.InputImage) {
			return "input image price is missing, negative or not finite"
		}
	}
	if len(p.PerImage.Output) == 0 {
		return "per-image pricing has no output rates"
	}
	type tier struct{ quality, size string }
	seen := map[tier]bool{}
	for _, row := range p.PerImage.Output {
		key := tier{row.Quality, row.ImageSize}
		if seen[key] {
			return "a per-image output rate is declared more than once"
		}
		seen[key] = true
		if (row.Quality != "low" && row.Quality != "medium") || !slices.Contains([]string{"1k", "1.5k", "2k"}, row.ImageSize) || !validRate(row.Price) {
			return "per-image output rate has an unsupported tier or missing, negative or non-finite price"
		}
	}
	qualities := m.Capabilities.Qualities
	if slices.Contains(qualities, "auto") {
		qualities = []string{"low", "medium"}
	}
	for _, quality := range qualities {
		for _, size := range m.Capabilities.ImageSizes {
			if !seen[tier{quality, size}] {
				return "an authorized image quality and resolution has no price"
			}
		}
	}
	return ""
}

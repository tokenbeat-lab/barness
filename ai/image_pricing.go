package ai

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
	return ""
}

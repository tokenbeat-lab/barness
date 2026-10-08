package ai

// The fixed Sunburst snapshot is included after this route's own JSON edit,
// generation and mask live acceptance on 2026-10-08 (ADR-0022 / issue 11).
// Its first release deliberately claims only the size, counts and qualities
// exercised by that smoke, even though the official API permits larger values.
// Sources and retrieval date are pinned in testdata/openai-images/official.json.
// Input fidelity remains omitted: Sunburst's values are not documented/tested.
func builtinOpenAIImageModels() []ImageModel {
	return []ImageModel{{
		Provider: ProviderOpenAI, API: APIOpenAIImages, ID: "gpt-image-2.5-sunburst-2026-09-08", Name: "GPT Image 2.5 Sunburst (2026-09-08)",
		Input: []Modality{ModalityText, ModalityImage}, Output: []Modality{ModalityImage},
		Capabilities: ImageCapabilities{MaxReferenceImages: 1, MaxOutputImages: 1, Sizes: []string{"1024x1024"}, Qualities: []string{"low", "medium"}, Mask: true, TransparentBackground: true},
		Pricing:      ImagePricing{InputText: Value(5.0), InputImage: Value(8.0), OutputText: Value(0.0), OutputImage: Value(30.0), Source: "https://developers.openai.com/api/docs/guides/image-generation; https://developers.openai.com/api/docs/models/gpt-image-2.5-sunburst; retrieved 2026-10-08"},
	}}
}

// Only this route's own generation/reference-edit PASS permits inclusion.
// Narrow to one reference/output, 1K and 1:1, as exercised on 2026-10-08.
// Documented broader capabilities and source conflicts stay in official.json.
// OutputText is required even for image-only responses: thought is billed text.
func builtinGoogleImageModels() []ImageModel {
	return []ImageModel{{
		Provider: ProviderGoogle, API: APIGoogleInteractions, ID: "gemini-nano-banana-2.1", Name: "Gemini Nano Banana 2.1",
		Input: []Modality{ModalityText, ModalityImage}, Output: []Modality{ModalityText, ModalityImage},
		Capabilities: ImageCapabilities{MaxReferenceImages: 1, MaxOutputImages: 1, ImageSizes: []string{"1K"}, AspectRatios: []string{"1:1"}},
		Pricing:      ImagePricing{InputText: Value(1.5), InputImage: Value(1.5), OutputText: Value(7.5), OutputImage: Value(30.0), Source: "https://ai.google.dev/gemini-api/docs/pricing?hl=en; retrieved 2026-10-08"},
	}}
}

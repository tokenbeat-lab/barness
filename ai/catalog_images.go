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

// These bounds are the route's own six real candidate captures on 2026-10-29:
// generation, ordered edits, five references, two outputs and the listed tiers.
// Source and hashes are pinned in e2e/testdata/xai-images/{official,live-contract}.json.
// This release literal intentionally duplicates the candidate's prices: the
// candidate keeps its original version/hash and broader limits independently.
// Listing this identity does not rewrite its candidate report as formal proof.
func builtinXAIImageModels() []ImageModel {
	return []ImageModel{{
		Provider: ProviderXAI, API: APIXAIImages, ID: "grok-imagine-image-2.0", Name: "Grok Imagine Image 2.0",
		Input: []Modality{ModalityText, ModalityImage}, Output: []Modality{ModalityImage},
		Capabilities: ImageCapabilities{MaxReferenceImages: 5, MaxOutputImages: 2, ImageSizes: []string{"1k", "1.5k", "2k"}, Qualities: []string{"low", "medium", "auto"}, AspectRatios: []string{"1:1", "16:9"}},
		Pricing: ImagePricing{Source: "https://docs.x.ai/developers/pricing; retrieved 2026-10-29", PerImage: &ImagePerImagePricing{
			InputImage: Value(0.01), Output: []ImageOutputPrice{
				{Quality: "low", ImageSize: "1k", Price: Value(0.04)},
				{Quality: "low", ImageSize: "1.5k", Price: Value(0.05)},
				{Quality: "low", ImageSize: "2k", Price: Value(0.06)},
				{Quality: "medium", ImageSize: "1k", Price: Value(0.06)},
				{Quality: "medium", ImageSize: "1.5k", Price: Value(0.07)},
				{Quality: "medium", ImageSize: "2k", Price: Value(0.08)},
			},
		}},
	}}
}

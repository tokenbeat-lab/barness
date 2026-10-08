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

package ai

// The candidate's own nine-scenario live and clean integrity/audit accepted
// these facts on 2026-10-13 (issue 09; sources.json pin 61be0c194e7231aa).
// Context is frozen pi-ai 1.1.0 classifier metadata, not chat context or a
// measured service limit. Score 2–10 and PNG/JPEG/WebP are supported subsets.
// Decisions guide input-only $0.10/M and Luna's >272000 whole-request 2x
// multiplier yield $0.20/M; chat output/cache prices do not apply. See
// ai/live/testdata/decisions/{sources.json,guide.md,model.md,reference.txt}.
// The undated alias's weights and regional premiums are not frozen by Hash.
func builtinDecisionsModels() []ClassifierModel {
	return []ClassifierModel{{
		Provider: ProviderOpenAI, API: APIOpenAIDecisions, ID: "gpt-6-luna", Name: "GPT-6 Luna", ContextWindow: 922000,
		Input: []Modality{ModalityText, ModalityImage},
		Capabilities: ClassifierCapabilities{
			Kinds:      []ClassifierQuestionKind{ClassifierQuestionChoice, ClassifierQuestionScore, ClassifierQuestionBool},
			MinChoices: 2, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10,
			DescriptionFormats: []ClassifierDescriptionFormat{ClassifierDescriptionString},
			MaxInputImages:     128, ImageMIMETypes: []string{"image/png", "image/jpeg", "image/webp"},
		},
		Cost: ModelCost{CostRates: CostRates{Input: 0.1}, Tiers: []CostTier{{InputTokensAbove: 272000, CostRates: CostRates{Input: 0.2}}}},
	}}
}

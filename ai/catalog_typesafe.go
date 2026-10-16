package ai

// Only fixed releases are included. Moving aliases would change decisions
// and pricing without changing the host's catalog snapshot. Rates and limits
// were read from https://docs.typesafe.ai/models and /api on 2026-10-08;
// official fixtures and this route's own live evidence are referenced by ADR-0021.
func builtinTypeSafeModels() []ClassifierModel {
	return []ClassifierModel{{
		Provider: ProviderTypeSafe, API: APITypeSafeSystemOne, ID: "jev-1.13.0", Name: "Jev", ContextWindow: 64000,
		Capabilities: ClassifierCapabilities{Kinds: []ClassifierQuestionKind{ClassifierQuestionChoice, ClassifierQuestionScore, ClassifierQuestionBool}, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10},
		Cost:         ModelCost{CostRates: CostRates{Input: 0.042, Output: 0}},
	}, {
		// Images/Box remain experimental even in the fixed catalog. Sources:
		// https://docs.typesafe.ai/models, /images and /primitives/box, read
		// 2026-10-17; own candidate/effect evidence: ticket 13 admission receipt.
		// 24 sources is our bounded supported subset, not a vendor count limit.
		Provider: ProviderTypeSafe, API: APITypeSafeSystemOne, ID: "jev-1.14.0", Name: "Jev 1.14", ContextWindow: 64000,
		Input: []Modality{ModalityText, ModalityImage},
		Capabilities: ClassifierCapabilities{Kinds: []ClassifierQuestionKind{ClassifierQuestionChoice, ClassifierQuestionScore, ClassifierQuestionBool, ClassifierQuestionBox}, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10,
			MaxInputImages: 24, ImageMIMETypes: []string{"image/png", "image/jpeg"}, ImageLocations: []ClassifierImageLocation{ClassifierImageLocationContext, ClassifierImageLocationState, ClassifierImageLocationInstructions, ClassifierImageLocationCriteria}},
		Cost: ModelCost{CostRates: CostRates{Input: 0.042, Output: 0}},
	}}
}

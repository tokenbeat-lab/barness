package ai

// Only the fixed release is included. Moving aliases would change decisions
// and pricing without changing the host's catalog snapshot. Rates and limits
// were read from https://docs.typesafe.ai/models and /api on 2026-10-08;
// official fixtures and this route's own live evidence are referenced by ADR-0021.
func builtinTypeSafeModels() []ClassifierModel {
	return []ClassifierModel{{
		Provider: ProviderTypeSafe, API: APITypeSafeSystemOne, ID: "jev-1.13.0", Name: "Jev", ContextWindow: 64000,
		Capabilities: ClassifierCapabilities{Kinds: []ClassifierQuestionKind{ClassifierQuestionChoice, ClassifierQuestionScore, ClassifierQuestionBool}, MaxChoices: 255, MinScoreLevels: 2, MaxScoreLevels: 10},
		Cost:         ModelCost{CostRates: CostRates{Input: 0.042, Output: 0}},
	}}
}

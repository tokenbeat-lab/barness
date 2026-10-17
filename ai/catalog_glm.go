package ai

// builtinGLMModels publishes six independent China standard API identities.
// Source and USD/capacity bounds: ADR-0036 and live/testdata/glm-candidate/
// sources.json (fixed CNY8/28/2 perM, CFETS2026-10-16 USD/CNY6.7220).
// Cache-write maps new-content input computation; estimates are not invoices.
// Repetition is deliberate: models and protocols evolve independently.
func builtinGLMModels() []Model {
	return []Model{
		{Provider: ProviderZhipu, API: APIOpenAICompletions, ID: "glm-5.3", Name: "GLM-5.3 Chat", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 131072,
			ThinkingLevelMap: ThinkingLevelMap{Off: Null[string](), Minimal: Null[string](), Low: Value("low"), Medium: Null[string](), High: Value("high"), XHigh: Null[string](), Max: Value("max")},
			Compat:           ModelCompat{RequiresThinking: true, SupportsLongCacheRetention: Value(false)}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
		{Provider: ProviderZhipu, API: APIOpenAICompletions, ID: "glm-5.2", Name: "GLM-5.2 Chat", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 131072,
			ThinkingLevelMap: ThinkingLevelMap{Off: Value("none"), Minimal: Value("minimal"), Low: Value("high"), Medium: Value("high"), High: Value("high"), XHigh: Value("max"), Max: Value("max")},
			Compat:           ModelCompat{SupportsLongCacheRetention: Value(false)}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
		{Provider: ProviderZhipu, API: APIOpenAIResponses, ID: "glm-5.3", Name: "GLM-5.3 Responses", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 131072,
			ThinkingLevelMap: ThinkingLevelMap{Off: Null[string](), Minimal: Null[string](), Low: Value("low"), Medium: Null[string](), High: Value("high"), XHigh: Null[string](), Max: Value("max")},
			Compat:           ModelCompat{RequiresThinking: true, SupportsLongCacheRetention: Value(false)}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
		{Provider: ProviderZhipu, API: APIOpenAIResponses, ID: "glm-5.2", Name: "GLM-5.2 Responses", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 131072,
			ThinkingLevelMap: ThinkingLevelMap{Off: Value("none"), Minimal: Value("minimal"), Low: Value("high"), Medium: Value("high"), High: Value("high"), XHigh: Value("max"), Max: Value("max")},
			Compat:           ModelCompat{SupportsLongCacheRetention: Value(false)}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
		{Provider: ProviderZhipu, API: APIAnthropicMessages, ID: "glm-5.3", Name: "GLM-5.3 Messages", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 128000,
			ThinkingLevelMap: ThinkingLevelMap{Off: Null[string](), Minimal: Null[string](), Low: Null[string](), Medium: Null[string](), High: Null[string](), XHigh: Null[string](), Max: Null[string]()},
			Compat:           ModelCompat{SupportsTemperature: Value(false), SupportsLongCacheRetention: Value(false)}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
		{Provider: ProviderZhipu, API: APIAnthropicMessages, ID: "glm-5.2", Name: "GLM-5.2 Messages", Reasoning: true, Input: []Modality{ModalityText}, ContextWindow: 1000000, MaxTokens: 128000,
			ThinkingLevelMap: ThinkingLevelMap{Off: Null[string](), Minimal: Null[string](), Low: Null[string](), Medium: Null[string](), High: Null[string](), XHigh: Null[string](), Max: Null[string]()},
			Compat:           ModelCompat{SupportsTemperature: Value(false), SupportsLongCacheRetention: Value(false), MessagesThinkingBudgetTokens: 1024}, Cost: ModelCost{CostRates: CostRates{Input: 1.1901219875, Output: 4.1654269563, CacheRead: .2975304969, CacheWrite: 1.1901219875}}},
	}
}

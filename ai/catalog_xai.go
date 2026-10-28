package ai

import "slices"

// xAI metadata may narrow the initial Grok Responses subset. Catalog data is
// Host input: it cannot enable another wire effort or provider capability.
func xaiModelProblem(m Model) string {
	if m.Provider != ProviderXAI {
		return ""
	}
	if m.ID != "grok-4.7" || m.API != APIOpenAIResponses {
		return "xAI model or protocol is outside the implemented subset"
	}
	if !slices.Contains(m.Input, ModalityText) || len(m.Input) > 2 || m.ContextWindow <= 0 || m.ContextWindow > 500000 || m.MaxTokens <= 0 || m.MaxTokens > 128000 {
		return "xAI capacity is outside the implemented subset"
	}
	if !m.Reasoning || !m.Compat.RequiresThinking || m.DefaultThinkingLevel != ThinkingHigh {
		return "Grok 4.7 requires reasoning and the high default"
	}
	if high, ok := m.ThinkingLevelMap.High.Get(); !ok || high != "high" {
		return "Grok 4.7 must retain high effort"
	}
	for _, level := range modelThinkingLevels {
		entry := m.ThinkingLevelMap.entry(level)
		value, ok := entry.Get()
		switch level {
		case ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh:
			if !entry.IsNull() && (!ok || value != string(level)) {
				return "xAI effort mapping is outside the implemented subset"
			}
		default:
			if !entry.IsNull() {
				return "unsupported xAI effort must be disabled"
			}
		}
	}
	c := m.Compat
	if c.ThinkingSwitchOnly || c.SupportsStrictMode || c.SupportsExplicitPromptCacheMode || c.ForceAdaptiveThinking || c.SupportsMidConvoEffort || c.MessagesThinkingBudgetTokens != 0 {
		return "xAI compatibility flags cannot expand the implemented subset"
	}
	if value, ok := c.SupportsLongCacheRetention.Get(); ok && value {
		return "xAI cache retention duration is not implemented"
	}
	return ""
}

// builtinXAICatalog is the deliberately prepared final snapshot on the
// unreleased integration branch (ADR-0054). Candidate JSON stays independent:
// its historical qualification and version must not follow Builtin changes.
func builtinXAICatalog() Catalog {
	return Catalog{Version: "2026-10-26.1", Models: []Model{{
		Provider: ProviderXAI, API: APIOpenAIResponses, ID: "grok-4.7", Name: "Grok 4.7", Reasoning: true,
		Input: []Modality{ModalityText, ModalityImage}, ContextWindow: 500000, MaxTokens: 128000,
		ThinkingLevelMap:     ThinkingLevelMap{Off: Null[string](), Minimal: Null[string](), Low: Value("low"), Medium: Value("medium"), High: Value("high"), XHigh: Value("xhigh"), Max: Null[string]()},
		DefaultThinkingLevel: ThinkingHigh,
		Compat:               ModelCompat{RequiresThinking: true, SupportsMidConvoSystemMessages: true, SupportsLongCacheRetention: Value(false)},
		Cost:                 ModelCost{CostRates: CostRates{Input: 2, Output: 6, CacheRead: .5}, Tiers: []CostTier{{InputTokensAbove: 199999, CostRates: CostRates{Input: 4, Output: 12, CacheRead: 1}}}},
	}}}
}

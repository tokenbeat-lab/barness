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

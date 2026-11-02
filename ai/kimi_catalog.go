package ai

func isKimi(p ProviderID) bool { return p == ProviderMoonshot || p == ProviderMoonshotCN }

// Catalogs may name Kimi models within the implemented text subset. Metadata
// cannot expand the protocol, modality, effort shape or pricing tier.
func kimiModelProblem(m Model) string {
	if !isKimi(m.Provider) {
		return ""
	}
	if m.API != APIOpenAICompletions && m.API != APIOpenAIResponses && m.API != APIAnthropicMessages {
		return "Kimi protocol is outside the implemented subset"
	}
	if len(m.Input) != 1 || m.Input[0] != ModalityText || m.ContextWindow <= 0 || m.ContextWindow > 1048576 || m.MaxTokens <= 0 || m.MaxTokens > 131072 {
		return "Kimi text capacity is outside the implemented subset"
	}
	if !m.Reasoning || !m.Compat.RequiresThinking || m.DefaultThinkingLevel != ThinkingMax {
		return "Kimi text models require reasoning and the max default"
	}
	if value, ok := m.ThinkingLevelMap.Max.Get(); !ok || value != "max" {
		return "Kimi text models must retain max effort"
	}
	for _, l := range modelThinkingLevels {
		e := m.ThinkingLevelMap.entry(l)
		value, ok := e.Get()
		if l == ThinkingLow || l == ThinkingHigh || l == ThinkingMax {
			if !e.IsNull() && (!ok || value != string(l)) {
				return "Kimi effort mapping is outside the implemented subset"
			}
		} else if !e.IsNull() {
			return "Kimi unsupported effort must be disabled"
		}
	}
	c := m.Compat
	if c.ThinkingSwitchOnly || c.SupportsStrictMode || c.SupportsMidConvoSystemMessages || c.SupportsExplicitPromptCacheMode || c.ForceAdaptiveThinking || c.SupportsMidConvoEffort || c.MessagesThinkingBudgetTokens != 0 {
		return "Kimi compatibility flags cannot expand the implemented subset"
	}
	if m.API == APIAnthropicMessages && (len(m.SamplingParams) > 0 || len(m.SamplingParamsByThinkingLevel) > 0) {
		return "Kimi Messages sampling defaults are not implemented"
	}
	if len(m.Cost.Tiers) > 0 {
		return "Kimi pricing tiers are not implemented"
	}
	for _, rate := range m.Cost.CostRates.list() {
		if rate <= 0 {
			return "Kimi pricing requires all four positive USD rates"
		}
	}
	return ""
}

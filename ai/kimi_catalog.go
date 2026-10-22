package ai

func isKimi(p ProviderID) bool { return p == ProviderMoonshot || p == ProviderMoonshotCN }

// Kimi metadata may narrow the implemented K3 subset. It cannot grant another
// model, protocol, modality, effort shape or pricing tier through compat flags.
func kimiModelProblem(m Model) string {
	if !isKimi(m.Provider) {
		return ""
	}
	if m.ID != "kimi-k3" || (m.API != APIOpenAICompletions && m.API != APIOpenAIResponses && m.API != APIAnthropicMessages) {
		return "Kimi model or protocol is outside the implemented subset"
	}
	if len(m.Input) != 1 || m.Input[0] != ModalityText || m.ContextWindow <= 0 || m.ContextWindow > 1048576 || m.MaxTokens <= 0 || m.MaxTokens > 131072 {
		return "Kimi text capacity is outside the implemented subset"
	}
	if !m.Reasoning {
		return "Kimi K3 requires reasoning"
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

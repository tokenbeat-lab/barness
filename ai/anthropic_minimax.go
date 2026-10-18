package ai

import (
	"slices"
)

// MiniMax's Messages adaptive switch and effort are independent of Claude's
// native display, budget, beta and managed per-turn effort (spec R15).
// Validation runs after authorization and before reading a credential.
func validateMiniMaxMessagesOptions(m Model, options Options) *Error {
	if m.Provider != ProviderMiniMax || m.API != APIAnthropicMessages {
		return nil
	}
	o, _ := options.(AnthropicOptions)
	var problem string
	switch {
	case o.ThinkingBudgetTokens != 0, o.ThinkingDisplay != "", !o.InterleavedThinking.IsZero() && !o.InterleavedThinking.IsNull():
		problem = "MiniMax Messages native thinking extensions are not supported"
	case !m.Reasoning && (!o.ThinkingEnabled.IsZero() || o.Effort != ""):
		problem = "model does not support thinking"
	}
	if enabled, explicit := o.ThinkingEnabled.Get(); explicit && !enabled && m.Compat.RequiresThinking {
		problem = "model requires thinking"
	}
	if o.Effort != "" {
		if !miniMaxMessagesHasEffort(m) || !miniMaxMessagesEffort(m, string(o.Effort)) {
			problem = "model does not support the requested Messages effort"
		}
	}
	if n, present := o.MaxTokens.Get(); present && (n < 1 || n > m.MaxTokens) {
		problem = "max_tokens must be an integer from 1 through the model output limit"
	}
	if supported, declared := m.Compat.SupportsTemperature.Get(); declared && !supported && !o.Temperature.IsZero() {
		problem = "model does not support temperature"
	}
	if n, present := o.Temperature.Get(); present && (n < 0 || n > 2) {
		problem = "temperature must be from 0 through 2"
	}
	if c, present := o.ToolChoice.Get(); present && (c.Tool != "" || !slices.Contains([]string{"auto", "none"}, c.Mode)) {
		problem = "MiniMax Messages supports only auto or none tool choice"
	}
	if problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	return nil
}

func miniMaxMessagesEffort(m Model, effort string) bool {
	level := ThinkingLevel(effort)
	return AnthropicEffort(effort).valid() && slices.Contains(supportedThinkingLevels(m), level)
}

// ThinkingSwitchOnly is independent of mandatory thinking and of which
// depths a trusted Host retains in the model's supported map.
func miniMaxMessagesHasEffort(m Model) bool {
	return m.Reasoning && !m.Compat.ThinkingSwitchOnly
}

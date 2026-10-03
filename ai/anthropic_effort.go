package ai

import "cmp"

// Managed mid-conversation effort (Model.Compat.SupportsMidConvoEffort,
// ADR-0019): the Anthropic vendor manages the effort per turn, so each turn
// records the effort it ran at and the replayed conversation names it.

// pi's MID_CONVERSATION_OUTPUT_CONFIG_BETA and
// THINKING_BINDING_CONTROLS_BETA, which managed effort needs.
const (
	anthropicMidConvoOutputConfigBeta   = "mid-conversation-output-config-2026-07-01"
	anthropicThinkingBindingControlBeta = "thinking-binding-controls-2026-08-01"
)

// anthropicBlockBinding is how the vendor treats thinking blocks bound to a
// prefix that no longer matches; a managed-effort model drops them, as pi
// does, instead of refusing the request.
type anthropicBlockBinding struct {
	PrefixMismatchBehavior string `json:"prefix_mismatch_behavior"`
}

// anthropicThinkingLevel is pi's providerThinkingLevel: the effort a
// managed-effort model's turn runs at, the effort option or high; empty for
// any other model.
func anthropicThinkingLevel(m Model, o AnthropicOptions) AnthropicEffort {
	if !m.Compat.SupportsMidConvoEffort {
		return ""
	}
	return cmp.Or(o.Effort, AnthropicEffortHigh)
}

// withThinkingLevels is pi's insertThinkingLevelMessages: an empty system
// message with each recorded effort goes right before its assistant turn,
// and one with the call's effort ends the conversation. It runs after the
// cache marker was placed, which therefore stays on the last real turn.
func withThinkingLevels(msgs []anthropicMessage, levels map[int]AnthropicEffort, active AnthropicEffort) []anthropicMessage {
	marker := func(e AnthropicEffort) anthropicMessage {
		return anthropicMessage{Role: "system", Content: []any{}, OutputConfig: &anthropicOutputConfig{Effort: e}}
	}
	out := make([]anthropicMessage, 0, len(msgs)+len(levels)+1)
	for i, m := range msgs {
		if e, ok := levels[i]; ok {
			out = append(out, marker(e))
		}
		out = append(out, m)
	}
	return append(out, marker(active))
}

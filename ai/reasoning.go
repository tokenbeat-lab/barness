package ai

import "slices"

// ThinkingLevel is a protocol-neutral reasoning level, as in pi-ai. "off" is
// not one: it is a model capability and mapping value, never a simple input
// (spec I7).
type ThinkingLevel string

const (
	ThinkingMinimal ThinkingLevel = "minimal"
	ThinkingLow     ThinkingLevel = "low"
	ThinkingMedium  ThinkingLevel = "medium"
	ThinkingHigh    ThinkingLevel = "high"
	ThinkingXHigh   ThinkingLevel = "xhigh"
	ThinkingMax     ThinkingLevel = "max"
)

// thinkingOff is pi's model-level "off". It only appears as a clamp result.
const thinkingOff ThinkingLevel = "off"

// modelThinkingLevels are pi's EXTENDED_THINKING_LEVELS in ascending order;
// clamping searches this order.
var modelThinkingLevels = []ThinkingLevel{thinkingOff, ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh, ThinkingXHigh, ThinkingMax}

func (l ThinkingLevel) valid() bool {
	return l != thinkingOff && slices.Contains(modelThinkingLevels, l)
}

// ThinkingLevelMap is a reasoning model's level mapping (pi-ai's
// thinkingLevelMap). Each entry has three states: unset keeps pi's default,
// Null marks the level unsupported, and a value is the provider's wire value
// for the level. xhigh and max are only supported when mapped to a value.
type ThinkingLevelMap struct {
	Off     Nullable[string] `json:"off,omitzero"`
	Minimal Nullable[string] `json:"minimal,omitzero"`
	Low     Nullable[string] `json:"low,omitzero"`
	Medium  Nullable[string] `json:"medium,omitzero"`
	High    Nullable[string] `json:"high,omitzero"`
	XHigh   Nullable[string] `json:"xhigh,omitzero"`
	Max     Nullable[string] `json:"max,omitzero"`
}

func (m ThinkingLevelMap) entry(l ThinkingLevel) Nullable[string] {
	switch l {
	case thinkingOff:
		return m.Off
	case ThinkingMinimal:
		return m.Minimal
	case ThinkingLow:
		return m.Low
	case ThinkingMedium:
		return m.Medium
	case ThinkingHigh:
		return m.High
	case ThinkingXHigh:
		return m.XHigh
	case ThinkingMax:
		return m.Max
	}
	return Nullable[string]{}
}

// wireValue is pi's `thinkingLevelMap?.[l] ?? l`: the mapped value, or the
// level itself when the entry is unset or null.
func (m ThinkingLevelMap) wireValue(l ThinkingLevel) string {
	if v, ok := m.entry(l).Get(); ok {
		return v
	}
	return string(l)
}

// supportedThinkingLevels ports pi's getSupportedThinkingLevels: a model
// without reasoning supports only off; otherwise every level not mapped to
// null, where xhigh and max also need a mapping.
func supportedThinkingLevels(m Model) []ThinkingLevel {
	if !m.Reasoning {
		return []ThinkingLevel{thinkingOff}
	}
	var out []ThinkingLevel
	for _, l := range modelThinkingLevels {
		e := m.ThinkingLevelMap.entry(l)
		if e.IsNull() {
			continue
		}
		if _, mapped := e.Get(); (l == ThinkingXHigh || l == ThinkingMax) && !mapped {
			continue
		}
		out = append(out, l)
	}
	return out
}

// clampThinkingLevel ports pi's clampThinkingLevel: a supported level is
// kept; otherwise the nearest supported level above it, else the nearest
// below, else off.
func clampThinkingLevel(m Model, l ThinkingLevel) ThinkingLevel {
	supported := supportedThinkingLevels(m)
	if slices.Contains(supported, l) {
		return l
	}
	requested := slices.Index(modelThinkingLevels, l)
	for _, c := range modelThinkingLevels[requested+1:] {
		if slices.Contains(supported, c) {
			return c
		}
	}
	for i := requested - 1; i >= 0; i-- {
		if c := modelThinkingLevels[i]; slices.Contains(supported, c) {
			return c
		}
	}
	if len(supported) > 0 {
		return supported[0]
	}
	return thinkingOff
}

// ThinkingBudgets are custom token budgets per level for token-budget
// protocols (pi-ai's thinkingBudgets). An unset level keeps the default
// (minimal 1024, low 2048, medium 8192, high 16384); xhigh and max use high's.
// Null and negative budgets are refused. Protocols that take an effort level
// instead, such as Responses, ignore them.
type ThinkingBudgets struct {
	Minimal Nullable[int] `json:"minimal,omitzero"`
	Low     Nullable[int] `json:"low,omitzero"`
	Medium  Nullable[int] `json:"medium,omitzero"`
	High    Nullable[int] `json:"high,omitzero"`
}

func (b ThinkingBudgets) validate() string {
	for _, v := range []Nullable[int]{b.Minimal, b.Low, b.Medium, b.High} {
		if n, ok := v.Get(); v.IsNull() || ok && n < 0 {
			return "thinking budgets must be non-negative token counts"
		}
	}
	return ""
}

// minAnswerTokens is pi's MIN_ANSWER_TOKENS: tokens always left for the
// answer when a thinking budget shares the response ceiling.
const minAnswerTokens = 1024

// thinkingBudgetForLevel ports pi's thinkingBudgetForLevel: xhigh and max
// count as high, and a custom budget replaces the default for its level.
func thinkingBudgetForLevel(l ThinkingLevel, custom ThinkingBudgets) int {
	budget, defaultBudget := custom.High, 16384
	switch l {
	case ThinkingMinimal:
		budget, defaultBudget = custom.Minimal, 1024
	case ThinkingLow:
		budget, defaultBudget = custom.Low, 2048
	case ThinkingMedium:
		budget, defaultBudget = custom.Medium, 8192
	}
	if v, ok := budget.Get(); ok {
		return v
	}
	return defaultBudget
}

// adjustMaxTokensForThinking ports pi's adjustMaxTokensForThinking for
// token-budget protocols. Without a caller cap (hasBase false) the model's
// cap is the ceiling and thinking fits inside it; with one, the thinking
// budget is added on top, up to the model's cap. When the ceiling leaves no
// room beyond the budget, the budget shrinks to keep minAnswerTokens for the
// answer.
func adjustMaxTokensForThinking(base int, hasBase bool, modelMax int, l ThinkingLevel, custom ThinkingBudgets) (maxTokens, thinkingBudget int) {
	thinkingBudget = thinkingBudgetForLevel(l, custom)
	maxTokens = modelMax
	if hasBase {
		maxTokens = min(base+thinkingBudget, modelMax)
	}
	if maxTokens <= thinkingBudget {
		thinkingBudget = min(thinkingBudget, max(0, maxTokens-minAnswerTokens))
	}
	return maxTokens, thinkingBudget
}

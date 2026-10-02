package ai

import (
	"regexp"
	"strings"
)

// GeminiOptions are the full options for the Gemini Developer API (pi-ai's
// GoogleOptions without transport, credentials, headers, callbacks, retry
// settings or a custom fetch, which are not request options here). pi's
// Google path ignores the shared timeoutMs, cacheRetention, sessionId,
// metadata and samplingParams (spec I7), so they have no field.
type GeminiOptions struct {
	// Temperature is sent as generationConfig.temperature; null sends
	// nothing, as the Google SDK drops null config fields.
	Temperature Nullable[float64] `json:"temperature,omitzero"`
	// MaxTokens is sent as generationConfig.maxOutputTokens, 0 included;
	// null sends nothing.
	MaxTokens Nullable[int] `json:"maxTokens,omitzero"`
	// ToolChoice is sent as the function calling mode when tools are
	// declared; unset and null send none.
	ToolChoice Nullable[GeminiToolChoice] `json:"toolChoice,omitzero"`
	// Thinking configures a reasoning model's thinking. Unset or null sends
	// no thinkingConfig; on a model without reasoning it is ignored.
	Thinking Nullable[GeminiThinking] `json:"thinking,omitzero"`

	// unsupported is pi's refusal of a level mapping the simple entry met;
	// validate reports it.
	unsupported string
}

// GeminiToolChoice is the Gemini function calling mode.
type GeminiToolChoice string

const (
	GeminiToolChoiceAuto GeminiToolChoice = "auto"
	GeminiToolChoiceNone GeminiToolChoice = "none"
	GeminiToolChoiceAny  GeminiToolChoice = "any"
)

// GeminiThinking is pi's thinking option. Enabled true sends
// includeThoughts with Level, else BudgetTokens; Enabled false sends the
// model's disabled configuration.
type GeminiThinking struct {
	Enabled bool `json:"enabled"`
	// BudgetTokens is thinkingBudget, -1 meaning dynamic; null is sent as
	// null. It is ignored when Level is set or null.
	BudgetTokens Nullable[int] `json:"budgetTokens,omitzero"`
	// Level is thinkingLevel; null sends neither a level nor a budget, as in
	// pi.
	Level Nullable[GeminiThinkingLevel] `json:"level,omitzero"`
}

// GeminiThinkingLevel is a Gemini thinkingLevel.
type GeminiThinkingLevel string

const (
	GeminiThinkingUnspecified GeminiThinkingLevel = "THINKING_LEVEL_UNSPECIFIED"
	GeminiThinkingMinimal     GeminiThinkingLevel = "MINIMAL"
	GeminiThinkingLow         GeminiThinkingLevel = "LOW"
	GeminiThinkingMedium      GeminiThinkingLevel = "MEDIUM"
	GeminiThinkingHigh        GeminiThinkingLevel = "HIGH"
)

func (GeminiOptions) api() API { return APIGoogleGenerativeAI }

func (o GeminiOptions) clone() Options { return o }

func (o GeminiOptions) validate() string {
	if o.unsupported != "" {
		return o.unsupported
	}
	if c, ok := o.ToolChoice.Get(); ok {
		switch c {
		case GeminiToolChoiceAuto, GeminiToolChoiceNone, GeminiToolChoiceAny:
		default:
			return "tool choice must be auto, none or any"
		}
	}
	if th, ok := o.Thinking.Get(); ok {
		if b, ok := th.BudgetTokens.Get(); ok && b < -1 {
			return "thinking budgetTokens must be -1 (dynamic) or a token count"
		}
		if l, ok := th.Level.Get(); ok {
			switch l {
			case GeminiThinkingUnspecified, GeminiThinkingMinimal, GeminiThinkingLow, GeminiThinkingMedium, GeminiThinkingHigh:
			default:
				return "thinking level must be THINKING_LEVEL_UNSPECIFIED, MINIMAL, LOW, MEDIUM or HIGH"
			}
		}
	}
	return validateCommon(o.Temperature, o.MaxTokens, "", nil, 0)
}

// simpleOptions is pi's google-generative-ai streamSimple: without a
// reasoning level, or one the model clamps to off, thinking is disabled;
// otherwise the clamped level, resolved through the model's level map, is
// sent as a thinkingLevel to a model using levels and as a token budget to
// any other. The output budget is not adjusted for thinking. Sampling
// parameters, cache retention, the session id, metadata and timeoutMs have
// no Gemini field.
func (geminiAdapter) simpleOptions(m Model, o SimpleOptions, _ []Message) Options {
	full := GeminiOptions{Temperature: o.Temperature, MaxTokens: o.MaxTokens}
	if c, ok := o.ToolChoice.Get(); ok {
		full.ToolChoice = Value(GeminiToolChoice(c))
	}
	disabled := Value(GeminiThinking{Enabled: false})
	if o.Reasoning == "" {
		full.Thinking = disabled
		return full
	}
	level := clampThinkingLevel(m, o.Reasoning)
	if level == thinkingOff {
		full.Thinking = disabled
		return full
	}
	resolved, problem := resolveGeminiThinkingLevel(m, level)
	if problem != "" {
		full.unsupported = problem
		return full
	}
	if usesGeminiThinkingLevel(m) {
		full.Thinking = Value(GeminiThinking{Enabled: true, Level: Value(geminiLevel(resolved))})
		return full
	}
	full.Thinking = Value(GeminiThinking{Enabled: true, BudgetTokens: Value(geminiBudget(m, resolved, o.ThinkingBudgets))})
	return full
}

// resolveGeminiThinkingLevel ports pi's resolveGoogleThinkingLevel: the
// model's mapped value for l, lowercased, else l itself, which must be one
// of Gemini's four levels.
func resolveGeminiThinkingLevel(m Model, l ThinkingLevel) (ThinkingLevel, string) {
	mapped, isMapped := m.ThinkingLevelMap.entry(l).Get()
	resolved := l
	if isMapped {
		resolved = ThinkingLevel(strings.ToLower(mapped))
	}
	switch resolved {
	case ThinkingMinimal, ThinkingLow, ThinkingMedium, ThinkingHigh:
		return resolved, ""
	}
	shown := "undefined" // JavaScript's String(undefined)
	switch e := m.ThinkingLevelMap.entry(l); {
	case isMapped:
		shown = mapped
	case e.IsNull():
		shown = "null"
	}
	return "", "Unsupported Google thinking level mapping for " + string(m.Provider) + "/" + m.ID + ": " + string(l) + " -> " + shown
}

// geminiThinkingLevelModel matches the Gemini 3 Pro and Flash ids, with or
// without a minor version.
var geminiThinkingLevelModel = regexp.MustCompile(`gemini-3(?:\.\d+)?-(?:pro|flash)`)

// gemma4Model matches both hosted Gemma 4 naming forms.
var gemma4Model = regexp.MustCompile(`gemma-?4`)

// usesGeminiThinkingLevel ports pi's usesGoogleThinkingLevel: the models
// steered by a discrete thinkingLevel instead of a token budget.
func usesGeminiThinkingLevel(m Model) bool {
	id := strings.ToLower(m.ID)
	return geminiThinkingLevelModel.MatchString(id) || id == "gemini-flash-latest" ||
		id == "gemini-flash-lite-latest" || gemma4Model.MatchString(id)
}

// geminiLevel is a resolved level's wire value (pi's toGoogleThinkingLevel).
func geminiLevel(l ThinkingLevel) GeminiThinkingLevel {
	return GeminiThinkingLevel(strings.ToUpper(string(l)))
}

// geminiBudget ports pi's getGoogleBudget: a custom budget for the level,
// else the Gemini 2.5 family's defaults, else -1 (dynamic).
func geminiBudget(m Model, l ThinkingLevel, custom ThinkingBudgets) int {
	var budget Nullable[int]
	switch l {
	case ThinkingMinimal:
		budget = custom.Minimal
	case ThinkingLow:
		budget = custom.Low
	case ThinkingMedium:
		budget = custom.Medium
	case ThinkingHigh:
		budget = custom.High
	}
	if v, ok := budget.Get(); ok {
		return v
	}
	var defaults map[ThinkingLevel]int
	switch {
	case strings.Contains(m.ID, "2.5-pro"):
		defaults = map[ThinkingLevel]int{ThinkingMinimal: 128, ThinkingLow: 2048, ThinkingMedium: 8192, ThinkingHigh: 32768}
	case strings.Contains(m.ID, "2.5-flash-lite"):
		defaults = map[ThinkingLevel]int{ThinkingMinimal: 512, ThinkingLow: 2048, ThinkingMedium: 8192, ThinkingHigh: 24576}
	case strings.Contains(m.ID, "2.5-flash"):
		defaults = map[ThinkingLevel]int{ThinkingMinimal: 128, ThinkingLow: 2048, ThinkingMedium: 8192, ThinkingHigh: 24576}
	default:
		return -1
	}
	return defaults[l]
}

// geminiDisabledThinking ports pi's getDisabledGoogleThinkingConfig: a zero
// budget, except that a model steered by levels which cannot turn thinking
// off gets its lowest supported level instead.
func geminiDisabledThinking(m Model) geminiThinkingConfig {
	zero := 0
	if !usesGeminiThinkingLevel(m) {
		return geminiThinkingConfig{ThinkingBudget: Value(zero)}
	}
	fallback := clampThinkingLevel(m, thinkingOff)
	if fallback == thinkingOff {
		return geminiThinkingConfig{ThinkingBudget: Value(zero)}
	}
	resolved, problem := resolveGeminiThinkingLevel(m, fallback)
	if problem != "" {
		// pi throws here while building the request; such a level map
		// cannot reach the wire.
		return geminiThinkingConfig{invalid: problem}
	}
	return geminiThinkingConfig{ThinkingLevel: geminiLevel(resolved)}
}

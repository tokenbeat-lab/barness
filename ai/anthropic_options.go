package ai

import (
	"encoding/json"
	"errors"
	"time"
)

// AnthropicOptions are the full options for the Anthropic Messages protocol
// (pi-ai's AnthropicOptions without transport, credentials, headers,
// callbacks, retry settings or an injected client, which are not request
// options here). pi's sessionId only feeds session affinity headers, which
// Anthropic's own API does not take, so it has no field; samplingParams are
// ignored by this protocol (spec I7), so neither.
type AnthropicOptions struct {
	// Temperature is sent as temperature, null included, unless thinking is
	// enabled or the model does not support it or manages its effort
	// (Model.Compat).
	Temperature Nullable[float64] `json:"temperature,omitzero"`
	// MaxTokens is sent as max_tokens, 0 included; unset or null sends the
	// model's maximum.
	MaxTokens Nullable[int] `json:"maxTokens,omitzero"`
	// CacheRetention marks the system prompt, the last tool and the last
	// user turn with an ephemeral cache_control; "long" asks a model that
	// supports it for a one-hour ttl, "none" marks nothing. Empty is "short".
	CacheRetention CacheRetention `json:"cacheRetention,omitempty"`
	// Metadata's user_id is sent when it is a string; nothing else is. The
	// library never adds the tenant.
	Metadata map[string]json.RawMessage `json:"metadata,omitempty"`
	// ThinkingEnabled true enables thinking on a reasoning model (adaptive
	// or budget-based, per Model.Compat.ForceAdaptiveThinking); false sends
	// thinking disabled unless the model's level map marks off unsupported.
	// Unset or null sends no thinking field. A managed-effort model ignores
	// it: its thinking is always adaptive.
	ThinkingEnabled Nullable[bool] `json:"thinkingEnabled,omitzero"`
	// ThinkingBudgetTokens is a budget-based model's budget_tokens; 0 is
	// 1024.
	ThinkingBudgetTokens int `json:"thinkingBudgetTokens,omitempty"`
	// Effort is an adaptive model's output_config effort; empty sends none.
	// A managed-effort model (Model.Compat.SupportsMidConvoEffort) runs the
	// turn at it, empty being high, and records it on the message.
	Effort AnthropicEffort `json:"effort,omitempty"`
	// ThinkingDisplay is "summarized" (the default) or "omitted".
	ThinkingDisplay string `json:"thinkingDisplay,omitempty"`
	// InterleavedThinking requests the interleaved-thinking beta for a
	// budget-based model with thinking enabled; unset or null is true.
	InterleavedThinking Nullable[bool] `json:"interleavedThinking,omitzero"`
	// ToolChoice is sent as tool_choice; unset and null send nothing.
	ToolChoice Nullable[AnthropicToolChoice] `json:"toolChoice,omitzero"`
	// TimeoutMs bounds each attempt until its response headers arrived, as
	// the Anthropic SDK's request timeout; 0 is its default of 10 minutes.
	// The policy's ResponseHeaderTimeout applies when it is earlier. It must
	// not be negative.
	TimeoutMs int `json:"timeoutMs,omitempty"`
}

// AnthropicEffort is an adaptive thinking effort level.
type AnthropicEffort string

const (
	AnthropicEffortLow    AnthropicEffort = "low"
	AnthropicEffortMedium AnthropicEffort = "medium"
	AnthropicEffortHigh   AnthropicEffort = "high"
	AnthropicEffortXHigh  AnthropicEffort = "xhigh"
	AnthropicEffortMax    AnthropicEffort = "max"
)

// anthropicDefaultTimeout is the Anthropic SDK's DEFAULT_TIMEOUT, which pi
// leaves in place unless timeoutMs is set.
const anthropicDefaultTimeout = 10 * time.Minute

func (o AnthropicOptions) requestTimeout() time.Duration {
	if o.TimeoutMs > 0 {
		return time.Duration(o.TimeoutMs) * time.Millisecond
	}
	return anthropicDefaultTimeout
}

func (AnthropicOptions) api() API { return APIAnthropicMessages }

func (o AnthropicOptions) clone() Options {
	o.Metadata = cloneJSONValues(o.Metadata)
	return o
}

// valid reports whether e is one of the Anthropic efforts (pi's
// isAnthropicEffort).
func (e AnthropicEffort) valid() bool {
	switch e {
	case AnthropicEffortLow, AnthropicEffortMedium, AnthropicEffortHigh, AnthropicEffortXHigh, AnthropicEffortMax:
		return true
	}
	return false
}

func (o AnthropicOptions) validate() string {
	if o.Effort != "" && !o.Effort.valid() {
		return "effort must be low, medium, high, xhigh or max"
	}
	switch o.ThinkingDisplay {
	case "", "summarized", "omitted":
	default:
		return "thinking display must be summarized or omitted"
	}
	if o.ThinkingBudgetTokens < 0 {
		return "thinkingBudgetTokens must not be negative"
	}
	if c, ok := o.ToolChoice.Get(); ok {
		if problem := c.validate(); problem != "" {
			return problem
		}
	}
	if problem := validateJSONValues("metadata", o.Metadata); problem != "" {
		return problem
	}
	return validateCommon(o.Temperature, o.MaxTokens, o.CacheRetention, nil, o.TimeoutMs)
}

// simpleOptions is pi's anthropic-messages streamSimple: without a reasoning
// level thinking is disabled; an adaptive model gets the level's effort; a
// budget-based model gets the level's thinking budget on top of the output
// budget, clamped to the context window again, with room for the answer.
// pi does not clamp the level to the model here, unlike Responses. Sampling
// parameters and the session id have no Anthropic field.
func (anthropicAdapter) simpleOptions(m Model, o SimpleOptions, history []Message) Options {
	full := AnthropicOptions{Temperature: o.Temperature, MaxTokens: o.MaxTokens, CacheRetention: o.CacheRetention,
		Metadata: o.Metadata, TimeoutMs: o.TimeoutMs}
	if c, ok := o.ToolChoice.Get(); ok {
		full.ToolChoice = Value(AnthropicToolChoice{Mode: string(c)})
	}
	if o.Reasoning == "" {
		full.ThinkingEnabled = Value(false)
		return full
	}
	full.ThinkingEnabled = Value(true)
	if m.Compat.ForceAdaptiveThinking {
		full.Effort = anthropicEffort(m, o.Reasoning)
		return full
	}
	// resolve always sets the output budget.
	base, _ := o.MaxTokens.Get()
	maxTokens, budget := adjustMaxTokensForThinking(base, true, m.MaxTokens, o.Reasoning, o.ThinkingBudgets)
	maxTokens = clampMaxTokensToContext(m, history, maxTokens)
	full.MaxTokens = Value(maxTokens)
	full.ThinkingBudgetTokens = min(budget, max(0, maxTokens-minAnswerTokens))
	return full
}

// anthropicEffort ports pi's mapThinkingLevelToEffort: the model's mapped
// value, else minimal and low are low, medium is medium, and everything
// higher is high.
func anthropicEffort(m Model, l ThinkingLevel) AnthropicEffort {
	if v, ok := m.ThinkingLevelMap.entry(l).Get(); ok {
		return AnthropicEffort(v)
	}
	switch l {
	case ThinkingMinimal, ThinkingLow:
		return AnthropicEffortLow
	case ThinkingMedium:
		return AnthropicEffortMedium
	}
	return AnthropicEffortHigh
}

// AnthropicToolChoice is the Anthropic tool_choice. Exactly one form is set.
// Its JSON is pi's option value: a mode string or {"type":"tool","name":…}.
type AnthropicToolChoice struct {
	// Mode is "auto", "any" or "none".
	Mode string
	// Tool forces a call of the named tool.
	Tool string
}

func (c AnthropicToolChoice) validate() string {
	switch {
	case (c.Mode == "") == (c.Tool == ""):
		return "tool choice must set exactly one of a mode or a tool"
	case c.Mode != "" && c.Mode != "auto" && c.Mode != "any" && c.Mode != "none":
		return "tool choice mode must be auto, any or none"
	}
	return ""
}

type anthropicToolChoiceObject struct {
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// MarshalJSON encodes pi's option value.
func (c AnthropicToolChoice) MarshalJSON() ([]byte, error) {
	if c.Tool != "" {
		return json.Marshal(anthropicToolChoiceObject{Type: "tool", Name: c.Tool})
	}
	return json.Marshal(c.Mode)
}

// UnmarshalJSON decodes pi's option value.
func (c *AnthropicToolChoice) UnmarshalJSON(data []byte) error {
	var mode string
	if json.Unmarshal(data, &mode) == nil {
		*c = AnthropicToolChoice{Mode: mode}
		return nil
	}
	var obj anthropicToolChoiceObject
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	if obj.Type != "tool" {
		return errors.New("ai: unsupported tool choice type")
	}
	*c = AnthropicToolChoice{Tool: obj.Name}
	return nil
}

// wire is the tool_choice request value: {"type":mode} or the tool form.
func (c AnthropicToolChoice) wire() anthropicToolChoiceObject {
	if c.Tool != "" {
		return anthropicToolChoiceObject{Type: "tool", Name: c.Tool}
	}
	return anthropicToolChoiceObject{Type: c.Mode}
}

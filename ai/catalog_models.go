package ai

import (
	"encoding/json"
	"slices"
)

// Modality is an input or output kind a model declares.
type Modality string

const (
	ModalityText  Modality = "text"
	ModalityImage Modality = "image"
)

// Model is secret-free chat metadata located by (chat, Provider, API, ID).
type Model struct {
	Provider      ProviderID `json:"provider"`
	API           API        `json:"api"`
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Reasoning     bool       `json:"reasoning"`
	Input         []Modality `json:"input"`
	ContextWindow int        `json:"contextWindow"`
	MaxTokens     int        `json:"maxTokens"`
	// ThinkingLevelMap maps a reasoning model's levels to provider values
	// and marks unsupported ones.
	ThinkingLevelMap ThinkingLevelMap `json:"thinkingLevelMap,omitzero"`
	// DefaultThinkingLevel selects the sampling profile when reasoning is
	// omitted. It labels the service default without requesting a wire effort
	// or recording a caller choice. Empty preserves the existing off/clamp
	// selection; an explicit label must be canonical and model-supported.
	DefaultThinkingLevel ThinkingLevel `json:"defaultThinkingLevel,omitempty"`
	// SamplingParams are the model's default sampling parameters; full and
	// simple chat entries apply call-level values over them key by key.
	SamplingParams map[string]json.RawMessage `json:"samplingParams,omitempty"`
	// SamplingParamsByThinkingLevel overrides general defaults for the
	// model-supported canonical level, before call SamplingParams. Only
	// off/minimal/low/medium/high/xhigh/max keys and non-nil objects are valid.
	// OpenAI/DeepSeek Chat and Responses consume it; native protocols ignore it.
	SamplingParamsByThinkingLevel map[ThinkingLevel]map[string]json.RawMessage `json:"samplingParamsByThinkingLevel,omitempty"`
	Compat                        ModelCompat                                  `json:"compat,omitzero"`
	// Cost is the model's price, which estimates every call's Usage.Cost.
	Cost ModelCost `json:"cost"`
}

// ModelCompat holds model constraints and the protocol compatibility flags
// of pi-ai's model data that barness-ai implements. An absent pi flag has
// pi's default; additional semantic constraints document their zero behavior.
type ModelCompat struct {
	// MessagesThinkingBudgetTokens opts a model into one verified Messages
	// enabled/budget wire value, within its provider's implemented subset.
	// Zero opts into none. This is exact shape evidence, not a budget range,
	// enforcement guarantee or canonical intensity mapping (ADR-0035).
	MessagesThinkingBudgetTokens int `json:"messagesThinkingBudgetTokens,omitempty"`
	// RequiresThinking: the model cannot disable reasoning. This semantic
	// constraint is independent of the provider's wire switch and level map;
	// zero preserves existing models. Implemented by Zhipu Chat initially.
	RequiresThinking bool `json:"requiresThinking,omitempty"`
	// ThinkingSwitchOnly: canonical high enables reasoning without exposing
	// distinct effort depths. This is independent of RequiresThinking and of
	// a Host narrowing the supported level map; zero retains existing models'
	// depth semantics. Provider capability still limits the actual wire shape.
	ThinkingSwitchOnly bool `json:"thinkingSwitchOnly,omitempty"`
	// SupportsStrictMode: Responses function tools carry a "strict" field.
	SupportsStrictMode bool `json:"supportsStrictMode,omitempty"`
	// SupportsMidConvoSystemMessages: later system messages are sent in place
	// as instruction updates instead of folding into the leading one; the
	// request still declares the tools current after the last change. pi
	// pairs it with optional ways to declare tool changes in place (OpenAI's
	// additional_tools and tool search, Anthropic's native tool changes),
	// which barness-ai does not implement: it sends the current tool list,
	// as pi does with them off (ADR-0018).
	SupportsMidConvoSystemMessages bool `json:"supportsMidConvoSystemMessages,omitempty"`
	// SupportsLongCacheRetention: long cache retention is requested
	// (Responses: a 24h retention; Anthropic: a 1h cache_control ttl). Unset
	// or null means true, as in pi.
	SupportsLongCacheRetention Nullable[bool] `json:"supportsLongCacheRetention,omitzero"`
	// SupportsExplicitPromptCacheMode: cache retention is sent as
	// prompt_cache_options instead of prompt_cache_retention.
	SupportsExplicitPromptCacheMode bool `json:"supportsExplicitPromptCacheMode,omitempty"`
	// ForceAdaptiveThinking: Anthropic thinking is adaptive, steered by an
	// effort level instead of a token budget.
	ForceAdaptiveThinking bool `json:"forceAdaptiveThinking,omitempty"`
	// SupportsTemperature: Anthropic accepts a temperature. Unset or null
	// means true, as in pi.
	SupportsTemperature Nullable[bool] `json:"supportsTemperature,omitzero"`
	// SupportsMidConvoEffort: Anthropic manages the effort per turn. Thinking
	// is always adaptive with prefix-mismatched thinking dropped instead of
	// refused, no temperature is sent, every message records the turn's
	// effort (AssistantMessage.ProviderThinkingLevel) and the replayed
	// conversation names the effort of each turn (ADR-0019). It is a hard
	// constraint: without it a host changing the effort between turns gets
	// persistent 400 responses (ADR-0018).
	SupportsMidConvoEffort bool `json:"supportsMidConvoEffort,omitempty"`
}

// supportsLongCacheRetention is the model's compat flag; unset or null is
// true, as in pi.
func (m Model) supportsLongCacheRetention() bool {
	if v, ok := m.Compat.SupportsLongCacheRetention.Get(); ok {
		return v
	}
	return true
}

// acceptsImages reports whether the model takes image input; any other model
// gets pi-ai's placeholder text instead of images.
func (m Model) acceptsImages() bool { return slices.Contains(m.Input, ModalityImage) }

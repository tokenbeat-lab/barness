package ai

import (
	"cmp"
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"time"
)

// ResponsesOptions are the full options for the OpenAI Responses protocol
// (pi-ai's OpenAIResponsesOptions without transport, credentials, headers,
// callbacks or retry settings, which are not request options here).
type ResponsesOptions struct {
	// Temperature is sent as temperature, null included.
	Temperature Nullable[float64] `json:"temperature,omitzero"`
	// MaxTokens is sent as max_output_tokens, raised to the protocol's
	// minimum of 16. Unset, null and 0 send nothing.
	MaxTokens Nullable[int] `json:"maxTokens,omitzero"`
	// SamplingParams override named request fields and model defaults, on
	// both full and simple entries (pi-ai parity). Keys that change the
	// model, history, tools, stored state or cache key are refused.
	SamplingParams map[string]json.RawMessage `json:"samplingParams,omitempty"`
	// CacheRetention maps onto the prompt cache fields the model supports;
	// empty is "short".
	CacheRetention CacheRetention `json:"cacheRetention,omitempty"`
	// SessionID, unless retention is none, sends a prompt cache key and the
	// session affinity headers, all derived from it, the tenant and the
	// vendor account rather than the raw value.
	SessionID string `json:"sessionId,omitempty"`
	// ReasoningEffort is sent as reasoning.effort through the model's level
	// map, unclamped; a model without reasoning ignores it. Empty with no
	// summary sends the model's "off" mapping, unless that is null.
	ReasoningEffort ThinkingLevel `json:"reasoningEffort,omitempty"`
	// ReasoningSummary is "auto", "detailed" or "concise"; empty (or null in
	// JSON) is unset. Set alone it asks for medium effort.
	ReasoningSummary string `json:"reasoningSummary,omitempty"`
	// ServiceTier is sent as service_tier, null included.
	ServiceTier Nullable[string] `json:"serviceTier,omitzero"`
	// ToolChoice is sent as tool_choice, null included.
	ToolChoice Nullable[ResponsesToolChoice] `json:"toolChoice,omitzero"`
	// TimeoutMs bounds each attempt until its response headers arrived, as
	// openai-node's request timeout; 0 is its default of 10 minutes. The
	// policy's ResponseHeaderTimeout applies when it is earlier. It must not
	// be negative.
	TimeoutMs int `json:"timeoutMs,omitempty"`
}

// responsesDefaultTimeout is openai-node's DEFAULT_TIMEOUT, which pi leaves
// in place unless timeoutMs is set.
const responsesDefaultTimeout = 10 * time.Minute

// requestTimeout is the protocol's request timeout for each attempt.
func (o ResponsesOptions) requestTimeout() time.Duration {
	if o.TimeoutMs > 0 {
		return time.Duration(o.TimeoutMs) * time.Millisecond
	}
	return responsesDefaultTimeout
}

func (ResponsesOptions) api() API { return APIOpenAIResponses }

func (o ResponsesOptions) clone() Options {
	o.SamplingParams = cloneJSONValues(o.SamplingParams)
	if c, ok := o.ToolChoice.Get(); ok && c.AllowedTools != nil {
		allowed := *c.AllowedTools
		allowed.Functions = slices.Clone(allowed.Functions)
		c.AllowedTools = &allowed
		o.ToolChoice = Value(c)
	}
	return o
}

// responsesReservedSampling are the Responses fields samplingParams may not
// set (see checkReservedKeys). stream is the adapter's own contract (it only
// reads event streams); store and background keep the response on the vendor
// account, which may be shared by several tenants.
var responsesReservedSampling = []string{"model", "input", "stream", "tools", "previous_response_id", "conversation", "prompt", "prompt_cache_key", "store", "background"}

func (o ResponsesOptions) validate() string {
	if o.ReasoningEffort != "" && !o.ReasoningEffort.valid() {
		return "reasoning effort must be minimal, low, medium, high, xhigh or max"
	}
	switch o.ReasoningSummary {
	case "", "auto", "detailed", "concise":
	default:
		return "reasoning summary must be auto, detailed or concise"
	}
	if c, ok := o.ToolChoice.Get(); ok {
		if problem := c.validate(); problem != "" {
			return problem
		}
	}
	if problem := validateCommon(o.Temperature, o.MaxTokens, o.CacheRetention, o.SamplingParams, o.TimeoutMs); problem != "" {
		return problem
	}
	return checkReservedKeys(o.SamplingParams, responsesReservedSampling)
}

// simpleOptions is pi's openai-responses streamSimple: the resolved simple
// options pass through, the tool choice becomes a mode, and the reasoning
// level is clamped to the model; off sends no effort. Metadata and thinking
// budgets have no Responses field.
func (responsesAdapter) simpleOptions(m Model, o SimpleOptions, _ []Message) Options {
	full := ResponsesOptions{Temperature: o.Temperature, MaxTokens: o.MaxTokens, SamplingParams: o.SamplingParams,
		CacheRetention: o.CacheRetention, SessionID: o.SessionID, TimeoutMs: o.TimeoutMs}
	if o.ToolChoice.IsNull() {
		full.ToolChoice = Null[ResponsesToolChoice]()
	} else if c, ok := o.ToolChoice.Get(); ok {
		full.ToolChoice = Value(ResponsesToolChoice{Mode: string(c)})
	}
	if o.Reasoning != "" {
		if l := clampThinkingLevel(m, o.Reasoning); l != thinkingOff {
			full.ReasoningEffort = l
		}
	}
	return full
}

// cacheKey is the prompt cache key to send, "" for none: a session id
// gives one unless retention is none or the provider serves no prompt
// cache.
func (o ResponsesOptions) cacheKey(scope cacheScope, caps responsesCapabilities) string {
	if o.SessionID == "" || o.CacheRetention == CacheRetentionNone || !caps.promptCache {
		return ""
	}
	return scope.key(o.SessionID)
}

// The Responses option fields of the request body.
type responsesReasoning struct {
	Effort  string `json:"effort"`
	Summary string `json:"summary,omitempty"`
}

type responsesPromptCacheOptions struct {
	Mode string `json:"mode,omitempty"`
	TTL  string `json:"ttl,omitempty"`
}

// responsesMinOutputTokens: Responses rejects max_output_tokens below 16.
const responsesMinOutputTokens = 16

// encode sets the option fields of body for model m, as pi's buildParams
// does before samplingParams, leaving out what the provider does not serve.
func (o ResponsesOptions) encode(body *responsesBody, m Model, caps responsesCapabilities, cacheKey string) {
	retention := cmp.Or(o.CacheRetention, CacheRetentionShort)
	longRetention := m.supportsLongCacheRetention()
	explicit := m.Compat.SupportsExplicitPromptCacheMode
	body.PromptCacheKey = cacheKey
	switch {
	case !caps.promptCache:
		// No retention fields either: the provider has no prompt cache.
	case retention == CacheRetentionLong && longRetention && !explicit:
		body.PromptCacheRetention = "24h"
	case explicit && retention == CacheRetentionNone:
		body.PromptCacheOptions = &responsesPromptCacheOptions{Mode: "explicit"}
	case explicit && retention == CacheRetentionLong && longRetention:
		body.PromptCacheOptions = &responsesPromptCacheOptions{TTL: "30m"}
	}

	if n, ok := o.MaxTokens.Get(); ok && n != 0 {
		body.MaxOutputTokens = max(n, responsesMinOutputTokens)
	}
	body.Temperature = o.Temperature
	if caps.serviceTier {
		// Without the capability a requested tier was refused before (unsupportedOption).
		body.ServiceTier = o.ServiceTier
	}
	body.ToolChoice = o.ToolChoice

	if !m.Reasoning {
		return
	}
	switch {
	case o.ReasoningEffort != "" || o.ReasoningSummary != "":
		effort := "medium"
		if o.ReasoningEffort != "" {
			effort = m.ThinkingLevelMap.wireValue(o.ReasoningEffort)
		}
		body.Reasoning = &responsesReasoning{Effort: effort, Summary: cmp.Or(o.ReasoningSummary, "auto")}
		if caps.encryptedReasoning {
			body.Include = []string{"reasoning.encrypted_content"}
		}
	case !m.ThinkingLevelMap.Off.IsNull():
		// pi's `thinkingLevelMap?.off ?? "none"`: only an unset entry
		// falls back.
		effort, ok := m.ThinkingLevelMap.Off.Get()
		if !ok {
			effort = "none"
		}
		body.Reasoning = &responsesReasoning{Effort: effort}
	}
}

// applySamplingParams sets each sampling parameter on the encoded body,
// replacing a named field of the same name (pi's Object.assign after
// buildParams). Keys are applied in sorted order; the body's own key order
// is kept and new keys are appended.
func applySamplingParams(body []byte, sources ...map[string]json.RawMessage) ([]byte, error) {
	count := 0
	for _, params := range sources {
		count += len(params)
	}
	if count == 0 {
		return body, nil
	}
	v, err := parseJSON(string(body))
	if err != nil {
		return nil, err
	}
	obj := v.(*jsonObject)
	for _, params := range sources {
		for _, k := range slices.Sorted(maps.Keys(params)) {
			value, err := parseJSON(string(params[k]))
			if err != nil {
				return nil, err
			}
			obj.set(k, value)
		}
	}
	return []byte(stringifyJSON(obj)), nil
}

// ResponsesToolChoice is the Responses tool_choice. Exactly one form is set.
// Its JSON is the wire value: a mode string, {"type":"function","name":…}
// or {"type":"allowed_tools","mode":…,"tools":[{"type":"function","name":…}]}.
type ResponsesToolChoice struct {
	// Mode is "auto", "none" or "required".
	Mode string
	// Function forces a call of the named function.
	Function string
	// AllowedTools restricts the model to some of the declared functions.
	AllowedTools *ResponsesAllowedTools
}

// ResponsesAllowedTools is the allowed_tools form of a tool choice.
type ResponsesAllowedTools struct {
	// Mode is "auto" or "required".
	Mode string
	// Functions are the names of the functions the model may call.
	Functions []string
}

func (c ResponsesToolChoice) validate() string {
	forms := 0
	for _, set := range []bool{c.Mode != "", c.Function != "", c.AllowedTools != nil} {
		if set {
			forms++
		}
	}
	switch {
	case forms != 1:
		return "tool choice must set exactly one of a mode, a function or allowed tools"
	case c.Mode != "" && c.Mode != "auto" && c.Mode != "none" && c.Mode != "required":
		return "tool choice mode must be auto, none or required"
	case c.AllowedTools == nil:
		return ""
	case c.AllowedTools.Mode != "auto" && c.AllowedTools.Mode != "required":
		return "allowed tools mode must be auto or required"
	case len(c.AllowedTools.Functions) == 0 || slices.Contains(c.AllowedTools.Functions, ""):
		return "allowed tools must name at least one function, each by name"
	}
	return ""
}

type responsesNamedTool struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type responsesToolChoiceObject struct {
	Type  string               `json:"type"`
	Name  string               `json:"name,omitempty"`
	Mode  string               `json:"mode,omitempty"`
	Tools []responsesNamedTool `json:"tools,omitempty"`
}

// MarshalJSON encodes the wire value of the form that is set.
func (c ResponsesToolChoice) MarshalJSON() ([]byte, error) {
	switch {
	case c.Mode != "":
		return json.Marshal(c.Mode)
	case c.Function != "":
		return json.Marshal(responsesToolChoiceObject{Type: "function", Name: c.Function})
	case c.AllowedTools != nil:
		tools := make([]responsesNamedTool, len(c.AllowedTools.Functions))
		for i, name := range c.AllowedTools.Functions {
			tools[i] = responsesNamedTool{Type: "function", Name: name}
		}
		return json.Marshal(responsesToolChoiceObject{Type: "allowed_tools", Mode: c.AllowedTools.Mode, Tools: tools})
	}
	return []byte("{}"), nil
}

// UnmarshalJSON decodes a wire value; forms for tools barness-ai cannot
// declare (hosted, custom or MCP tools) are refused.
func (c *ResponsesToolChoice) UnmarshalJSON(data []byte) error {
	var mode string
	if json.Unmarshal(data, &mode) == nil {
		*c = ResponsesToolChoice{Mode: mode}
		return nil
	}
	var obj responsesToolChoiceObject
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	switch obj.Type {
	case "function":
		*c = ResponsesToolChoice{Function: obj.Name}
	case "allowed_tools":
		allowed := &ResponsesAllowedTools{Mode: obj.Mode, Functions: []string{}}
		for _, t := range obj.Tools {
			if t.Type != "function" {
				return errors.New("ai: allowed tools may only name functions")
			}
			allowed.Functions = append(allowed.Functions, t.Name)
		}
		*c = ResponsesToolChoice{AllowedTools: allowed}
	default:
		return errors.New("ai: unsupported tool choice type")
	}
	return nil
}

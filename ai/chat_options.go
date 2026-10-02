package ai

import (
	"cmp"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
)

// ChatOptions are the full options for the OpenAI Chat Completions protocol
// (pi-ai's OpenAICompletionsOptions without transport, credentials, headers,
// callbacks or retry settings, which are not request options here). pi's
// Chat path sends no metadata, and its thinkingBudgets only feed budget
// fields of compat flags no built-in Chat model sets, so neither has a
// field.
type ChatOptions struct {
	// Temperature is sent as temperature, null included.
	Temperature Nullable[float64] `json:"temperature,omitzero"`
	// MaxTokens is sent as max_completion_tokens (max_tokens to DeepSeek);
	// unset, null and 0 send nothing.
	MaxTokens Nullable[int] `json:"maxTokens,omitzero"`
	// SamplingParams are applied over the request body after every named
	// field, so a key here overrides it (pi-ai parity). Unlike the simple
	// entry, the model's own samplingParams are not merged in. Keys that
	// would change the model, history, tools, stored state or the cache key
	// are refused.
	SamplingParams map[string]json.RawMessage `json:"samplingParams,omitempty"`
	// CacheRetention decides when a prompt cache key is sent; "long" asks a
	// model that supports it for a 24h retention. Empty is "short".
	CacheRetention CacheRetention `json:"cacheRetention,omitempty"`
	// SessionID sends a prompt cache key derived from it, the tenant and
	// the vendor account rather than the raw value, when the endpoint is
	// OpenAI's own or the retention is long.
	SessionID string `json:"sessionId,omitempty"`
	// ReasoningEffort is sent as reasoning_effort through the model's level
	// map, unclamped; a model without reasoning ignores it. Empty sends the
	// model's "off" mapping when that is a value. To DeepSeek it also sends
	// thinking enabled, and empty sends thinking disabled unless the model
	// maps off to null.
	ReasoningEffort ThinkingLevel `json:"reasoningEffort,omitempty"`
	// ToolChoice is sent as tool_choice; unset and null send nothing.
	ToolChoice Nullable[ChatToolChoice] `json:"toolChoice,omitzero"`
	// TimeoutMs bounds each attempt until its response headers arrived, as
	// openai-node's request timeout; 0 is its default of 10 minutes. The
	// policy's ResponseHeaderTimeout applies when it is earlier. It must not
	// be negative.
	TimeoutMs int `json:"timeoutMs,omitempty"`
}

// requestTimeout is the protocol's request timeout for each attempt; the
// Chat path shares openai-node's default with Responses.
func (o ChatOptions) requestTimeout() time.Duration {
	if o.TimeoutMs > 0 {
		return time.Duration(o.TimeoutMs) * time.Millisecond
	}
	return responsesDefaultTimeout
}

func (ChatOptions) api() API { return APIOpenAICompletions }

func (o ChatOptions) clone() Options {
	o.SamplingParams = cloneJSONValues(o.SamplingParams)
	if c, ok := o.ToolChoice.Get(); ok && c.AllowedTools != nil {
		allowed := *c.AllowedTools
		allowed.Functions = slices.Clone(allowed.Functions)
		c.AllowedTools = &allowed
		o.ToolChoice = Value(c)
	}
	return o
}

// chatReservedSampling are the Chat fields samplingParams may not set (see
// checkReservedKeys). stream is the adapter's own contract; functions is the
// legacy form of tools; store keeps the completion on the vendor account,
// which may be shared by several tenants; web_search_options turns on the
// hosted web search, which only the binding can allow (ADR-0005).
var chatReservedSampling = []string{"model", "messages", "stream", "tools", "functions", "prompt_cache_key", "store", "web_search_options"}

func (o ChatOptions) validate() string {
	if o.ReasoningEffort != "" && !o.ReasoningEffort.valid() {
		return "reasoning effort must be minimal, low, medium, high, xhigh or max"
	}
	if c, ok := o.ToolChoice.Get(); ok {
		if problem := c.validate(); problem != "" {
			return problem
		}
	}
	if problem := validateCommon(o.Temperature, o.MaxTokens, o.CacheRetention, o.SamplingParams, o.TimeoutMs); problem != "" {
		return problem
	}
	return checkReservedKeys(o.SamplingParams, chatReservedSampling)
}

// simpleOptions is pi's openai-completions streamSimple: the resolved simple
// options pass through with the tool choice, and the reasoning level is
// clamped to the model; off sends no effort. Metadata has no Chat field.
func (chatAdapter) simpleOptions(m Model, o SimpleOptions, _ []Message) Options {
	full := ChatOptions{Temperature: o.Temperature, MaxTokens: o.MaxTokens, SamplingParams: o.SamplingParams,
		CacheRetention: o.CacheRetention, SessionID: o.SessionID, TimeoutMs: o.TimeoutMs}
	if c, ok := o.ToolChoice.Get(); ok {
		full.ToolChoice = Value(ChatToolChoice{Mode: string(c)})
	}
	if o.Reasoning != "" {
		if l := clampThinkingLevel(m, o.Reasoning); l != thinkingOff {
			full.ReasoningEffort = l
		}
	}
	return full
}

// openAIEndpointMarker is how pi recognizes OpenAI's own API in a model's
// base URL: a prompt cache key is sent there for any retention but none.
const openAIEndpointMarker = "api.openai.com"

// cacheKey is the prompt cache key to send, "" for none. pi sends one when
// the base URL is OpenAI's and retention is not none, or when retention is
// long and the model supports long retention; only with a session id.
func (o ChatOptions) cacheKey(scope cacheScope, endpoint string, m Model) string {
	retention := cmp.Or(o.CacheRetention, CacheRetentionShort)
	onOpenAI := strings.Contains(endpoint, openAIEndpointMarker) && retention != CacheRetentionNone
	if o.SessionID == "" || !onOpenAI && !(retention == CacheRetentionLong && m.supportsLongCacheRetention()) {
		return ""
	}
	return scope.key(o.SessionID)
}

// encode sets the option fields of body for model m, as pi's buildParams
// does before samplingParams, for the provider's compat: the output budget,
// the tool choice and the reasoning switch of its thinking format.
func (o ChatOptions) encode(body *chatBody, m Model, compat chatCompat, cacheKey string) {
	body.PromptCacheKey = cacheKey
	if o.CacheRetention == CacheRetentionLong && m.supportsLongCacheRetention() {
		body.PromptCacheRetention = "24h"
	}
	if n, ok := o.MaxTokens.Get(); ok && n != 0 {
		if compat.maxTokens {
			body.MaxTokens = n
		} else {
			body.MaxCompletionTokens = n
		}
	}
	body.Temperature = o.Temperature
	if c, ok := o.ToolChoice.Get(); ok {
		body.ToolChoice = &c
	}
	if !m.Reasoning {
		return
	}
	if compat.deepseekThinking {
		// No effort switches thinking off unless the model's level map
		// says off is unsupported (null); off is never an effort here.
		if o.ReasoningEffort != "" {
			body.Thinking = &chatThinking{Type: "enabled"}
			body.ReasoningEffort = m.ThinkingLevelMap.wireValue(o.ReasoningEffort)
		} else if !m.ThinkingLevelMap.Off.IsNull() {
			body.Thinking = &chatThinking{Type: "disabled"}
		}
		return
	}
	if o.ReasoningEffort != "" {
		body.ReasoningEffort = m.ThinkingLevelMap.wireValue(o.ReasoningEffort)
	} else if off, ok := m.ThinkingLevelMap.Off.Get(); ok {
		body.ReasoningEffort = off
	}
}

// ChatToolChoice is the Chat Completions tool_choice. Exactly one form is
// set. Its JSON is the wire value: a mode string,
// {"type":"function","function":{"name":…}} or
// {"type":"allowed_tools","allowed_tools":{"mode":…,"tools":[…]}}.
type ChatToolChoice struct {
	// Mode is "auto", "none" or "required".
	Mode string
	// Function forces a call of the named function.
	Function string
	// AllowedTools restricts the model to some of the declared functions;
	// its Mode is "auto" or "required". Both OpenAI protocols take the same
	// allowed-tools choice, only its wire shape differs, so Chat shares the
	// Responses type and its validation.
	AllowedTools *ResponsesAllowedTools
}

func (c ChatToolChoice) validate() string {
	return ResponsesToolChoice{Mode: c.Mode, Function: c.Function, AllowedTools: c.AllowedTools}.validate()
}

type chatFunctionName struct {
	Name string `json:"name"`
}

type chatNamedTool struct {
	Type     string           `json:"type"`
	Function chatFunctionName `json:"function"`
}

type chatAllowedTools struct {
	Mode  string          `json:"mode"`
	Tools []chatNamedTool `json:"tools"`
}

type chatToolChoiceObject struct {
	Type         string            `json:"type"`
	Function     *chatFunctionName `json:"function,omitempty"`
	AllowedTools *chatAllowedTools `json:"allowed_tools,omitempty"`
}

// MarshalJSON encodes the wire value of the form that is set.
func (c ChatToolChoice) MarshalJSON() ([]byte, error) {
	switch {
	case c.Mode != "":
		return json.Marshal(c.Mode)
	case c.Function != "":
		return json.Marshal(chatToolChoiceObject{Type: "function", Function: &chatFunctionName{Name: c.Function}})
	case c.AllowedTools != nil:
		tools := make([]chatNamedTool, len(c.AllowedTools.Functions))
		for i, name := range c.AllowedTools.Functions {
			tools[i] = chatNamedTool{Type: "function", Function: chatFunctionName{Name: name}}
		}
		return json.Marshal(chatToolChoiceObject{Type: "allowed_tools", AllowedTools: &chatAllowedTools{Mode: c.AllowedTools.Mode, Tools: tools}})
	}
	return []byte("{}"), nil
}

// UnmarshalJSON decodes a wire value; forms for tools barness-ai cannot
// declare (custom tools) are refused.
func (c *ChatToolChoice) UnmarshalJSON(data []byte) error {
	var mode string
	if json.Unmarshal(data, &mode) == nil {
		*c = ChatToolChoice{Mode: mode}
		return nil
	}
	var obj chatToolChoiceObject
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	switch {
	case obj.Type == "function" && obj.Function != nil:
		*c = ChatToolChoice{Function: obj.Function.Name}
	case obj.Type == "allowed_tools" && obj.AllowedTools != nil:
		allowed := &ResponsesAllowedTools{Mode: obj.AllowedTools.Mode, Functions: []string{}}
		for _, t := range obj.AllowedTools.Tools {
			if t.Type != "function" {
				return errors.New("ai: allowed tools may only name functions")
			}
			allowed.Functions = append(allowed.Functions, t.Function.Name)
		}
		*c = ChatToolChoice{AllowedTools: allowed}
	default:
		return errors.New("ai: unsupported tool choice type")
	}
	return nil
}

package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"github.com/buger/jsonparser"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

// Kimi's requested cache TTL expresses a preference; it is not proof of the
// actual paid write TTL. Chat usage therefore never derives actual TTL here.
type chatKimiCacheOptions struct {
	Mode string `json:"mode"`
	TTL  string `json:"ttl"`
}

func validateKimiChatOptions(m Model, options Options) *Error {
	o, _ := options.(ChatOptions)
	invalid := o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort)
	if n, ok := o.MaxTokens.Get(); ok && (n < 1 || n > min(m.MaxTokens, 131072)) {
		invalid = true
	}
	if o.MaxTokens.IsNull() {
		invalid = true
	}
	if t, ok := o.Temperature.Get(); ok && t != 1 {
		invalid = true
	}
	if o.Temperature.IsNull() {
		invalid = true
	}
	if o.CacheRetention == CacheRetentionNone || (o.CacheRetention == CacheRetentionLong && !m.supportsLongCacheRetention()) {
		invalid = true
	}
	if o.ToolChoice.IsNull() {
		invalid = true
	}
	if c, ok := o.ToolChoice.Get(); ok && (c.Function != "" || c.AllowedTools != nil) {
		invalid = true
	}
	if invalid {
		return newError(CodeInvalidRequest, PhaseCapability, "Kimi Chat option is outside the integrated model capabilities")
	}
	return nil
}

// Retain exact JSON numbers from every sampling layer. The standard codec's
// JavaScript conversion would otherwise round fractional budgets or fixed
// sampling values before the final guard can reject them.
func preserveKimiChatSampling(body []byte, m Model, o ChatOptions) ([]byte, error) {
	level := clampThinkingLevel(m, cmp.Or(o.ReasoningEffort, m.DefaultThinkingLevel, thinkingOff))
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], o.SamplingParams} {
		for _, field := range slices.Sorted(maps.Keys(source)) {
			var err error
			body, err = jsonparser.Set(body, source[field], field)
			if err != nil {
				return nil, err
			}
		}
	}
	return body, nil
}
func validateKimiChatFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if dec.Decode(&body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	for _, check := range []func() string{func() string { return validateKimiChatCapabilities(m, body) }, func() string { return validateKimiChatTools(body) }, func() string { return validateKimiChatMessages(body) }} {
		if problem := check(); problem != "" {
			return newError(CodeInvalidRequest, PhaseRequest, problem)
		}
	}
	return nil
}
func validateKimiChatCapabilities(m Model, b map[string]any) string {
	for field := range b {
		if !slices.Contains([]string{"model", "messages", "stream", "stream_options", "max_completion_tokens", "temperature", "top_p", "n", "presence_penalty", "frequency_penalty", "tools", "tool_choice", "reasoning_effort", "prompt_cache_key", "prompt_cache_options"}, field) {
			return "Kimi Chat request contains a field outside the integrated capabilities"
		}
	}
	if b["stream"] != true {
		return "Kimi Chat requires stream:true"
	}
	stream, ok := b["stream_options"].(map[string]any)
	if !ok || len(stream) != 1 || stream["include_usage"] != true {
		return "Kimi Chat requires stream_options.include_usage:true"
	}
	cache, ok := b["prompt_cache_options"].(map[string]any)
	if !ok || len(cache) != 2 || cache["mode"] != "implicit" || (cache["ttl"] != "5m" && cache["ttl"] != "1h") || (cache["ttl"] == "1h" && !m.supportsLongCacheRetention()) {
		return "Kimi Chat requires an integrated implicit cache TTL"
	}
	if raw, present := b["max_completion_tokens"]; present {
		n, ok := raw.(json.Number)
		if !ok || !outputBudgetNumber(json.RawMessage(n), min(m.MaxTokens, 131072)) {
			return "max_completion_tokens must be an integer from 1 through the model output limit"
		}
	}
	if raw, present := b["reasoning_effort"]; present {
		effort, ok := raw.(string)
		if !ok || !slices.Contains([]string{"low", "high", "max"}, effort) || !slices.Contains(supportedThinkingLevels(m), ThinkingLevel(effort)) {
			return "Kimi Chat reasoning effort is not supported by this model"
		}
	}
	for field, want := range map[string]string{"temperature": "1", "top_p": "0.95", "n": "1", "presence_penalty": "0", "frequency_penalty": "0"} {
		if raw, present := b[field]; present && !kimiChatFixedNumber(raw, want) {
			return "Kimi Chat sampling must retain the documented fixed value"
		}
	}
	return ""
}
func kimiChatFixedNumber(raw any, want string) bool {
	n, ok := raw.(json.Number)
	if !ok {
		return false
	}
	s := string(n)
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exp, err := strconv.Atoi(s[i+1:])
		if err != nil || exp > len(s)+10 || exp < -len(s)-10 {
			return false
		}
	}
	exact, ok := new(big.Rat).SetString(s)
	fixed, _ := new(big.Rat).SetString(want)
	return ok && exact.Cmp(fixed) == 0
}

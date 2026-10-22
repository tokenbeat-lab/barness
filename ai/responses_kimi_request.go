package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"github.com/buger/jsonparser"
	"maps"
	"slices"
)

func validateKimiResponsesOptions(m Model, options Options) *Error {
	o, _ := options.(ResponsesOptions)
	invalid := o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort)
	if n, ok := o.MaxTokens.Get(); ok && (n < 1 || n > min(m.MaxTokens, 131072)) {
		invalid = true
	}
	invalid = invalid || o.MaxTokens.IsNull() || !o.Temperature.IsZero() || o.ReasoningSummary != "" || !o.ServiceTier.IsZero() || o.CacheRetention == CacheRetentionNone || (o.CacheRetention == CacheRetentionLong && !m.supportsLongCacheRetention())
	if !o.ToolChoice.IsZero() {
		c, ok := o.ToolChoice.Get()
		invalid = invalid || !ok || c.Mode != "auto" || c.Function != "" || c.AllowedTools != nil
	}
	if invalid {
		return newError(CodeInvalidRequest, PhaseCapability, "Kimi Responses option is outside the integrated model capabilities")
	}
	return nil
}

// Exact sampling numbers must survive the standard JavaScript merge until
// the final protocol guard. Responses owns its field selection independently.
func preserveKimiResponsesSampling(body []byte, m Model, level ThinkingLevel, overrides map[string]json.RawMessage) ([]byte, error) {
	level = clampThinkingLevel(m, cmp.Or(level, m.DefaultThinkingLevel, thinkingOff))
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], overrides} {
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
func validateKimiResponsesFinalBody(m Model, raw []byte, authorize func(map[string]any) string) *Error {
	if !isKimi(m.Provider) {
		return nil
	}
	var body map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if d.Decode(&body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	for _, check := range []func() string{func() string { return validateKimiResponsesControls(m, body) }, func() string { return validateKimiResponsesInput(body) }} {
		if problem := check(); problem != "" {
			return newError(CodeInvalidRequest, PhaseRequest, problem)
		}
	}
	return nil
}
func validateKimiResponsesControls(m Model, b map[string]any) string {
	for field := range b {
		if !slices.Contains([]string{"model", "input", "stream", "max_output_tokens", "reasoning", "tools", "tool_choice", "prompt_cache_key", "prompt_cache_options"}, field) {
			return "Kimi Responses request contains an unintegrated field"
		}
	}
	if b["stream"] != true {
		return "Kimi Responses requires stream:true"
	}
	if v, present := b["max_output_tokens"]; present {
		n, ok := v.(json.Number)
		if !ok || !outputBudgetNumber(json.RawMessage(n), min(m.MaxTokens, 131072)) {
			return "max_output_tokens must be an integer from 1 through the model output limit"
		}
	}
	if v, present := b["reasoning"]; present {
		r, ok := v.(map[string]any)
		if !ok || len(r) != 1 {
			return "Kimi Responses reasoning must contain only supported effort"
		}
		effort, ok := r["effort"].(string)
		if !ok || !slices.Contains([]string{"low", "high", "max"}, effort) || !slices.Contains(supportedThinkingLevels(m), ThinkingLevel(effort)) {
			return "Kimi Responses effort is not supported by this model"
		}
	}
	if v, present := b["tool_choice"]; present && v != "auto" {
		return "Kimi Responses supports only omitted or auto tool choice"
	}
	cache, ok := b["prompt_cache_options"].(map[string]any)
	if !ok || len(cache) != 2 || cache["mode"] != "implicit" || (cache["ttl"] != "5m" && cache["ttl"] != "1h") || (cache["ttl"] == "1h" && !m.supportsLongCacheRetention()) {
		return "Kimi Responses requires an integrated implicit cache TTL"
	}
	if raw, present := b["tools"]; present {
		tools, ok := raw.([]any)
		if !ok {
			return "Kimi Responses tools must be an array"
		}
		names := map[string]bool{}
		for _, raw := range tools {
			t, ok := raw.(map[string]any)
			if !ok || t["type"] != "function" {
				return "Kimi Responses supports only function tools"
			}
			for field := range t {
				if !slices.Contains([]string{"type", "name", "description", "parameters", "strict"}, field) {
					return "Kimi Responses function tool contains an unintegrated field"
				}
			}
			name, _ := t["name"].(string)
			if name == "" || names[name] {
				return "Kimi Responses tools require unique names"
			}
			names[name] = true
			if _, ok := t["parameters"].(map[string]any); !ok {
				return "Kimi Responses function parameters must be an object"
			}
			if v, has := t["description"]; has {
				if _, ok := v.(string); !ok {
					return "Kimi Responses description must be text"
				}
			}
			if v, has := t["strict"]; has && v != false {
				return "Kimi Responses strict sampling is not integrated"
			}
		}
	}
	return ""
}
func validateKimiResponsesInput(b map[string]any) string {
	items, ok := b["input"].([]any)
	if !ok || len(items) == 0 {
		return "Kimi Responses input must be a nonempty item array"
	}
	ids, pending := map[string]bool{}, map[string]bool{}
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			return "Kimi Responses input item must be an object"
		}
		typ, _ := item["type"].(string)
		switch typ {
		case "", "message":
			for field := range item {
				if !slices.Contains([]string{"type", "role", "content"}, field) {
					return "Kimi Responses input message contains an output-only or unintegrated field"
				}
			}
			role, _ := item["role"].(string)
			if !slices.Contains([]string{"system", "user", "assistant"}, role) {
				return "Kimi Responses input message requires an integrated role"
			}
			switch v := item["content"].(type) {
			case string:
			case []any:
				for _, raw := range v {
					part, ok := raw.(map[string]any)
					if !ok || len(part) != 2 || (part["type"] != "input_text" && part["type"] != "output_text") {
						return "Kimi Responses accepts only text input parts"
					}
					if _, ok := part["text"].(string); !ok {
						return "Kimi Responses input part requires text"
					}
				}
			default:
				return "Kimi Responses message content must be text"
			}
		case "reasoning":
			if len(item) != 2 {
				return "Kimi Responses input reasoning requires only type and summary"
			}
			parts, ok := item["summary"].([]any)
			if !ok {
				return "Kimi Responses input reasoning requires a summary array"
			}
			for _, raw := range parts {
				part, ok := raw.(map[string]any)
				if !ok || len(part) != 2 || part["type"] != "summary_text" {
					return "Kimi Responses summary requires summary_text parts"
				}
				if _, ok := part["text"].(string); !ok {
					return "Kimi Responses summary part requires text"
				}
			}
		case "function_call":
			if len(item) != 4 {
				return "Kimi Responses input function call contains unintegrated fields"
			}
			id, _ := item["call_id"].(string)
			name, _ := item["name"].(string)
			args, ok := item["arguments"].(string)
			if id == "" || ids[id] || name == "" || !ok || !kimiResponsesArguments(args) {
				return "Kimi Responses function call requires a unique call id and JSON object arguments"
			}
			ids[id] = true
			pending[id] = true
		case "function_call_output":
			if len(item) != 3 {
				return "Kimi Responses function result contains unintegrated fields"
			}
			id, _ := item["call_id"].(string)
			if !pending[id] {
				return "Kimi Responses tool result must match a pending call"
			}
			if _, ok := item["output"].(string); !ok {
				return "Kimi Responses tool result must be text"
			}
			delete(pending, id)
		default:
			return "Kimi Responses input item type is not integrated"
		}
	}
	if len(pending) > 0 {
		return "Kimi Responses tool calls require matching results"
	}
	return ""
}
func kimiResponsesArguments(raw string) bool {
	var obj map[string]json.RawMessage
	return json.Unmarshal([]byte(raw), &obj) == nil && obj != nil
}

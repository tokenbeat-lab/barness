package ai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"maps"
	"math/big"
	"slices"
	"strconv"
	"strings"
)

func validateXAIResponsesOptions(m Model, options Options) *Error {
	o, _ := options.(ResponsesOptions)
	if problem := responsesCapabilitiesOf(m.Provider).unsupportedOption(o); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	if o.ReasoningEffort != "" && !slices.Contains(supportedThinkingLevels(m), o.ReasoningEffort) {
		return newError(CodeInvalidRequest, PhaseCapability, "reasoning effort is not supported by this model")
	}
	return nil
}

// xAI's exact sampling merge owns its wire numbers and explicit nulls. It
// intentionally does not change the established JS serialization of other routes.
func applyXAIResponsesSampling(raw []byte, m Model, level ThinkingLevel, overrides map[string]json.RawMessage) ([]byte, error) {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	level = clampThinkingLevel(m, cmp.Or(level, m.DefaultThinkingLevel, thinkingOff))
	for _, source := range []map[string]json.RawMessage{m.SamplingParams, m.SamplingParamsByThinkingLevel[level], overrides} {
		for _, key := range slices.Sorted(maps.Keys(source)) {
			body[key] = source[key]
		}
	}
	return marshalJS(body)
}

func validateXAIResponsesFinalBody(ctx context.Context, m Model, raw []byte, authorize func(map[string]any) string, callback bool, limits byteLimits) *Error {
	if m.Provider != ProviderXAI {
		return nil
	}
	var body map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	code := CodeInvalidRequest
	if callback {
		code = CodeCallbackFailed
	}
	if d.Decode(&body) != nil || body == nil {
		return newError(code, PhaseRequest, "xAI request must be an object")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if problem := validateXAIResponsesControls(m, body); problem != "" {
		return newError(code, PhaseRequest, problem)
	}
	if failure := validateXAIResponsesImages(ctx, m, body, limits, callback); failure != nil {
		return failure
	}
	return limits.checkRequestBody(raw)
}

func validateXAIResponsesControls(m Model, b map[string]any) string {
	if b["stream"] != true || b["store"] != false {
		return "xAI requires stream:true and stateless store:false"
	}
	include, ok := b["include"].([]any)
	if !ok || len(include) != 1 || include[0] != "reasoning.encrypted_content" {
		return "xAI requires encrypted reasoning include"
	}
	for k := range b {
		if !slices.Contains([]string{"model", "input", "stream", "store", "include", "max_output_tokens", "reasoning", "temperature", "top_p", "tools", "tool_choice", "parallel_tool_calls", "prompt_cache_key", "text"}, k) {
			return "xAI Responses field is not integrated: " + k
		}
	}
	n, ok := b["max_output_tokens"].(json.Number)
	if !ok || !outputBudgetNumber(json.RawMessage(n), m.MaxTokens) {
		return "max_output_tokens must be an integer from 1 through the model output limit"
	}
	for _, field := range []string{"temperature", "top_p"} {
		if v, present := b[field]; present {
			limit := "1"
			zero := true
			if field == "top_p" {
				limit = "0.95"
				zero = false
			}
			if !xaiSamplingNumber(v, limit, zero) {
				return "xAI sampling value is outside the integrated range"
			}
		}
	}
	if r, present := b["reasoning"]; present {
		obj, ok := r.(map[string]any)
		if !ok || len(obj) != 1 {
			return "xAI reasoning accepts only a supported effort"
		}
		effort, ok := obj["effort"].(string)
		valid := false
		if ok {
			for _, level := range supportedThinkingLevels(m) {
				if m.ThinkingLevelMap.wireValue(level) == effort {
					valid = true
				}
			}
		}
		if !valid {
			return "reasoning effort is not supported by this model"
		}
	}
	count := 0
	if raw, present := b["tools"]; present {
		tools, ok := raw.([]any)
		if !ok || len(tools) > 350 {
			return "xAI accepts at most 350 function tools"
		}
		count = len(tools)
		seen := map[string]bool{}
		for _, v := range tools {
			tool, ok := v.(map[string]any)
			if !ok || tool["type"] != "function" {
				return "xAI serves only function tools"
			}
			name, ok := tool["name"].(string)
			if !ok || name == "" || seen[name] {
				return "xAI function tools require unique names"
			}
			seen[name] = true
			for k := range tool {
				if !slices.Contains([]string{"type", "name", "description", "parameters"}, k) {
					return "xAI function tool field is not integrated: " + k
				}
			}
			if desc, present := tool["description"]; present {
				if _, ok := desc.(string); !ok {
					return "xAI function description must be text"
				}
			}
			params, ok := tool["parameters"].(map[string]any)
			if !ok || params["type"] != "object" {
				return "xAI function parameters require an object schema"
			}
		}
	}
	if c, present := b["tool_choice"]; present {
		choice, ok := c.(string)
		if !ok || !slices.Contains([]string{"auto", "none", "required"}, choice) || (choice == "required" && count == 0) {
			return "xAI tool choice requires auto, none or required with tools"
		}
	}
	if v, present := b["parallel_tool_calls"]; present {
		if _, ok := v.(bool); !ok {
			return "xAI parallel_tool_calls must be boolean"
		}
	}
	if v, present := b["text"]; present {
		obj, ok := v.(map[string]any)
		format, _ := obj["format"].(map[string]any)
		if !ok || len(obj) != 1 || len(format) != 1 || format["type"] != "text" {
			return "xAI supports only default text output"
		}
	}
	return validateXAIResponsesInput(b["input"])
}

func xaiSamplingNumber(value any, upper string, allowZero bool) bool {
	n, ok := value.(json.Number)
	if !ok {
		return false
	}
	s := string(n)
	if !json.Valid([]byte(s)) {
		return false
	}
	// Bound exponent expansion before exact rational comparison.
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		e, err := strconv.Atoi(s[i+1:])
		if err != nil || e > len(s)+8 || e < -len(s)-8 {
			return false
		}
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok || r.Sign() < 0 || (!allowZero && r.Sign() == 0) {
		return false
	}
	limit, _ := new(big.Rat).SetString(upper)
	return r.Cmp(limit) <= 0
}

func authorizeXAIResponsesPayload(m Model, key string, hosted []string, baseline ...[]byte) func(map[string]any) string {
	var native map[string]bool
	if len(baseline) > 0 {
		native = xaiNativeReferences(baseline[0])
	}
	return func(b map[string]any) string {
		if b["model"] != m.ID {
			return "payload may not change the authorized model"
		}
		value, present := b["prompt_cache_key"]
		if (key == "" && present) || (key != "" && value != key) {
			return "payload may not change the prompt cache key"
		}
		if v, present := b["store"]; present && v != false {
			return "payload may not store the response on the vendor"
		}
		for _, field := range responsesPayloadReferences {
			if _, present := b[field]; present {
				return "payload may not reference vendor-side state"
			}
		}
		if list, ok := b["tools"].([]any); ok {
			for _, raw := range list {
				tool, _ := raw.(map[string]any)
				typ, _ := tool["type"].(string)
				if typ != "function" && typ != "" && !slices.Contains(hosted, typ) {
					return "payload may not add unauthorized hosted tools"
				}
			}
		}
		if native != nil {
			if input, ok := b["input"].([]any); ok {
				for _, raw := range input {
					if item, ok := raw.(map[string]any); ok {
						if key := xaiNativeReferenceKey(item); key != "" && !native[key] {
							return "payload may not expand native reference authorization"
						}
					}
				}
			}
		}
		if field := vendorReference(b["input"]); field != "" {
			return "payload may not reference vendor-side state"
		}
		return ""
	}
}

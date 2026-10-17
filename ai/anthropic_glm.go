package ai

import (
	"encoding/json"
	"math"
	"net/http"
	"slices"
)

// GLM's basic Messages contract is deliberately narrower than Claude's.
// Wire evidence proves default omission, not canonical effort/budget mappings
// or arbitrary Claude extension fields. Ticket 07 owns bounded native/tool
// expansion; changing the model compat cannot opt into Provider features.
func validateGLMMessagesOptions(m Model, o AnthropicOptions) *Error {
	if m.Provider != ProviderZhipu {
		return nil
	}
	var problem string
	switch {
	case !o.ThinkingEnabled.IsZero() && !o.ThinkingEnabled.IsNull(), o.ThinkingBudgetTokens != 0, o.Effort != "", o.ThinkingDisplay != "", !o.InterleavedThinking.IsZero() && !o.InterleavedThinking.IsNull():
		problem = "explicit Messages thinking controls have no supported GLM mapping"
	case o.CacheRetention != "" && o.CacheRetention != CacheRetentionNone:
		problem = "GLM Messages cache controls are not supported"
	case !o.Temperature.IsZero(), len(o.Metadata) != 0:
		problem = "GLM Messages sampling and metadata controls are not supported"
	}
	if choice, ok := o.ToolChoice.Get(); ok && (choice.Mode != "auto" || choice.Tool != "") {
		problem = "GLM Messages supports only the verified automatic tool choice"
	}
	if problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	return nil
}

// Validate the final payload even without OnPayload. The checks happen after
// ordinary authorization so capability failures cannot mask another model or
// vendor-side state reference. Header transforms and betas extracted from a
// replacement payload are checked at the same final boundary.
func validateGLMMessagesBody(m Model, raw []byte, header http.Header, beta string, authorize func(map[string]any) string) *Error {
	if m.Provider != ProviderZhipu {
		return nil
	}
	var body map[string]any
	if json.Unmarshal(raw, &body) != nil || body == nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if beta != "" || header.Get("Anthropic-Dangerous-Direct-Browser-Access") != "" {
		return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages beta and browser access headers are not supported")
	}
	for key := range body {
		if !slices.Contains([]string{"model", "max_tokens", "messages", "stream", "system", "tools", "tool_choice"}, key) {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages request field is not supported: "+key)
		}
	}
	n, ok := body["max_tokens"].(float64)
	if !ok || math.Trunc(n) != n || n < 1 || n > float64(m.MaxTokens) {
		return newError(CodeInvalidRequest, PhaseRequest, "max_tokens must be an integer from 1 through the model output limit")
	}
	if body["stream"] != true {
		return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages requires the streaming wire contract")
	}
	if raw, present := body["tool_choice"]; present {
		choice, ok := raw.(map[string]any)
		if !ok || len(choice) != 1 || choice["type"] != "auto" {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages supports only the verified automatic tool choice")
		}
	}
	for _, field := range []string{"messages", "system", "tools"} {
		if unsupported := glmMessagesExtension(body[field]); unsupported != "" {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages extension is not supported: "+unsupported)
		}
	}
	return nil
}

// Inspect wire containers without walking a custom tool's arbitrary JSON
// schema/input values: property names there are user data, not extensions.
func glmMessagesExtension(v any) string {
	switch v := v.(type) {
	case []any:
		for _, item := range v {
			if field := glmMessagesExtension(item); field != "" {
				return field
			}
		}
	case map[string]any:
		for _, key := range []string{"cache_control", "eager_input_streaming", "display", "block_binding", "output_config"} {
			if _, present := v[key]; present {
				return key
			}
		}
		if field := glmMessagesExtension(v["content"]); field != "" {
			return field
		}
	}
	return ""
}

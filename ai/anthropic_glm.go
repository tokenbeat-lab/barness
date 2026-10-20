package ai

import (
	"encoding/json"
	"net/http"
	"slices"
)

// GLM's Messages contract is deliberately narrower than Claude's.
// Wire evidence proves default omission, not canonical effort/budget mappings
// or arbitrary Claude extension fields. Model metadata only narrows the
// provider's exact proven basic budget subset (ADR-0035).
func validateGLMMessagesOptions(m Model, o AnthropicOptions) *Error {
	if anthropicCapabilitiesOf(m.Provider).implementation != anthropicGLM {
		return nil
	}
	var problem string
	switch {
	case !glmMessagesThinkingOptions(m, o), o.Effort != "", o.ThinkingDisplay != "", !o.InterleavedThinking.IsZero() && !o.InterleavedThinking.IsNull():
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

func glmMessagesThinkingOptions(m Model, o AnthropicOptions) bool {
	if (o.ThinkingEnabled.IsZero() || o.ThinkingEnabled.IsNull()) && o.ThinkingBudgetTokens == 0 {
		return true
	}
	enabled, _ := o.ThinkingEnabled.Get()
	output, _ := o.MaxTokens.Get()
	return enabled && glmMessagesThinkingShape(m, o.ThinkingBudgetTokens, output)
}

// Provider implementation bounds the subset; metadata can only opt a model
// into that subset. Own wire evidence covers this exact pair, not all budgets
// below an arbitrary maximum or another protocol's canonical levels.
func glmMessagesThinkingShape(m Model, budget, output int) bool {
	return anthropicCapabilitiesOf(m.Provider).glmBudget1024 && m.Reasoning && m.Compat.MessagesThinkingBudgetTokens == 1024 && budget == 1024 && output == 4096
}

// Validate the final payload even without OnPayload. The checks happen after
// ordinary authorization so capability failures cannot mask another model or
// vendor-side state reference. Header transforms and betas extracted from a
// replacement payload are checked at the same final boundary.
func validateGLMMessagesBody(m Model, raw []byte, header http.Header, beta string, authorize func(map[string]any) string) *Error {
	if anthropicCapabilitiesOf(m.Provider).implementation != anthropicGLM {
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
		if !slices.Contains([]string{"model", "max_tokens", "messages", "stream", "system", "tools", "tool_choice", "thinking"}, key) {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages request field is not supported: "+key)
		}
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if !outputBudgetNumber(fields["max_tokens"], m.MaxTokens) {
		return newError(CodeInvalidRequest, PhaseRequest, "max_tokens must be an integer from 1 through the model output limit")
	}
	if thinking, present := body["thinking"]; present {
		t, ok := thinking.(map[string]any)
		if !ok || len(t) != 2 || t["type"] != "enabled" || t["budget_tokens"] != float64(1024) ||
			!glmMessagesThinkingShape(m, 1024, intNumber(body["max_tokens"])) {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages thinking requires the model's verified enabled/budget1024/max4096 shape")
		}
		var exactThinking map[string]json.RawMessage
		_ = json.Unmarshal(fields["thinking"], &exactThinking)
		if !outputBudgetNumber(exactThinking["budget_tokens"], 1024) {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages thinking budget must be the exact verified integer")
		}
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
	if tools, ok := body["tools"].([]any); ok {
		for _, tool := range tools {
			obj, _ := tool.(map[string]any) // authorization checked object shape
			for field := range obj {
				if !slices.Contains([]string{"name", "description", "input_schema"}, field) {
					return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages tool field is not supported: "+field)
				}
			}
		}
	}
	for _, field := range []string{"messages", "system", "tools"} {
		if unsupported := glmMessagesExtension(body[field], m.acceptsImages()); unsupported != "" {
			return newError(CodeInvalidRequest, PhaseRequest, "GLM Messages extension is not supported: "+unsupported)
		}
	}
	return nil
}

func intNumber(v any) int {
	n, _ := v.(float64)
	return int(n)
}

// Inspect wire containers without walking a custom tool's arbitrary JSON
// schema/input values: property names there are user data, not extensions.
func glmMessagesExtension(v any, allowImages bool) string {
	switch v := v.(type) {
	case []any:
		for _, item := range v {
			if field := glmMessagesExtension(item, allowImages); field != "" {
				return field
			}
		}
	case map[string]any:
		if !allowImages && v["type"] == "image" {
			return "image input for a text-only model"
		}
		if v["type"] == "redacted_thinking" {
			// No own Messages sample returned or replayed this opaque form.
			// A trusted Host can omit it; cross-origin preparation drops it.
			return "redacted_thinking"
		}
		if v["type"] == "thinking" {
			for key := range v {
				if !slices.Contains([]string{"type", "thinking", "signature"}, key) {
					return key
				}
			}
			if signature, present := v["signature"]; present {
				s, ok := signature.(string)
				if !ok || isBlank(s) {
					return "signature"
				}
			}
		}
		for _, key := range []string{"cache_control", "eager_input_streaming", "display", "block_binding", "output_config"} {
			if _, present := v[key]; present {
				return key
			}
		}
		if field := glmMessagesExtension(v["content"], allowImages); field != "" {
			return field
		}
	}
	return ""
}

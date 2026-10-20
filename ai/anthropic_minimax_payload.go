package ai

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
)

// This boundary checks the actual bytes after hooks. UseNumber preserves
// decimal precision: a near-integer output limit must not round into validity.
func validateMiniMaxMessagesBody(m Model, raw []byte, header http.Header, beta string, authorize func(map[string]any) string) *Error {
	if anthropicCapabilitiesOf(m.Provider).implementation != anthropicMiniMax {
		return nil
	}
	invalid := func(problem string) *Error { return newError(CodeInvalidRequest, PhaseRequest, problem) }
	var body map[string]any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if dec.Decode(&body) != nil || body == nil {
		return invalid("request could not be encoded")
	}
	if problem := authorize(body); problem != "" {
		return newError(CodeTenantDenied, PhaseRequest, problem)
	}
	if beta != "" || header.Get("Anthropic-Dangerous-Direct-Browser-Access") != "" {
		return invalid("MiniMax Messages beta and browser access headers are not supported")
	}
	for key := range body {
		if !slices.Contains([]string{"model", "messages", "system", "max_tokens", "stream", "tools", "tool_choice", "thinking", "output_config", "metadata", "temperature", "top_p", "service_tier"}, key) {
			return invalid("MiniMax Messages request field is not supported: " + key)
		}
	}
	var fields map[string]json.RawMessage
	_ = json.Unmarshal(raw, &fields)
	if !outputBudgetNumber(fields["max_tokens"], m.MaxTokens) {
		return invalid("max_tokens must be an integer from 1 through the model output limit")
	}
	if body["stream"] != true {
		return invalid("MiniMax Messages requires the streaming wire contract")
	}
	if _, present := body["temperature"]; present {
		if supported, declared := m.Compat.SupportsTemperature.Get(); declared && !supported {
			return invalid("model does not support temperature")
		}
	}
	for _, s := range []struct {
		field string
		max   int64
	}{{"temperature", 2}, {"top_p", 1}} {
		if v, present := body[s.field]; present {
			// Temperature has the established Nullable option contract; top_p
			// has no named null meaning in this protocol subset.
			if v == nil && s.field == "temperature" {
				continue
			}
			if !miniMaxSamplingNumber(v, s.max, true) {
				return invalid(s.field + " is outside the MiniMax Messages range")
			}
		}
	}
	if v, present := body["service_tier"]; present && v != "standard" {
		return invalid("MiniMax Messages supports only standard service tier")
	}
	if v, present := body["thinking"]; present {
		t, ok := v.(map[string]any)
		if !ok || len(t) != 1 || (t["type"] != "adaptive" && t["type"] != "disabled") || !m.Reasoning {
			return invalid("MiniMax Messages supports only adaptive or disabled thinking")
		}
		if t["type"] == "disabled" && m.Compat.RequiresThinking {
			return invalid("model requires thinking")
		}
	}
	if v, present := body["output_config"]; present {
		config, ok := v.(map[string]any)
		effort, _ := config["effort"].(string)
		if !ok || len(config) != 1 || !miniMaxMessagesHasEffort(m) || !miniMaxMessagesEffort(m, effort) {
			return invalid("model does not support the requested Messages effort")
		}
	}
	if v, present := body["tool_choice"]; present {
		choice, ok := v.(map[string]any)
		if !ok || len(choice) != 1 || !slices.Contains([]any{"auto", "none"}, choice["type"]) {
			return invalid("MiniMax Messages supports only auto or none tool choice")
		}
	}
	if problem := miniMaxMessagesShape(body["messages"]); problem != "" {
		return invalid(problem)
	}
	if value, present := body["tools"]; present {
		tools, ok := value.([]any)
		if !ok {
			return invalid("MiniMax Messages tools must be an array")
		}
		for _, tool := range tools {
			obj, ok := tool.(map[string]any)
			if !ok || !miniMaxMessagesString(obj["name"]) {
				return invalid("MiniMax Messages function tool requires a name")
			}
			schema, ok := obj["input_schema"].(map[string]any)
			if !ok || schema["type"] != "object" {
				return invalid("MiniMax Messages function tool requires an object input_schema")
			}
			if description, present := obj["description"]; present {
				if _, ok := description.(string); !ok {
					return invalid("MiniMax Messages function tool description must be a string")
				}
			}
			if properties, present := schema["properties"]; present {
				if _, ok := properties.(map[string]any); !ok {
					return invalid("MiniMax Messages schema properties must be an object")
				}
			}
			if required, present := schema["required"]; present {
				list, ok := required.([]any)
				if !ok {
					return invalid("MiniMax Messages schema required must be a string array")
				}
				for _, value := range list {
					if _, ok := value.(string); !ok {
						return invalid("MiniMax Messages schema required must be a string array")
					}
				}
			}
			for key := range obj {
				if !slices.Contains([]string{"name", "description", "input_schema"}, key) {
					return invalid("MiniMax Messages tool field is not supported: " + key)
				}
			}
			// Schema properties are business data and deliberately not traversed.
			if schema, ok := obj["input_schema"].(map[string]any); ok {
				if _, present := schema["strict"]; present {
					return invalid("MiniMax Messages strict tools are not supported")
				}
			}
		}
	}
	for _, key := range []string{"system", "messages"} {
		if problem := miniMaxMessagesContent(body[key]); problem != "" {
			return invalid("MiniMax Messages extension is not supported: " + problem)
		}
	}
	return nil
}

// Walk wire containers only. Tool input and JSON schema property names never
// become protocol fields merely because they resemble a vendor extension.
func miniMaxMessagesContent(v any) string {
	switch v := v.(type) {
	case []any:
		for _, item := range v {
			if problem := miniMaxMessagesContent(item); problem != "" {
				return problem
			}
		}
	case map[string]any:
		for _, key := range []string{"cache_control", "eager_input_streaming", "display", "block_binding", "output_config"} {
			if _, present := v[key]; present {
				return key
			}
		}
		if typ, present := v["type"]; present {
			switch typ {
			case "text", "tool_use", "tool_result":
			case "thinking":
				signature, ok := v["signature"].(string)
				if !ok || isBlank(signature) {
					return "unsigned thinking replay is not verified"
				}
				for key := range v {
					if !slices.Contains([]string{"type", "thinking", "signature"}, key) {
						return key
					}
				}
			default:
				return "content type"
			}
		}
		return miniMaxMessagesContent(v["content"])
	}
	return ""
}

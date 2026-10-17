package ai

import "slices"

// These checks describe the integrated China Chat subset, not everything the
// endpoint might accept. Host allowance establishes ownership, independently
// of provider capability. Inspect only protocol containers: identically named
// function-schema properties and JSON arguments are ordinary Host data.
// Source: ChatCompletionTextRequest / FunctionToolSchema, retrieved 2026-10-17,
// https://docs.bigmodel.cn/openapi/openapi.json (ADR-0033).
func validateGLMChatCapabilities(body map[string]any) string {
	if body["stream"] != true {
		return "Zhipu Chat requires stream:true"
	}
	for _, field := range []string{"functions", "function_call", "store", "prompt_cache_key", "prompt_cache_retention", "prompt_cache_options", "service_tier", "parallel_tool_calls", "context_management", "web_search_options"} {
		if _, present := body[field]; present {
			return "Zhipu Chat has not integrated " + field
		}
	}
	if raw, present := body["response_format"]; present {
		format, ok := raw.(map[string]any)
		if !ok || len(format) != 1 || (format["type"] != "text" && format["type"] != "json_object") {
			return "Zhipu Chat supports only text or json_object response_format"
		}
	}
	if raw, present := body["tools"]; present {
		tools, ok := raw.([]any)
		if !ok {
			return "Zhipu Chat tools must be a function array"
		}
		for _, raw := range tools {
			tool, ok := raw.(map[string]any)
			if !ok || tool["type"] != "function" {
				return "Zhipu Chat has integrated only function tools"
			}
			for field := range tool {
				if !slices.Contains([]string{"type", "function"}, field) {
					return "Zhipu Chat tool field is not integrated: " + field
				}
			}
			function, ok := tool["function"].(map[string]any)
			if !ok {
				return "Zhipu Chat function tool requires a function object"
			}
			for field := range function {
				if !slices.Contains([]string{"name", "description", "parameters"}, field) {
					return "Zhipu Chat function field is not integrated: " + field
				}
			}
			name, _ := function["name"].(string)
			if name == "" {
				return "Zhipu Chat function tool requires a name"
			}
			if _, present := function["strict"]; present {
				return "Zhipu Chat function strict is not integrated"
			}
			if _, ok := function["parameters"].(map[string]any); !ok {
				return "Zhipu Chat function parameters must be an object"
			}
		}
	}
	messages, ok := body["messages"].([]any)
	if !ok {
		return "Zhipu Chat messages must be an array"
	}
	for _, raw := range messages {
		message, ok := raw.(map[string]any)
		if !ok {
			return "Zhipu Chat message must be an object"
		}
		switch message["role"] {
		case "system", "user", "assistant", "tool":
		default:
			return "Zhipu Chat requires system, user, assistant or tool message roles"
		}
		switch message["content"].(type) {
		case string, nil, []any:
		default:
			return "Zhipu Chat message content must be text or text parts"
		}
		if content, ok := message["content"].([]any); ok {
			for _, raw := range content {
				part, ok := raw.(map[string]any)
				if !ok || part["type"] != "text" {
					return "Zhipu Chat models accept only text content"
				}
			}
		}
	}
	return ""
}

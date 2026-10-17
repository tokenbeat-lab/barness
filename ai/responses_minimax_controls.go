package ai

import (
	"encoding/json"
	"slices"
	"strconv"
)

// Provider request capability is independent from a custom catalog. Validate
// after all sampling and hooks; JSON Schema business keys remain untouched.
func validateMiniMaxResponsesControls(m Model, body map[string]any) string {
	if body["stream"] != true {
		return "MiniMax Responses requires stream:true"
	}
	allowed := []string{"model", "input", "instructions", "stream", "max_output_tokens", "temperature", "top_p", "tools", "tool_choice", "metadata", "reasoning", "text", "prompt_cache_key", "service_tier"}
	for field := range body {
		if !slices.Contains(allowed, field) {
			return "MiniMax Responses field is not integrated: " + field
		}
	}
	if tier, ok := body["service_tier"]; ok && tier != "standard" {
		return "MiniMax Responses supports only standard service tier"
	}
	if choice, ok := body["tool_choice"]; ok && choice != "auto" && choice != "none" {
		return "MiniMax Responses supports only default, auto or none tool choice"
	}
	for _, field := range []string{"temperature", "top_p"} {
		if raw, ok := body[field]; ok {
			n, ok := raw.(json.Number)
			if !ok {
				return "MiniMax Responses " + field + " must be a number in (0,1]"
			}
			f, e := strconv.ParseFloat(string(n), 64)
			if e != nil || f <= 0 || f > 1 {
				return "MiniMax Responses " + field + " must be a number in (0,1]"
			}
		}
	}
	if raw, ok := body["reasoning"]; ok {
		reasoning, ok := raw.(map[string]any)
		if !ok || len(reasoning) != 1 {
			return "MiniMax reasoning must contain only a supported effort"
		}
		effort, ok := reasoning["effort"].(string)
		valid := false
		if ok {
			for _, level := range supportedThinkingLevels(m) {
				if m.ThinkingLevelMap.wireValue(level) == effort {
					valid = true
				}
			}
		}
		if m.Compat.RequiresThinking && effort == "none" {
			valid = false
		}
		if !valid {
			return "reasoning effort is not supported by this model"
		}
	}
	if raw, ok := body["tools"]; ok {
		tools, ok := raw.([]any)
		if !ok {
			return "MiniMax Responses tools must be function declarations"
		}
		for _, raw := range tools {
			t, ok := raw.(map[string]any)
			if !ok || t["type"] != "function" {
				return "MiniMax Responses serves only function tools"
			}
			name, _ := t["name"].(string)
			if name == "" {
				return "MiniMax function tools must name a function"
			}
			for key := range t {
				if !slices.Contains([]string{"type", "name", "description", "parameters"}, key) {
					return "MiniMax function tool field is not integrated: " + key
				}
			}
			if params, ok := t["parameters"]; ok {
				if _, ok := params.(map[string]any); !ok {
					return "MiniMax function parameters must be a JSON Schema object"
				}
			}
		}
	}
	if raw, ok := body["text"]; ok {
		text, ok := raw.(map[string]any)
		format, _ := text["format"].(map[string]any)
		if !ok || len(text) != 1 || len(format) != 1 || format["type"] != "text" {
			return "MiniMax Responses supports only text output format"
		}
	}
	if raw, ok := body["metadata"]; ok {
		meta, ok := raw.(map[string]any)
		if !ok {
			return "MiniMax metadata must contain string values"
		}
		for _, value := range meta {
			if _, ok := value.(string); !ok {
				return "MiniMax metadata must contain string values"
			}
		}
	}
	return validateMiniMaxResponsesInput(body["input"])
}

func validateMiniMaxResponsesInput(value any) string {
	input, ok := value.([]any)
	if !ok {
		return "MiniMax Responses requires complete input history"
	}
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			return "MiniMax Responses requires documented input items"
		}
		switch item["type"] {
		case "reasoning":
			summary, ok := item["summary"].([]any)
			if !ok || len(item) != 2 {
				return "MiniMax input reasoning requires summary parts"
			}
			for _, raw := range summary {
				part, ok := raw.(map[string]any)
				_, text := part["text"].(string)
				if !ok || len(part) != 2 || part["type"] != "summary_text" || !text {
					return "MiniMax input reasoning requires summary_text parts"
				}
			}
		case "function_call":
			call, callOK := item["call_id"].(string)
			name, nameOK := item["name"].(string)
			args, argsOK := item["arguments"].(string)
			if len(item) != 4 || !callOK || call == "" || !nameOK || name == "" || !argsOK || !json.Valid([]byte(args)) {
				return "MiniMax function calls require name, call_id and arguments"
			}
		case "function_call_output":
			call, callOK := item["call_id"].(string)
			if len(item) != 3 || !callOK || call == "" {
				return "MiniMax function outputs require call_id and output"
			}
			if !miniMaxResponsesTextContent(item["output"]) {
				return "MiniMax Responses accepts text tool output only"
			}
		case nil, "message":
			for key := range item {
				if !slices.Contains([]string{"type", "role", "content"}, key) {
					return "MiniMax input messages require only type, role and content"
				}
			}
			if !slices.Contains([]any{"user", "assistant", "system", "developer"}, item["role"]) {
				return "MiniMax message role is not integrated"
			}
			switch content := item["content"].(type) {
			case string:
			case []any:
				for _, raw := range content {
					part, ok := raw.(map[string]any)
					_, text := part["text"].(string)
					if !ok || !slices.Contains([]any{"input_text", "output_text"}, part["type"]) || !text {
						return "MiniMax Responses accepts text input only"
					}
				}
			default:
				return "MiniMax Responses message content must be text"
			}
		default:
			return "MiniMax Responses input item type is not integrated"
		}
	}
	return ""
}

func miniMaxResponsesTextContent(value any) bool {
	switch content := value.(type) {
	case string:
		return true
	case []any:
		for _, raw := range content {
			part, ok := raw.(map[string]any)
			_, text := part["text"].(string)
			if !ok || len(part) != 2 || !slices.Contains([]any{"input_text", "output_text"}, part["type"]) || !text {
				return false
			}
		}
		return true
	default:
		return false
	}
}

package ai

import (
	"encoding/json"
	"slices"
)

// This is the integrated China text/function subset, not a list of everything
// the API may accept. Only protocol containers are checked: schema properties,
// function arguments and ordinary text may contain any of these field names.
func validateMiniMaxChatCapabilities(body map[string]any) string {
	if body["stream"] != true {
		return "MiniMax Chat requires stream:true"
	}
	for field := range body {
		if !slices.Contains([]string{"model", "messages", "stream", "stream_options", "max_completion_tokens", "temperature", "top_p", "tools", "tool_choice", "thinking", "reasoning_effort", "reasoning_split", "service_tier", "n"}, field) {
			return "MiniMax Chat has not integrated " + field
		}
	}
	stream, ok := body["stream_options"].(map[string]any)
	if !ok || len(stream) != 1 || stream["include_usage"] != true {
		return "MiniMax Chat requires stream_options.include_usage:true"
	}
	if tier, present := body["service_tier"]; present && tier != "standard" {
		return "MiniMax Chat supports only standard service_tier"
	}
	if choice, present := body["tool_choice"]; present && choice != "auto" {
		return "MiniMax Chat supports only the default or auto tool choice"
	}
	if n, present := body["n"]; present {
		number, ok := n.(json.Number)
		if !ok || !outputBudgetNumber(json.RawMessage(number), 1) {
			return "MiniMax Chat supports only n=1"
		}
	}
	for _, sample := range []struct {
		key      string
		max      float64
		nullable bool
	}{{"temperature", 2, true}, {"top_p", 1, false}} {
		if raw, present := body[sample.key]; present {
			if raw == nil && sample.nullable {
				continue
			}
			if !miniMaxSamplingNumber(raw, int64(sample.max), true) {
				return sample.key + " is outside the integrated numeric range"
			}
		}
	}
	if problem := validateMiniMaxChatTools(body); problem != "" {
		return problem
	}
	return validateMiniMaxChatMessages(body)
}

func validateMiniMaxChatTools(body map[string]any) string {
	raw, present := body["tools"]
	if !present {
		return ""
	}
	tools, ok := raw.([]any)
	if !ok {
		return "MiniMax tools must be a function array"
	}
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok || len(tool) != 2 || tool["type"] != "function" {
			return "MiniMax Chat has integrated only nested function tools"
		}
		function, ok := tool["function"].(map[string]any)
		if !ok {
			return "MiniMax function tool requires a function object"
		}
		for field := range function {
			if !slices.Contains([]string{"name", "description", "parameters"}, field) {
				return "MiniMax function field is not integrated: " + field
			}
		}
		name, _ := function["name"].(string)
		if name == "" {
			return "MiniMax function requires a name"
		}
		if _, ok := function["parameters"].(map[string]any); !ok {
			return "MiniMax function parameters must be an object"
		}
	}
	return ""
}

func validateMiniMaxChatMessages(body map[string]any) string {
	messages, ok := body["messages"].([]any)
	if !ok {
		return "MiniMax messages must be an array"
	}
	for _, raw := range messages {
		message, ok := raw.(map[string]any)
		if !ok {
			return "MiniMax message must be an object"
		}
		role, _ := message["role"].(string)
		if !slices.Contains([]string{"system", "user", "assistant", "tool"}, role) {
			return "MiniMax Chat requires system, user, assistant or tool roles"
		}
		if message["content"] == nil && role != "assistant" {
			return "MiniMax text messages require content"
		}
		if raw, present := message["reasoning_content"]; present {
			if _, ok := raw.(string); !ok || role != "assistant" {
				return "MiniMax reasoning_content must be assistant text"
			}
		}
		if raw, present := message["reasoning_details"]; present {
			encoded, err := json.Marshal(raw)
			if _, ok := chatReasoningDetails(string(encoded)); err != nil || !ok || role != "assistant" {
				return "MiniMax reasoning_details must preserve typed assistant entries"
			}
			if _, present := message["reasoning_content"]; present {
				return "MiniMax assistant reasoning formats cannot be combined"
			}
		}
		if raw, present := message["tool_calls"]; present {
			if role != "assistant" {
				return "MiniMax tool_calls require an assistant"
			}
			calls, ok := raw.([]any)
			if !ok {
				return "MiniMax tool_calls must be an array"
			}
			for _, raw := range calls {
				call, ok := raw.(map[string]any)
				if !ok || len(call) != 3 || call["type"] != "function" {
					return "MiniMax tool call must be a nested function"
				}
				id, _ := call["id"].(string)
				function, ok := call["function"].(map[string]any)
				if id == "" || !ok || len(function) != 2 {
					return "MiniMax tool call requires an id and function"
				}
				name, _ := function["name"].(string)
				_, args := function["arguments"].(string)
				if name == "" || !args {
					return "MiniMax tool call requires a name and string arguments"
				}
			}
		}
		if raw, present := message["tool_call_id"]; present {
			id, ok := raw.(string)
			if !ok || id == "" || role != "tool" {
				return "MiniMax tool_call_id requires a tool result"
			}
		}
		for field := range message {
			if !slices.Contains([]string{"role", "content", "tool_calls", "tool_call_id", "name", "reasoning_content", "reasoning_details"}, field) {
				return "MiniMax message field is not integrated: " + field
			}
		}
		switch content := message["content"].(type) {
		case string, nil:
		case []any:
			for _, raw := range content {
				part, ok := raw.(map[string]any)
				if !ok || part["type"] != "text" || len(part) != 2 {
					return "MiniMax Chat accepts only text content"
				}
				if _, ok := part["text"].(string); !ok {
					return "MiniMax text content requires text"
				}
			}
		default:
			return "MiniMax message content must be text or text parts"
		}
	}
	return ""
}

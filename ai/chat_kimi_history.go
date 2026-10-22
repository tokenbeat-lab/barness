package ai

import (
	"encoding/json"
	"slices"
)

func validateKimiChatTools(b map[string]any) string {
	tools := []any{}
	if raw, has := b["tools"]; has {
		var ok bool
		tools, ok = raw.([]any)
		if !ok {
			return "Kimi Chat tools must be an array"
		}
	}
	names := map[string]bool{}
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok || len(tool) != 2 || tool["type"] != "function" {
			return "Kimi Chat supports only function tools"
		}
		f, ok := tool["function"].(map[string]any)
		if !ok {
			return "Kimi Chat function tool requires a function object"
		}
		for field := range f {
			if !slices.Contains([]string{"name", "description", "parameters", "strict"}, field) {
				return "Kimi Chat function tool contains an unsupported field"
			}
		}
		name, _ := f["name"].(string)
		if name == "" || names[name] {
			return "Kimi Chat function tools require unique names"
		}
		names[name] = true
		if _, ok := f["parameters"].(map[string]any); !ok {
			return "Kimi Chat function parameters must be an object"
		}
		if v, present := f["strict"]; present && v != false {
			return "Kimi Chat strict function sampling is not integrated"
		}
		if v, present := f["description"]; present {
			if _, ok := v.(string); !ok {
				return "Kimi Chat function description must be text"
			}
		}
	}
	if c, present := b["tool_choice"]; present {
		mode, ok := c.(string)
		if !ok || !slices.Contains([]string{"auto", "none", "required"}, mode) || (mode == "required" && len(tools) == 0) {
			return "Kimi Chat tool choice requires an integrated mode and declared tools"
		}
	}
	return ""
}
func validateKimiChatMessages(b map[string]any) string {
	messages, ok := b["messages"].([]any)
	if !ok || len(messages) == 0 {
		return "Kimi Chat messages must be a nonempty array"
	}
	pending := map[string]bool{}
	ids := map[string]bool{}
	for _, raw := range messages {
		m, ok := raw.(map[string]any)
		if !ok {
			return "Kimi Chat message must be an object"
		}
		role, _ := m["role"].(string)
		if !slices.Contains([]string{"system", "user", "assistant", "tool"}, role) {
			return "Kimi Chat requires system, user, assistant or tool roles"
		}
		for field := range m {
			if !slices.Contains([]string{"role", "content", "reasoning_content", "tool_calls", "tool_call_id"}, field) {
				return "Kimi Chat message contains an unsupported field"
			}
		}
		switch content := m["content"].(type) {
		case string:
		case nil:
			if role != "assistant" {
				return "Kimi Chat text message requires content"
			}
		case []any:
			for _, raw := range content {
				part, ok := raw.(map[string]any)
				if !ok || len(part) != 2 || part["type"] != "text" {
					return "Kimi Chat supports only text content"
				}
				if _, ok := part["text"].(string); !ok {
					return "Kimi Chat content part requires text"
				}
			}
		default:
			return "Kimi Chat content must be text"
		}
		if v, present := m["reasoning_content"]; present {
			if _, ok := v.(string); !ok || role != "assistant" {
				return "Kimi Chat reasoning_content requires assistant text"
			}
		}
		if raw, present := m["tool_calls"]; present {
			calls, ok := raw.([]any)
			if !ok || role != "assistant" {
				return "Kimi Chat tool calls require an assistant array"
			}
			for _, raw := range calls {
				c, ok := raw.(map[string]any)
				if !ok || len(c) != 3 || c["type"] != "function" {
					return "Kimi Chat tool call requires a nested function"
				}
				id, _ := c["id"].(string)
				f, ok := c["function"].(map[string]any)
				if !ok || len(f) != 2 || id == "" || ids[id] {
					return "Kimi Chat tool calls require unique ids"
				}
				name, _ := f["name"].(string)
				args, ok := f["arguments"].(string)
				if name == "" || !ok || !kimiChatToolArguments(args) {
					return "Kimi Chat tool call requires a name and JSON object arguments"
				}
				ids[id] = true
				pending[id] = true
			}
		}
		if raw, present := m["tool_call_id"]; present {
			id, ok := raw.(string)
			if !ok || id == "" || role != "tool" || !pending[id] {
				return "Kimi Chat tool result must match one pending call"
			}
			delete(pending, id)
		} else if role == "tool" {
			return "Kimi Chat tool result requires a call id"
		}
	}
	if len(pending) > 0 {
		return "Kimi Chat tool calls require matching results"
	}
	return ""
}
func kimiChatToolArguments(raw string) bool {
	var obj map[string]json.RawMessage
	return json.Unmarshal([]byte(raw), &obj) == nil && obj != nil
}

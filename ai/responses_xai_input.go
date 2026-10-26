package ai

import (
	"bytes"
	"encoding/json"
	"slices"
)

func validateXAIResponsesInput(value any) string {
	input, ok := value.([]any)
	if !ok {
		return "xAI requires a complete input array"
	}
	calls, outputs := map[string]bool{}, map[string]bool{}
	for _, raw := range input {
		item, ok := raw.(map[string]any)
		if !ok {
			return "xAI input must contain objects"
		}
		switch item["type"] {
		case "reasoning":
			raw, err := marshalJS(item)
			if err != nil {
				return "xAI reasoning cannot be encoded"
			}
			if _, err := xaiResponsesReasoning(raw); err != nil {
				return "invalid xAI native reasoning item"
			}
		case "function_call":
			if !xaiFieldsOnly(item, "type", "call_id", "id", "name", "arguments") {
				return "xAI function call contains an unsupported field"
			}
			call, ok := item["call_id"].(string)
			name, nameOK := item["name"].(string)
			args, argsOK := item["arguments"].(string)
			if !ok || call == "" || !nameOK || name == "" || !argsOK || !xaiArgumentsObject(args) || calls[call] {
				return "xAI function calls require unique associations and complete object arguments"
			}
			calls[call] = true
			if id, present := item["id"]; present {
				if v, ok := id.(string); !ok || v == "" {
					return "xAI function item ID must be a nonempty string"
				}
			}
		case "function_call_output":
			if !xaiFieldsOnly(item, "type", "call_id", "output") {
				return "xAI function output contains an unsupported field"
			}
			call, ok := item["call_id"].(string)
			_, text := item["output"].(string)
			if !ok || !calls[call] || outputs[call] || !text {
				return "xAI function output must answer one preceding call with text"
			}
			outputs[call] = true
		case nil, "message":
			if !xaiFieldsOnly(item, "type", "role", "content", "id") {
				return "xAI input message contains an unsupported field"
			}
			role, ok := item["role"].(string)
			if !ok || !slices.Contains([]string{"system", "user", "assistant"}, role) {
				return "xAI message role is outside the integrated subset"
			}
			if id, present := item["id"]; present {
				if v, ok := id.(string); !ok || v == "" || role != "assistant" {
					return "xAI only assistant messages may carry native IDs"
				}
			}
			switch content := item["content"].(type) {
			case string:
			case []any:
				for _, rawPart := range content {
					part, ok := rawPart.(map[string]any)
					if !ok {
						return "xAI message content must contain parts"
					}
					switch part["type"] {
					case "input_text", "output_text":
						_, text := part["text"].(string)
						if !text || len(part) != 2 || (role == "assistant" && part["type"] != "output_text") || (role != "assistant" && part["type"] != "input_text") {
							return "xAI text part does not match its message role"
						}
					case "input_image":
						if role != "user" || !xaiFieldsOnly(part, "type", "detail", "image_url") {
							return "xAI images require user input_image parts"
						}
						if _, ok := part["image_url"].(string); !ok {
							return "xAI image URL must be inline text"
						}
						if detail, present := part["detail"]; present && detail != "auto" {
							return "xAI image detail control is not integrated"
						}
					default:
						return "xAI message part is outside the integrated subset"
					}
				}
			default:
				return "xAI message content must be text or parts"
			}
		default:
			return "xAI input item is outside the integrated subset"
		}
	}
	for call := range calls {
		if !outputs[call] {
			return "xAI input contains a function call without its output"
		}
	}
	return ""
}
func xaiFieldsOnly(item map[string]any, keys ...string) bool {
	for key := range item {
		if !slices.Contains(keys, key) {
			return false
		}
	}
	return true
}

// Match the source reference fields, not unrelated business keys in a schema
// or in function argument/output JSON strings. Native reasoning remains one
// indivisible signed/encrypted object at the callback boundary.
func xaiNativeReferenceKey(item map[string]any) string {
	typ, _ := item["type"].(string)
	var reference any
	switch typ {
	case "reasoning":
		reference = item
	case "function_call":
		if id, ok := item["id"]; ok {
			reference = []any{typ, id, item["call_id"]}
		}
	case "message":
		if id, ok := item["id"]; ok {
			reference = []any{typ, id}
		}
	}
	if reference == nil {
		return ""
	}
	raw, err := marshalJS(reference)
	if err != nil {
		return "invalid"
	}
	return string(raw)
}
func xaiNativeReferences(raw []byte) map[string]bool {
	var b map[string]any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	_ = d.Decode(&b)
	allowed := map[string]bool{}
	if items, ok := b["input"].([]any); ok {
		for _, value := range items {
			if item, ok := value.(map[string]any); ok {
				if key := xaiNativeReferenceKey(item); key != "" {
					allowed[key] = true
				}
			}
		}
	}
	return allowed
}

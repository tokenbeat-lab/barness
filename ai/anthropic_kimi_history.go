package ai

// Check only protocol containers. Arbitrary function input is opaque JSON;
// text-only tool results, thinking and IDs each have their own wire schema.
func kimiMessagesBlocks(value any, role string) bool {
	if _, ok := value.(string); ok {
		return true
	}
	blocks, ok := value.([]any)
	if !ok || len(blocks) == 0 {
		return false
	}
	for _, value := range blocks {
		block, ok := value.(map[string]any)
		if !ok {
			return false
		}
		fields := []string{"type"}
		switch block["type"] {
		case "text":
			fields = append(fields, "text")
			if _, ok := block["text"].(string); !ok {
				return false
			}
		case "thinking":
			fields = append(fields, "thinking", "signature")
			if role != "assistant" {
				return false
			}
			if _, ok := block["thinking"].(string); !ok {
				return false
			}
			if value, present := block["signature"]; present {
				if _, ok := value.(string); !ok {
					return false
				}
			}
		case "tool_use":
			fields = append(fields, "id", "name", "input")
			if role != "assistant" || !miniMaxMessagesString(block["id"]) || !miniMaxMessagesString(block["name"]) {
				return false
			}
			if _, ok := block["input"].(map[string]any); !ok {
				return false
			}
		case "tool_result":
			fields = append(fields, "tool_use_id", "is_error", "content")
			if role != "user" || !miniMaxMessagesString(block["tool_use_id"]) {
				return false
			}
			if value, present := block["is_error"]; present {
				if _, ok := value.(bool); !ok {
					return false
				}
			}
			if !kimiMessagesBlocks(block["content"], "result") {
				return false
			}
		default:
			return false
		}
		if role == "system" || role == "result" {
			if block["type"] != "text" {
				return false
			}
		}
		if !kimiMessagesFields(block, fields) {
			return false
		}
	}
	return true
}

func kimiMessagesHistory(value any) bool {
	messages, ok := value.([]any)
	if !ok || len(messages) == 0 {
		return false
	}
	pending := map[string]bool{}
	seen := map[string]bool{}
	for _, value := range messages {
		message, ok := value.(map[string]any)
		if !ok || !kimiMessagesFields(message, []string{"role", "content"}) {
			return false
		}
		role, _ := message["role"].(string)
		if role != "user" && role != "assistant" || !kimiMessagesBlocks(message["content"], role) {
			return false
		}
		blocks, _ := message["content"].([]any)
		for _, value := range blocks {
			block := value.(map[string]any)
			switch block["type"] {
			case "tool_use":
				id := block["id"].(string)
				if seen[id] {
					return false
				}
				seen[id] = true
				pending[id] = true
			case "tool_result":
				id := block["tool_use_id"].(string)
				if !pending[id] {
					return false
				}
				delete(pending, id)
			}
		}
	}
	return len(pending) == 0
}

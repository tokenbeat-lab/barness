package ai

// Only MiniMax's final wire containers are checked here. Function inputs and
// schema properties stay business JSON; other Providers retain their contracts.
func miniMaxMessagesShape(value any) string {
	messages, ok := value.([]any)
	if !ok || len(messages) == 0 {
		return "MiniMax Messages requires a nonempty messages array"
	}
	for _, value := range messages {
		message, ok := value.(map[string]any)
		if !ok {
			return "MiniMax Messages message must be an object"
		}
		if message["role"] != "user" && message["role"] != "assistant" {
			return "MiniMax Messages role must be user or assistant"
		}
		if _, ok := message["content"].(string); ok {
			continue
		}
		blocks, ok := message["content"].([]any)
		if !ok || len(blocks) == 0 {
			return "MiniMax Messages content must be text or a nonempty block array"
		}
		for _, block := range blocks {
			if problem := miniMaxMessageBlock(block, message["role"]); problem != "" {
				return problem
			}
		}
	}
	return ""
}

func miniMaxMessageBlock(value, role any) string {
	block, ok := value.(map[string]any)
	if !ok {
		return "MiniMax Messages content block must be an object"
	}
	switch block["type"] {
	case "text":
		if _, ok := block["text"].(string); !ok {
			return "MiniMax Messages text block requires text"
		}
	case "thinking":
		if _, ok := block["thinking"].(string); !ok || role != "assistant" {
			return "MiniMax Messages thinking block requires assistant text"
		}
	case "tool_use":
		_, object := block["input"].(map[string]any)
		if role != "assistant" || !miniMaxMessagesString(block["id"]) || !miniMaxMessagesString(block["name"]) || !object {
			return "MiniMax Messages tool use requires assistant id, name and input object"
		}
		// Do not traverse input: names and values inside it are business data.
	case "tool_result":
		if role != "user" || !miniMaxMessagesString(block["tool_use_id"]) {
			return "MiniMax Messages tool result requires user tool_use_id"
		}
		if value, present := block["is_error"]; present {
			if _, ok := value.(bool); !ok {
				return "MiniMax Messages tool result is_error must be boolean"
			}
		}
		if content, present := block["content"]; present {
			if _, ok := content.(string); ok {
				return ""
			}
			list, ok := content.([]any)
			if !ok {
				return "MiniMax Messages tool result content must be text or text blocks"
			}
			for _, value := range list {
				text, ok := value.(map[string]any)
				if !ok || text["type"] != "text" {
					return "MiniMax Messages tool result permits only text blocks"
				}
				if _, ok := text["text"].(string); !ok {
					return "MiniMax Messages tool result text must be a string"
				}
			}
		}
	default:
		return "MiniMax Messages content block type is not supported"
	}
	return ""
}

func miniMaxMessagesString(value any) bool {
	text, ok := value.(string)
	return ok && !isBlank(text)
}

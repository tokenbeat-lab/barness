package ai

import (
	"encoding/json"
	"errors"
)

// normalizeGLMResponsesEvent translates only the SDK's parsing copy. Native
// reasoning is retained from the original JSON, because GLM's replay contract
// uses its own content shape, not OpenAI's normalized array representation.
func (p *responsesParser) normalizeGLMResponsesEvent(raw []byte) ([]byte, error) {
	var event map[string]json.RawMessage
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}
	var eventType string
	if json.Unmarshal(event["type"], &eventType) != nil {
		return nil, errors.New("invalid GLM event type")
	}
	if err := validateGLMResponsesTextEvent(eventType, event); err != nil {
		return nil, err
	}
	if item, ok := event["item"]; ok {
		normalized, native, err := normalizeGLMResponsesItem(item)
		if err != nil {
			return nil, err
		}
		event["item"] = normalized
		if native != "" && eventType == "response.output_item.done" {
			var idx int64
			_ = json.Unmarshal(event["output_index"], &idx)
			if p.glmNative == nil {
				p.glmNative = map[responsesIndex]string{}
			}
			p.glmNative[responsesIndex{kind: "number", value: idx}] = native
		}
	}
	if response, ok := event["response"]; ok {
		var body map[string]json.RawMessage
		if json.Unmarshal(response, &body) != nil || body == nil {
			return nil, errors.New("invalid GLM response")
		}
		if eventType == "response.completed" || eventType == "response.incomplete" {
			var status string
			_ = json.Unmarshal(body["status"], &status)
			if (eventType == "response.completed" && status != "completed") || (eventType == "response.incomplete" && status != "incomplete") {
				return nil, errors.New("invalid GLM terminal status")
			}
		}
		if output, ok := body["output"]; ok {
			var items []json.RawMessage
			if json.Unmarshal(output, &items) != nil || string(output) == "null" {
				return nil, errors.New("invalid GLM response output")
			}
			for i, item := range items {
				normalized, native, err := normalizeGLMResponsesItem(item)
				if err != nil {
					return nil, err
				}
				items[i] = normalized
				if native != "" {
					if p.glmNative == nil {
						p.glmNative = map[responsesIndex]string{}
					}
					p.glmNative[responsesIndex{kind: "number", value: int64(i)}] = native
				}
			}
			body["output"], _ = json.Marshal(items)
		}
		event["response"], _ = json.Marshal(body)
	}
	return json.Marshal(event)
}

func normalizeGLMResponsesItem(raw json.RawMessage) (json.RawMessage, string, error) {
	var item map[string]json.RawMessage
	if json.Unmarshal(raw, &item) != nil || item == nil {
		return nil, "", errors.New("invalid GLM output item")
	}
	var typ string
	_ = json.Unmarshal(item["type"], &typ)
	if typ != "reasoning" && typ != "message" {
		return raw, "", nil
	}
	native := ""
	if typ == "reasoning" {
		native = string(raw)
	}
	content, present := item["content"]
	if !present {
		return raw, native, nil
	}
	var parts []json.RawMessage
	switch {
	case len(content) > 0 && content[0] == '{':
		parts = []json.RawMessage{content}
	case len(content) > 0 && content[0] == '[':
		if err := json.Unmarshal(content, &parts); err != nil {
			return nil, "", err
		}
	default:
		return nil, "", errors.New("invalid GLM content shape")
	}
	for _, part := range parts {
		var c map[string]json.RawMessage
		var partType, text string
		if json.Unmarshal(part, &c) != nil || c == nil || json.Unmarshal(c["type"], &partType) != nil {
			return nil, "", errors.New("invalid GLM content part")
		}
		textField := "text"
		if typ == "reasoning" && partType != "reasoning_text" {
			return nil, "", errors.New("invalid GLM reasoning part")
		}
		if typ == "message" {
			switch partType {
			case "output_text":
			case "refusal":
				textField = "refusal"
			default:
				return nil, "", errors.New("invalid GLM message part")
			}
		}
		if json.Unmarshal(c[textField], &text) != nil || string(c[textField]) == "null" {
			return nil, "", errors.New("invalid GLM content text")
		}
	}
	item["content"], _ = json.Marshal(parts)
	normalized, err := json.Marshal(item)
	return normalized, native, err
}

func validateGLMResponsesTextEvent(typ string, event map[string]json.RawMessage) error {
	field := ""
	indexRequired := false
	switch typ {
	case "response.output_item.added", "response.output_item.done":
		indexRequired = true
	case "response.output_text.delta", "response.refusal.delta", "response.reasoning_text.delta", "response.reasoning_summary_text.delta":
		field = "delta"
	case "response.output_text.done", "response.reasoning_text.done", "response.reasoning_summary_text.done":
		field = "text"
	case "response.refusal.done":
		field = "refusal"
	}
	if field == "" && !indexRequired {
		return nil
	}
	var index int64
	if json.Unmarshal(event["output_index"], &index) != nil || string(event["output_index"]) == "null" || index < 0 {
		return errors.New("invalid GLM output_index")
	}
	if field == "" {
		return nil
	}
	var text string
	if json.Unmarshal(event[field], &text) != nil || string(event[field]) == "null" {
		return errors.New("invalid GLM text event")
	}
	return nil
}

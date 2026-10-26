package supportmatrix

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai"
)

// The fixed formal suite serves text, inline images and native function items.
// Opaque reasoning objects deliberately retain unknown future fields verbatim.
func verifyXAIInputItem(item map[string]json.RawMessage) error {
	var kind, role string
	json.Unmarshal(item["type"], &kind)
	json.Unmarshal(item["role"], &role)
	switch kind {
	case "reasoning", "function_call", "function_call_output":
		if role != "" {
			return refuse("xAI native item acquired a role")
		}
	case "", "message":
		if role != "system" && role != "user" && role != "assistant" {
			return refuse("xAI message role invalid")
		}
		raw := item["content"]
		var text string
		if json.Unmarshal(raw, &text) == nil && string(raw) != "null" {
			if role != "system" {
				return refuse("xAI non-system message lost typed content")
			}
			return nil
		}
		var blocks []map[string]json.RawMessage
		if regionalDecode(raw, &blocks) != nil || len(blocks) == 0 {
			return refuse("xAI typed input content missing")
		}
		for _, b := range blocks {
			var part string
			json.Unmarshal(b["type"], &part)
			switch part {
			case "input_text", "output_text":
				if len(b) != 2 || json.Unmarshal(b["text"], &text) != nil || string(b["text"]) == "null" || (role == "assistant") != (part == "output_text") {
					return refuse("xAI input text role or fields differ")
				}
			case "input_image":
				if role != "user" {
					return refuse("xAI input image role differs")
				}
				if _, err := xaiImages(raw); err != nil {
					return err
				}
			default:
				return refuse("xAI unsupported input modality")
			}
		}
	default:
		return refuse("xAI unsupported native input kind")
	}
	return nil
}

func xaiNativeReplay(first, next xaiCall) bool {
	var old, input []map[string]json.RawMessage
	if regionalDecode(first.request["input"], &old) != nil || regionalDecode(next.request["input"], &input) != nil || len(input) < len(old)+len(first.output) {
		return false
	}
	for i, item := range old {
		if !xaiEqual(item, input[i]) {
			return false
		}
	}
	for i, item := range first.output {
		var kind string
		json.Unmarshal(item["type"], &kind)
		in := input[len(old)+i]
		switch kind {
		case "reasoning":
			if !xaiEqual(item, in) {
				return false
			}
		case "function_call":
			if !xaiFunctionProjection(item, in) {
				return false
			}
		case "message":
			if !xaiMessageProjection(item, in) {
				return false
			}
		default:
			return false
		}
	}
	for _, item := range input[len(old)+len(first.output):] {
		var kind, role string
		json.Unmarshal(item["type"], &kind)
		json.Unmarshal(item["role"], &role)
		if kind == "reasoning" || kind == "function_call" || role == "assistant" {
			return false
		}
	}
	return true
}

func xaiMessageProjection(output, input map[string]json.RawMessage) bool {
	if len(input) != 3 || !xaiJSONEqual(output["type"], input["type"]) || !xaiJSONEqual(output["role"], input["role"]) {
		return false
	}
	var old, next []map[string]json.RawMessage
	if regionalDecode(output["content"], &old) != nil || regionalDecode(input["content"], &next) != nil || len(old) != len(next) {
		return false
	}
	for i, item := range old {
		if len(next[i]) != 2 || !xaiJSONEqual(item["type"], next[i]["type"]) || !xaiJSONEqual(item["text"], next[i]["text"]) {
			return false
		}
	}
	return true
}

func xaiFunctionTools(raw json.RawMessage) ([]ai.Tool, error) {
	var tools []map[string]json.RawMessage
	if regionalDecode(raw, &tools) != nil || tools == nil || len(tools) > 350 {
		return nil, refuse("xAI function tool list invalid")
	}
	out := make([]ai.Tool, 0, len(tools))
	seen := map[string]bool{}
	for _, item := range tools {
		for key := range item {
			if key != "type" && key != "name" && key != "description" && key != "parameters" {
				return nil, refuse("xAI unsupported function tool field")
			}
		}
		var kind, name, description string
		var schema map[string]json.RawMessage
		if json.Unmarshal(item["type"], &kind) != nil || kind != "function" || json.Unmarshal(item["name"], &name) != nil || name == "" || seen[name] || regionalDecode(item["parameters"], &schema) != nil || string(schema["type"]) != `"object"` {
			return nil, refuse("xAI hosted or malformed function tool")
		}
		if d, ok := item["description"]; ok && (string(d) == "null" || json.Unmarshal(d, &description) != nil) {
			return nil, refuse("xAI function description is not text")
		}
		seen[name] = true
		out = append(out, ai.Tool{Name: name, Description: description, Parameters: item["parameters"]})
	}
	return out, nil
}

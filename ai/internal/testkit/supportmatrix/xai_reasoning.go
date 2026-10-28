package supportmatrix

import (
	"bytes"
	"encoding/json"
)

// Native reasoning remains indivisible JSON. Only documented structural types
// and resource references are validated; future non-resource fields survive.
func verifyXAIReasoning(item map[string]json.RawMessage) error {
	if xaiWireString(item["type"]) != "reasoning" {
		return refuse("xAI native reasoning discriminator invalid")
	}
	for _, key := range []string{"id", "status", "encrypted_content"} {
		if raw, ok := item[key]; ok {
			if key == "encrypted_content" && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				continue
			}
			if _, valid := xaiWireText(raw); !valid {
				return refuse("xAI native reasoning field type invalid")
			}
		}
	}
	for _, key := range []string{"summary", "content"} {
		raw, ok := item[key]
		if !ok {
			continue
		}
		var parts []map[string]json.RawMessage
		if regionalDecode(raw, &parts) != nil || parts == nil {
			return refuse("xAI native reasoning parts malformed")
		}
		want := "summary_text"
		if key == "content" {
			want = "reasoning_text"
		}
		for _, p := range parts {
			if xaiWireString(p["type"]) != want {
				return refuse("xAI native reasoning part kind invalid")
			}
			if _, valid := xaiWireText(p["text"]); !valid {
				return refuse("xAI native reasoning text invalid")
			}
		}
	}
	raw, _ := json.Marshal(item)
	if xaiResourceReference(raw) {
		return refuse("xAI native reasoning contains a vendor resource reference")
	}
	return nil
}
func xaiResourceReference(raw json.RawMessage) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) == nil && object != nil {
		if _, present := object["file_id"]; present {
			return true
		}
		if xaiWireString(object["type"]) == "item_reference" {
			return true
		}
		for _, value := range object {
			if xaiResourceReference(value) {
				return true
			}
		}
		return false
	}
	var list []json.RawMessage
	if json.Unmarshal(raw, &list) == nil {
		for _, value := range list {
			if xaiResourceReference(value) {
				return true
			}
		}
	}
	return false
}

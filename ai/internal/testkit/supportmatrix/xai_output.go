package supportmatrix

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyXAIOutput(c xaiCall) error {
	id, model := xaiWireString(c.terminal["id"]), xaiWireString(c.terminal["model"])
	if id == "" || c.message.ResponseID != id || c.message.ResponseModel != "" && c.message.ResponseModel != model || len(c.output) != len(c.message.Content) {
		return refuse("xAI response identity or block count differs")
	}
	seenItem, seenCall := map[string]bool{}, map[string]bool{}
	tools := 0
	for n, item := range c.output {
		kind, itemID := xaiWireString(item["type"]), xaiWireString(item["id"])
		if itemID == "" || seenItem[itemID] {
			return refuse("xAI output item identity missing or duplicated")
		}
		seenItem[itemID] = true
		switch kind {
		case "reasoning":
			encrypted := xaiWireString(item["encrypted_content"])
			if encrypted == "" {
				return refuse("xAI own encrypted reasoning missing")
			}
			block, ok := c.message.Content[n].(ai.Thinking)
			raw, _ := json.Marshal(item)
			if !ok || block.Redacted || block.Thinking != c.thinking[itemID] || !xaiJSONEqual(raw, []byte(block.Signature)) {
				return refuse("xAI visible reasoning, native signature or block order differs")
			}
		case "message":
			if xaiWireString(item["status"]) != "completed" || xaiWireString(item["role"]) != "assistant" {
				return refuse("xAI output message not complete")
			}
			var content []map[string]json.RawMessage
			if regionalDecode(item["content"], &content) != nil || content == nil {
				return refuse("xAI output content malformed")
			}
			text := ""
			for _, part := range content {
				k := xaiWireString(part["type"])
				field := "text"
				if k == "refusal" {
					field = "refusal"
				} else if k != "output_text" {
					return refuse("xAI unsupported output modality")
				}
				value, ok := xaiWireText(part[field])
				if !ok {
					return refuse("xAI output text invalid")
				}
				text += value
			}
			block, ok := c.message.Content[n].(ai.Text)
			if !ok || block.Text != text {
				return refuse("xAI text or cross-kind block order differs")
			}
		case "function_call":
			tools++
			call := xaiWireString(item["call_id"])
			if call == "" || seenCall[call] {
				return refuse("xAI duplicate or missing function correlation")
			}
			seenCall[call] = true
			block, ok := c.message.Content[n].(ai.ToolCall)
			if !ok {
				return refuse("xAI function block order differs")
			}
			args, valid := xaiWireText(item["arguments"])
			ids, err := xaiOpaque(block.ID)
			var object map[string]json.RawMessage
			var display ai.UncheckedArguments
			if !valid || err != nil || len(ids) != 2 || ids[0] != call || ids[1] != itemID || block.Name != xaiWireString(item["name"]) || regionalDecode([]byte(args), &object) != nil || object == nil || json.Unmarshal([]byte(args), &display) != nil || !xaiJSONEqual([]byte(args), []byte(block.RawArguments)) || !xaiJSONEqual([]byte(display), []byte(block.Arguments)) {
				return refuse("xAI raw/display arguments or opaque function identity differs")
			}
		default:
			return refuse("xAI unsupported output kind")
		}
	}
	if tools > 0 && c.message.StopReason != ai.StopReasonToolUse || tools == 0 && c.message.StopReason != ai.StopReasonStop {
		return refuse("xAI legal stop reason differs")
	}
	return nil
}
func xaiOpaque(id string) ([]string, error) {
	prefix, encoded, ok := strings.Cut(id, "|")
	if !ok || prefix != "xar1" {
		return nil, refuse("xAI opaque call version missing")
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(encoded)
	var parts []string
	if err != nil || regionalDecode(raw, &parts) != nil {
		return nil, refuse("xAI opaque call pair malformed")
	}
	return parts, nil
}
func xaiJSONEqual(a, b []byte) bool {
	// Validate duplicate keys before decoding independently with UseNumber.
	// Unknown native fields must never lose integer precision through float64.
	var left, right json.RawMessage
	if regionalDecode(a, &left) != nil || regionalDecode(b, &right) != nil {
		return false
	}
	var x, y any
	da, db := json.NewDecoder(bytes.NewReader(left)), json.NewDecoder(bytes.NewReader(right))
	da.UseNumber()
	db.UseNumber()
	if da.Decode(&x) != nil || db.Decode(&y) != nil {
		return false
	}
	return reflect.DeepEqual(x, y)
}
func xaiJSONSequence(a, b []json.RawMessage) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !xaiJSONEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

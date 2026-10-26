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
	var id, model string
	json.Unmarshal(c.terminal["id"], &id)
	json.Unmarshal(c.terminal["model"], &model)
	if id == "" || c.message.ResponseID != id || c.message.ResponseModel != "" && c.message.ResponseModel != model {
		return refuse("xAI response identity not retained")
	}
	text := ""
	reasoning := []json.RawMessage{}
	calls := []map[string]json.RawMessage{}
	seenItem, seenCall := map[string]bool{}, map[string]bool{}
	for _, item := range c.output {
		var kind, itemID string
		json.Unmarshal(item["type"], &kind)
		json.Unmarshal(item["id"], &itemID)
		if itemID == "" || seenItem[itemID] {
			return refuse("xAI output item identity missing or duplicated")
		}
		seenItem[itemID] = true
		switch kind {
		case "reasoning":
			var encrypted string
			json.Unmarshal(item["encrypted_content"], &encrypted)
			if encrypted == "" {
				return refuse("xAI own encrypted reasoning missing")
			}
			raw, _ := json.Marshal(item)
			reasoning = append(reasoning, raw)
		case "message":
			var status, role string
			json.Unmarshal(item["status"], &status)
			json.Unmarshal(item["role"], &role)
			if status != "completed" || role != "assistant" {
				return refuse("xAI output message not complete")
			}
			var content []map[string]json.RawMessage
			if regionalDecode(item["content"], &content) != nil {
				return refuse("xAI output content malformed")
			}
			for _, b := range content {
				var k, v string
				json.Unmarshal(b["type"], &k)
				if k != "output_text" || json.Unmarshal(b["text"], &v) != nil {
					return refuse("xAI unsupported output modality")
				}
				text += v
			}
		case "function_call":
			var call string
			json.Unmarshal(item["call_id"], &call)
			if call == "" || seenCall[call] {
				return refuse("xAI duplicate or missing function correlation")
			}
			seenCall[call] = true
			calls = append(calls, item)
		default:
			return refuse("xAI unsupported output kind")
		}
	}
	actualText := ""
	var actualReasoning []json.RawMessage
	var actualCalls []ai.ToolCall
	for _, b := range c.message.Content {
		switch v := b.(type) {
		case ai.Text:
			actualText += v.Text
		case ai.Thinking:
			actualReasoning = append(actualReasoning, json.RawMessage(v.Signature))
		case ai.ToolCall:
			actualCalls = append(actualCalls, v)
		default:
			return refuse("xAI Result unsupported block")
		}
	}
	if text != actualText || !xaiJSONSequence(reasoning, actualReasoning) || len(calls) != len(actualCalls) {
		return refuse("xAI terminal output differs from Result")
	}
	for i, w := range calls {
		var call, item, name, args string
		json.Unmarshal(w["call_id"], &call)
		json.Unmarshal(w["id"], &item)
		json.Unmarshal(w["name"], &name)
		json.Unmarshal(w["arguments"], &args)
		parts, err := xaiOpaque(actualCalls[i].ID)
		var object map[string]json.RawMessage
		if err != nil || len(parts) != 2 || parts[0] != call || parts[1] != item || actualCalls[i].Name != name || regionalDecode([]byte(args), &object) != nil || object == nil || !xaiJSONEqual([]byte(args), []byte(actualCalls[i].RawArguments)) {
			return refuse("xAI opaque tool identity or executable arguments differ")
		}
	}
	if len(calls) > 0 && c.message.StopReason != ai.StopReasonToolUse || len(calls) == 0 && c.message.StopReason != ai.StopReasonStop {
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

package supportmatrix

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

func verifyXAIEvents(root *os.Root, name string, c xaiCall) error {
	var events []map[string]json.RawMessage
	if xaiReadJSON(root, name, &events) != nil || len(events) < 2 {
		return refuse("xAI stream events missing")
	}
	kind := func(e map[string]json.RawMessage) string { var s string; json.Unmarshal(e["type"], &s); return s }
	if kind(events[0]) != "start" || kind(events[len(events)-1]) != "done" {
		return refuse("xAI stream lifecycle lacks typed terminals")
	}
	opened := map[int]string{}
	closed := map[int]bool{}
	terminals := 0
	for i, e := range events {
		k := kind(e)
		if k == "done" || k == "error" {
			terminals++
			if i != len(events)-1 || k != "done" {
				return refuse("xAI stream has conflicting or early terminal")
			}
			var message regionalMessage
			var reason ai.StopReason
			if regionalDecode(e["Reason"], &reason) != nil || reason != c.message.StopReason || regionalDecode(e["Message"], &message) != nil || !xaiEqual(message, c.result.Message) {
				return refuse("xAI stream terminal differs from Result")
			}
			continue
		}
		if i == 0 {
			continue
		}
		rawIndex, err := xaiCount(e["ContentIndex"])
		if err != nil || rawIndex >= int64(len(c.message.Content)) {
			return refuse("xAI stream block index invalid")
		}
		index := int(rawIndex)
		parts := strings.Split(k, "_")
		if len(parts) != 2 || parts[0] != xaiBlockKind(c.message.Content[index]) {
			return refuse("xAI stream event type differs from Result block")
		}
		if parts[1] == "end" && !xaiEventEnd(e, c.message.Content[index]) {
			return refuse("xAI stream closed content differs from Result")
		}
		if parts[1] == "delta" {
			var delta string
			if json.Unmarshal(e["Delta"], &delta) != nil || string(e["Delta"]) == "null" {
				return refuse("xAI stream delta missing")
			}
		}
		switch parts[1] {
		case "start":
			if opened[index] != "" || closed[index] {
				return refuse("xAI stream duplicate block start")
			}
			opened[index] = parts[0]
		case "delta", "end":
			if opened[index] != parts[0] || closed[index] {
				return refuse("xAI stream block lifecycle differs")
			}
			if parts[1] == "end" {
				closed[index] = true
			}
		default:
			return refuse("xAI stream event kind unknown")
		}
	}
	if terminals != 1 || len(opened) != len(closed) {
		return refuse("xAI stream terminal or block completion differs")
	}
	return nil
}

func xaiBlockKind(b ai.AssistantContent) string {
	switch b.(type) {
	case ai.Text:
		return "text"
	case ai.Thinking:
		return "thinking"
	case ai.ToolCall:
		return "toolcall"
	}
	return ""
}
func xaiEventEnd(e map[string]json.RawMessage, b ai.AssistantContent) bool {
	switch v := b.(type) {
	case ai.Text:
		var text string
		return json.Unmarshal(e["Content"], &text) == nil && text == v.Text
	case ai.Thinking:
		var text string
		return json.Unmarshal(e["Content"], &text) == nil && text == v.Thinking
	case ai.ToolCall:
		raw, _ := json.Marshal(v)
		return xaiJSONEqual(e["ToolCall"], raw)
	}
	return false
}

package supportmatrix

import (
	"encoding/json"

	"github.com/tokenbeat-lab/barness/ai"
)

// Stored wire evidence is decoded at this boundary. The domain's closed
// content interfaces deliberately do not accept arbitrary persisted objects.
type regionalResult struct {
	Message  regionalMessage `json:"message"`
	Metadata ai.CallMetadata `json:"metadata"`
}
type regionalMessage struct {
	ai.AssistantMessage
	Content []json.RawMessage `json:"content"`
}

func (m regionalMessage) assistant() (ai.AssistantMessage, error) {
	out := m.AssistantMessage
	for _, raw := range m.Content {
		var keys map[string]json.RawMessage
		if json.Unmarshal(raw, &keys) != nil {
			return ai.AssistantMessage{}, refuse("regional captured content invalid")
		}
		n := 0
		for _, k := range []string{"text", "thinking", "id"} {
			if _, ok := keys[k]; ok {
				n++
			}
		}
		if n != 1 {
			return ai.AssistantMessage{}, refuse("regional captured content ambiguous")
		}
		switch {
		case keys["text"] != nil:
			var b ai.Text
			if json.Unmarshal(raw, &b) != nil {
				return ai.AssistantMessage{}, refuse("regional captured text invalid")
			}
			out.Content = append(out.Content, b)
		case keys["thinking"] != nil:
			var b ai.Thinking
			if json.Unmarshal(raw, &b) != nil {
				return ai.AssistantMessage{}, refuse("regional captured reasoning invalid")
			}
			out.Content = append(out.Content, b)
		case keys["id"] != nil:
			var b ai.ToolCall
			if json.Unmarshal(raw, &b) != nil {
				return ai.AssistantMessage{}, refuse("regional captured tool invalid")
			}
			out.Content = append(out.Content, b)
		}
	}
	return out, nil
}

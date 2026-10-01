package e2e

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// piContext is the pi Context for the same logical input barness receives.
// Messages carry pi's required timestamp; it never reaches the wire. A nil
// content slice is sent as null, so pi's own null normalization is what the
// differential compares with. An assistant message is replayed in pi's
// message shape.
func piContext(t *testing.T, req ai.Request) json.RawMessage {
	t.Helper()
	messages := []any{}
	for _, m := range req.Messages {
		switch m := m.(type) {
		case ai.UserMessage:
			messages = append(messages, map[string]any{"role": "user", "content": piInputBlocks(t, anySlice(m.Content)), "timestamp": 0})
		case ai.AssistantMessage:
			pm := piMessage(t, m)
			if m.Content == nil {
				pm["content"] = nil
			}
			messages = append(messages, pm)
		case ai.ToolResultMessage:
			messages = append(messages, map[string]any{"role": "toolResult", "toolCallId": m.ToolCallID, "toolName": m.ToolName,
				"content": piInputBlocks(t, anySlice(m.Content)), "isError": m.IsError, "timestamp": 0})
		case ai.SystemMessage:
			messages = append(messages, piSystemMessage(t, m))
		default:
			t.Fatalf("piContext: unsupported message %T", m)
		}
	}
	ctx := map[string]any{"messages": messages}
	if req.SystemPrompt != "" {
		ctx["systemPrompt"] = req.SystemPrompt
	}
	if len(req.Tools) > 0 {
		ctx["tools"] = req.Tools
	}
	return mustMarshal(t, ctx)
}

// piInputBlocks are user or tool result blocks in pi's shape; nil stays null.
func piInputBlocks(t *testing.T, blocks []any) []map[string]any {
	if blocks == nil {
		return nil
	}
	content := []map[string]any{}
	for _, c := range blocks {
		switch c := c.(type) {
		case ai.Text:
			content = append(content, map[string]any{"type": "text", "text": c.Text})
		case ai.Image:
			content = append(content, map[string]any{"type": "image", "data": c.Data, "mimeType": c.MimeType})
		default:
			t.Fatalf("piContext: unsupported content %T", c)
		}
	}
	return content
}

// piSystemMessage is a SystemMessage in pi's shape. Sections become a JSON
// object written in the slice's order, so pi sees the keys in the order the
// host gave them (and JavaScript then applies its own key order); a removed
// section is null.
func piSystemMessage(t *testing.T, m ai.SystemMessage) map[string]any {
	out := map[string]any{"role": "system", "content": m.Content, "timestamp": 0}
	if m.Sections != nil {
		var obj bytes.Buffer
		obj.WriteByte('{')
		for i, s := range m.Sections {
			if i > 0 {
				obj.WriteByte(',')
			}
			obj.Write(mustMarshal(t, s.Name))
			obj.WriteByte(':')
			if s.Removed {
				obj.WriteString("null")
			} else {
				obj.Write(mustMarshal(t, s.Text))
			}
		}
		obj.WriteByte('}')
		out["sections"] = json.RawMessage(obj.Bytes())
	}
	if len(m.ToolsAdded) > 0 {
		out["toolsAdded"] = m.ToolsAdded
	}
	if len(m.ToolsRemoved) > 0 {
		refs := make([]map[string]string, len(m.ToolsRemoved))
		for i, name := range m.ToolsRemoved {
			refs[i] = map[string]string{"name": name}
		}
		out["toolsRemoved"] = refs
	}
	return out
}

func anySlice[T any](in []T) []any {
	if in == nil {
		return nil
	}
	out := make([]any, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

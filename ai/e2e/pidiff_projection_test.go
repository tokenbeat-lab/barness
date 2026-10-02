package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The differential compares barness-ai's public results in pi-ai's JSON
// shape; these helpers project them onto it.

// piEvent is a barness event in pi-ai's event shape, taken when the consumer
// receives it. Like the pi runner, which serializes each event as received, the
// live partial becomes a snapshot of the view at that moment. Barness-only
// extensions (ErrorEvent.Err with Code/Phase) are not part of the pi golden
// (spec §5) and are asserted by the offline E2E instead. An event kind without
// a mapping shows up as a difference rather than being dropped.
func piEvent(t *testing.T, e ai.Event) map[string]any {
	block := func(typ string, index int, partial *ai.PartialView, fields ...any) map[string]any {
		out := map[string]any{"type": typ, "contentIndex": index, "partial": piMessage(t, partial.Snapshot())}
		for i := 0; i < len(fields); i += 2 {
			out[fields[i].(string)] = fields[i+1]
		}
		return out
	}
	switch e := e.(type) {
	case ai.StartEvent:
		return map[string]any{"type": "start", "partial": piMessage(t, e.Partial.Snapshot())}
	case ai.TextStartEvent:
		return block("text_start", e.ContentIndex, e.Partial)
	case ai.TextDeltaEvent:
		return block("text_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.TextEndEvent:
		return block("text_end", e.ContentIndex, e.Partial, "content", e.Content)
	case ai.ThinkingStartEvent:
		return block("thinking_start", e.ContentIndex, e.Partial)
	case ai.ThinkingDeltaEvent:
		return block("thinking_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.ThinkingEndEvent:
		return block("thinking_end", e.ContentIndex, e.Partial, "content", e.Content)
	case ai.ToolCallStartEvent:
		return block("toolcall_start", e.ContentIndex, e.Partial)
	case ai.ToolCallDeltaEvent:
		return block("toolcall_delta", e.ContentIndex, e.Partial, "delta", e.Delta)
	case ai.ToolCallEndEvent:
		return block("toolcall_end", e.ContentIndex, e.Partial, "toolCall", piBlock(t, e.ToolCall))
	case ai.DoneEvent:
		return map[string]any{"type": "done", "reason": e.Reason, "message": piMessage(t, e.Message)}
	case ai.ErrorEvent:
		return map[string]any{"type": "error", "reason": e.Reason, "error": piMessage(t, e.Message)}
	default:
		return map[string]any{"type": string(e.Type()), "barnessUnmapped": fmt.Sprintf("%T", e)}
	}
}

// piMessage is a barness AssistantMessage in pi-ai's message shape. Every
// field barness serializes is kept, so a field pi lacks surfaces as a
// difference; only pi's role and content-block type discriminators are added.
func piMessage(t *testing.T, m ai.AssistantMessage) map[string]any {
	t.Helper()
	out := jsonObject(t, m)
	out["role"] = "assistant"
	content := make([]any, 0, len(m.Content))
	for _, c := range m.Content {
		content = append(content, piBlock(t, c))
	}
	out["content"] = content
	return out
}

// piBlock is a content block in pi's shape: barness's fields plus pi's type
// discriminator.
func piBlock(t *testing.T, c ai.AssistantContent) map[string]any {
	block := jsonObject(t, c)
	switch c.(type) {
	case ai.Text:
		block["type"] = "text"
	case ai.Thinking:
		block["type"] = "thinking"
	case ai.ToolCall:
		block["type"] = "toolCall"
	default:
		block["type"] = fmt.Sprintf("barness-unmapped:%T", c)
	}
	return block
}

// jsonObject is v's JSON serialization as a generic object.
func jsonObject(t *testing.T, v any) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(mustMarshal(t, v), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The same words denote protocol extensions in wire containers and ordinary
// data inside a custom tool schema or function arguments (spec R20).
func TestMiniMaxMessagesBusinessFields(t *testing.T) {
	ev := run.Case(t, "MM04-business-fields")
	w := newMiniMaxMessagesWorld(t, nil)
	tool := ai.ToolCall{ID: "business-call", Name: "inspect", Arguments: ai.UncheckedArguments(`{"type":"file","file_id":"business-id","thinking":{"type":"disabled"},"cache_control":"business","stop_sequences":["business"]}`)}
	req := miniMaxMessagesRequest()
	req.Tools = []ai.Tool{{Name: "inspect", Description: "Inspect business fields.", Parameters: json.RawMessage(`{"type":"object","properties":{"file_id":{"type":"string"},"type":{"type":"string"},"thinking":{"type":"object"},"cache_control":{"type":"string"},"stop_sequences":{"type":"array","items":{"type":"string"}}},"required":["file_id"]}`)}}
	req.Messages = append(req.Messages, ai.AssistantMessage{Provider: ai.ProviderMiniMax, API: ai.APIAnthropicMessages, Model: "MiniMax-M3", Content: []ai.AssistantContent{tool}, StopReason: ai.StopReasonToolUse}, ai.ToolResultText(tool, "Synthetic result.", false))
	enqueue(ev, w, miniMaxMessagesReply(t, "MiniMax-M3"))
	o := drain(w.client.Stream(ctxFor(t), textScope("mm04-business-fields"), miniMaxMessagesTarget("MiniMax-M3"), req, nil))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("business names remain legal through public Client", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		return
	}
	var body struct {
		Messages []struct {
			Content []struct {
				Type  string
				Input json.RawMessage
			}
		}
	}
	mustUnmarshal(t, w.provider.Requests()[0].Body, &body)
	ev.Check("function arguments retain exact business value", len(body.Messages) > 1 && jsonEqual(body.Messages[1].Content[0].Input, []byte(tool.Arguments)), "%s", w.provider.Requests()[0].Body)
}

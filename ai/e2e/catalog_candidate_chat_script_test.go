package e2e

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Literal, synthetic protocol messages form an independent response oracle.
// Token counts represent a small response even when exercising a price tier;
// no design-load payload, real credential or vendor request is used.
func catalogChatReply(t *testing.T, m ai.Model, input int64, tool bool) provider.Reply {
	t.Helper()
	id := string(mustMarshal(t, m.ID))
	total := input + 4
	switch m.API {
	case ai.APIOpenAIResponses:
		item := `{"id":"catalog-message","type":"message","role":"assistant","status":"completed","content":[{"type":"output_text","text":"Catalog reply.","annotations":[]}]}`
		if tool {
			item = `{"id":"catalog-function","type":"function_call","status":"completed","call_id":"catalog-call","name":"catalog_tool","arguments":"{}"}`
		}
		frames := []json.RawMessage{}
		if tool {
			frames = append(frames, json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":`+item+`}`), json.RawMessage(`{"type":"response.function_call_arguments.done","output_index":0,"item_id":"catalog-function","arguments":"{}"}`))
		} else {
			frames = append(frames, json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"catalog-message","type":"message","role":"assistant","content":[]}}`), json.RawMessage(`{"type":"response.output_text.delta","output_index":0,"item_id":"catalog-message","content_index":0,"delta":"Catalog reply."}`))
		}
		frames = append(frames, json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":`+item+`}`), json.RawMessage(fmt.Sprintf(`{"type":"response.completed","response":{"id":"catalog-response","model":%s,"status":"completed","output":[%s],"usage":{"input_tokens":%d,"output_tokens":4,"total_tokens":%d,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}}}`, id, item, input, total)))
		return sseEvents(t, frames, provider.FramingLF)
	case ai.APIOpenAICompletions:
		delta := `{"role":"assistant","content":"Catalog reply."}`
		stop := "stop"
		if tool {
			delta = `{"role":"assistant","reasoning_content":"Catalog thought.","tool_calls":[{"index":0,"id":"catalog-call","type":"function","function":{"name":"catalog_tool","arguments":"{}"}}]}`
			stop = "tool_calls"
		}
		return chatEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"id":"catalog-chat","model":%s,"choices":[{"index":0,"delta":%s,"finish_reason":null}]}`, id, delta)), json.RawMessage(fmt.Sprintf(`{"id":"catalog-chat","model":%s,"choices":[{"index":0,"delta":{},"finish_reason":%q}],"usage":{"prompt_tokens":%d,"completion_tokens":4,"total_tokens":%d,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}}`, id, stop, input, total))}, provider.FramingLF)
	case ai.APIAnthropicMessages:
		block := `{"type":"text","text":""}`
		delta := `{"type":"text_delta","text":"Catalog reply."}`
		stop := "end_turn"
		if tool {
			block = `{"type":"tool_use","id":"catalog-call","name":"catalog_tool","input":{}}`
			delta = `{"type":"input_json_delta","partial_json":"{}"}`
			stop = "tool_use"
		}
		return sseEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"type":"message_start","message":{"id":"catalog-message","model":%s,"type":"message","role":"assistant","content":[],"usage":{"input_tokens":%d,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`, id, input)), json.RawMessage(`{"type":"content_block_start","index":0,"content_block":` + block + `}`), json.RawMessage(`{"type":"content_block_delta","index":0,"delta":` + delta + `}`), json.RawMessage(`{"type":"content_block_stop","index":0}`), json.RawMessage(fmt.Sprintf(`{"type":"message_delta","delta":{"stop_reason":%q},"usage":{"input_tokens":%d,"output_tokens":4,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`, stop, input)), json.RawMessage(`{"type":"message_stop"}`)}, provider.FramingLF)
	case ai.APIGoogleGenerativeAI:
		part := `{"text":"Catalog reply."}`
		if tool {
			part = `{"functionCall":{"name":"catalog_tool","args":{}},"thoughtSignature":"catalog-thought-signature"}`
		}
		return dataEvents(t, []json.RawMessage{json.RawMessage(fmt.Sprintf(`{"candidates":[{"index":0,"content":{"role":"model","parts":[%s]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":%d,"candidatesTokenCount":4,"totalTokenCount":%d,"cachedContentTokenCount":0,"thoughtsTokenCount":0}}`, part, input, total))}, provider.FramingLF)
	}
	t.Fatalf("code_gap: chat contract unavailable for %s", m.API)
	return provider.Reply{}
}

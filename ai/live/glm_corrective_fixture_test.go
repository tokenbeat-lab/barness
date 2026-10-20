//go:build live

package live

import (
	"encoding/json"
	"strings"
)

// Bounded synthetic output, never an alteration of a vendor capture. 4094 tiny
// chunks exercise the measured quadratic Partial duplication; optional padding
// proves wire capture bounds independently of public content size.
func glmCorrectiveSyntheticResponse(name string) string {
	var body strings.Builder
	emit := func(value any) {
		raw, _ := json.Marshal(value)
		body.WriteString("data: ")
		body.Write(raw)
		body.WriteString("\n\n")
	}
	thought := map[string]any{"type": "reasoning", "id": "rs_corrective", "content": map[string]any{"type": "reasoning_text", "text": strings.Repeat("x", 4094)}}
	emit(map[string]any{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"type": "reasoning", "id": "rs_corrective", "content": []any{}}})
	padding := strings.Repeat("p", 400)
	if name == "corrective-overcap" {
		padding = strings.Repeat("p", 600)
	}
	for n := 0; n < 4094; n++ {
		emit(map[string]any{"type": "response.reasoning_text.delta", "output_index": 0, "delta": "x", "capture_padding": padding})
	}
	emit(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": thought})
	output := []any{thought}
	status := "completed"
	terminal := "response.completed"
	response := map[string]any{"id": "response_corrective", "model": "glm-5.3", "status": status, "output": output, "usage": map[string]any{"input_tokens": 49, "output_tokens": 4096, "total_tokens": 4145, "output_tokens_details": map[string]any{"reasoning_tokens": 4094}}}
	if name == "corrective-length" {
		terminal = "response.incomplete"
		response["status"] = "incomplete"
		response["incomplete_details"] = map[string]any{"reason": "max_output_tokens"}
	} else {
		text := map[string]any{"type": "message", "id": "msg_corrective", "role": "assistant", "content": map[string]any{"type": "output_text", "text": "42"}}
		emit(map[string]any{"type": "response.output_item.done", "output_index": 1, "item": text})
		response["output"] = append(output, text)
	}
	emit(map[string]any{"type": terminal, "response": response})
	return body.String()
}

func glmCorrectiveMessageOverflow() string {
	var body strings.Builder
	emit := func(value map[string]any) {
		raw, _ := json.Marshal(value)
		body.WriteString("event: ")
		body.WriteString(value["type"].(string))
		body.WriteString("\ndata: ")
		body.Write(raw)
		body.WriteString("\n\n")
	}
	emit(map[string]any{"type": "message_start", "message": map[string]any{"id": "msg_eventcap", "type": "message", "role": "assistant", "model": "glm-5.3", "content": []any{}, "usage": map[string]any{"input_tokens": 20, "output_tokens": 0}}})
	emit(map[string]any{"type": "content_block_start", "index": 0, "content_block": map[string]any{"type": "thinking", "thinking": "", "signature": ""}})
	for n := 0; n < 9000; n++ {
		emit(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "thinking_delta", "thinking": "x"}})
	}
	emit(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "signature_delta", "signature": "controlled-signature"}})
	emit(map[string]any{"type": "content_block_stop", "index": 0})
	emit(map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": "end_turn"}, "usage": map[string]any{"output_tokens": 10}})
	emit(map[string]any{"type": "message_stop"})
	return body.String()
}

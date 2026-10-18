//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"io"
	"net/http"
	"strings"
)

// Synthetic protocol peer at the transport seam. It reads requests unchanged,
// cannot network, and keeps cancellation in flight after an observable frame.
func installGLMControlled(rec *recorder, name string) {
	rec.next.(*http.Transport).CloseIdleConnections()
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		raw, _ := io.ReadAll(req.Body)
		req.Body.Close()
		var b map[string]any
		json.Unmarshal(raw, &b)
		status := 200
		body := ""
		tool := strings.Contains(string(raw), "Call the add function") && !strings.Contains(string(raw), "tool_result") && !strings.Contains(string(raw), "function_call_output") && !strings.Contains(string(raw), `"role":"tool"`)
		if name == "429" {
			status = 429
			body = `{"error":{"message":"synthetic upstream rate limit"}}`
		} else {
			emit := func(v any) { x, _ := json.Marshal(v); body += "data: " + string(x) + "\n\n" }
			switch glmSelected.api {
			case ai.APIOpenAICompletions:
				emit(map[string]any{"id": "controlled_chat", "model": b["model"], "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"reasoning_content": "Check arithmetic."}, "finish_reason": nil}}})
				if tool {
					emit(map[string]any{"model": b["model"], "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "call_controlled", "type": "function", "function": map[string]any{"name": "add", "arguments": `{"a":17,"b":25}`}}}}, "finish_reason": "tool_calls"}}})
				} else {
					emit(map[string]any{"model": b["model"], "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"content": "42"}, "finish_reason": "stop"}}})
				}
				emit(map[string]any{"model": b["model"], "choices": []any{}, "usage": map[string]any{"prompt_tokens": 20, "completion_tokens": 10, "total_tokens": 30, "prompt_tokens_details": map[string]any{"cached_tokens": 0}, "completion_tokens_details": map[string]any{"reasoning_tokens": 8}}})
				body += "data: [DONE]\n\n"
			case ai.APIOpenAIResponses:
				thinking := map[string]any{"type": "reasoning", "id": "rs_controlled", "content": map[string]any{"type": "reasoning_text", "text": "Check arithmetic."}}
				output := []any{thinking}
				emit(map[string]any{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"type": "reasoning", "id": "rs_controlled", "content": []any{}}})
				emit(map[string]any{"type": "response.reasoning_text.delta", "output_index": 0, "delta": "Check arithmetic."})
				emit(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": thinking})
				if tool {
					call := map[string]any{"type": "function_call", "id": "fc_controlled", "call_id": "call_controlled", "name": "add", "arguments": `{"a":17,"b":25}`, "status": "completed"}
					emit(map[string]any{"type": "response.output_item.done", "output_index": 1, "item": call})
					output = append(output, call)
				} else {
					text := map[string]any{"type": "message", "id": "msg_controlled", "role": "assistant", "content": map[string]any{"type": "output_text", "text": "42"}}
					emit(map[string]any{"type": "response.output_item.done", "output_index": 1, "item": text})
					output = append(output, text)
				}
				emit(map[string]any{"type": "response.completed", "response": map[string]any{"id": "response_controlled", "model": b["model"], "status": "completed", "output": output, "usage": map[string]any{"input_tokens": 20, "output_tokens": 10, "total_tokens": 30, "input_tokens_details": map[string]any{"cached_tokens": 0}, "output_tokens_details": map[string]any{"reasoning_tokens": 8}}}})
			default:
				// Anthropic SDK requires named SSE events; rewrite only our generated
				// response script framing, never the production request.
				anthropic := func(v map[string]any) {
					x, _ := json.Marshal(v)
					body += "event: " + fmt.Sprint(v["type"]) + "\ndata: " + string(x) + "\n\n"
				}
				anthropic(map[string]any{"type": "message_start", "message": map[string]any{"id": "msg_controlled", "type": "message", "role": "assistant", "model": b["model"], "content": []any{}, "usage": map[string]any{"input_tokens": 20, "output_tokens": 0}}})
				anthropic(map[string]any{"type": "content_block_start", "index": 0, "content_block": map[string]any{"type": "thinking", "thinking": "", "signature": ""}})
				anthropic(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "thinking_delta", "thinking": "Check arithmetic."}})
				anthropic(map[string]any{"type": "content_block_delta", "index": 0, "delta": map[string]any{"type": "signature_delta", "signature": "controlled-signature"}})
				anthropic(map[string]any{"type": "content_block_stop", "index": 0})
				if tool {
					anthropic(map[string]any{"type": "content_block_start", "index": 1, "content_block": map[string]any{"type": "tool_use", "id": "call_controlled", "name": "add", "input": map[string]any{}}})
					anthropic(map[string]any{"type": "content_block_delta", "index": 1, "delta": map[string]any{"type": "input_json_delta", "partial_json": `{"a":17,"b":25}`}})
				} else {
					anthropic(map[string]any{"type": "content_block_start", "index": 1, "content_block": map[string]any{"type": "text", "text": ""}})
					anthropic(map[string]any{"type": "content_block_delta", "index": 1, "delta": map[string]any{"type": "text_delta", "text": "42"}})
				}
				anthropic(map[string]any{"type": "content_block_stop", "index": 1})
				stop := "end_turn"
				if tool {
					stop = "tool_use"
				}
				anthropic(map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": stop}, "usage": map[string]any{"output_tokens": 10}})
				anthropic(map[string]any{"type": "message_stop"})
			}
		}
		if strings.HasPrefix(name, "corrective-") {
			if reasoning, ok := b["reasoning"].(map[string]any); ok && reasoning["effort"] == "low" && b["model"] == "glm-5.3" {
				body = glmCorrectiveSyntheticResponse(name)
			}
		}
		if name == "missing-text" && !tool {
			body = strings.ReplaceAll(body, `"42"`, `""`)
		}
		var responseBody io.ReadCloser = io.NopCloser(strings.NewReader(body))
		if strings.Contains(string(raw), "Write the integers") && status == 200 {
			pr, pw := io.Pipe()
			responseBody = pr
			go func() {
				idx := strings.Index(body, "\n\n")
				if glmSelected.api == ai.APIOpenAIResponses {
					idx = strings.Index(body, "response.reasoning_text.delta")
					idx += strings.Index(body[idx:], "\n\n")
				}
				if glmSelected.api == ai.APIAnthropicMessages {
					idx = strings.Index(body, "thinking_delta")
					idx += strings.Index(body[idx:], "\n\n")
				}
				io.WriteString(pw, body[:idx+2])
				<-req.Context().Done()
				pw.CloseWithError(req.Context().Err())
			}()
		}
		return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"text/event-stream"}, "X-Request-Id": {"controlled-request"}}, Body: responseBody, Request: req}, nil
	})
}

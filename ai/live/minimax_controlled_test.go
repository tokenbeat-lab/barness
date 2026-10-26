//go:build live

package live

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// MiniMax's synthetic official-schema peer is isolated at the Host transport
// seam; it has no network path and does not borrow GLM's response shapes.
func installMiniMaxControlled(rec *recorder, name string) {
	toolCalls := 0
	requests := 0
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests++
		if requests == 1 && (name == "permanent-429" || name == "unknown-429" || name == "typed-503") {
			req.Body.Close()
			status, body := 429, `{"error":{"type":"permission_error","message":"synthetic"}}`
			if name == "unknown-429" {
				body = `{"error":{"message":"synthetic"}}`
			}
			if name == "typed-503" {
				status = 503
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
		}
		if requests == 1 && name == "success-transient-transport" {
			req.Body.Close()
			return nil, fmt.Errorf("synthetic connection failure")
		}

		if name == "success-transient-request" && requests == 1 || name == "transient-exhausted" && requests <= 2 {
			return &http.Response{StatusCode: 503, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"message":"synthetic unavailable"}}`)), Request: req}, nil
		}
		raw, _ := io.ReadAll(req.Body)
		req.Body.Close()
		var b map[string]any
		json.Unmarshal(raw, &b)
		model := fmt.Sprint(b["model"])
		if name == "wrong-actual-model" {
			model = "unrequested-actual-model"
		}
		tool := strings.Contains(string(raw), "MUST call the add") && !strings.Contains(string(raw), `"role":"tool"`) && !strings.Contains(string(raw), `"type":"function_call_output"`) && !strings.Contains(string(raw), `"type":"tool_result"`)
		truncate := strings.Contains(string(raw), "long explanation of all integers")
		thinking := name != "missing-thinking"
		if minimaxReport.Identity.Model == "MiniMax-M3" {
			if minimaxReport.Identity.Route == "responses" {
				_, thinking = b["reasoning"]
			} else {
				v, _ := b["thinking"].(map[string]any)
				thinking = v["type"] == "adaptive"
			}
			if name == "missing-thinking" {
				thinking = false
			}
		}
		// Adaptive thinking may legitimately omit reasoning for a direct-answer
		// task. Keep the probe's own positive observation requirement exercised.
		if strings.HasPrefix(name, "adaptive-probe") &&
			(strings.Contains(string(raw), "Synthetic protocol probe. Reply with the integer 42.") ||
				(name == "adaptive-probe-empty" && strings.Contains(string(raw), "Synthetic reasoning task."))) {
			thinking = false
		}
		if tool {
			toolCalls++
			if name == "success-rerun-tool" && toolCalls == 1 {
				thinking = false
			}
		}
		if tool && toolCalls == 1 && (name == "tool-auth401" || name == "tool-permanent429") {
			status := 401
			if name == "tool-permanent429" {
				status = 429
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"type":"permission_error","message":"synthetic"}}`)), Request: req}, nil
		}
		body := ""
		input := 20
		if name == "known-input-overrun" {
			input = 100000
		}
		emit := func(v map[string]any) {
			if name == "success-one-hour-zero" && v["type"] == "message_start" {
				v["message"].(map[string]any)["usage"].(map[string]any)["cache_creation"] = map[string]any{"ephemeral_1h_input_tokens": 0}
			}
			x, _ := json.Marshal(v)
			if minimaxReport.Identity.Route == "messages" {
				body += "event: " + fmt.Sprint(v["type"]) + "\n"
			}
			body += "data: " + string(x) + "\n\n"
		}
		switch minimaxReport.Identity.Route {
		case "chat":
			if thinking {
				emit(map[string]any{"id": "synthetic-chat", "model": model, "choices": []any{map[string]any{"index": 0, "delta": map[string]any{"reasoning_content": "Check arithmetic."}}}})
			}
			delta := map[string]any{"content": "42"}
			stop := "stop"
			if tool {
				delta = map[string]any{"tool_calls": []any{map[string]any{"index": 0, "id": "call:原|%", "type": "function", "function": map[string]any{"name": "add", "arguments": `{"a":17,"b":25}`}}}}
				stop = "tool_calls"
			}
			if truncate {
				stop = "length"
			}
			emit(map[string]any{"id": "synthetic-chat", "model": model, "choices": []any{map[string]any{"index": 0, "delta": delta, "finish_reason": stop}}})
			emit(map[string]any{"model": model, "choices": []any{}, "usage": map[string]any{"prompt_tokens": input, "completion_tokens": 10, "total_tokens": input + 10, "prompt_tokens_details": map[string]any{"cached_tokens": 0}}})
			body += "data: [DONE]\n\n"
		case "responses":
			output := []any{}
			if thinking {
				item := map[string]any{"type": "reasoning", "id": "reasoning:原|%", "summary": []any{}, "content": []any{map[string]any{"type": "reasoning_text", "text": "Check arithmetic."}}}
				emit(map[string]any{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"type": "reasoning", "id": "reasoning:原|%", "summary": []any{}, "content": []any{}}})
				emit(map[string]any{"type": "response.reasoning_text.delta", "output_index": 0, "delta": "Check arithmetic."})
				emit(map[string]any{"type": "response.output_item.done", "output_index": 0, "item": item})
				output = append(output, item)
			}
			var item map[string]any
			if tool {
				item = map[string]any{"type": "function_call", "id": "item:原|%", "call_id": "call:原|%", "name": "add", "arguments": `{"a":17,"b":25}`, "status": "completed"}
			} else {
				item = map[string]any{"type": "message", "id": "message:原|%", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "42"}}}
			}
			if !tool {
				emit(map[string]any{"type": "response.output_item.added", "output_index": len(output), "item": map[string]any{"type": "message", "id": "message:原|%", "role": "assistant", "content": []any{}}})
				emit(map[string]any{"type": "response.output_text.delta", "output_index": len(output), "delta": "42"})
			}
			emit(map[string]any{"type": "response.output_item.done", "output_index": len(output), "item": item})
			output = append(output, item)
			status, typ := "completed", "response.completed"
			response := map[string]any{"id": "synthetic-response", "model": model, "status": status, "output": output, "usage": map[string]any{"input_tokens": input, "output_tokens": 10, "total_tokens": input + 10, "input_tokens_details": map[string]any{"cached_tokens": 0}}}
			if truncate {
				typ = "response.incomplete"
				response["status"] = "incomplete"
				response["incomplete_details"] = map[string]any{"reason": "max_output_tokens"}
			}
			emit(map[string]any{"type": typ, "response": response})
		default:
			emit(map[string]any{"type": "message_start", "message": map[string]any{"id": "synthetic-message", "type": "message", "role": "assistant", "model": model, "content": []any{}, "usage": map[string]any{"input_tokens": input, "output_tokens": 0}}})
			index := 0
			if thinking {
				emit(map[string]any{"type": "content_block_start", "index": index, "content_block": map[string]any{"type": "thinking", "thinking": ""}})
				emit(map[string]any{"type": "content_block_delta", "index": index, "delta": map[string]any{"type": "thinking_delta", "thinking": "Check arithmetic."}})
				emit(map[string]any{"type": "content_block_delta", "index": index, "delta": map[string]any{"type": "signature_delta", "signature": "synthetic-signature"}})
				emit(map[string]any{"type": "content_block_stop", "index": index})
				index++
			}
			if tool {
				emit(map[string]any{"type": "content_block_start", "index": index, "content_block": map[string]any{"type": "tool_use", "id": "call:原|%", "name": "add", "input": map[string]any{}}})
				emit(map[string]any{"type": "content_block_delta", "index": index, "delta": map[string]any{"type": "input_json_delta", "partial_json": `{"a":17,"b":25}`}})
			} else {
				emit(map[string]any{"type": "content_block_start", "index": index, "content_block": map[string]any{"type": "text", "text": ""}})
				emit(map[string]any{"type": "content_block_delta", "index": index, "delta": map[string]any{"type": "text_delta", "text": "42"}})
			}
			emit(map[string]any{"type": "content_block_stop", "index": index})
			stop := "end_turn"
			if tool {
				stop = "tool_use"
			}
			if truncate {
				stop = "max_tokens"
			}
			emit(map[string]any{"type": "message_delta", "delta": map[string]any{"stop_reason": stop}, "usage": map[string]any{"output_tokens": 10}})
			emit(map[string]any{"type": "message_stop"})
		}
		if tool && toolCalls == 1 && name == "tool-stream-error" {
			body = `data: {"model":"MiniMax-M3","choices":[{"index":0,"delta":{"reasoning_content":"Partial synthetic reasoning."}}]}` + "\n\n" +
				`data: {"error":{"message":"synthetic stream failure","code":"synthetic_error"}}` + "\n\n"
		}
		if name == "missing-terminal" || name == "transient-stream" {
			idx := strings.LastIndex(body, "data:")
			if idx > 0 {
				body = body[:idx]
			}
			if minimaxReport.Identity.Route == "chat" {
				body = strings.ReplaceAll(body, `"finish_reason":"stop"`, `"finish_reason":null`)
			}
		}
		var reader io.ReadCloser = io.NopCloser(strings.NewReader(body))
		if strings.Contains(string(raw), "Write the integers") {
			pr, pw := io.Pipe()
			reader = pr
			prefix := body
			marker := "\n\n"
			idx := strings.Index(body, marker)
			if minimaxReport.Identity.Route == "messages" {
				pos := strings.Index(body, "thinking_delta")
				if pos < 0 {
					pos = strings.Index(body, "text_delta")
				}
				idx = pos + strings.Index(body[pos:], marker)
			}
			if minimaxReport.Identity.Route == "responses" {
				pos := strings.Index(body, "response.reasoning_text.delta")
				if pos < 0 {
					pos = strings.Index(body, "response.output_text.delta")
				}
				idx = pos + strings.Index(body[pos:], marker)
			}
			if idx >= 0 {
				prefix = body[:idx+2]
			}
			go func() { io.WriteString(pw, prefix); <-req.Context().Done(); pw.CloseWithError(req.Context().Err()) }()
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"text/event-stream"}, "X-Request-Id": {"synthetic-provider-id"}}, Body: reader, Request: req}, nil
	})
}

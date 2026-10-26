package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func prependXAIReviewFrames(t *testing.T, dir string, makeFrames func(map[string]any) []map[string]any) {
	t.Helper()
	editXAINativeArray(t, dir, "E17-xai-entry-stream/text-http.json", func(v []any) {
		e := v[0].(map[string]any)
		body := e["responseBody"].(string)
		var terminal map[string]any
		for _, chunk := range strings.Split(body, "\n\n") {
			for _, line := range strings.Split(chunk, "\n") {
				if strings.HasPrefix(line, "data: {") {
					var f map[string]any
					mustUnmarshal(t, []byte(strings.TrimPrefix(line, "data: ")), &f)
					if f["type"] == "response.completed" {
						terminal = f["response"].(map[string]any)
					}
				}
			}
		}
		if terminal == nil {
			t.Fatal("synthetic terminal missing")
		}
		prefix := ""
		for _, f := range makeFrames(terminal) {
			b, err := json.Marshal(f)
			if err != nil {
				t.Fatal(err)
			}
			prefix += "data: " + string(b) + "\n\n"
		}
		e["responseBody"] = prefix + body
	})
}
func TestXAIFormalNativeReviewSSE(t *testing.T) {
	frame := func(kind string, extra map[string]any) map[string]any { extra["type"] = kind; return extra }
	withMessage := func(r map[string]any) map[string]any { out := r["output"].([]any); return out[1].(map[string]any) }
	add := func(r map[string]any) map[string]any {
		m := withMessage(r)
		return frame("response.output_item.added", map[string]any{"output_index": 1, "item": map[string]any{"type": "message", "id": m["id"], "role": "assistant", "content": []any{}}})
	}
	runXAINativeReview(t, "sse", []xaiNativeReviewCase{
		{"final-only", true, nil},
		{"done-only", true, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				return []map[string]any{frame("response.output_item.done", map[string]any{"output_index": 0, "item": r["output"].([]any)[0]}), frame("response.output_item.done", map[string]any{"output_index": 1, "item": withMessage(r)})}
			})
		}},
		{"incremental-prefix", true, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				reason := r["output"].([]any)[0].(map[string]any)
				addedReason := map[string]any{"type": "reasoning", "id": reason["id"], "summary": []any{}}
				return []map[string]any{frame("response.output_item.added", map[string]any{"output_index": 0, "item": addedReason}), add(r), frame("response.output_text.delta", map[string]any{"output_index": 1, "content_index": 0, "item_id": withMessage(r)["id"], "delta": "391."})}
			})
		}},
		{"alternate-channel-out-of-range", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				reason := r["output"].([]any)[0].(map[string]any)
				return []map[string]any{frame("response.output_item.added", map[string]any{"output_index": 0, "item": map[string]any{"type": "reasoning", "id": reason["id"], "summary": []any{}}}), frame("response.reasoning_text.delta", map[string]any{"output_index": 0, "content_index": 1, "item_id": reason["id"], "delta": "\x00"}), add(r)}
			})
		}},
		{"new-channel-after-item-done", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				reason := r["output"].([]any)[0].(map[string]any)
				text := reason["summary"].([]any)[0].(map[string]any)["text"]
				return []map[string]any{frame("response.output_item.done", map[string]any{"output_index": 0, "item": reason}), frame("response.reasoning_text.done", map[string]any{"output_index": 0, "content_index": 0, "item_id": reason["id"], "text": text}), frame("response.output_item.done", map[string]any{"output_index": 1, "item": withMessage(r)})}
			})
		}},

		{"unknown-event", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				return []map[string]any{frame("response.magic", map[string]any{})}
			})
		}},
		{"orphan-delta", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				return []map[string]any{frame("response.output_text.delta", map[string]any{"output_index": 1, "content_index": 0, "item_id": withMessage(r)["id"], "delta": "391."})}
			})
		}},
		{"wrong-associated-id", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				return []map[string]any{add(r), frame("response.output_text.delta", map[string]any{"output_index": 1, "content_index": 0, "item_id": "SYNTHETIC_wrong", "delta": "391."})}
			})
		}},
		{"contradictory-prefix", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				return []map[string]any{add(r), frame("response.output_text.delta", map[string]any{"output_index": 1, "content_index": 0, "item_id": withMessage(r)["id"], "delta": "SYNTHETIC_wrong"})}
			})
		}},
		{"done-authority-changed", false, func(d string) {
			prependXAIReviewFrames(t, d, func(r map[string]any) []map[string]any {
				m := withMessage(r)
				m["content"].([]any)[0].(map[string]any)["text"] = "SYNTHETIC_wrong"
				return []map[string]any{frame("response.output_item.done", map[string]any{"output_index": 1, "item": m})}
			})
		}},
	})
}

package e2e

import (
	"encoding/json"
	"testing"
)

func rewriteXAIReviewReasoning(t *testing.T, dir, prefix string, edit func(map[string]any), continued bool) {
	t.Helper()
	var native map[string]any
	editXAINativeTerminalAt(t, dir, prefix+"-http.json", func(v map[string]any) {
		for _, raw := range v["output"].([]any) {
			item := raw.(map[string]any)
			if item["type"] == "reasoning" {
				edit(item)
				native = item
				break
			}
		}
	})
	encoded, err := json.Marshal(native)
	if err != nil {
		t.Fatal(err)
	}
	mutateXAIReviewViews(t, dir, prefix, func(v []any) {
		for _, raw := range v {
			b := raw.(map[string]any)
			if _, ok := b["thinkingSignature"]; ok {
				b["thinkingSignature"] = string(encoded)
			}
		}
	})
	if !continued {
		return
	}
	editXAINativeJSON(t, dir, "E17-xai-encrypted-continuation/persisted-turn.json", func(v map[string]any) {
		for _, raw := range v["blocks"].([]any) {
			b := raw.(map[string]any)
			if b["kind"] == "thinking" {
				b["data"].(map[string]any)["thinkingSignature"] = string(encoded)
			}
		}
	})
	editXAINativeHTTPBody(t, dir, "E17-xai-encrypted-continuation/continued-http.json", func(v map[string]any) {
		input := v["input"].([]any)
		for i, raw := range input {
			if raw.(map[string]any)["type"] == "reasoning" {
				input[i] = native
				break
			}
		}
	})
	// The helper only changes the synthetic copy; the retained source is immutable.
}
func TestXAIFormalNativeReviewReasoning(t *testing.T) {
	entry := func(d string, edit func(map[string]any)) {
		rewriteXAIReviewReasoning(t, d, "E17-xai-entry-complete/text", edit, false)
	}
	runXAINativeReview(t, "reasoning", []xaiNativeReviewCase{
		{"null-cipher-unrelated", true, func(d string) { entry(d, func(v map[string]any) { v["encrypted_content"] = nil }) }},
		{"missing-cipher-unrelated", true, func(d string) { entry(d, func(v map[string]any) { delete(v, "encrypted_content") }) }},
		{"empty-cipher-unrelated", true, func(d string) { entry(d, func(v map[string]any) { v["encrypted_content"] = "" }) }},
		{"opaque-nonresource-extension", true, func(d string) {
			entry(d, func(v map[string]any) {
				v["future"] = map[string]any{"description": "file_id is a business word", "large": json.Number("1e400")}
			})
		}},
		{"resource-extension", false, func(d string) {
			entry(d, func(v map[string]any) { v["future"] = map[string]any{"file_id": "SYNTHETIC_foreign"} })
		}},
		{"item-reference-extension", false, func(d string) {
			entry(d, func(v map[string]any) {
				v["future"] = map[string]any{"type": "item_reference", "id": "SYNTHETIC_foreign"}
			})
		}},
		{"dedicated-encrypted-proof-empty", false, func(d string) {
			rewriteXAIReviewReasoning(t, d, "E17-xai-encrypted-continuation/first", func(v map[string]any) { v["encrypted_content"] = "" }, true)
		}},
	})
}

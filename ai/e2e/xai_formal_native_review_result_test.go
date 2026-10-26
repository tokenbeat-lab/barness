package e2e

import (
	"os"
	"testing"
)

func mutateXAIReviewViews(t *testing.T, dir, prefix string, change func([]any)) {
	t.Helper()
	editXAINativeJSON(t, dir, prefix+"-result.json", func(v map[string]any) { change(v["message"].(map[string]any)["content"].([]any)) })
	path := dir + "/" + prefix + "-events.json"
	if _, err := os.Stat(path); err != nil {
		return
	}
	editXAINativeArray(t, dir, prefix+"-events.json", func(events []any) {
		for _, raw := range events {
			e := raw.(map[string]any)
			for _, key := range []string{"Partial", "Message"} {
				if m, ok := e[key].(map[string]any); ok {
					change(m["content"].([]any))
				}
			}
			if call, ok := e["ToolCall"].(map[string]any); ok {
				change([]any{call})
			}
		}
	})
}
func TestXAIFormalNativeReviewResult(t *testing.T) {
	runXAINativeReview(t, "result", []xaiNativeReviewCase{
		{"complete", true, nil},
		{"visible-thinking-changed", false, func(d string) {
			mutateXAIReviewViews(t, d, "E17-xai-entry-complete/text", func(v []any) { v[0].(map[string]any)["thinking"] = "SYNTHETIC_wrong" })
		}},
		{"visible-thinking-redacted", false, func(d string) {
			mutateXAIReviewViews(t, d, "E17-xai-entry-complete/text", func(v []any) { v[0].(map[string]any)["redacted"] = true })
		}},
		{"cross-kind-order", false, func(d string) {
			mutateXAIReviewViews(t, d, "E17-xai-entry-complete/text", func(v []any) { v[0], v[1] = v[1], v[0] })
		}},
		{"display-arguments-changed", false, func(d string) {
			mutateXAIReviewViews(t, d, "E17-xai-tools-required/calls", func(v []any) {
				for _, raw := range v {
					b := raw.(map[string]any)
					if _, ok := b["arguments"]; ok {
						b["arguments"] = map[string]any{"city": "SYNTHETIC_wrong"}
					}
				}
			})
		}},
	})
}

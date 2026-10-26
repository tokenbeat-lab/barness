package e2e

import (
	"path/filepath"
	"testing"
)

func TestXAIFormalNativeReviewTools(t *testing.T) {
	runXAINativeReview(t, "tools", []xaiNativeReviewCase{
		{"reversed-parallel-results", true, nil},
		{"results-choice-auto", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-tools-required/results-http.json", func(v map[string]any) { v["tool_choice"] = "auto" })
		}},
		{"parallel-marker-swap", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-tools-parallel/results-http.json", func(v map[string]any) {
				var a, b map[string]any
				for _, raw := range v["input"].([]any) {
					m := raw.(map[string]any)
					if m["type"] == "function_call_output" {
						if a == nil {
							a = m
						} else {
							b = m
						}
					}
				}
				a["output"], b["output"] = b["output"], a["output"]
			})
		}},
		{"host-marker-changed", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-tools-required/results-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					m := raw.(map[string]any)
					if m["type"] == "function_call_output" {
						m["output"] = "SYNTHETIC_wrong"
					}
				}
			})
		}},
		{"results-calls-again", false, func(d string) {
			first := readXAINativeObject(t, filepath.Join(d, "E17-xai-tools-required/calls-result.json"))
			content := first["message"].(map[string]any)["content"]
			editXAINativeJSON(t, d, "E17-xai-tools-required/results-result.json", func(v map[string]any) {
				m := v["message"].(map[string]any)
				m["content"] = content
				m["stopReason"] = "toolUse"
			})
			editXAINativeArray(t, d, "E17-xai-tools-required/results-observer.json", func(v []any) { v[3].(map[string]any)["stopReason"] = "toolUse" })
			var output any
			editXAINativeTerminalAt(t, d, "E17-xai-tools-required/calls-http.json", func(v map[string]any) { output = v["output"] })
			editXAINativeTerminalAt(t, d, "E17-xai-tools-required/results-http.json", func(v map[string]any) { v["output"] = output })
		}},
	})
}

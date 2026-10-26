package e2e

import "testing"

func TestXAIFormalNativeReviewUsage(t *testing.T) {
	edit := func(d string, f func(map[string]any)) { editXAINativeTerminal(t, d, f) }
	runXAINativeReview(t, "usage", []xaiNativeReviewCase{
		{"retained-real-format", true, func(d string) { useXAIReviewRetainedEntry(t, d) }},
		{"default-tier", true, func(d string) { edit(d, func(r map[string]any) { r["service_tier"] = "default" }) }},
		{"null-tier", true, func(d string) { edit(d, func(r map[string]any) { r["service_tier"] = nil }) }},
		{"invoice-context-metadata", true, func(d string) {
			edit(d, func(r map[string]any) {
				u := r["usage"].(map[string]any)
				u["cost_in_usd_ticks"] = 12345
				u["context_details"] = map[string]any{"input_tokens": 32, "output_tokens": 1109}
				u["num_sources_used"] = 0
				u["num_server_side_tools_used"] = 0
				u["cache_write_tokens"] = 0
				u["server_side_tool_usage"] = map[string]any{}
				u["input_tokens_details"].(map[string]any)["audio_tokens"] = 0
			})
		}},
		{"priority-tier", false, func(d string) { edit(d, func(r map[string]any) { r["service_tier"] = "priority" }) }},
		{"fast-tier", false, func(d string) { edit(d, func(r map[string]any) { r["service_tier"] = "fast" }) }},
		{"cache-write", false, func(d string) {
			edit(d, func(r map[string]any) { r["usage"].(map[string]any)["cache_write_tokens"] = 1 })
		}},
		{"hosted-tool-usage", false, func(d string) {
			edit(d, func(r map[string]any) {
				r["usage"].(map[string]any)["server_side_tool_usage"] = map[string]any{"web_search": 1}
			})
		}},
		{"hosted-source-count", false, func(d string) {
			edit(d, func(r map[string]any) { r["usage"].(map[string]any)["num_sources_used"] = 1 })
		}},
		{"hosted-tool-count", false, func(d string) {
			edit(d, func(r map[string]any) { r["usage"].(map[string]any)["num_server_side_tools_used"] = 1 })
		}},
		{"extra-input-billing", false, func(d string) {
			edit(d, func(r map[string]any) {
				r["usage"].(map[string]any)["input_tokens_details"].(map[string]any)["audio_tokens"] = 1
			})
		}},
		{"extra-output-billing", false, func(d string) {
			edit(d, func(r map[string]any) {
				r["usage"].(map[string]any)["output_tokens_details"].(map[string]any)["prediction_tokens"] = 1
			})
		}},
	})
}

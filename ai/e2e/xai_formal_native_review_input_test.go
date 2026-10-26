package e2e

import (
	"encoding/json"
	"testing"
)

func TestXAIFormalNativeReviewInput(t *testing.T) {
	initial := func(d string, edit func(map[string]any)) {
		editXAINativeHTTPBody(t, d, "E17-xai-entry-complete/text-http.json", edit)
	}
	pair := func(v map[string]any, edit func(map[string]any, map[string]any)) {
		call := map[string]any{"type": "function_call", "call_id": "SYNTHETIC_history", "id": "SYNTHETIC_item", "name": "get_weather", "arguments": `{"file_id":"business value"}`}
		output := map[string]any{"type": "function_call_output", "call_id": "SYNTHETIC_history", "output": `{"file_id":"business text"}`}
		edit(call, output)
		v["input"] = append(v["input"].([]any), call, output)
	}
	runXAINativeReview(t, "input", []xaiNativeReviewCase{
		{"business-json-text", true, func(d string) {
			initial(d, func(v map[string]any) { pair(v, func(map[string]any, map[string]any) {}) })
		}},
		{"business-schema-property", true, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-tools-required/calls-http.json", func(v map[string]any) {
				v["tools"].([]any)[0].(map[string]any)["parameters"].(map[string]any)["properties"].(map[string]any)["file_id"] = map[string]any{"type": "string"}
			})
		}},
		{"reasoning-opaque-extension", true, func(d string) {
			initial(d, func(v map[string]any) {
				v["input"] = append(v["input"].([]any), map[string]any{"type": "reasoning", "summary": []any{}, "encrypted_content": nil, "future": map[string]any{"quantity": json.Number("1e400"), "description": "file_id is text"}})
			})
		}},
		{"native-output-file-id", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-tools-required/results-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					m := raw.(map[string]any)
					if m["type"] == "function_call_output" {
						m["file_id"] = "SYNTHETIC_foreign"
					}
				}
			})
		}},
		{"native-call-control", false, func(d string) {
			initial(d, func(v map[string]any) { pair(v, func(c, o map[string]any) { c["store"] = true }) })
		}},
		{"native-output-control", false, func(d string) {
			initial(d, func(v map[string]any) {
				pair(v, func(c, o map[string]any) { o["previous_response_id"] = "SYNTHETIC_foreign" })
			})
		}},
		{"native-output-object", false, func(d string) {
			initial(d, func(v map[string]any) {
				pair(v, func(c, o map[string]any) { o["output"] = map[string]any{"file_id": "SYNTHETIC_foreign"} })
			})
		}},
		{"native-reasoning-resource", false, func(d string) {
			initial(d, func(v map[string]any) {
				v["input"] = append(v["input"].([]any), map[string]any{"type": "reasoning", "summary": []any{}, "future": map[string]any{"file_id": "SYNTHETIC_foreign"}})
			})
		}},
		{"message-control", false, func(d string) {
			initial(d, func(v map[string]any) { v["input"].([]any)[0].(map[string]any)["store"] = true })
		}},
		{"image-resource", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-user-png/image-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					m := raw.(map[string]any)
					if parts, ok := m["content"].([]any); ok {
						for _, raw := range parts {
							p := raw.(map[string]any)
							if p["type"] == "input_image" {
								p["file_id"] = "SYNTHETIC_foreign"
							}
						}
					}
				}
			})
		}},
		{"image-detail", false, func(d string) {
			editXAINativeHTTPBody(t, d, "E17-xai-user-png/image-http.json", func(v map[string]any) {
				for _, raw := range v["input"].([]any) {
					m := raw.(map[string]any)
					if parts, ok := m["content"].([]any); ok {
						for _, raw := range parts {
							p := raw.(map[string]any)
							if p["type"] == "input_image" {
								p["detail"] = "high"
							}
						}
					}
				}
			})
		}},
	})
}

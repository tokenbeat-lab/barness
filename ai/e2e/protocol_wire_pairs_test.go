package e2e

import "testing"

// Inspect native reasoning and tool identities only at captured public wire
// boundaries. The three protocol containers are shared by provider suites;
// this helper does not infer provider capability or native-state trust.
func protocolWirePairs(t *testing.T, protocol string, raw []byte) (native int, calls, results []string) {
	t.Helper()
	body := bodyFields(t, raw)
	var input []map[string]any
	if protocol == "responses" {
		mustUnmarshal(t, body["input"], &input)
	} else {
		mustUnmarshal(t, body["messages"], &input)
	}
	for _, item := range input {
		if protocol == "responses" {
			switch item["type"] {
			case "reasoning":
				native++
			case "function_call":
				calls = append(calls, item["call_id"].(string))
			case "function_call_output":
				results = append(results, item["call_id"].(string))
			}
			continue
		}
		if protocol == "chat" {
			if v, ok := item["reasoning_content"].(string); ok && v != "" {
				native++
			}
			if list, ok := item["tool_calls"].([]any); ok {
				for _, v := range list {
					calls = append(calls, v.(map[string]any)["id"].(string))
				}
			}
			if item["role"] == "tool" {
				results = append(results, item["tool_call_id"].(string))
			}
			continue
		}
		content, _ := item["content"].([]any)
		for _, v := range content {
			b := v.(map[string]any)
			switch b["type"] {
			case "thinking":
				native++
			case "tool_use":
				calls = append(calls, b["id"].(string))
			case "tool_result":
				results = append(results, b["tool_use_id"].(string))
			}
		}
	}
	return
}

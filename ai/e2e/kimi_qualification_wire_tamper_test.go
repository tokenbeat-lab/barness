package e2e

import (
	"encoding/json"
	"strings"
	"testing"
)

func kimiQualificationWireTamper(t *testing.T, value any, combo, mode, file string) {
	t.Helper()
	messages, responses := strings.HasSuffix(combo, "messages"), strings.HasSuffix(combo, "responses")
	if file == "options.json" {
		field := "reasoningEffort"
		if messages {
			field = "effort"
		}
		value.(map[string]any)[field] = "low"
		return
	}
	x := value.([]any)[0].(map[string]any)
	var body map[string]any
	mustUnmarshal(t, []byte(x["requestBody"].(string)), &body)
	switch mode {
	case "wire-budget":
		field := "max_completion_tokens"
		if messages {
			field = "max_tokens"
		}
		if responses {
			field = "max_output_tokens"
		}
		body[field] = 1
	case "wire-tool-choice":
		if responses {
			body["tool_choice"] = "required"
		} else {
			delete(body, "tool_choice")
		}
	case "wire-tools":
		body["tools"] = []any{}
	case "wire-unsupported":
		body["unintegrated_private_control"] = true
	default:
		effort := "low"
		if mode == "wire-simple-low" {
			effort = "high"
		}
		if mode == "wire-default-effort" {
			effort = "max"
		}
		field := "reasoning_effort"
		var v any = effort
		if messages {
			field = "output_config"
			v = map[string]any{"effort": effort}
		}
		if responses {
			field = "reasoning"
			v = map[string]any{"effort": effort}
		}
		if mode == "wire-effort-removed" {
			delete(body, field)
		} else {
			body[field] = v
		}
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	x["requestBody"] = string(raw)
}

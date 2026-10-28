package grokhistory

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// Verify independently checks actual sent protocol fields. A safe opaque
// correlation value may encode source identity (MiniMax/Kimi Responses); that
// is different from forwarding a foreign native resource id or ciphertext.
func Verify(api ai.API, raw []byte) error {
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil {
		return fmt.Errorf("xAI history: invalid request JSON")
	}
	var calls, results, outputs, names []string
	var arguments []json.RawMessage
	addCall := func(id, name string, args json.RawMessage, encoded bool) {
		calls, names = append(calls, id), append(names, name)
		if encoded {
			var text string
			if json.Unmarshal(args, &text) != nil {
				args = nil
			} else {
				args = json.RawMessage(text)
			}
		}
		arguments = append(arguments, args)
	}
	addResult := func(id string, content any) {
		text, _ := json.Marshal(content)
		results = append(results, id)
		outputs = append(outputs, string(text))
	}
	switch api {
	case ai.APIOpenAIResponses:
		var input []map[string]json.RawMessage
		if json.Unmarshal(body["input"], &input) != nil {
			return fmt.Errorf("xAI history: missing Responses input")
		}
		for _, item := range input {
			switch str(item["type"]) {
			case "function_call":
				if item["id"] != nil {
					return fmt.Errorf("xAI history: foreign function resource id")
				}
				addCall(str(item["call_id"]), str(item["name"]), item["arguments"], true)
			case "function_call_output":
				addResult(str(item["call_id"]), item["output"])
			}
		}
	case ai.APIOpenAICompletions:
		var messages []struct {
			Role      string
			ToolCalls []struct {
				ID       string
				Function struct {
					Name      string
					Arguments json.RawMessage
				}
			} `json:"tool_calls"`
			ToolCallID string `json:"tool_call_id"`
			Content    json.RawMessage
		}
		if json.Unmarshal(body["messages"], &messages) != nil {
			return fmt.Errorf("xAI history: missing Chat messages")
		}
		for _, m := range messages {
			for _, c := range m.ToolCalls {
				addCall(c.ID, c.Function.Name, c.Function.Arguments, true)
			}
			if m.Role == "tool" {
				addResult(m.ToolCallID, m.Content)
			}
		}
	case ai.APIAnthropicMessages:
		var messages []struct{ Content []map[string]json.RawMessage }
		if json.Unmarshal(body["messages"], &messages) != nil {
			return fmt.Errorf("xAI history: missing Messages content")
		}
		for _, m := range messages {
			for _, b := range m.Content {
				switch str(b["type"]) {
				case "tool_use":
					addCall(str(b["id"]), str(b["name"]), b["input"], false)
				case "tool_result":
					addResult(str(b["tool_use_id"]), b["content"])
				}
			}
		}
	case ai.APIGoogleGenerativeAI:
		var contents []struct {
			Parts []struct {
				Call *struct {
					ID, Name string
					Args     json.RawMessage
				} `json:"functionCall"`
				Result *struct {
					ID, Name string
					Response json.RawMessage
				} `json:"functionResponse"`
			}
		}
		if json.Unmarshal(body["contents"], &contents) != nil {
			return fmt.Errorf("xAI history: missing Gemini contents")
		}
		for _, c := range contents {
			for _, p := range c.Parts {
				if p.Call != nil {
					addCall(p.Call.ID, p.Call.Name, p.Call.Args, false)
				}
				if p.Result != nil {
					if p.Result.Name != "inspect_xai_history" {
						return fmt.Errorf("xAI history: result function name differs")
					}
					addResult(p.Result.ID, p.Result.Response)
				}
			}
		}
	default:
		return fmt.Errorf("xAI history: unsupported proof protocol %s", api)
	}
	if len(calls) != 3 || !slices.Equal(calls, results) || len(outputs) != 3 {
		return fmt.Errorf("xAI history: three ordered call/result associations required")
	}
	// These values are the fixed public qualification contract, not values
	// inferred from the captured IDs or the producer's claimed PASS.
	labels := []string{"first", "second", "first"}
	for i := range calls {
		var args map[string]json.RawMessage
		if names[i] != "inspect_xai_history" || json.Unmarshal(arguments[i], &args) != nil || len(args) != 1 || str(args["label"]) != labels[i] {
			return fmt.Errorf("xAI history: function name or complete arguments differ at call %d", i)
		}
	}
	seen := map[string]bool{}
	for i, id := range calls {
		if id == "" || seen[id] {
			return fmt.Errorf("xAI history: call identities empty or collide")
		}
		seen[id] = true
		for _, r := range id {
			if r < 33 || r > 126 {
				return fmt.Errorf("xAI history: nonprintable or non-ASCII association")
			}
		}
		// Responses has destination-specific opaque association syntax. The
		// remaining codecs have the reviewed 64-character ASCII identifier rule.
		if api != ai.APIOpenAIResponses {
			if len(id) > 64 {
				return fmt.Errorf("xAI history: identifier exceeds protocol bound")
			}
			for _, r := range id {
				if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
					return fmt.Errorf("xAI history: illegal protocol identifier")
				}
			}
		}
		if !strings.Contains(outputs[i], resultMarkers[i]) {
			return fmt.Errorf("xAI history: result marker/order differs")
		}
	}
	text := string(raw)
	if !strings.Contains(text, visibleReasoning) || !strings.Contains(text, visibleAnswer) {
		return fmt.Errorf("xAI history: visible source content missing")
	}
	for _, secret := range []string{cipher, reasoningID, messageID} {
		if strings.Contains(text, secret) {
			return fmt.Errorf("xAI history: foreign native state leaked")
		}
	}
	// The top-level include selector requests the destination's own future
	// output. Only an encrypted_content field inside replayed history is foreign
	// native state; substring matching would reject valid full OpenAI options.
	for _, key := range []string{"input", "messages", "contents"} {
		var history any
		if value := body[key]; value != nil && json.Unmarshal(value, &history) == nil && encryptedHistoryField(history) {
			return fmt.Errorf("xAI history: foreign encrypted history field")
		}
	}
	return nil
}

func encryptedHistoryField(value any) bool {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			if key == "encrypted_content" || encryptedHistoryField(child) {
				return true
			}
		}
	case []any:
		for _, child := range value {
			if encryptedHistoryField(child) {
				return true
			}
		}
	}
	return false
}

func str(raw json.RawMessage) string { var out string; _ = json.Unmarshal(raw, &out); return out }

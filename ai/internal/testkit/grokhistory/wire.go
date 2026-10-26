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
	var calls, results, outputs []string
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
				calls = append(calls, str(item["call_id"]))
			case "function_call_output":
				addResult(str(item["call_id"]), item["output"])
			}
		}
	case ai.APIOpenAICompletions:
		var messages []struct {
			Role       string
			ToolCalls  []struct{ ID string } `json:"tool_calls"`
			ToolCallID string                `json:"tool_call_id"`
			Content    json.RawMessage
		}
		if json.Unmarshal(body["messages"], &messages) != nil {
			return fmt.Errorf("xAI history: missing Chat messages")
		}
		for _, m := range messages {
			for _, c := range m.ToolCalls {
				calls = append(calls, c.ID)
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
					calls = append(calls, str(b["id"]))
				case "tool_result":
					addResult(str(b["tool_use_id"]), b["content"])
				}
			}
		}
	case ai.APIGoogleGenerativeAI:
		var contents []struct {
			Parts []struct {
				Call   *struct{ ID string } `json:"functionCall"`
				Result *struct {
					ID       string
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
					calls = append(calls, p.Call.ID)
				}
				if p.Result != nil {
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
	seen := map[string]bool{}
	for i, id := range calls {
		if id == "" || seen[id] {
			return fmt.Errorf("xAI history: call identities empty or collide")
		}
		seen[id] = true
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
	for _, secret := range []string{cipher, reasoningID, messageID, "encrypted_content"} {
		if strings.Contains(text, secret) {
			return fmt.Errorf("xAI history: foreign native state leaked")
		}
	}
	return nil
}

func str(raw json.RawMessage) string { var out string; _ = json.Unmarshal(raw, &out); return out }

package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"
)

// reasoning is stored as the complete native item, including documented
// ciphertext and future non-resource JSON fields. Validation never rebuilds
// it through the SDK's narrower union, or manufactures native state from text.
func xaiResponsesReasoning(raw json.RawMessage) (json.RawMessage, error) {
	invalid := errors.New("invalid xAI native reasoning item")
	var item map[string]json.RawMessage
	if json.Unmarshal(raw, &item) != nil || item == nil {
		return nil, invalid
	}
	typ, ok := rawString(item["type"])
	if !ok || typ != "reasoning" {
		return nil, invalid
	}
	for _, key := range []string{"id", "status", "encrypted_content"} {
		if v, present := item[key]; present {
			if _, ok := rawString(v); !ok {
				return nil, invalid
			}
		}
	}
	for _, key := range []string{"summary", "content"} {
		if v, present := item[key]; present {
			var parts []map[string]json.RawMessage
			if json.Unmarshal(v, &parts) != nil || parts == nil {
				return nil, invalid
			}
			for _, part := range parts {
				kind, ok := rawString(part["type"])
				want := "summary_text"
				if key == "content" {
					want = "reasoning_text"
				}
				if !ok || kind != want {
					return nil, invalid
				}
				if _, ok := rawString(part["text"]); !ok {
					return nil, invalid
				}
			}
		}
	}
	var decoded map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil || vendorReference(decoded) != "" {
		return nil, invalid
	}
	return slices.Clone(raw), nil
}

func xaiResponsesAssistantItems(m replayedAssistant) ([]any, error) {
	var items []any
	for _, block := range m.Content {
		switch b := block.(type) {
		case Thinking:
			if b.Signature == "" {
				if b.Thinking != "" {
					items = append(items, map[string]any{"type": "message", "role": "assistant", "content": []responsesInputText{{Type: "output_text", Text: b.Thinking}}})
				}
				continue
			}
			if b.Redacted {
				return nil, errors.New("xAI does not accept redacted reasoning")
			}
			native, err := xaiResponsesReasoning(json.RawMessage(b.Signature))
			if err != nil {
				return nil, err
			}
			items = append(items, native)
		case Text:
			item := map[string]any{"type": "message", "role": "assistant", "content": []responsesInputText{{Type: "output_text", Text: b.Text}}}
			id, _ := parseTextSignature(b.Signature)
			if id != "" {
				item["id"] = id
			}
			items = append(items, item)
		case ToolCall:
			call, item, err := xaiResponsesIDPair(b.ID)
			if err != nil {
				return nil, err
			}
			if !xaiArgumentsObject(string(b.Arguments)) {
				return nil, errors.New("xAI tool arguments must be a complete JSON object")
			}
			if !m.sameModel {
				item = ""
			}
			items = append(items, responsesFunctionCall{Type: "function_call", CallID: call, ID: item, Name: b.Name, Arguments: string(b.Arguments)})
		}
	}
	return items, nil
}
func xaiArgumentsObject(raw string) bool {
	var obj map[string]json.RawMessage
	return json.Unmarshal([]byte(raw), &obj) == nil && obj != nil
}

// Check the Host's source transcript before shared normalization can supply a
// missing result. A complete xAI transcript has one unambiguous answer per call.
func validateXAIHistory(req Request) string {
	pending := map[string]string{}
	answered := map[string]bool{}
	closePending := func() bool {
		for id := range pending {
			if !answered[id] {
				return false
			}
		}
		return true
	}
	for _, m := range req.Messages {
		switch m := m.(type) {
		case AssistantMessage:
			if !closePending() {
				return "xAI history contains a tool call without a result"
			}
			pending = map[string]string{}
			answered = map[string]bool{}
			if m.StopReason == StopReasonError || m.StopReason == StopReasonAborted {
				continue
			}
			calls := map[string]bool{}
			for _, b := range m.Content {
				if call, ok := b.(ToolCall); ok {
					wire := call.ID
					if m.Provider == ProviderXAI && m.API == APIOpenAIResponses {
						var err error
						wire, _, err = xaiResponsesIDPair(call.ID)
						if err != nil {
							return "xAI history has an invalid opaque tool ID"
						}
					}
					if calls[wire] || call.ID == "" {
						return "xAI history repeats a function call association"
					}
					calls[wire] = true
					pending[call.ID] = wire
				}
			}
		case ToolResultMessage:
			if _, ok := pending[m.ToolCallID]; !ok || answered[m.ToolCallID] {
				return "xAI history has an orphan or duplicate tool result"
			}
			answered[m.ToolCallID] = true
		case UserMessage:
			if !closePending() {
				return "xAI history contains a tool call without a result"
			}
			pending = map[string]string{}
			answered = map[string]bool{}
		}
	}
	if !closePending() {
		return "xAI history contains a tool call without a result"
	}
	return ""
}

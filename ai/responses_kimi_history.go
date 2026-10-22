package ai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// Opaque call/item IDs use a versioned JSON pair. The delimiter only marks
// native state for provenance; it never separates untrusted vendor strings.
func kimiResponsesToolID(callID, itemID string) string {
	pair, _ := json.Marshal([2]string{callID, itemID})
	return "kmr1|" + base64.RawURLEncoding.EncodeToString(pair)
}
func kimiResponsesCallID(id string) (string, error) {
	invalid := errors.New("invalid Kimi Responses tool ID")
	encoded, ok := strings.CutPrefix(id, "kmr1|")
	if !ok {
		return "", invalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return "", invalid
	}
	var pair []string
	if json.Unmarshal(raw, &pair) != nil || len(pair) != 2 || pair[0] == "" {
		return "", invalid
	}
	return pair[0], nil
}
func kimiResponsesHistoryToolID(id string, source AssistantMessage) string {
	if source.API == APIOpenAIResponses && isKimi(source.Provider) {
		if _, err := kimiResponsesCallID(id); err == nil {
			return id
		}
	}
	return kimiResponsesToolID(id, "")
}

type kimiResponsesSummaryPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type kimiResponsesInputReasoning struct {
	Type    string                     `json:"type"`
	Summary []kimiResponsesSummaryPart `json:"summary"`
}

// Summary is the documented Kimi output representation. Project input fields
// without rewriting the raw signature or synthesizing content from visible text.
func kimiResponsesReasoning(b Thinking) (kimiResponsesInputReasoning, error) {
	var r struct {
		Type    string          `json:"type"`
		Summary json.RawMessage `json:"summary"`
	}
	invalid := errors.New("Kimi Responses reasoning lacks a valid native summary")
	if b.Redacted || json.Unmarshal([]byte(b.Signature), &r) != nil || r.Type != "reasoning" {
		return kimiResponsesInputReasoning{}, invalid
	}
	var parts []map[string]json.RawMessage
	if json.Unmarshal(r.Summary, &parts) != nil || parts == nil {
		return kimiResponsesInputReasoning{}, invalid
	}
	out := kimiResponsesInputReasoning{Type: "reasoning", Summary: []kimiResponsesSummaryPart{}}
	for _, part := range parts {
		var typ, text string
		if len(part) != 2 || json.Unmarshal(part["type"], &typ) != nil || typ != "summary_text" || json.Unmarshal(part["text"], &text) != nil || string(part["text"]) == "null" {
			return kimiResponsesInputReasoning{}, invalid
		}
		out.Summary = append(out.Summary, kimiResponsesSummaryPart{Type: typ, Text: text})
	}
	return out, nil
}
func kimiResponsesAssistantItems(m replayedAssistant) ([]any, error) {
	var out []any
	for _, block := range m.Content {
		switch b := block.(type) {
		case Thinking:
			if b.Signature == "" {
				if b.Thinking != "" {
					out = append(out, responsesUserMessage{Role: "assistant", Content: []any{responsesInputText{Type: "output_text", Text: b.Thinking}}})
				}
				continue
			}
			item, err := kimiResponsesReasoning(b)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		case Text:
			out = append(out, responsesUserMessage{Role: "assistant", Content: []any{responsesInputText{Type: "output_text", Text: b.Text}}})
		case ToolCall:
			call, err := kimiResponsesCallID(b.ID)
			if err != nil {
				return nil, err
			}
			args := string(b.Arguments)
			if !kimiResponsesArguments(args) {
				return nil, errors.New("Kimi Responses history tool arguments must be a JSON object")
			}
			out = append(out, responsesFunctionCall{Type: "function_call", CallID: call, Name: b.Name, Arguments: args})
		}
	}
	return out, nil
}

package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

// glmResponsesReasoning consumes only trusted same-model state, already
// checked by prepareTranscript. Output/native and input are separate schemas:
// GLM accepts a single reasoning_text object, without output IDs or encrypted
// fields. Decode into owned values; never change the preserved signature.
func glmResponsesReasoning(thought Thinking) (any, error) {
	var item struct {
		Type    string          `json:"type"`
		Content json.RawMessage `json:"content"`
	}
	if json.Unmarshal([]byte(thought.Signature), &item) != nil || item.Type != "reasoning" {
		return nil, errors.New("invalid GLM native reasoning")
	}
	content := bytes.TrimSpace(item.Content)
	var parts []json.RawMessage
	switch {
	case len(content) == 0:
		// A done-only stream can expose visible thinking without native
		// reasoning content. Keep it as ordinary history, not fabricated state.
	case content[0] == '{':
		parts = []json.RawMessage{content}
	case content[0] == '[':
		if json.Unmarshal(content, &parts) != nil {
			return nil, errors.New("invalid GLM native reasoning content")
		}
	default:
		return nil, errors.New("invalid GLM native reasoning content")
	}
	if len(parts) == 0 {
		if thought.Thinking == "" {
			return nil, nil
		}
		return glmResponsesAssistantText(thought.Thinking), nil
	}
	text := make([]string, len(parts))
	for i, raw := range parts {
		var part struct {
			Type string  `json:"type"`
			Text *string `json:"text"`
		}
		if json.Unmarshal(raw, &part) != nil || part.Type != "reasoning_text" || part.Text == nil {
			return nil, errors.New("invalid GLM native reasoning text")
		}
		text[i] = *part.Text
	}
	return struct {
		Type    string             `json:"type"`
		Content responsesInputText `json:"content"`
	}{Type: "reasoning", Content: responsesInputText{Type: "reasoning_text", Text: strings.Join(text, "\n\n")}}, nil
}

// GLM InputMessage has only type, role and content; output item IDs/status and
// OpenAI's generated msg_pi_* IDs do not belong to this independent schema.
type glmResponsesMessage struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content any    `json:"content"`
}

func glmResponsesAssistantText(text string) glmResponsesMessage {
	return glmResponsesMessage{Type: "message", Role: "assistant", Content: []responsesInputText{{Type: "output_text", Text: text}}}
}

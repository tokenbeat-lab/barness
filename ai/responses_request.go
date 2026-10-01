package ai

import (
	"bytes"
	"encoding/json"
)

// Responses wire DTOs. The adapter owns these and marshals them itself, so the
// exact request — field presence included — is decided here rather than by
// the SDK's typed params. The SDK only carries the bytes, authentication and
// SSE decoding. Shapes follow pi-ai 0.87.1 openai-responses.ts buildParams and
// openai-responses-shared.ts convertResponsesMessages.
type responsesBody struct {
	Model  string `json:"model"`
	Input  []any  `json:"input"`
	Stream bool   `json:"stream"`
	Store  bool   `json:"store"`
}

// responsesInstruction is a system or developer instruction item.
type responsesInstruction struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responsesUserMessage struct {
	Role    string               `json:"role"`
	Content []responsesInputText `json:"content"`
}

type responsesInputText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// buildResponsesBody encodes the request for model. Options carry no fields
// yet (ticket 09); with no options pi sends no prompt-cache fields because the
// default cache retention is "short" without a session id.
func buildResponsesBody(model Model, req Request) ([]byte, error) {
	body := responsesBody{Model: model.ID, Input: []any{}, Stream: true, Store: false}

	// pi: reasoning models (that support it) take instructions as "developer".
	role := "system"
	if model.Reasoning {
		role = "developer"
	}
	if req.SystemPrompt != "" {
		body.Input = append(body.Input, responsesInstruction{Role: role, Content: req.SystemPrompt})
	}

	for _, m := range req.Messages {
		switch m := m.(type) {
		case UserMessage:
			content := make([]responsesInputText, 0, len(m.Content))
			for _, c := range m.Content {
				if t, ok := c.(Text); ok {
					content = append(content, responsesInputText{Type: "input_text", Text: t.Text})
				}
			}
			if len(content) == 0 {
				continue
			}
			body.Input = append(body.Input, responsesUserMessage{Role: "user", Content: content})
		}
	}
	return marshalJS(body)
}

// marshalJS encodes like JavaScript's JSON.stringify for these values: no
// HTML escaping and no trailing newline.
func marshalJS(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

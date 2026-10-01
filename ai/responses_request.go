package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"strconv"
	"strings"
	"unicode/utf16"
)

// Responses wire DTOs. The adapter owns these and marshals them itself, so the
// exact request — field presence included — is decided here rather than by
// the SDK's typed params. The SDK only carries the bytes, authentication and
// SSE decoding. Shapes follow pi-ai 0.87.1 openai-responses.ts buildParams and
// openai-responses-shared.ts convertResponsesMessages / convertResponsesTools.
type responsesBody struct {
	Model      string          `json:"model"`
	Input      []any           `json:"input"`
	Stream     bool            `json:"stream"`
	Store      bool            `json:"store"`
	Tools      []responsesTool `json:"tools,omitempty"`
	ToolChoice any             `json:"tool_choice,omitempty"`
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

// responsesOutputMessage replays an assistant text block.
type responsesOutputMessage struct {
	Type    string                `json:"type"`
	Role    string                `json:"role"`
	Content []responsesOutputText `json:"content"`
	Status  string                `json:"status"`
	ID      string                `json:"id"`
	Phase   string                `json:"phase,omitempty"`
}

type responsesOutputText struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
}

// responsesFunctionCall replays a tool call.
type responsesFunctionCall struct {
	Type      string `json:"type"`
	ID        string `json:"id,omitempty"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type responsesFunctionCallOutput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

type responsesTool struct {
	Type        string          `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	// Strict is sent only for models whose compat supports strict mode. It is
	// false: strict JSON-schema sampling is not offered yet.
	Strict *bool `json:"strict,omitempty"`
}

// buildResponsesBody encodes the request for model. Only tool_choice of the
// options is implemented (ticket 09 brings the rest); with no options pi sends
// no prompt-cache fields because the default cache retention is "short"
// without a session id.
func buildResponsesBody(model Model, req Request, opts ResponsesOptions) ([]byte, error) {
	body := responsesBody{Model: model.ID, Input: []any{}, Stream: true, Store: false}

	// pi: reasoning models (that support it) take instructions as "developer".
	role := "system"
	if model.Reasoning {
		role = "developer"
	}
	if req.SystemPrompt != "" {
		body.Input = append(body.Input, responsesInstruction{Role: role, Content: req.SystemPrompt})
	}

	// msgIndex counts the messages that produced input items, as pi's does;
	// it names assistant text replayed without a signature.
	msgIndex := 0
	for _, m := range req.Messages {
		var items []any
		switch m := m.(type) {
		case UserMessage:
			content := make([]responsesInputText, 0, len(m.Content))
			for _, c := range m.Content {
				if t, ok := c.(Text); ok {
					content = append(content, responsesInputText{Type: "input_text", Text: t.Text})
				}
			}
			if len(content) > 0 {
				items = append(items, responsesUserMessage{Role: "user", Content: content})
			}
		case AssistantMessage:
			var err error
			if items, err = responsesAssistantItems(model, m, msgIndex); err != nil {
				return nil, err
			}
		case ToolResultMessage:
			items = append(items, responsesToolResult(m))
		}
		if len(items) == 0 {
			continue
		}
		body.Input = append(body.Input, items...)
		msgIndex++
	}

	for _, t := range req.Tools {
		tool := responsesTool{Type: "function", Name: t.Name, Description: t.Description, Parameters: t.Parameters}
		if model.Compat.SupportsStrictMode {
			tool.Strict = new(bool)
		}
		body.Tools = append(body.Tools, tool)
	}
	if c := opts.ToolChoice; c != nil {
		if c.Function != "" {
			body.ToolChoice = map[string]string{"type": "function", "name": c.Function}
		} else {
			body.ToolChoice = c.Mode
		}
	}
	return marshalJS(body)
}

// responsesAssistantItems replays a previous turn: text as output messages and
// tool calls as function_call items. The generic history transform (tool ID
// normalization across models, synthetic missing results, skipping failed
// turns) is ticket 08's; replaying reasoning state needs ticket 07's trusted
// native-state envelope, so thinking blocks are not sent until then.
func responsesAssistantItems(model Model, m AssistantMessage, msgIndex int) ([]any, error) {
	sameProviderAPI := m.Provider == model.Provider && m.API == model.API
	differentModel := sameProviderAPI && m.Model != model.ID
	var items []any
	textIndex := 0
	for _, block := range m.Content {
		switch b := block.(type) {
		case Text:
			id, phase := parseTextSignature(b.Signature)
			switch {
			case id == "" && textIndex == 0:
				id = "msg_pi_" + strconv.Itoa(msgIndex)
			case id == "":
				id = "msg_pi_" + strconv.Itoa(msgIndex) + "_" + strconv.Itoa(textIndex)
			case len(id) > 64: // OpenAI's item id limit
				id = "msg_" + shortHash(id)
			}
			textIndex++
			items = append(items, responsesOutputMessage{Type: "message", Role: "assistant", Status: "completed", ID: id, Phase: phase,
				Content: []responsesOutputText{{Type: "output_text", Text: b.Text, Annotations: []any{}}}})
		case ToolCall:
			parts := strings.Split(b.ID, "|")
			itemID := ""
			if len(parts) > 1 {
				itemID = parts[1]
			}
			// A different model's fc_ id would trip OpenAI's reasoning pairing
			// check, and function_call item ids must start with fc_.
			if differentModel || !strings.HasPrefix(itemID, "fc_") {
				itemID = ""
			}
			// Arguments that are not JSON (only a host-built call can have
			// them) make the request unencodable: invalid_request, never a
			// silent {}.
			args, err := parseJSON(string(cmp.Or(b.Arguments, "{}")))
			if err != nil {
				return nil, err
			}
			items = append(items, responsesFunctionCall{Type: "function_call", ID: itemID, CallID: parts[0], Name: b.Name, Arguments: stringifyJSON(args)})
		}
	}
	return items, nil
}

// responsesToolResult is a function_call_output: the result's text blocks
// joined by newlines, or pi's placeholder when there are none. Image results
// arrive with ticket 08.
func responsesToolResult(m ToolResultMessage) responsesFunctionCallOutput {
	var texts []string
	for _, c := range m.Content {
		if t, ok := c.(Text); ok {
			texts = append(texts, t.Text)
		}
	}
	output := strings.Join(texts, "\n")
	if output == "" {
		output = "(no tool output)"
	}
	callID, _, _ := strings.Cut(m.ToolCallID, "|")
	return responsesFunctionCallOutput{Type: "function_call_output", CallID: callID, Output: output}
}

// parseTextSignature reads pi-ai's TextSignatureV1 ({"v":1,"id":…,"phase":…});
// any other non-empty signature is a legacy bare item id.
func parseTextSignature(sig string) (id, phase string) {
	if sig == "" {
		return "", ""
	}
	if strings.HasPrefix(sig, "{") {
		var v struct {
			V     any `json:"v"`
			ID    any `json:"id"`
			Phase any `json:"phase"`
		}
		if json.Unmarshal([]byte(sig), &v) == nil && v.V == 1.0 {
			if id, ok := v.ID.(string); ok {
				if p, _ := v.Phase.(string); p == "commentary" || p == "final_answer" {
					return id, p
				}
				return id, ""
			}
		}
	}
	return sig, ""
}

// shortHash ports pi-ai's shortHash, a deterministic cyrb53-style hash over
// UTF-16 code units, used to shorten over-long item ids.
func shortHash(s string) string {
	h1, h2 := uint32(0xdeadbeef), uint32(0x41c6ce57)
	for _, ch := range utf16.Encode([]rune(s)) {
		h1 = (h1 ^ uint32(ch)) * 2654435761
		h2 = (h2 ^ uint32(ch)) * 1597334677
	}
	h1 = ((h1 ^ (h1 >> 16)) * 2246822507) ^ ((h2 ^ (h2 >> 13)) * 3266489909)
	h2 = ((h2 ^ (h2 >> 16)) * 2246822507) ^ ((h1 ^ (h1 >> 13)) * 3266489909)
	return strconv.FormatUint(uint64(h2), 36) + strconv.FormatUint(uint64(h1), 36)
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

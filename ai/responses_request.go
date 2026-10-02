package ai

import (
	"bytes"
	"cmp"
	"encoding/json"
	"slices"
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
	Model                string                        `json:"model"`
	Input                []any                         `json:"input"`
	Stream               bool                          `json:"stream"`
	PromptCacheKey       string                        `json:"prompt_cache_key,omitempty"`
	PromptCacheRetention string                        `json:"prompt_cache_retention,omitempty"`
	PromptCacheOptions   *responsesPromptCacheOptions  `json:"prompt_cache_options,omitempty"`
	Store                *bool                         `json:"store,omitempty"`
	MaxOutputTokens      int                           `json:"max_output_tokens,omitempty"`
	Temperature          Nullable[float64]             `json:"temperature,omitzero"`
	ServiceTier          Nullable[string]              `json:"service_tier,omitzero"`
	Tools                []responsesTool               `json:"tools,omitempty"`
	ToolChoice           Nullable[ResponsesToolChoice] `json:"tool_choice,omitzero"`
	Reasoning            *responsesReasoning           `json:"reasoning,omitempty"`
	Include              []string                      `json:"include,omitempty"`
}

// responsesInstruction is a system or developer instruction item.
type responsesInstruction struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responsesUserMessage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type responsesInputText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesInputImage struct {
	Type     string `json:"type"`
	Detail   string `json:"detail"`
	ImageURL string `json:"image_url"`
}

func newResponsesInputImage(img Image) responsesInputImage {
	return responsesInputImage{Type: "input_image", Detail: "auto", ImageURL: "data:" + img.MimeType + ";base64," + img.Data}
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
	// Output is the result text, or input_text and input_image items when
	// the result holds images the model can see.
	Output any `json:"output"`
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

// buildResponsesBody encodes the prepared history and options for model, as
// pi's buildParams does, samplingParams last; caps leaves out what the
// provider does not serve. cacheKey is the derived prompt cache key, "" for
// none.
func buildResponsesBody(model Model, history transcript, opts ResponsesOptions, caps responsesCapabilities, cacheKey string) ([]byte, error) {
	body := responsesBody{Model: model.ID, Input: []any{}, Stream: true}
	if caps.store {
		body.Store = new(bool)
	}

	// pi: reasoning models (that support it) take instructions as "developer".
	role := "system"
	if model.Reasoning {
		role = "developer"
	}
	vision := model.acceptsImages()

	// msgIndex counts the messages after a leading system message, except
	// user and assistant messages that produced no items, as pi's does; it
	// names assistant text replayed without a signature.
	msgIndex := 0
	for i, m := range history.messages {
		var items []any
		switch m := m.(type) {
		case SystemMessage:
			// A leading message is the complete prompt; a later one (only
			// kept for a model taking them mid-conversation) is an update.
			text := m.updateText()
			if i == 0 {
				text = m.promptText()
			}
			if text != "" {
				body.Input = append(body.Input, responsesInstruction{Role: role, Content: text})
			}
			if i > 0 {
				msgIndex++
			}
			continue
		case UserMessage:
			content := make([]any, 0, len(m.Content))
			for _, c := range m.Content {
				switch c := c.(type) {
				case Text:
					content = append(content, responsesInputText{Type: "input_text", Text: c.Text})
				case Image:
					content = append(content, newResponsesInputImage(c))
				}
			}
			if len(content) > 0 {
				items = append(items, responsesUserMessage{Role: "user", Content: content})
			}
		case replayedAssistant:
			var err error
			if items, err = responsesAssistantItems(model, m, msgIndex); err != nil {
				return nil, err
			}
		case ToolResultMessage:
			items = append(items, responsesToolResult(m, vision))
		}
		if len(items) == 0 {
			continue
		}
		body.Input = append(body.Input, items...)
		msgIndex++
	}

	for _, t := range history.tools {
		tool := responsesTool{Type: "function", Name: t.Name, Description: t.Description, Parameters: t.Parameters}
		if model.Compat.SupportsStrictMode {
			tool.Strict = new(bool)
		}
		body.Tools = append(body.Tools, tool)
	}
	opts.encode(&body, model, caps, cacheKey)
	out, err := marshalJS(body)
	if err != nil {
		return nil, err
	}
	return applySamplingParams(out, opts.SamplingParams)
}

// responsesAssistantItems replays a previous turn prepared by
// prepareTranscript: a signed thinking block as its reasoning item, text as
// output messages and tool calls as function_call items.
func responsesAssistantItems(model Model, m replayedAssistant, msgIndex int) ([]any, error) {
	sameProviderAPI := m.Provider == model.Provider && m.API == model.API
	// A same provider/API message whose native state was downgraded is
	// replayed as pi replays another model's.
	differentModel := sameProviderAPI && !m.sameModel
	var items []any
	textIndex := 0
	for _, block := range m.Content {
		switch b := block.(type) {
		case Thinking:
			if b.Signature == "" {
				continue
			}
			// pi sends JSON.parse(thinkingSignature). Only same-model state
			// under a trusted envelope gets here, so a signature that is not
			// JSON is a host fault: invalid_request, never a silent drop.
			item, err := parseJSON(b.Signature)
			if err != nil {
				return nil, err
			}
			items = append(items, json.RawMessage(stringifyJSON(item)))
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

// responsesToolResult is a function_call_output (pi convertToolResultOutput).
// Without images, or for a model without image input, the output is the text
// blocks joined by newlines, else "(see attached image)" or "(no tool
// output)". With images the model can see, it is the joined text as one
// input_text item, if any, followed by every image.
func responsesToolResult(m ToolResultMessage, vision bool) responsesFunctionCallOutput {
	var texts []string
	var images []Image
	for _, c := range m.Content {
		switch c := c.(type) {
		case Text:
			texts = append(texts, c.Text)
		case Image:
			images = append(images, c)
		}
	}
	text := strings.Join(texts, "\n")
	callID, _, _ := strings.Cut(m.ToolCallID, "|")
	out := responsesFunctionCallOutput{Type: "function_call_output", CallID: callID}
	switch {
	case len(images) > 0 && vision:
		var items []any
		if text != "" {
			items = append(items, responsesInputText{Type: "input_text", Text: text})
		}
		for _, img := range images {
			items = append(items, newResponsesInputImage(img))
		}
		out.Output = items
	case text != "":
		out.Output = text
	case len(images) > 0:
		out.Output = "(see attached image)"
	default:
		out.Output = "(no tool output)"
	}
	return out
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

// responsesPayloadReferences are request fields that point at, or create,
// state stored on the vendor account (an earlier response, a conversation, a
// stored prompt, a background response). The account may be shared by
// several tenants and the library cannot tell whose state a reference
// names, so a payload callback may not add them; barness-ai never sends them.
var responsesPayloadReferences = []string{"previous_response_id", "conversation", "prompt", "background"}

// authorizeResponsesPayload checks a body a payload callback produced. It
// must still name the call's authorized model, keep the tenant-scoped cache
// key and stateless storage, declare only function tools and the hosted tool
// types the binding allows, and carry no vendor-side reference: none of the
// fields above, and no input item_reference or file_id anywhere in the input.
// Other hosted tools (file search, code interpreter, MCP) are refused because
// they name vendor resources or network targets only the binding's allowance
// can vouch for (ADR-0005).
func authorizeResponsesPayload(m Model, cacheKey string, hostedTools []string) func(map[string]any) string {
	return func(body map[string]any) string {
		if model, _ := body["model"].(string); model != m.ID {
			return "payload callback may not change the authorized model"
		}
		key, hasKey := body["prompt_cache_key"]
		if (cacheKey == "" && hasKey) || (cacheKey != "" && key != cacheKey) {
			return "payload callback may not change the prompt cache key"
		}
		if store, ok := body["store"]; ok && store != false {
			return "payload callback may not store the response on the vendor"
		}
		for _, field := range responsesPayloadReferences {
			if _, ok := body[field]; ok {
				return "payload callback may not reference vendor-side state (" + field + ")"
			}
		}
		if tools, ok := body["tools"]; ok {
			list, _ := tools.([]any)
			if list == nil && tools != nil {
				return "payload callback produced malformed tools"
			}
			for _, t := range list {
				obj, _ := t.(map[string]any)
				typ, _ := obj["type"].(string)
				if typ != "function" && (typ == "" || !slices.Contains(hostedTools, typ)) {
					return "payload callback may only declare function tools and the binding's hosted tools"
				}
			}
		}
		if field := vendorReference(body["input"]); field != "" {
			return "payload callback may not reference vendor-side state (" + field + ")"
		}
		return ""
	}
}

// vendorReference finds an item_reference or a file_id anywhere in v.
func vendorReference(v any) string {
	switch v := v.(type) {
	case []any:
		for _, e := range v {
			if field := vendorReference(e); field != "" {
				return field
			}
		}
	case map[string]any:
		if v["type"] == "item_reference" {
			return "item_reference"
		}
		if _, ok := v["file_id"]; ok {
			return "file_id"
		}
		for _, e := range v {
			if field := vendorReference(e); field != "" {
				return field
			}
		}
	}
	return ""
}

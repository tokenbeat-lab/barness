package ai

import (
	"cmp"
	"encoding/json"
	"slices"
	"strings"
	"unicode/utf16"
)

// Chat Completions wire DTOs. The adapter owns these and marshals them
// itself, so the exact request — field presence included — is decided here
// rather than by the SDK's typed params, which carry none of pi's
// non-standard reasoning fields. Shapes follow pi-ai 0.87.1
// openai-completions.ts buildParams and convertMessages for the standard
// OpenAI compat (ADR-0013).
type chatBody struct {
	Model                string            `json:"model"`
	Messages             []any             `json:"messages"`
	Stream               bool              `json:"stream"`
	PromptCacheKey       string            `json:"prompt_cache_key,omitempty"`
	PromptCacheRetention string            `json:"prompt_cache_retention,omitempty"`
	StreamOptions        chatStreamOptions `json:"stream_options"`
	Store                bool              `json:"store"`
	MaxCompletionTokens  int               `json:"max_completion_tokens,omitempty"`
	Temperature          Nullable[float64] `json:"temperature,omitzero"`
	Tools                *[]chatTool       `json:"tools,omitempty"`
	ToolChoice           *ChatToolChoice   `json:"tool_choice,omitempty"`
	ReasoningEffort      string            `json:"reasoning_effort,omitempty"`
}

// chatStreamOptions asks for the usage chunk at the end of the stream.
type chatStreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatTextMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatUserMessage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type chatTextPart struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type chatImagePart struct {
	Type     string       `json:"type"`
	ImageURL chatImageURL `json:"image_url"`
}

type chatImageURL struct {
	URL string `json:"url"`
}

func newChatImagePart(img Image) chatImagePart {
	return chatImagePart{Type: "image_url", ImageURL: chatImageURL{URL: "data:" + img.MimeType + ";base64," + img.Data}}
}

// chatAssistantMessage replays a previous turn. Content is null or the
// turn's text; at most one reasoning field carries the visible reasoning
// back under the name the provider streamed it with.
type chatAssistantMessage struct {
	Role             string          `json:"role"`
	Content          *string         `json:"content"`
	ReasoningContent *string         `json:"reasoning_content,omitempty"`
	Reasoning        *string         `json:"reasoning,omitempty"`
	ReasoningText    *string         `json:"reasoning_text,omitempty"`
	ToolCalls        []chatToolCall  `json:"tool_calls,omitempty"`
	ReasoningDetails json.RawMessage `json:"reasoning_details,omitempty"`
}

type chatToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function chatCallFunction `json:"function"`
}

type chatCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatToolMessage struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	ToolCallID string `json:"tool_call_id"`
}

type chatTool struct {
	Type     string           `json:"type"`
	Function chatToolFunction `json:"function"`
}

type chatToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	// Strict is sent only for models whose compat supports strict mode. It
	// is false: strict JSON-schema sampling is not offered yet.
	Strict *bool `json:"strict,omitempty"`
}

// chatToolImagesIntro introduces the images of a run of tool results, which
// Chat Completions can only carry in a user message.
const chatToolImagesIntro = "Attached image(s) from tool result:"

// buildChatBody encodes the prepared history and options for model, as pi's
// buildParams does, samplingParams last. cacheKey is the derived prompt
// cache key, "" for none.
func buildChatBody(model Model, history transcript, opts ChatOptions, cacheKey string) ([]byte, error) {
	messages, err := chatMessages(model, history)
	if err != nil {
		return nil, err
	}
	body := chatBody{Model: model.ID, Messages: messages, Stream: true,
		StreamOptions: chatStreamOptions{IncludeUsage: true}, Store: false}
	switch {
	case len(history.tools) > 0:
		tools := make([]chatTool, len(history.tools))
		for i, t := range history.tools {
			tools[i] = chatTool{Type: "function", Function: chatToolFunction{Name: t.Name, Description: t.Description, Parameters: t.Parameters}}
			if model.Compat.SupportsStrictMode {
				tools[i].Function.Strict = new(bool)
			}
		}
		body.Tools = &tools
	case history.toolHistory:
		// pi declares an empty tool list for a history with tool calls or
		// results, which some proxies require.
		body.Tools = &[]chatTool{}
	}
	opts.encode(&body, model, cacheKey)
	out, err := marshalJS(body)
	if err != nil {
		return nil, err
	}
	return applySamplingParams(out, opts.SamplingParams)
}

// chatMessages ports pi's convertMessages for the prepared history.
// Instructions go out as "developer" to a reasoning model, else "system".
func chatMessages(model Model, history transcript) ([]any, error) {
	role := "system"
	if model.Reasoning {
		role = "developer"
	}
	vision := model.acceptsImages()
	out := []any{}
	msgs := history.messages
	for i := 0; i < len(msgs); i++ {
		switch m := msgs[i].(type) {
		case SystemMessage:
			// A leading message is the complete prompt; a later one (only
			// kept for a model taking them mid-conversation) is an update.
			text := m.updateText()
			if i == 0 {
				text = m.promptText()
			}
			if text != "" {
				out = append(out, chatTextMessage{Role: role, Content: text})
			}
		case UserMessage:
			content := []any{}
			for _, c := range m.Content {
				switch c := c.(type) {
				case Text:
					if c.Text != "" {
						content = append(content, chatTextPart{Type: "text", Text: c.Text})
					}
				case Image:
					content = append(content, newChatImagePart(c))
				}
			}
			if len(content) > 0 {
				out = append(out, chatUserMessage{Role: "user", Content: content})
			}
		case replayedAssistant:
			msg, ok, err := chatAssistant(m)
			if err != nil {
				return nil, err
			}
			if ok {
				out = append(out, msg)
			}
		case ToolResultMessage:
			// A run of results goes out together: one tool message each,
			// then their images in one user message.
			j := i
			var images []any
			for ; j < len(msgs); j++ {
				r, isResult := msgs[j].(ToolResultMessage)
				if !isResult {
					break
				}
				text, resultImages := chatToolResult(r)
				out = append(out, chatToolMessage{Role: "tool", Content: text, ToolCallID: r.ToolCallID})
				if vision {
					images = append(images, resultImages...)
				}
			}
			i = j - 1
			if len(images) > 0 {
				out = append(out, chatUserMessage{Role: "user", Content: append([]any{chatTextPart{Type: "text", Text: chatToolImagesIntro}}, images...)})
			}
		}
	}
	return out, nil
}

// chatToolResult is a result's text (its text blocks joined by newlines,
// else "(see attached image)" or "(no tool output)") and its images.
func chatToolResult(m ToolResultMessage) (string, []any) {
	var texts []string
	var images []any
	for _, c := range m.Content {
		switch c := c.(type) {
		case Text:
			texts = append(texts, c.Text)
		case Image:
			images = append(images, newChatImagePart(c))
		}
	}
	text := strings.Join(texts, "\n")
	switch {
	case text != "":
		return text, images
	case len(images) > 0:
		return "(see attached image)", images
	}
	return "(no tool output)", images
}

// chatReasoningFields are the delta fields pi reads reasoning from, in the
// order it tries them; a thinking block's signature names the one it came
// from, and replay sends the reasoning back under that name.
var chatReasoningFields = []string{"reasoning_content", "reasoning", "reasoning_text"}

// chatAssistant replays a previous turn prepared by prepareTranscript. Its
// non-blank text becomes the content string; reasoning goes back as the
// provider's reasoning_details when a thinking block (or, in older history,
// a tool call) holds them, else as the reasoning field its thinking came
// from; tool calls as function calls with their parsed arguments. A turn
// with neither content nor tool calls is skipped.
func chatAssistant(m replayedAssistant) (chatAssistantMessage, bool, error) {
	msg := chatAssistantMessage{Role: "assistant"}
	var text strings.Builder
	var thinking []Thinking
	var calls []ToolCall
	for _, block := range m.Content {
		switch b := block.(type) {
		case Text:
			if !isBlank(b.Text) {
				text.WriteString(b.Text)
			}
		case Thinking:
			thinking = append(thinking, b)
		case ToolCall:
			calls = append(calls, b)
		}
	}
	var details json.RawMessage
	for _, t := range thinking {
		if d, ok := chatReasoningDetails(t.Signature); ok {
			details = d
			break
		}
	}
	if details == nil {
		details = chatLegacyReasoningDetails(calls)
	}
	if text.Len() > 0 {
		s := text.String()
		msg.Content = &s
	}
	visible := slices.DeleteFunc(slices.Clone(thinking), func(t Thinking) bool { return isBlank(t.Thinking) })
	if len(visible) > 0 && details == nil {
		texts := make([]string, len(visible))
		for i, t := range visible {
			texts[i] = t.Thinking
		}
		joined := strings.Join(texts, "\n")
		switch visible[0].Signature {
		case "reasoning_content":
			msg.ReasoningContent = &joined
		case "reasoning":
			msg.Reasoning = &joined
		case "reasoning_text":
			msg.ReasoningText = &joined
		}
	}
	for _, c := range calls {
		// Arguments that are not JSON (only a host-built call can have
		// them) make the request unencodable: invalid_request, never a
		// silent {}.
		args, err := parseJSON(string(cmp.Or(c.Arguments, "{}")))
		if err != nil {
			return msg, false, err
		}
		msg.ToolCalls = append(msg.ToolCalls, chatToolCall{ID: c.ID, Type: "function",
			Function: chatCallFunction{Name: c.Name, Arguments: stringifyJSON(args)}})
	}
	msg.ReasoningDetails = details
	return msg, msg.Content != nil || len(msg.ToolCalls) > 0, nil
}

// chatToolCallID ports pi's Chat normalizeToolCallId for a call replayed by
// the cross-model rules. A Responses id ("<call_id>|<item id>") becomes both
// parts in OpenAI's alphabet joined by "_", shortened with a hash when over
// 40 characters, so calls sharing a call id stay distinct; any other id is
// cut to 40 characters for OpenAI and kept for other providers.
func chatToolCallID(target Model, id string) string {
	callID, itemID, piped := strings.Cut(id, "|")
	if !piped {
		if target.Provider == ProviderOpenAI {
			return truncateUTF16Units(id, 40)
		}
		return id
	}
	callID, itemID = chatIDPart(callID), chatIDPart(itemID)
	combined := callID
	if itemID != "" {
		combined += "_" + itemID
	}
	if len(combined) <= 40 {
		return combined
	}
	hash := shortHash(id)[:8]
	return callID[:min(len(callID), max(1, 40-len(hash)-1))] + "_" + hash
}

// chatIDPart replaces every UTF-16 unit outside [A-Za-z0-9_-] with "_", as
// pi's regular expression does.
func chatIDPart(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		case utf16.RuneLen(r) == 2:
			b.WriteString("__")
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// truncateUTF16Units is JavaScript's s.slice(0, n). A cut through a
// surrogate pair keeps the whole character, which Go cannot split. Unlike
// truncateUTF16 (pi's error text truncation) it adds no marker: an id is
// cut silently.
func truncateUTF16Units(s string, n int) string {
	units := 0
	for i, r := range s {
		if units >= n {
			return s[:i]
		}
		units += utf16.RuneLen(r)
	}
	return s
}

func (chatAdapter) historyRules(m Model) historyRules {
	return historyRules{
		midConvoSystem: m.Compat.SupportsMidConvoSystemMessages,
		normalizeToolCallID: func(id string, _ AssistantMessage) string {
			return chatToolCallID(m, id)
		},
	}
}

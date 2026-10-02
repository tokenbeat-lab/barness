package ai

import (
	"cmp"
	"encoding/json"
	"slices"
	"strings"
)

// Anthropic Messages wire DTOs. The adapter owns these and marshals them
// itself, so the exact request — field presence included — is decided here
// rather than by the SDK's typed params. Shapes follow pi-ai 0.87.1
// anthropic-messages.ts buildParams, convertMessages, convertContentBlocks
// and convertTools for an API key (pi's OAuth, Copilot, managed-effort,
// fallback and native tool-change paths are out of scope).
type anthropicBody struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream"`
	// Betas is pi's request parameter the SDK turns into the anthropic-beta
	// header. It is only in the body a payload callback sees; the adapter
	// moves it to the header before sending.
	Betas        []string                   `json:"betas,omitempty"`
	System       []anthropicText            `json:"system,omitempty"`
	Temperature  Nullable[float64]          `json:"temperature,omitzero"`
	Tools        []anthropicTool            `json:"tools,omitempty"`
	Thinking     *anthropicThinking         `json:"thinking,omitempty"`
	OutputConfig *anthropicOutputConfig     `json:"output_config,omitempty"`
	Metadata     *anthropicMetadata         `json:"metadata,omitempty"`
	ToolChoice   *anthropicToolChoiceObject `json:"tool_choice,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type anthropicCacheControl struct {
	Type string `json:"type"`
	TTL  string `json:"ttl,omitempty"`
}

type anthropicText struct {
	Type         string                 `json:"type"`
	Text         string                 `json:"text"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicImage struct {
	Type         string                 `json:"type"`
	Source       anthropicImageSource   `json:"source"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

func newAnthropicImage(img Image) anthropicImage {
	return anthropicImage{Type: "image", Source: anthropicImageSource{Type: "base64", MediaType: img.MimeType, Data: img.Data}}
}

type anthropicToolResult struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	// Content is the result text, or text and image blocks.
	Content      any                    `json:"content"`
	IsError      bool                   `json:"is_error"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicThinkingBlock struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

type anthropicRedactedThinking struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type anthropicToolUse struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type anthropicInputSchema struct {
	Type       string          `json:"type"`
	Properties json.RawMessage `json:"properties"`
	Required   json.RawMessage `json:"required"`
}

type anthropicTool struct {
	Name                string                 `json:"name"`
	Description         string                 `json:"description"`
	EagerInputStreaming bool                   `json:"eager_input_streaming"`
	InputSchema         anthropicInputSchema   `json:"input_schema"`
	CacheControl        *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicThinking struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens,omitempty"`
	Display      string `json:"display,omitempty"`
}

type anthropicOutputConfig struct {
	Effort AnthropicEffort `json:"effort"`
}

type anthropicMetadata struct {
	UserID string `json:"user_id"`
}

// anthropicInterleavedThinkingBeta is pi's INTERLEAVED_THINKING_BETA.
const anthropicInterleavedThinkingBeta = "interleaved-thinking-2025-05-14"

// anthropicBetas ports pi's getBetaFeatures for the paths barness-ai
// serves: only a budget-based reasoning model with thinking enabled asks for
// interleaved thinking. (Fine-grained tool streaming is for models without
// eager input streaming, which every Anthropic model supports in pi's data.)
// A beta the trusted header transform configures replaces this list, as pi's
// configured anthropic-beta header does.
func anthropicBetas(m Model, o AnthropicOptions) []string {
	enabled, _ := o.ThinkingEnabled.Get()
	interleaved, set := o.InterleavedThinking.Get()
	if m.Reasoning && enabled && (interleaved || !set) && !m.Compat.ForceAdaptiveThinking {
		return []string{anthropicInterleavedThinkingBeta}
	}
	return nil
}

// normalizeBetas is pi's reading of a configured anthropic-beta value: the
// comma-separated features, trimmed, without empty or repeated ones.
func normalizeBetas(values []string) []string {
	var out []string
	for _, f := range strings.Split(strings.Join(values, ","), ",") {
		if f = strings.TrimSpace(f); f != "" && !slices.Contains(out, f) {
			out = append(out, f)
		}
	}
	return out
}

// cacheControl is pi's getCacheControl: nil for retention none, else an
// ephemeral marker with a one-hour ttl for long retention on a model that
// supports it. pi's PI_CACHE_RETENTION environment fallback is never read.
func (o AnthropicOptions) cacheControl(m Model) *anthropicCacheControl {
	retention := cmp.Or(o.CacheRetention, CacheRetentionShort)
	if retention == CacheRetentionNone {
		return nil
	}
	cc := &anthropicCacheControl{Type: "ephemeral"}
	long := true
	if v, ok := m.Compat.SupportsLongCacheRetention.Get(); ok {
		long = v
	}
	if retention == CacheRetentionLong && long {
		cc.TTL = "1h"
	}
	return cc
}

// buildAnthropicBody encodes the prepared history and options for model, as
// pi's buildParams does. betas are included for a payload callback to see.
func buildAnthropicBody(model Model, history transcript, opts AnthropicOptions, betas []string) ([]byte, error) {
	cc := opts.cacheControl(model)
	body := anthropicBody{Model: model.ID, MaxTokens: model.MaxTokens, Stream: true, Betas: betas}
	if n, ok := opts.MaxTokens.Get(); ok {
		body.MaxTokens = n
	}

	msgs := history.messages
	if len(msgs) > 0 {
		if sm, ok := msgs[0].(SystemMessage); ok {
			if text := sm.promptText(); text != "" {
				body.System = []anthropicText{{Type: "text", Text: text, CacheControl: cc}}
			}
			msgs = msgs[1:]
		}
	}
	var err error
	if body.Messages, err = anthropicMessages(msgs, cc); err != nil {
		return nil, err
	}

	enabled, _ := opts.ThinkingEnabled.Get()
	supportsTemperature := true
	if v, ok := model.Compat.SupportsTemperature.Get(); ok {
		supportsTemperature = v
	}
	if !enabled && supportsTemperature {
		body.Temperature = opts.Temperature
	}

	for i, t := range history.tools {
		tool, err := newAnthropicTool(t)
		if err != nil {
			return nil, err
		}
		if i == len(history.tools)-1 {
			tool.CacheControl = cc
		}
		body.Tools = append(body.Tools, tool)
	}

	if model.Reasoning {
		display := cmp.Or(opts.ThinkingDisplay, "summarized")
		switch {
		case enabled && model.Compat.ForceAdaptiveThinking:
			body.Thinking = &anthropicThinking{Type: "adaptive", Display: display}
			if opts.Effort != "" {
				body.OutputConfig = &anthropicOutputConfig{Effort: opts.Effort}
			}
		case enabled:
			body.Thinking = &anthropicThinking{Type: "enabled", BudgetTokens: cmp.Or(opts.ThinkingBudgetTokens, 1024), Display: display}
		case opts.ThinkingEnabled.IsZero() || opts.ThinkingEnabled.IsNull():
		case !model.ThinkingLevelMap.Off.IsNull():
			body.Thinking = &anthropicThinking{Type: "disabled"}
		}
	}

	var userID string
	if raw, ok := opts.Metadata["user_id"]; ok && json.Unmarshal(raw, &userID) == nil {
		body.Metadata = &anthropicMetadata{UserID: userID}
	}
	if c, ok := opts.ToolChoice.Get(); ok {
		wire := c.wire()
		body.ToolChoice = &wire
	}
	return marshalJS(body)
}

// anthropicMessages is pi's convertMessages for the conversation after the
// leading system message. A later system message (kept only for a model
// taking them mid-conversation) waits until the next assistant turn, since a
// tool_result must directly follow its tool_use. The last user or system
// message's last block carries the cache marker.
func anthropicMessages(msgs []Message, cc *anthropicCacheControl) ([]anthropicMessage, error) {
	out := []anthropicMessage{}
	var pending []anthropicMessage
	flush := func() {
		out = append(out, pending...)
		pending = nil
	}
	for i := 0; i < len(msgs); i++ {
		switch m := msgs[i].(type) {
		case SystemMessage:
			if text := m.updateText(); text != "" {
				pending = append(pending, anthropicMessage{Role: "system", Content: []any{anthropicText{Type: "text", Text: text}}})
			}
		case UserMessage:
			var blocks []any
			for _, c := range m.Content {
				switch c := c.(type) {
				case Text:
					if !isBlank(c.Text) {
						blocks = append(blocks, anthropicText{Type: "text", Text: c.Text})
					}
				case Image:
					blocks = append(blocks, newAnthropicImage(c))
				}
			}
			if len(blocks) > 0 {
				out = append(out, anthropicMessage{Role: "user", Content: blocks})
			}
		case replayedAssistant:
			flush()
			blocks, err := anthropicAssistantBlocks(m)
			if err != nil {
				return nil, err
			}
			if len(blocks) > 0 {
				out = append(out, anthropicMessage{Role: "assistant", Content: blocks})
			}
		case ToolResultMessage:
			// Consecutive results travel in one user message.
			var results []any
			for ; i < len(msgs); i++ {
				r, ok := msgs[i].(ToolResultMessage)
				if !ok {
					break
				}
				results = append(results, anthropicToolResultBlock(r))
			}
			i--
			out = append(out, anthropicMessage{Role: "user", Content: results})
		}
	}
	flush()
	if cc != nil && len(out) > 0 {
		last := &out[len(out)-1]
		if (last.Role == "user" || last.Role == "system") && len(last.Content) > 0 {
			last.Content[len(last.Content)-1] = withCacheControl(last.Content[len(last.Content)-1], cc)
		}
	}
	return out, nil
}

// withCacheControl marks a block that can carry a cache breakpoint.
func withCacheControl(block any, cc *anthropicCacheControl) any {
	switch b := block.(type) {
	case anthropicText:
		b.CacheControl = cc
		return b
	case anthropicImage:
		b.CacheControl = cc
		return b
	case anthropicToolResult:
		b.CacheControl = cc
		return b
	}
	return block
}

// anthropicAssistantBlocks replays a previous turn prepared by
// prepareTranscript: blank text is skipped; redacted thinking returns its
// payload; signed thinking is replayed with its signature even when its text
// is blank; unsigned thinking (only same-model state reaches here as
// thinking) becomes text; tool calls become tool_use blocks.
func anthropicAssistantBlocks(m replayedAssistant) ([]any, error) {
	var blocks []any
	for _, block := range m.Content {
		switch b := block.(type) {
		case Text:
			if !isBlank(b.Text) {
				blocks = append(blocks, anthropicText{Type: "text", Text: b.Text})
			}
		case Thinking:
			signed := !isBlank(b.Signature)
			switch {
			case b.Redacted:
				blocks = append(blocks, anthropicRedactedThinking{Type: "redacted_thinking", Data: b.Signature})
			case isBlank(b.Thinking) && !signed:
			case !signed:
				blocks = append(blocks, anthropicText{Type: "text", Text: b.Thinking})
			default:
				blocks = append(blocks, anthropicThinkingBlock{Type: "thinking", Thinking: b.Thinking, Signature: b.Signature})
			}
		case ToolCall:
			// Arguments that are not JSON (only a host-built call can have
			// them) make the request unencodable: invalid_request, never a
			// silent {}.
			args, err := parseJSON(string(cmp.Or(b.Arguments, "{}")))
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, anthropicToolUse{Type: "tool_use", ID: b.ID, Name: b.Name, Input: json.RawMessage(stringifyJSON(args))})
		}
	}
	return blocks, nil
}

// anthropicToolResultBlock is pi's convertToolResult: text alone is joined by
// newlines into a string; with images the blocks are kept, led by "(see
// attached image)" when there is no text. Images a model cannot see were
// already replaced by placeholder text.
func anthropicToolResultBlock(m ToolResultMessage) anthropicToolResult {
	r := anthropicToolResult{Type: "tool_result", ToolUseID: m.ToolCallID, IsError: m.IsError}
	var texts []string
	var blocks []any
	hasImage, hasText := false, false
	for _, c := range m.Content {
		switch c := c.(type) {
		case Text:
			texts = append(texts, c.Text)
			blocks = append(blocks, anthropicText{Type: "text", Text: c.Text})
			hasText = true
		case Image:
			blocks = append(blocks, newAnthropicImage(c))
			hasImage = true
		}
	}
	switch {
	case !hasImage:
		r.Content = strings.Join(texts, "\n")
	case !hasText:
		r.Content = append([]any{anthropicText{Type: "text", Text: "(see attached image)"}}, blocks...)
	default:
		r.Content = blocks
	}
	return r
}

// newAnthropicTool is pi's convertTools without constrained sampling: the
// schema is reduced to an object of its properties and required list.
func newAnthropicTool(t Tool) (anthropicTool, error) {
	// Parameters were checked to be a JSON object when the call was received.
	var schema map[string]json.RawMessage
	if err := json.Unmarshal(t.Parameters, &schema); err != nil {
		return anthropicTool{}, err
	}
	orDefault := func(raw json.RawMessage, def string) json.RawMessage {
		if len(raw) == 0 || string(raw) == "null" {
			return json.RawMessage(def)
		}
		return raw
	}
	return anthropicTool{Name: t.Name, Description: t.Description, EagerInputStreaming: true,
		InputSchema: anthropicInputSchema{Type: "object", Properties: orDefault(schema["properties"], "{}"), Required: orDefault(schema["required"], "[]")}}, nil
}

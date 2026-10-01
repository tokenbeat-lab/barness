package ai

import "unicode/utf16"

// The simple entry's output budget (pi-ai simple-options.ts and
// utils/estimate.ts). The estimate only sizes max output tokens; it never
// trims, drops or reorders history (spec I7).
const (
	// contextSafetyTokens are always kept free of the context window.
	contextSafetyTokens = 4096
	// estimatedImageChars is what pi counts for one image.
	estimatedImageChars = 4800
	charsPerToken       = 4
)

// clampMaxTokensToContext ports pi's clampMaxTokensToContext: at most what
// the context window leaves after the estimated input and the safety margin,
// and at least 1. A model without a positive context window is not clamped
// beyond that minimum.
func clampMaxTokensToContext(m Model, msgs []Message, maxTokens int) int {
	if m.ContextWindow <= 0 {
		return max(1, maxTokens)
	}
	available := m.ContextWindow - estimateContextTokens(msgs) - contextSafetyTokens
	return min(maxTokens, max(1, available))
}

// estimateContextTokens ports pi's estimateContextTokens over the normalized
// transcript (SystemPrompt and Tools folded into a leading system message,
// nothing else converted). The usage the last successful assistant turn
// reported stands for everything up to it; later messages are estimated at
// four characters per token.
//
// pi skips a turn's usage when a message before it has a later timestamp
// (for example, an inserted compaction summary). Only assistant messages
// carry timestamps here; every other message counts as 0, as it does in
// the pi differential.
func estimateContextTokens(msgs []Message) int {
	usageTokens, usageIndex := 0, -1
	latest := int64(0)
	for i, m := range msgs {
		a, ok := m.(AssistantMessage)
		if !ok {
			continue
		}
		failed := a.StopReason == StopReasonError || a.StopReason == StopReasonAborted
		if tokens := contextTokens(a.Usage); a.Timestamp >= latest && !failed && tokens > 0 {
			usageTokens, usageIndex = tokens, i
		}
		latest = max(latest, a.Timestamp)
	}
	tokens := usageTokens
	for _, m := range msgs[usageIndex+1:] {
		tokens += estimateMessageTokens(m)
	}
	return tokens
}

// contextTokens is pi's calculateContextTokens.
func contextTokens(u Usage) int {
	if u.TotalTokens != 0 {
		return int(u.TotalTokens)
	}
	return int(u.Input + u.Output + u.CacheRead + u.CacheWrite)
}

func estimateMessageTokens(m Message) int {
	chars := 0
	switch m := m.(type) {
	case SystemMessage:
		return ceilTokens(jsLength(m.promptText())) + estimateToolsTokens(m.ToolsAdded) + estimateRemovedToolsTokens(m.ToolsRemoved)
	case UserMessage:
		for _, c := range m.Content {
			chars += contentChars(c)
		}
	case ToolResultMessage:
		for _, c := range m.Content {
			chars += contentChars(c)
		}
	case AssistantMessage:
		for _, c := range m.Content {
			switch c := c.(type) {
			case Text:
				chars += jsLength(c.Text)
			case Thinking:
				chars += jsLength(c.Thinking)
			case ToolCall:
				args, _ := c.Arguments.MarshalJSON()
				chars += jsLength(c.Name) + jsLength(string(args))
			}
		}
	}
	return ceilTokens(chars)
}

// contentChars counts a user or tool result block: its text, or a fixed
// amount for an image.
func contentChars(c any) int {
	if t, ok := c.(Text); ok {
		return jsLength(t.Text)
	}
	return estimatedImageChars
}

// estimateToolsTokens sizes declared tools as pi's JSON.stringify of
// {name, description, parameters}.
func estimateToolsTokens(tools []Tool) int {
	if len(tools) == 0 {
		return 0
	}
	list := make([]any, len(tools))
	for i, t := range tools {
		o := newJSONObject()
		o.set("name", t.Name)
		o.set("description", t.Description)
		// Parameters were checked to be a JSON object when the call was
		// received.
		params, _ := parseJSON(string(t.Parameters))
		o.set("parameters", params)
		list[i] = o
	}
	return ceilTokens(jsLength(stringifyJSON(list)))
}

// estimateRemovedToolsTokens sizes removed tools as pi's [{name}] references.
func estimateRemovedToolsTokens(names []string) int {
	if len(names) == 0 {
		return 0
	}
	list := make([]any, len(names))
	for i, n := range names {
		o := newJSONObject()
		o.set("name", n)
		list[i] = o
	}
	return ceilTokens(jsLength(stringifyJSON(list)))
}

func ceilTokens(chars int) int { return (chars + charsPerToken - 1) / charsPerToken }

// jsLength is JavaScript's String#length: UTF-16 code units.
func jsLength(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

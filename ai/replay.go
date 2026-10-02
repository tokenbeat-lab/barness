package ai

import (
	"strings"
	"unicode/utf16"
)

// replayedAssistant is a history assistant message prepared for the call's
// target: its content already follows pi's same-model or cross-model rules,
// and sameModel tells the adapter which ones applied. Adapters only ever see
// history assistant messages in this form.
type replayedAssistant struct {
	AssistantMessage
	sameModel bool
	// downgrade is why native state was not replayed natively, or
	// noDowngrade. It is counted only if the message is replayed at all.
	downgrade downgradeReason
}

// replayAssistant applies the per-message half of pi-ai's transformMessages
// to one history assistant message, with one barness-ai addition: a message
// carrying native state counts as same-model only when nativeReplay vouches
// for it, and is otherwise downgraded exactly as pi downgrades another
// model's history (spec I6). A message without native state has nothing to
// vouch for and follows pi's field comparison alone. renamed records each
// tool call id the target's rules rewrote, for the results that answer it.
func (o replayOrigin) replayAssistant(a AssistantMessage, rules historyRules, renamed map[string]string) replayedAssistant {
	same := o.isTarget(a.Provider, a.API, a.Model)
	reason := noDowngrade
	if carriesNativeState(a) {
		reason = o.nativeReplay(a)
		same = reason == noDowngrade
	}
	source := a
	a.Content = replayContent(a.Content, same)
	if !same && rules.normalizeToolCallID != nil {
		for i, block := range a.Content {
			call, ok := block.(ToolCall)
			if !ok {
				continue
			}
			if id := rules.normalizeToolCallID(call.ID, source); id != call.ID {
				renamed[call.ID] = id
				call.ID = id
				a.Content[i] = call
			}
		}
	}
	return replayedAssistant{AssistantMessage: a, sameModel: same, downgrade: reason}
}

// replayContent converts one message's blocks into a new slice: same-model
// keeps redacted and signed thinking (even with no visible text) and drops
// other blank thinking; cross-model turns visible thinking into plain text,
// drops redacted and blank thinking, strips text of its signature and
// removes a tool call's non-empty thought signature. As pi tests the
// signature for truthiness, a null or "" one is kept as it is.
func replayContent(content []AssistantContent, sameModel bool) []AssistantContent {
	out := make([]AssistantContent, 0, len(content))
	for _, block := range content {
		switch b := block.(type) {
		case Thinking:
			switch {
			case b.Redacted:
				if sameModel {
					out = append(out, b)
				}
			case sameModel && b.Signature != "":
				out = append(out, b)
			case isBlank(b.Thinking):
			case sameModel:
				out = append(out, b)
			default:
				out = append(out, Text{Text: b.Thinking})
			}
		case Text:
			if !sameModel {
				b = Text{Text: b.Text}
			}
			out = append(out, b)
		case ToolCall:
			if sig, _ := b.ThoughtSignature.Get(); !sameModel && sig != "" {
				b.ThoughtSignature = Nullable[string]{}
			}
			out = append(out, b)
		default:
			out = append(out, block)
		}
	}
	return out
}

// carriesNativeState reports whether m holds provider state bound to the
// account and model that produced it: a reasoning signature or redacted
// block, a text signature or item id, or a tool call's provider item id or
// thought signature.
func carriesNativeState(m AssistantMessage) bool {
	for _, block := range m.Content {
		switch b := block.(type) {
		case Thinking:
			if b.Signature != "" || b.Redacted {
				return true
			}
		case Text:
			if b.Signature != "" {
				return true
			}
		case ToolCall:
			if _, item, _ := strings.Cut(b.ID, "|"); item != "" {
				return true
			}
			if sig, _ := b.ThoughtSignature.Get(); sig != "" {
				return true
			}
		}
	}
	return false
}

// isBlank is JavaScript's `s.trim().length === 0`.
func isBlank(s string) bool { return strings.TrimFunc(s, isJSSpace) == "" }

// sanitizeToolCallID is pi-ai's Anthropic normalizeToolCallId for a call
// replayed by the cross-model rules, which Responses also builds on: every character outside [A-Za-z0-9_-]
// becomes "_" and at most 64 are kept. Like pi's regular expression it works
// on UTF-16 code units, so a character outside the Basic Multilingual Plane
// becomes two underscores.
func sanitizeToolCallID(id string) string {
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-':
			b.WriteRune(r)
		case utf16.RuneLen(r) == 2:
			b.WriteString("__")
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	return s[:min(len(s), 64)]
}

package ai

import "strings"

// replayedAssistant is a history assistant message prepared for the call's
// target: its content already follows pi's same-model or cross-model rules,
// and sameModel tells the adapter which ones applied. Adapters only ever see
// history assistant messages in this form.
type replayedAssistant struct {
	AssistantMessage
	sameModel bool
}

// replayHistory prepares the host's history for the target. It ports the
// thinking and text rules of pi-ai's transformMessages, with one barness-ai
// addition: a message carrying native state counts as same-model only when
// nativeReplay vouches for it, and is otherwise downgraded exactly as pi
// downgrades another model's history (spec I6). A message without native
// state has nothing to vouch for and follows pi's field comparison alone.
// The caller's messages are never modified.
//
// The rest of pi's history transform — tool call ID normalization, synthetic
// results for unanswered calls, skipping error/aborted turns, image
// placeholders — is ticket 08's. Known defect until then: an error/aborted
// turn, which pi skips, is replayed and may add to the downgrade counts.
func (o replayOrigin) replayHistory(msgs []Message) ([]Message, NativeStateDowngrades) {
	var count NativeStateDowngrades
	out := make([]Message, len(msgs))
	for i, m := range msgs {
		a, ok := m.(AssistantMessage)
		if !ok {
			out[i] = m
			continue
		}
		same := o.isTarget(a.Provider, a.API, a.Model)
		if carriesNativeState(a) {
			same = o.nativeReplay(a, &count)
		}
		a.Content = replayContent(a.Content, same)
		out[i] = replayedAssistant{AssistantMessage: a, sameModel: same}
	}
	return out, count
}

// replayContent converts one message's blocks into a new slice: same-model
// keeps redacted and signed thinking (even with no visible text) and drops
// other blank thinking; cross-model turns visible thinking into plain text,
// drops redacted and blank thinking, and strips text of its signature.
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
			case strings.TrimFunc(b.Thinking, isJSSpace) == "":
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
		default:
			out = append(out, block)
		}
	}
	return out
}

// carriesNativeState reports whether m holds provider state bound to the
// account and model that produced it: a reasoning signature or redacted
// block, a text item id, or a tool call's provider item id.
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
		}
	}
	return false
}

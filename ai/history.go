package ai

// historyRules are the target-specific inputs of pi-ai's history transform,
// supplied by the adapter for the call's model.
type historyRules struct {
	// midConvoSystem keeps later system messages in place (pi's
	// supportsMidConvoSystemMessages); otherwise they all fold into one
	// leading system message.
	midConvoSystem bool
	// normalizeToolCallID rewrites a tool call id of a message replayed by
	// the cross-model rules into one the target accepts; source is that
	// message. nil keeps ids unchanged.
	normalizeToolCallID func(id string, source AssistantMessage) string
}

// transcript is a call's history prepared for its target. messages holds
// SystemMessage, UserMessage, replayedAssistant and ToolResultMessage values,
// all new or unmodified copies: the caller's history is never written. tools
// is the tool set current after the last system message.
type transcript struct {
	messages []Message
	tools    []Tool
}

// The placeholders pi-ai puts in place of images for a model without image
// input.
const (
	userImagePlaceholder = "(image omitted: model does not support images)"
	toolImagePlaceholder = "(tool image omitted: model does not support images)"
)

// prepareTranscript ports pi-ai's history normalization for the target:
// normalizeContext folds SystemPrompt and Tools into a leading system
// message, resolveTranscript collapses later ones unless the target takes
// them mid-conversation, and transformMessages downgrades images the model
// cannot see, applies the same-/cross-model content rules with the target's
// tool call ids, skips error and aborted turns and answers unanswered tool
// calls. It returns the native state downgrades of the assistant messages
// actually replayed.
func (o replayOrigin) prepareTranscript(req Request, rules historyRules) (transcript, NativeStateDowngrades) {
	msgs := normalizeTranscript(req)
	if !rules.midConvoSystem {
		msgs = collapseSystemMessages(msgs)
	}
	tools := currentTools(msgs)

	vision := o.model.acceptsImages()
	renamed := map[string]string{}
	converted := make([]Message, len(msgs))
	for i, m := range msgs {
		switch m := m.(type) {
		case UserMessage:
			if !vision {
				m.Content = placeholderImages(m.Content, userImagePlaceholder, func(s string) UserContent { return Text{Text: s} })
			}
			converted[i] = m
		case ToolResultMessage:
			if !vision {
				m.Content = placeholderImages(m.Content, toolImagePlaceholder, func(s string) ToolResultContent { return Text{Text: s} })
			}
			if id, ok := renamed[m.ToolCallID]; ok {
				m.ToolCallID = id
			}
			converted[i] = m
		case AssistantMessage:
			converted[i] = o.replayAssistant(m, rules, renamed)
		default:
			converted[i] = m
		}
	}
	out, downgrades := settleTurns(converted)
	return transcript{messages: out, tools: tools}, downgrades
}

// settleTurns is the second pass of pi's transformMessages. It drops
// error and aborted assistant turns, which are incomplete and must not be
// replayed, and gives every tool call left without a result a synthetic
// error result ("No result provided"), placed after the results that did
// arrive and before the next user or assistant message. A system message
// arriving while calls are open is held back until they are answered, so it
// never separates a call from its results. It counts the native state
// downgrades of the assistant messages it keeps.
func settleTurns(msgs []Message) ([]Message, NativeStateDowngrades) {
	var downgrades NativeStateDowngrades
	out := make([]Message, 0, len(msgs))
	var pending []ToolCall
	answered := map[string]bool{}
	var held []Message
	closePending := func() {
		for _, call := range pending {
			if !answered[call.ID] {
				out = append(out, ToolResultMessage{ToolCallID: call.ID, ToolName: call.Name,
					Content: []ToolResultContent{Text{Text: "No result provided"}}, IsError: true})
			}
		}
		pending, answered = nil, map[string]bool{}
		out = append(out, held...)
		held = nil
	}
	for _, m := range msgs {
		switch m := m.(type) {
		case replayedAssistant:
			closePending()
			if m.StopReason == StopReasonError || m.StopReason == StopReasonAborted {
				continue
			}
			downgrades.add(m.downgrade)
			for _, block := range m.Content {
				if call, ok := block.(ToolCall); ok {
					pending = append(pending, call)
				}
			}
			out = append(out, m)
		case ToolResultMessage:
			answered[m.ToolCallID] = true
			out = append(out, m)
		case SystemMessage:
			if len(pending) > 0 {
				held = append(held, m)
			} else {
				out = append(out, m)
			}
		default: // UserMessage: a new user turn interrupts the tool flow
			closePending()
			out = append(out, m)
		}
	}
	closePending()
	return out, downgrades
}

// placeholderImages replaces each image block with placeholder text, keeping
// one placeholder for a run of images, including a run that follows text
// already equal to the placeholder (pi replaceImagesWithPlaceholder). It
// returns a new slice.
func placeholderImages[C any](content []C, placeholder string, text func(string) C) []C {
	out := make([]C, 0, len(content))
	previousWasPlaceholder := false
	for _, block := range content {
		switch b := any(block).(type) {
		case Image:
			if !previousWasPlaceholder {
				out = append(out, text(placeholder))
			}
			previousWasPlaceholder = true
			continue
		case Text:
			previousWasPlaceholder = b.Text == placeholder
		default:
			previousWasPlaceholder = false
		}
		out = append(out, block)
	}
	return out
}

package ai

import (
	"cmp"
	"slices"
)

// assembler is the single production and merge process behind every entry
// point. Adapters report normalized protocol progress to it; it maintains the
// assistant message in the call's PartialView and publishes events. Stream
// gives it a sink; Complete gives it none, so no event is ever queued for a
// caller that will not read it.
//
// It also holds the call's resource-limit failure: the first limit reached
// while merging (a tool call's argument JSON) or publishing (the Stream's
// event queue) is latched, publishing stops, and the adapter must end the
// call at its next check of failure. The message keeps everything received
// up to that point, including the content whose event did not fit.
type assembler struct {
	view *PartialView
	// call is the attribution every published event is enveloped with; it
	// is replaced, never changed, once the call resolved.
	call CallAttribution
	// emit queues a non-terminal or terminal event for the Stream's consumer
	// and reports a non-terminal one that did not fit; nil for Complete.
	emit     func(EventEnvelope) *Error
	toolJSON int64 // the policy's MaxToolJSONBytes
	limited  *Error
}

func newAssembler(timestamp int64, call CallAttribution, emit func(EventEnvelope) *Error, toolJSON int64) *assembler {
	return &assembler{
		view:     newPartialView(AssistantMessage{StopReason: StopReasonPending, Timestamp: timestamp}),
		call:     call,
		emit:     emit,
		toolJSON: toolJSON,
	}
}

func (a *assembler) publish(e Event) {
	if a.emit == nil || a.limited != nil {
		return
	}
	a.fail(a.emit(EventEnvelope{Call: a.call, Event: e}))
}

// fail latches the first resource-limit failure.
func (a *assembler) fail(f *Error) {
	if a.limited == nil {
		a.limited = f
	}
}

// failure is the latched resource-limit failure, or nil. An adapter checks it
// after each unit of protocol progress and ends the call with it.
func (a *assembler) failure() *Error { return a.limited }

// fitsToolJSON reports whether a tool call's argument JSON of size bytes is
// within the policy, latching the failure when it is not. Adapters ask before
// they grow their own copy of the text.
func (a *assembler) fitsToolJSON(size int) bool {
	if int64(size) <= a.toolJSON {
		return true
	}
	a.fail(limitFailure(PhaseStream, "MaxToolJSONBytes", a.toolJSON))
	return false
}

// identify records what actually serves the call, once resolved: later
// events carry call, and the message's native state is vouched for with the
// same provenance: it was produced here, for this tenant, account and model.
func (a *assembler) identify(origin NativeStateEnvelope, call CallAttribution) {
	a.call = call
	a.view.update(func(m *AssistantMessage) {
		m.API, m.Provider, m.Model = origin.API, origin.ProviderID, origin.ModelID
		m.NativeState = TrustedNativeState{env: origin}
	})
}

func (a *assembler) start() { a.publish(StartEvent{Partial: a.view}) }

func (a *assembler) responseModel(model string) {
	a.view.update(func(m *AssistantMessage) { m.ResponseModel = model })
}

func (a *assembler) responseID(id string) {
	a.view.update(func(m *AssistantMessage) { m.ResponseID = id })
}

func (a *assembler) usage(u Usage) {
	a.view.update(func(m *AssistantMessage) { m.Usage = u })
}

// stop records a successful protocol terminal's stop reason. Failed
// terminals are returned by the adapter as an *Error instead.
func (a *assembler) stop(reason StopReason) {
	a.view.update(func(m *AssistantMessage) { m.StopReason = reason })
}

// rawStopReason records the provider's own terminal status.
func (a *assembler) rawStopReason(raw string) {
	a.view.update(func(m *AssistantMessage) { m.RawStopReason = raw })
}

// open appends a new block and returns its content index, which never changes
// however other blocks interleave.
func (a *assembler) open(block AssistantContent) (i int) {
	a.view.update(func(m *AssistantMessage) {
		m.Content = append(m.Content, block)
		i = len(m.Content) - 1
	})
	return i
}

// textStart opens a text block holding what the provider announced with it
// (usually nothing).
func (a *assembler) textStart(initial Text) int {
	i := a.open(initial)
	a.publish(TextStartEvent{ContentIndex: i, Partial: a.view})
	return i
}

func (a *assembler) textDelta(i int, delta string) {
	a.view.update(func(m *AssistantMessage) {
		t := m.Content[i].(Text)
		t.Text += delta
		m.Content[i] = t
	})
	a.publish(TextDeltaEvent{ContentIndex: i, Delta: delta, Partial: a.view})
}

// textContent is the text streamed into block i so far.
func (a *assembler) textContent(i int) (text string) {
	a.view.read(func(m *AssistantMessage) { text = m.Content[i].(Text).Text })
	return text
}

// setTextSignature replaces block i's signature without an event, as pi
// updates a Gemini thought signature arriving on a later part of the block.
func (a *assembler) setTextSignature(i int, sig string) {
	a.view.update(func(m *AssistantMessage) {
		t := m.Content[i].(Text)
		t.Signature = sig
		m.Content[i] = t
	})
}

func (a *assembler) textEnd(i int, text, signature string) {
	a.view.update(func(m *AssistantMessage) { m.Content[i] = Text{Text: text, Signature: signature} })
	a.publish(TextEndEvent{ContentIndex: i, Content: text, Partial: a.view})
}

// thinkingStart opens a thinking block holding what the provider announced
// with it: initial text and signature, or a redacted block's payload.
func (a *assembler) thinkingStart(initial Thinking) int {
	i := a.open(initial)
	a.publish(ThinkingStartEvent{ContentIndex: i, Partial: a.view})
	return i
}

func (a *assembler) thinkingDelta(i int, delta string) {
	a.view.update(func(m *AssistantMessage) {
		t := m.Content[i].(Thinking)
		t.Thinking += delta
		m.Content[i] = t
	})
	a.publish(ThinkingDeltaEvent{ContentIndex: i, Delta: delta, Partial: a.view})
}

// thinkingText is the thinking streamed into block i so far.
func (a *assembler) thinkingText(i int) (text string) {
	a.view.read(func(m *AssistantMessage) { text = m.Content[i].(Thinking).Thinking })
	return text
}

// thinkingEnd closes block i with its final text and signature; a redacted
// block stays redacted.
func (a *assembler) thinkingEnd(i int, thinking, signature string) {
	a.view.update(func(m *AssistantMessage) {
		t := m.Content[i].(Thinking)
		t.Thinking, t.Signature = thinking, signature
		m.Content[i] = t
	})
	a.publish(ThinkingEndEvent{ContentIndex: i, Content: thinking, Partial: a.view})
}

// thinkingSignature is block i's signature so far.
func (a *assembler) thinkingSignature(i int) (sig string) {
	a.view.read(func(m *AssistantMessage) { sig = m.Content[i].(Thinking).Signature })
	return sig
}

// setThinkingSignature replaces block i's signature without an event, as pi
// updates a signature it learns only at the terminal response (Responses)
// or from signature deltas (Anthropic).
func (a *assembler) setThinkingSignature(i int, sig string) {
	a.view.update(func(m *AssistantMessage) {
		t := m.Content[i].(Thinking)
		t.Signature = sig
		m.Content[i] = t
	})
}

// toolCallStart opens a tool call block. As in pi, its arguments start as {}
// whatever raw text the provider announced with the item.
func (a *assembler) toolCallStart(id, name, raw string) int {
	if !a.fitsToolJSON(len(raw)) {
		raw = ""
	}
	i := a.open(ToolCall{ID: id, Name: name, Arguments: "{}", RawArguments: raw})
	a.publish(ToolCallStartEvent{ContentIndex: i, Partial: a.view})
	return i
}

// setToolCallSignature sets block i's thought signature without an event;
// a Gemini call arrives with it.
func (a *assembler) setToolCallSignature(i int, sig string) {
	a.view.update(func(m *AssistantMessage) {
		c := m.Content[i].(ToolCall)
		c.ThoughtSignature = Value(sig)
		m.Content[i] = c
	})
}

// toolCallArguments replaces block i's raw argument text and its display parse.
func (a *assembler) toolCallArguments(i int, raw string) {
	if a.fitsToolJSON(len(raw)) {
		a.replaceToolCallArguments(i, raw)
	}
}

// replaceToolCallArguments sets block i's raw text, already within the limit.
func (a *assembler) replaceToolCallArguments(i int, raw string) {
	args := displayArguments(raw)
	a.view.update(func(m *AssistantMessage) {
		c := m.Content[i].(ToolCall)
		c.RawArguments, c.Arguments = raw, args
		m.Content[i] = c
	})
}

// toolCallDelta sets block i's raw argument text to raw, which ends with
// delta, and publishes the fragment.
func (a *assembler) toolCallDelta(i int, delta, raw string) {
	if !a.fitsToolJSON(len(raw)) {
		return
	}
	a.replaceToolCallArguments(i, raw)
	a.publish(ToolCallDeltaEvent{ContentIndex: i, Delta: delta, Partial: a.view})
}

// toolCallEnd closes block i with the provider's final raw arguments; an
// empty text means no arguments, which pi parses as {}.
func (a *assembler) toolCallEnd(i int, raw string) {
	if !a.fitsToolJSON(len(raw)) {
		return
	}
	args := displayArguments(cmp.Or(raw, "{}"))
	var call ToolCall
	a.view.update(func(m *AssistantMessage) {
		call = m.Content[i].(ToolCall)
		call.RawArguments, call.Arguments = raw, args
		m.Content[i] = call
	})
	a.publish(ToolCallEndEvent{ContentIndex: i, ToolCall: call, Partial: a.view})
}

// blockCount is the number of content blocks opened so far.
func (a *assembler) blockCount() (n int) {
	a.view.read(func(m *AssistantMessage) { n = len(m.Content) })
	return n
}

// hasToolCall reports whether the message holds a tool call block.
func (a *assembler) hasToolCall() (found bool) {
	a.view.read(func(m *AssistantMessage) {
		found = slices.ContainsFunc(m.Content, func(c AssistantContent) bool { _, ok := c.(ToolCall); return ok })
	})
	return found
}

// finish settles the message, publishes the single terminal event and returns
// the call's result. A nil failure means the adapter reached a successful
// protocol terminal. A latched resource-limit failure came first and wins
// over whatever the adapter returned. The terminal always fits the event
// queue, which reserves its place. After finish the view never changes again.
func (a *assembler) finish(meta CallMetadata, failure *Error) (Result, error) {
	if a.limited != nil {
		failure = a.limited
	}
	var final AssistantMessage
	a.view.update(func(m *AssistantMessage) {
		if failure != nil {
			m.StopReason = StopReasonError
			if failure.aborts() {
				m.StopReason = StopReasonAborted
			}
			m.ErrorMessage = failure.Message
		}
		final = m.clone()
	})
	res := Result{Message: final, Metadata: meta}
	var terminal Event = ErrorEvent{Reason: final.StopReason, Message: final.clone(), Err: failure}
	if failure == nil {
		terminal = DoneEvent{Reason: final.StopReason, Message: final.clone()}
	}
	if a.emit != nil {
		a.emit(EventEnvelope{Call: meta.CallAttribution, Event: terminal})
	}
	if failure == nil {
		return res, nil
	}
	return res, failure
}

// aborts reports whether a failure ends the message as aborted rather than
// error. As in pi-ai, cancellation observed once the adapter is talking to the
// provider aborts (waiting for admission is part of that, inside the retry of
// the initial request); cancellation during setup is an error. An attempt's
// own time limit is pi's SDK timeout, an error.
func (e *Error) aborts() bool {
	return (e.Code == CodeCanceled || e.Code == CodeDeadlineExceeded) && !e.attemptTimeout &&
		(e.Phase == PhaseAdmission || e.Phase == PhaseRequest || e.Phase == PhaseStream)
}

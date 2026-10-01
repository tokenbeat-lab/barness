package ai

// assembler is the single production and merge process behind every entry
// point. Adapters report normalized protocol progress to it; it maintains the
// assistant message in the call's PartialView and publishes events. Stream
// gives it a sink; Complete gives it none, so no event is ever queued for a
// caller that will not read it.
type assembler struct {
	view *PartialView
	emit func(Event) // nil for Complete
}

func newAssembler(timestamp int64, emit func(Event)) *assembler {
	return &assembler{
		view: newPartialView(AssistantMessage{StopReason: StopReasonPending, Timestamp: timestamp}),
		emit: emit,
	}
}

func (a *assembler) publish(e Event) {
	if a.emit != nil {
		a.emit(e)
	}
}

// identify records what actually serves the call, once resolved.
func (a *assembler) identify(api API, provider ProviderID, model string) {
	a.view.update(func(m *AssistantMessage) { m.API, m.Provider, m.Model = api, provider, model })
}

func (a *assembler) start() { a.publish(StartEvent{Partial: a.view}) }

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

// open appends a new block and returns its content index, which never changes
// however other blocks interleave.
func (a *assembler) open(block AssistantContent) (i int) {
	a.view.update(func(m *AssistantMessage) {
		m.Content = append(m.Content, block)
		i = len(m.Content) - 1
	})
	return i
}

func (a *assembler) textStart() int {
	i := a.open(Text{})
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

func (a *assembler) textEnd(i int, text, signature string) {
	a.view.update(func(m *AssistantMessage) { m.Content[i] = Text{Text: text, Signature: signature} })
	a.publish(TextEndEvent{ContentIndex: i, Content: text, Partial: a.view})
}

func (a *assembler) thinkingStart() int {
	i := a.open(Thinking{})
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

func (a *assembler) thinkingEnd(i int, thinking, signature string) {
	a.view.update(func(m *AssistantMessage) { m.Content[i] = Thinking{Thinking: thinking, Signature: signature} })
	a.publish(ThinkingEndEvent{ContentIndex: i, Content: thinking, Partial: a.view})
}

// finish settles the message, publishes the single terminal event and returns
// the call's result. A nil failure means the adapter reached a successful
// protocol terminal. After finish the view never changes again.
func (a *assembler) finish(meta CallMetadata, failure *Error) (Result, error) {
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
	if failure == nil {
		a.publish(DoneEvent{Reason: final.StopReason, Message: final.clone()})
		return res, nil
	}
	a.publish(ErrorEvent{Reason: final.StopReason, Message: final.clone(), Err: failure})
	return res, failure
}

// aborts reports whether a failure ends the message as aborted rather than
// error. As in pi-ai, cancellation observed once the adapter is talking to the
// provider aborts; cancellation during setup is an error.
func (e *Error) aborts() bool {
	return (e.Code == CodeCanceled || e.Code == CodeDeadlineExceeded) &&
		(e.Phase == PhaseRequest || e.Phase == PhaseStream)
}

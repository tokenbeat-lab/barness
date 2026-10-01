package ai

// assembler is the single production and merge process behind every entry
// point. Adapters report normalized protocol progress to it; it maintains the
// assistant message and publishes events. Stream gives it a sink; Complete
// gives it none, so no event is ever queued for a caller that will not read it.
type assembler struct {
	msg  AssistantMessage
	emit func(Event) // nil for Complete
}

func (a *assembler) publish(e Event) {
	if a.emit != nil {
		a.emit(e)
	}
}

// identify records what actually serves the call, once resolved.
func (a *assembler) identify(api API, provider ProviderID, model string) {
	a.msg.API, a.msg.Provider, a.msg.Model = api, provider, model
}

func (a *assembler) start() { a.publish(StartEvent{}) }

func (a *assembler) responseID(id string) { a.msg.ResponseID = id }

func (a *assembler) usage(u Usage) { a.msg.Usage = u }

// stop records a successful protocol terminal's stop reason. Failed
// terminals are returned by the adapter as an *Error instead.
func (a *assembler) stop(reason StopReason) { a.msg.StopReason = reason }

func (a *assembler) textStart() int {
	a.msg.Content = append(a.msg.Content, Text{})
	i := len(a.msg.Content) - 1
	a.publish(TextStartEvent{ContentIndex: i})
	return i
}

func (a *assembler) textDelta(i int, delta string) {
	t := a.msg.Content[i].(Text)
	t.Text += delta
	a.msg.Content[i] = t
	a.publish(TextDeltaEvent{ContentIndex: i, Delta: delta})
}

func (a *assembler) textEnd(i int, text, signature string) {
	a.msg.Content[i] = Text{Text: text, Signature: signature}
	a.publish(TextEndEvent{ContentIndex: i, Content: text})
}

// finish publishes the single terminal event and returns the call's result.
// A nil failure means the adapter reached a successful protocol terminal.
func (a *assembler) finish(meta CallMetadata, failure *Error) (Result, error) {
	if failure == nil {
		res := Result{Message: a.msg.clone(), Metadata: meta}
		a.publish(DoneEvent{Reason: a.msg.StopReason, Message: a.msg.clone()})
		return res, nil
	}
	a.msg.StopReason = StopReasonError
	if failure.aborts() {
		a.msg.StopReason = StopReasonAborted
	}
	a.msg.ErrorMessage = failure.Message
	res := Result{Message: a.msg.clone(), Metadata: meta}
	a.publish(ErrorEvent{Reason: a.msg.StopReason, Message: a.msg.clone(), Err: failure})
	return res, failure
}

// aborts reports whether a failure ends the message as aborted rather than
// error. As in pi-ai, cancellation observed once the adapter is talking to the
// provider aborts; cancellation during setup is an error.
func (e *Error) aborts() bool {
	return (e.Code == CodeCanceled || e.Code == CodeDeadlineExceeded) &&
		(e.Phase == PhaseRequest || e.Phase == PhaseStream)
}

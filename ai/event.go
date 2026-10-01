package ai

// EventType names a stream event. Values match pi-ai's event types.
type EventType string

const (
	EventStart         EventType = "start"
	EventTextStart     EventType = "text_start"
	EventTextDelta     EventType = "text_delta"
	EventTextEnd       EventType = "text_end"
	EventThinkingStart EventType = "thinking_start"
	EventThinkingDelta EventType = "thinking_delta"
	EventThinkingEnd   EventType = "thinking_end"
	EventDone          EventType = "done"
	EventError         EventType = "error"
)

// Event is one stream event. The set of event kinds is closed; use a type
// switch on the concrete types below.
//
// A successful stream starts with StartEvent; no block event precedes it.
// Blocks may interleave, and each keeps the ContentIndex it was opened with.
// A fully consumed stream ends with exactly one DoneEvent or ErrorEvent; a
// call refused before generation starts publishes only the ErrorEvent.
//
// Every non-terminal event refers, through Partial, to the call's single live
// PartialView. Its other fields (ContentIndex, Delta, end Content) are the
// event's own values and never change.
type Event interface {
	Type() EventType
}

// StartEvent opens a stream.
type StartEvent struct {
	Partial *PartialView
}

// TextStartEvent opens text block ContentIndex.
type TextStartEvent struct {
	ContentIndex int
	Partial      *PartialView
}

// TextDeltaEvent appends Delta to text block ContentIndex.
type TextDeltaEvent struct {
	ContentIndex int
	Delta        string
	Partial      *PartialView
}

// TextEndEvent closes text block ContentIndex with its final Content.
type TextEndEvent struct {
	ContentIndex int
	Content      string
	Partial      *PartialView
}

// ThinkingStartEvent opens thinking block ContentIndex.
type ThinkingStartEvent struct {
	ContentIndex int
	Partial      *PartialView
}

// ThinkingDeltaEvent appends Delta to thinking block ContentIndex.
type ThinkingDeltaEvent struct {
	ContentIndex int
	Delta        string
	Partial      *PartialView
}

// ThinkingEndEvent closes thinking block ContentIndex with its final Content,
// which may differ from the concatenated deltas (as in pi, the provider's
// final reasoning text replaces what was streamed).
type ThinkingEndEvent struct {
	ContentIndex int
	Content      string
	Partial      *PartialView
}

// DoneEvent is the successful terminal event; Message equals Result's message.
type DoneEvent struct {
	Reason  StopReason
	Message AssistantMessage
}

// ErrorEvent is the failed terminal event; Message still holds everything
// received so far plus ErrorMessage.
type ErrorEvent struct {
	Reason  StopReason
	Message AssistantMessage
	Err     *Error
}

func (StartEvent) Type() EventType         { return EventStart }
func (TextStartEvent) Type() EventType     { return EventTextStart }
func (TextDeltaEvent) Type() EventType     { return EventTextDelta }
func (TextEndEvent) Type() EventType       { return EventTextEnd }
func (ThinkingStartEvent) Type() EventType { return EventThinkingStart }
func (ThinkingDeltaEvent) Type() EventType { return EventThinkingDelta }
func (ThinkingEndEvent) Type() EventType   { return EventThinkingEnd }
func (DoneEvent) Type() EventType          { return EventDone }
func (ErrorEvent) Type() EventType         { return EventError }

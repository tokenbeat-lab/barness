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
	EventToolCallStart EventType = "toolcall_start"
	EventToolCallDelta EventType = "toolcall_delta"
	EventToolCallEnd   EventType = "toolcall_end"
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

// EventEnvelope is one stream event together with the attribution of the
// call that produced it, so a host merging several streams routes every event
// by its own immutable reference instead of by arrival order or a "current
// tenant" (spec I2). Call is the attribution as it stood when the event was
// published: every event of a resolved call names what serves it, and an
// error terminal of a call refused before resolution reports Resolved=false.
// The terminal's Call equals the Result's Metadata.CallAttribution.
//
// The envelope is a barness extension; the events themselves are pi's. Which
// of its fields reach the host's own clients is the host's decision.
type EventEnvelope struct {
	Call  CallAttribution
	Event Event
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

// ToolCallStartEvent opens tool call block ContentIndex. The call's ID and
// name are in the view's block from the start.
type ToolCallStartEvent struct {
	ContentIndex int
	Partial      *PartialView
}

// ToolCallDeltaEvent appends Delta, a raw fragment of the argument JSON as the
// provider sent it, to tool call block ContentIndex. Fragments may split JSON
// tokens and UTF-8 text anywhere a JSON string allows.
type ToolCallDeltaEvent struct {
	ContentIndex int
	Delta        string
	Partial      *PartialView
}

// ToolCallEndEvent closes tool call block ContentIndex with the final call. A
// closed call is still not executable: only ValidateToolCall on the finished
// message decides that.
type ToolCallEndEvent struct {
	ContentIndex int
	ToolCall     ToolCall
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
func (ToolCallStartEvent) Type() EventType { return EventToolCallStart }
func (ToolCallDeltaEvent) Type() EventType { return EventToolCallDelta }
func (ToolCallEndEvent) Type() EventType   { return EventToolCallEnd }
func (DoneEvent) Type() EventType          { return EventDone }
func (ErrorEvent) Type() EventType         { return EventError }

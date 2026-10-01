package ai

// EventType names a stream event. Values match pi-ai's event types.
type EventType string

const (
	EventStart     EventType = "start"
	EventTextStart EventType = "text_start"
	EventTextDelta EventType = "text_delta"
	EventTextEnd   EventType = "text_end"
	EventDone      EventType = "done"
	EventError     EventType = "error"
)

// Event is one stream event. The set of event kinds is closed; use a type
// switch on the concrete types below.
type Event interface {
	Type() EventType
}

// StartEvent opens a stream. No content event precedes it.
type StartEvent struct{}

// TextStartEvent opens text block ContentIndex.
type TextStartEvent struct {
	ContentIndex int
}

// TextDeltaEvent appends Delta to text block ContentIndex.
type TextDeltaEvent struct {
	ContentIndex int
	Delta        string
}

// TextEndEvent closes text block ContentIndex with its final Content.
type TextEndEvent struct {
	ContentIndex int
	Content      string
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

func (StartEvent) Type() EventType     { return EventStart }
func (TextStartEvent) Type() EventType { return EventTextStart }
func (TextDeltaEvent) Type() EventType { return EventTextDelta }
func (TextEndEvent) Type() EventType   { return EventTextEnd }
func (DoneEvent) Type() EventType      { return EventDone }
func (ErrorEvent) Type() EventType     { return EventError }

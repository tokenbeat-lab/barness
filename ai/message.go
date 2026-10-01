package ai

// Message is one entry of the transcript a host passes for a call:
// UserMessage, AssistantMessage (a previous turn's result) or
// ToolResultMessage. The set of message kinds is closed.
type Message interface {
	isMessage()
}

// UserContent is a content block allowed in a UserMessage.
type UserContent interface {
	isUserContent()
}

// AssistantContent is a content block allowed in an AssistantMessage.
type AssistantContent interface {
	isAssistantContent()
}

// Text is a text content block.
type Text struct {
	Text string `json:"text"`
	// Signature is opaque provider state attached to assistant text (for
	// Responses, the output message id as pi's TextSignatureV1 JSON). It is
	// empty on user text.
	Signature string `json:"textSignature,omitempty"`
}

func (Text) isUserContent()      {}
func (Text) isAssistantContent() {}

// Thinking is a reasoning content block of an assistant message.
type Thinking struct {
	Thinking string `json:"thinking"`
	// Signature is opaque provider replay state for the block (for Responses,
	// the reasoning item JSON as in pi). Replaying it is ticket 07's native
	// state contract; until then it is only reported.
	Signature string `json:"thinkingSignature,omitempty"`
}

func (Thinking) isAssistantContent() {}

// UserMessage is user input.
type UserMessage struct {
	Content []UserContent `json:"content"`
}

func (UserMessage) isMessage() {}

// UserText builds a UserMessage holding one text block.
func UserText(text string) UserMessage {
	return UserMessage{Content: []UserContent{Text{Text: text}}}
}

// StopReason is how a generation turn ended. Values match pi-ai.
type StopReason string

const (
	// StopReasonPending is the stop reason of a message still being produced,
	// as seen through a PartialView. A finished message never carries it.
	StopReasonPending StopReason = "pending"
	StopReasonStop    StopReason = "stop"
	StopReasonLength  StopReason = "length"
	// StopReasonToolUse: the turn completed normally and holds tool calls the
	// host may validate and execute (see ValidateToolCall).
	StopReasonToolUse StopReason = "toolUse"
	StopReasonError   StopReason = "error"
	StopReasonAborted StopReason = "aborted"
)

// Usage is provider-reported token usage converted to pi-ai's shape. Zero
// values do not imply the call was free or that usage was reported.
type Usage struct {
	Input       int64 `json:"input"`
	Output      int64 `json:"output"`
	CacheRead   int64 `json:"cacheRead"`
	CacheWrite  int64 `json:"cacheWrite"`
	TotalTokens int64 `json:"totalTokens"`
}

// AssistantMessage is the final (or partial) result of one generation turn.
// Failed calls still produce one, carrying StopReason error/aborted and
// ErrorMessage.
type AssistantMessage struct {
	Content []AssistantContent `json:"content"`
	// API, Provider and Model identify what actually served the call; they are
	// empty when the call failed before the binding and model were resolved.
	API        API        `json:"api"`
	Provider   ProviderID `json:"provider"`
	Model      string     `json:"model"`
	ResponseID string     `json:"responseId,omitempty"`
	Usage      Usage      `json:"usage"`
	StopReason StopReason `json:"stopReason"`
	// RawStopReason is the provider's own terminal status as pi-ai reports
	// it (Responses: response.status, plus ".<reason>" for an incomplete
	// response with a reason). Empty when no terminal status was received.
	RawStopReason string `json:"rawStopReason,omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
	// Timestamp is when the call started producing this message, in Unix
	// milliseconds as in pi-ai. Every message of a call — terminal event,
	// Result, failures included — carries the same value.
	Timestamp int64 `json:"timestamp"`
}

// An AssistantMessage is also history: a previous turn's Result.Message
// replayed in a later call.
func (AssistantMessage) isMessage() {}

// clone copies the message so the copy shares no mutable storage with m.
// Content blocks are immutable values; a block type holding references (a
// slice or map) must be deep-copied here.
func (m AssistantMessage) clone() AssistantMessage {
	m.Content = append([]AssistantContent(nil), m.Content...)
	return m
}

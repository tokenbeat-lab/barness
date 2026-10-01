package ai

// Message is one entry of the transcript a host passes for a call. The set of
// message kinds is closed; it grows as protocol support lands (assistant and
// tool-result history arrive with tickets 06–08).
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
	StopReasonStop    StopReason = "stop"
	StopReasonLength  StopReason = "length"
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
	API          API        `json:"api"`
	Provider     ProviderID `json:"provider"`
	Model        string     `json:"model"`
	ResponseID   string     `json:"responseId,omitempty"`
	Usage        Usage      `json:"usage"`
	StopReason   StopReason `json:"stopReason"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
}

func (m AssistantMessage) clone() AssistantMessage {
	m.Content = append([]AssistantContent(nil), m.Content...)
	return m
}

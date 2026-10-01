package ai

import "encoding/json"

// Tool declares a function the model may call. Declarations are protocol
// data: barness-ai sends them and reports the model's calls, but never runs a
// tool (spec I6). The host executes calls and returns ToolResultMessages in its
// next logical call.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Parameters is the JSON Schema of the arguments object, sent as is.
	Parameters json.RawMessage `json:"parameters"`
}

func (t Tool) clone() Tool {
	t.Parameters = append(json.RawMessage(nil), t.Parameters...)
	return t
}

// ToolCall is a tool call content block of an assistant message.
//
// Neither field holding arguments makes the call executable. Arguments is a
// best-effort parse that may come from a truncated or malformed stream; it is
// for display and history replay only. Only ValidateToolCall turns a call of a
// finished message into a ValidToolCall a host may decide to execute.
type ToolCall struct {
	// ID is the provider's call identity. For Responses it is pi-ai's
	// "<call_id>|<item id>"; a ToolResultMessage refers to it.
	ID   string `json:"id"`
	Name string `json:"name"`
	// Arguments is pi-ai's `arguments`: the arguments parsed so far, repaired
	// and completed where possible, as JSON text.
	Arguments UncheckedArguments `json:"arguments"`
	// RawArguments is the argument JSON exactly as the provider sent it: the
	// concatenated deltas while streaming, the provider's final text once the
	// call ended. It may be incomplete. Empty when the provider sent none.
	RawArguments string `json:"rawArguments,omitempty"`
}

func (ToolCall) isAssistantContent() {}

// UncheckedArguments is the JSON text of a tool call's arguments as parsed for
// display, in JavaScript JSON.stringify form (pi-ai parity). It is not proof
// that the arguments are complete or match the tool's schema.
type UncheckedArguments string

// MarshalJSON emits the arguments as a JSON value; empty means {}.
func (a UncheckedArguments) MarshalJSON() ([]byte, error) {
	if a == "" {
		return []byte("{}"), nil
	}
	return []byte(a), nil
}

// UnmarshalJSON accepts any JSON value and keeps it in JSON.stringify form, so
// a message read back from storage replays exactly like the original.
func (a *UncheckedArguments) UnmarshalJSON(data []byte) error {
	v, err := parseJSON(string(data))
	if err != nil {
		return err
	}
	*a = UncheckedArguments(stringifyJSON(v))
	return nil
}

// ToolResultMessage is the host's result for one tool call, sent in the next
// logical call after the assistant message holding the call.
type ToolResultMessage struct {
	// ToolCallID is the ToolCall.ID this result answers.
	ToolCallID string              `json:"toolCallId"`
	ToolName   string              `json:"toolName"`
	Content    []ToolResultContent `json:"content"`
	// IsError marks a failed tool execution. Responses has no wire field for
	// it; the content carries the failure.
	IsError bool `json:"isError"`
}

func (ToolResultMessage) isMessage() {}

// ToolResultContent is a content block allowed in a ToolResultMessage: Text
// or Image.
type ToolResultContent interface {
	isToolResultContent()
}

func (Text) isToolResultContent() {}

// ToolResultText builds a ToolResultMessage holding one text block.
func ToolResultText(call ToolCall, text string, isError bool) ToolResultMessage {
	return ToolResultMessage{ToolCallID: call.ID, ToolName: call.Name, Content: []ToolResultContent{Text{Text: text}}, IsError: isError}
}

// ToolChoice is the protocol-neutral tool selection of SimpleOptions, as in
// pi-ai.
type ToolChoice string

const (
	ToolChoiceAuto ToolChoice = "auto"
	ToolChoiceNone ToolChoice = "none"
)

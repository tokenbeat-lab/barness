package ai

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// Request is the host-selected input of one generation turn. The library
// converts it; it never selects, trims or stores history.
type Request struct {
	// SystemPrompt is a convenience input normalized, as in pi-ai, into the
	// leading system instruction.
	SystemPrompt string
	Messages     []Message
	// Tools are the functions the model may call in this turn, sent in order.
	Tools []Tool
}

// clone copies everything the caller could mutate after handing the request
// over, so a call in flight never observes later changes.
func (r Request) clone() Request {
	out := Request{SystemPrompt: r.SystemPrompt, Messages: make([]Message, len(r.Messages))}
	for i, m := range r.Messages {
		switch m := m.(type) {
		case UserMessage:
			m.Content = append([]UserContent(nil), m.Content...)
			out.Messages[i] = m
		case AssistantMessage:
			out.Messages[i] = m.clone()
		case ToolResultMessage:
			m.Content = append([]ToolResultContent(nil), m.Content...)
			out.Messages[i] = m
		default:
			out.Messages[i] = m
		}
	}
	if r.Tools != nil {
		out.Tools = make([]Tool, len(r.Tools))
		for i, t := range r.Tools {
			out.Tools[i] = t.clone()
		}
	}
	return out
}

// Options are the complete, protocol-specific options for the full entry
// points (Stream, Complete). Each implementation belongs to exactly one API
// and must match the binding's API. A nil Options means protocol defaults.
type Options interface {
	api() API
	// validate reports why the options are unusable, or "".
	validate() string
}

// ResponsesOptions are the full options for the OpenAI Responses protocol.
// The remaining fields arrive with ticket 09; the type already fixes the seam
// so the full entry points cannot be called with another protocol's options.
type ResponsesOptions struct {
	// ToolChoice is tool_choice; nil leaves it to the provider.
	ToolChoice *ResponsesToolChoice
}

func (ResponsesOptions) api() API { return APIOpenAIResponses }

func (o ResponsesOptions) validate() string {
	c := o.ToolChoice
	switch {
	case c == nil:
		return ""
	case c.Function != "" && c.Mode != "":
		return "tool choice sets both a mode and a function"
	case c.Function != "" || c.Mode == "auto" || c.Mode == "none" || c.Mode == "required":
		return ""
	}
	return "tool choice mode must be auto, none or required"
}

// SimpleOptions are the protocol-neutral options for StreamSimple and
// CompleteSimple, mapped per model onto the binding's protocol. The remaining
// fields arrive with ticket 09.
type SimpleOptions struct {
	// ToolChoice selects auto or no tool use; empty leaves it to the provider.
	ToolChoice ToolChoice
}

func (o SimpleOptions) validate() string {
	switch o.ToolChoice {
	case "", ToolChoiceAuto, ToolChoiceNone:
		return ""
	}
	return "tool choice must be auto or none"
}

// validate checks the request's structure: every tool needs a unique name and
// a JSON object schema, as the wire protocols require. History content is
// checked where it is encoded.
func (r Request) validate() string {
	seen := map[string]bool{}
	for _, t := range r.Tools {
		switch {
		case t.Name == "":
			return "a tool has no name"
		case seen[t.Name]:
			return "tool " + strconv.Quote(t.Name) + " is declared twice"
		case !isJSONObject(t.Parameters):
			return "tool " + strconv.Quote(t.Name) + " parameters must be a JSON Schema object"
		}
		seen[t.Name] = true
	}
	return ""
}

func isJSONObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	return json.Valid(raw) && len(trimmed) > 0 && trimmed[0] == '{'
}

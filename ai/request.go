package ai

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
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
		case SystemMessage:
			out.Messages[i] = m.clone()
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

// validate checks the request's structure: every tool needs a unique name and
// a JSON object schema, as the wire protocols require; every message and
// block must be set; a system message's changes must be expressible in pi-ai's
// shape; an image needs an image media type, as it is spliced into a data
// URL. Values that only some targets accept are checked where they are
// encoded.
func (r Request) validate() string {
	if problem := validateTools(r.Tools); problem != "" {
		return problem
	}
	for i, m := range r.Messages {
		if problem := validateMessage(m); problem != "" {
			return "message " + strconv.Itoa(i) + ": " + problem
		}
	}
	return ""
}

func validateTools(tools []Tool) string {
	seen := map[string]bool{}
	for _, t := range tools {
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

func validateMessage(m Message) string {
	var blocks []any
	switch m := m.(type) {
	case nil:
		return "no message"
	case UserMessage:
		blocks = anyBlocks(m.Content)
	case AssistantMessage:
		blocks = anyBlocks(m.Content)
	case ToolResultMessage:
		blocks = anyBlocks(m.Content)
	case SystemMessage:
		return validateSystemMessage(m)
	}
	for _, b := range blocks {
		switch b := b.(type) {
		case nil:
			return "a content block is nil"
		case Image:
			if !isImageMediaType(b.MimeType) {
				return "image media type " + strconv.Quote(b.MimeType) + " is not image/<subtype>"
			}
		}
	}
	return ""
}

func validateSystemMessage(m SystemMessage) string {
	if problem := validateTools(m.ToolsAdded); problem != "" {
		return problem
	}
	for _, name := range m.ToolsRemoved {
		if name == "" {
			return "a removed tool has no name"
		}
	}
	seen := map[string]bool{}
	for _, s := range m.Sections {
		switch {
		case seen[s.Name]:
			return "section " + strconv.Quote(s.Name) + " appears twice"
		case s.Removed && s.Text != "":
			return "section " + strconv.Quote(s.Name) + " is removed but has text"
		}
		seen[s.Name] = true
	}
	return ""
}

func anyBlocks[T any](in []T) []any {
	out := make([]any, len(in))
	for i, b := range in {
		out[i] = b
	}
	return out
}

// isImageMediaType accepts "image/<subtype>" with RFC 6838 name characters
// and no parameters.
func isImageMediaType(s string) bool {
	sub, ok := strings.CutPrefix(s, "image/")
	if !ok || sub == "" || len(sub) > 127 {
		return false
	}
	for _, c := range sub {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case strings.ContainsRune("!#$&-^_.+", c):
		default:
			return false
		}
	}
	return true
}

func isJSONObject(raw json.RawMessage) bool {
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	return json.Valid(raw) && len(trimmed) > 0 && trimmed[0] == '{'
}

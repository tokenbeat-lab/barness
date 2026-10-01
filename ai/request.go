package ai

// Request is the host-selected input of one generation turn. The library
// converts it; it never selects, trims or stores history.
type Request struct {
	// SystemPrompt is a convenience input normalized, as in pi-ai, into the
	// leading system instruction.
	SystemPrompt string
	Messages     []Message
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
		default:
			out.Messages[i] = m
		}
	}
	return out
}

// Options are the complete, protocol-specific options for the full entry
// points (Stream, Complete). Each implementation belongs to exactly one API
// and must match the binding's API. A nil Options means protocol defaults.
type Options interface {
	api() API
}

// ResponsesOptions are the full options for the OpenAI Responses protocol.
// Fields arrive with ticket 09; the type already fixes the seam so the full
// entry points cannot be called with another protocol's options.
type ResponsesOptions struct{}

func (ResponsesOptions) api() API { return APIOpenAIResponses }

// SimpleOptions are the protocol-neutral options for StreamSimple and
// CompleteSimple, mapped per model onto the binding's protocol (ticket 09).
type SimpleOptions struct{}

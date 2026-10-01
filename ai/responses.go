package ai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

// responsesAdapter implements the OpenAI Responses protocol on the OpenAI Go
// SDK. It deliberately avoids openai.NewClient, whose defaults read
// OPENAI_API_KEY, OPENAI_BASE_URL, OPENAI_ORG_ID, OPENAI_PROJECT_ID and
// OPENAI_CUSTOM_HEADERS from the environment. A fresh, unauthenticated-until-
// now service is built per call from explicit options only, so no
// authenticated client outlives the call.
type responsesAdapter struct{}

// simpleOptions maps the protocol-neutral options; pi passes toolChoice
// through unchanged.
func (responsesAdapter) simpleOptions(_ Model, o SimpleOptions) Options {
	var full ResponsesOptions
	if o.ToolChoice != "" {
		full.ToolChoice = &ResponsesToolChoice{Mode: string(o.ToolChoice)}
	}
	return full
}

func (responsesAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(ResponsesOptions) // nil means protocol defaults
	body, err := buildResponsesBody(ac.model, ac.request, opts)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	var res *http.Response
	svc := responses.NewResponseService(
		option.WithHTTPClient(ac.http),
		option.WithBaseURL(ac.endpoint),
		option.WithAPIKey(ac.apiKey.reveal()),
		// Retries are barness-ai's decision (default 0, ticket 11), never the SDK's.
		option.WithMaxRetries(0),
		option.WithHeader("User-Agent", userAgent),
		option.WithRequestBody("application/json", body),
		option.WithResponseInto(&res),
	)
	failures := responsesFailures{provider: ac.model.Provider, apiKey: ac.apiKey}
	// Close releases the response body on every path; the SDK has already
	// closed the body of an error status.
	stream := svc.NewStreaming(ctx, responses.ResponseNewParams{})
	defer stream.Close()
	if err := stream.Err(); err != nil {
		return failures.request(ctx, err, res)
	}
	if res.StatusCode >= 300 {
		return failures.status(res)
	}

	out.start()
	failure := readResponsesStream(ctx, stream, responsesParser{out: out, failures: failures, slots: map[int64]responsesSlot{}})
	if failure != nil {
		failure.ProviderRequestID = res.Header.Get("x-request-id")
	}
	return failure
}

// readResponsesStream feeds the stream to p until a protocol terminal, a
// failure or the end of the body. As in pi, only a terminal response event is
// success: the SDK ends quietly at EOF, and when the call is interrupted.
func readResponsesStream(ctx context.Context, stream *ssestream.Stream[responses.ResponseStreamEventUnion], p responsesParser) *Error {
	for stream.Next() {
		if failure := p.handle(stream.Current()); failure != nil {
			return failure
		}
	}
	if ctx.Err() != nil {
		msg := msgNoTerminal
		if p.terminal {
			msg = msgAbortedAfterTerminal
		}
		return interrupted(ctx, PhaseStream, msg)
	}
	if err := stream.Err(); err != nil {
		return p.failures.stream(err)
	}
	if p.failure != nil {
		return p.failure
	}
	if !p.terminal {
		return newError(CodeProtocol, PhaseStream, msgNoTerminal)
	}
	return nil
}

// responsesParser normalizes Responses stream events into assembler calls. It
// ports the text, reasoning, function call and terminal paths of pi-ai's
// processResponsesStream. Unknown events and items are ignored; so are custom
// (grammar) tool calls and tool namespaces, which only tools barness-ai does
// not declare produce.
type responsesParser struct {
	out      *assembler
	failures responsesFailures
	slots    map[int64]responsesSlot // output_index → open block
	terminal bool
	// failure is a terminal status that maps to an error. As in pi, the stream
	// is still read to its end; the failure is reported afterwards.
	failure *Error
}

// responsesSlot is the open content block an output item streams into.
type responsesSlot struct {
	kind  string // the item type: "message", "reasoning" or "function_call"
	index int    // content index in the assistant message
	raw   string // a function call's argument JSON received so far (pi's partialJson)
}

// slot returns the open block for outputIndex if it has the given kind.
func (p *responsesParser) slot(outputIndex int64, kind string) (int, bool) {
	s, ok := p.slots[outputIndex]
	return s.index, ok && s.kind == kind
}

// open starts a block for a message, reasoning or function call item; other
// items are not handled.
func (p *responsesParser) open(outputIndex int64, item responses.ResponseOutputItemUnion) (int, bool) {
	s := responsesSlot{kind: item.Type}
	switch item.Type {
	case "message":
		s.index = p.out.textStart()
	case "reasoning":
		s.index = p.out.thinkingStart()
	case "function_call":
		// pi's tool call id joins the call id the result must answer with
		// the item id same-model replay needs.
		s.raw = item.Arguments.OfString
		s.index = p.out.toolCallStart(item.CallID+"|"+item.ID, item.Name, s.raw)
	default:
		return 0, false
	}
	p.slots[outputIndex] = s
	return s.index, true
}

// argumentsDelta appends a function call's argument fragment.
func (p *responsesParser) argumentsDelta(outputIndex int64, delta string) {
	s, ok := p.slots[outputIndex]
	if !ok || s.kind != "function_call" {
		return
	}
	s.raw += delta
	p.slots[outputIndex] = s
	p.out.toolCallDelta(s.index, delta, s.raw)
}

// argumentsDone takes the provider's complete argument text. As in pi, the
// part beyond what was streamed is published as a final delta; a text that
// does not extend it replaces it silently.
func (p *responsesParser) argumentsDone(outputIndex int64, arguments string) {
	s, ok := p.slots[outputIndex]
	if !ok || s.kind != "function_call" {
		return
	}
	previous := s.raw
	s.raw = arguments
	p.slots[outputIndex] = s
	if rest, extends := strings.CutPrefix(arguments, previous); extends && rest != "" {
		p.out.toolCallDelta(s.index, rest, arguments)
		return
	}
	p.out.toolCallArguments(s.index, arguments)
}

func (p *responsesParser) handle(ev responses.ResponseStreamEventUnion) *Error {
	switch ev.Type {
	case "response.created":
		p.out.responseID(ev.Response.ID)
	case "response.output_item.added":
		p.open(ev.OutputIndex, ev.Item)
	case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		if i, ok := p.slot(ev.OutputIndex, "reasoning"); ok {
			p.out.thinkingDelta(i, ev.Delta)
		}
	case "response.reasoning_summary_part.done":
		// pi separates summary parts with a blank line.
		if i, ok := p.slot(ev.OutputIndex, "reasoning"); ok {
			p.out.thinkingDelta(i, "\n\n")
		}
	case "response.output_text.delta", "response.refusal.delta":
		if i, ok := p.slot(ev.OutputIndex, "message"); ok {
			p.out.textDelta(i, ev.Delta)
		}
	case "response.function_call_arguments.delta":
		p.argumentsDelta(ev.OutputIndex, ev.Delta)
	case "response.function_call_arguments.done":
		p.argumentsDone(ev.OutputIndex, ev.Arguments)
	case "response.output_item.done":
		p.itemDone(ev.OutputIndex, ev.Item)
	case "response.completed", "response.incomplete":
		return p.finalize(ev.Response)
	case "error":
		return p.failures.upstream("Error Code " + jsText(ev.JSON.Code, ev.Code) + ": " + jsText(ev.JSON.Message, ev.Message))
	case "response.failed":
		return p.failed(ev.Response)
	}
	return nil
}

// failed ends the call at a response.failed terminal with pi's message. As in
// pi, the failed response's usage and id are not taken over.
func (p *responsesParser) failed(r responses.Response) *Error {
	p.terminal = true
	p.out.rawStopReason(string(r.Status))
	switch {
	case r.JSON.Error.Valid():
		return p.failures.upstream(cmp.Or(string(r.Error.Code), "unknown") + ": " + cmp.Or(r.Error.Message, "no message"))
	case r.IncompleteDetails.Reason != "":
		return p.failures.upstream("incomplete: " + string(r.IncompleteDetails.Reason))
	}
	return p.failures.upstream("Unknown error (no error details in response)")
}

// itemDone closes the item's block with the item's authoritative content. As
// pi's getOrCreateSlot, it opens the block first when the item was never
// announced, and ignores the item when its output index holds a block of
// another kind.
func (p *responsesParser) itemDone(outputIndex int64, item responses.ResponseOutputItemUnion) {
	s, ok := p.slots[outputIndex]
	if ok && s.kind != item.Type {
		return
	}
	if !ok {
		if _, ok = p.open(outputIndex, item); !ok {
			return
		}
		s = p.slots[outputIndex]
	}
	i := s.index
	delete(p.slots, outputIndex)
	switch item.Type {
	case "message":
		var text strings.Builder
		for _, c := range item.Content {
			if c.Type == "output_text" {
				text.WriteString(c.Text)
			} else {
				text.WriteString(c.Refusal)
			}
		}
		p.out.textEnd(i, text.String(), encodeTextSignature(item.ID, string(item.Phase)))
	case "reasoning":
		summary := make([]string, len(item.Summary))
		for j, s := range item.Summary {
			summary[j] = s.Text
		}
		content := make([]string, len(item.Content))
		for j, c := range item.Content {
			content[j] = c.Text
		}
		thinking := cmp.Or(strings.Join(summary, "\n\n"), strings.Join(content, "\n\n"), p.out.thinkingText(i))
		p.out.thinkingEnd(i, thinking, reasoningSignature(item.RawJSON()))
	case "function_call":
		p.out.toolCallEnd(i, cmp.Or(item.Arguments.OfString, s.raw))
	}
}

// finalize records a completed or incomplete response. A status pi cannot map
// fails at once; a status that maps to an error fails once the stream ends.
func (p *responsesParser) finalize(r responses.Response) *Error {
	p.terminal = true
	if r.ID != "" {
		p.out.responseID(r.ID)
	}
	if r.JSON.Usage.Valid() {
		u := r.Usage
		cached, cacheWrite := u.InputTokensDetails.CachedTokens, u.InputTokensDetails.CacheWriteTokens
		// OpenAI counts cached and cache-write tokens inside input_tokens.
		p.out.usage(Usage{
			Input:       max(0, u.InputTokens-cached-cacheWrite),
			Output:      u.OutputTokens,
			CacheRead:   cached,
			CacheWrite:  cacheWrite,
			TotalTokens: u.TotalTokens,
		})
	}
	status, reason := string(r.Status), string(r.IncompleteDetails.Reason)
	raw := status
	if reason != "" {
		raw += "." + reason
	}
	p.out.rawStopReason(raw)
	stop, msg, known := mapResponsesStatus(status, reason)
	switch {
	case !known:
		return newError(CodeProtocol, PhaseStream, msg)
	case stop == StopReasonError:
		// pi-ai: throw new Error(output.errorMessage || "An unknown error occurred").
		p.failure = p.failures.upstream(cmp.Or(msg, "An unknown error occurred"))
	case stop == StopReasonStop && p.out.hasToolCall():
		// pi: a normally completed turn holding tool calls is tool use.
		p.out.stop(StopReasonToolUse)
	default:
		p.out.stop(stop)
	}
	return nil
}

// mapResponsesStatus ports pi-ai's mapStopReason; known is false for a
// status pi throws on.
func mapResponsesStatus(status, incompleteReason string) (stop StopReason, msg string, known bool) {
	switch status {
	case "", "completed", "in_progress", "queued":
		return StopReasonStop, "", true
	case "incomplete":
		if incompleteReason == "max_output_tokens" {
			return StopReasonLength, "", true
		}
		if incompleteReason != "" {
			return StopReasonError, "Response incomplete: " + incompleteReason, true
		}
		return StopReasonError, "Response incomplete without a provider reason", true
	case "failed", "cancelled":
		return StopReasonError, "", true
	default:
		return StopReasonError, "Unhandled stop reason: " + status, false
	}
}

// reasoningSignature is pi's thinkingSignature for a Responses reasoning item:
// JSON.stringify of the item as received. The raw item JSON keeps the wire's
// key order; compacting removes the framing's whitespace. Escapes and number
// spellings stay as sent, where JSON.stringify would normalize them — exact
// replay encoding is ticket 07's native state contract.
func reasoningSignature(raw string) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(raw)); err != nil {
		return raw
	}
	return buf.String()
}

// encodeTextSignature produces pi-ai's TextSignatureV1 JSON: {"v":1,"id":…[,"phase":…]}.
func encodeTextSignature(id, phase string) string {
	sig := struct {
		V     int    `json:"v"`
		ID    string `json:"id"`
		Phase string `json:"phase,omitempty"`
	}{V: 1, ID: id, Phase: phase}
	data, _ := marshalJS(sig)
	return string(data)
}

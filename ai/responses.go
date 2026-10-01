package ai

import (
	"cmp"
	"context"
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

// headers are the bearer authentication, barness-ai's User-Agent and pi's
// OpenAI session affinity headers carrying the derived cache key.
func (responsesAdapter) headers(ac adapterCall) http.Header {
	h := http.Header{}
	h.Set("Authorization", "Bearer "+ac.apiKey.reveal())
	h.Set("User-Agent", userAgent)
	opts, _ := ac.options.(ResponsesOptions) // nil means protocol defaults
	if key := opts.cacheKey(ac.cache); key != "" {
		h.Set("session_id", key)
		h.Set("x-client-request-id", key)
	}
	return h
}

func (responsesAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(ResponsesOptions) // nil means protocol defaults
	cacheKey := opts.cacheKey(ac.cache)
	body, err := buildResponsesBody(ac.model, ac.history, opts, cacheKey)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	// As in pi, the payload callback runs once per logical call, outside
	// any retry of the initial request (ticket 11).
	body, failure := ac.hooks.payload(ctx, body, authorizeResponsesPayload(ac.model, cacheKey, ac.hostedTools))
	if failure != nil {
		return failure
	}
	var res *http.Response
	reqOpts := []option.RequestOption{
		option.WithHTTPClient(ac.http),
		option.WithBaseURL(ac.endpoint),
		// Retries are barness-ai's decision (default 0, ticket 11), never the SDK's.
		option.WithMaxRetries(0),
		option.WithRequestBody("application/json", body),
		option.WithResponseInto(&res),
	}
	// Authentication travels in ac.header, so the SDK's own API key setting
	// stays empty and adds no second Authorization.
	reqOpts = append(reqOpts, headerOptions(ac.header)...)
	svc := responses.NewResponseService(reqOpts...)
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
	if failure := ac.hooks.response(ctx, res); failure != nil {
		failure.ProviderRequestID = res.Header.Get("x-request-id")
		return failure
	}

	out.start()
	failure = readResponsesStream(ctx, stream, responsesParser{out: out, failures: failures, slots: map[int64]responsesSlot{}, reasoning: map[string]int{}})
	if failure != nil {
		failure.ProviderRequestID = res.Header.Get("x-request-id")
	}
	return failure
}

// headerOptions sends header exactly: each name with its values in order,
// and a name with no values suppressed.
func headerOptions(header http.Header) []option.RequestOption {
	var opts []option.RequestOption
	for name, values := range header {
		if len(values) == 0 {
			opts = append(opts, option.WithHeaderDel(name))
			continue
		}
		opts = append(opts, option.WithHeader(name, values[0]))
		for _, v := range values[1:] {
			opts = append(opts, option.WithHeaderAdd(name, v))
		}
	}
	return opts
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
	// reasoning maps a finished reasoning item's id to its thinking block,
	// for the terminal response's encrypted content backfill.
	reasoning map[string]int
	terminal  bool
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
		p.reasoning[item.ID] = i
	case "function_call":
		p.out.toolCallEnd(i, cmp.Or(item.Arguments.OfString, s.raw))
	}
}

// finalize records a completed or incomplete response. A status pi cannot map
// fails at once; a status that maps to an error fails once the stream ends.
func (p *responsesParser) finalize(r responses.Response) *Error {
	p.terminal = true
	p.backfillReasoning(r.Output)
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
// JSON.stringify of the item as received, so key order, escapes and number
// spellings are JavaScript's and same-model replay sends exactly what pi
// sends.
func reasoningSignature(raw string) string {
	item, err := parseJSON(raw)
	if err != nil {
		return raw // the SDK decoded it, so it is JSON; kept as is otherwise
	}
	return stringifyJSON(item)
}

// backfillReasoning ports pi's backfillReasoningSignatures: some services
// omit a reasoning item's encrypted_content from output_item.done and send it
// only in the terminal response's output. It is added to the finished block's
// signature unless that already holds a truthy one, in place when the key is
// present and at the end otherwise, so stateless (store:false) replay keeps
// the encrypted reasoning.
func (p *responsesParser) backfillReasoning(output []responses.ResponseOutputItemUnion) {
	for _, item := range output {
		if item.Type != "reasoning" {
			continue
		}
		encrypted, err := parseJSON(item.JSON.EncryptedContent.Raw())
		i, ok := p.reasoning[item.ID]
		if err != nil || !jsTruthy(encrypted) || !ok {
			continue
		}
		if sig, ok := withEncryptedContent(p.out.thinkingSignature(i), encrypted); ok {
			p.out.setThinkingSignature(i, sig)
		}
	}
}

// withEncryptedContent is pi's {...storedItem, encrypted_content}: the stored
// reasoning item with encrypted set, in place when the key exists. ok is
// false when the signature is empty, not an object, or already holds a
// truthy encrypted_content.
func withEncryptedContent(signature string, encrypted any) (string, bool) {
	stored, err := parseJSON(signature)
	obj, isObject := stored.(*jsonObject)
	if signature == "" || err != nil || !isObject {
		return "", false
	}
	if current, _ := obj.get("encrypted_content"); jsTruthy(current) {
		return "", false
	}
	obj.set("encrypted_content", encrypted)
	return stringifyJSON(obj), true
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

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
	caps := responsesCapabilitiesOf(ac.model.Provider)
	if key := opts.cacheKey(ac.cache, caps); key != "" && caps.sessionAffinity {
		h.Set("session_id", key)
		h.Set("x-client-request-id", key)
	}
	return h
}

func (responsesAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	opts, _ := ac.options.(ResponsesOptions) // nil means protocol defaults
	caps := responsesCapabilitiesOf(ac.model.Provider)
	if problem := caps.unsupportedOption(opts); problem != "" {
		return newError(CodeInvalidRequest, PhaseRequest, problem)
	}
	cacheKey := opts.cacheKey(ac.cache, caps)
	body, err := buildResponsesBody(ac.model, ac.history, opts, caps, cacheKey)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	// As in pi, the payload callback runs once per logical call, outside
	// any retry of the initial request.
	authorize := authorizeResponsesPayload(ac.model, cacheKey, ac.hostedTools)
	body, failure := ac.hooks.payload(ctx, body, payloadAuthorization(authorize))
	if failure != nil {
		return failure
	}
	if failure := validateGLMResponsesFinalBody(ac.model, body, authorize); failure != nil {
		return failure
	}
	if failure := validateMiniMaxResponsesFinalBody(ac.model, body, authorize); failure != nil {
		return failure
	}
	if failure := validateKimiResponsesFinalBody(ac.model, body, authorize); failure != nil {
		return failure
	}
	if failure := ac.limits.checkRequestBody(body); failure != nil {
		return failure
	}
	var delivered deliveredBody
	reqOpts := openAIRequestOptions(ac, body, opts.TimeoutMs, opts.requestTimeout(), &delivered)
	svc := responses.NewResponseService(reqOpts...)
	failures := newResponsesFailures(ac.model.Provider, ac.apiKey, ac.initial.clock)
	ctx = withProtocolTimeout(ctx, opts.requestTimeout())
	var stream *ssestream.Stream[responses.ResponseStreamEventUnion]
	var res *http.Response
	failure = ac.initial.send(ctx, func(ctx context.Context) attemptOutcome {
		res = nil
		stream = svc.NewStreaming(ctx, responses.ResponseNewParams{}, option.WithResponseInto(&res))
		if err := stream.Err(); err != nil {
			// The SDK has already closed the body of an error status, but
			// not of a response it dropped as ctx ended.
			_ = stream.Close()
			o := failures.request(ctx, err, res)
			delivered.close()
			return o
		}
		if res.StatusCode >= 300 {
			o := failures.status(res)
			_ = stream.Close()
			return o
		}
		return attemptOutcome{status: res.StatusCode, header: res.Header, providerRequestID: res.Header.Get(requestIDHeaderOf(ac.model.Provider))}
	})
	if failure != nil {
		return failure
	}
	// Close releases the response body on every later path. The typed
	// stream stays unread; the raw SDK decoder preserves SSE event names.
	defer stream.Close()
	requestID := res.Header.Get(requestIDHeaderOf(ac.model.Provider))
	if failure := ac.hooks.response(ctx, res); failure != nil {
		failure.ProviderRequestID = requestID
		if responsesCapabilitiesOf(ac.model.Provider).implementation == responsesGLM || responsesCapabilitiesOf(ac.model.Provider).implementation == responsesMiniMax {
			failure.HTTPStatus = res.StatusCode
		}
		return failure
	}

	out.start()
	failure = readResponsesStream(ctx, ssestream.NewDecoder(res), responsesParser{out: out, failures: failures, slots: map[responsesIndex]responsesSlot{}, reasoning: map[string]int{},
		kimi: newKimiResponsesState(ac.model.Provider), glm: responsesCapabilitiesOf(ac.model.Provider).implementation == responsesGLM, minimax: newMiniMaxResponsesState(ac.model.Provider), usage: responsesUsage{model: ac.model, pricesByTier: caps.serviceTier && responsesCapabilitiesOf(ac.model.Provider).implementation != responsesMiniMax, requestedTier: opts.ServiceTier}, attempt: ac.initial})
	if failure != nil {
		failure.ProviderRequestID = requestID
		if responsesCapabilitiesOf(ac.model.Provider).implementation == responsesGLM || responsesCapabilitiesOf(ac.model.Provider).implementation == responsesMiniMax {
			failure.HTTPStatus = res.StatusCode
		}
	}
	return failure
}

// responsesParser normalizes Responses stream events into assembler calls. It
// ports the text, reasoning, function call and terminal paths of pi-ai's
// processResponsesStream. Unknown events and items are ignored; so are custom
// (grammar) tool calls and tool namespaces, which only tools barness-ai does
// not declare produce.
type responsesParser struct {
	kimi        *kimiResponsesState
	minimax     *miniMaxResponsesState
	glm         bool
	glmNative   map[responsesIndex]string
	glmFinished map[responsesIndex]bool
	out         *assembler
	failures    responsesFailures
	slots       map[responsesIndex]responsesSlot // output_index → open block
	unfinished  map[int]responsesToolRef         // content index → call lacking item.done
	// reasoning maps a finished reasoning item's id to its thinking block,
	// for the terminal response's encrypted content backfill.
	reasoning map[string]int
	terminal  bool
	stop      StopReason
	// failure is a terminal status that maps to an error. As in pi, the stream
	// is still read to its end; the failure is reported afterwards.
	failure *Error
	usage   responsesUsage
	// attempt records the usage the terminal response reported for the
	// attempt being read.
	attempt *initialRequest
}

type responsesToolRef struct{ name, id string }

// responsesIndex preserves pi's Map key distinction between undefined, null
// and a numeric index. The SDK's int64 alone collapses all three to zero and
// lets a done on another key incorrectly complete an unfinished tool block.
type responsesIndex struct {
	kind  string
	value int64
}

func outputIndexOf(ev responses.ResponseStreamEventUnion) responsesIndex {
	switch ev.JSON.OutputIndex.Raw() {
	case "":
		return responsesIndex{kind: "missing"}
	case "null":
		return responsesIndex{kind: "null"}
	default:
		return responsesIndex{kind: "number", value: ev.OutputIndex}
	}
}

// responsesSlot is the open content block an output item streams into.
type responsesSlot struct {
	kind  string // the item type: "message", "reasoning" or "function_call"
	index int    // content index in the assistant message
	// raw is a function call's argument JSON received so far (pi's partialJson).
	raw  *argumentText
	done *string
}

// slot returns the open block for outputIndex if it has the given kind.
func (p *responsesParser) slot(outputIndex responsesIndex, kind string) (int, bool) {
	s, ok := p.slots[outputIndex]
	return s.index, ok && s.kind == kind
}

// open starts a block for a message, reasoning or function call item; other
// items are not handled.
func (p *responsesParser) open(outputIndex responsesIndex, item responses.ResponseOutputItemUnion) (int, bool) {
	s := responsesSlot{kind: item.Type}
	switch item.Type {
	case "message":
		s.index = p.out.textStart(Text{})
	case "reasoning":
		s.index = p.out.thinkingStart(Thinking{})
	case "function_call":
		// pi's tool call id joins the call id the result must answer with
		// the item id same-model replay needs.
		s.raw = newArgumentText(item.Arguments.OfString)
		id := item.CallID + "|" + item.ID
		if p.kimi != nil {
			id = kimiResponsesToolID(item.CallID, item.ID)
		}
		if p.minimax != nil {
			id = miniMaxResponsesToolID(item.CallID, item.ID)
		}
		if p.glm {
			id = glmResponsesToolID(item.CallID, item.ID)
		}
		s.index = p.out.toolCallStart(id, item.Name, s.raw.String())
		if p.unfinished == nil {
			p.unfinished = make(map[int]responsesToolRef)
		}
		p.unfinished[s.index] = responsesToolRef{name: item.Name, id: id}
	default:
		return 0, false
	}
	p.slots[outputIndex] = s
	return s.index, true
}

// argumentsDelta appends a function call's argument fragment.
func (p *responsesParser) argumentsDelta(outputIndex responsesIndex, delta string) {
	s, ok := p.slots[outputIndex]
	if !ok || s.kind != "function_call" || !p.out.fitsToolJSON(s.raw.Len()+len(delta)) {
		return
	}
	p.out.toolCallDelta(s.index, delta, s.raw.append(delta))
}

// argumentsDone takes the provider's complete argument text. As in pi, the
// part beyond what was streamed is published as a final delta; a text that
// does not extend it replaces it silently.
func (p *responsesParser) argumentsDone(outputIndex responsesIndex, arguments string) {
	s, ok := p.slots[outputIndex]
	if !ok || s.kind != "function_call" {
		return
	}
	previous := s.raw.String()
	s.raw = newArgumentText(arguments)
	p.slots[outputIndex] = s
	if rest, extends := strings.CutPrefix(arguments, previous); extends && rest != "" {
		p.out.toolCallDelta(s.index, rest, arguments)
		return
	}
	p.out.toolCallArguments(s.index, arguments)
}

func (p *responsesParser) handle(ev responses.ResponseStreamEventUnion) *Error {
	index := outputIndexOf(ev)
	switch ev.Type {
	case "response.created":
		p.out.responseID(ev.Response.ID)
		if p.minimax != nil {
			p.out.responseModel(ev.Response.Model)
		}
	case "response.output_item.added":
		p.open(index, ev.Item)
	// response.reasoning_text.done (DeepSeek) repeats the streamed text;
	// as in pi, the reasoning item's output_item.done closes the block.
	case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		if i, ok := p.slot(index, "reasoning"); ok {
			p.out.thinkingDelta(i, ev.Delta)
		}
	case "response.reasoning_summary_part.done":
		// pi separates summary parts with a blank line.
		if i, ok := p.slot(index, "reasoning"); ok {
			p.out.thinkingDelta(i, "\n\n")
		}
	case "response.output_text.delta", "response.refusal.delta":
		if i, ok := p.slot(index, "message"); ok {
			p.out.textDelta(i, ev.Delta)
		}
	case "response.output_text.done", "response.refusal.done", "response.reasoning_text.done", "response.reasoning_summary_text.done":
		if p.kimi != nil {
			p.kimiResponsesTextDone(ev)
		}
		if p.glm {
			p.glmTextDone(ev)
		}
		if p.minimax != nil {
			p.miniMaxTextDone(ev)
		}
	case "response.function_call_arguments.delta":
		p.argumentsDelta(index, ev.Delta)
	case "response.function_call_arguments.done":
		p.argumentsDone(index, ev.Arguments)
	case "response.output_item.done":
		p.itemDone(index, ev.Item)
	case "response.completed", "response.incomplete":
		return p.finalize(ev.Response)
	case "error":
		if p.glm {
			return p.failures.providerCodeError(ev.Code)
		}
		return p.failures.upstream("Error Code " + jsText(ev.JSON.Code, ev.Code) + ": " + jsText(ev.JSON.Message, ev.Message))
	case "response.failed":
		return p.failed(ev.Response)
	}
	return nil
}

// failed ends the call at a response.failed terminal with pi's message. As in
// pi, the failed response's usage and id are not taken over, so its attempt
// stays unreported even when the response carried usage (maintainer
// decision, ADR-0010).
func (p *responsesParser) failed(r responses.Response) *Error {
	if p.minimax != nil {
		p.out.responseModel(r.Model)
	}
	p.terminal = true
	p.out.rawStopReason(string(r.Status))
	switch {
	case r.JSON.Error.Valid():
		if p.glm {
			return p.failures.providerCodeError(string(r.Error.Code))
		}
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
func (p *responsesParser) itemDone(outputIndex responsesIndex, item responses.ResponseOutputItemUnion) {
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
	if p.minimax != nil {
		p.minimax.finished[outputIndex] = s
	}
	if p.kimi != nil {
		p.kimi.finished[outputIndex] = s
	}
	if p.glm {
		if p.glmFinished == nil {
			p.glmFinished = map[responsesIndex]bool{}
		}
		p.glmFinished[outputIndex] = true
	}
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
		value := text.String()
		if (p.glm || p.minimax != nil) && len(item.Content) == 0 {
			value = p.out.textContent(i)
			if s.done != nil {
				value = *s.done
			}
		}
		p.out.textEnd(i, value, encodeTextSignature(item.ID, string(item.Phase)))
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
		if p.glm || p.minimax != nil {
			// Explicit content parts are authoritative even when their text
			// is empty. An empty part list still uses a separate text.done.
			if len(item.Content) > 0 {
				thinking = strings.Join(content, "\n\n")
			} else if s.done != nil {
				thinking = *s.done
			}
		}
		native := item.RawJSON()
		if p.glm && p.glmNative[outputIndex] != "" {
			native = p.glmNative[outputIndex]
		}
		signature := reasoningSignature(native)
		if p.minimax != nil || p.kimi != nil {
			signature = native
		}
		p.out.thinkingEnd(i, thinking, signature)
		p.reasoning[item.ID] = i
	case "function_call":
		p.out.toolCallEnd(i, cmp.Or(item.Arguments.OfString, s.raw.String()))
		delete(p.unfinished, i)
	}
}

// finalize records a completed or incomplete response. A status pi cannot map
// fails at once; a status that maps to an error fails once the stream ends.
func (p *responsesParser) finalize(r responses.Response) *Error {
	if p.minimax != nil {
		p.out.responseModel(r.Model)
	}
	p.terminal = true
	if p.kimi != nil {
		p.finishKimiOutput(r.Output)
	} else if p.minimax != nil {
		p.finishMiniMaxOutput(r.Output)
	} else if p.glm {
		p.finishGLMOutput(r.Output)
	} else {
		p.backfillReasoning(r.Output)
	}
	if r.ID != "" {
		p.out.responseID(r.ID)
	}
	if p.kimi != nil {
		if failure := p.reportKimiResponsesUsage(r); failure != nil {
			return failure
		}
	} else {
		u, reporting, ok := p.usage.of(r)
		if ok {
			p.out.usage(u)
		}
		p.attempt.usage(reporting, u)
	}
	status, reason := string(r.Status), string(r.IncompleteDetails.Reason)
	raw := status
	if reason != "" {
		raw += "." + reason
	}
	p.out.rawStopReason(raw)
	stop, msg, known := mapResponsesStatus(status, reason)
	p.stop = stop
	switch {
	case !known:
		return newError(CodeProtocol, PhaseStream, msg)
	case stop == StopReasonError:
		// pi-ai: throw new Error(output.errorMessage || "An unknown error occurred").
		p.failure = p.failures.upstream(cmp.Or(msg, "An unknown error occurred"))
	case stop == StopReasonStop && p.out.hasToolCall():
		// pi: a normally completed turn holding tool calls is tool use.
		p.stop = StopReasonToolUse
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

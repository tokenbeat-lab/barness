package ai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// responsesAdapter implements the OpenAI Responses protocol on the OpenAI Go
// SDK. It deliberately avoids openai.NewClient, whose defaults read
// OPENAI_API_KEY, OPENAI_BASE_URL, OPENAI_ORG_ID, OPENAI_PROJECT_ID and
// OPENAI_CUSTOM_HEADERS from the environment. A fresh, unauthenticated-until-
// now service is built per call from explicit options only, so no
// authenticated client outlives the call.
type responsesAdapter struct{}

func (responsesAdapter) simpleOptions(Model, SimpleOptions) Options { return ResponsesOptions{} }

func (responsesAdapter) stream(ctx context.Context, ac adapterCall, out *assembler) *Error {
	body, err := buildResponsesBody(ac.model, ac.request)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "request could not be encoded")
	}
	svc := responses.NewResponseService(
		option.WithHTTPClient(ac.http),
		option.WithBaseURL(ac.endpoint),
		option.WithAPIKey(ac.apiKey.reveal()),
		// Retries are barness-ai's decision (default 0, ticket 11), never the SDK's.
		option.WithMaxRetries(0),
		option.WithHeader("User-Agent", userAgent),
		option.WithRequestBody("application/json", body),
	)
	stream := svc.NewStreaming(ctx, responses.ResponseNewParams{})
	defer stream.Close()
	if err := stream.Err(); err != nil {
		return classifyResponsesError(ctx, err, PhaseRequest)
	}

	out.start()
	p := responsesParser{out: out, slots: map[int64]responsesSlot{}}
	for stream.Next() {
		if failure := p.handle(stream.Current()); failure != nil {
			return failure
		}
	}
	if err := stream.Err(); err != nil {
		return classifyResponsesError(ctx, err, PhaseStream)
	}
	if e := contextError(ctx, PhaseStream); e != nil {
		return e
	}
	// The SDK ends quietly at EOF; only a protocol terminal event is success.
	if !p.terminal {
		return newError(CodeProtocol, PhaseStream, "OpenAI Responses stream ended before a terminal response event")
	}
	if p.failure != "" {
		return newError(CodeProtocol, PhaseStream, p.failure)
	}
	return nil
}

// classifyResponsesError maps an SDK/transport error to a classified Error.
func classifyResponsesError(ctx context.Context, err error, phase Phase) *Error {
	if e := contextError(ctx, phase); e != nil {
		return e
	}
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		code := CodeProtocol
		switch apiErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			code = CodeUpstreamAuth
		case http.StatusTooManyRequests:
			code = CodeRateLimited
		}
		e := newError(code, phase, fmt.Sprintf("OpenAI API error: %d %s", apiErr.StatusCode, http.StatusText(apiErr.StatusCode)))
		e.HTTPStatus = apiErr.StatusCode
		return e
	}
	if phase == PhaseStream {
		return newError(CodeProtocol, phase, "OpenAI API error: "+err.Error())
	}
	return newError(CodeTransport, phase, "OpenAI API error: "+err.Error())
}

// responsesParser normalizes Responses stream events into assembler calls. It
// ports the text and reasoning paths of pi-ai's processResponsesStream; tool
// items arrive with ticket 06, and the error/response.failed terminals with
// ticket 05 (until then such a stream fails as having no terminal event).
// Unknown events are ignored.
type responsesParser struct {
	out      *assembler
	slots    map[int64]responsesSlot // output_index → open block
	terminal bool
	failure  string // set when the terminal status is a failure
}

// responsesSlot is the open content block an output item streams into.
type responsesSlot struct {
	kind  string // the item type: "message" or "reasoning"
	index int    // content index in the assistant message
}

// slot returns the open block for outputIndex if it has the given kind.
func (p *responsesParser) slot(outputIndex int64, kind string) (int, bool) {
	s, ok := p.slots[outputIndex]
	return s.index, ok && s.kind == kind
}

// open starts a block for a message or reasoning item; other items are not
// handled yet.
func (p *responsesParser) open(outputIndex int64, kind string) (int, bool) {
	var i int
	switch kind {
	case "message":
		i = p.out.textStart()
	case "reasoning":
		i = p.out.thinkingStart()
	default:
		return 0, false
	}
	p.slots[outputIndex] = responsesSlot{kind: kind, index: i}
	return i, true
}

func (p *responsesParser) handle(ev responses.ResponseStreamEventUnion) *Error {
	switch ev.Type {
	case "response.created":
		p.out.responseID(ev.Response.ID)
	case "response.output_item.added":
		p.open(ev.OutputIndex, ev.Item.Type)
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
	case "response.output_item.done":
		p.itemDone(ev.OutputIndex, ev.Item)
	case "response.completed", "response.incomplete":
		p.finalize(ev.Response)
	}
	return nil
}

// itemDone closes the item's block with the item's authoritative content. As
// pi's getOrCreateSlot, it opens the block first when the item was never
// announced, and ignores the item when its output index holds a block of
// another kind.
func (p *responsesParser) itemDone(outputIndex int64, item responses.ResponseOutputItemUnion) {
	var i int
	if s, ok := p.slots[outputIndex]; ok {
		if s.kind != item.Type {
			return
		}
		i = s.index
	} else if i, ok = p.open(outputIndex, item.Type); !ok {
		return
	}
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
	}
}

func (p *responsesParser) finalize(r responses.Response) {
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
	stop, msg := mapResponsesStatus(string(r.Status), string(r.IncompleteDetails.Reason))
	if stop == StopReasonError {
		// pi-ai: throw new Error(output.errorMessage || "An unknown error occurred").
		p.failure = cmp.Or(msg, "An unknown error occurred")
		return
	}
	p.out.stop(stop)
}

// mapResponsesStatus ports pi-ai's mapStopReason.
func mapResponsesStatus(status, incompleteReason string) (StopReason, string) {
	switch status {
	case "", "completed", "in_progress", "queued":
		return StopReasonStop, ""
	case "incomplete":
		if incompleteReason == "max_output_tokens" {
			return StopReasonLength, ""
		}
		if incompleteReason != "" {
			return StopReasonError, "Response incomplete: " + incompleteReason
		}
		return StopReasonError, "Response incomplete without a provider reason"
	case "failed", "cancelled":
		return StopReasonError, ""
	default:
		return StopReasonError, "Unhandled stop reason: " + status
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

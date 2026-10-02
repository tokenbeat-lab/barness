package ai

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strconv"
)

// geminiChunk is the part of a streamed GenerateContentResponse pi reads:
// the response id, the first candidate's parts and finish reason, and the
// usage.
type geminiChunk struct {
	ResponseID string `json:"responseId"`
	Candidates []struct {
		Content *struct {
			Parts []geminiWirePart `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata *geminiWireUsage `json:"usageMetadata"`
}

// geminiWirePart is one response part. Fields pi tests by JavaScript
// truthiness or strict equality stay raw, so a value of an unexpected type
// is ignored as pi ignores it instead of failing the chunk.
type geminiWirePart struct {
	Text             *string         `json:"text"`
	Thought          json.RawMessage `json:"thought"`
	ThoughtSignature json.RawMessage `json:"thoughtSignature"`
	FunctionCall     *struct {
		ID   json.RawMessage `json:"id"`
		Name json.RawMessage `json:"name"`
		Args json.RawMessage `json:"args"`
	} `json:"functionCall"`
}

// thinking is pi's isThinkingPart: thought is exactly true. A thought
// signature alone does not make a part thinking.
func (p geminiWirePart) thinking() bool { return string(p.Thought) == "true" }

// signature is the part's thought signature when it is a non-empty string.
func (p geminiWirePart) signature() string { return jsonString(p.ThoughtSignature) }

// jsonString is raw's value when it is a JSON string, else "".
func jsonString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}
	return s
}

// geminiWireUsage is usageMetadata; a nil count was missing or null.
type geminiWireUsage struct {
	PromptTokenCount        *int64 `json:"promptTokenCount"`
	CandidatesTokenCount    *int64 `json:"candidatesTokenCount"`
	ThoughtsTokenCount      *int64 `json:"thoughtsTokenCount"`
	CachedContentTokenCount *int64 `json:"cachedContentTokenCount"`
	TotalTokenCount         *int64 `json:"totalTokenCount"`
}

// Error texts barness-ai shares with pi-ai's Google path.
const (
	// pi's check after the stream once its signal was aborted.
	msgGeminiAborted  = "Request was aborted"
	msgGeminiNoFinish = "Google stream ended without a finish reason"
)

// Error texts pi takes from its JavaScript runtime (JSON.parse, undici),
// which barness cannot reproduce; recorded as approved differences in the pi
// ledger.
const (
	msgGeminiConnectionLost = "Connection lost while reading the Gemini stream"
	msgGeminiInvalidJSON    = "Could not parse Gemini stream chunk: invalid JSON"
)

// geminiBlock is the open text or thinking block pi calls currentBlock.
type geminiBlock struct {
	index     int
	kind      blockKind
	text      string
	signature string
}

// geminiParser normalizes Gemini stream chunks into assembler calls. It
// ports the stream loop of pi-ai's google-generative-ai: consecutive text
// parts of one kind share a block, a function call closes the open block
// and arrives whole, and only a finish reason ends the call successfully.
type geminiParser struct {
	out      *assembler
	failures httpFailures
	model    Model
	attempt  *initialRequest
	open     *geminiBlock
	// stop is the mapped finish reason of the last chunk that had one; ""
	// while none arrived. finishReason is that finish reason as sent.
	stop         StopReason
	finishReason string
	// identified is set once the response id was taken.
	identified bool
	usage      geminiUsage
	// generated counts the tool call ids made up for calls without one.
	generated int
}

// read feeds the stream to p until its end, a failure or an interruption.
func (p *geminiParser) read(ctx context.Context, dec *geminiDecoder) *Error {
	for dec.next() {
		if failure := p.handle(dec.chunk()); failure != nil {
			return failure
		}
		if failure := p.out.failure(); failure != nil {
			return failure
		}
	}
	if ctx.Err() != nil {
		return interrupted(ctx, PhaseStream, msgGeminiAborted)
	}
	if dec.err != nil {
		var sdk *geminiStreamError
		if errors.As(dec.err, &sdk) {
			return sdk.failure
		}
		return bodyReadFailure(dec.err, msgGeminiConnectionLost)
	}
	// As in pi, the open block ends before the terminal is judged.
	p.closeBlock()
	switch p.stop {
	case "":
		return newError(CodeProtocol, PhaseStream, msgGeminiNoFinish)
	case StopReasonError:
		return p.failures.upstream("Provider stopped with: " + p.finishReason)
	}
	return nil
}

func (p *geminiParser) handle(data []byte) *Error {
	var c geminiChunk
	if json.Unmarshal(data, &c) != nil {
		return newError(CodeProtocol, PhaseStream, msgGeminiInvalidJSON)
	}
	// pi keeps the first non-empty response id.
	if c.ResponseID != "" && !p.identified {
		p.identified = true
		p.out.responseID(c.ResponseID)
	}
	if len(c.Candidates) == 0 {
		p.takeUsage(c.UsageMetadata)
		return nil
	}
	cand := c.Candidates[0]
	if cand.Content != nil {
		for _, part := range cand.Content.Parts {
			if part.Text != nil {
				p.text(part)
			}
			if part.FunctionCall != nil {
				p.functionCall(part)
			}
		}
	}
	if cand.FinishReason != "" {
		p.finishReason = cand.FinishReason
		p.out.rawStopReason(cand.FinishReason)
		stop, known := mapGeminiFinishReason(cand.FinishReason)
		if !known {
			return newError(CodeProtocol, PhaseStream, "Unhandled stop reason: "+cand.FinishReason)
		}
		if stop == StopReasonStop && p.out.hasToolCall() {
			stop = StopReasonToolUse
		}
		p.stop = stop
		if stop != StopReasonError {
			p.out.stop(stop)
		}
	}
	p.takeUsage(c.UsageMetadata)
	return nil
}

// text appends a text part to the open block of its kind, closing a block of
// the other kind first. A signature arriving on a later part replaces the
// block's; an empty or missing one keeps it.
func (p *geminiParser) text(part geminiWirePart) {
	kind := blockText
	if part.thinking() {
		kind = blockThinking
	}
	if p.open == nil || p.open.kind != kind {
		p.closeBlock()
		b := &geminiBlock{kind: kind}
		if kind == blockThinking {
			b.index = p.out.thinkingStart(Thinking{})
		} else {
			b.index = p.out.textStart(Text{})
		}
		p.open = b
	}
	b := p.open
	b.text += *part.Text
	if sig := part.signature(); sig != "" {
		b.signature = sig
	}
	if kind == blockThinking {
		p.out.setThinkingSignature(b.index, b.signature)
		p.out.thinkingDelta(b.index, *part.Text)
	} else {
		p.out.setTextSignature(b.index, b.signature)
		p.out.textDelta(b.index, *part.Text)
	}
}

// closeBlock ends the open text or thinking block, if any.
func (p *geminiParser) closeBlock() {
	b := p.open
	if b == nil {
		return
	}
	p.open = nil
	if b.kind == blockThinking {
		p.out.thinkingEnd(b.index, b.text, b.signature)
	} else {
		p.out.textEnd(b.index, b.text, b.signature)
	}
}

// functionCall reports a whole function call: start, its arguments as one
// delta (pi's JSON.stringify of them), end. A call without an id, or
// repeating one already in the message, gets pi's generated
// "<name>_<epoch ms>_<n>"; n counts within the call where pi counts within
// its process, so with the millisecond the id stays unique in a
// conversation. The raw arguments are the provider's JSON text.
func (p *geminiParser) functionCall(part geminiWirePart) {
	p.closeBlock()
	fc := part.FunctionCall
	name := jsonString(fc.Name)
	id := jsonString(fc.ID)
	if id == "" || p.hasToolCallID(id) {
		p.generated++
		// pi interpolates the name as sent: a missing one is "undefined".
		id = jsTemplateString(fc.Name) + "_" + strconv.FormatInt(p.attempt.clock.Now().UnixMilli(), 10) + "_" + strconv.Itoa(p.generated)
	}
	raw := string(fc.Args)
	if raw == "null" {
		raw = ""
	}
	delta := "{}"
	if raw != "" {
		if v, err := parseJSON(raw); err == nil {
			delta = stringifyJSON(v)
		}
	}
	i := p.out.toolCallStart(id, name, "")
	if sig := part.signature(); sig != "" {
		p.out.setToolCallSignature(i, sig)
	}
	p.out.toolCallDelta(i, delta, raw)
	p.out.toolCallEnd(i, raw)
}

func (p *geminiParser) hasToolCallID(id string) (found bool) {
	p.out.view.read(func(m *AssistantMessage) {
		found = slices.ContainsFunc(m.Content, func(c AssistantContent) bool {
			call, ok := c.(ToolCall)
			return ok && call.ID == id
		})
	})
	return found
}

// mapGeminiFinishReason ports pi's mapStopReason over the Google SDK's
// FinishReason values; known is false for a value pi throws on.
func mapGeminiFinishReason(reason string) (stop StopReason, known bool) {
	switch reason {
	case "STOP":
		return StopReasonStop, true
	case "MAX_TOKENS":
		return StopReasonLength, true
	case "BLOCKLIST", "PROHIBITED_CONTENT", "SPII", "SAFETY", "IMAGE_SAFETY", "IMAGE_PROHIBITED_CONTENT",
		"IMAGE_RECITATION", "IMAGE_OTHER", "RECITATION", "FINISH_REASON_UNSPECIFIED", "OTHER", "LANGUAGE",
		"MALFORMED_FUNCTION_CALL", "UNEXPECTED_TOOL_CALL", "TOO_MANY_TOOL_CALLS", "NO_IMAGE":
		return StopReasonError, true
	}
	return StopReasonError, false
}

// takeUsage replaces the usage with a chunk's usageMetadata, as pi does, and
// publishes it on the message and the attempt being read (ADR-0010).
func (p *geminiParser) takeUsage(w *geminiWireUsage) {
	if w == nil {
		return
	}
	p.usage.take(p.model, w)
	p.out.usage(p.usage.u)
	p.attempt.usage(p.usage.reporting, p.usage.u)
}

// geminiUsage is one stream's usage. Each usageMetadata replaces it whole,
// missing or null counts as 0: input excludes cached tokens, output includes
// thinking tokens (reported again as the reasoning subset), and the total is
// the provider's. Gemini reports no cache writes.
type geminiUsage struct {
	u         Usage
	reporting UsageReporting
}

func (g *geminiUsage) take(m Model, w *geminiWireUsage) {
	n := func(c *int64) int64 {
		if c == nil {
			return 0
		}
		return *c
	}
	g.u = Usage{
		Input:       n(w.PromptTokenCount) - n(w.CachedContentTokenCount),
		Output:      n(w.CandidatesTokenCount) + n(w.ThoughtsTokenCount),
		CacheRead:   n(w.CachedContentTokenCount),
		Reasoning:   Value(n(w.ThoughtsTokenCount)),
		TotalTokens: n(w.TotalTokenCount),
	}
	g.u.Cost = m.Cost.estimate(g.u)
	// Complete when the prompt, candidates and total counts were reported
	// (0 included). Cached and thinking counts are sent only when there are
	// some, so their absence reads as none.
	g.reporting = UsagePartial
	if w.PromptTokenCount != nil && w.CandidatesTokenCount != nil && w.TotalTokenCount != nil {
		g.reporting = UsageComplete
	}
}

func (geminiAdapter) historyRules(m Model) historyRules {
	rules := historyRules{}
	if geminiRequiresToolCallID(m.ID) {
		rules.normalizeToolCallID = func(id string, _ AssistantMessage) string { return sanitizeToolCallID(id) }
	}
	return rules
}

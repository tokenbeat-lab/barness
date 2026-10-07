package ai

import (
	"context"
	"encoding/json"
	"errors"
)

// anthropicMessageEvents are the SSE event names pi handles; it skips every
// other name (ping, proxy events, …) without parsing its data.
var anthropicMessageEvents = map[string]bool{
	"message_start": true, "message_delta": true, "message_stop": true,
	"content_block_start": true, "content_block_delta": true, "content_block_stop": true,
}

// anthropicEvent is the union of the stream events' fields pi reads.
type anthropicEvent struct {
	Type    string `json:"type"`
	Index   int64  `json:"index"`
	Message *struct {
		ID    string              `json:"id"`
		Model *string             `json:"model"`
		Usage *anthropicWireUsage `json:"usage"`
	} `json:"message"`
	ContentBlock *struct {
		Type      string `json:"type"`
		Text      string `json:"text"`
		Thinking  string `json:"thinking"`
		Signature string `json:"signature"`
		Data      string `json:"data"`
		ID        string `json:"id"`
		Name      string `json:"name"`
	} `json:"content_block"`
	Delta *struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		Thinking    string `json:"thinking"`
		PartialJSON string `json:"partial_json"`
		Signature   string `json:"signature"`
		StopReason  string `json:"stop_reason"`
		StopDetails *struct {
			Explanation string `json:"explanation"`
		} `json:"stop_details"`
	} `json:"delta"`
	Usage *anthropicWireUsage `json:"usage"`
}

// anthropicWireUsage is a usage object; a nil count was missing or null.
type anthropicWireUsage struct {
	InputTokens   *int64 `json:"input_tokens"`
	OutputTokens  *int64 `json:"output_tokens"`
	CacheRead     *int64 `json:"cache_read_input_tokens"`
	CacheCreation *int64 `json:"cache_creation_input_tokens"`
	CacheDetails  *struct {
		Ephemeral1h *int64 `json:"ephemeral_1h_input_tokens"`
	} `json:"cache_creation"`
	OutputDetails *struct {
		ThinkingTokens *int64 `json:"thinking_tokens"`
	} `json:"output_tokens_details"`
}

// decodeAnthropicEvent is pi's parseJsonWithRepair: invalid JSON is parsed
// again after repairing its string literals.
func decodeAnthropicEvent(data string, e *anthropicEvent) error {
	err := json.Unmarshal([]byte(data), e)
	var syntax *json.SyntaxError
	if errors.As(err, &syntax) {
		if repaired := RepairToolJSON(data); repaired != data {
			*e = anthropicEvent{}
			return json.Unmarshal([]byte(repaired), e)
		}
	}
	return err
}

// anthropicSlot is an open content block: its content index and kind.
type anthropicSlot struct {
	index int
	kind  blockKind
}

// blockKind is the kind of content block an Anthropic stream block opened.
type blockKind int

const (
	blockText blockKind = iota
	blockThinking
	blockToolCall
)

// anthropicParser normalizes Anthropic stream events into assembler calls. It
// ports the event handling of pi-ai's anthropic-messages stream for an API
// key. Blocks of other types (server tool use and results, …) and unknown
// deltas are ignored, as in pi.
type anthropicParser struct {
	out      *assembler
	failures httpFailures
	model    Model
	// open maps a stream block index to its open blocks, oldest first: like
	// pi's search for the first block with that index, a repeated index
	// addresses the older block until it stops.
	open map[int64][]anthropicSlot
	// raw is each tool call's argument JSON received so far (pi's partialJson).
	raw     map[int]*argumentText
	usage   anthropicUsage
	attempt *initialRequest
	started bool // message_start was received
	stopped bool // message_stop was received
	// stop is the mapped stop reason of the last message_delta that had one;
	// "" while none arrived. stopMessage is an error stop reason's message.
	stop        StopReason
	stopMessage string
}

// read feeds the stream to p until its end, a failure, or an interruption.
// As in pi, only a stop reason makes it a success, and a stream that started
// a message must also stop it.
func (p *anthropicParser) read(ctx context.Context, dec *sseDecoder) *Error {
	for dec.next() {
		if failure := p.handle(dec.event()); failure != nil {
			return failure
		}
		if failure := p.out.failure(); failure != nil {
			return failure
		}
	}
	if ctx.Err() != nil {
		return interrupted(ctx, PhaseStream, msgAnthropicAborted)
	}
	if dec.err != nil {
		return bodyReadFailure(dec.err, msgAnthropicConnectionLost)
	}
	switch {
	case p.started && !p.stopped:
		return newError(CodeProtocol, PhaseStream, msgAnthropicNoMessageStop)
	case p.stop == "":
		return newError(CodeProtocol, PhaseStream, msgAnthropicNoStopReason)
	case p.stop == StopReasonError:
		return p.failures.upstream(p.stopMessage)
	}
	return nil
}

func (p *anthropicParser) handle(ev sseEvent) *Error {
	// pi fails on an error event whatever came before, its data as the
	// message.
	if ev.name == "error" {
		return p.failures.upstream(ev.data)
	}
	if !anthropicMessageEvents[ev.name] {
		return nil
	}
	var e anthropicEvent
	if err := decodeAnthropicEvent(ev.data, &e); err != nil {
		return newError(CodeProtocol, PhaseStream, msgAnthropicInvalidJSON(ev.name))
	}
	switch e.Type {
	case "message_start":
		p.started = true
		if e.Message != nil {
			p.messageStart(e)
		}
	case "content_block_start":
		if e.ContentBlock != nil {
			return p.blockStart(e)
		}
	case "content_block_delta":
		if e.Delta != nil {
			p.blockDelta(e)
		}
	case "content_block_stop":
		p.blockStop(e.Index)
	case "message_delta":
		return p.messageDelta(e)
	case "message_stop":
		p.stopped = true
	}
	return nil
}

func (p *anthropicParser) messageStart(e anthropicEvent) {
	p.out.responseID(e.Message.ID)
	if m := e.Message.Model; m != nil && *m != p.model.ID {
		p.out.responseModel(*m)
	}
	// pi takes the initial usage here, so an interrupted stream still
	// reports its input.
	p.usage.start(e.Message.Usage)
	p.reportUsage()
}

func (p *anthropicParser) blockStart(e anthropicEvent) *Error {
	b := e.ContentBlock
	var slot anthropicSlot
	switch b.Type {
	case "fallback":
		// Only a fallback before any output can be served by the model it
		// names; barness-ai requests none (see BuiltinCatalog).
		if p.out.blockCount() > 0 {
			return newError(CodeProtocol, PhaseStream, msgAnthropicFallback)
		}
		return nil
	case "text":
		slot = anthropicSlot{index: p.out.textStart(Text{Text: b.Text}), kind: blockText}
	case "thinking":
		slot = anthropicSlot{index: p.out.thinkingStart(Thinking{Thinking: b.Thinking, Signature: b.Signature}), kind: blockThinking}
	case "redacted_thinking":
		slot = anthropicSlot{index: p.out.thinkingStart(Thinking{Thinking: "[Reasoning redacted]", Signature: b.Data, Redacted: true}), kind: blockThinking}
	case "tool_use":
		slot = anthropicSlot{index: p.out.toolCallStart(b.ID, b.Name, ""), kind: blockToolCall}
		if p.raw == nil {
			p.raw = map[int]*argumentText{}
		}
		p.raw[slot.index] = newArgumentText("")
	default:
		return nil
	}
	p.open[e.Index] = append(p.open[e.Index], slot)
	return nil
}

// slot is the oldest open block with the stream index, if it has kind.
func (p *anthropicParser) slot(index int64, kind blockKind) (int, bool) {
	slots := p.open[index]
	if len(slots) == 0 || slots[0].kind != kind {
		return 0, false
	}
	return slots[0].index, true
}

func (p *anthropicParser) blockDelta(e anthropicEvent) {
	d := e.Delta
	switch d.Type {
	case "text_delta":
		if i, ok := p.slot(e.Index, blockText); ok {
			p.out.textDelta(i, d.Text)
		}
	case "thinking_delta":
		if i, ok := p.slot(e.Index, blockThinking); ok {
			p.out.thinkingDelta(i, d.Thinking)
		}
	case "input_json_delta":
		if i, ok := p.slot(e.Index, blockToolCall); ok {
			raw := p.raw[i]
			if !p.out.fitsToolJSON(raw.Len() + len(d.PartialJSON)) {
				return
			}
			p.out.toolCallDelta(i, d.PartialJSON, raw.append(d.PartialJSON))
		}
	case "signature_delta":
		// The signature grows without an event, as in pi.
		if i, ok := p.slot(e.Index, blockThinking); ok {
			p.out.setThinkingSignature(i, p.out.thinkingSignature(i)+d.Signature)
		}
	}
}

// blockStop closes the oldest open block with the stream index.
func (p *anthropicParser) blockStop(index int64) {
	slots := p.open[index]
	if len(slots) == 0 {
		return
	}
	s := slots[0]
	if p.open[index] = slots[1:]; len(p.open[index]) == 0 {
		delete(p.open, index)
	}
	switch s.kind {
	case blockText:
		p.out.textEnd(s.index, p.out.textContent(s.index), "")
	case blockThinking:
		p.out.thinkingEnd(s.index, p.out.thinkingText(s.index), p.out.thinkingSignature(s.index))
	case blockToolCall:
		p.out.toolCallEnd(s.index, p.raw[s.index].String())
	}
}

// messageDelta records a stop reason and the usage counts it carries. An
// unknown stop reason fails at once, as pi throws on it; a stop reason that
// maps to an error fails once the stream ends.
func (p *anthropicParser) messageDelta(e anthropicEvent) *Error {
	if d := e.Delta; d != nil && d.StopReason != "" {
		p.out.rawStopReason(d.StopReason)
		explanation := ""
		if d.StopDetails != nil {
			explanation = d.StopDetails.Explanation
		}
		stop, msg, known := mapAnthropicStopReason(d.StopReason, explanation)
		if !known {
			return newError(CodeProtocol, PhaseStream, msg)
		}
		p.stop, p.stopMessage = stop, msg
		if stop != StopReasonError {
			p.out.stop(stop)
		}
	}
	p.usage.delta(e.Usage)
	p.reportUsage()
	return nil
}

// mapAnthropicStopReason ports pi-ai's mapStopReason; known is false for a
// stop reason pi throws on.
func mapAnthropicStopReason(reason, explanation string) (stop StopReason, msg string, known bool) {
	switch reason {
	case "end_turn", "pause_turn", "stop_sequence":
		return StopReasonStop, "", true
	case "max_tokens":
		return StopReasonLength, "", true
	case "tool_use":
		return StopReasonToolUse, "", true
	case "refusal":
		if explanation == "" {
			explanation = msgAnthropicRefusalDefault
		}
		return StopReasonError, explanation, true
	case "sensitive":
		return StopReasonError, msgAnthropicSensitive, true
	}
	return StopReasonError, "Unhandled stop reason: " + reason, false
}

// reportUsage publishes the usage so far on the message and on the attempt
// being read, which keep the same value (ADR-0010).
func (p *anthropicParser) reportUsage() {
	p.out.usage(p.usage.u)
	p.attempt.usage(p.usage.reporting(), p.usage.u)
}

// anthropicUsage accumulates one stream's usage as pi does: message_start
// sets every count (missing or null as 0), each message_delta replaces the
// counts it reports. The total is computed from the parts, which Anthropic
// does not report, and the cost recomputed at the model's rates.
type anthropicUsage struct {
	model Model
	u     Usage
	// seen records which of input, output, cache read and cache write
	// tokens were reported at least once; reported whether any usage was.
	seen     [4]bool
	reported bool
}

func (a *anthropicUsage) start(w *anthropicWireUsage) {
	a.u.Input, a.u.Output, a.u.CacheRead, a.u.CacheWrite = 0, 0, 0, 0
	var long int64
	if w != nil {
		a.reported = true
		a.take(w)
		if w.CacheDetails != nil && w.CacheDetails.Ephemeral1h != nil {
			long = *w.CacheDetails.Ephemeral1h
		}
	}
	a.u.CacheWrite1h = Value(long)
	a.settle()
}

func (a *anthropicUsage) delta(w *anthropicWireUsage) {
	if w != nil {
		a.reported = true
		a.take(w)
		// Gateway deltas may supply or correct the TTL breakdown. This is
		// an authoritative count, not an additional cache write (pi 1.0).
		if w.CacheDetails != nil && w.CacheDetails.Ephemeral1h != nil {
			a.u.CacheWrite1h = Value(*w.CacheDetails.Ephemeral1h)
		}
		// Anthropic reports reasoning tokens as a subset of output tokens.
		if w.OutputDetails != nil && w.OutputDetails.ThinkingTokens != nil {
			a.u.Reasoning = Value(*w.OutputDetails.ThinkingTokens)
		}
	}
	a.settle()
}

// take replaces each count w reports and marks it seen.
func (a *anthropicUsage) take(w *anthropicWireUsage) {
	counts := []*int64{w.InputTokens, w.OutputTokens, w.CacheRead, w.CacheCreation}
	targets := []*int64{&a.u.Input, &a.u.Output, &a.u.CacheRead, &a.u.CacheWrite}
	for i, c := range counts {
		if c != nil {
			*targets[i], a.seen[i] = *c, true
		}
	}
}

func (a *anthropicUsage) settle() {
	a.u.TotalTokens = a.u.Input + a.u.Output + a.u.CacheRead + a.u.CacheWrite
	a.u.Cost = a.model.Cost.estimate(a.u)
}

// reporting is complete when input, output, cache read and cache write
// tokens were each reported (0 included), partial when usage was reported
// without some of them, and unreported when no usage arrived. The one-hour
// cache write and thinking token details joined the protocol later and are
// not required: their absence reads as none.
func (a *anthropicUsage) reporting() UsageReporting {
	switch {
	case !a.reported:
		return UsageUnreported
	case a.seen == [4]bool{true, true, true, true}:
		return UsageComplete
	}
	return UsagePartial
}

func (anthropicAdapter) historyRules(m Model) historyRules {
	return historyRules{
		midConvoSystem: m.Compat.SupportsMidConvoSystemMessages,
		normalizeToolCallID: func(id string, _ AssistantMessage) string {
			return sanitizeToolCallID(id)
		},
	}
}

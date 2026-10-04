package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"

	"github.com/openai/openai-go/v3/packages/ssestream"
)

// Error texts barness-ai shares with pi-ai's Chat Completions path.
const (
	// pi's check after the stream once its signal was aborted; openai-node
	// ends an aborted stream quietly, so it is the text of every
	// cancellation while reading.
	msgChatAborted           = "Request was aborted"
	msgChatNoFinishReason    = "Stream ended without finish_reason"
	msgChatErrorStopFallback = "Provider returned an error stop reason"
)

// Error texts pi takes from its JavaScript runtime (JSON.parse, undici),
// which barness cannot reproduce; recorded as approved differences in the
// pi ledger.
const (
	msgChatInvalidJSON    = "OpenAI Chat Completions stream event is not valid JSON"
	msgChatConnectionLost = "Connection lost while reading the OpenAI Chat Completions stream"
)

// chatParser reads a Chat Completions stream into assembler calls. It ports
// the chunk loop of pi-ai's openai-completions stream: one text block and
// one thinking block per turn, whatever interleaves them, and a tool call
// block per call, matched by its stream index or id. Every block is closed
// only once the stream ended, in content order; a failure while reading
// leaves them open, as pi does.
type chatParser struct {
	out      *assembler
	model    Model
	failures chatFailures
	// attempt records the usage of the attempt being read.
	attempt *initialRequest

	text, thinking int // content index of the turn's block; -1 before it opens
	calls          []chatCall
	// byIndex and byID find the call a delta continues, as pi's maps do:
	// a call registers its first stream index, and every id a delta of it
	// carries, the latest registration winning.
	byIndex map[float64]int
	byID    map[string]int
	details chatDetails // streamed reasoning_details, nil before any
	// responseID and responseModel were taken from a chunk: pi keeps the
	// first truthy ones.
	responseID, responseModel bool
	finished                  bool       // a finish_reason arrived
	stop                      StopReason // the last finish_reason's stop reason
	stopMessage               string     // the last error finish_reason's message
	usage                     chatUsage
}

// chatCall is one streamed tool call: its block, the argument text so far
// (pi's partialArgs) and whether it was given a stream index yet.
type chatCall struct {
	block   int
	raw     *argumentText
	indexed bool
}

func newChatParser(out *assembler, model Model, usage chatUsage, failures chatFailures, attempt *initialRequest) *chatParser {
	return &chatParser{out: out, model: model, failures: failures, attempt: attempt, text: -1, thinking: -1,
		byIndex: map[float64]int{}, byID: map[string]int{}, usage: usage}
}

// read feeds the SSE events of dec to the parser until the body ends, and
// settles the turn as pi does. Like openai-node it stops at data starting
// with "[DONE]" but reads the body to its end, parses every other event,
// named or not, and fails at a chunk with a truthy error field.
func (p *chatParser) read(ctx context.Context, dec ssestream.Decoder) *Error {
	done := false
	for dec.Next() {
		if done {
			continue
		}
		data := dec.Event().Data
		if bytes.HasPrefix(data, []byte("[DONE]")) {
			done = true
			continue
		}
		if failure := p.event(dec.Event().Type, data); failure != nil {
			return p.fail(failure)
		}
		if failure := p.out.failure(); failure != nil {
			return p.fail(failure)
		}
	}
	if ctx.Err() != nil {
		// openai-node ends an aborted stream quietly; pi closes the blocks
		// and then reports the abort.
		p.closeBlocks()
		return interrupted(ctx, PhaseStream, msgChatAborted)
	}
	if err := dec.Err(); err != nil {
		return p.fail(bodyReadFailure(err, msgChatConnectionLost))
	}
	p.closeBlocks()
	return p.settle()
}

// fail is pi's catch path: the thinking block takes the streamed details
// as its signature, without an end event.
func (p *chatParser) fail(failure *Error) *Error {
	if p.thinking >= 0 && p.details != nil {
		p.out.setThinkingSignature(p.thinking, p.details.signature())
	}
	return failure
}

// settle decides the terminal once the stream ended cleanly: an error
// finish reason fails, a missing one is a protocol error.
func (p *chatParser) settle() *Error {
	switch {
	case p.stop == StopReasonError:
		msg := p.stopMessage
		if msg == "" {
			msg = msgChatErrorStopFallback
		}
		return p.failures.upstream(msg)
	case !p.finished:
		return newError(CodeProtocol, PhaseStream, msgChatNoFinishReason)
	}
	p.out.stop(p.stop)
	return nil
}

// chatChunk is the part of a chunk pi reads. Fields stay raw: pi tests
// them by JavaScript type and truthiness.
type chatChunk struct {
	Error   json.RawMessage `json:"error"`
	ID      json.RawMessage `json:"id"`
	Model   json.RawMessage `json:"model"`
	Usage   json.RawMessage `json:"usage"`
	Choices json.RawMessage `json:"choices"`
}

type chatChoice struct {
	Delta        json.RawMessage `json:"delta"`
	FinishReason json.RawMessage `json:"finish_reason"`
	// Usage is where some providers (Moonshot) report usage instead.
	Usage json.RawMessage `json:"usage"`
}

type chatDelta struct {
	Content          json.RawMessage `json:"content"`
	ReasoningContent json.RawMessage `json:"reasoning_content"`
	Reasoning        json.RawMessage `json:"reasoning"`
	ReasoningText    json.RawMessage `json:"reasoning_text"`
	ToolCalls        json.RawMessage `json:"tool_calls"`
	ReasoningDetails json.RawMessage `json:"reasoning_details"`
}

type chatToolCallDelta struct {
	Index    json.RawMessage `json:"index"`
	ID       json.RawMessage `json:"id"`
	Function *struct {
		Name      json.RawMessage `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
	Custom json.RawMessage `json:"custom"`
}

// event handles one SSE event's data.
func (p *chatParser) event(eventType string, data []byte) *Error {
	if !json.Valid(data) {
		return newError(CodeProtocol, PhaseStream, msgChatInvalidJSON)
	}
	var c chatChunk
	if json.Unmarshal(data, &c) != nil {
		// Not an object: pi skips it.
		return nil
	}
	// openai-node fails at any chunk with a truthy error, but passes the
	// events of OpenAI's Assistants stream (named thread.*) on wrapped,
	// where pi finds nothing to read.
	if strings.HasPrefix(eventType, "thread.") {
		return nil
	}
	if rawTruthy(c.Error) {
		return p.failures.inBand(c.Error)
	}
	return p.chunk(c)
}

func (p *chatParser) chunk(c chatChunk) *Error {
	if id, _ := rawString(c.ID); id != "" && !p.responseID {
		p.responseID = true
		p.out.responseID(id)
	}
	if model, _ := rawString(c.Model); model != "" && model != p.model.ID && !p.responseModel {
		p.responseModel = true
		p.out.responseModel(model)
	}
	if rawTruthy(c.Usage) {
		p.report(c.Usage)
	}
	var choices []json.RawMessage
	if json.Unmarshal(c.Choices, &choices) != nil || len(choices) == 0 || !rawTruthy(choices[0]) {
		return nil
	}
	var choice chatChoice
	if json.Unmarshal(choices[0], &choice) != nil {
		return nil
	}
	if !rawTruthy(c.Usage) && rawTruthy(choice.Usage) {
		p.report(choice.Usage)
	}
	if rawTruthy(choice.FinishReason) {
		p.finishReason(choice.FinishReason)
	}
	var delta chatDelta
	if !rawTruthy(choice.Delta) || json.Unmarshal(choice.Delta, &delta) != nil {
		return nil
	}
	if content, _ := rawString(delta.Content); content != "" {
		if p.text < 0 {
			p.text = p.out.textStart(Text{})
		}
		p.out.textDelta(p.text, content)
	}
	p.reasoning(delta)
	p.toolCalls(delta.ToolCalls)
	p.reasoningDetails(delta.ReasoningDetails)
	return nil
}

// finishReason records pi's mapStopReason of a truthy finish_reason. As in
// pi the last one wins, and an error message stays once set.
func (p *chatParser) finishReason(raw json.RawMessage) {
	reason, isString := rawString(raw)
	if !isString {
		reason = compactJSON(raw)
	}
	p.out.rawStopReason(reason)
	p.finished = true
	switch reason {
	case "stop", "end":
		p.stop = StopReasonStop
	case "length":
		p.stop = StopReasonLength
	case "function_call", "tool_calls":
		p.stop = StopReasonToolUse
	default:
		p.stop = StopReasonError
		p.stopMessage = "Provider finish_reason: " + reason
	}
}

// reasoning appends the first non-empty reasoning field to the thinking
// block, which remembers the field's name as its signature.
func (p *chatParser) reasoning(d chatDelta) {
	for i, raw := range []json.RawMessage{d.ReasoningContent, d.Reasoning, d.ReasoningText} {
		delta, _ := rawString(raw)
		if delta == "" {
			continue
		}
		if p.thinking < 0 {
			p.thinking = p.out.thinkingStart(Thinking{Signature: chatReasoningFields[i]})
		}
		p.out.thinkingDelta(p.thinking, delta)
		return
	}
}

// reasoningDetails keeps the valid entries of a reasoning_details delta;
// they open the thinking block without text.
func (p *chatParser) reasoningDetails(raw json.RawMessage) {
	if len(raw) == 0 || raw[0] != '[' {
		return
	}
	v, err := parseJSON(string(raw))
	if err != nil {
		return
	}
	for _, e := range v.([]any) {
		d, ok := chatReasoningDetail(e)
		if !ok {
			continue
		}
		if p.thinking < 0 {
			p.thinking = p.out.thinkingStart(Thinking{})
		}
		if p.details == nil {
			p.details = chatDetails{}
		}
		p.details.add(d)
	}
}

// toolCalls applies a tool_calls delta: each entry finds its call by stream
// index, else by id, else opens one, and appends its argument fragment. pi
// publishes a delta for every entry, an empty one included. Custom (grammar)
// tool calls, which only tools barness-ai does not declare produce, are
// skipped, as on the Responses path.
func (p *chatParser) toolCalls(raw json.RawMessage) {
	var entries []json.RawMessage
	if json.Unmarshal(raw, &entries) != nil {
		return
	}
	for _, e := range entries {
		var tc chatToolCallDelta
		if json.Unmarshal(e, &tc) != nil {
			continue
		}
		if tc.Function == nil && rawTruthy(tc.Custom) {
			continue
		}
		var name, args string
		if tc.Function != nil {
			name, _ = rawString(tc.Function.Name)
			args, _ = rawString(tc.Function.Arguments)
		}
		id, _ := rawString(tc.ID)
		var index float64
		indexed := len(tc.Index) > 0 && json.Unmarshal(tc.Index, &index) == nil
		k := p.findCall(indexed, index, id)
		if k < 0 {
			k = len(p.calls)
			p.calls = append(p.calls, chatCall{block: p.out.toolCallStart(id, name, ""), raw: newArgumentText("")})
		}
		call := &p.calls[k]
		if indexed && !call.indexed {
			call.indexed = true
			p.byIndex[index] = k
		}
		if id != "" {
			p.byID[id] = k
		}
		p.out.fillToolCallIdentity(call.block, id, name)
		if args != "" && !p.out.fitsToolJSON(call.raw.Len()+len(args)) {
			return
		}
		p.out.toolCallDelta(call.block, args, call.raw.append(args))
	}
}

// findCall is the call a delta continues: the one registered for its
// stream index, else for its id; -1 for a new call.
func (p *chatParser) findCall(indexed bool, index float64, id string) int {
	if k, ok := p.byIndex[index]; ok && indexed {
		return k
	}
	if k, ok := p.byID[id]; ok && id != "" {
		return k
	}
	return -1
}

// closeBlocks is pi's finishBlock for every block, in content order: the
// thinking block takes the streamed details as its signature, a tool call
// its final argument parse.
func (p *chatParser) closeBlocks() {
	for i := range p.out.blockCount() {
		switch {
		case i == p.text:
			p.out.textEnd(i, p.out.textContent(i), "")
		case i == p.thinking:
			sig := p.out.thinkingSignature(i)
			if p.details != nil {
				sig = p.details.signature()
			}
			p.out.thinkingEnd(i, p.out.thinkingText(i), sig)
		default:
			for _, c := range p.calls {
				if c.block == i {
					p.out.toolCallEnd(i, c.raw.String())
				}
			}
		}
	}
}

package ai

import (
	"encoding/json"
	"strings"
)

func (s *xaiResponsesState) arguments(p *responsesParser, typ string, ev map[string]json.RawMessage) *Error {
	item, ok := s.associated(ev, "function_call")
	if !ok {
		return xaiStreamFailure()
	}
	done := strings.HasSuffix(typ, ".done")
	field := "delta"
	if done {
		field = "arguments"
	}
	value, ok := rawString(ev[field])
	if !ok {
		return xaiStreamFailure()
	}
	if item.done && (!done || value != item.args.String()) {
		return xaiStreamFailure()
	}
	if done {
		return s.callSnapshot(p, item, value)
	}
	if item.argsDone {
		return xaiStreamFailure()
	}
	if !p.out.fitsToolJSON(item.args.Len() + len(value)) {
		return p.out.failure()
	}
	p.out.toolCallDelta(item.index, value, item.args.append(value))
	return nil
}
func (s *xaiResponsesState) callSnapshot(p *responsesParser, item *xaiStreamItem, value string) *Error {
	previous := item.args.String()
	if item.argsDone && value != previous || !strings.HasPrefix(value, previous) {
		return xaiStreamFailure()
	}
	if !p.out.fitsToolJSON(len(value)) {
		return p.out.failure()
	}
	if rest := strings.TrimPrefix(value, previous); rest != "" {
		p.out.toolCallDelta(item.index, rest, value)
	}
	item.args = newArgumentText(value)
	item.argsDone = true
	return nil
}
func (s *xaiResponsesState) closeCall(p *responsesParser, item *xaiStreamItem, obj map[string]json.RawMessage, raw json.RawMessage) *Error {
	args, ok := rawString(obj["arguments"])
	if !ok {
		return xaiStreamFailure()
	}
	if _, ok := xaiStreamObject([]byte(args)); !ok {
		if !s.incomplete || item.done {
			return xaiStreamFailure()
		}
		// A length/filtered terminal is a partial result, never an executable
		// function completion. Preserve its authoritative raw prefix only.
		return s.callSnapshot(p, item, args)
	}
	if failure := s.callSnapshot(p, item, args); failure != nil {
		return failure
	}
	if !item.done {
		p.out.toolCallEnd(item.index, args)
	}
	item.done = true
	item.native = append(json.RawMessage(nil), raw...)
	return nil
}

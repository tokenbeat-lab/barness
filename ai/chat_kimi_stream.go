package ai

import "encoding/json"

// Strict Kimi tool validation is scoped to this provider. The shared partial
// JSON display parser cannot prove that a normal terminal is executable.
func (p *chatParser) settleKimiChat() *Error {
	if !p.finished || p.stop == StopReasonError {
		return nil
	}
	if (p.stop == StopReasonStop && len(p.calls) > 0) || (p.stop == StopReasonToolUse && len(p.calls) == 0) || (p.kimiRequired && p.stop == StopReasonStop && len(p.calls) == 0) {
		return newError(CodeProtocol, PhaseStream, "Kimi Chat finish reason conflicts with the required tool calls")
	}
	return nil
}
func (p *chatParser) closeKimiChatTools() *Error {
	if p.stop == StopReasonLength {
		return nil
	}
	for _, c := range p.calls {
		if c.id == "" || c.name == "" || !kimiChatToolArguments(c.raw.String()) {
			return newError(CodeProtocol, PhaseStream, "Kimi Chat tool call is incomplete or unpairable")
		}
	}
	return nil
}
func (p *chatParser) validateKimiChatCall(tc chatToolCallDelta, indexed bool, index float64, id, name string, k int) *Error {
	if !indexed || index < 0 || index != float64(int64(index)) || tc.Function == nil || rawTruthy(tc.Custom) {
		return newError(CodeProtocol, PhaseStream, "Kimi Chat tool delta requires an indexed function")
	}
	if k >= 0 {
		c := p.calls[k]
		if (c.id != "" && id != "" && c.id != id) || (c.indexed && c.index != index) || (c.name != "" && name != "" && c.name != name) {
			return newError(CodeProtocol, PhaseStream, "Kimi Chat tool delta has conflicting identity")
		}
	}
	for _, v := range []json.RawMessage{tc.ID, tc.Function.Name, tc.Function.Arguments} {
		if len(v) > 0 {
			if _, ok := rawString(v); !ok {
				return newError(CodeProtocol, PhaseStream, "Kimi Chat tool fields must be strings")
			}
		}
	}
	return nil
}
func (p *chatParser) validateKimiChatDelta(d chatDelta) *Error {
	for _, v := range []json.RawMessage{d.Content, d.ReasoningContent} {
		if len(v) > 0 && string(v) != "null" {
			if _, ok := rawString(v); !ok {
				return newError(CodeProtocol, PhaseStream, "Kimi Chat content and reasoning must be text")
			}
		}
	}
	if rawTruthy(d.Reasoning) || rawTruthy(d.ReasoningText) || rawTruthy(d.ReasoningDetails) {
		return newError(CodeProtocol, PhaseStream, "Kimi Chat returned an unsupported reasoning format")
	}
	return nil
}

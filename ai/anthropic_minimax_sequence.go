package ai

import "encoding/json"

// MiniMax validates its own event contract without changing the permissive
// native Anthropic/pi parser. Synthetic protocol faults prove these guards;
// they do not assert that the service has emitted those faulty sequences.
func (p *anthropicParser) miniMaxSequence(ev sseEvent, e anthropicEvent) *Error {
	invalid := func() *Error {
		return newError(CodeProtocol, PhaseStream, "MiniMax Messages stream has an invalid event sequence")
	}
	if ev.name != e.Type || p.stopped {
		return invalid()
	}
	if ev.name == "ping" {
		return nil
	}
	if ev.name == "message_start" {
		if p.started || e.Message == nil {
			return invalid()
		}
		return nil
	}
	if !p.started {
		return invalid()
	}
	switch ev.name {
	case "message_delta":
		if len(p.open) != 0 || e.Delta == nil || (p.stop != "" && e.Delta.StopReason != "") {
			return invalid()
		}
	case "message_stop":
		if len(p.open) != 0 || p.stop == "" {
			return invalid()
		}
	case "content_block_start", "content_block_delta", "content_block_stop":
		if p.stop != "" || e.Index < 0 {
			return invalid()
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal([]byte(ev.data), &fields)
		if raw := fields["index"]; len(raw) == 0 || string(raw) == "null" {
			return invalid()
		}
		switch ev.name {
		case "content_block_start":
			if e.ContentBlock == nil || p.miniMaxSeen[e.Index] != "" || !miniMaxMessagesEventContent(ev.data, ev.name) {
				return invalid()
			}
			switch e.ContentBlock.Type {
			case "text", "thinking", "redacted_thinking", "tool_use":
			default:
				return invalid()
			}
			if p.miniMaxSeen == nil {
				p.miniMaxSeen = map[int64]string{}
			}
			p.miniMaxSeen[e.Index] = e.ContentBlock.Type
			if e.ContentBlock.Type == "tool_use" && len(e.ContentBlock.Input) > 0 {
				var input map[string]json.RawMessage
				if json.Unmarshal(e.ContentBlock.Input, &input) != nil || input == nil {
					return invalid()
				}
				if len(input) > 0 {
					if p.miniMaxInitial == nil {
						p.miniMaxInitial = map[int64]bool{}
					}
					p.miniMaxInitial[e.Index] = true
				}
			}
		case "content_block_delta":
			if e.Delta == nil || len(p.open[e.Index]) == 0 || !miniMaxMessagesEventContent(ev.data, ev.name) {
				return invalid()
			}
			var want string
			switch e.Delta.Type {
			case "text_delta":
				want = "text"
			case "thinking_delta", "signature_delta":
				want = "thinking"
			case "input_json_delta":
				want = "tool_use"
			default:
				return invalid()
			}
			if p.miniMaxSeen[e.Index] != want || (want == "tool_use" && p.miniMaxInitial[e.Index]) {
				return invalid()
			}
		case "content_block_stop":
			if len(p.open[e.Index]) == 0 {
				return invalid()
			}
			slot := p.open[e.Index][0]
			if slot.kind == blockToolCall && p.raw[slot.index].Len() > 0 {
				// Preserve fragments while streaming, but require a complete object
				// before accepting the provider's block stop. Native pi repair stays
				// unchanged; a malformed MiniMax turn is never promoted to success.
				var input map[string]json.RawMessage
				if json.Unmarshal([]byte(p.raw[slot.index].String()), &input) != nil || input == nil {
					return invalid()
				}
			}
		}
	default:
		return invalid()
	}
	return nil
}

// Required payload strings must actually be present and strings; decoding a
// JSON null into a Go string would otherwise silently produce an empty value.
func miniMaxMessagesEventContent(raw string, name string) bool {
	var root map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &root) != nil {
		return false
	}
	field := "content_block"
	if name == "content_block_delta" {
		field = "delta"
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(root[field], &obj) != nil || obj == nil {
		return false
	}
	var kind string
	if json.Unmarshal(obj["type"], &kind) != nil {
		return false
	}
	text := func(key string, required bool) bool {
		value, present := obj[key]
		if !present {
			return !required
		}
		var str string
		return string(value) != "null" && json.Unmarshal(value, &str) == nil
	}
	switch kind {
	case "text":
		return text("text", true)
	case "thinking":
		return text("thinking", true) && text("signature", false)
	case "redacted_thinking":
		var data string
		return json.Unmarshal(obj["data"], &data) == nil && !isBlank(data)
	case "tool_use":
		var id, name string
		return json.Unmarshal(obj["id"], &id) == nil && !isBlank(id) && json.Unmarshal(obj["name"], &name) == nil && !isBlank(name)
	case "text_delta":
		return text("text", true)
	case "thinking_delta":
		return text("thinking", true)
	case "signature_delta":
		return text("signature", true)
	case "input_json_delta":
		return text("partial_json", true)
	}
	return false
}

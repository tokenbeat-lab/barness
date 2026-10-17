package ai

// glmSequence validates the shape/order before any event mutates the public
// result. Own Messages fixtures establish complete blocks then final stop;
// permissive native pi parsing must not silently bless a broken GLM turn.
// Different tool indexes may interleave; argument assemblers remain separate.
func (p *anthropicParser) glmSequence(name string, e anthropicEvent) *Error {
	invalid := func() *Error {
		return newError(CodeProtocol, PhaseStream, "GLM Messages stream has an invalid event sequence")
	}
	if name != e.Type || p.stopped {
		return invalid()
	}
	if name == "message_start" {
		if p.started || e.Message == nil {
			return invalid()
		}
		return nil
	}
	if !p.started {
		return invalid()
	}
	switch name {
	case "message_delta":
		if len(p.open) != 0 || e.Delta == nil {
			return invalid()
		}
	case "message_stop":
		if len(p.open) != 0 || p.stop == "" {
			return invalid()
		}
	default:
		if p.stop != "" || e.Index < 0 {
			return invalid()
		}
		switch name {
		case "content_block_start":
			if e.ContentBlock == nil || p.glmSeen[e.Index] {
				return invalid()
			}
			switch e.ContentBlock.Type {
			case "text", "thinking", "redacted_thinking", "tool_use":
			default:
				return invalid()
			}
			if p.glmSeen == nil {
				p.glmSeen = map[int64]bool{}
			}
			p.glmSeen[e.Index] = true
		case "content_block_delta":
			if e.Delta == nil || len(p.open[e.Index]) == 0 {
				return invalid()
			}
			var kind blockKind
			switch e.Delta.Type {
			case "text_delta":
				kind = blockText
			case "thinking_delta", "signature_delta":
				kind = blockThinking
			case "input_json_delta":
				kind = blockToolCall
			default:
				return invalid()
			}
			if p.open[e.Index][0].kind != kind {
				return invalid()
			}
		case "content_block_stop":
			if len(p.open[e.Index]) == 0 {
				return invalid()
			}
		}
	}
	return nil
}

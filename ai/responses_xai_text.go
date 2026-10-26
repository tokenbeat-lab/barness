package ai

import (
	"encoding/json"
	"strings"
)

func (s *xaiResponsesState) text(p *responsesParser, typ string, ev map[string]json.RawMessage) *Error {
	kind, channel, indexField := "message", "content", "content_index"
	if strings.Contains(typ, "reasoning_") {
		kind = "reasoning"
		if strings.Contains(typ, "summary") {
			channel, indexField = "summary", "summary_index"
		}
	}
	item, ok := s.associated(ev, kind)
	if !ok {
		return xaiStreamFailure()
	}
	index, ok := xaiStreamIndex(ev[indexField])
	if !ok {
		return xaiStreamFailure()
	}
	field := "delta"
	done := strings.HasSuffix(typ, ".done")
	if done {
		field = "text"
		if strings.Contains(typ, "refusal") {
			field = "refusal"
		}
	}
	text := ""
	if strings.Contains(typ, "_part.") {
		part, ok := xaiStreamObject(ev["part"])
		if !ok {
			return xaiStreamFailure()
		}
		partKind, ok := rawString(part["type"])
		if !ok {
			return xaiStreamFailure()
		}
		if kind == "reasoning" && partKind != "summary_text" || kind == "message" && partKind != "output_text" && partKind != "refusal" {
			return xaiStreamFailure()
		}
		field = "text"
		if partKind == "refusal" {
			field = "refusal"
		}
		text, ok = rawString(part[field])
		if !ok {
			return xaiStreamFailure()
		}
		if !done && text == "" {
			return nil
		}
	} else {
		text, ok = rawString(ev[field])
		if !ok {
			return xaiStreamFailure()
		}
	}
	return s.setPart(p, item, channel, index, text, done)
}
func (s *xaiResponsesState) setPart(p *responsesParser, item *xaiStreamItem, channel string, index int64, value string, done bool) *Error {
	if item.parts[channel] == nil {
		item.parts[channel] = map[int64]*xaiStreamPart{}
	}
	part := item.parts[channel][index]
	if part == nil {
		part = &xaiStreamPart{text: newArgumentText("")}
		item.parts[channel][index] = part
	}
	previous := part.text.String()
	if done {
		if (part.done && previous != value) || !strings.HasPrefix(value, previous) {
			return xaiStreamFailure()
		}
		part.text = newArgumentText(value)
		part.done = true
	} else {
		if part.done {
			return xaiStreamFailure()
		}
		part.text.append(value)
	}
	if item.primary == "" {
		item.primary = channel
	}
	if channel != item.primary {
		return nil
	}
	text := s.visible(item, channel)
	if !strings.HasPrefix(text, item.emitted) {
		return xaiStreamFailure()
	}
	if delta := strings.TrimPrefix(text, item.emitted); delta != "" {
		s.emit(p, item, delta)
		item.emitted = text
	}
	return nil
}
func (s *xaiResponsesState) visible(item *xaiStreamItem, channel string) string {
	var b strings.Builder
	for index := int64(0); ; index++ {
		part := item.parts[channel][index]
		if part == nil {
			break
		}
		if index > 0 && item.kind == "reasoning" {
			b.WriteString("\n\n")
		}
		b.WriteString(part.text.String())
		if !part.done {
			break
		}
	}
	return b.String()
}
func (s *xaiResponsesState) emit(p *responsesParser, item *xaiStreamItem, text string) {
	if item.kind == "message" {
		p.out.textDelta(item.index, text)
	} else {
		p.out.thinkingDelta(item.index, text)
	}
}

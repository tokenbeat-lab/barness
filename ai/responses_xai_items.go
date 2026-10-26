package ai

import (
	"encoding/json"
	"strings"
)

func (s *xaiResponsesState) item(p *responsesParser, key int64, raw json.RawMessage) (*xaiStreamItem, *Error) {
	obj, ok := xaiStreamObject(raw)
	if !ok {
		return nil, xaiStreamFailure()
	}
	kind, ok := rawString(obj["type"])
	if !ok || (kind != "message" && kind != "reasoning" && kind != "function_call") {
		return nil, xaiStreamFailure()
	}
	id := ""
	if v, has := obj["id"]; has {
		id, ok = rawString(v)
		if !ok {
			return nil, xaiStreamFailure()
		}
	}
	call, name := "", ""
	if kind == "function_call" {
		call, ok = rawString(obj["call_id"])
		if !ok || call == "" {
			return nil, xaiStreamFailure()
		}
		name, ok = rawString(obj["name"])
		if !ok || name == "" {
			return nil, xaiStreamFailure()
		}
	}
	if raw, has := obj["status"]; has {
		if _, ok := rawString(raw); !ok {
			return nil, xaiStreamFailure()
		}
	}
	if kind == "message" {
		if raw, has := obj["role"]; has {
			role, ok := rawString(raw)
			if !ok || role != "assistant" {
				return nil, xaiStreamFailure()
			}
		}
	}
	cipher := ""
	if kind == "reasoning" {
		if _, err := xaiResponsesReasoning(raw); err != nil {
			return nil, xaiStreamFailure()
		}
		if raw, has := obj["encrypted_content"]; has && string(raw) != "null" {
			cipher, ok = rawString(raw)
			if !ok {
				return nil, xaiStreamFailure()
			}
		}
	}
	if item, has := s.items[key]; has {
		if item.kind != kind || item.id != id || item.call != call || item.name != name || item.cipher != "" && cipher != "" && item.cipher != cipher {
			return nil, xaiStreamFailure()
		}
		if cipher != "" {
			item.cipher = cipher
		}
		return item, nil
	}
	if prior, has := s.ids[id]; id != "" && has && prior != key {
		return nil, xaiStreamFailure()
	}
	if prior, has := s.calls[call]; call != "" && has && prior != key {
		return nil, xaiStreamFailure()
	}
	item := &xaiStreamItem{kind: kind, id: id, call: call, name: name, cipher: cipher, parts: map[string]map[int64]*xaiStreamPart{}}
	switch kind {
	case "message":
		item.index = p.out.textStart(Text{})
	case "reasoning":
		item.index = p.out.thinkingStart(Thinking{})
	case "function_call":
		args := ""
		if raw, has := obj["arguments"]; has {
			args, ok = rawString(raw)
			if !ok {
				return nil, xaiStreamFailure()
			}
		}
		if !p.out.fitsToolJSON(len(args)) {
			return nil, p.out.failure()
		}
		item.args = newArgumentText(args)
		item.index = p.out.toolCallStart(xaiResponsesToolID(call, id), name, args)
	}
	s.items[key] = item
	if id != "" {
		s.ids[id] = key
	}
	if call != "" {
		s.calls[call] = key
	}
	return item, nil
}
func (s *xaiResponsesState) closeItem(p *responsesParser, item *xaiStreamItem, raw json.RawMessage) *Error {
	obj, _ := xaiStreamObject(raw)
	if raw, has := obj["status"]; has {
		status, ok := rawString(raw)
		if !ok || status != "completed" && !(s.incomplete && status == "incomplete") {
			return xaiStreamFailure()
		}
	}
	if item.kind == "message" {
		if _, has := obj["content"]; !has {
			return xaiStreamFailure()
		}
	}
	if item.kind == "function_call" {
		return s.closeCall(p, item, obj, raw)
	}
	if item.done {
		prior, _ := xaiStreamObject(item.native)
		for _, field := range []string{"content", "summary", "arguments", "encrypted_content"} {
			old, has := prior[field]
			if !has || string(old) == "null" || string(old) == `""` {
				continue
			}
			if !xaiStreamEqual(old, obj[field]) {
				return xaiStreamFailure()
			}
		}
	}
	channels := []string{"content"}
	if item.kind == "reasoning" {
		channels = []string{"summary", "content"}
	}
	values := map[string]string{}
	for _, channel := range channels {
		rawParts, has := obj[channel]
		if !has {
			continue
		}
		var parts []json.RawMessage
		if json.Unmarshal(rawParts, &parts) != nil || parts == nil {
			return xaiStreamFailure()
		}
		texts := make([]string, len(parts))
		for i, rawPart := range parts {
			part, ok := xaiStreamObject(rawPart)
			if !ok {
				return xaiStreamFailure()
			}
			kind, ok := rawString(part["type"])
			if !ok {
				return xaiStreamFailure()
			}
			field := "text"
			if item.kind == "message" {
				if kind == "refusal" {
					field = "refusal"
				} else if kind != "output_text" {
					return xaiStreamFailure()
				}
			} else if kind != "summary_text" && kind != "reasoning_text" {
				return xaiStreamFailure()
			}
			text, ok := rawString(part[field])
			if !ok {
				return xaiStreamFailure()
			}
			texts[i] = text
			if !s.partKind(item, channel, int64(i), kind) {
				return xaiStreamFailure()
			}
			if failure := s.setPart(p, item, channel, int64(i), text, true); failure != nil {
				return failure
			}
		}
		for index := range item.parts[channel] {
			if index >= int64(len(parts)) {
				return xaiStreamFailure()
			}
		}
		separator := ""
		if item.kind == "reasoning" {
			separator = "\n\n"
		}
		values[channel] = strings.Join(texts, separator)
	}
	if item.kind == "reasoning" {
		if cipher, has := obj["encrypted_content"]; has && string(cipher) != "null" {
			if _, ok := rawString(cipher); !ok {
				return xaiStreamFailure()
			}
		}
	}
	primary := item.primary
	if primary == "" {
		primary = "content"
		if item.kind == "reasoning" {
			primary = "summary"
			if _, has := values[primary]; !has {
				primary = "content"
			}
		}
	}
	if _, has := values[primary]; !has {
		for _, channel := range channels {
			if _, has := values[channel]; has {
				primary = channel
				break
			}
		}
	}
	text, has := values[primary]
	if !has {
		text = s.visible(item, primary)
	}
	// When only one native representation is reported, the other stream channel
	// describes the same logical parts and must agree with that authority.
	for channel, parts := range item.parts {
		for i, part := range parts {
			target := item.parts[channel][i]
			if _, has := values[channel]; !has {
				target = item.parts[primary][i]
			}
			if target == nil || (part.done && target.text.String() != part.text.String()) || !strings.HasPrefix(target.text.String(), part.text.String()) {
				return xaiStreamFailure()
			}
		}
	}
	if !strings.HasPrefix(text, item.emitted) {
		return xaiStreamFailure()
	}
	if rest := strings.TrimPrefix(text, item.emitted); rest != "" {
		s.emit(p, item, rest)
		item.emitted = text
	}
	if item.done {
		if item.kind == "reasoning" {
			p.out.setThinkingSignature(item.index, string(raw))
		}
		item.native = append(json.RawMessage(nil), raw...)
		return nil
	}
	if item.kind == "message" {
		p.out.textEnd(item.index, text, "")
	} else {
		p.out.thinkingEnd(item.index, text, string(raw))
	}
	item.done = true
	item.native = append(json.RawMessage(nil), raw...)
	return nil
}

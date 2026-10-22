package ai

import (
	"encoding/json"
	"slices"
	"strings"
)

type kimiResponsesIdentity struct{ kind, id, call, name string }

func kimiResponsesProtocolFailure() *Error {
	return newError(CodeProtocol, PhaseStream, "Kimi Responses stream contains malformed or contradictory data")
}
func kimiResponsesObject(raw json.RawMessage) (map[string]json.RawMessage, bool) {
	var obj map[string]json.RawMessage
	ok := json.Unmarshal(raw, &obj) == nil && obj != nil
	return obj, ok
}
func kimiResponsesString(raw json.RawMessage) (string, bool) { s, ok := rawString(raw); return s, ok }
func kimiResponsesIndex(raw json.RawMessage) (int64, bool) {
	n, err := parseKimiCount(raw)
	return n.n, err == nil && n.present
}

func (p *responsesParser) validateKimiResponsesEvent(raw []byte) *Error {
	bad := kimiResponsesProtocolFailure
	ev, ok := kimiResponsesObject(raw)
	if !ok || p.terminal {
		return bad()
	}
	typ, ok := kimiResponsesString(ev["type"])
	if !ok {
		return bad()
	}
	if !slices.Contains([]string{"response.created", "response.in_progress", "response.output_item.added", "response.output_item.done", "response.content_part.added", "response.content_part.done", "response.output_text.delta", "response.output_text.done", "response.reasoning_summary_part.added", "response.reasoning_summary_part.done", "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done", "response.function_call_arguments.delta", "response.function_call_arguments.done", "response.completed", "response.incomplete", "response.failed", "error"}, typ) {
		return bad()
	}
	if sequence, present := ev["sequence_number"]; present {
		n, ok := kimiResponsesIndex(sequence)
		if !ok || (p.kimi.hasSequence && n <= p.kimi.sequence) {
			return bad()
		}
		p.kimi.sequence, p.kimi.hasSequence = n, true
	}
	indexed := strings.HasPrefix(typ, "response.output_item.") || strings.HasPrefix(typ, "response.content_part.") || strings.Contains(typ, "_summary_") || strings.Contains(typ, "output_text.") || strings.Contains(typ, "function_call_arguments.")
	if indexed {
		i, ok := kimiResponsesIndex(ev["output_index"])
		if !ok {
			return bad()
		}
		key := responsesIndex{kind: "number", value: i}
		slot, open := p.slots[key]
		if _, done := p.kimi.finished[key]; done {
			return bad()
		}
		if strings.HasPrefix(typ, "response.output_item.") {
			id, ok := p.kimi.recordItem(key, ev["item"])
			if !ok {
				return bad()
			}
			if open && slot.kind != id.kind {
				return bad()
			}
			if typ == "response.output_item.added" && open {
				return bad()
			}
			if typ == "response.output_item.done" {
				if !p.kimi.completedTextAgrees(key, ev["item"]) {
					return bad()
				}
				if slot.kind == "function_call" && p.kimi.partDone(key, ev) {
					item, _ := kimiResponsesObject(ev["item"])
					args, _ := kimiResponsesString(item["arguments"])
					if args != slot.raw.String() {
						return bad()
					}
				}
				p.kimi.doneItems[key] = ev["item"]
			}
		} else {
			kind := "message"
			if strings.Contains(typ, "reasoning_summary") {
				kind = "reasoning"
			}
			if strings.Contains(typ, "function_call_arguments") {
				kind = "function_call"
			}
			if open && slot.kind != kind {
				return bad()
			}
			if strings.HasSuffix(typ, ".delta") && (!open || p.kimi.partDone(key, ev)) {
				return bad()
			}
			if typ == "response.function_call_arguments.done" && !open {
				return bad()
			}
			if raw, has := ev["item_id"]; has {
				itemID, ok := kimiResponsesString(raw)
				if !ok || (p.kimi.identities[key].id != "" && itemID != p.kimi.identities[key].id) {
					return bad()
				}
			}
			if strings.HasSuffix(typ, ".delta") || strings.HasSuffix(typ, "_text.done") || typ == "response.function_call_arguments.done" {
				field := "delta"
				if strings.HasSuffix(typ, ".done") {
					field = "text"
				}
				if typ == "response.function_call_arguments.done" {
					field = "arguments"
				}
				if _, ok := kimiResponsesString(ev[field]); !ok {
					return bad()
				}
			}
			if strings.Contains(typ, "content_part.") || strings.Contains(typ, "summary_part.") {
				part, ok := kimiResponsesObject(ev["part"])
				want := "output_text"
				if kind == "reasoning" {
					want = "summary_text"
				}
				if !ok || string(part["type"]) != `"`+want+`"` {
					return bad()
				}
				if _, ok := kimiResponsesString(part["text"]); !ok {
					return bad()
				}
				if strings.HasSuffix(typ, ".done") {
					text, _ := kimiResponsesString(part["text"])
					if !p.kimi.recordCompletedText(key, ev, text) {
						return bad()
					}
					p.kimi.markPartDone(key, ev)
				}
			}
			if _, present := ev["content_index"]; present {
				if _, ok := kimiResponsesIndex(ev["content_index"]); !ok {
					return bad()
				}
			}
			if _, present := ev["summary_index"]; present {
				if _, ok := kimiResponsesIndex(ev["summary_index"]); !ok {
					return bad()
				}
			}
			if strings.HasSuffix(typ, "_text.done") || typ == "response.function_call_arguments.done" {
				if p.kimi.partDone(key, ev) {
					return bad()
				}
				p.kimi.markPartDone(key, ev)
			}
		}
	}
	if typ == "response.completed" || typ == "response.incomplete" || typ == "response.failed" {
		r, ok := kimiResponsesObject(ev["response"])
		if !ok {
			return bad()
		}
		status, ok := kimiResponsesString(r["status"])
		if !ok || typ != "response."+status {
			return bad()
		}
		if raw, has := r["output"]; has {
			var output []json.RawMessage
			if json.Unmarshal(raw, &output) != nil || output == nil {
				return bad()
			}
			for i, item := range output {
				key := responsesIndex{kind: "number", value: int64(i)}
				if !p.kimi.completedTextAgrees(key, item) {
					return bad()
				}
				if _, ok := p.kimi.recordItem(key, item); !ok {
					return bad()
				}
				if prior := p.kimi.doneItems[key]; prior != nil && !kimiResponsesItemsAgree(prior, item) {
					return bad()
				}
				if status == "completed" && !kimiResponsesCompletedItem(item) {
					return bad()
				}
			}
		}
		if status == "completed" {
			for _, item := range p.kimi.doneItems {
				if !kimiResponsesCompletedItem(item) {
					return bad()
				}
			}
		}
	}
	return nil
}
func (s *kimiResponsesState) recordItem(key responsesIndex, raw json.RawMessage) (kimiResponsesIdentity, bool) {
	item, ok := kimiResponsesObject(raw)
	if !ok {
		return kimiResponsesIdentity{}, false
	}
	kind, ok := kimiResponsesString(item["type"])
	if !ok || !slices.Contains([]string{"message", "reasoning", "function_call"}, kind) {
		return kimiResponsesIdentity{}, false
	}
	id := ""
	if v, present := item["id"]; present {
		var ok bool
		id, ok = kimiResponsesString(v)
		if !ok {
			return kimiResponsesIdentity{}, false
		}
	}
	identity := kimiResponsesIdentity{kind: kind, id: id}
	if kind == "function_call" {
		call, ok := kimiResponsesString(item["call_id"])
		if !ok || call == "" {
			return identity, false
		}
		name, ok := kimiResponsesString(item["name"])
		if !ok || name == "" {
			return identity, false
		}
		identity.call, identity.name = call, name
		if v, has := item["arguments"]; has {
			if _, ok := kimiResponsesString(v); !ok {
				return identity, false
			}
		}
	}
	if kind == "reasoning" {
		if encrypted, has := item["encrypted_content"]; has && string(encrypted) != "null" {
			return identity, false
		}
		if content, has := item["content"]; has && string(content) != "[]" {
			return identity, false
		}
	}
	for _, field := range []string{"content", "summary"} {
		if v, has := item[field]; has {
			var parts []map[string]json.RawMessage
			if json.Unmarshal(v, &parts) != nil || parts == nil {
				return identity, false
			}
			for _, part := range parts {
				typ, ok := kimiResponsesString(part["type"])
				want := "output_text"
				if kind == "reasoning" {
					want = "summary_text"
				}
				if !ok || typ != want {
					return identity, false
				}
				if _, ok := kimiResponsesString(part["text"]); !ok {
					return identity, false
				}
			}
		}
	}
	if prior, exists := s.identities[key]; exists && (prior.kind != identity.kind || (prior.id != "" && identity.id != "" && prior.id != identity.id) || prior.call != identity.call || prior.name != identity.name) {
		return identity, false
	}
	for other, prior := range s.identities {
		if other != key && ((id != "" && prior.id == id) || (identity.call != "" && identity.call == prior.call)) {
			return identity, false
		}
	}
	s.identities[key] = identity
	return identity, true
}
func kimiResponsesCompletedItem(raw json.RawMessage) bool {
	item, ok := kimiResponsesObject(raw)
	if !ok {
		return false
	}
	if string(item["type"]) == `"function_call"` {
		args, ok := kimiResponsesString(item["arguments"])
		return ok && kimiResponsesArguments(args)
	}
	return true
}
func kimiResponsesItemsAgree(first, last json.RawMessage) bool {
	a, ok := kimiResponsesObject(first)
	if !ok {
		return false
	}
	b, ok := kimiResponsesObject(last)
	if !ok {
		return false
	}
	for _, field := range []string{"type", "id", "call_id", "name", "arguments", "summary", "content"} {
		if av, has := a[field]; has {
			if bv, present := b[field]; present && compactJSON(av) != compactJSON(bv) {
				return false
			}
		}
	}
	return true
}
func (s *kimiResponsesState) partDone(key responsesIndex, event map[string]json.RawMessage) bool {
	i := int64(0)
	raw := event["content_index"]
	if raw == nil {
		raw = event["summary_index"]
	}
	if raw != nil {
		i, _ = kimiResponsesIndex(raw)
	}
	return s.doneParts[key][i]
}
func (s *kimiResponsesState) markPartDone(key responsesIndex, event map[string]json.RawMessage) {
	i := int64(0)
	raw := event["content_index"]
	if raw == nil {
		raw = event["summary_index"]
	}
	if raw != nil {
		i, _ = kimiResponsesIndex(raw)
	}
	if s.doneParts[key] == nil {
		s.doneParts[key] = map[int64]bool{}
	}
	s.doneParts[key][i] = true
}

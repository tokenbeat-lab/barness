package ai

import "encoding/json"

// Completed parts are immutable facts, even when an item or the final response
// supplies a larger content snapshot. Compare each known part independently so
// concatenation cannot hide changed text, moved boundaries or dropped parts.
func (s *kimiResponsesState) completedTextAgrees(key responsesIndex, raw json.RawMessage) bool {
	if len(s.textParts[key]) == 0 {
		return true
	}
	item, ok := kimiResponsesObject(raw)
	if !ok {
		return false
	}
	field := "content"
	if string(item["type"]) == `"reasoning"` {
		field = "summary"
	}
	var parts []map[string]json.RawMessage
	if json.Unmarshal(item[field], &parts) != nil {
		return false
	}
	for index, text := range s.textParts[key] {
		if index < 0 || index >= int64(len(parts)) {
			return false
		}
		got, ok := kimiResponsesString(parts[index]["text"])
		if !ok || got != text {
			return false
		}
	}
	return true
}

func (s *kimiResponsesState) recordCompletedText(key responsesIndex, ev map[string]json.RawMessage, text string) bool {
	raw := ev["content_index"]
	if raw == nil {
		raw = ev["summary_index"]
	}
	index := int64(0)
	if raw != nil {
		var ok bool
		index, ok = kimiResponsesIndex(raw)
		if !ok {
			return false
		}
	}
	if s.textParts[key] == nil {
		s.textParts[key] = map[int64]string{}
	}
	if previous, has := s.textParts[key][index]; has && previous != text {
		return false
	}
	s.textParts[key][index] = text
	return true
}

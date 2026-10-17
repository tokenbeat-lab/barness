package ai

import (
	"encoding/json"
	"errors"
	"strings"
)

// MiniMax uses the documented array content shape directly. This validator
// checks the unmodified frame before SDK decoding; it never repairs wire JSON.
func (p *responsesParser) validateMiniMaxEvent(raw []byte) error {
	var ev map[string]json.RawMessage
	if json.Unmarshal(raw, &ev) != nil || ev == nil {
		return errors.New("invalid event")
	}
	var typ string
	if json.Unmarshal(ev["type"], &typ) != nil || typ == "" {
		return errors.New("missing event type")
	}
	if p.terminal {
		return errors.New("event after terminal")
	}
	indexEvent := strings.HasPrefix(typ, "response.output_item.") || strings.Contains(typ, "_text.") || strings.Contains(typ, "function_call_arguments.") || strings.HasPrefix(typ, "response.refusal.")
	if indexEvent {
		var index int64
		if json.Unmarshal(ev["output_index"], &index) != nil || string(ev["output_index"]) == "null" || index < 0 {
			return errors.New("invalid output index")
		}
		key := responsesIndex{kind: "number", value: index}
		slot, open := p.slots[key]
		if _, finished := p.minimax.finished[key]; finished {
			return errors.New("event after item done")
		}
		kind := "message"
		if strings.Contains(typ, "reasoning_") {
			kind = "reasoning"
		}
		if strings.Contains(typ, "function_call_arguments") {
			kind = "function_call"
		}
		if strings.HasPrefix(typ, "response.output_item.") {
			itemKind, id, err := validateMiniMaxOutputItem(ev["item"], typ == "response.output_item.added")
			if err != nil {
				return err
			}
			if open && (slot.kind != itemKind || (p.minimax.ids[key] != "" && id != "" && p.minimax.ids[key] != id)) {
				return errors.New("mismatched output item")
			}
			if typ == "response.output_item.added" && open {
				return errors.New("duplicate output item")
			}
			p.minimax.ids[key] = id
		} else {
			if strings.HasSuffix(typ, ".delta") && (!open || slot.kind != kind || p.miniMaxPartFinished(key, ev)) {
				return errors.New("delta has no matching open item")
			}
			if typ == "response.function_call_arguments.done" && !open {
				return errors.New("arguments have no associated call")
			}
			if strings.HasSuffix(typ, ".done") && p.miniMaxPartFinished(key, ev) {
				return errors.New("duplicate content done")
			}
			if open && slot.kind != kind {
				return errors.New("content kind mismatch")
			}
			var itemID string
			if rawID, ok := ev["item_id"]; ok && (json.Unmarshal(rawID, &itemID) != nil || (p.minimax.ids[key] != "" && itemID != p.minimax.ids[key])) {
				return errors.New("content item ID mismatch")
			}
			field := "text"
			if strings.HasSuffix(typ, ".delta") {
				field = "delta"
			}
			if strings.HasPrefix(typ, "response.refusal.") && strings.HasSuffix(typ, ".done") {
				field = "refusal"
			}
			if typ == "response.function_call_arguments.done" {
				field = "arguments"
			}
			var value string
			if json.Unmarshal(ev[field], &value) != nil || string(ev[field]) == "null" {
				return errors.New("invalid content event text")
			}
		}
	}
	if typ == "response.completed" || typ == "response.incomplete" || typ == "response.failed" {
		var response map[string]json.RawMessage
		if json.Unmarshal(ev["response"], &response) != nil || response == nil {
			return errors.New("invalid terminal response")
		}
		var status string
		if json.Unmarshal(response["status"], &status) != nil || typ != "response."+status {
			return errors.New("terminal status mismatch")
		}
		if rawOutput, ok := response["output"]; ok {
			var output []json.RawMessage
			if json.Unmarshal(rawOutput, &output) != nil || string(rawOutput) == "null" {
				return errors.New("invalid terminal output")
			}
			for i, item := range output {
				kind, id, err := validateMiniMaxOutputItem(item, false)
				if err != nil {
					return err
				}
				key := responsesIndex{kind: "number", value: int64(i)}
				if slot, ok := p.minimax.finished[key]; ok && slot.kind != kind {
					return errors.New("terminal finished item kind mismatch")
				}
				if slot, ok := p.slots[key]; ok && slot.kind != kind {
					return errors.New("terminal item kind mismatch")
				}
				if prior := p.minimax.ids[key]; prior != "" && id != "" && prior != id {
					return errors.New("terminal item ID mismatch")
				}
			}
		}
	}
	return nil
}

func validateMiniMaxOutputItem(raw json.RawMessage, added bool) (string, string, error) {
	var item map[string]json.RawMessage
	invalid := errors.New("invalid MiniMax output item")
	if json.Unmarshal(raw, &item) != nil || item == nil {
		return "", "", invalid
	}
	var typ, id string
	if json.Unmarshal(item["type"], &typ) != nil {
		return "", "", invalid
	}
	if value, ok := item["id"]; ok && json.Unmarshal(value, &id) != nil {
		return "", "", invalid
	}
	switch typ {
	case "message", "reasoning":
		for _, field := range []string{"content", "summary"} {
			content, ok := item[field]
			if !ok {
				continue
			}
			var parts []map[string]json.RawMessage
			if json.Unmarshal(content, &parts) != nil || string(content) == "null" {
				return "", "", invalid
			}
			for _, part := range parts {
				var kind, text string
				if json.Unmarshal(part["type"], &kind) != nil {
					return "", "", invalid
				}
				textField := "text"
				valid := kind == "output_text"
				if typ == "reasoning" {
					valid = (field == "content" && kind == "reasoning_text") || (field == "summary" && kind == "summary_text")
				}
				if typ == "message" && kind == "refusal" {
					valid = true
					textField = "refusal"
				}
				if !valid || json.Unmarshal(part[textField], &text) != nil || string(part[textField]) == "null" {
					return "", "", invalid
				}
			}
		}
	case "function_call":
		var call, name, args string
		if json.Unmarshal(item["call_id"], &call) != nil || call == "" || json.Unmarshal(item["name"], &name) != nil || name == "" {
			return "", "", invalid
		}
		if value, ok := item["arguments"]; ok {
			if json.Unmarshal(value, &args) != nil || (!added && !json.Valid([]byte(args))) {
				return "", "", invalid
			}
		} else if !added {
			return "", "", invalid
		}
	default:
		return "", "", invalid
	}
	return typ, id, nil
}

func (p *responsesParser) miniMaxPartFinished(key responsesIndex, event map[string]json.RawMessage) bool {
	var index int64
	if raw, ok := event["content_index"]; ok {
		if json.Unmarshal(raw, &index) != nil || string(raw) == "null" || index < 0 {
			return true
		}
	}
	_, done := p.minimax.doneParts[key][index]
	return done
}

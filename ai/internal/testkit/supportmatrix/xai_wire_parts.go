package supportmatrix

import (
	"encoding/json"
	"strings"
)

func (i *xaiWireItem) partKind(channel, kind string) bool {
	if i.kind == "message" {
		return channel == "content" && (kind == "output_text" || kind == "refusal")
	}
	return i.kind == "reasoning" && (channel == "summary" && kind == "summary_text" || channel == "content" && kind == "reasoning_text")
}
func (i *xaiWireItem) text(typ string, f map[string]json.RawMessage) error {
	channel, indexField, partKind := "content", "content_index", "output_text"
	if strings.Contains(typ, "reasoning_") {
		if i.kind != "reasoning" {
			return refuse("xAI reasoning frame associated with non-reasoning item")
		}
		partKind = "reasoning_text"
		if strings.Contains(typ, "summary") {
			channel, indexField, partKind = "summary", "summary_index", "summary_text"
		}
	} else if i.kind != "message" {
		return refuse("xAI text frame associated with non-message item")
	}
	if strings.Contains(typ, "refusal") {
		partKind = "refusal"
	}
	index, err := xaiWireIndex(f[indexField])
	if err != nil {
		return err
	}
	done := strings.HasSuffix(typ, ".done")
	field := "delta"
	if done {
		field = "text"
		if partKind == "refusal" {
			field = "refusal"
		}
	}
	raw := f[field]
	if strings.Contains(typ, "_part.") {
		var part map[string]json.RawMessage
		if regionalDecode(f["part"], &part) != nil {
			return refuse("xAI raw part invalid")
		}
		partKind = xaiWireString(part["type"])
		field = "text"
		if partKind == "refusal" {
			field = "refusal"
		}
		raw = part[field]
	}
	value, ok := xaiWireText(raw)
	if !ok || !i.partKind(channel, partKind) {
		return refuse("xAI raw part kind or value invalid")
	}
	if i.done {
		p := i.parts[channel][index]
		if !done || p == nil || !p.done || p.text != value {
			return refuse("xAI raw part changed after completed item")
		}
	}
	// An empty added part carries no visible text and cannot select the
	// primary reasoning channel. The first effective text event selects it.
	if strings.Contains(typ, "_part.added") && value == "" {
		return nil
	}
	if i.primary == "" {
		i.primary = channel
	}
	return i.set(channel, index, partKind, value, done)
}
func (i *xaiWireItem) set(channel string, index int64, kind, value string, done bool) error {
	if i.parts[channel] == nil {
		i.parts[channel] = map[int64]*xaiWirePart{}
	}
	p := i.parts[channel][index]
	if p == nil {
		p = &xaiWirePart{kind: kind}
		i.parts[channel][index] = p
	}
	if p.kind != kind {
		return refuse("xAI raw part kind changed")
	}
	if done {
		if p.done && value != p.text || !strings.HasPrefix(value, p.text) {
			return refuse("xAI raw snapshot contradicts prefix")
		}
		p.text = value
		p.done = true
	} else {
		if p.done {
			return refuse("xAI raw delta follows part done")
		}
		p.text += value
	}
	target, has, err := i.authority(i.final, channel, index)
	if err != nil {
		return err
	}
	if has && (!strings.HasPrefix(target, p.text) || done && target != p.text) {
		return refuse("xAI raw text contradicts final authority")
	}
	return nil
}
func (i *xaiWireItem) authority(item map[string]json.RawMessage, channel string, index int64) (string, bool, error) {
	raw, ok := item[channel]
	if !ok && i.kind == "reasoning" {
		raw, ok = item["summary"]
		if !ok {
			raw, ok = item["content"]
		}
	}
	if !ok {
		return "", false, nil
	}
	var parts []map[string]json.RawMessage
	if regionalDecode(raw, &parts) != nil || index >= int64(len(parts)) {
		return "", true, refuse("xAI raw part index exceeds authoritative snapshot")
	}
	p := parts[index]
	field := "text"
	if xaiWireString(p["type"]) == "refusal" {
		field = "refusal"
	}
	return xaiWireString(p[field]), true, nil
}
func (i *xaiWireItem) visible() string {
	primary := i.primary
	if primary == "" {
		primary = "summary"
		if _, ok := i.final[primary]; !ok {
			primary = "content"
		}
	}
	if _, ok := i.final[primary]; !ok {
		for _, c := range []string{"summary", "content"} {
			if _, ok := i.final[c]; ok {
				primary = c
				break
			}
		}
	}
	var parts []map[string]json.RawMessage
	if raw, ok := i.final[primary]; ok && regionalDecode(raw, &parts) == nil {
		texts := make([]string, 0, len(parts))
		for _, p := range parts {
			texts = append(texts, xaiWireString(p["text"]))
		}
		return strings.Join(texts, "\n\n")
	}
	var text []string
	for n := int64(0); ; n++ {
		p := i.parts[primary][n]
		if p == nil {
			break
		}
		text = append(text, p.text)
		if !p.done {
			break
		}
	}
	return strings.Join(text, "\n\n")
}

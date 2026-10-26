package supportmatrix

import (
	"encoding/json"
	"strings"
)

type xaiWireItem struct {
	final    map[string]json.RawMessage
	snapshot map[string]json.RawMessage
	kind     string
	done     bool
	cipher   string
	args     string
	argsDone bool
	primary  string
	parts    map[string]map[int64]*xaiWirePart
}
type xaiWirePart struct {
	kind, text string
	done       bool
}

// The consumer reconciles original frames against final authority independently
// of the producer's PASS and the runtime's parser. Sparse/final-only streams are
// valid; an observed association or authoritative value cannot be contradicted.
func verifyXAIWire(frames []map[string]json.RawMessage, terminal map[string]json.RawMessage, output []map[string]json.RawMessage) ([]map[string]json.RawMessage, map[string]string, error) {
	items := map[int64]*xaiWireItem{}
	var order []int64
	fail := func() ([]map[string]json.RawMessage, map[string]string, error) {
		return nil, nil, refuse("xAI raw stream association or authority contradicts completion")
	}
	for _, f := range frames {
		kind := xaiWireString(f["type"])
		if n, ok := f["sequence_number"]; ok {
			if _, err := xaiWireIndex(n); err != nil {
				return fail()
			}
		}
		switch kind {
		case "response.completed":
			continue
		case "response.created", "response.in_progress":
			var r map[string]json.RawMessage
			if regionalDecode(f["response"], &r) != nil || r == nil {
				return fail()
			}
			for _, key := range []string{"id", "model"} {
				if raw, ok := r[key]; ok && !xaiJSONEqual(raw, terminal[key]) {
					return fail()
				}
			}
			continue
		}
		key, err := xaiWireIndex(f["output_index"])
		if err != nil || key >= int64(len(output)) {
			return fail()
		}
		item := items[key]
		switch kind {
		case "response.output_item.added", "response.output_item.done":
			var v map[string]json.RawMessage
			if regionalDecode(f["item"], &v) != nil || v == nil {
				return fail()
			}
			if item == nil {
				item = &xaiWireItem{final: output[key], kind: xaiWireString(output[key]["type"]), parts: map[string]map[int64]*xaiWirePart{}}
				items[key] = item
				order = append(order, key)
				if item.kind == "function_call" {
					item.args = xaiWireString(v["arguments"])
				}
			}
			if !item.identity(v) {
				return fail()
			}
			if kind == "response.output_item.added" {
				if item.done {
					return fail()
				}
				continue
			}
			if err := item.close(v); err != nil {
				return fail()
			}
		case "response.output_text.delta", "response.output_text.done", "response.refusal.delta", "response.refusal.done",
			"response.reasoning_text.delta", "response.reasoning_text.done", "response.reasoning_summary_text.delta", "response.reasoning_summary_text.done",
			"response.content_part.added", "response.content_part.done", "response.reasoning_summary_part.added", "response.reasoning_summary_part.done":
			if item == nil || !item.associated(f) || item.text(kind, f) != nil {
				return fail()
			}
		case "response.function_call_arguments.delta", "response.function_call_arguments.done":
			if item == nil || item.kind != "function_call" || !item.associated(f) {
				return fail()
			}
			field := "delta"
			done := strings.HasSuffix(kind, ".done")
			if done {
				field = "arguments"
			}
			value, ok := xaiWireText(f[field])
			if !ok {
				return fail()
			}
			if done {
				if (item.argsDone && item.args != value) || !strings.HasPrefix(value, item.args) {
					return fail()
				}
				item.args = value
				item.argsDone = true
			} else {
				if item.done || item.argsDone {
					return fail()
				}
				item.args += value
			}
		default:
			return fail()
		}
	}
	for i, final := range output {
		k := int64(i)
		item := items[k]
		if item == nil {
			item = &xaiWireItem{final: final, kind: xaiWireString(final["type"]), parts: map[string]map[int64]*xaiWirePart{}}
			items[k] = item
			order = append(order, k)
		}
		if !item.identity(final) || item.close(final) != nil {
			return fail()
		}
	}
	ordered := make([]map[string]json.RawMessage, 0, len(output))
	visible := map[string]string{}
	for _, k := range order {
		item := items[k]
		ordered = append(ordered, item.final)
		if item.kind == "reasoning" {
			visible[xaiWireString(item.final["id"])] = item.visible()
		}
	}
	return ordered, visible, nil
}
func xaiWireString(raw json.RawMessage) string { v, _ := xaiWireText(raw); return v }
func xaiWireText(raw json.RawMessage) (string, bool) {
	var s string
	return s, len(raw) > 0 && string(raw) != "null" && json.Unmarshal(raw, &s) == nil
}
func xaiWireIndex(raw json.RawMessage) (int64, error) {
	var n int64
	if len(raw) == 0 || string(raw) == "null" || json.Unmarshal(raw, &n) != nil || n < 0 {
		return 0, refuse("xAI raw stream index invalid")
	}
	return n, nil
}
func (i *xaiWireItem) associated(f map[string]json.RawMessage) bool {
	for _, key := range []string{"item_id", "call_id"} {
		if raw, ok := f[key]; ok {
			want := "id"
			if key == "call_id" {
				want = key
			}
			if !xaiJSONEqual(raw, i.final[want]) {
				return false
			}
		}
	}
	return true
}
func (i *xaiWireItem) identity(v map[string]json.RawMessage) bool {
	for _, key := range []string{"type", "id"} {
		if !xaiJSONEqual(v[key], i.final[key]) {
			return false
		}
	}
	if i.kind == "function_call" {
		for _, key := range []string{"call_id", "name"} {
			if !xaiJSONEqual(v[key], i.final[key]) {
				return false
			}
		}
	}
	if role, ok := v["role"]; ok && xaiWireString(role) != "assistant" {
		return false
	}
	if status, ok := v["status"]; ok {
		if _, valid := xaiWireText(status); !valid {
			return false
		}
	}
	if i.kind == "reasoning" {
		if verifyXAIReasoning(v) != nil {
			return false
		}
		if raw, ok := v["encrypted_content"]; ok && string(raw) != "null" {
			cipher, valid := xaiWireText(raw)
			if !valid || i.cipher != "" && cipher != "" && cipher != i.cipher {
				return false
			}
			if cipher != "" {
				i.cipher = cipher
			}
		}
	}
	return true
}
func (i *xaiWireItem) close(v map[string]json.RawMessage) error {
	bad := func() error { return refuse("xAI raw item snapshot contradicts stream") }
	if raw, ok := v["status"]; ok && xaiWireString(raw) != "completed" {
		return bad()
	}
	if i.done {
		for _, k := range []string{"content", "summary", "arguments", "encrypted_content"} {
			old, ok := i.snapshot[k]
			if ok && string(old) != "null" && string(old) != `""` && !xaiJSONEqual(old, v[k]) {
				return bad()
			}
		}
	}
	if i.kind == "function_call" {
		args, ok := xaiWireText(v["arguments"])
		if !ok || !strings.HasPrefix(args, i.args) || i.argsDone && i.args != args {
			return bad()
		}
		var object map[string]json.RawMessage
		if regionalDecode([]byte(args), &object) != nil || object == nil {
			return bad()
		}
		i.args = args
		i.argsDone = true
	} else {
		if i.kind == "message" {
			if _, ok := v["content"]; !ok {
				return bad()
			}
		}
		for _, channel := range []string{"summary", "content"} {
			raw, ok := v[channel]
			if !ok {
				continue
			}
			var parts []map[string]json.RawMessage
			if regionalDecode(raw, &parts) != nil || parts == nil {
				return bad()
			}
			for n, p := range parts {
				kind := xaiWireString(p["type"])
				field := "text"
				if kind == "refusal" {
					field = "refusal"
				}
				text, valid := xaiWireText(p[field])
				if !valid || !i.partKind(channel, kind) || i.set(channel, int64(n), kind, text, true) != nil {
					return bad()
				}
			}
			for n := range i.parts[channel] {
				if n >= int64(len(parts)) {
					return bad()
				}
			}
		}
		// If only one reasoning representation is authoritative, both channels
		// describe those same logical parts. No absent channel can hide a conflict.
		for channel, parts := range i.parts {
			for n, p := range parts {
				target, has, err := i.authority(v, channel, n)
				if err != nil || has && (!strings.HasPrefix(target, p.text) || p.done && p.text != target) {
					return bad()
				}
			}
		}
	}
	i.done = true
	i.snapshot = v
	return nil
}

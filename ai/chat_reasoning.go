package ai

import "encoding/json"

// Chat Completions carries replayable reasoning as reasoning_details, an
// array of typed entries (OpenRouter's format, which other providers adopt):
// reasoning.text, reasoning.summary and opaque reasoning.encrypted. pi keeps
// them, merged as they stream, as the JSON signature of the turn's thinking
// block and sends them back unchanged to the same model. Entries are
// *jsonObject values so key order and number spellings survive exactly as
// JavaScript handles them.

// chatReasoningDetail validates one entry as pi's isOpenAIReasoningDetail:
// an object whose id is absent, null or a string, format absent or a string,
// index absent or a number, and whose type's payload is a string.
func chatReasoningDetail(v any) (*jsonObject, bool) {
	d, ok := v.(*jsonObject)
	if !ok {
		return nil, false
	}
	if id, has := d.get("id"); has && id != nil {
		if _, isString := id.(string); !isString {
			return nil, false
		}
	}
	if format, has := d.get("format"); has {
		if _, isString := format.(string); !isString {
			return nil, false
		}
	}
	if index, has := d.get("index"); has {
		if _, isNumber := index.(float64); !isNumber {
			return nil, false
		}
	}
	isString := func(key string) bool {
		v, _ := d.get(key)
		_, ok := v.(string)
		return ok
	}
	typ, _ := d.get("type")
	switch typ {
	case "reasoning.summary":
		return d, isString("summary")
	case "reasoning.encrypted":
		return d, isString("data")
	case "reasoning.text":
		sig, has := d.get("signature")
		return d, isString("text") && (!has || sig == nil || isString("signature"))
	}
	return nil, false
}

// chatDetails are the reasoning_details a stream delivered so far, merged
// as pi's appendOpenAIReasoningDetail merges them: consecutive text or
// summary entries extend one logical entry; encrypted entries stay
// discrete. Each entry is pi's shallow copy of the first delta.
type chatDetails []*jsonObject

func (ds *chatDetails) add(d *jsonObject) {
	if n := len(*ds); n > 0 {
		last := (*ds)[n-1]
		lastType, _ := last.get("type")
		typ, _ := d.get("type")
		if typ == lastType && (typ == "reasoning.text" || typ == "reasoning.summary") {
			field := "text"
			if typ == "reasoning.summary" {
				field = "summary"
			}
			a, _ := last.get(field)
			b, _ := d.get(field)
			last.set(field, a.(string)+b.(string))
			if typ == "reasoning.text" {
				orAssign(last, d, "signature")
			}
			nullishAssign(last, d, "id")
			orAssign(last, d, "format")
			nullishAssign(last, d, "index")
			return
		}
	}
	copied := newJSONObject()
	for _, k := range d.keys {
		copied.set(k, d.values[k])
	}
	*ds = append(*ds, copied)
}

// sourceValue is source[key], undefined when absent.
func sourceValue(source *jsonObject, key string) any {
	if v, ok := source.get(key); ok {
		return v
	}
	return jsUndefined
}

// orAssign is JavaScript's target[key] ||= source[key].
func orAssign(target, source *jsonObject, key string) {
	if v, ok := target.get(key); !ok || !jsTruthy(v) || v == jsUndefined {
		target.set(key, sourceValue(source, key))
	}
}

// nullishAssign is JavaScript's target[key] ??= source[key].
func nullishAssign(target, source *jsonObject, key string) {
	if v, ok := target.get(key); !ok || v == nil || v == jsUndefined {
		target.set(key, sourceValue(source, key))
	}
}

// signature is the details as pi stores them on the thinking block,
// JSON.stringify of the array.
func (ds chatDetails) signature() string {
	arr := make([]any, len(ds))
	for i, d := range ds {
		arr[i] = d
	}
	return stringifyJSON(arr)
}

// chatReasoningDetails reads a thinking signature as pi's
// parseOpenAIReasoningDetails: a non-empty JSON array of valid entries,
// returned as the JSON to replay.
func chatReasoningDetails(signature string) (json.RawMessage, bool) {
	if signature == "" {
		return nil, false
	}
	v, err := parseJSON(signature)
	arr, isArray := v.([]any)
	if err != nil || !isArray || len(arr) == 0 {
		return nil, false
	}
	for _, e := range arr {
		if _, ok := chatReasoningDetail(e); !ok {
			return nil, false
		}
	}
	return json.RawMessage(stringifyJSON(arr)), true
}

// chatLegacyReasoningDetails reads the encrypted reasoning older pi versions
// kept on tool calls' thought signatures (pi's
// parseLegacyEncryptedReasoningDetail): each a reasoning.encrypted entry with
// a non-empty id and data. nil when no call holds one.
func chatLegacyReasoningDetails(calls []ToolCall) json.RawMessage {
	var found []any
	for _, c := range calls {
		sig, _ := c.ThoughtSignature.Get()
		if sig == "" {
			continue
		}
		v, err := parseJSON(sig)
		d, ok := chatReasoningDetail(v)
		if err != nil || !ok {
			continue
		}
		typ, _ := d.get("type")
		id, _ := d.get("id")
		data, _ := d.get("data")
		if typ == "reasoning.encrypted" && id != "" && id != nil && data != "" {
			if _, isString := id.(string); isString {
				found = append(found, d)
			}
		}
	}
	if found == nil {
		return nil
	}
	return json.RawMessage(stringifyJSON(found))
}

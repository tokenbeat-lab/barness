package ai

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
)

// Tool argument JSON is handled as JavaScript handles it, because pi-ai's
// observable results depend on it: arguments are replayed as
// JSON.stringify(JSON.parse(raw)), so key order, duplicate keys and number
// spelling must follow JavaScript rather than encoding/json's map semantics.
//
// A JSON value here is nil, bool, float64, string, []any or *jsonObject, as
// JSON.parse produces them. Numbers are IEEE doubles, exactly like JavaScript.
// The exact text the provider sent stays available as ToolCall.RawArguments.

// jsonObject is a JavaScript object: keys keep first-insertion order, a later
// duplicate key replaces the value in place.
type jsonObject struct {
	keys   []string
	values map[string]any
}

func newJSONObject() *jsonObject { return &jsonObject{values: map[string]any{}} }

func (o *jsonObject) set(key string, v any) {
	if _, ok := o.values[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.values[key] = v
}

func (o *jsonObject) get(key string) (any, bool) {
	v, ok := o.values[key]
	return v, ok
}

func (o *jsonObject) remove(key string) {
	if _, ok := o.values[key]; ok {
		delete(o.values, key)
		o.keys = slices.DeleteFunc(o.keys, func(k string) bool { return k == key })
	}
}

// ordered is the property order JavaScript enumerates: integer-index keys in
// ascending numeric order first, then the other keys in insertion order.
func (o *jsonObject) ordered() []string {
	var indexes, names []string
	for _, k := range o.keys {
		if isArrayIndex(k) {
			indexes = append(indexes, k)
		} else {
			names = append(names, k)
		}
	}
	slices.SortFunc(indexes, func(a, b string) int {
		x, _ := strconv.ParseUint(a, 10, 32)
		y, _ := strconv.ParseUint(b, 10, 32)
		return int(int64(x) - int64(y))
	})
	return append(indexes, names...)
}

// isArrayIndex reports whether k is a canonical uint32 array index below 2³²−1.
func isArrayIndex(k string) bool {
	if k == "" || (len(k) > 1 && k[0] == '0') {
		return false
	}
	n, err := strconv.ParseUint(k, 10, 32)
	return err == nil && n < math.MaxUint32
}

// maxJSONDepth bounds nesting like encoding/json, so hostile input cannot
// exhaust the stack.
const maxJSONDepth = 10000

var errJSONDepth = errors.New("ai: JSON nested too deeply")

// parseJSON is JavaScript's JSON.parse: one JSON text, surrounding whitespace
// allowed, nothing after it.
func parseJSON(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	v, err := decodeJSONValue(dec, 0)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, errors.New("ai: unexpected data after JSON value")
	}
	return v, nil
}

func decodeJSONValue(dec *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, errJSONDepth
	}
	tok, err := dec.Token()
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch tok := tok.(type) {
	case json.Delim:
		switch tok {
		case '{':
			obj := newJSONObject()
			for dec.More() {
				key, err := dec.Token()
				if err != nil {
					return nil, err
				}
				v, err := decodeJSONValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				obj.set(key.(string), v)
			}
			_, err := dec.Token()
			return obj, err
		case '[':
			arr := []any{}
			for dec.More() {
				v, err := decodeJSONValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			}
			_, err := dec.Token()
			return arr, err
		}
		return nil, errors.New("ai: unexpected JSON delimiter")
	case json.Number:
		return jsNumberLiteral(string(tok))
	default:
		return tok, nil // nil, bool or string
	}
}

// jsNumberLiteral converts a JSON number literal as JavaScript does; a literal
// too large for a double becomes ±Infinity rather than an error.
func jsNumberLiteral(lit string) (float64, error) {
	f, err := strconv.ParseFloat(lit, 64)
	var numErr *strconv.NumError
	if err != nil && !(errors.As(err, &numErr) && numErr.Err == strconv.ErrRange) {
		return 0, err
	}
	return f, nil
}

// stringifyJSON is JavaScript's JSON.stringify for a JSON value.
func stringifyJSON(v any) string {
	var b strings.Builder
	writeJSON(&b, v)
	return b.String()
}

func writeJSON(b *strings.Builder, v any) {
	switch v := v.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		b.WriteString(strconv.FormatBool(v))
	case float64:
		writeJSNumber(b, v)
	case string:
		writeJSString(b, v)
	case []any:
		b.WriteByte('[')
		for i, e := range v {
			if i > 0 {
				b.WriteByte(',')
			}
			writeJSON(b, e)
		}
		b.WriteByte(']')
	case *jsonObject:
		b.WriteByte('{')
		first := true
		for _, k := range v.ordered() {
			if v.values[k] == jsUndefined {
				continue
			}
			if !first {
				b.WriteByte(',')
			}
			first = false
			writeJSString(b, k)
			b.WriteByte(':')
			writeJSON(b, v.values[k])
		}
		b.WriteByte('}')
	}
}

// jsUndefined is a property JavaScript code assigned undefined: it holds its
// place in the key order, so a later value keeps that position, but
// JSON.stringify leaves it out. JSON.parse never produces it.
var jsUndefined any = jsUndefinedValue{}

type jsUndefinedValue struct{}

// writeJSNumber formats like JavaScript's Number#toString in JSON.stringify:
// non-finite numbers are null and -0 is 0. encoding/json already formats
// finite doubles as ES6 does.
func writeJSNumber(b *strings.Builder, f float64) {
	switch {
	case math.IsNaN(f) || math.IsInf(f, 0):
		b.WriteString("null")
	case f == 0:
		b.WriteByte('0')
	default:
		out, _ := json.Marshal(f)
		b.Write(out)
	}
}

// writeJSString quotes like JSON.stringify: only the quote, the backslash and
// control characters are escaped.
func writeJSString(b *strings.Builder, s string) {
	const hex = "0123456789abcdef"
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if c < 0x20 {
				b.WriteString(`\u00`)
				b.WriteByte(hex[c>>4])
				b.WriteByte(hex[c&0xf])
			} else {
				b.WriteByte(c)
			}
		}
	}
	b.WriteByte('"')
}

// RepairToolJSON ports pi-ai's repairJson: inside string literals it escapes
// raw control characters and doubles a backslash that starts no valid escape.
// It never completes truncated JSON and leaves everything outside strings
// untouched.
func RepairToolJSON(raw string) string {
	var b strings.Builder
	inString := false
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if !inString {
			b.WriteByte(c)
			inString = c == '"'
			continue
		}
		switch {
		case c == '"':
			b.WriteByte(c)
			inString = false
		case c == '\\':
			if i+1 == len(raw) {
				b.WriteString(`\\`)
				continue
			}
			next := raw[i+1]
			if next == 'u' && i+6 <= len(raw) && isHex4(raw[i+2:i+6]) {
				b.WriteString(raw[i : i+6])
				i += 5
				continue
			}
			// As in pi, "u" counts as a valid escape even without four hex
			// digits, so such text stays invalid rather than being rewritten.
			if strings.IndexByte(`"\/bfnrtu`, next) >= 0 {
				b.WriteByte('\\')
				b.WriteByte(next)
				i++
				continue
			}
			b.WriteString(`\\`)
		case c < 0x20:
			b.WriteString(escapeControl(c))
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func isHex4(s string) bool {
	for i := 0; i < 4; i++ {
		if !strings.ContainsRune("0123456789abcdefABCDEF", rune(s[i])) {
			return false
		}
	}
	return true
}

func escapeControl(c byte) string {
	var b strings.Builder
	writeJSString(&b, string([]byte{c}))
	s := b.String()
	return s[1 : len(s)-1]
}

// parseJSONWithRepair is pi-ai's parseJsonWithRepair: JSON.parse, then once
// more after RepairToolJSON when repair changed anything. It never accepts
// incomplete JSON.
func parseJSONWithRepair(s string) (any, error) {
	v, err := parseJSON(s)
	if err == nil {
		return v, nil
	}
	if repaired := RepairToolJSON(s); repaired != s {
		return parseJSON(repaired)
	}
	return nil, err
}

// displayArguments is a tool call's Arguments for raw argument text.
func displayArguments(raw string) UncheckedArguments {
	return UncheckedArguments(stringifyJSON(parseStreamingJSON(raw)))
}

// parseStreamingJSON is pi-ai's parseStreamingJson, the display parse of
// possibly incomplete argument JSON: a complete (or repairable) text parses
// exactly; otherwise the partial parser recovers what it can; failing that,
// an empty object. It never fails.
func parseStreamingJSON(s string) any {
	if strings.TrimSpace(s) == "" {
		return newJSONObject()
	}
	if v, err := parseJSONWithRepair(s); err == nil {
		return v
	}
	for _, text := range []string{s, RepairToolJSON(s)} {
		if v, err := parsePartialJSON(text); err == nil {
			if v == nil {
				return newJSONObject()
			}
			return v
		}
	}
	return newJSONObject()
}

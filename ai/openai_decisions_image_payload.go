package ai

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/buger/jsonparser"
)

// A preflight owns only counters. It walks ordinary callback containers and
// raw JSON without serializing or copying image strings. Custom encoders run
// once in boundHooks; their frozen bytes undergo the same checks afterwards.
type decisionsImageBudget struct {
	policy    *ResourcePolicy
	count     int
	remaining int64
}

func (c *Client) imageBudget() *decisionsImageBudget {
	b := &decisionsImageBudget{policy: &c.policy}
	if c.policy.Classifier.Images != nil {
		b.remaining = c.policy.Classifier.Images.MaxTotalImageBytes
	}
	return b
}

func (c *Client) checkDecisionsInputPayload(input any) *Error {
	b := c.imageBudget()
	v, failure := imagePayloadValue(reflect.ValueOf(input))
	if failure != nil || !v.IsValid() {
		return failure
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			part, kind, _, err := jsonparser.Get(raw)
			if err != nil {
				return decisionsSchemaFailure()
			}
			if kind == jsonparser.String {
				return b.rawText(part)
			}
		}
	}
	if !usesCustomJSON(v) && v.Kind() == reflect.String {
		return c.checkDecisionsText(v.String())
	}
	return b.value(v, 0)
}

func (c *Client) checkDecisionsRawPayload(body []byte) *Error {
	raw, kind, _, err := jsonparser.Get(body, "input")
	if err != nil {
		return decisionsSchemaFailure()
	}
	if kind == jsonparser.String {
		return c.imageBudget().rawText(raw)
	}
	return c.imageBudget().raw(raw, kind, 0)
}

func (b *decisionsImageBudget) takeImage(size int64) *Error {
	b.count++
	p := b.policy
	if size > p.MaxImageBytes {
		return limitFailure(PhaseRequest, "MaxImageBytes", p.MaxImageBytes)
	}
	if ip := p.Classifier.Images; ip != nil {
		if b.count > ip.MaxImages {
			return limitFailure(PhaseRequest, "Classifier.Images.MaxImages", int64(ip.MaxImages))
		}
		if size > b.remaining {
			return limitFailure(PhaseRequest, "Classifier.Images.MaxTotalImageBytes", ip.MaxTotalImageBytes)
		}
		b.remaining -= size
	}
	return nil
}

func (b *decisionsImageBudget) field(key string, value reflect.Value, depth int) *Error {
	if key == "file_id" || key == "item_reference" {
		return decisionsImageDenied()
	}
	v, failure := imagePayloadValue(value)
	if failure != nil || !v.IsValid() {
		return failure
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			part, kind, _, err := jsonparser.Get(raw)
			if err != nil {
				return decisionsSchemaFailure()
			}
			return b.rawField(key, part, kind, depth)
		}
	}
	if usesCustomJSON(v) {
		return nil
	}
	if key == "image_url" && v.Kind() == reflect.String {
		url := v.String()
		if !strings.HasPrefix(url, "data:") {
			return decisionsImageDenied()
		}
		_, data, ok := strings.Cut(url, ";base64,")
		if !ok {
			return decisionsSchemaFailure()
		}
		return b.takeImage(base64Size(data))
	}
	if key == "text" && v.Kind() == reflect.String && int64(v.Len()) > b.policy.Classifier.MaxStateBytes {
		return limitFailure(PhaseRequest, "Classifier.MaxStateBytes", b.policy.Classifier.MaxStateBytes)
	}
	return b.value(v, depth+1)
}

func (b *decisionsImageBudget) value(value reflect.Value, depth int) *Error {
	if depth > maxJSONDepth {
		return limitFailure(PhaseRequest, "MaxJSONDepth", maxJSONDepth)
	}
	v, failure := imagePayloadValue(value)
	if failure != nil || !v.IsValid() {
		return failure
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			part, kind, _, err := jsonparser.Get(raw)
			if err != nil {
				return decisionsSchemaFailure()
			}
			return b.raw(part, kind, depth)
		}
	}
	if usesCustomJSON(v) {
		return nil
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if int64(v.Len()) > b.policy.MaxRequestBytes {
			return limitFailure(PhaseRequest, "MaxRequestBytes", b.policy.MaxRequestBytes)
		}
		for i := 0; i < v.Len(); i++ {
			if failure := b.value(v.Index(i), depth+1); failure != nil {
				return failure
			}
		}
	case reflect.Map:
		entries := v.MapRange()
		for entries.Next() {
			key := ""
			if entries.Key().Kind() == reflect.String {
				key = entries.Key().String()
			}
			if failure := b.field(key, entries.Value(), depth+1); failure != nil {
				return failure
			}
		}
	case reflect.Struct:
		// Selected tagged/promoted fields share the existing image boundary's
		// encoder-aware selection. Unknown fields fail the final closed schema.
		for _, key := range []string{"role", "content", "type", "text", "image_url", "file_id", "item_reference"} {
			field, failure := imageStructField(v, key)
			if failure != nil {
				return failure
			}
			if field.IsValid() {
				if failure := b.field(key, field, depth+1); failure != nil {
					return failure
				}
			}
		}
	}
	return nil
}

func (b *decisionsImageBudget) raw(raw []byte, kind jsonparser.ValueType, depth int) *Error {
	if depth > maxJSONDepth {
		return limitFailure(PhaseRequest, "MaxJSONDepth", maxJSONDepth)
	}
	var failure *Error
	switch kind {
	case jsonparser.Object:
		err := jsonparser.ObjectEach(raw, func(key, value []byte, kind jsonparser.ValueType, _ int) error {
			return b.rawField(string(key), value, kind, depth+1)
		})
		if err != nil {
			if failure, ok := err.(*Error); ok {
				return failure
			}
			return decisionsSchemaFailure()
		}
	case jsonparser.Array:
		_, err := jsonparser.ArrayEach(raw, func(value []byte, kind jsonparser.ValueType, _ int, err error) {
			if failure == nil && err == nil {
				failure = b.raw(value, kind, depth+1)
			}
		})
		if failure != nil {
			return failure
		}
		if err != nil {
			return decisionsSchemaFailure()
		}
	}
	return nil
}

func (b *decisionsImageBudget) rawField(key string, raw []byte, kind jsonparser.ValueType, depth int) *Error {
	if key == "file_id" || key == "item_reference" {
		return decisionsImageDenied()
	}
	if key == "text" && kind == jsonparser.String {
		return b.rawText(raw)
	}
	if key != "image_url" || kind != jsonparser.String {
		return b.raw(raw, kind, depth)
	}
	// Decode only individual escapes and a bounded prefix. No allocation
	// scales with the base64 length, including JSON-escaped callback data.
	var prefix [128]byte
	var escaped [6]byte
	n, padding, prefixLen := int64(0), int64(0), 0
	for i := 0; i < len(raw); {
		ch := raw[i]
		i++
		if ch == '\\' {
			length := 2
			if i < len(raw) && raw[i] == 'u' {
				length = 6
			}
			if i-1+length > len(raw) {
				return decisionsSchemaFailure()
			}
			decoded, err := jsonparser.Unescape(raw[i-1:i-1+length], escaped[:])
			if err != nil || len(decoded) != 1 {
				return decisionsSchemaFailure()
			}
			ch = decoded[0]
			i += length - 1
		}
		if n < int64(len(prefix)) {
			prefix[n] = ch
		}
		n++
		if prefixLen == 0 && n >= 8 && n <= int64(len(prefix)) && string(prefix[n-8:n]) == ";base64," {
			prefixLen = int(n)
		}
		if ch == '=' {
			padding++
		} else {
			padding = 0
		}
	}
	if n < 5 || string(prefix[:5]) != "data:" {
		return decisionsImageDenied()
	}
	if prefixLen == 0 {
		return decisionsSchemaFailure()
	}
	length := n - int64(prefixLen)
	return b.takeImage(max(0, length/4*3+(length%4)*3/4-min(padding, 2)))
}

// Count the UTF-8 bytes the standard decoder will produce, including escaped
// surrogate pairs and its replacement of lone surrogates/invalid UTF-8. Only
// a two-escape window is ever decoded; an oversized string is never copied.
func (b *decisionsImageBudget) rawText(raw []byte) *Error {
	remaining := b.policy.Classifier.MaxStateBytes
	for i := 0; i < len(raw); {
		n, consumed := int64(1), 1
		if raw[i] == '\\' {
			consumed = 2
			if i+1 < len(raw) && raw[i+1] == 'u' {
				consumed = 6
				if i+12 <= len(raw) && raw[i+6] == '\\' && raw[i+7] == 'u' {
					first, _ := strconv.ParseUint(string(raw[i+2:i+6]), 16, 16)
					second, _ := strconv.ParseUint(string(raw[i+8:i+12]), 16, 16)
					if first >= 0xd800 && first <= 0xdbff && second >= 0xdc00 && second <= 0xdfff {
						consumed = 12
					}
				}
			}
			if i+consumed > len(raw) {
				return decisionsSchemaFailure()
			}
			var window [14]byte
			window[0] = '"'
			copy(window[1:], raw[i:i+consumed])
			window[consumed+1] = '"'
			var text string
			if json.Unmarshal(window[:consumed+2], &text) != nil {
				return decisionsSchemaFailure()
			}
			n = int64(len(text))
		} else if raw[i] >= utf8.RuneSelf {
			r, size := utf8.DecodeRune(raw[i:])
			consumed = size
			n = int64(utf8.RuneLen(r))
		}
		if n > remaining {
			return limitFailure(PhaseRequest, "Classifier.MaxStateBytes", b.policy.Classifier.MaxStateBytes)
		}
		remaining -= n
		i += consumed
	}
	return nil
}

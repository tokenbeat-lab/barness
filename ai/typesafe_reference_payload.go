package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/buger/jsonparser"
)

// Inspect known borrowed content before marshalJS creates independent bytes.
// This pass only accounts whole-value occurrences; the final native guard
// still owns schema, placement, visibility, private namespace and file validity.
// Custom encoders execute once at the normal encoder boundary and are checked
// there. Their output cannot be predicted by inspecting their Go representation.
func checkTypeSafeReferencePayload(ctx context.Context, body map[string]any, sizes map[string]int64, limits ClassifierImageReferencePolicy, rawLimit int64) *Error {
	count, remaining := 0, limits.MaxExpandedImageBytes
	take := func(key string) *Error {
		if len(key) > typeSafeKeyBytes {
			return nil
		} // Cannot match a legal native key.
		key = typeSafeJSONText(key)
		size, known := sizes[key]
		if !known {
			return nil
		}
		if count >= limits.MaxReferences {
			return limitFailure(PhaseRequest, "Classifier.Images.References.MaxReferences", int64(limits.MaxReferences))
		}
		count++
		if size > remaining {
			return limitFailure(PhaseRequest, "Classifier.Images.References.MaxExpandedImageBytes", limits.MaxExpandedImageBytes)
		}
		remaining -= size
		return nil
	}
	var raw func([]byte, jsonparser.ValueType, int) *Error
	raw = func(value []byte, kind jsonparser.ValueType, depth int) *Error {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return failure
		}
		if depth > maxJSONDepth {
			return limitFailure(PhaseRequest, "MaxJSONDepth", maxJSONDepth)
		}
		var failure *Error
		var err error
		switch kind {
		case jsonparser.String:
			// A key is at most 1024 bytes (six escaped bytes per character).
			// Avoid copying an unrelated large business string for comparison.
			if len(value) > 6*typeSafeKeyBytes {
				return nil
			}
			decoded, failure := typeSafeReferenceString(value)
			if failure != nil {
				return failure
			}
			return take(decoded)
		case jsonparser.Object:
			err = jsonparser.ObjectEach(value, func(_, child []byte, childKind jsonparser.ValueType, _ int) error {
				failure = raw(child, childKind, depth+1)
				if failure != nil {
					return failure
				}
				return nil
			})
		case jsonparser.Array:
			_, err = jsonparser.ArrayEach(value, func(child []byte, childKind jsonparser.ValueType, _ int, parseErr error) {
				if failure != nil {
					return
				}
				if parseErr != nil {
					failure = typeSafeRequestFailure()
					return
				}
				failure = raw(child, childKind, depth+1)
			})
		}
		if failure != nil {
			return failure
		}
		if err != nil {
			return typeSafeRequestFailure()
		}
		return nil
	}
	fields := map[reflect.Type][]imageJSONField{}
	var walk func(reflect.Value, int) *Error
	walk = func(value reflect.Value, depth int) *Error {
		if failure := contextError(ctx, PhaseRequest); failure != nil {
			return failure
		}
		if depth > maxJSONDepth {
			return limitFailure(PhaseRequest, "MaxJSONDepth", maxJSONDepth)
		}
		v, failure := imagePayloadValue(value)
		if failure != nil || !v.IsValid() {
			return failure
		}
		if v.CanInterface() {
			if _, number := v.Interface().(json.Number); number {
				return nil
			}
			if data, ok := v.Interface().(json.RawMessage); ok {
				if int64(len(data)) > rawLimit {
					return limitFailure(PhaseRequest, "MaxRequestBytes", rawLimit)
				}
				value, kind, _, err := jsonparser.Get(data)
				if err != nil {
					return typeSafeRequestFailure()
				}
				return raw(value, kind, depth)
			}
		}
		if usesCustomJSON(v) {
			return nil
		}
		switch v.Kind() {
		case reflect.String:
			return take(v.String())
		case reflect.Map:
			it := v.MapRange()
			for it.Next() {
				if failure := walk(it.Value(), depth+1); failure != nil {
					return failure
				}
			}
		case reflect.Slice, reflect.Array:
			if byteSliceEncodesBase64(v) {
				// JSON encodes ordinary byte slices as Base64 strings. Compare
				// only bounded possible keys; large business bytes need no copy.
				if v.Len() > typeSafeKeyBytes/4*3 {
					return nil
				}
				var data [typeSafeKeyBytes / 4 * 3]byte
				for i := 0; i < v.Len(); i++ {
					data[i] = byte(v.Index(i).Uint())
				}
				var encoded [typeSafeKeyBytes]byte
				base64.StdEncoding.Encode(encoded[:], data[:v.Len()])
				return take(string(encoded[:base64.StdEncoding.EncodedLen(v.Len())]))
			}
			for i := 0; i < v.Len(); i++ {
				if failure := walk(v.Index(i), depth+1); failure != nil {
					return failure
				}
			}
		case reflect.Struct:
			selected, ok := fields[v.Type()]
			if !ok {
				selected = imageJSONFields(v.Type())
				fields[v.Type()] = selected
			}
			for _, field := range selected {
				value := field.value(v)
				text, quoted, failure := field.quotedString(value, typeSafeKeyBytes)
				if failure != nil {
					return failure
				}
				if quoted {
					if failure := take(text); failure != nil {
						return failure
					}
					continue
				}
				if failure := walk(value, depth+1); failure != nil {
					return failure
				}
			}
		}
		return nil
	}
	for _, name := range []string{"state", "questions"} {
		if failure := walk(reflect.ValueOf(body[name]), 0); failure != nil {
			return failure
		}
	}
	return nil
}

// encoding/json replaces each invalid UTF-8 byte separately. ToValidUTF8
// merges runs and would assign different image identities to the same JSON.
func typeSafeJSONText(value string) string {
	if utf8.ValidString(value) {
		return value
	}
	return strings.Map(func(r rune) rune { return r }, value)
}

// Use the native decoder's Unicode semantics for bounded raw strings too.
// In particular, an unpaired surrogate is U+FFFD, not a malformed request.
func typeSafeReferenceString(value []byte) (string, *Error) {
	var literal [6*typeSafeKeyBytes + 2]byte
	literal[0] = '"'
	copy(literal[1:], value)
	literal[len(value)+1] = '"'
	var text string
	if json.Unmarshal(literal[:len(value)+2], &text) != nil {
		return "", typeSafeRequestFailure()
	}
	return text, nil
}

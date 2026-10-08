package ai

import (
	"encoding"
	"encoding/json"
	"reflect"
)

// A conservative size from ordinary JSON values, without encoding strings,
// cloning maps or collecting keys. Raw JSON counts its supplied bytes, as at
// the public classifier input boundary. Unknown/custom encodings are checked
// against the final body limit before decoding; they cannot be predicted here.
func jsonPayloadSize(value any, limit int64) (int64, bool) {
	size := int64(0)
	take := func(n int64) bool {
		if n > limit-size {
			return false
		}
		size += n
		return true
	}
	var walk func(reflect.Value, int) bool
	walk = func(v reflect.Value, depth int) bool {
		if depth > maxJSONDepth {
			return false
		}
		if !v.IsValid() {
			return take(4)
		}
		if v.Kind() == reflect.Interface {
			if v.IsNil() {
				return take(4)
			}
			return walk(v.Elem(), depth+1)
		}
		if v.CanInterface() {
			switch x := v.Interface().(type) {
			case json.RawMessage:
				return take(int64(len(x)))
			case *json.RawMessage:
				if x == nil {
					return take(4)
				}
				return take(int64(len(*x)))
			case ClassifierQuestion:
				if v.Kind() == reflect.Pointer {
					if v.IsNil() {
						return take(4)
					}
					return walk(v.Elem(), depth+1)
				}
				n, ok := classifierQuestionSize(x, limit-size)
				return ok && take(n)
			case json.Number:
				return take(int64(len(x)))
			}
		}
		// Known raw/question sizes above must precede this fallback. The JSON
		// encoder also uses pointer methods on addressable slice/array elements.
		if usesCustomJSON(v) {
			return take(1)
		}
		switch v.Kind() {
		case reflect.String:
			return take(2) && take(int64(v.Len()))
		case reflect.Bool:
			return take(4)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return take(1)
		case reflect.Map:
			if !take(2) || !take(int64(max(0, v.Len()-1))) {
				return false
			}
			entries := v.MapRange()
			for entries.Next() {
				key := entries.Key()
				// encoding/json uses string map keys directly, even when their
				// named type implements TextMarshaler.
				keyFits := false
				if key.Kind() == reflect.String {
					keyFits = take(2) && take(int64(key.Len()))
				} else {
					keyFits = walk(key, depth+1)
				}
				if !keyFits || !take(1) || !walk(entries.Value(), depth+1) {
					return false
				}
			}
		case reflect.Slice, reflect.Array:
			if !take(2) {
				return false
			}
			if v.Type().Elem().Kind() == reflect.Uint8 {
				return take(int64(v.Len()))
			}
			if !take(int64(max(0, v.Len()-1))) {
				return false
			}
			for i := 0; i < v.Len(); i++ {
				if !walk(v.Index(i), depth+1) {
					return false
				}
			}
		case reflect.Pointer:
			if !v.IsNil() {
				return walk(v.Elem(), depth+1)
			}
		default:
			// Struct field selection and custom encodings belong to the JSON
			// encoder. Guessing their shape can reject valid small payloads.
			return take(1)
		}
		return true
	}
	ok := walk(reflect.ValueOf(value), 0)
	return size, ok
}

// Both request sizing and image preflight must honor the encoder's method
// selection, including pointer methods on addressable slice elements.
func usesCustomJSON(v reflect.Value) bool {
	custom := func(value reflect.Value) bool {
		if value.CanInterface() {
			switch value.Interface().(type) {
			case json.Marshaler, encoding.TextMarshaler:
				return true
			}
		}
		return false
	}
	return custom(v) || (v.Kind() != reflect.Pointer && v.CanAddr() && custom(v.Addr()))
}

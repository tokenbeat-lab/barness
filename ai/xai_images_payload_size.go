package ai

import (
	"encoding/json"
	"reflect"
)

// The generic JSON estimate deliberately treats ordinary structs as opaque.
// Image resources are a second real use of typed resource containers: count
// their selected fields and their surrounding request before invoking Host
// marshalers. Custom encodings remain unknown until final byte validation.
func xaiImagesPayloadFits(body map[string]any, limit int64) bool {
	remaining := limit - 2 - int64(max(0, len(body)-1))
	for key, value := range body {
		remaining -= jsonStringBytes(key) + 1
		if remaining < 0 {
			return false
		}
		size, ok := jsonPayloadSize(value, remaining)
		if key == "image" || key == "images" {
			size, ok = xaiImageResourceSize(reflect.ValueOf(value), remaining, 0)
		}
		if !ok || size > remaining {
			return false
		}
		remaining -= size
	}
	return true
}
func xaiImageResourceSize(value reflect.Value, limit int64, depth int) (int64, bool) {
	if depth > maxJSONDepth {
		return 0, false
	}
	v, failure := imagePayloadValue(value)
	if failure != nil {
		return 0, false
	}
	if !v.IsValid() {
		return 4, limit >= 4
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			return int64(len(raw)), int64(len(raw)) <= limit
		}
	}
	if usesCustomJSON(v) {
		return 1, limit >= 1
	}
	size := int64(0)
	take := func(n int64) bool {
		if n > limit-size {
			return false
		}
		size += n
		return true
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if byteSliceEncodesBase64(v) {
			return jsonPayloadValueSize(v, limit)
		}
		if !take(2 + int64(max(0, v.Len()-1))) {
			return 0, false
		}
		for i := 0; i < v.Len(); i++ {
			n, ok := xaiImageResourceSize(v.Index(i), limit-size, depth+1)
			if !ok || !take(n) {
				return 0, false
			}
		}
	case reflect.Struct:
		if !take(2) {
			return 0, false
		}
		count := 0
		for _, field := range imageJSONFields(v.Type()) {
			part := field.value(v)
			if !part.IsValid() {
				continue
			}
			n, ok := xaiImageResourceSize(part, limit-size, depth+1)
			if !ok || !take(jsonStringBytes(field.name)+1+n) {
				return 0, false
			}
			count++
		}
		if !take(int64(max(0, count-1))) {
			return 0, false
		}
	case reflect.Map:
		if !take(2 + int64(max(0, v.Len()-1))) {
			return 0, false
		}
		fields := v.MapRange()
		for fields.Next() {
			key := fields.Key()
			keySize := int64(1)
			if key.Kind() == reflect.String {
				keySize = jsonStringBytes(key.String())
			}
			if !take(keySize + 1) {
				return 0, false
			}
			n, ok := xaiImageResourceSize(fields.Value(), limit-size, depth+1)
			if !ok || !take(n) {
				return 0, false
			}
		}
	default:
		return jsonPayloadValueSize(v, limit)
	}
	return size, true
}

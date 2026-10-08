package ai

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Check borrowed callback values before encoding or converting them into
// independent question records. Exact encoded sizes and schema are checked
// afterwards; custom marshalers remain responsible for their own execution.
func (p *ClassifierPolicy) checkPayload(body map[string]any, requestLimit int64) *Error {
	invalid := func() *Error {
		return newError(CodeCallbackFailed, PhaseRequest, "payload callback exceeded classifier input limits")
	}
	if _, ok := classifierPayloadSize(body["state"], p.MaxStateBytes); !ok {
		return invalid()
	}
	questions := reflect.ValueOf(body["questions"])
	if questions.IsValid() && questions.Kind() == reflect.Map {
		if questions.Len() > p.MaxQuestions {
			return invalid()
		}
		entries := questions.MapRange()
		for entries.Next() {
			if _, ok := classifierPayloadSize(entries.Value().Interface(), p.MaxQuestionBytes); !ok {
				return invalid()
			}
		}
	}
	if _, ok := classifierPayloadSize(body, requestLimit); !ok {
		return limitFailure(PhaseRequest, "MaxRequestBytes", requestLimit)
	}
	return nil
}

// A conservative size from ordinary JSON values, without encoding strings,
// cloning maps or collecting keys. Raw JSON counts its supplied bytes, as at
// the public classifier input boundary. Unknown/custom encodings are checked
// against the final body limit before decoding; they cannot be predicted here.
func classifierPayloadSize(value any, limit int64) (int64, bool) {
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
			case ClassifierQuestion:
				n, ok := classifierQuestionSize(x, limit-size)
				return ok && take(n)
			case json.Number:
				return take(int64(len(x)))
			case json.Marshaler:
				return true
			}
		}
		switch v.Kind() {
		case reflect.String:
			return take(int64(v.Len()))
		case reflect.Map:
			entries := v.MapRange()
			for entries.Next() {
				if !walk(entries.Key(), depth+1) || !walk(entries.Value(), depth+1) {
					return false
				}
			}
		case reflect.Slice, reflect.Array:
			if v.Type().Elem().Kind() == reflect.Uint8 {
				return take(int64(v.Len()))
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
		case reflect.Struct:
			typ := v.Type()
			for i := 0; i < v.NumField(); i++ {
				field := typ.Field(i)
				tag := field.Tag.Get("json")
				if field.PkgPath != "" || tag == "-" {
					continue
				}
				if (strings.Contains(tag, ",omitempty") || strings.Contains(tag, ",omitzero")) && v.Field(i).IsZero() {
					continue
				}
				// Field names can be omitted by promotion; counting only their values
				// stays a lower bound regardless of encoding/json's field selection.
				if !walk(v.Field(i), depth+1) {
					return false
				}
			}
		}
		return true
	}
	ok := walk(reflect.ValueOf(value), 0)
	return size, ok
}

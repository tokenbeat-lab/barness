package ai

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/buger/jsonparser"
)

// Bound known resource containers before the encoder executes Host methods.
// Their final JSON bytes still pass the independent native schema guard.
func (p *ResourcePolicy) checkXAIResourcePayload(value reflect.Value, many bool, caps ImageCapabilities) *Error {
	v, failure := imagePayloadValue(value)
	if failure != nil {
		return failure
	}
	if !v.IsValid() {
		return imageInputFailure(PhaseRequest, "Imagine resources must not be null")
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			if !many {
				return p.checkRawXAIResource(raw)
			}
			limit := min(5, caps.MaxReferenceImages, p.Image.MaxInputImages)
			if _, _, _, err := jsonparser.Get(raw, "["+strconv.Itoa(limit)+"]"); err == nil {
				return p.checkXAIImageCount(limit+1, caps, PhaseRequest)
			}
			var failure *Error
			count := 0
			_, err := jsonparser.ArrayEach(raw, func(part []byte, kind jsonparser.ValueType, _ int, err error) {
				count++
				if failure == nil {
					if err != nil || kind != jsonparser.Object {
						failure = imageInputFailure(PhaseRequest, "Imagine resources require objects")
					} else {
						failure = p.checkRawXAIResource(part)
					}
				}
			})
			if failure != nil {
				return failure
			}
			if err != nil {
				return imageInputFailure(PhaseRequest, "Imagine resources require an array")
			}
			return p.checkXAIImageCount(count, caps, PhaseRequest)
		}
	}
	if usesCustomJSON(v) {
		return nil
	}
	if many {
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return imageInputFailure(PhaseRequest, "Imagine images require an array")
		}
		if failure := p.checkXAIImageCount(v.Len(), caps, PhaseRequest); failure != nil {
			return failure
		}
		for i := 0; i < v.Len(); i++ {
			if failure := p.checkXAIResourcePayload(v.Index(i), false, caps); failure != nil {
				return failure
			}
		}
		return nil
	}
	check := func(key string, value reflect.Value) *Error {
		if key == "file_id" {
			return imageAuthorityFailure()
		}
		if key == "url" {
			return p.checkImageURLPayload(value, true)
		}
		return nil
	}
	switch v.Kind() {
	case reflect.Map:
		fields := v.MapRange()
		for fields.Next() {
			if fields.Key().Kind() == reflect.String {
				if failure := check(fields.Key().String(), fields.Value()); failure != nil {
					return failure
				}
			} else {
				if failure := p.checkImageURLPayload(fields.Value(), false); failure != nil {
					return failure
				}
			}
		}
	case reflect.Struct:
		for _, key := range []string{"url", "file_id"} {
			field, failure := imageStructField(v, key)
			if failure != nil {
				return failure
			}
			if field.IsValid() {
				if failure := check(key, field); failure != nil {
					return failure
				}
			}
		}
	default:
		return imageInputFailure(PhaseRequest, "Imagine resource requires an object")
	}
	return nil
}
func (p *ResourcePolicy) checkRawXAIResource(raw []byte) *Error {
	var failure *Error
	err := jsonparser.ObjectEach(raw, func(key, value []byte, kind jsonparser.ValueType, _ int) error {
		if string(key) == "file_id" {
			return imageAuthorityFailure()
		}
		if string(key) == "url" {
			if kind != jsonparser.String {
				return imageInputFailure(PhaseRequest, "Imagine url must be a string")
			}
			failure = p.checkRawInlineURL(value, true)
			return failure
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if ae, ok := err.(*Error); ok {
		return ae
	}
	if err != nil {
		return imageInputFailure(PhaseRequest, "Imagine resource requires an object")
	}
	return nil
}

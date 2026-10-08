package ai

import (
	"encoding/json"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

// Only the small envelope is encoded for this estimate. Validated base64 is
// ASCII without JSON escapes; substituting one character preserves every
// delimiter and field overhead while avoiding any proportional image copy.
func (p *ResourcePolicy) checkGoogleImagesEncoding(wire googleImagesRequest) *Error {
	wire.Input = slices.Clone(wire.Input)
	extra := int64(0)
	for i := 1; i < len(wire.Input); i++ {
		extra += int64(len(wire.Input[i].Data)) - 1
		wire.Input[i].Data = "A"
	}
	small, err := json.Marshal(wire)
	if err != nil {
		return newError(CodeInvalidRequest, PhaseRequest, "Google image request could not be encoded")
	}
	size := int64(len(small)) + extra
	if size > p.MaxRequestBytes {
		return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
	}
	if size > googleImagesRequestLimit {
		return limitFailure(PhaseRequest, "Google.MaxInlineRequestBytes", googleImagesRequestLimit)
	}
	return nil
}

// Known host containers and raw JSON are bounded before hooks freeze them.
// Custom marshalers are checked on their final bytes, without running twice.
func (p *ResourcePolicy) checkGoogleInputPayload(input any, caps ImageCapabilities) *Error {
	remaining := min(p.MaxRequestBytes, googleImagesRequestLimit)
	v, failure := imagePayloadValue(reflect.ValueOf(input))
	if failure != nil || !v.IsValid() {
		return failure // Final schema check rejects missing/null input.
	}
	if raw, ok := v.Interface().(json.RawMessage); ok {
		limit := min(p.Image.MaxInputImages, caps.MaxReferenceImages, googleImagesReferenceLimit)
		if _, _, _, err := jsonparser.Get(raw, "["+strconv.Itoa(limit+1)+"]"); err == nil {
			return p.checkGoogleImageCount(limit+1, caps, PhaseRequest)
		}
		var problem *Error
		index := 0
		_, err := jsonparser.ArrayEach(raw, func(part []byte, kind jsonparser.ValueType, _ int, err error) {
			if problem == nil && err == nil && kind == jsonparser.Object {
				problem = p.checkRawGoogleInput(part, index, &remaining)
			}
			index++
		})
		if problem != nil {
			return problem
		}
		if err != nil {
			return imageInputFailure(PhaseRequest, "Google image input must be an array")
		}
		return nil
	}
	if usesCustomJSON(v) || (v.Kind() != reflect.Slice && v.Kind() != reflect.Array) {
		return nil
	}
	if failure := p.checkGoogleImageCount(max(0, v.Len()-1), caps, PhaseRequest); failure != nil {
		return failure
	}
	for i := 0; i < v.Len(); i++ {
		part, failure := imagePayloadValue(v.Index(i))
		if failure != nil {
			return failure
		}
		if !part.IsValid() {
			continue
		}
		if raw, ok := part.Interface().(json.RawMessage); ok {
			if failure := p.checkRawGoogleInput(raw, i, &remaining); failure != nil {
				return failure
			}
			continue
		}
		if usesCustomJSON(part) {
			continue
		}
		if part.Kind() == reflect.Map && part.Type().Key().Kind() == reflect.String {
			fields := part.MapRange()
			for fields.Next() {
				if fields.Key().String() == "data" {
					if failure := p.checkGoogleDataPayload(fields.Value(), &remaining); failure != nil {
						return failure
					}
				}
			}
		}
		if part.Kind() == reflect.Struct {
			// Ordinary tagged data fields have known sizes. Promoted/custom
			// encodings still undergo the independent final schema validation.
			for j := 0; j < part.NumField(); j++ {
				field := part.Type().Field(j)
				key, _, _ := strings.Cut(field.Tag.Get("json"), ",")
				if field.PkgPath == "" && key == "data" {
					if failure := p.checkGoogleDataPayload(part.Field(j), &remaining); failure != nil {
						return failure
					}
				}
			}
		}
	}
	return nil
}

func (p *ResourcePolicy) checkGoogleDataPayload(value reflect.Value, remaining *int64) *Error {
	v, failure := imagePayloadValue(value)
	if failure != nil || !v.IsValid() {
		return failure
	}
	if raw, ok := v.Interface().(json.RawMessage); ok {
		data, kind, _, err := jsonparser.Get(raw)
		if err != nil || kind != jsonparser.String {
			return imageInputFailure(PhaseRequest, "inline image data must be a string")
		}
		if failure := p.checkRawGoogleData(data); failure != nil {
			return failure
		}
		return p.takeGoogleInputBytes(int64(len(data)), remaining)
	}
	if !usesCustomJSON(v) && v.Kind() == reflect.String {
		if base64Size(v.String()) > p.MaxImageBytes {
			return limitFailure(PhaseRequest, "MaxImageBytes", p.MaxImageBytes)
		}
		return p.takeGoogleInputBytes(int64(v.Len()), remaining)
	}
	return nil
}

func (p *ResourcePolicy) checkRawGoogleInput(raw []byte, index int, remaining *int64) *Error {
	allowed := []string{"type", "text"}
	if index > 0 {
		allowed = []string{"type", "mime_type", "data"}
	}
	err := jsonparser.ObjectEach(raw, func(key, value []byte, kind jsonparser.ValueType, _ int) error {
		if failure := googleImagesField(string(key), allowed); failure != nil {
			return failure
		}
		if string(key) == "data" && kind == jsonparser.String {
			if failure := p.checkRawGoogleData(value); failure != nil {
				return failure
			}
			return p.takeGoogleInputBytes(int64(len(value)), remaining)
		}
		return nil
	})
	if failure, ok := err.(*Error); ok {
		return failure
	}
	if err != nil {
		return imageInputFailure(PhaseRequest, "invalid inline Google image input")
	}
	return nil
}

func (p *ResourcePolicy) checkRawGoogleData(raw []byte) *Error {
	// Count unescaped base64 characters and padding with fixed memory. This is
	// intentionally separate from OpenAI's data-URL prefix/URL length policy.
	n, padding := int64(0), int64(0)
	var escaped [6]byte
	for i := 0; i < len(raw); {
		b := raw[i]
		i++
		if b == '\\' {
			length := 2
			if i < len(raw) && raw[i] == 'u' {
				length = 6
			}
			if i-1+length > len(raw) {
				return imageInputFailure(PhaseRequest, "invalid image data escape")
			}
			decoded, err := jsonparser.Unescape(raw[i-1:i-1+length], escaped[:])
			if err != nil || len(decoded) != 1 {
				return imageInputFailure(PhaseRequest, "inline image data must contain ASCII")
			}
			b = decoded[0]
			i += length - 1
		}
		n++
		if b == '=' {
			padding++
		} else {
			padding = 0
		}
	}
	if n*3/4-min(padding, 2) > p.MaxImageBytes {
		return limitFailure(PhaseRequest, "MaxImageBytes", p.MaxImageBytes)
	}
	return nil
}

func (p *ResourcePolicy) takeGoogleInputBytes(size int64, remaining *int64) *Error {
	if size > *remaining {
		if p.MaxRequestBytes < googleImagesRequestLimit {
			return limitFailure(PhaseRequest, "MaxRequestBytes", p.MaxRequestBytes)
		}
		return limitFailure(PhaseRequest, "Google.MaxInlineRequestBytes", googleImagesRequestLimit)
	}
	*remaining -= size
	return nil
}

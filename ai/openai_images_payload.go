package ai

import (
	"bytes"
	"encoding"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
)

// Struct fields are not traversed by the generic request size estimate.
// Bound their pointer/interface expansion too, including cyclic host values.
func imagePayloadValue(v reflect.Value) (reflect.Value, *Error) {
	for depth := 0; v.IsValid(); depth++ {
		if depth > maxJSONDepth {
			return reflect.Value{}, limitFailure(PhaseRequest, "MaxJSONDepth", maxJSONDepth)
		}
		if v.Kind() == reflect.Interface {
			if v.IsNil() {
				return reflect.Value{}, nil
			}
			v = v.Elem()
			continue
		}
		if v.Kind() != reflect.Pointer {
			return v, nil
		}
		if v.IsNil() {
			return reflect.Value{}, nil
		}
		if v.Type().Elem() != reflect.TypeFor[json.RawMessage]() && usesCustomJSON(v) {
			return v, nil
		}
		v = v.Elem()
	}
	return v, nil
}

func (p *ResourcePolicy) checkImageReferencePayload(value reflect.Value, mask bool, caps ImageCapabilities) *Error {
	v, failure := imagePayloadValue(value)
	if failure != nil {
		return failure
	}
	if !v.IsValid() {
		return imageAuthorityFailure()
	}
	if raw, ok := v.Interface().(json.RawMessage); ok {
		return p.checkRawImageReferences(raw, mask, caps)
	}
	if usesCustomJSON(v) {
		return nil
	} // Its representation belongs to the encoder and final guard.
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return imageInputFailure(PhaseRequest, "images must be an array")
	}
	if v.Len() == 0 {
		return imageAuthorityFailure()
	}
	if failure := p.checkFinalImageCount(v.Len(), mask, caps); failure != nil {
		return failure
	}
	for i := 0; i < v.Len(); i++ {
		if failure := p.checkInlineImagePayload(v.Index(i)); failure != nil {
			return failure
		}
	}
	return nil
}

func (p *ResourcePolicy) checkInlineImagePayload(value reflect.Value) *Error {
	v, failure := imagePayloadValue(value)
	if failure != nil {
		return failure
	}
	if !v.IsValid() {
		return imageInputFailure(PhaseRequest, "inline image must be an object")
	}
	if raw, ok := v.Interface().(json.RawMessage); ok {
		return p.checkRawInlineImage(raw)
	}
	if usesCustomJSON(v) {
		return nil
	}
	checkField := func(key string, value reflect.Value) *Error {
		if key == "file_id" {
			return imageAuthorityFailure()
		}
		if key != "image_url" {
			return imageInputFailure(PhaseRequest, "unsupported inline image field")
		}
		return p.checkImageURLPayload(value, true)
	}
	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			if v.Type().Key().Implements(reflect.TypeFor[encoding.TextMarshaler]()) {
				// Key encoding is deferred, but ordinary URL values still have
				// known lengths. Do not execute a host key marshaler twice.
				fields := v.MapRange()
				for fields.Next() {
					if failure := p.checkImageURLPayload(fields.Value(), false); failure != nil {
						return failure
					}
				}
				return nil
			}
			return imageInputFailure(PhaseRequest, "inline image requires string keys")
		}
		fields := v.MapRange()
		for fields.Next() {
			if failure := checkField(fields.Key().String(), fields.Value()); failure != nil {
				return failure
			}
		}
	case reflect.Struct:
		field, failure := inlineStructURL(v)
		if failure != nil {
			return failure
		}
		if field.IsValid() {
			return checkField("image_url", field)
		}
	default:
		return imageInputFailure(PhaseRequest, "inline image must be an object")
	}
	return nil
}

// Raw JSON strings have known URL lengths, unlike host-defined marshalers.
// requireInline validates a known image_url key; custom key names are frozen
// by the encoder later, but their ordinary values retain capacity checks.
func (p *ResourcePolicy) checkImageURLPayload(value reflect.Value, requireInline bool) *Error {
	field, failure := imagePayloadValue(value)
	if failure != nil {
		return failure
	}
	if !field.IsValid() {
		return nil
	}
	if field.CanInterface() {
		if raw, ok := field.Interface().(json.RawMessage); ok {
			url, kind, _, err := jsonparser.Get(raw)
			if err != nil || kind != jsonparser.String {
				return imageInputFailure(PhaseRequest, "image_url must be a string")
			}
			return p.checkRawInlineURL(url, requireInline)
		}
	}
	if !usesCustomJSON(field) && field.Kind() == reflect.String {
		if requireInline {
			return p.checkInlineURL(field.String())
		}
		return p.checkInlineURLSize(field.String())
	}
	return nil // Unknown field encodings are validated after freezing.
}

func inlineStructURL(v reflect.Value) (reflect.Value, *Error) {
	// JSON selects promoted fields by their JSON name, independently of Go
	// name shadowing. Prefer the shallowest tagged resource field; equal-depth
	// duplicates cannot encode an unambiguous image_url.
	var index []int
	duplicate := false
	visiting := map[reflect.Type]bool{}
	var walk func(reflect.Type, []int)
	walk = func(t reflect.Type, path []int) {
		if visiting[t] {
			return
		}
		visiting[t] = true
		defer delete(visiting, t)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			key, _, _ := strings.Cut(field.Tag.Get("json"), ",")
			next := append(append([]int(nil), path...), i)
			if field.PkgPath == "" && key == "image_url" {
				if index == nil || len(next) < len(index) {
					index = next
					duplicate = false
				} else if len(next) == len(index) {
					duplicate = true
				}
			}
			embedded := field.Type
			if embedded.Kind() == reflect.Pointer {
				embedded = embedded.Elem()
			}
			if field.Anonymous && key == "" && embedded.Kind() == reflect.Struct {
				walk(embedded, next)
			}
		}
	}
	walk(v.Type(), nil)
	if duplicate {
		return reflect.Value{}, imageInputFailure(PhaseRequest, "ambiguous inline image fields")
	}
	if index == nil {
		return reflect.Value{}, nil
	}
	field, err := v.FieldByIndexErr(index)
	if err != nil {
		return reflect.Value{}, nil
	} // A nil anonymous pointer is omitted by encoding/json.
	return field, nil
}

func (p *ResourcePolicy) checkRawImageReferences(raw json.RawMessage, mask bool, caps ImageCapabilities) *Error {
	// jsonparser returns views of the supplied bytes. Probe the first excess
	// index before walking entries, without decoding or allocating URLs.
	limit := min(openAIReferenceLimit, caps.MaxReferenceImages, p.Image.MaxInputImages)
	if mask {
		limit = min(limit, p.Image.MaxInputImages-1)
	}
	if _, kind, _, err := jsonparser.Get(raw, "["+strconv.Itoa(limit)+"]"); err == nil && kind != jsonparser.NotExist {
		return p.checkFinalImageCount(limit+1, mask, caps)
	}
	n := 0
	var failure *Error
	_, err := jsonparser.ArrayEach(raw, func(value []byte, kind jsonparser.ValueType, _ int, err error) {
		if failure != nil {
			return
		}
		n++
		if err != nil || kind != jsonparser.Object {
			failure = imageInputFailure(PhaseRequest, "inline image must be an object")
			return
		}
		failure = p.checkRawInlineImage(value)
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return imageInputFailure(PhaseRequest, "images must be a valid array")
	}
	if n == 0 {
		return imageAuthorityFailure()
	}
	return p.checkFinalImageCount(n, mask, caps)
}

func (p *ResourcePolicy) checkRawInlineImage(raw []byte) *Error {
	var url []byte
	err := jsonparser.ObjectEach(raw, func(key, value []byte, kind jsonparser.ValueType, _ int) error {
		if string(key) == "file_id" {
			return imageAuthorityFailure()
		}
		if string(key) != "image_url" || url != nil || kind != jsonparser.String {
			return imageInputFailure(PhaseRequest, "inline image requires one image_url string")
		}
		url = value
		return nil
	})
	if failure, ok := err.(*Error); ok {
		return failure
	}
	if err != nil || url == nil {
		return imageInputFailure(PhaseRequest, "inline image must be an object")
	}
	return p.checkRawInlineURL(url, true)
}

func (p *ResourcePolicy) checkRawInlineURL(raw []byte, requireInline bool) *Error {
	// Inline URLs contain ASCII MIME/base64. Count decoded characters using
	// the existing JSON unescaper in a fixed buffer; retain only the prefix
	// and padding. Even escaped multi-MiB URLs require no proportional copy.
	var prefix [160]byte
	var escaped [6]byte
	n, padding := int64(0), int64(0)
	for i := 0; i < len(raw); {
		b := raw[i]
		i++
		if b == '\\' {
			length := 2
			if i < len(raw) && raw[i] == 'u' {
				length = 6
			}
			if i-1+length > len(raw) {
				return imageInputFailure(PhaseRequest, "invalid inline URL escape")
			}
			decoded, err := jsonparser.Unescape(raw[i-1:i-1+length], escaped[:])
			if err != nil || len(decoded) != 1 {
				return imageInputFailure(PhaseRequest, "inline image URLs must contain ASCII")
			}
			b = decoded[0]
			i += length - 1
		}
		if b >= 128 {
			return imageInputFailure(PhaseRequest, "inline image URLs must contain ASCII")
		}
		if n < int64(len(prefix)) {
			prefix[n] = b
		}
		n++
		if n > openAIInlineURLLimit {
			return limitFailure(PhaseRequest, "OpenAI.MaxInlineImageURLCharacters", openAIInlineURLLimit)
		}
		if b == '=' {
			padding++
		} else {
			padding = 0
		}
	}
	header := prefix[:min(n, int64(len(prefix)))]
	if !bytes.HasPrefix(header, []byte("data:")) {
		if !requireInline {
			return nil
		}
		return imageAuthorityFailure()
	}
	index := bytes.Index(header, []byte(";base64,"))
	if index < 5 || !isImageMediaType(string(header[5:index])) {
		if !requireInline {
			return nil
		}
		return imageInputFailure(PhaseRequest, "input image must be an inline base64 data URL")
	}
	dataLength := n - int64(index+len(";base64,"))
	if dataLength*3/4-min(padding, 2) > p.MaxImageBytes {
		return limitFailure(PhaseRequest, "MaxImageBytes", p.MaxImageBytes)
	}
	return nil
}

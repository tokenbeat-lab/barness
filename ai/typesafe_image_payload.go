package ai

import (
	"encoding/json"
	"reflect"
	"strconv"

	"github.com/buger/jsonparser"
)

// Inspect borrowed native assets before serialization. Unlike Decisions' URL
// content tree, this boundary is a source map of raw Base64 files. The two
// parsers deliberately evolve independently with their protocol schemas.
type typeSafeImageBudget struct {
	policy    *ResourcePolicy
	remaining int64
	count     int
	sizes     map[string]int64
}

func (c *Client) checkTypeSafeAssetPayload(value any) (map[string]int64, *Error) {
	if c.policy.Classifier.Images == nil {
		return nil, newError(CodeTenantDenied, PhaseRequest, "Jev images are not authorized")
	}
	b := typeSafeImageBudget{policy: &c.policy, remaining: c.policy.Classifier.Images.MaxTotalImageBytes, sizes: map[string]int64{}}
	failure := b.assets(value)
	return b.sizes, failure
}

func (b *typeSafeImageBudget) assets(value any) *Error {
	v, failure := imagePayloadValue(reflect.ValueOf(value))
	if failure != nil || !v.IsValid() {
		return failure
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			return b.rawAssets(raw)
		}
	}
	if usesCustomJSON(v) {
		return nil
	} // Closed final decoder checks unknown shapes/encoders.
	if v.Kind() == reflect.Struct {
		for _, field := range imageJSONFields(v.Type()) {
			before, count := b.remaining, b.count
			if failure := b.asset(field.value(v)); failure != nil {
				return failure
			}
			if b.count != count {
				b.sizes[field.name] = before - b.remaining
			}
		}
		return nil
	}
	if v.Kind() != reflect.Map {
		return nil
	}
	if v.Len() > b.policy.Classifier.Images.MaxImages {
		return limitFailure(PhaseRequest, "Classifier.Images.MaxImages", int64(b.policy.Classifier.Images.MaxImages))
	}
	it := v.MapRange()
	for it.Next() {
		before, count := b.remaining, b.count
		if failure := b.asset(it.Value()); failure != nil {
			return failure
		}
		if b.count != count {
			key := it.Key()
			switch key.Kind() {
			case reflect.String:
				b.sizes[key.String()] = before - b.remaining
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if !usesCustomJSON(key) {
					b.sizes[strconv.FormatInt(key.Int(), 10)] = before - b.remaining
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
				if !usesCustomJSON(key) {
					b.sizes[strconv.FormatUint(key.Uint(), 10)] = before - b.remaining
				}
			}
		}
	}
	return nil
}

func (b *typeSafeImageBudget) take(size int64) *Error {
	b.count++
	ip := b.policy.Classifier.Images
	if b.count > ip.MaxImages {
		return limitFailure(PhaseRequest, "Classifier.Images.MaxImages", int64(ip.MaxImages))
	}
	if size > b.policy.MaxImageBytes {
		return limitFailure(PhaseRequest, "MaxImageBytes", b.policy.MaxImageBytes)
	}
	if size > typeSafeImageBytes {
		return limitFailure(PhaseRequest, "Jev.MaxImageBytes", typeSafeImageBytes)
	}
	if size > b.remaining {
		return limitFailure(PhaseRequest, "Classifier.Images.MaxTotalImageBytes", ip.MaxTotalImageBytes)
	}
	b.remaining -= size
	return nil
}

func (b *typeSafeImageBudget) asset(value reflect.Value) *Error {
	v, failure := imagePayloadValue(value)
	if failure != nil || !v.IsValid() {
		return failure
	}
	if v.CanInterface() {
		if raw, ok := v.Interface().(json.RawMessage); ok {
			return b.rawAsset(raw)
		}
	}
	if usesCustomJSON(v) {
		return nil
	}
	var data reflect.Value
	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() == reflect.String {
			data = v.MapIndex(reflect.ValueOf("data_base64").Convert(v.Type().Key()))
		}
	case reflect.Struct:
		data, failure = imageStructField(v, "data_base64")
	}
	if failure != nil {
		return failure
	}
	data, failure = imagePayloadValue(data)
	if failure != nil || !data.IsValid() {
		return failure
	}
	if data.CanInterface() {
		if raw, ok := data.Interface().(json.RawMessage); ok {
			part, kind, _, err := jsonparser.Get(raw)
			if err == nil && kind == jsonparser.String {
				return b.rawData(part)
			}
			return nil
		}
	}
	if !usesCustomJSON(data) && data.Kind() == reflect.String {
		return b.take(base64Size(data.String()))
	}
	return nil
}

func (b *typeSafeImageBudget) rawAssets(raw []byte) *Error {
	var failure *Error
	err := jsonparser.ObjectEach(raw, func(key, value []byte, _ jsonparser.ValueType, _ int) error {
		before, count := b.remaining, b.count
		failure = b.rawAsset(value)
		if failure != nil {
			return failure
		}
		if b.count != count {
			b.sizes[string(key)] = before - b.remaining
		}
		return nil
	})
	if failure != nil {
		return failure
	}
	if err != nil {
		return typeSafeRequestFailure()
	}
	return nil
}

func (b *typeSafeImageBudget) rawAsset(raw []byte) *Error {
	data, kind, _, err := jsonparser.Get(raw, "data_base64")
	if err != nil || kind != jsonparser.String {
		return nil
	} // Final schema owns invalid fields.
	return b.rawData(data)
}

func (b *typeSafeImageBudget) rawData(raw []byte) *Error {
	var escaped [6]byte
	n, padding := int64(0), int64(0)
	for i := 0; i < len(raw); {
		ch := raw[i]
		i++
		if ch == '\\' {
			length := 2
			if i < len(raw) && raw[i] == 'u' {
				length = 6
			}
			if i-1+length > len(raw) {
				return typeSafeRequestFailure()
			}
			decoded, err := jsonparser.Unescape(raw[i-1:i-1+length], escaped[:])
			if err != nil || len(decoded) != 1 {
				return typeSafeRequestFailure()
			}
			ch = decoded[0]
			i += length - 1
		}
		n++
		if ch == '=' {
			padding++
		} else {
			padding = 0
		}
	}
	return b.take(n/4*3 - padding)
}

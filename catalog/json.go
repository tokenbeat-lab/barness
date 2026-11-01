package catalog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// jsonShape validates exact field spelling and non-null scalar/struct fields.
// encoding/json alone accepts case-insensitive keys and null scalar values,
// which would discard unknown constraints or silently turn missing data to zero.
// Custom JSON values (Nullable and RawMessage) retain their own contracts.
type unknownFieldError struct{ field string }

func (e *unknownFieldError) Error() string { return fmt.Sprintf("json: unknown field %q", e.field) }

func jsonShape(raw []byte, typ reflect.Type) error {
	if typ.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil
		}
		return jsonShape(raw, typ.Elem())
	}
	unmarshaler := reflect.TypeFor[json.Unmarshaler]()
	if reflect.PointerTo(typ).Implements(unmarshaler) {
		return nil
	}
	isNull := bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
	switch typ.Kind() {
	case reflect.Struct:
		if isNull {
			return fmt.Errorf("object cannot be null")
		}
		fields, err := objectFields(raw)
		if err != nil {
			return err
		}
		known := jsonFields(typ)
		for key, value := range fields {
			field, ok := known[key]
			if !ok {
				return &unknownFieldError{field: key}
			}
			if err := jsonShape(value, field); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	case reflect.Slice, reflect.Array:
		if isNull {
			return nil
		}
		var values []json.RawMessage
		if err := json.Unmarshal(raw, &values); err != nil {
			return err
		}
		for i, value := range values {
			if err := jsonShape(value, typ.Elem()); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
	case reflect.Map:
		if isNull {
			return nil
		}
		values, err := objectFields(raw)
		if err != nil {
			return err
		}
		for key, value := range values {
			if err := jsonShape(value, typ.Elem()); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
	default:
		if isNull {
			return fmt.Errorf("scalar value cannot be null")
		}
	}
	return nil
}

func jsonFields(typ reflect.Type) map[string]reflect.Type {
	fields := map[string]reflect.Type{}
	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}
		if field.Anonymous && name == "" && field.Type.Kind() == reflect.Struct {
			for key, nested := range jsonFields(field.Type) {
				fields[key] = nested
			}
			continue
		}
		if name == "" {
			name = field.Name
		}
		fields[name] = field.Type
	}
	return fields
}

func objectFields(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, fmt.Errorf("definition must be an object")
	}
	fields := map[string]json.RawMessage{}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("object key must be a string")
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("duplicate JSON field %q", key)
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		fields[key] = value
	}
	_, err = decoder.Token()
	return fields, err
}

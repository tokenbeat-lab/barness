// Package jsonshape validates raw definitions before typed decoding can erase constraints.
package jsonshape

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
)

// Validate validates exact field spelling and non-null scalar/struct fields.
// encoding/json alone accepts case-insensitive keys and null scalar values,
// which would discard unknown constraints or silently turn missing data to zero.
// Custom JSON values (Nullable and RawMessage) retain their own contracts.
type UnknownFieldError struct{ field string }

func (e *UnknownFieldError) Error() string { return fmt.Sprintf("json: unknown field %q", e.field) }

func Validate(raw []byte, typ reflect.Type) error {
	if !json.Valid(raw) {
		return fmt.Errorf("expected exactly one valid JSON value")
	}
	if typ.Kind() == reflect.Pointer {
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return nil
		}
		return Validate(raw, typ.Elem())
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
		fields, err := ObjectFields(raw)
		if err != nil {
			return err
		}
		known := jsonFields(typ)
		for _, key := range slices.Sorted(maps.Keys(fields)) {
			value := fields[key]
			field, ok := known[key]
			if !ok {
				return &UnknownFieldError{field: key}
			}
			if err := Validate(value, field); err != nil {
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
			if err := Validate(value, typ.Elem()); err != nil {
				return fmt.Errorf("element %d: %w", i, err)
			}
		}
	case reflect.Map:
		if isNull {
			return nil
		}
		values, err := ObjectFields(raw)
		if err != nil {
			return err
		}
		for _, key := range slices.Sorted(maps.Keys(values)) {
			value := values[key]
			if err := Validate(value, typ.Elem()); err != nil {
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

func ObjectFields(raw []byte) (map[string]json.RawMessage, error) {
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

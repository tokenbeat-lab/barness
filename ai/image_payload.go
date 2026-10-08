package ai

import (
	"encoding/json"
	"reflect"
	"strings"
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

func imageStructField(v reflect.Value, name string) (reflect.Value, *Error) {
	// JSON selects promoted fields by their JSON name, independently of Go
	// name shadowing. Prefer the shallowest tagged resource field; equal-depth
	// duplicates cannot encode an unambiguous resource field.
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
			if field.PkgPath == "" && key == name {
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
	_, options, _ := strings.Cut(v.Type().FieldByIndex(index).Tag.Get("json"), ",")
	for _, option := range strings.Split(options, ",") {
		if option == "omitzero" {
			zeroer := reflect.TypeFor[interface{ IsZero() bool }]()
			if field.Type().Implements(zeroer) || reflect.PointerTo(field.Type()).Implements(zeroer) {
				// Host omission methods execute once in the encoder. Their
				// unknown field presence belongs to the final byte/schema guard.
				return reflect.Value{}, nil
			}
			if field.IsZero() {
				return reflect.Value{}, nil
			}
		}
		if option == "omitempty" {
			empty := false
			switch field.Kind() {
			case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
				empty = field.Len() == 0
			case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
				reflect.Float32, reflect.Float64, reflect.Interface, reflect.Pointer:
				empty = field.IsZero()
			}
			if empty {
				return reflect.Value{}, nil
			}
		}
	}
	return field, nil
}

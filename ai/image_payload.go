package ai

import (
	"encoding/json"
	"reflect"
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
	for _, field := range imageJSONFields(v.Type()) {
		if field.name == name {
			return field.value(v), nil
		}
	}
	return reflect.Value{}, nil
}

package ai

import (
	"bytes"
	"encoding/json"
)

// Nullable is an option value that keeps pi-ai's three states apart: unset
// (the zero value), explicit null, and a value — including the type's zero
// value. Fields whose wire encoding differs among the three use it; omitempty
// alone cannot tell them apart (spec I7).
//
// In JSON an unset Nullable is omitted (with the omitzero tag), null is null
// and a value is the value, so options decoded from pi-shaped JSON keep the
// distinction.
type Nullable[T any] struct {
	state nullableState
	value T
}

type nullableState uint8

const (
	nullableUnset nullableState = iota
	nullableNull
	nullableValue
)

// Value returns a Nullable holding v.
func Value[T any](v T) Nullable[T] { return Nullable[T]{state: nullableValue, value: v} }

// Null returns an explicit null.
func Null[T any]() Nullable[T] { return Nullable[T]{state: nullableNull} }

// Get returns the value and true, or the zero value and false when unset or
// null.
func (n Nullable[T]) Get() (T, bool) { return n.value, n.state == nullableValue }

// IsNull reports an explicit null.
func (n Nullable[T]) IsNull() bool { return n.state == nullableNull }

// IsZero reports an unset value; encoding/json's omitzero omits it.
func (n Nullable[T]) IsZero() bool { return n.state == nullableUnset }

// MarshalJSON encodes null for null (and unset) and the value otherwise.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if n.state != nullableValue {
		return []byte("null"), nil
	}
	return json.Marshal(n.value)
}

// UnmarshalJSON decodes null as Null and anything else as a value.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*n = Null[T]()
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*n = Value(v)
	return nil
}

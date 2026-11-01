package maintenance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tokenbeat-lab/barness/catalog/data"
)

// BuiltinState returns a fresh copy of the checked-in maintenance source.
func BuiltinState() (State, error) {
	state := State{RulesRevision: "barness-catalog-2026-11-01"}
	names, err := data.Providers.ReadDir("providers")
	if err != nil {
		return State{}, err
	}
	for _, name := range names {
		if name.IsDir() {
			continue
		}
		raw, err := data.Providers.ReadFile("providers/" + name.Name())
		if err != nil {
			return State{}, err
		}
		var provider ProviderRecord
		if err := decode(raw, &provider); err != nil {
			return State{}, fmt.Errorf("provider %s: %w", name.Name(), err)
		}
		state.Providers = append(state.Providers, provider)
	}
	return state, nil
}

// DecodeState strictly reads a saved maintenance source. Runtime model
// definitions are independently validated when Compose is called.
func DecodeState(raw []byte) (State, error) {
	var state State
	if err := decode(raw, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func decode(raw []byte, value any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON value")
	}
	return nil
}

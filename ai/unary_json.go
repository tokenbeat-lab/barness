package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Duplicate envelope fields are omitted, so ambiguous usage or response
// identity cannot become authoritative. Unique usage survives answer failures.
func unaryJSONEnvelope(raw []byte) (map[string]json.RawMessage, bool, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, false, fmt.Errorf("invalid unary envelope")
	}
	fields, seen := map[string]json.RawMessage{}, map[string]bool{}
	duplicate := false
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return nil, false, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, false, fmt.Errorf("invalid unary envelope key")
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, false, err
		}
		if seen[key] {
			delete(fields, key)
			duplicate = true
		} else {
			fields[key] = value
		}
		seen[key] = true
	}
	if _, err := d.Token(); err != nil {
		return nil, false, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, false, fmt.Errorf("unary envelope must contain one value")
	}
	return fields, duplicate, nil
}

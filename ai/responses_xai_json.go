package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Check protocol objects before decoding them into maps. Schema attributes,
// argument/output strings and unknown native extensions are business JSON;
// their key names and exact numbers do not acquire protocol meaning here.
func xaiDuplicateControls(raw []byte, shape string) (bool, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	opening, err := d.Token()
	if err != nil {
		return false, err
	}
	if opening != json.Delim('{') && opening != json.Delim('[') {
		return false, nil // The normal field validator owns type errors.
	}
	duplicate := false
	seen := map[string]bool{}
	for d.More() {
		key := ""
		if opening == json.Delim('{') {
			token, err := d.Token()
			if err != nil {
				return false, err
			}
			key = token.(string)
			// Detection precedes assignment/folding, including escaped keys.
			duplicate = duplicate || seen[key]
			seen[key] = true
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return false, err
		}
		child := xaiControlChild(shape, key)
		if child != "" {
			found, err := xaiDuplicateControls(value, child)
			if err != nil {
				return false, err
			}
			duplicate = duplicate || found
		}
	}
	if _, err := d.Token(); err != nil {
		return false, err
	}
	if _, err := d.Token(); err != io.EOF {
		return false, fmt.Errorf("trailing xAI JSON")
	}
	return duplicate, nil
}

func xaiControlChild(shape, key string) string {
	switch shape {
	case "request":
		switch key {
		case "reasoning", "text":
			return key
		case "input", "tools":
			return key + "-array"
		}
	case "input-array":
		return "input-item"
	case "tools-array":
		return "tool"
	case "input-item":
		if key == "content" || key == "summary" {
			return "parts-array"
		}
	case "parts-array":
		return "part"
	case "text":
		if key == "format" {
			return "format"
		}
	}
	return ""
}

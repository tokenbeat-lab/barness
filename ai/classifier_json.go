package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// decodeClassifierJSON rejects unknown fields and trailing JSON. Every wire
// question is independently converted to its public domain value.
func decodeClassifierJSON(raw []byte, out any) error {
	if err := uniqueClassifierJSON(raw); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("classifier JSON must contain one value")
	}
	return nil
}

func (r *ClassifierRequest) UnmarshalJSON(raw []byte) error {
	var wire struct {
		State     json.RawMessage            `json:"state"`
		Questions map[string]json.RawMessage `json:"questions"`
	}
	if err := decodeClassifierJSON(raw, &wire); err != nil {
		return err
	}
	out := ClassifierRequest{State: wire.State, Questions: make(map[string]ClassifierQuestion, len(wire.Questions))}
	for key, raw := range wire.Questions {
		var q struct {
			Type         string                     `json:"type"`
			Instructions json.RawMessage            `json:"instructions"`
			Criteria     map[string]json.RawMessage `json:"criteria"`
		}
		if err := decodeClassifierJSON(raw, &q); err != nil {
			return err
		}
		if q.Type != "choice" {
			return fmt.Errorf("unsupported classifier question type")
		}
		out.Questions[key] = ChoiceQuestion{Instructions: q.Instructions, Criteria: q.Criteria}
	}
	*r = out
	return nil
}

// uniqueClassifierJSON prevents repeated question/answer/option keys from
// being silently collapsed by encoding/json's last-value-wins behavior.
func uniqueClassifierJSON(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	return uniqueClassifierValue(d)
}
func uniqueClassifierValue(d *json.Decoder) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delim {
	case '{':
		seen := map[string]bool{}
		for d.More() {
			token, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok || seen[key] {
				return fmt.Errorf("duplicate classifier JSON field")
			}
			seen[key] = true
			if err := uniqueClassifierValue(d); err != nil {
				return err
			}
		}
	case '[':
		for d.More() {
			if err := uniqueClassifierValue(d); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("invalid classifier JSON delimiter")
	}
	_, err = d.Token()
	return err
}

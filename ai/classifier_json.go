package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
)

// decodeClassifierJSON rejects unknown fields and trailing JSON. Every wire
// question is independently converted to its public domain value.
func decodeClassifierJSON(raw []byte, out any) error {
	if err := uniqueClassifierJSON(raw); err != nil {
		return err
	}
	if err := exactClassifierFields(raw, out); err != nil {
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

// encoding/json treats struct keys case-insensitively even with
// DisallowUnknownFields. Closed wire records require their exact JSON tags,
// so an alias cannot override an invalid value or masquerade as a known field.
// Arbitrary state/description objects and question names stay raw/map values.
func exactClassifierFields(raw []byte, out any) error {
	t := reflect.TypeOf(out)
	if t == nil || t.Kind() != reflect.Pointer {
		return nil
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return nil
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	allowed := make(map[string]bool, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name != "" && name != "-" {
			allowed[name] = true
		}
	}
	for key := range fields {
		if !allowed[key] {
			return fmt.Errorf("unknown classifier record field")
		}
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
	questions, err := decodeClassifierQuestions(wire.Questions, false)
	if err != nil {
		return err
	}
	out := ClassifierRequest{State: wire.State, Questions: questions}
	*r = out
	return nil
}

func decodeClassifierQuestions(raw map[string]json.RawMessage, native bool) (map[string]ClassifierQuestion, error) {
	out := make(map[string]ClassifierQuestion, len(raw))
	for key, raw := range raw {
		var tag struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &tag); err != nil {
			return nil, err
		}
		var question ClassifierQuestion
		switch tag.Type {
		case "choice":
			var wire struct {
				Type         string                     `json:"type"`
				Instructions json.RawMessage            `json:"instructions"`
				Criteria     map[string]json.RawMessage `json:"criteria"`
			}
			if err := decodeClassifierJSON(raw, &wire); err != nil {
				return nil, err
			}
			question = ChoiceQuestion{Instructions: wire.Instructions, Criteria: wire.Criteria}
		case "score":
			var wire struct {
				Type         string            `json:"type"`
				Instructions json.RawMessage   `json:"instructions"`
				Criteria     []json.RawMessage `json:"criteria"`
			}
			if err := decodeClassifierJSON(raw, &wire); err != nil {
				return nil, err
			}
			question = ScoreQuestion{Instructions: wire.Instructions, Criteria: wire.Criteria}
		case "bool", "noul":
			if (tag.Type == "noul") != native {
				return nil, fmt.Errorf("invalid bool question encoding")
			}
			var wire struct {
				Type         string          `json:"type"`
				Instructions json.RawMessage `json:"instructions"`
				Criteria     json.RawMessage `json:"criteria,omitempty"`
			}
			if err := decodeClassifierJSON(raw, &wire); err != nil {
				return nil, err
			}
			var criteria *BoolCriteria
			if len(wire.Criteria) > 0 {
				if bytes.Equal(bytes.TrimSpace(wire.Criteria), []byte("null")) {
					return nil, fmt.Errorf("bool criteria must be an object when supplied")
				}
				criteria = new(BoolCriteria)
				if err := decodeClassifierJSON(wire.Criteria, criteria); err != nil {
					return nil, err
				}
			}
			question = BoolQuestion{Instructions: wire.Instructions, Criteria: criteria}
		default:
			return nil, fmt.Errorf("unsupported classifier question type")
		}
		out[key] = question
	}
	return out, nil
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

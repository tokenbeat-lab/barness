package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// These records belong to the TypeSafe boundary. Decoding the final bytes
// gives them independent ownership, including raw JSON numbers and criteria;
// they never become Host requests or borrow the callback's mutable values.
type typeSafeRequest struct {
	Model     string                      `json:"model"`
	State     json.RawMessage             `json:"state"`
	Questions map[string]typeSafeQuestion `json:"questions"`
	Assets    map[string]typeSafeAsset    `json:"assets,omitempty"`
}

// The native discriminator selects exactly one criteria shape. Keep the
// decoded variants private so answer validation needs no public request.
type typeSafeQuestion struct {
	kind         string
	instructions json.RawMessage
	choices      map[string]json.RawMessage
	levels       []json.RawMessage
	boolean      *typeSafeBoolCriteria
}

type typeSafeBoolCriteria struct {
	True  json.RawMessage `json:"true"`
	False json.RawMessage `json:"false"`
}

func encodeTypeSafeRequest(model string, req ClassifierRequest) ([]byte, error) {
	wire := typeSafeRequest{Model: model, State: req.State, Questions: make(map[string]typeSafeQuestion, len(req.Questions))}
	for name, question := range req.Questions {
		var native typeSafeQuestion
		switch q := question.(type) {
		case ChoiceQuestion:
			native = typeSafeQuestion{kind: "choice", instructions: q.Instructions, choices: q.Criteria}
		case ScoreQuestion:
			native = typeSafeQuestion{kind: "score", instructions: q.Instructions, levels: q.Criteria}
		case BoolQuestion:
			native = typeSafeQuestion{kind: "noul", instructions: q.Instructions}
			if q.Criteria != nil {
				native.boolean = &typeSafeBoolCriteria{True: q.Criteria.True, False: q.Criteria.False}
			}
		default:
			return nil, fmt.Errorf("unsupported TypeSafe question")
		}
		wire.Questions[name] = native
	}
	return json.Marshal(wire)
}

func (q typeSafeQuestion) MarshalJSON() ([]byte, error) {
	var criteria any
	switch q.kind {
	case "choice":
		criteria = q.choices
	case "score":
		criteria = q.levels
	case "noul":
		if q.boolean != nil {
			criteria = q.boolean
		}
	}
	return json.Marshal(struct {
		Type         string          `json:"type"`
		Instructions json.RawMessage `json:"instructions"`
		Criteria     any             `json:"criteria,omitempty"`
	}{q.kind, q.instructions, criteria})
}

func (q *typeSafeQuestion) UnmarshalJSON(raw []byte) error {
	var wire struct {
		Type         string          `json:"type"`
		Instructions json.RawMessage `json:"instructions"`
		Criteria     json.RawMessage `json:"criteria"`
	}
	if err := decodeClassifierJSON(raw, &wire); err != nil {
		return err
	}
	value := typeSafeQuestion{kind: wire.Type, instructions: wire.Instructions}
	var err error
	switch wire.Type {
	case "choice":
		err = decodeClassifierJSON(wire.Criteria, &value.choices)
	case "score":
		err = decodeClassifierJSON(wire.Criteria, &value.levels)
	case "noul":
		if len(wire.Criteria) > 0 {
			if bytes.Equal(bytes.TrimSpace(wire.Criteria), []byte("null")) {
				return fmt.Errorf("TypeSafe noul criteria must be an object when supplied")
			}
			value.boolean = new(typeSafeBoolCriteria)
			err = decodeClassifierJSON(wire.Criteria, value.boolean)
		}
	default:
		return fmt.Errorf("unsupported TypeSafe question type")
	}
	if err != nil {
		return err
	}
	*q = value
	return nil
}

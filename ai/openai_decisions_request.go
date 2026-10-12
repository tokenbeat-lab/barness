package ai

import (
	"bytes"
	"encoding/json"
	"maps"
	"slices"
	"strconv"
)

// These private records own native protocol data. They never borrow the
// caller's maps or become the public domain representation.
type decisionsRequest struct {
	Model     string              `json:"model"`
	Input     *string             `json:"input"`
	Questions []decisionsQuestion `json:"questions"`
}
type decisionsQuestion struct {
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Instructions *string           `json:"instructions"`
	Choices      []decisionsChoice `json:"choices"`
}
type decisionsChoice struct {
	Value       *string `json:"value"`
	Description *string `json:"description,omitempty"`
}

func encodeDecisionsRequest(model string, req ClassifierRequest) ([]byte, error) {
	var text string
	state := bytes.TrimSpace(req.State)
	if state[0] == '"' {
		if err := json.Unmarshal(state, &text); err != nil {
			return nil, err
		}
	} else {
		var compact bytes.Buffer
		if err := json.Compact(&compact, state); err != nil {
			return nil, err
		}
		text = compact.String()
	}
	wire := decisionsRequest{Model: model, Input: &text}
	for _, name := range decisionsKeys(req.Questions) {
		q := req.Questions[name].(ChoiceQuestion) // capability checked before pin
		var instructions string
		if err := json.Unmarshal(q.Instructions, &instructions); err != nil {
			return nil, err
		}
		question := decisionsQuestion{Type: "choice", Name: name, Instructions: &instructions}
		for _, value := range decisionsKeys(q.Criteria) {
			choice := decisionsChoice{Value: &value}
			if !bytes.Equal(bytes.TrimSpace(q.Criteria[value]), []byte("null")) {
				var description string
				if err := json.Unmarshal(q.Criteria[value], &description); err != nil {
					return nil, err
				}
				if description != "" {
					choice.Description = &description
				}
			}
			question.Choices = append(question.Choices, choice)
		}
		wire.Questions = append(wire.Questions, question)
	}
	return marshalJS(wire)
}

// Go public maps promise no insertion order. Their frozen JSON object keys
// are lexical; JS canonical array-index names enumerate before those keys.
func decisionsKeys[V any](values map[string]V) []string {
	keys := slices.Sorted(maps.Keys(values))
	slices.SortStableFunc(keys, func(a, b string) int {
		aok, bok := isArrayIndex(a), isArrayIndex(b)
		if aok && bok {
			ai, _ := strconv.ParseUint(a, 10, 32)
			bi, _ := strconv.ParseUint(b, 10, 32)
			if ai < bi {
				return -1
			}
			if ai > bi {
				return 1
			}
			return 0
		}
		if aok {
			return -1
		}
		if bok {
			return 1
		}
		return 0
	})
	return keys
}

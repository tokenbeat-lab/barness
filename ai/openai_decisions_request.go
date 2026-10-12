package ai

import (
	"bytes"
	"encoding/json"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// These private records own native protocol data. They never borrow the
// caller's maps or become the public domain representation.
type decisionsRequest struct {
	Model     string              `json:"model"`
	Input     json.RawMessage     `json:"input"`
	Questions []decisionsQuestion `json:"questions"`
}
type decisionsQuestion struct {
	Type         string            `json:"type"`
	Name         string            `json:"name"`
	Instructions *string           `json:"instructions"`
	Choices      []decisionsChoice `json:"choices,omitempty"`
	Levels       []decisionsLevel  `json:"levels,omitempty"`
}
type decisionsLevel struct {
	Label       *string `json:"label"`
	Description *string `json:"description,omitempty"`
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
	input, err := marshalJS(text)
	if err != nil {
		return nil, err
	}
	if len(req.Images) > 0 {
		content := []decisionsContent{{Type: "input_text", Text: &text}}
		for _, img := range req.Images {
			url := "data:" + img.MimeType + ";base64," + img.Data
			content = append(content, decisionsContent{Type: "input_image", ImageURL: &url})
		}
		input, err = marshalJS([]decisionsMessage{{Role: "user", Content: content}})
		if err != nil {
			return nil, err
		}
	}
	wire := decisionsRequest{Model: model, Input: input}
	for _, name := range decisionsKeys(req.Questions) {
		question, err := encodeDecisionsQuestion(name, req.Questions[name])
		if err != nil {
			return nil, err
		}
		wire.Questions = append(wire.Questions, question)
	}
	return marshalJS(wire)
}

// Capability validation guarantees string descriptions before conversion;
// native score and predicate retain their own semantics, never boolean choices.
func encodeDecisionsQuestion(name string, public ClassifierQuestion) (decisionsQuestion, error) {
	question := decisionsQuestion{Name: name}
	var raw json.RawMessage
	switch q := public.(type) {
	case ChoiceQuestion:
		question.Type, raw = "choice", q.Instructions
		for _, value := range decisionsKeys(q.Criteria) {
			choice := decisionsChoice{Value: &value}
			if !bytes.Equal(bytes.TrimSpace(q.Criteria[value]), []byte("null")) {
				var description string
				if err := json.Unmarshal(q.Criteria[value], &description); err != nil {
					return question, err
				}
				if description != "" {
					choice.Description = &description
				}
			}
			question.Choices = append(question.Choices, choice)
		}
	case ScoreQuestion:
		question.Type, raw = "score", q.Instructions
		for _, criterion := range q.Criteria {
			var label string
			if err := json.Unmarshal(criterion, &label); err != nil {
				return question, err
			}
			question.Levels = append(question.Levels, decisionsLevel{Label: &label})
		}
	case BoolQuestion:
		question.Type, raw = "predicate", q.Instructions
	}
	var instructions string
	if err := json.Unmarshal(raw, &instructions); err != nil {
		return question, err
	}
	if q, ok := public.(BoolQuestion); ok && q.Criteria != nil {
		var yes, no string
		if err := json.Unmarshal(q.Criteria.True, &yes); err != nil {
			return question, err
		}
		if err := json.Unmarshal(q.Criteria.False, &no); err != nil {
			return question, err
		}
		lines := []string{}
		if yes != "" {
			lines = append(lines, "True means: "+yes)
		}
		if no != "" {
			lines = append(lines, "False means: "+no)
		}
		if len(lines) > 0 {
			instructions += "\n\n" + strings.Join(lines, "\n")
		}
	}
	question.Instructions = &instructions
	return question, nil
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

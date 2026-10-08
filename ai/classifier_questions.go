package ai

import "encoding/json"

// ClassifierQuestion is a closed choice, score or bool question.
type ClassifierQuestion interface{ classifierQuestion() }

// ChoiceQuestion asks for one option. Instructions and criteria descriptions
// accept JSON strings, objects or arrays; a null criterion uses its name alone.
type ChoiceQuestion struct {
	Instructions json.RawMessage            `json:"instructions"`
	Criteria     map[string]json.RawMessage `json:"criteria"`
}

// ScoreQuestion rates an ordered rubric of 2–10 levels. Each criterion's
// position is its score, starting at zero; descriptions accept native JSON.
type ScoreQuestion struct {
	Instructions json.RawMessage   `json:"instructions"`
	Criteria     []json.RawMessage `json:"criteria"`
}

// BoolQuestion asks for the probability of yes. Optional criteria clarify
// both alternatives; when supplied, both must be JSON strings, objects or arrays.
type BoolQuestion struct {
	Instructions json.RawMessage `json:"instructions"`
	Criteria     *BoolCriteria   `json:"criteria,omitempty"`
}

type BoolCriteria struct {
	True  json.RawMessage `json:"true"`
	False json.RawMessage `json:"false"`
}

func (ChoiceQuestion) classifierQuestion() {}
func (ScoreQuestion) classifierQuestion()  {}
func (BoolQuestion) classifierQuestion()   {}

// Concrete tagged encoders intentionally keep each closed type's fields explicit;
// a shared untyped field-merging encoder would obscure the public JSON contract.
func (q ChoiceQuestion) MarshalJSON() ([]byte, error) {
	type question ChoiceQuestion
	return json.Marshal(struct {
		Type string `json:"type"`
		question
	}{"choice", question(q)})
}
func (q ScoreQuestion) MarshalJSON() ([]byte, error) {
	type question ScoreQuestion
	return json.Marshal(struct {
		Type string `json:"type"`
		question
	}{"score", question(q)})
}
func (q BoolQuestion) MarshalJSON() ([]byte, error) {
	type question BoolQuestion
	return json.Marshal(struct {
		Type string `json:"type"`
		question
	}{"bool", question(q)})
}

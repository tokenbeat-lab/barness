package ai

import "encoding/json"

// ClassifierAnswer is a validated answer to a requested question.
type ClassifierAnswer interface{ classifierAnswer() }

type ChoiceAnswer struct {
	Choice        string             `json:"choice"`
	Probabilities map[string]float64 `json:"probabilities"`
	Confidence    float64            `json:"confidence"`
}

// ScoreAnswer retains the expected score and any reported level distribution
// and rubric. Legend values preserve the final criteria's native JSON.
type ScoreAnswer struct {
	Score         float64                 `json:"score"`
	Confidence    float64                 `json:"confidence"`
	Probabilities map[int]float64         `json:"probabilities,omitempty"`
	Legend        map[int]json.RawMessage `json:"legend,omitempty"`
}

// BoolAnswer is the probability of yes, rather than a thresholded decision.
type BoolAnswer struct {
	Probability float64 `json:"probability"`
}

func (ChoiceAnswer) classifierAnswer() {}
func (ScoreAnswer) classifierAnswer()  {}
func (BoolAnswer) classifierAnswer()   {}

// Concrete tagged encoders intentionally keep each closed type's fields explicit;
// a shared untyped field-merging encoder would obscure the public JSON contract.
func (a ChoiceAnswer) MarshalJSON() ([]byte, error) {
	type answer ChoiceAnswer
	return json.Marshal(struct {
		Type string `json:"type"`
		answer
	}{"choice", answer(a)})
}
func (a ScoreAnswer) MarshalJSON() ([]byte, error) {
	type answer ScoreAnswer
	return json.Marshal(struct {
		Type string `json:"type"`
		answer
	}{"score", answer(a)})
}
func (a BoolAnswer) MarshalJSON() ([]byte, error) {
	type answer BoolAnswer
	return json.Marshal(struct {
		Type string `json:"type"`
		answer
	}{"bool", answer(a)})
}

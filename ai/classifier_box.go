package ai

import (
	"encoding/json"
	"fmt"
)

// BoxQuestion locates one target. Each Box must see an image in shared context,
// State or its own Instructions. Instructions follow the other question types'
// nonempty JSON and explicit image-reference contract; Box has no criteria.
type BoxQuestion struct {
	Instructions json.RawMessage `json:"instructions"`
}

// Box describes normalized image edges measured from the top-left. The Host
// owns search-image attribution and any display, crop or pixel transformations.
type Box struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// BoxAnswer retains one positive-area rectangle even when a separate BoolAnswer
// reports low or zero existence. Confidence is required but may be explicit null;
// it is not an existence probability. No image identity is inferred here.
type BoxAnswer struct {
	Box        Box               `json:"box"`
	Confidence Nullable[float64] `json:"confidence"`
}

func (BoxQuestion) classifierQuestion() {}
func (BoxAnswer) classifierAnswer()     {}

func (q BoxQuestion) MarshalJSON() ([]byte, error) {
	type question BoxQuestion
	return json.Marshal(struct {
		Type string `json:"type"`
		question
	}{"box", question(q)})
}

func (a BoxAnswer) MarshalJSON() ([]byte, error) {
	confidence, numeric := a.Confidence.Get()
	if !a.Box.valid() || a.Confidence.IsZero() || (numeric && !unitProbability(confidence)) {
		return nil, fmt.Errorf("invalid Box answer")
	}
	type answer BoxAnswer
	return json.Marshal(struct {
		Type string `json:"type"`
		answer
	}{"box", answer(a)})
}

func (b Box) valid() bool {
	return unitProbability(b.Top) && unitProbability(b.Bottom) && unitProbability(b.Left) && unitProbability(b.Right) && b.Left < b.Right && b.Top < b.Bottom
}

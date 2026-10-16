package ai

import "encoding/json"

func decodeBoxAnswer(raw json.RawMessage) (BoxAnswer, *Error) {
	var wire struct {
		Type       string            `json:"type"`
		Box        json.RawMessage   `json:"box"`
		Confidence Nullable[float64] `json:"confidence"`
	}
	if decodeClassifierJSON(raw, &wire) != nil || wire.Type != "box" || wire.Confidence.IsZero() {
		return BoxAnswer{}, classifierProtocolFailure()
	}
	// Native edges use pointers to distinguish an absent/null edge from numeric
	// zero. Only fully validated native values become a public value rectangle.
	var edges struct {
		Top    *float64 `json:"top"`
		Bottom *float64 `json:"bottom"`
		Left   *float64 `json:"left"`
		Right  *float64 `json:"right"`
	}
	if decodeClassifierJSON(wire.Box, &edges) != nil || edges.Top == nil || edges.Bottom == nil || edges.Left == nil || edges.Right == nil {
		return BoxAnswer{}, classifierProtocolFailure()
	}
	box := Box{Top: *edges.Top, Bottom: *edges.Bottom, Left: *edges.Left, Right: *edges.Right}
	confidence, numeric := wire.Confidence.Get()
	if !box.valid() || (numeric && !unitProbability(confidence)) {
		return BoxAnswer{}, classifierProtocolFailure()
	}
	return BoxAnswer{Box: box, Confidence: wire.Confidence}, nil
}

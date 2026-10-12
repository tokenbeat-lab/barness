package ai

import (
	"encoding/json"
	"math"
	"slices"
	"strings"
)

// Decisions responses may add vendor metadata. Known answer semantics stay
// exact and kind-specific; contradictory fields and case aliases never become
// metadata. Total response bytes are bounded by the unary transport, and the
// envelope has already rejected duplicate keys throughout the JSON tree.
// TypeSafe deliberately continues using its independent closed decoder.
func decisionsAnswerFields(raw []byte, allowed ...string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return false
	}
	for key := range fields {
		if slices.Contains(allowed, key) {
			continue
		}
		for _, semantic := range []string{"type", "name", "choice", "score", "confidence", "probabilities", "probability", "value", "label", "legend", "noul", "reason", "message", "answer", "refusal", "decision"} {
			if strings.EqualFold(key, semantic) {
				return false
			}
		}
	}
	return true
}

func decodeDecisionsChoice(raw []byte, q decisionsQuestion) (ClassifierAnswer, *Error) {
	var wire struct {
		Choice        *string           `json:"choice"`
		Probabilities []json.RawMessage `json:"probabilities"`
		Confidence    *float64          `json:"confidence"`
	}
	if !decisionsAnswerFields(raw, "type", "name", "choice", "probabilities", "confidence") || json.Unmarshal(raw, &wire) != nil || wire.Choice == nil || wire.Confidence == nil || !unitProbability(*wire.Confidence) || len(wire.Probabilities) != len(q.Choices) {
		return nil, classifierProtocolFailure()
	}
	values := make(map[string]bool, len(q.Choices))
	for _, c := range q.Choices {
		values[*c.Value] = true
	}
	if !values[*wire.Choice] {
		return nil, classifierProtocolFailure()
	}
	answer := ChoiceAnswer{Choice: *wire.Choice, Confidence: *wire.Confidence, Probabilities: make(map[string]float64, len(q.Choices))}
	sum, highest := 0.0, 0.0
	for _, raw := range wire.Probabilities {
		var entry struct {
			Value       *string  `json:"value"`
			Probability *float64 `json:"probability"`
		}
		if !decisionsAnswerFields(raw, "value", "probability") || json.Unmarshal(raw, &entry) != nil || entry.Value == nil || entry.Probability == nil || !unitProbability(*entry.Probability) || !values[*entry.Value] {
			return nil, classifierProtocolFailure()
		}
		if _, ok := answer.Probabilities[*entry.Value]; ok {
			return nil, classifierProtocolFailure()
		}
		answer.Probabilities[*entry.Value] = *entry.Probability
		sum += *entry.Probability
		highest = max(highest, *entry.Probability)
	}
	if math.Abs(sum-1) > 1e-6 || answer.Probabilities[answer.Choice] < highest {
		return nil, classifierProtocolFailure()
	}
	return answer, nil
}

func decodeDecisionsPredicate(raw []byte) (ClassifierAnswer, *Error) {
	var wire struct {
		Probability *float64 `json:"probability"`
	}
	if !decisionsAnswerFields(raw, "type", "name", "probability") || json.Unmarshal(raw, &wire) != nil || wire.Probability == nil || !unitProbability(*wire.Probability) {
		return nil, classifierProtocolFailure()
	}
	return BoolAnswer{Probability: *wire.Probability}, nil
}

func decodeDecisionsScore(raw []byte, q decisionsQuestion) (ClassifierAnswer, *Error) {
	var wire struct {
		Score         *float64          `json:"score"`
		Confidence    *float64          `json:"confidence"`
		Probabilities []json.RawMessage `json:"probabilities"`
	}
	if !decisionsAnswerFields(raw, "type", "name", "score", "confidence", "probabilities") || json.Unmarshal(raw, &wire) != nil || wire.Score == nil || wire.Confidence == nil || !unitProbability(*wire.Confidence) || math.IsNaN(*wire.Score) || math.IsInf(*wire.Score, 0) || *wire.Score < 0 || *wire.Score > float64(len(q.Levels)-1) || len(wire.Probabilities) != len(q.Levels) {
		return nil, classifierProtocolFailure()
	}
	answer := ScoreAnswer{Score: *wire.Score, Confidence: *wire.Confidence, Probabilities: make(map[int]float64, len(q.Levels)), Legend: make(map[int]json.RawMessage, len(q.Levels))}
	for _, raw := range wire.Probabilities {
		var entry struct {
			Value       *int     `json:"value"`
			Label       *string  `json:"label"`
			Probability *float64 `json:"probability"`
		}
		if !decisionsAnswerFields(raw, "value", "label", "probability") || json.Unmarshal(raw, &entry) != nil || entry.Value == nil || *entry.Value < 0 || *entry.Value >= len(q.Levels) || entry.Label == nil || *entry.Label != *q.Levels[*entry.Value].Label || entry.Probability == nil || !unitProbability(*entry.Probability) {
			return nil, classifierProtocolFailure()
		}
		if _, exists := answer.Probabilities[*entry.Value]; exists {
			return nil, classifierProtocolFailure()
		}
		answer.Probabilities[*entry.Value] = *entry.Probability
		// Encode the verified native label as independent public rubric JSON.
		label, err := json.Marshal(*entry.Label)
		if err != nil {
			return nil, classifierProtocolFailure()
		}
		answer.Legend[*entry.Value] = label
	}
	sum, expected := 0.0, 0.0
	// Use frozen index order for deterministic numeric accumulation even when
	// the provider sends its distribution in a different order.
	for i := range q.Levels {
		p := answer.Probabilities[i]
		sum += p
		expected += float64(i) * p
	}
	if math.Abs(sum-1) > 1e-6 || math.Abs(expected-answer.Score) > 1e-6 {
		return nil, classifierProtocolFailure()
	}
	return answer, nil
}

package ai

import (
	"encoding/json"
	"math"
	"strings"
)

func decodeDecisionsResponse(raw []byte, model ClassifierModel, initial *initialRequest, req decisionsRequest, out *ClassifierResult) *Error {
	fields, duplicate, err := unaryJSONEnvelope(raw)
	if err != nil {
		return classifierProtocolFailure()
	}
	reporting, usage, failure := decisionsUsage(fields["usage"], model.Cost)
	out.Usage = usage
	initial.usage(reporting, usage)
	if failure != nil {
		return failure
	}
	if rawModel, ok := fields["model"]; ok {
		var value *string
		if json.Unmarshal(rawModel, &value) != nil || value == nil {
			return classifierProtocolFailure()
		}
		out.ResponseModel = Value(*value)
	}
	if duplicate || uniqueClassifierJSON(raw) != nil {
		return classifierProtocolFailure()
	}
	for key := range fields {
		for _, known := range []string{"answers", "model", "usage"} {
			if key != known && strings.EqualFold(key, known) {
				return classifierProtocolFailure()
			}
		}
	}
	var answers []json.RawMessage
	if json.Unmarshal(fields["answers"], &answers) != nil || len(answers) != len(req.Questions) {
		return classifierProtocolFailure()
	}
	questions := make(map[string]decisionsQuestion, len(req.Questions))
	for _, q := range req.Questions {
		questions[q.Name] = q
	}
	validated := make(map[string]ClassifierAnswer, len(answers))
	for _, raw := range answers {
		var tag struct {
			Type string  `json:"type"`
			Name *string `json:"name"`
		}
		if json.Unmarshal(raw, &tag) != nil || tag.Name == nil {
			return classifierProtocolFailure()
		}
		q, ok := questions[*tag.Name]
		if !ok || validated[*tag.Name] != nil {
			return classifierProtocolFailure()
		}
		if tag.Type == "refusal" {
			if decodeClassifierJSON(raw, &tag) != nil {
				return classifierProtocolFailure()
			}
			validated[*tag.Name] = RefusalAnswer{}
			continue
		}
		answer, failure := decodeDecisionsChoice(raw, q)
		if failure != nil {
			return failure
		}
		validated[*tag.Name] = answer
	}
	out.Answers = validated
	return nil
}

func decodeDecisionsChoice(raw []byte, q decisionsQuestion) (ClassifierAnswer, *Error) {
	var wire struct {
		Type          string            `json:"type"`
		Name          *string           `json:"name"`
		Choice        *string           `json:"choice"`
		Probabilities []json.RawMessage `json:"probabilities"`
		Confidence    *float64          `json:"confidence"`
	}
	if decodeClassifierJSON(raw, &wire) != nil || wire.Type != "choice" || wire.Choice == nil || wire.Confidence == nil || !unitProbability(*wire.Confidence) || len(wire.Probabilities) != len(q.Choices) {
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
		if decodeClassifierJSON(raw, &entry) != nil || entry.Value == nil || entry.Probability == nil || !unitProbability(*entry.Probability) || !values[*entry.Value] {
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

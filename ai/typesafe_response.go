package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
)

// Usage is registered before parsing any answer. A syntactically valid
// envelope with invalid answers retains consumption and response identity.
func decodeTypeSafeResponse(raw []byte, model ClassifierModel, initial *initialRequest, req ClassifierRequest, out *ClassifierResult) *Error {
	fields, duplicate, err := typeSafeEnvelope(raw)
	if err != nil {
		return classifierProtocolFailure()
	}
	reporting, usage, failure := typeSafeUsage(fields["usage"], model.Cost)
	out.Usage = usage
	initial.usage(reporting, usage)
	if failure != nil {
		return failure
	}
	if rawModel, ok := fields["model"]; ok {
		var responseModel string
		if string(rawModel) == "null" || json.Unmarshal(rawModel, &responseModel) != nil {
			return classifierProtocolFailure()
		}
		out.ResponseModel = Value(responseModel)
	}
	if duplicate {
		return classifierProtocolFailure()
	}
	var answers map[string]json.RawMessage
	if decodeClassifierJSON(fields["answers"], &answers) != nil || len(answers) != len(req.Questions) {
		return classifierProtocolFailure()
	}
	validated := make(map[string]ClassifierAnswer, len(answers))
	for key, question := range req.Questions {
		raw, ok := answers[key]
		if !ok {
			return classifierProtocolFailure()
		}
		var answer ClassifierAnswer
		var failure *Error
		switch q := question.(type) {
		case ChoiceQuestion:
			answer, failure = decodeChoiceAnswer(raw, q)
		case ScoreQuestion:
			answer, failure = decodeScoreAnswer(raw, q)
		case BoolQuestion:
			answer, failure = decodeBoolAnswer(raw)
		default:
			return classifierProtocolFailure()
		}
		if failure != nil {
			return failure
		}
		validated[key] = answer
	}
	out.Answers = validated
	return nil
}

// Duplicate envelope fields are omitted, so ambiguous usage or response
// identity cannot become authoritative. Unique usage survives answer failures.
func typeSafeEnvelope(raw []byte) (map[string]json.RawMessage, bool, error) {
	d := json.NewDecoder(bytes.NewReader(raw))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, false, fmt.Errorf("invalid classifier envelope")
	}
	fields, seen := map[string]json.RawMessage{}, map[string]bool{}
	duplicate := false
	for d.More() {
		token, err := d.Token()
		if err != nil {
			return nil, false, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, false, fmt.Errorf("invalid classifier envelope key")
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, false, err
		}
		if seen[key] {
			delete(fields, key)
			duplicate = true
		} else {
			fields[key] = value
		}
		seen[key] = true
	}
	if _, err := d.Token(); err != nil {
		return nil, false, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, false, fmt.Errorf("classifier envelope must contain one value")
	}
	return fields, duplicate, nil
}

func typeSafeUsage(raw json.RawMessage, cost ModelCost) (UsageReporting, Usage, *Error) {
	if len(raw) == 0 || string(raw) == "null" {
		return UsageUnreported, Usage{}, nil
	}
	var fields map[string]json.RawMessage
	if decodeClassifierJSON(raw, &fields) != nil || fields == nil {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	u := Usage{}
	input, inOK := typeSafeTokenCount(fields["input_tokens"])
	output, outOK := typeSafeTokenCount(fields["output_tokens"])
	if !inOK || !outOK || input > math.MaxInt64-output {
		return UsagePartial, u, classifierProtocolFailure()
	}
	u.Input, u.Output, u.TotalTokens = input, output, input+output
	// TypeSafe only charges input. Host rates remain authoritative, with
	// output/cache rates excluded even if a host supplied nonzero ones.
	cost.Output, cost.CacheRead, cost.CacheWrite = 0, 0, 0
	for i := range cost.Tiers {
		cost.Tiers[i].Output = 0
		cost.Tiers[i].CacheRead = 0
		cost.Tiers[i].CacheWrite = 0
	}
	u.Cost = cost.estimate(u)
	reporting := UsagePartial
	if len(fields["input_tokens"]) > 0 && string(fields["input_tokens"]) != "null" && len(fields["output_tokens"]) > 0 && string(fields["output_tokens"]) != "null" {
		reporting = UsageComplete
	}
	return reporting, u, nil
}

func typeSafeTokenCount(raw json.RawMessage) (int64, bool) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, true
	}
	var n int64
	if json.Unmarshal(raw, &n) != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func decodeChoiceAnswer(raw json.RawMessage, q ChoiceQuestion) (ChoiceAnswer, *Error) {
	var wire struct {
		Type          string                     `json:"type"`
		Choice        *string                    `json:"choice"`
		Probabilities map[string]json.RawMessage `json:"probabilities"`
		Confidence    *float64                   `json:"confidence"`
	}
	if decodeClassifierJSON(raw, &wire) != nil || wire.Type != "choice" || wire.Choice == nil || wire.Confidence == nil || !unitProbability(*wire.Confidence) {
		return ChoiceAnswer{}, classifierProtocolFailure()
	}
	if _, ok := q.Criteria[*wire.Choice]; !ok {
		return ChoiceAnswer{}, classifierProtocolFailure()
	}
	if len(wire.Probabilities) != len(q.Criteria) {
		return ChoiceAnswer{}, classifierProtocolFailure()
	}
	answer := ChoiceAnswer{Choice: *wire.Choice, Confidence: *wire.Confidence, Probabilities: make(map[string]float64, len(q.Criteria))}
	sum, highest := 0.0, 0.0
	for key := range q.Criteria {
		raw, ok := wire.Probabilities[key]
		var probability *float64
		if !ok || json.Unmarshal(raw, &probability) != nil || probability == nil || !unitProbability(*probability) {
			return ChoiceAnswer{}, classifierProtocolFailure()
		}
		answer.Probabilities[key] = *probability
		sum += *probability
		highest = max(highest, *probability)
	}
	if math.Abs(sum-1) > 1e-6 || answer.Probabilities[answer.Choice] < highest {
		return ChoiceAnswer{}, classifierProtocolFailure()
	}
	return answer, nil
}
func unitProbability(p float64) bool { return !math.IsNaN(p) && !math.IsInf(p, 0) && p >= 0 && p <= 1 }
func classifierProtocolFailure() *Error {
	return newError(CodeProtocol, PhaseResponse, "invalid classifier response")
}

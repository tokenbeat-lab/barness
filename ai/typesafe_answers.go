package ai

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
)

func decodeBoolAnswer(raw json.RawMessage) (BoolAnswer, *Error) {
	var wire struct {
		Type string   `json:"type"`
		Noul *float64 `json:"noul"`
	}
	if decodeClassifierJSON(raw, &wire) != nil || wire.Type != "noul" || wire.Noul == nil || !unitProbability(*wire.Noul) {
		return BoolAnswer{}, classifierProtocolFailure()
	}
	return BoolAnswer{Probability: *wire.Noul}, nil
}

func decodeScoreAnswer(raw json.RawMessage, q ScoreQuestion) (ScoreAnswer, *Error) {
	var wire struct {
		Type          string          `json:"type"`
		Score         *float64        `json:"score"`
		Confidence    *float64        `json:"confidence"`
		Probabilities json.RawMessage `json:"probabilities"`
		Legend        json.RawMessage `json:"legend"`
	}
	if decodeClassifierJSON(raw, &wire) != nil || wire.Type != "score" || wire.Score == nil || wire.Confidence == nil || !unitProbability(*wire.Confidence) {
		return ScoreAnswer{}, classifierProtocolFailure()
	}
	if math.IsNaN(*wire.Score) || math.IsInf(*wire.Score, 0) || *wire.Score < 0 || *wire.Score > float64(len(q.Criteria)-1) {
		return ScoreAnswer{}, classifierProtocolFailure()
	}
	a := ScoreAnswer{Score: *wire.Score, Confidence: *wire.Confidence}
	if len(wire.Probabilities) > 0 {
		var probabilities map[string]*float64
		if decodeClassifierJSON(wire.Probabilities, &probabilities) != nil || len(probabilities) != len(q.Criteria) {
			return ScoreAnswer{}, classifierProtocolFailure()
		}
		a.Probabilities = make(map[int]float64, len(probabilities))
		sum, expected := 0.0, 0.0
		for i := range q.Criteria {
			p := probabilities[strconv.Itoa(i)]
			if p == nil || !unitProbability(*p) {
				return ScoreAnswer{}, classifierProtocolFailure()
			}
			a.Probabilities[i] = *p
			sum += *p
			expected += float64(i) * *p
		}
		if math.Abs(sum-1) > 1e-6 || (math.Abs(expected-a.Score) > 1e-6 && !typeSafeRoundedScoreConsistent(a.Score, a.Probabilities)) {
			return ScoreAnswer{}, classifierProtocolFailure()
		}
	}
	if len(wire.Legend) > 0 {
		var legend map[string]json.RawMessage
		if a.Probabilities == nil || decodeClassifierJSON(wire.Legend, &legend) != nil || len(legend) != len(q.Criteria) {
			return ScoreAnswer{}, classifierProtocolFailure()
		}
		a.Legend = make(map[int]json.RawMessage, len(legend))
		for i, criterion := range q.Criteria {
			value, ok := legend[strconv.Itoa(i)]
			if !ok || !classifierJSONEqual(value, criterion) {
				return ScoreAnswer{}, classifierProtocolFailure()
			}
			a.Legend[i] = value
		}
	}
	return a, nil
}

// Compare native rubric JSON by value, ignoring object order and formatting.
// Decimal coefficients and exponents stay symbolic; even a huge exponent
// requires space proportional only to its JSON representation.
func classifierJSONEqual(a, b json.RawMessage) bool {
	decode := func(raw json.RawMessage) (any, error) {
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		var value any
		err := d.Decode(&value)
		return value, err
	}
	x, err := decode(a)
	if err != nil {
		return false
	}
	y, err := decode(b)
	return err == nil && classifierJSONValueEqual(x, y)
}

func classifierJSONValueEqual(a, b any) bool {
	switch x := a.(type) {
	case json.Number:
		y, ok := b.(json.Number)
		if !ok {
			return false
		}
		return classifierDecimal(x) == classifierDecimal(y)
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for key, value := range x {
			other, ok := y[key]
			if !ok || !classifierJSONValueEqual(value, other) {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i, value := range x {
			if !classifierJSONValueEqual(value, y[i]) {
				return false
			}
		}
		return true
	case string:
		y, ok := b.(string)
		return ok && x == y
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	default:
		return a == nil && b == nil
	}
}

// Inputs have already passed JSON validation. Normalize the coefficient and
// decimal exponent without constructing 10^exponent or rounding to float64.
func classifierDecimal(n json.Number) string {
	value := strings.ToLower(string(n))
	mantissa, power, hasPower := strings.Cut(value, "e")
	exponent := new(big.Int)
	if hasPower {
		exponent.SetString(power, 10)
	}
	negative := strings.HasPrefix(mantissa, "-")
	mantissa = strings.TrimPrefix(mantissa, "-")
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		exponent.Sub(exponent, big.NewInt(int64(len(mantissa)-dot-1)))
		mantissa = mantissa[:dot] + mantissa[dot+1:]
	}
	mantissa = strings.TrimLeft(mantissa, "0")
	if mantissa == "" {
		return "0"
	}
	trimmed := strings.TrimRight(mantissa, "0")
	exponent.Add(exponent, big.NewInt(int64(len(mantissa)-len(trimmed))))
	if negative {
		trimmed = "-" + trimmed
	}
	return trimmed + "e" + exponent.String()
}

package ai

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// kimiCount keeps absence/null apart from an explicit zero. Raw numeric
// syntax is checked exactly before conversion, including the int64 boundary.
type kimiCount struct {
	n       int64
	present bool
}

type kimiUsageReport struct {
	input, output, read, write, reasoning, total, short, long kimiCount
}

func kimiUsage(model Model, raw json.RawMessage, actual json.RawMessage) (Usage, UsageReporting, CostEstimate, *Error) {
	unreported := CostEstimate{Status: CostUnreported}
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Usage{}, UsageUnreported, unreported, nil
	}
	r, err := decodeKimiUsage(model.API, raw)
	if err != nil {
		return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
	}
	u := Usage{Input: r.input.n, Output: r.output.n, CacheRead: r.read.n, CacheWrite: r.write.n}
	cacheTotal, ok := kimiSum(r.read.n, r.write.n)
	if !ok {
		return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
	}
	if model.API != APIAnthropicMessages {
		if r.input.present && cacheTotal > r.input.n {
			return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
		}
		if r.input.present {
			u.Input -= cacheTotal
		}
	}
	if r.reasoning.present {
		if r.output.present && r.reasoning.n > r.output.n {
			return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
		}
		u.Reasoning = Value(r.reasoning.n)
	} else if model.API != APIAnthropicMessages {
		u.Reasoning = Value(int64(0))
	}
	sum, ok := kimiSum(u.Input, u.Output, u.CacheRead, u.CacheWrite)
	if !ok {
		return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
	}
	u.TotalTokens = sum
	billable := r.input.present && r.output.present && r.read.present && r.write.present
	if r.total.present {
		if billable && r.total.n != sum {
			return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
		}
		u.TotalTokens = r.total.n
	}
	reporting := UsagePartial
	if billable && (model.API == APIAnthropicMessages || r.reasoning.present) && (model.API != APIOpenAIResponses || r.total.present) {
		reporting = UsageComplete
	}
	// Response actual options describe the server's applied tier, never the
	// request preference. Chat cannot supply this evidence.
	if model.API == APIOpenAIResponses && r.write.present && len(actual) > 0 {
		var applied struct {
			Mode string `json:"mode"`
			TTL  string `json:"ttl"`
		}
		if json.Unmarshal(actual, &applied) == nil && applied.Mode == "implicit" {
			if applied.TTL == "5m" {
				r.short = kimiCount{r.write.n, true}
			}
			if applied.TTL == "1h" {
				r.long = kimiCount{r.write.n, true}
			}
		}
	}
	ttlSum, ok := kimiSum(r.short.n, r.long.n)
	if !ok || (r.write.present && (ttlSum > r.write.n || (r.short.present && r.long.present && ttlSum != r.write.n))) {
		return Usage{}, UsageUnreported, unreported, kimiUsageFailure()
	}
	if r.long.present && r.write.present {
		u.CacheWrite1h = Value(r.long.n)
	}
	u.Cost = model.Cost.estimate(u)
	estimate := CostEstimate{Status: CostPartial}
	if billable {
		unknown := r.write.n - ttlSum
		lower := u.Cost
		upperUsage := u
		// A bound is a pricing calculation, not an observed one-hour counter.
		upperUsage.CacheWrite1h = Value(r.long.n + unknown)
		upper := model.Cost.estimate(upperUsage)
		estimate = CostEstimate{Status: CostResolved, Minimum: Value(lower), Maximum: Value(upper)}
		if unknown > 0 {
			estimate.Status = CostBounded
		}
	}
	return u, reporting, estimate, nil
}

func kimiUsageFailure() *Error {
	return newError(CodeProtocol, PhaseStream, "Kimi usage counts are invalid or contradictory")
}

func kimiSum(values ...int64) (int64, bool) {
	var sum int64
	for _, v := range values {
		if v < 0 || sum > math.MaxInt64-v {
			return 0, false
		}
		sum += v
	}
	return sum, true
}

// parseKimiCount rejects malformed or fractional tokens rather than silently
// truncating them. Float parsing bounds exponent work before exact rational
// validation, so adversarial exponents cannot allocate unbounded big integers.
func parseKimiCount(raw json.RawMessage) (kimiCount, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return kimiCount{}, nil
	}
	invalid := func() (kimiCount, error) { return kimiCount{}, kimiUsageFailure() }
	if raw[0] == '"' || raw[0] == 't' || raw[0] == 'f' || !json.Valid(raw) {
		return invalid()
	}
	s := string(raw)
	if v, err := strconv.ParseInt(s, 10, 64); err == nil && v >= 0 {
		return kimiCount{v, true}, nil
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(n, 0) || n < 0 || math.Trunc(n) != n || n > float64(math.MaxInt64) {
		return invalid()
	}
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		e, err := strconv.ParseInt(s[i+1:], 10, 32)
		if err != nil || e > 256 || e < -256 {
			return invalid()
		}
	}
	exact, ok := new(big.Rat).SetString(s)
	if !ok || !exact.IsInt() || !exact.Num().IsInt64() || exact.Sign() < 0 {
		return invalid()
	}
	return kimiCount{exact.Num().Int64(), true}, nil
}

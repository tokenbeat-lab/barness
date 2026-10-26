package ai

import (
	"bytes"
	"encoding/json"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// xAI owns presence and arithmetic independently of the other providers: a
// missing cache/reasoning count cannot inherit their protocol defaults.
type xaiCount struct {
	n       int64
	present bool
}
type xaiUsageReport struct{ input, output, total, cache, reasoning xaiCount }

func xaiUsage(model Model, raw json.RawMessage) (Usage, UsageReporting, CostEstimate, *Error) {
	unreported := CostEstimate{Status: CostUnreported}
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return Usage{}, UsageUnreported, unreported, nil
	}
	r, ok := xaiDecodeUsage(raw)
	if !ok {
		return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
	}
	if !r.input.present && !r.output.present && !r.total.present && !r.cache.present && !r.reasoning.present {
		return Usage{}, UsageUnreported, unreported, nil
	}
	u := Usage{Input: r.input.n, CacheRead: r.cache.n, Output: max(r.output.n, r.reasoning.n), TotalTokens: r.total.n}
	if r.reasoning.present {
		u.Reasoning = Value(r.reasoning.n)
	}
	if r.input.present {
		if r.cache.n > r.input.n {
			return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
		}
		u.Input -= r.cache.n
	}
	known := r.input.present && r.total.present
	if known {
		if r.total.n < r.input.n {
			return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
		}
		b := r.total.n - r.input.n
		if r.reasoning.n > b || r.output.n > b {
			return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
		}
		if r.output.present && r.reasoning.present && b != r.output.n && b-r.output.n != r.reasoning.n {
			return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
		}
		u.Output = b
	} else if r.total.present {
		// Even incomplete reports must not contradict the known lower bound.
		if r.cache.n > r.total.n || u.Output > r.total.n-r.cache.n {
			return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
		}
	}
	reporting := UsagePartial
	if r.input.present && r.output.present && r.total.present && r.cache.present && r.reasoning.present {
		reporting = UsageComplete
	}
	u.Cost = model.Cost.estimate(u)
	if !xaiFiniteCost(u.Cost) {
		return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
	}
	estimate := CostEstimate{Status: CostPartial}
	if known {
		low, high := u.Cost, u.Cost
		if !r.cache.present {
			allCache := u
			allCache.Input, allCache.CacheRead = 0, r.input.n
			other := model.Cost.estimate(allCache)
			if !xaiFiniteCost(other) {
				return Usage{}, UsageUnreported, unreported, xaiUsageFailure()
			}
			if other.Total < low.Total {
				low = other
			}
			if other.Total > high.Total {
				high = other
			}
		}
		estimate.Status = CostBounded
		if low.Total == high.Total {
			estimate.Status = CostResolved
			high = low
		}
		estimate.Minimum, estimate.Maximum = Value(low), Value(high)
		u.Cost = low
	}
	return u, reporting, estimate, nil
}
func xaiDecodeUsage(raw json.RawMessage) (xaiUsageReport, bool) {
	var r xaiUsageReport
	obj, ok := xaiStreamObject(raw)
	if !ok {
		return r, false
	}
	for name, target := range map[string]*xaiCount{"input_tokens": &r.input, "output_tokens": &r.output, "total_tokens": &r.total} {
		count, ok := xaiUsageCount(obj[name])
		if !ok {
			return r, false
		}
		*target = count
	}
	for _, detail := range []struct {
		name, field string
		target      *xaiCount
	}{{"input_tokens_details", "cached_tokens", &r.cache}, {"output_tokens_details", "reasoning_tokens", &r.reasoning}} {
		raw := bytes.TrimSpace(obj[detail.name])
		if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
			continue
		}
		fields, ok := xaiStreamObject(raw)
		if !ok {
			return r, false
		}
		for name, value := range fields {
			count, ok := xaiUsageCount(value)
			if !ok {
				return r, false
			}
			if name == detail.field {
				*detail.target = count
			} else if count.present && count.n > 0 {
				return r, false
			}
		}
	}
	if value, has := obj["cache_write_tokens"]; has {
		count, ok := xaiUsageCount(value)
		if !ok || count.n > 0 {
			return r, false
		}
	}
	// Host functions are the only tool class in this route. Native tool usage
	// would add a billable category that this estimate cannot represent.
	if value, has := obj["server_side_tool_usage"]; has && string(value) != "null" && string(value) != "{}" && string(value) != "[]" {
		return r, false
	}
	return r, true
}
func xaiUsageFailure() *Error {
	return newError(CodeProtocol, PhaseStream, "xAI usage counts or costs are invalid or contradictory")
}

// Bound exponent work before exact rational conversion. Float64 alone would
// accept fractional source numbers rounded to integers or the int64 boundary.
func xaiUsageCount(raw json.RawMessage) (xaiCount, bool) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return xaiCount{}, true
	}
	if !json.Valid(raw) || raw[0] == '"' || raw[0] == 't' || raw[0] == 'f' {
		return xaiCount{}, false
	}
	s := string(raw)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil && n >= 0 {
		return xaiCount{n, true}, true
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsInf(n, 0) || n < 0 || math.Trunc(n) != n || n > float64(math.MaxInt64) {
		return xaiCount{}, false
	}
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		e, err := strconv.ParseInt(s[i+1:], 10, 32)
		if err != nil || e > 256 || e < -256 {
			return xaiCount{}, false
		}
	}
	exact, ok := new(big.Rat).SetString(s)
	if !ok || !exact.IsInt() || !exact.Num().IsInt64() || exact.Sign() < 0 {
		return xaiCount{}, false
	}
	return xaiCount{exact.Num().Int64(), true}, true
}
func xaiFiniteCost(cost UsageCost) bool {
	for _, v := range []float64{cost.Input, cost.Output, cost.CacheRead, cost.CacheWrite, cost.Total} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return false
		}
	}
	return true
}

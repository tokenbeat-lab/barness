package ai

import (
	"bytes"
	"encoding/json"
	"math"
	"strings"
)

// Decisions reports inclusive input and free output. Cache/reasoning details
// validate that report and its completeness; cache never becomes a second
// canonical bucket or a second charge. This conversion intentionally differs
// from chat and TypeSafe, whose wire counts and presence have other meanings.
func decisionsUsage(raw json.RawMessage, cost ModelCost) (UsageReporting, Usage, *Error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return UsageUnreported, Usage{}, nil
	}
	fields, failure := decisionsUsageFields(raw, "input_tokens", "output_tokens", "total_tokens", "input_tokens_details", "output_tokens_details")
	if failure != nil {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	inputDetails, failure := decisionsUsageFields(fields["input_tokens_details"], "cached_tokens", "cache_write_tokens")
	if failure != nil {
		return UsageUnreported, Usage{}, failure
	}
	outputDetails, failure := decisionsUsageFields(fields["output_tokens_details"], "reasoning_tokens")
	if failure != nil {
		return UsageUnreported, Usage{}, failure
	}
	u := Usage{}
	var cached, written, reasoning int64
	const (
		inputCount = iota
		outputCount
		totalCount
		cachedCount
		writtenCount
		reasoningCount
	)
	counts := [6]struct {
		raw      json.RawMessage
		value    *int64
		reported bool
	}{
		inputCount:     {raw: fields["input_tokens"], value: &u.Input},
		outputCount:    {raw: fields["output_tokens"], value: &u.Output},
		totalCount:     {raw: fields["total_tokens"], value: &u.TotalTokens},
		cachedCount:    {raw: inputDetails["cached_tokens"], value: &cached},
		writtenCount:   {raw: inputDetails["cache_write_tokens"], value: &written},
		reasoningCount: {raw: outputDetails["reasoning_tokens"], value: &reasoning},
	}
	reporting := UsageComplete
	for i := range counts {
		count := &counts[i]
		value := bytes.TrimSpace(count.raw)
		if len(value) == 0 || bytes.Equal(value, []byte("null")) {
			reporting = UsagePartial
			continue
		}
		if json.Unmarshal(value, count.value) != nil || *count.value < 0 {
			return UsageUnreported, Usage{}, classifierProtocolFailure()
		}
		count.reported = true
	}
	if u.Input > math.MaxInt64-u.Output || cached > math.MaxInt64-written {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	if counts[inputCount].reported && cached+written > u.Input || counts[outputCount].reported && reasoning > u.Output {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	if counts[totalCount].reported {
		if counts[inputCount].reported && counts[outputCount].reported && u.TotalTokens != u.Input+u.Output {
			return UsageUnreported, Usage{}, classifierProtocolFailure()
		}
	} else {
		u.TotalTokens = u.Input + u.Output
	}
	if counts[reasoningCount].reported {
		u.Reasoning = Value(reasoning)
	}
	// The authorized classifier's snapshot selects a full-request tier from
	// inclusive input alone, even if the host supplied nonzero free rates.
	priced := cost.estimate(Usage{Input: u.Input})
	u.Cost = UsageCost{Input: priced.Input, Total: priced.Input}
	return reporting, u, nil
}

// Missing/null detail objects are partial, while any other non-object is
// invalid. Additive vendor metadata is allowed; known fields require exact
// names, and duplicate keys anywhere in usage invalidate the whole report.
func decisionsUsageFields(raw json.RawMessage, known ...string) (map[string]json.RawMessage, *Error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var fields map[string]json.RawMessage
	if decodeClassifierJSON(raw, &fields) != nil || fields == nil {
		return nil, classifierProtocolFailure()
	}
	for key := range fields {
		for _, name := range known {
			if key != name && strings.EqualFold(key, name) {
				return nil, classifierProtocolFailure()
			}
		}
	}
	return fields, nil
}

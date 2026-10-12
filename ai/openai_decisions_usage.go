package ai

import (
	"bytes"
	"encoding/json"
	"math"
)

// This first text slice preserves known basic counts. Missing detail fields
// never imply complete reporting; full presence/details are owned by issue 04.
func decisionsUsage(raw json.RawMessage, cost ModelCost) (UsageReporting, Usage, *Error) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return UsageUnreported, Usage{}, nil
	}
	var fields map[string]json.RawMessage
	if decodeClassifierJSON(raw, &fields) != nil || fields == nil {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	u := Usage{}
	counts := []struct {
		name  string
		value *int64
	}{{"input_tokens", &u.Input}, {"output_tokens", &u.Output}, {"total_tokens", &u.TotalTokens}}
	present := map[string]bool{}
	for _, count := range counts {
		value := bytes.TrimSpace(fields[count.name])
		if len(value) == 0 || bytes.Equal(value, []byte("null")) {
			continue
		}
		if json.Unmarshal(value, count.value) != nil || *count.value < 0 {
			return UsageUnreported, Usage{}, classifierProtocolFailure()
		}
		present[count.name] = true
	}
	if u.Input > math.MaxInt64-u.Output {
		return UsageUnreported, Usage{}, classifierProtocolFailure()
	}
	if present["total_tokens"] {
		if present["input_tokens"] && present["output_tokens"] && u.TotalTokens != u.Input+u.Output {
			return UsageUnreported, Usage{}, classifierProtocolFailure()
		}
	} else {
		u.TotalTokens = u.Input + u.Output
	}
	// Preserve host input tiers without borrowing or altering catalog rates.
	priced := cost.estimate(Usage{Input: u.Input})
	u.Cost = UsageCost{Input: priced.Input, Total: priced.Input}
	return UsagePartial, u, nil
}

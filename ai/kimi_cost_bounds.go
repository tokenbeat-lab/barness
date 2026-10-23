package ai

import "encoding/json"

// Missing writes have a bounded allocation when Chat/Responses report total
// input: 0 <= W <= P-R, with uncached input P-R-W. Linear prices reach their
// extrema at these endpoints. All scenarios preserve the input-side total and
// therefore retain the fixed Host rates. These are prices, never observations.
func kimiCostBounds(prices ModelCost, usage Usage, report kimiUsageReport, actual json.RawMessage) (CostEstimate, *Error) {
	var scenarios []Usage
	if report.write.present {
		unknown := report.write.n - report.short.n - report.long.n
		short, long := usage, usage
		short.CacheWrite1h = Value(report.long.n)
		long.CacheWrite1h = Value(report.long.n + unknown)
		scenarios = []Usage{short, long}
	} else {
		scenarios = append(scenarios, usage)
		allWrite := usage
		allWrite.Input, allWrite.CacheWrite = 0, usage.Input
		ttl := kimiAppliedTTL(actual)
		if ttl != "1h" {
			scenarios = append(scenarios, allWrite)
		}
		if ttl != "5m" {
			allWrite.CacheWrite1h = Value(allWrite.CacheWrite)
			scenarios = append(scenarios, allWrite)
		}
	}
	var minimum, maximum UsageCost
	for i, scenario := range scenarios {
		cost := prices.estimate(scenario)
		if !kimiFiniteCost(cost) {
			return CostEstimate{}, kimiCostFailure()
		}
		if i == 0 || cost.Total < minimum.Total {
			minimum = cost
		}
		if i == 0 || cost.Total > maximum.Total {
			maximum = cost
		}
	}
	status := CostBounded
	if minimum.Total == maximum.Total {
		// Equal total amounts are one reliable estimate even when feasible input
		// classifications differ. Use the same component snapshot for both bounds.
		status, maximum = CostResolved, minimum
	}
	return CostEstimate{Status: status, Minimum: Value(minimum), Maximum: Value(maximum)}, nil
}

// Only the response can prove an actual applied TTL. The request's preference
// is deliberately absent from this boundary.
func kimiAppliedTTL(raw json.RawMessage) string {
	var applied struct {
		Mode string `json:"mode"`
		TTL  string `json:"ttl"`
	}
	if json.Unmarshal(raw, &applied) == nil && applied.Mode == "implicit" && (applied.TTL == "5m" || applied.TTL == "1h") {
		return applied.TTL
	}
	return ""
}

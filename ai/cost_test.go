package ai

import (
	"math"
	"testing"
)

// TestCostEstimate checks the cost rule ported from pi-ai's calculateCost
// (spec I10) on the inputs no E2E can produce: catalogs with several tiers,
// with and without one-hour cache writes. TestUsageAndCost (Responses) and
// TestAnthropicUsageAndRetry (one-hour writes, P02) cover the rest end to
// end. This is the one isolated
// test. Expectations are independent literal monetary examples.
//
// Ways it could fail:
//
//	C1 a rate is applied per token instead of per million, or to the wrong
//	   count
//	C2 one-hour cache writes are priced at the cache-write rate instead of
//	   twice the input rate, or are priced twice (also as short writes)
//	C3 a tier applies at its threshold, counts input alone, or the first
//	   matching tier wins over the highest threshold exceeded
//	C4 reasoning tokens are priced on top of output
//	C5 the total is not the sum of the independently calculated parts
//	C6 a zero usage costs anything
func TestCostEstimate(t *testing.T) {
	base := ModelCost{CostRates: CostRates{Input: 3, Output: 15, CacheRead: 0.3, CacheWrite: 3.75}}
	tiered := ModelCost{CostRates: CostRates{Input: 2, Output: 8, CacheRead: 0.2, CacheWrite: 2.5}, Tiers: []CostTier{
		{InputTokensAbove: 500000, CostRates: CostRates{Input: 6, Output: 24, CacheRead: 0.6, CacheWrite: 7.5}},
		{InputTokensAbove: 200000, CostRates: CostRates{Input: 4, Output: 16, CacheRead: 0.4, CacheWrite: 5}},
	}}
	cases := []struct {
		name  string
		cost  ModelCost
		usage Usage
		want  UsageCost
	}{
		{"zero", base, Usage{}, UsageCost{}},
		{"base-rates", base, Usage{Input: 1000, Output: 200, CacheRead: 4000},
			UsageCost{Input: 0.003, Output: 0.003, CacheRead: 0.0012, Total: 0.0072}},
		{"short-cache-write", base, Usage{Input: 10, CacheWrite: 2000},
			UsageCost{Input: 0.00003, CacheWrite: 0.0075, Total: 0.00753}},
		{"reasoning-not-priced-again", base, Usage{Output: 100, Reasoning: Value[int64](60)},
			UsageCost{Output: 0.0015, Total: 0.0015}},
		{"below-tiers", tiered, Usage{Input: 200000, Output: 10},
			UsageCost{Input: 0.4, Output: 0.00008, Total: 0.40008}},
		{"lower-tier", tiered, Usage{Input: 150000, CacheRead: 50000, CacheWrite: 1, Output: 10},
			UsageCost{Input: 0.6, Output: 0.00016, CacheRead: 0.02, CacheWrite: 0.000005, Total: 0.620165}},
		{"highest-tier-exceeded", tiered, Usage{Input: 500001, Output: 10},
			UsageCost{Input: 3.000006, Output: 0.00024, Total: 3.000246}},
		{"tier-one-hour-write", tiered, Usage{Input: 300000, CacheWrite: 1000, CacheWrite1h: Value[int64](1000)},
			UsageCost{Input: 1.2, CacheWrite: 0.008, Total: 1.208}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.cost.estimate(c.usage)
			if !closeCost(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
			if got.Total != got.Input+got.Output+got.CacheRead+got.CacheWrite {
				t.Errorf("total %v is not the sum of %+v", got.Total, got)
			}
		})
	}

}

// closeCost compares two costs to a relative 1e-12: the hand-written
// expectations are worked decimal monetary examples.
func closeCost(a, b UsageCost) bool {
	near := func(x, y float64) bool { return math.Abs(x-y) <= 1e-12*math.Max(1, math.Max(math.Abs(x), math.Abs(y))) }
	return near(a.Input, b.Input) && near(a.Output, b.Output) && near(a.CacheRead, b.CacheRead) &&
		near(a.CacheWrite, b.CacheWrite) && near(a.Total, b.Total)
}

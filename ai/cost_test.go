package ai

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// TestCostEstimate checks the cost rule ported from pi-ai's calculateCost
// (spec I10) on the inputs no Responses E2E can produce: one-hour cache
// writes (only Anthropic reports them) and catalogs with several tiers.
// TestUsageAndCost covers the rest end to end. This is the one isolated
// test; with BARNESS_AI_PIDIFF=1 every case is also checked bit for bit
// against frozen pi-ai's own calculateCost, and the built-in catalog's
// prices against pi's frozen model data.
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
//	C5 the total is not the sum of the parts, or a component differs from
//	   pi's in the last bit (operation order, fused multiply-add)
//	C6 a zero usage costs anything
//	C7 a built-in price differs from pi's frozen model data
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
		{"one-hour-cache-write", base, Usage{Input: 10, CacheWrite: 2000, CacheWrite1h: Value[int64](2000)},
			UsageCost{Input: 0.00003, CacheWrite: 0.012, Total: 0.01203}},
		{"mixed-cache-write", base, Usage{CacheWrite: 3000, CacheWrite1h: Value[int64](1000)},
			UsageCost{CacheWrite: 0.0135, Total: 0.0135}},
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

	var pi pioracle.CostResults
	if pioracle.Enabled() {
		o, err := pioracle.Open()
		if err != nil {
			t.Fatal(err)
		}
		in := make([]pioracle.CostCase, len(cases))
		for i, c := range cases {
			u := c.usage
			u.Cost = UsageCost{}
			in[i] = pioracle.CostCase{Cost: mustJSON(t, c.cost), Usage: mustJSON(t, u)}
		}
		var ids []string
		for _, m := range BuiltinCatalog().Models {
			ids = append(ids, m.ID)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if pi, err = o.Costs(ctx, in, ids); err != nil {
			t.Fatal(err)
		}
	}

	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.cost.estimate(c.usage)
			if !closeCost(got, c.want) {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
			if got.Total != got.Input+got.Output+got.CacheRead+got.CacheWrite {
				t.Errorf("total %v is not the sum of %+v", got.Total, got)
			}
			if pi.Results != nil {
				var want UsageCost
				if err := json.Unmarshal(pi.Results[i], &want); err != nil {
					t.Fatal(err)
				}
				if got != want {
					t.Errorf("pi gives %+v, barness %+v", want, got)
				}
			}
		})
	}

	t.Run("builtin-prices-are-pi's", func(t *testing.T) {
		if pi.Models == nil {
			t.Skipf("frozen pi differential not requested (set %s=1)", pioracle.EnableEnv)
		}
		for _, m := range BuiltinCatalog().Models {
			var want ModelCost
			if err := json.Unmarshal(pi.Models[m.ID], &want); err != nil || pi.Models[m.ID] == nil {
				t.Errorf("%s: pi has no cost entry (%v)", m.ID, err)
				continue
			}
			if !reflect.DeepEqual(m.Cost, want) {
				t.Errorf("%s: barness %+v, pi %+v", m.ID, m.Cost, want)
			}
		}
	})
}

// closeCost compares two costs to a relative 1e-12: the hand-written
// expectations are decimals, not pi's exact binary values.
func closeCost(a, b UsageCost) bool {
	near := func(x, y float64) bool { return math.Abs(x-y) <= 1e-12*math.Max(1, math.Max(math.Abs(x), math.Abs(y))) }
	return near(a.Input, b.Input) && near(a.Output, b.Output) && near(a.CacheRead, b.CacheRead) &&
		near(a.CacheWrite, b.CacheWrite) && near(a.Total, b.Total)
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

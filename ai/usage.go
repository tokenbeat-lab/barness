package ai

import "math"

// Usage is provider-reported token usage converted to pi-ai's shape, with
// pi's estimated cost. A message starts with the zero value, which is also
// what it keeps when the provider reported nothing: zero values never imply
// the call was free or that usage was reported. How completely each attempt
// reported is in CallMetadata.Attempts (see UsageReporting).
type Usage struct {
	// Input excludes the tokens counted in CacheRead and CacheWrite.
	Input      int64 `json:"input"`
	Output     int64 `json:"output"`
	CacheRead  int64 `json:"cacheRead"`
	CacheWrite int64 `json:"cacheWrite"`
	// CacheWrite1h is the part of CacheWrite written with a one-hour
	// lifetime, which costs twice the input rate. Only protocols that report
	// it set it; Responses never does.
	CacheWrite1h Nullable[int64] `json:"cacheWrite1h,omitzero"`
	// Reasoning is the part of Output spent on reasoning: already counted in
	// Output and TotalTokens and priced as output, never added again. As in
	// pi, it is set (possibly 0) whenever the provider reported usage.
	Reasoning Nullable[int64] `json:"reasoning,omitzero"`
	// TotalTokens is the provider's own total.
	TotalTokens int64     `json:"totalTokens"`
	Cost        UsageCost `json:"cost"`
}

// UsageCost is the estimated cost of a Usage in US dollars, computed from
// the catalog the call was priced with (CallMetadata.CatalogVersion). It is
// an estimate, not a billing promise.
type UsageCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
	Total      float64 `json:"total"`
}

// UsageReporting is how completely the provider reported one attempt's
// usage. It lives beside pi's Usage instead of in it, so the message keeps
// pi's numbers while a zero can still be told from nothing reported.
type UsageReporting string

const (
	// UsageUnreported: the attempt reported no usage — it failed before a
	// response, its stream ended without one, the field was missing or
	// null, or the response was a failure whose usage pi does not take
	// over. Its Usage is zero, which says nothing about what it consumed.
	UsageUnreported UsageReporting = "unreported"
	// UsagePartial: usage was reported without some of the counts the
	// protocol defines; the missing ones read as 0, so the cost may be off.
	UsagePartial UsageReporting = "partial"
	// UsageComplete: every count the protocol defines was reported, zeros
	// included.
	UsageComplete UsageReporting = "complete"
)

// CostRates are prices in US dollars per million tokens.
type CostRates struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

// ModelCost is a model's price, in pi-ai's model data shape: base rates and
// optional tiers.
type ModelCost struct {
	CostRates
	// Tiers replace the base rates when a request's input-side tokens
	// (input, cache reads and cache writes together) exceed a threshold.
	Tiers []CostTier `json:"tiers,omitempty"`
}

// CostTier is a set of rates that applies strictly above InputTokensAbove
// input-side tokens; of several, the highest threshold exceeded wins.
type CostTier struct {
	InputTokensAbove int64 `json:"inputTokensAbove"`
	CostRates
}

// problem describes the first price that cannot yield a finite, non-negative
// cost, or is empty.
func (c ModelCost) problem() string {
	rates := c.CostRates.list()
	for _, t := range c.Tiers {
		if t.InputTokensAbove < 0 {
			return "a cost tier threshold is negative"
		}
		rates = append(rates, t.CostRates.list()...)
	}
	for _, r := range rates {
		if math.IsNaN(r) || math.IsInf(r, 0) || r < 0 {
			return "a price is negative or not finite"
		}
	}
	return ""
}

func (r CostRates) list() []float64 { return []float64{r.Input, r.Output, r.CacheRead, r.CacheWrite} }

// estimate ports pi-ai's calculateCost: u's cost at c's rates.
//
// The arithmetic keeps pi's operation order, and every product is rounded on
// its own (the float64 conversions around each product, not only its
// operands), so the result is the same IEEE value as JavaScript's: Go may
// otherwise fuse a multiply and a later add, even across statements, which
// rounds once and can differ in the last bit.
func (c ModelCost) estimate(u Usage) UsageCost {
	rates := c.CostRates
	inputTokens := u.Input + u.CacheRead + u.CacheWrite
	matched := int64(-1)
	for _, t := range c.Tiers {
		if inputTokens > t.InputTokensAbove && t.InputTokensAbove > matched {
			rates, matched = t.CostRates, t.InputTokensAbove
		}
	}
	// Anthropic charges twice the base input rate for one-hour cache writes.
	long, _ := u.CacheWrite1h.Get()
	short := u.CacheWrite - long
	var cost UsageCost
	cost.Input = float64(float64(rates.Input/1e6) * float64(u.Input))
	cost.Output = float64(float64(rates.Output/1e6) * float64(u.Output))
	cost.CacheRead = float64(float64(rates.CacheRead/1e6) * float64(u.CacheRead))
	cost.CacheWrite = (float64(rates.CacheWrite*float64(short)) + float64(float64(rates.Input*2)*float64(long))) / 1e6
	cost.Total = cost.Input + cost.Output + cost.CacheRead + cost.CacheWrite
	return cost
}

// scaled is cost with every component multiplied by m and the total summed
// again, as pi's per-protocol pricing adjustments do.
func (c UsageCost) scaled(m float64) UsageCost {
	c.Input = float64(c.Input * m)
	c.Output = float64(c.Output * m)
	c.CacheRead = float64(c.CacheRead * m)
	c.CacheWrite = float64(c.CacheWrite * m)
	c.Total = c.Input + c.Output + c.CacheRead + c.CacheWrite
	return c
}

package ai

import "github.com/openai/openai-go/v3/responses"

// responsesUsage converts, classifies and prices one call's Responses
// usage, with the model's rates and the service tier the call requested.
type responsesUsage struct {
	model Model
	// pricesByTier: the provider serves service tiers, so a tier prices
	// the usage.
	pricesByTier bool
	// requestedTier is ResponsesOptions.ServiceTier; unset or null when the
	// call requested none.
	requestedTier Nullable[string]
}

// of converts a terminal response's usage as pi's finalizeResponse does
// and says how completely it was reported. ok is false when the response
// carries no usage (missing or null): pi then keeps the message's zero
// value, whose cost is zero whatever the tier.
func (p responsesUsage) of(r responses.Response) (u Usage, reporting UsageReporting, ok bool) {
	if !r.JSON.Usage.Valid() {
		return Usage{}, UsageUnreported, false
	}
	ru := r.Usage
	cached, cacheWrite := ru.InputTokensDetails.CachedTokens, ru.InputTokensDetails.CacheWriteTokens
	// OpenAI counts cached and cache-write tokens inside input_tokens. A
	// count that is missing or null reads as 0, as pi's `|| 0`.
	u = Usage{
		Input:       max(0, ru.InputTokens-cached-cacheWrite),
		Output:      ru.OutputTokens,
		CacheRead:   cached,
		CacheWrite:  cacheWrite,
		Reasoning:   Value(ru.OutputTokensDetails.ReasoningTokens),
		TotalTokens: ru.TotalTokens,
	}
	u.Cost = p.model.Cost.estimate(u).scaled(serviceTierMultiplier(p.model, p.tier(r)))
	return u, responsesUsageReporting(ru), true
}

// responsesUsageReporting is complete when every count the Responses usage
// object has always defined is present: input, output and total tokens,
// cached tokens and reasoning tokens. cache_write_tokens joined the schema
// later and many responses lack it, so its absence reads as no cache writes
// rather than a partial report (maintainer decision 2026-10-02, ADR-0010).
func responsesUsageReporting(u responses.ResponseUsage) UsageReporting {
	if u.JSON.InputTokens.Valid() && u.JSON.OutputTokens.Valid() && u.JSON.TotalTokens.Valid() &&
		u.JSON.InputTokensDetails.Valid() && u.InputTokensDetails.JSON.CachedTokens.Valid() &&
		u.JSON.OutputTokensDetails.Valid() && u.OutputTokensDetails.JSON.ReasoningTokens.Valid() {
		return UsageComplete
	}
	return UsagePartial
}

// tier is pi's `response.service_tier ?? options.serviceTier`: the tier the
// provider says served the response, else the one the call requested.
func (p responsesUsage) tier(r responses.Response) string {
	if !p.pricesByTier {
		return ""
	}
	if r.JSON.ServiceTier.Valid() {
		return string(r.ServiceTier)
	}
	tier, _ := p.requestedTier.Get()
	return tier
}

// serviceTierMultiplier ports pi's OpenAI Responses pricing adjustment
// (getServiceTierCostMultiplier).
func serviceTierMultiplier(m Model, tier string) float64 {
	switch tier {
	case "flex":
		return 0.5
	case "priority":
		if m.ID == "gpt-5.5" {
			return 2.5
		}
		return 2
	default:
		return 1
	}
}

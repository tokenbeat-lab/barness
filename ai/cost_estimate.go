package ai

// CostEstimateStatus describes which amounts a usage report supports.
// It is independent of UsageReporting: reasoning can be absent while all
// billable tokens are known, and cache TTL can be absent with complete tokens.
type CostEstimateStatus string

const (
	CostResolved   CostEstimateStatus = "resolved"
	CostBounded    CostEstimateStatus = "bounded"
	CostPartial    CostEstimateStatus = "partial"
	CostUnreported CostEstimateStatus = "unreported"
)

// CostEstimate is a USD estimate, not an invoice. Only resolved and bounded
// carry amounts. On token-priced routes bounded Usage.Cost is the lower
// bound. xAI per-image prices live solely here and do not populate token
// Usage.Cost, even when token usage is absent. Budget reservations use Maximum.
// Its value-only fields need no shared ownership or cloning.
type CostEstimate struct {
	Status  CostEstimateStatus  `json:"status"`
	Minimum Nullable[UsageCost] `json:"minimum,omitzero"`
	Maximum Nullable[UsageCost] `json:"maximum,omitzero"`
}

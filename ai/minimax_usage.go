package ai

// MiniMax's dated M3 standard table supplies no cache-write rate. Active
// writes are excluded (spec R17–R19); a positive automatic write count is
// retained but cannot become a successful zero-cost estimate. Synthetic Host
// catalogs with an explicit positive rate still test usage normalization.
func miniMaxUnpricedWriteFailure(model Model, attempts []Attempt) *Error {
	if model.Provider != ProviderMiniMaxCN || model.Cost.CacheWrite != 0 {
		return nil
	}
	for _, tier := range model.Cost.Tiers {
		if tier.CacheWrite != 0 {
			return nil
		}
	}
	for _, attempt := range attempts {
		if attempt.Usage.CacheWrite > 0 {
			return newError(CodeProtocol, PhaseStream, "MiniMax automatic cache-write pricing is unverified")
		}
	}
	return nil
}

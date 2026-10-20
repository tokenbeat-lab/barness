package ai

// Regional standard tables supply no verified cache-write rate. This is a
// pricing-contract guard scoped to real service identities, not codec equality:
// positive automatic writes retain usage but cannot claim free success. A
// Host catalog with an explicit positive selected rate remains usable.
func providerUnpricedWriteFailure(model Model, attempts []Attempt) *Error {
	if model.Provider != ProviderMiniMax && model.Provider != ProviderMiniMaxCN && model.Provider != ProviderZAI {
		return nil
	}
	for _, attempt := range attempts {
		long, _ := attempt.Usage.CacheWrite1h.Get()
		if long > 0 {
			if model.Provider == ProviderZAI {
				return newError(CodeProtocol, PhaseStream, "Z.AI one-hour cache-write pricing is unverified")
			}
			return newError(CodeProtocol, PhaseStream, "MiniMax one-hour cache-write pricing is unverified")
		}
		if attempt.Usage.CacheWrite > 0 && model.Cost.ratesForUsage(attempt.Usage).CacheWrite == 0 {
			if model.Provider == ProviderZAI {
				return newError(CodeProtocol, PhaseStream, "Z.AI automatic cache-write pricing is unverified")
			}
			return newError(CodeProtocol, PhaseStream, "MiniMax automatic cache-write pricing is unverified")
		}
	}
	return nil
}

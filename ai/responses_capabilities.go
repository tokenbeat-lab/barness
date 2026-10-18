package ai

// responsesCapabilities are the server-side Responses features a provider
// serves beyond the protocol's core (input, tools, reasoning effort, output
// budget). The Responses adapter derives a request field from one of them
// only when the call's provider serves it, so another provider's request
// never carries OpenAI's storage, cache or tier fields by default
// (spec I4, ADR-0014).
//
// The profile is keyed by the binding's provider rather than by model
// compat: it is a fact about the vendor's endpoint, the same for every
// model it serves, and a host's own catalog cannot turn it on for a vendor
// that ignores it. It is a snapshot of the vendor's documentation at the
// time of writing; a change there is updated here and in the P05 fixtures.
type responsesCapabilities struct {
	// store: the request declares stateless storage (store:false), as pi
	// always does for OpenAI.
	store bool
	// encryptedReasoning: requests that set a reasoning effort or summary
	// ask for the encrypted reasoning content (include).
	encryptedReasoning bool
	// Cache routing does not imply a promised lifetime or header affinity.
	promptCacheKey       bool
	promptCacheRetention bool
	sessionAffinity      bool
	// reasoningSummary controls summary independently of encrypted content.
	reasoningSummary bool
	// serviceTier: ServiceTier may be requested, and the tier the response
	// names prices its usage.
	serviceTier bool
	// strictFunctionTools permits the existing explicit strict:false field.
	// Model compatibility can narrow a provider's profile, never expand it.
	strictFunctionTools bool
}

// responsesCapabilitiesOf is p's profile. DeepSeek's Responses guide
// (read 2026-10-02) lists store, include, prompt_cache_key,
// prompt_cache_retention and service_tier as silently ignored, with
// previous_response_id and conversation, which barness-ai never sends;
// history is always sent as complete input. Every other provider gets the
// OpenAI profile, which pi sends to whatever serves openai-responses.
func responsesCapabilitiesOf(p ProviderID) responsesCapabilities {
	if p == ProviderMiniMax {
		return responsesCapabilities{promptCacheKey: true, serviceTier: true}
	}
	if p == ProviderZhipu {
		return responsesCapabilities{store: true, promptCacheKey: true}
	}
	if p == ProviderDeepSeek {
		// DeepSeek accepts summary but does not generate one (ADR-0014).
		return responsesCapabilities{reasoningSummary: true, strictFunctionTools: true}
	}
	return responsesCapabilities{store: true, encryptedReasoning: true, promptCacheKey: true,
		promptCacheRetention: true, sessionAffinity: true, reasoningSummary: true, serviceTier: true, strictFunctionTools: true}
}

// unsupportedOption is why opts asks for something the provider does not
// serve, "" when it asks for nothing such. Only explicit requests are
// refused; what the provider merely ignores is not derived at all.
func (c responsesCapabilities) unsupportedOption(opts ResponsesOptions) string {
	if !c.reasoningSummary && opts.ReasoningSummary != "" {
		return "the provider does not serve a reasoning summary on this protocol"
	}
	if !c.serviceTier && !opts.ServiceTier.IsZero() {
		return "the provider does not serve a service tier on this protocol"
	}
	return ""
}

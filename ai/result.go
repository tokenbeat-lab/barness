package ai

import "slices"

// CallAttribution is who and what one logical call belongs to: the trusted
// scope, the binding it named and, once resolved, what actually serves it.
// Provider, API, model and account are only set once a consistent,
// authorized configuration snapshot was established; a call that failed
// earlier reports Resolved=false instead of echoing request values. It holds
// values only, so a copy never changes.
type CallAttribution struct {
	TenantID  string `json:"tenantId"`
	RequestID string `json:"requestId"`
	ActorID   string `json:"actorId,omitempty"`
	JobID     string `json:"jobId,omitempty"`
	BindingID string `json:"bindingId"`
	// Operation is set by the public entry even before resolving the binding.
	Operation  Operation  `json:"operation"`
	Resolved   bool       `json:"resolved"`
	ProviderID ProviderID `json:"providerId,omitempty"`
	API        API        `json:"api,omitempty"`
	ModelID    string     `json:"modelId,omitempty"`
	// AccountScopeID is the vendor account serving the call.
	AccountScopeID string `json:"accountScopeId,omitempty"`
}

// CallMetadata is the immutable attribution of one logical call together
// with its configuration snapshot, downgrades and attempts.
type CallMetadata struct {
	CallAttribution
	// BindingVersion and CredentialVersion identify the configuration
	// snapshot the call was pinned to. Set once resolved.
	BindingVersion    string `json:"bindingVersion,omitempty"`
	CredentialVersion string `json:"credentialVersion,omitempty"`
	// CatalogVersion and CatalogHash identify the model catalog and price
	// snapshot the call's costs were estimated with (Catalog.Version and
	// Catalog.Hash). Set once resolved.
	CatalogVersion string `json:"catalogVersion,omitempty"`
	CatalogHash    string `json:"catalogHash,omitempty"`
	// NativeStateDowngrades counts history messages whose native state was
	// downgraded instead of replayed, by reason. Set once resolved.
	NativeStateDowngrades NativeStateDowngrades `json:"nativeStateDowngrades,omitzero"`
	// Attempts are the HTTP requests the call sent, in order; retries of the
	// initial request add one each. Empty when nothing was sent. Each
	// records its own usage: the message's Usage is the streamed attempt's
	// as pi reports it, never a sum over attempts.
	Attempts []Attempt `json:"attempts,omitempty"`
}

// Result is the outcome of a call: the final assistant message, which exists
// even for failures, plus call attribution.
type Result struct {
	Message  AssistantMessage `json:"message"`
	Metadata CallMetadata     `json:"metadata"`
}

func (r Result) clone() Result {
	r.Message = r.Message.clone()
	r.Metadata.Attempts = slices.Clone(r.Metadata.Attempts)
	return r
}

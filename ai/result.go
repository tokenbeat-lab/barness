package ai

// CallMetadata is the immutable attribution of one logical call. Provider,
// API, model and account are only set once a consistent, authorized
// configuration snapshot was established; a call that failed earlier reports
// Resolved=false instead of echoing request values.
type CallMetadata struct {
	TenantID   string     `json:"tenantId"`
	RequestID  string     `json:"requestId"`
	ActorID    string     `json:"actorId,omitempty"`
	JobID      string     `json:"jobId,omitempty"`
	BindingID  string     `json:"bindingId"`
	Resolved   bool       `json:"resolved"`
	ProviderID ProviderID `json:"providerId,omitempty"`
	API        API        `json:"api,omitempty"`
	ModelID    string     `json:"modelId,omitempty"`
	// AccountScopeID is the vendor account serving the call.
	AccountScopeID string `json:"accountScopeId,omitempty"`
}

// Result is the outcome of a call: the final assistant message, which exists
// even for failures, plus call attribution.
type Result struct {
	Message  AssistantMessage `json:"message"`
	Metadata CallMetadata     `json:"metadata"`
}

func (r Result) clone() Result {
	r.Message = r.Message.clone()
	return r
}

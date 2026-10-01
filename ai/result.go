package ai

// CallMetadata is the immutable attribution of one logical call. Provider,
// API and model are only set once resolved from the tenant's binding; a call
// that failed earlier reports Resolved=false instead of echoing request values.
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

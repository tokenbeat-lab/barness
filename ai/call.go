package ai

import (
	"context"
	"time"
)

// call is one logical call's validated input, copied on receipt.
type call struct {
	scope  CallScope
	target Target
	req    Request
	full   Options        // full entry points; may be nil
	simple *SimpleOptions // simple entry points; nil for full
}

func newCall(scope CallScope, target Target, req Request, full Options, simple *SimpleOptions) call {
	if simple != nil {
		s := *simple
		simple = &s
	}
	return call{scope: scope, target: target, req: req.clone(), full: full, simple: simple}
}

// run executes a call to its terminal and returns the result. emit receives
// every event when the caller streams; it is nil for Complete.
func (c *Client) run(ctx context.Context, cl call, emit func(Event)) (Result, error) {
	asm := &assembler{msg: AssistantMessage{Timestamp: time.Now().UnixMilli()}, emit: emit}
	meta := CallMetadata{
		TenantID:  cl.scope.TenantID,
		RequestID: cl.scope.RequestID,
		ActorID:   cl.scope.ActorID,
		JobID:     cl.scope.JobID,
		BindingID: cl.target.BindingID,
	}
	failure := c.execute(ctx, cl, &meta, asm)
	return asm.finish(meta, failure)
}

// execute follows spec I3's order: scope → binding → capability/options →
// credential → snapshot consistency → send. Every step before send fails
// without contacting the provider, and identity is only reported resolved
// once the snapshot is consistent. Request structure/size checks (ticket 12)
// and admission (ticket 13) slot in where marked.
func (c *Client) execute(ctx context.Context, cl call, meta *CallMetadata, asm *assembler) *Error {
	if cl.scope.TenantID == "" || cl.scope.RequestID == "" {
		return newError(CodeInvalidRequest, PhaseScope, "call scope requires TenantID and RequestID")
	}
	// Request structure and size checks (ticket 12) go here.

	binding, failure := c.resolveBinding(ctx, cl.scope, cl.target.BindingID)
	if failure != nil {
		return failure
	}
	ad, ok := c.adapters[binding.API]
	if !ok {
		return newError(CodeInvalidRequest, PhaseCapability, "binding API is not supported: "+string(binding.API))
	}
	model, failure := c.authorizeModel(binding, cl.target.ModelID)
	if failure != nil {
		return failure
	}
	options := cl.full
	if cl.simple != nil {
		options = ad.simpleOptions(model, *cl.simple)
	} else if options != nil && options.api() != binding.API {
		return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
	}

	cred, failure := c.resolveCredential(ctx, cl.scope, binding)
	if failure != nil {
		return failure
	}
	if failure := checkSnapshot(cl.scope, binding, cred); failure != nil {
		return failure
	}
	meta.Resolved, meta.ProviderID, meta.API, meta.ModelID = true, binding.ProviderID, binding.API, model.ID
	meta.AccountScopeID = binding.AccountScopeID
	asm.identify(binding.API, binding.ProviderID, model.ID)

	// Admission for the attempt (ticket 13) goes here.

	// From here on the call is pinned to this snapshot: later updates or
	// revocations affect only new logical calls.
	return ad.stream(ctx, adapterCall{
		http:     c.http,
		endpoint: binding.Endpoint,
		apiKey:   cred.APIKey,
		model:    model,
		request:  cl.req,
		options:  options,
	}, asm)
}

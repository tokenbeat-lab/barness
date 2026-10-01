package ai

import (
	"context"
	"net"
	"net/url"
	"slices"
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
// credential → send. Every step before send fails without contacting the
// provider. Snapshot consistency (ticket 03) and admission (ticket 13) slot
// in before send.
func (c *Client) execute(ctx context.Context, cl call, meta *CallMetadata, asm *assembler) *Error {
	if cl.scope.TenantID == "" || cl.scope.RequestID == "" {
		return newError(CodeInvalidRequest, PhaseScope, "call scope requires TenantID and RequestID")
	}

	binding, err := c.bindings.ResolveBinding(ctx, cl.scope, cl.target.BindingID)
	if err != nil {
		if e := contextError(ctx, PhaseBinding); e != nil {
			return e
		}
		return newError(CodeBindingNotFound, PhaseBinding, "binding could not be resolved for this tenant")
	}
	ad, ok := c.adapters[binding.API]
	if !ok {
		return newError(CodeInvalidRequest, PhaseCapability, "binding API is not supported: "+string(binding.API))
	}
	if !endpointAllowed(binding.Endpoint, c.loopback) {
		return newError(CodeInvalidRequest, PhaseBinding, "binding endpoint must be https (http only for loopback)")
	}
	model, failure := c.authorizeModel(binding, cl.target.ModelID)
	if failure != nil {
		return failure
	}
	meta.Resolved, meta.ProviderID, meta.API, meta.ModelID = true, binding.ProviderID, binding.API, model.ID
	asm.identify(binding.API, binding.ProviderID, model.ID)

	options := cl.full
	if cl.simple != nil {
		options = ad.simpleOptions(model, *cl.simple)
	} else if options != nil && options.api() != binding.API {
		return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
	}

	cred, err := c.credentials.ResolveCredential(ctx, cl.scope, binding)
	if err != nil {
		if e := contextError(ctx, PhaseCredential); e != nil {
			return e
		}
		return newError(CodeCredentialUnavailable, PhaseCredential, "credential could not be resolved for this binding")
	}

	return ad.stream(ctx, adapterCall{
		http:     c.http,
		endpoint: binding.Endpoint,
		apiKey:   cred.APIKey,
		model:    model,
		request:  cl.req,
		options:  options,
	}, asm)
}

// authorizeModel returns the model if it is both allowed by the binding and
// present in the catalog for the binding's provider and API.
func (c *Client) authorizeModel(b Binding, modelID string) (Model, *Error) {
	if !slices.Contains(b.AllowedModels, modelID) {
		return Model{}, newError(CodeTenantDenied, PhaseCapability, "model is not allowed by the binding")
	}
	m, ok := c.catalog.lookup(b.ProviderID, b.API, modelID)
	if !ok {
		return Model{}, newError(CodeInvalidRequest, PhaseCapability, "model is not in the catalog for this provider and API")
	}
	return m, nil
}

// endpointAllowed accepts https endpoints, and plain http to loopback hosts
// only when the Client was test-assembled with AllowLoopbackHTTP.
func endpointAllowed(endpoint string, allowLoopbackHTTP bool) bool {
	u, err := url.Parse(endpoint)
	if err != nil || u.Host == "" {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		if !allowLoopbackHTTP {
			return false
		}
		ip := net.ParseIP(u.Hostname())
		return ip != nil && ip.IsLoopback()
	}
	return false
}

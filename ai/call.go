package ai

import (
	"context"
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
	// hooks are the trusted host's callbacks, pinned when the call starts.
	hooks Hooks
}

func newCall(scope CallScope, target Target, req Request, full Options, simple *SimpleOptions, hooks Hooks) call {
	if simple != nil {
		s := simple.clone()
		simple = &s
	}
	if full != nil {
		full = full.clone()
	}
	return call{scope: scope, target: target, req: req.clone(), full: full, simple: simple, hooks: hooks}
}

// run executes a call to its terminal and returns the result. emit queues
// every event when the caller streams; it is nil for Complete.
func (c *Client) run(ctx context.Context, cl call, emit func(Event) *Error) (Result, error) {
	c.probe.CallStarted()
	defer c.probe.CallEnded()
	// The policy's CallTimeout bounds the whole call, setup included; the
	// host's own deadline ends it earlier if it comes first.
	ctx, cancel := context.WithTimeout(ctx, c.policy.CallTimeout)
	defer cancel()
	started := time.Now()
	asm := newAssembler(started.UnixMilli(), emit, c.policy.byteLimits().toolJSON)
	meta := CallMetadata{
		TenantID:  cl.scope.TenantID,
		RequestID: cl.scope.RequestID,
		ActorID:   cl.scope.ActorID,
		JobID:     cl.scope.JobID,
		BindingID: cl.target.BindingID,
	}
	c.observations.callStarted(meta)
	failure := c.execute(ctx, cl, &meta, asm)
	res, err := asm.finish(meta, failure)
	c.observations.callFinished(res, err, started)
	return res, err
}

// execute follows spec I3's order: scope → binding → capability/options →
// credential → snapshot consistency → admission → send. Every step before
// send fails without contacting the provider, and identity is only reported
// resolved once the snapshot is consistent. Admission is per attempt, so it
// happens inside the adapter's send of the initial request (initialRequest),
// once the request body is built.
func (c *Client) execute(ctx context.Context, cl call, meta *CallMetadata, asm *assembler) *Error {
	if cl.scope.TenantID == "" || cl.scope.RequestID == "" {
		return newError(CodeInvalidRequest, PhaseScope, "call scope requires TenantID and RequestID")
	}
	if problem := cl.req.validate(); problem != "" {
		return newError(CodeInvalidRequest, PhaseScope, problem)
	}
	limits := c.policy.byteLimits()
	if failure := limits.checkImages(cl.req); failure != nil {
		return failure
	}

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
	// Simple options are mapped for the model first; either way the
	// protocol options are then checked, which also holds samplingParams
	// merged from the model to the call's boundary.
	options := cl.full
	if cl.simple != nil {
		if problem := cl.simple.validate(); problem != "" {
			return newError(CodeInvalidRequest, PhaseCapability, problem)
		}
		options = ad.simpleOptions(model, cl.simple.resolve(model, cl.req))
	}
	if options != nil {
		if options.api() != binding.API {
			return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
		}
		if problem := options.validate(); problem != "" {
			return newError(CodeInvalidRequest, PhaseCapability, problem)
		}
	}

	// History source: native state is replayed only where its envelope
	// matches this tenant, the binding's account and the model; the rest is
	// downgraded, never refused.
	// The history is converted for the target as pi does (spec I6).
	origin := replayOrigin{tenant: cl.scope.TenantID, account: binding.AccountScopeID, model: model}
	history, downgrades := origin.prepareTranscript(cl.req, ad.historyRules(model))

	cred, failure := c.resolveCredential(ctx, cl.scope, binding)
	if failure != nil {
		return failure
	}
	if failure := checkSnapshot(cl.scope, binding, cred); failure != nil {
		return failure
	}
	meta.Resolved, meta.ProviderID, meta.API, meta.ModelID = true, binding.ProviderID, binding.API, model.ID
	meta.AccountScopeID = binding.AccountScopeID
	meta.BindingVersion, meta.CredentialVersion = binding.Version, cred.Version
	meta.CatalogVersion, meta.CatalogHash = c.catalog.Version, c.catalogHash
	meta.NativeStateDowngrades = downgrades
	asm.identify(origin.envelope())

	// From here on the call is pinned to this snapshot: later updates or
	// revocations affect only new logical calls, and every retry of the
	// initial request reuses it, each attempt under its own permit.
	initial := &initialRequest{policy: binding.Retry.pinned(), clock: c.clock, requestID: cl.scope.RequestID,
		observe: c.observations.attempts(*meta),
		admit: func(ctx context.Context, attemptID string) (func(), *Error) {
			return c.admission.admit(ctx, AdmissionRequest{TenantID: cl.scope.TenantID, AccountScopeID: binding.AccountScopeID, AttemptID: attemptID})
		}}
	// The permit of the attempt whose stream the adapter read is held until
	// the adapter returned, which closed that stream.
	defer initial.done()
	ac := adapterCall{
		http:     c.http,
		endpoint: binding.Endpoint,
		apiKey:   cred.APIKey,
		model:    model,
		history:  history,
		options:  options,
		cache:    cacheScope{tenant: cl.scope.TenantID, account: binding.AccountScopeID},
		// Copied so a resolver that reuses its slice cannot change the
		// pinned snapshot.
		hostedTools: slices.Clone(binding.AllowedHostedTools),
		hooks:       boundHooks{hooks: cl.hooks, scope: cl.scope, model: model},
		initial:     initial,
		limits:      limits,
	}
	// The trusted header transform sees the merged authentication and
	// request headers before the adapter takes over, and may not change
	// what the binding authorized.
	if ac.header, failure = ac.hooks.headers(ctx, ad.headers(ac)); failure != nil {
		return failure
	}
	failure = ad.stream(ctx, ac, asm)
	meta.Attempts = initial.attempts
	return failure
}

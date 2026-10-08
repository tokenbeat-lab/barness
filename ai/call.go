package ai

import (
	"context"
	"slices"
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
func (c *Client) run(ctx context.Context, cl call, emit func(EventEnvelope) *Error) (Result, error) {
	ctx, runtime := c.beginCall(ctx, cl.scope, cl.target, OperationChat)
	defer runtime.close()
	meta := &runtime.meta
	asm := newAssembler(runtime.started.UnixMilli(), meta.CallAttribution, emit, c.policy.byteLimits().toolJSON)
	failure := c.execute(ctx, cl, runtime, asm)
	res, err := asm.finish(*meta, failure)
	classified, _ := err.(*Error)
	runtime.finish(callOutcome{stop: res.Message.StopReason, usage: res.Message.Usage, failure: classified})
	res.Metadata = runtime.meta
	return res, err
}

// execute follows spec I3's order: scope → binding → capability/options →
// credential → snapshot consistency → admission → send. Every step before
// send fails without contacting the provider, and identity is only reported
// resolved once the snapshot is consistent. Admission is per attempt, so it
// happens inside the adapter's send of the initial request (initialRequest),
// once the request body is built.
func (c *Client) execute(ctx context.Context, cl call, runtime *callRuntime, asm *assembler) *Error {
	meta := &runtime.meta
	if failure := validateScope(cl.scope); failure != nil {
		return failure
	}
	if problem := cl.req.validate(); problem != "" {
		return newError(CodeInvalidRequest, PhaseScope, problem)
	}
	limits := c.policy.byteLimits()
	if failure := limits.checkImages(cl.req); failure != nil {
		return failure
	}

	index, failure := runtime.resolve(ctx, cl.target)
	if failure != nil {
		return failure
	}
	binding := runtime.binding
	ad := c.adapters[binding.API]
	model := c.catalog.Models[index].clone()
	// Simple options are mapped for the model first; either way the
	// protocol options are then checked, which also holds samplingParams
	// merged from the model to the call's boundary.
	options := cl.full
	if cl.simple != nil {
		if problem := cl.simple.validate(); problem != "" {
			return newError(CodeInvalidRequest, PhaseCapability, problem)
		}
		options = ad.simpleOptions(model, cl.simple.resolve(model, cl.req), normalizeTranscript(cl.req))
	}
	if failure := validateOptions(options, binding.API); failure != nil {
		return failure
	}

	// History source: native state is replayed only where its envelope
	// matches this tenant, the binding's account and the model; the rest is
	// downgraded, never refused.
	// The history is converted for the target as pi does (spec I6).
	origin := replayOrigin{tenant: cl.scope.TenantID, account: binding.AccountScopeID, model: model}
	history, downgrades := origin.prepareTranscript(cl.req, ad.historyRules(model))

	meta.NativeStateDowngrades = downgrades
	cred, failure := runtime.pin(ctx, model.ID, cl.hooks)
	if failure != nil {
		meta.NativeStateDowngrades = NativeStateDowngrades{}
		return failure
	}
	asm.identify(origin.envelope(), meta.CallAttribution)
	initial := runtime.initial
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
		hooks:       runtime.hooks,
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

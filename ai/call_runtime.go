package ai

import (
	"context"
	"slices"
	"time"
)

// callRuntime owns the operation-independent lifetime and pinned snapshot.
// Protocol input conversion and result construction remain on their own paths.
type callRuntime struct {
	client  *Client
	scope   CallScope
	meta    CallMetadata
	started time.Time
	cancel  context.CancelFunc
	closed  bool
	binding Binding
	hooks   boundHooks
	initial *initialRequest
}

type callOutcome struct {
	stop    StopReason
	usage   Usage
	failure *Error
}

func (c *Client) beginCall(ctx context.Context, scope CallScope, target Target, op Operation) (context.Context, *callRuntime) {
	scope.Operation = op
	c.probe.CallStarted()
	ctx, cancel := context.WithTimeout(ctx, c.policy.CallTimeout)
	r := &callRuntime{client: c, scope: scope, started: time.Now(), cancel: cancel,
		meta: CallMetadata{CallAttribution: CallAttribution{TenantID: scope.TenantID, RequestID: scope.RequestID, ActorID: scope.ActorID, JobID: scope.JobID, BindingID: target.BindingID, Operation: op}}}
	c.observations.callStarted(r.meta)
	return ctx, r
}

func (r *callRuntime) finish(out callOutcome) {
	if r.initial != nil {
		r.initial.done()
		r.meta.Attempts = slices.Clone(r.initial.attempts)
	}
	r.client.observations.callFinished(r.meta, out, r.started)
	r.close()
}

// close also runs during panic unwinding; callbacks remain the host's fault,
// but neither a deadline nor an admission permit may outlive the call.
func (r *callRuntime) close() {
	if r.closed {
		return
	}
	r.closed = true
	if r.initial != nil {
		r.initial.done()
	}
	r.cancel()
	r.client.probe.CallEnded()
}

func validateScope(scope CallScope) *Error {
	if scope.TenantID == "" || scope.RequestID == "" {
		return newError(CodeInvalidRequest, PhaseScope, "call scope requires TenantID and RequestID")
	}
	return nil
}

// resolve selects the binding and model index before any credential read.
// Each operation takes the index from its own typed catalog collection.
func (r *callRuntime) resolve(ctx context.Context, target Target) (int, *Error) {
	b, failure := r.client.resolveBinding(ctx, r.scope, target.BindingID)
	if failure != nil {
		return 0, failure
	}
	if b.Operation != r.meta.Operation {
		return 0, newError(CodeTenantDenied, PhaseCapability, "binding does not allow this operation")
	}
	supported := false
	switch r.meta.Operation {
	case OperationChat:
		_, supported = r.client.adapters[b.API]
	case OperationClassifier:
		supported = b.ProviderID == ProviderTypeSafe && b.API == APITypeSafeSystemOne
	}
	if !supported {
		return 0, newError(CodeInvalidRequest, PhaseCapability, "binding API is not supported: "+string(b.API))
	}
	if !slices.Contains(b.AllowedModels, target.ModelID) {
		return 0, newError(CodeTenantDenied, PhaseCapability, "model is not allowed by the binding")
	}
	i, ok := r.client.modelIndex[modelKey{b.Operation, b.ProviderID, b.API, target.ModelID}]
	if !ok {
		return 0, newError(CodeInvalidRequest, PhaseCapability, "model is not in the catalog for this provider and API")
	}
	r.binding = b
	return i, nil
}

func validateOptions(opts Options, api API) *Error {
	if opts == nil {
		return nil
	}
	if opts.api() != api {
		return newError(CodeInvalidRequest, PhaseCapability, "options do not match the binding API")
	}
	if problem := opts.validate(); problem != "" {
		return newError(CodeInvalidRequest, PhaseCapability, problem)
	}
	return nil
}

// pin completes the consistency check and creates the one initial-request
// owner, reused by chat and unary calls for retries, permits and observations.
func (r *callRuntime) pin(ctx context.Context, modelID string, hooks Hooks) (Credential, *Error) {
	c, b := r.client, r.binding
	cred, failure := c.resolveCredential(ctx, r.scope, b)
	if failure != nil {
		return Credential{}, failure
	}
	if failure := checkSnapshot(r.scope, b, cred); failure != nil {
		return Credential{}, failure
	}
	m := &r.meta
	m.Resolved, m.ProviderID, m.API, m.ModelID = true, b.ProviderID, b.API, modelID
	m.AccountScopeID = b.AccountScopeID
	m.BindingVersion, m.CredentialVersion = b.Version, cred.Version
	m.CatalogVersion, m.CatalogHash = c.catalog.Version, c.catalogHash
	r.hooks = boundHooks{hooks: hooks, scope: r.scope, identity: m.CallAttribution}
	r.initial = &initialRequest{policy: b.Retry.pinned(), clock: c.clock, requestID: r.scope.RequestID, observe: c.observations.attempts(*m),
		admit: func(ctx context.Context, attemptID string) (func(), *Error) {
			return c.admission.admit(ctx, AdmissionRequest{TenantID: r.scope.TenantID, AccountScopeID: b.AccountScopeID, AttemptID: attemptID})
		}}
	return cred, nil
}

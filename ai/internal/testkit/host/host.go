// Package host is the trusted-host double for barness-ai E2E tests. It plays
// the host's role of resolving tenant-owned bindings and versioned credentials,
// so tests exercise the library's resolver contract without a real secret
// store (spec Testing Decisions §1, §3).
//
// Besides the well-behaved path it can deny actors, fail its secret backend,
// misfile records (a buggy backend whose key disagrees with the record) and
// run a hook between the binding and credential reads so updates and
// revocations interleave deterministically with a call's resolution.
package host

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/tokenbeat-lab/barness/ai"
)

// ErrNotFound is returned for lookups the host has no record of.
var ErrNotFound = errors.New("host: not found")

type bindingKey struct{ tenant, binding string }
type credentialKey struct{ tenant, credentialID string }

// Reads counts one RequestID's resolver calls.
type Reads struct {
	Bindings    int `json:"bindings"`
	Credentials int `json:"credentials"`
}

// Host stores bindings by (TenantID, BindingID) and credentials by
// (OwnerTenantID, CredentialID); a binding's CredentialRef names the latter.
// It is safe for concurrent use.
type Host struct {
	mu          sync.Mutex
	bindings    map[bindingKey]ai.Binding
	credentials map[credentialKey]ai.Credential
	actors      map[bindingKey][]string
	gate        chan struct{}
	credFailure error
	beforeCred  func()
	reads       map[string]*Reads
}

// New returns an empty host.
func New() *Host {
	return &Host{
		bindings:    map[bindingKey]ai.Binding{},
		credentials: map[credentialKey]ai.Credential{},
		actors:      map[bindingKey][]string{},
		reads:       map[string]*Reads{},
	}
}

// PutBinding stores b under (b.TenantID, b.BindingID), replacing any earlier
// version.
func (h *Host) PutBinding(b ai.Binding) { h.PutBindingAt(b.TenantID, b.BindingID, b) }

// PutBindingAt stores b under an explicit key, which may disagree with b's own
// fields to model a misbehaving host backend.
func (h *Host) PutBindingAt(tenant, bindingID string, b ai.Binding) {
	h.mu.Lock()
	defer h.mu.Unlock()
	b.AllowedModels = slices.Clone(b.AllowedModels)
	h.bindings[bindingKey{tenant, bindingID}] = b
}

// Binding returns the stored binding at (tenant, bindingID).
func (h *Host) Binding(tenant, bindingID string) ai.Binding {
	h.mu.Lock()
	defer h.mu.Unlock()
	b := h.bindings[bindingKey{tenant, bindingID}]
	b.AllowedModels = slices.Clone(b.AllowedModels)
	return b
}

// PutCredential stores c under (c.OwnerTenantID, c.CredentialID), replacing
// any earlier version.
func (h *Host) PutCredential(c ai.Credential) { h.PutCredentialAt(c.OwnerTenantID, c.CredentialID, c) }

// PutCredentialAt stores c under an explicit key, which may disagree with c's
// own fields to model a misbehaving host backend.
func (h *Host) PutCredentialAt(tenant, credentialID string, c ai.Credential) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.credentials[credentialKey{tenant, credentialID}] = c
}

// DeleteCredential removes the credential at (tenant, credentialID).
func (h *Host) DeleteCredential(tenant, credentialID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.credentials, credentialKey{tenant, credentialID})
}

// RestrictActors limits which ActorIDs may use (tenant, bindingID); any other
// actor is refused with ai.ErrAccessDenied.
func (h *Host) RestrictActors(tenant, bindingID string, actors ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.actors[bindingKey{tenant, bindingID}] = actors
}

// FailCredentials makes every credential read fail with err, standing in for
// a secret backend outage. Nil restores normal reads.
func (h *Host) FailCredentials(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.credFailure = err
}

// BeforeCredential runs fn at the start of every credential read, i.e. after
// the call's binding was resolved. Tests use it to update or revoke
// configuration between the two reads. Nil removes the hook.
func (h *Host) BeforeCredential(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.beforeCred = fn
}

// ReadsFor reports how often each resolver was called for requestID.
func (h *Host) ReadsFor(requestID string) Reads {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r, ok := h.reads[requestID]; ok {
		return *r
	}
	return Reads{}
}

// HoldBindings makes ResolveBinding block until the returned release func is
// called (or the call's context ends), standing in for a slow host backend.
func (h *Host) HoldBindings() (release func()) {
	gate := make(chan struct{})
	h.mu.Lock()
	h.gate = gate
	h.mu.Unlock()
	var once sync.Once
	return func() { once.Do(func() { close(gate) }) }
}

func (h *Host) countLocked(requestID string) *Reads {
	r, ok := h.reads[requestID]
	if !ok {
		r = &Reads{}
		h.reads[requestID] = r
	}
	return r
}

// ResolveBinding implements ai.BindingResolver.
func (h *Host) ResolveBinding(ctx context.Context, scope ai.CallScope, bindingID string) (ai.Binding, error) {
	h.mu.Lock()
	h.countLocked(scope.RequestID).Bindings++
	gate := h.gate
	h.mu.Unlock()
	if gate != nil {
		select {
		case <-gate:
		case <-ctx.Done():
			return ai.Binding{}, ctx.Err()
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	key := bindingKey{scope.TenantID, bindingID}
	b, ok := h.bindings[key]
	if !ok {
		return ai.Binding{}, ErrNotFound
	}
	if actors, restricted := h.actors[key]; restricted && !slices.Contains(actors, scope.ActorID) {
		return ai.Binding{}, ai.ErrAccessDenied
	}
	b.AllowedModels = slices.Clone(b.AllowedModels)
	return b, nil
}

// ResolveCredential implements ai.CredentialResolver. Like a careful backend,
// it refuses with ai.ErrSnapshotConflict when the binding it is asked about is
// no longer the current version, instead of pairing a stale binding with
// today's credential.
func (h *Host) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	h.mu.Lock()
	h.countLocked(scope.RequestID).Credentials++
	hook := h.beforeCred
	h.mu.Unlock()
	if hook != nil {
		hook()
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.credFailure != nil {
		return ai.Credential{}, h.credFailure
	}
	if current, ok := h.bindings[bindingKey{scope.TenantID, b.BindingID}]; !ok || current.Version != b.Version {
		return ai.Credential{}, ai.ErrSnapshotConflict
	}
	c, ok := h.credentials[credentialKey{scope.TenantID, b.CredentialRef}]
	if !ok {
		return ai.Credential{}, ErrNotFound
	}
	return c, nil
}

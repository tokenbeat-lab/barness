// Package host is the trusted-host double for barness-ai E2E tests. It plays
// the host's role of resolving tenant-owned bindings and versioned credentials,
// so tests exercise the library's resolver contract without a real secret
// store (spec Testing Decisions §1).
package host

import (
	"context"
	"errors"
	"sync"

	"github.com/tokenbeat-lab/barness/ai"
)

// ErrNotFound is returned for lookups the host has no record of.
var ErrNotFound = errors.New("host: not found")

type bindingKey struct{ tenant, binding string }
type credentialKey struct{ tenant, credentialID string }

// Host stores bindings by (TenantID, BindingID) and credentials by
// (OwnerTenantID, CredentialID); a binding's CredentialRef names the latter.
// It is safe for concurrent use.
type Host struct {
	mu          sync.Mutex
	bindings    map[bindingKey]ai.Binding
	credentials map[credentialKey]ai.Credential
	gate        chan struct{}
}

// New returns an empty host.
func New() *Host {
	return &Host{bindings: map[bindingKey]ai.Binding{}, credentials: map[credentialKey]ai.Credential{}}
}

// PutBinding stores b under (b.TenantID, b.BindingID).
func (h *Host) PutBinding(b ai.Binding) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.bindings[bindingKey{b.TenantID, b.BindingID}] = b
}

// PutCredential stores c under (c.OwnerTenantID, c.CredentialID).
func (h *Host) PutCredential(c ai.Credential) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.credentials[credentialKey{c.OwnerTenantID, c.CredentialID}] = c
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

// ResolveBinding implements ai.BindingResolver.
func (h *Host) ResolveBinding(ctx context.Context, scope ai.CallScope, bindingID string) (ai.Binding, error) {
	h.mu.Lock()
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
	b, ok := h.bindings[bindingKey{scope.TenantID, bindingID}]
	if !ok {
		return ai.Binding{}, ErrNotFound
	}
	return b, nil
}

// ResolveCredential implements ai.CredentialResolver.
func (h *Host) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	c, ok := h.credentials[credentialKey{scope.TenantID, b.CredentialRef}]
	if !ok {
		return ai.Credential{}, ErrNotFound
	}
	return c, nil
}

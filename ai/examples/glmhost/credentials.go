package glmhost

import (
	"context"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// CredentialRecord contains only independently stored ownership metadata.
// Record versions come from the trusted backend, never the caller's binding.
// This record stays separate from the other brand's example because purpose
// qualifications and model entitlement evolve independently.
type CredentialRecord struct {
	ProviderID                              ai.ProviderID
	TenantID, AccountScopeID, CredentialRef string
	Purpose                                 Purpose
	Version, BindingVersion                 string
	Active                                  bool
	AllowedAPIs                             []ai.API
}

// Credentials is an immutable resolved Host configuration snapshot. A Host
// with live updates resolves a fresh snapshot atomically in its own backend;
// it must not mutate this value concurrently with calls. ReadSecret is invoked
// only after all ownership, purpose, API and version checks succeed. It receives
// the tenant-owned record's reference; it must not select alternate keys.
type Credentials struct {
	Authorization Authorization
	Record        CredentialRecord
	ReadSecret    func(context.Context, string) (ai.Secret, error)
}

func (c Credentials) ResolveCredential(ctx context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	a, r := c.Authorization, c.Record
	routes, err := Bindings(a)
	if err != nil || scope.TenantID != a.TenantID || r.TenantID != a.TenantID || r.ProviderID != a.ProviderID || r.AccountScopeID != a.AccountScopeID || r.CredentialRef != a.CredentialRef || r.Purpose != a.Purpose || !r.Active || !slices.Contains(r.AllowedAPIs, b.API) || strings.TrimSpace(r.Version) == "" || c.ReadSecret == nil {
		return ai.Credential{}, ai.ErrAccessDenied
	}
	authorized := false
	for _, expected := range routes {
		if b.BindingID == expected.BindingID && b.TenantID == expected.TenantID && b.ProviderID == expected.ProviderID && b.API == expected.API && b.Operation == expected.Operation && b.AccountScopeID == expected.AccountScopeID && b.CredentialRef == expected.CredentialRef {
			authorized = true
			break
		}
	}
	if !authorized {
		return ai.Credential{}, ai.ErrAccessDenied
	}
	if b.Version != a.Version || r.BindingVersion != a.Version {
		return ai.Credential{}, ai.ErrSnapshotConflict
	}
	if err := ctx.Err(); err != nil {
		return ai.Credential{}, err
	}
	secret, err := c.ReadSecret(ctx, r.CredentialRef)
	if err != nil {
		return ai.Credential{}, err
	}
	return ai.Credential{OwnerTenantID: r.TenantID, CredentialID: r.CredentialRef, Version: r.Version, BindingVersion: r.BindingVersion, AccountScopeID: r.AccountScopeID, Active: r.Active, APIKey: secret}, nil
}

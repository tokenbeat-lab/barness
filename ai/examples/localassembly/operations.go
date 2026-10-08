package localassembly

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

// OperationsConfig explicitly assembles independent operation bindings. Keys
// are indexed by CredentialRef; only bindings of the same Provider/account
// may share a reference. The host reads each named source once, never using
// conventional environment fallbacks. Policy is required: use MixedPolicy
// only after checking its documented design load against the program's load.
type OperationsConfig struct {
	TenantID          string
	Bindings          []ai.Binding
	Keys              map[string]KeySource
	Policy            *ai.ResourcePolicy
	Catalog           *ai.Catalog
	Transport         http.RoundTripper
	AllowLoopbackHTTP bool
}

// Operations holds one read-only Client for all local model operations.
// Call Client.Complete, GenerateImages or Classify with a fresh Scope and an
// explicit ai.Target. Pass the user's cancellation context through and handle
// typed ai.Error codes; classification confidence thresholds, human review
// and business routing are decisions of this program, not of barness-ai.
type Operations struct {
	Client *ai.Client
	tenant string
}

func (a *Operations) Scope() ai.CallScope {
	return ai.CallScope{TenantID: a.tenant, RequestID: requestid.New()}
}

// OpenOperations sends nothing. This immutable local resolver is deliberately
// separate from the mutable cloud secret store: live revocation/rotation is a
// cloud host's responsibility; a local program reassembles after key changes.
func OpenOperations(cfg OperationsConfig) (*Operations, error) {
	if cfg.TenantID == "" || len(cfg.Bindings) == 0 || cfg.Policy == nil {
		return nil, errors.New("localassembly: tenant, bindings and explicit policy are required")
	}
	s := &operationsResolver{tenant: cfg.TenantID, bindings: map[string]ai.Binding{}, credentials: map[string]ai.Credential{}}
	providers := map[string]ai.ProviderID{}
	for _, original := range cfg.Bindings {
		b := original
		if b.BindingID == "" || b.CredentialRef == "" || b.AccountScopeID == "" || b.ProviderID == "" {
			return nil, errors.New("localassembly: binding, credential, Provider and account are required")
		}
		if _, exists := s.bindings[b.BindingID]; exists {
			return nil, errors.New("localassembly: duplicate binding")
		}
		b.TenantID, b.Version, b.AuthKind = cfg.TenantID, "local", ai.AuthAPIKey
		// Enabled remains the host's explicit choice, including for rollback.
		b.AllowedModels = slices.Clone(b.AllowedModels)
		b.AllowedHostedTools = slices.Clone(b.AllowedHostedTools)
		if b.Retry.MaxRetryDelay != nil {
			d := *b.Retry.MaxRetryDelay
			b.Retry.MaxRetryDelay = &d
		}
		if c, exists := s.credentials[b.CredentialRef]; exists {
			if c.AccountScopeID != b.AccountScopeID || providers[b.CredentialRef] != b.ProviderID {
				return nil, errors.New("localassembly: shared credential requires the same Provider and account")
			}
		} else {
			source, ok := cfg.Keys[b.CredentialRef]
			if !ok {
				return nil, ErrKeySource
			}
			key, err := readKey(source)
			if err != nil {
				return nil, err
			}
			s.credentials[b.CredentialRef] = ai.Credential{OwnerTenantID: cfg.TenantID, CredentialID: b.CredentialRef,
				Version: "local", BindingVersion: "local", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(key)}
			providers[b.CredentialRef] = b.ProviderID
		}
		s.bindings[b.BindingID] = b
	}
	client, err := ai.NewClient(ai.Config{Policy: cfg.Policy, Catalog: cfg.Catalog, Bindings: s, Credentials: s,
		Transport: cfg.Transport, AllowLoopbackHTTP: cfg.AllowLoopbackHTTP})
	if err != nil {
		return nil, err
	}
	return &Operations{Client: client, tenant: cfg.TenantID}, nil
}

type operationsResolver struct {
	tenant      string
	bindings    map[string]ai.Binding
	credentials map[string]ai.Credential
}

func (s *operationsResolver) ResolveBinding(_ context.Context, scope ai.CallScope, id string) (ai.Binding, error) {
	b, ok := s.bindings[id]
	if scope.TenantID != s.tenant || !ok {
		return ai.Binding{}, errNotFound
	}
	b.AllowedModels = slices.Clone(b.AllowedModels)
	b.AllowedHostedTools = slices.Clone(b.AllowedHostedTools)
	if b.Retry.MaxRetryDelay != nil {
		d := *b.Retry.MaxRetryDelay
		b.Retry.MaxRetryDelay = &d
	}
	return b, nil
}

func (s *operationsResolver) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	c, ok := s.credentials[b.CredentialRef]
	if scope.TenantID != s.tenant || !ok {
		return ai.Credential{}, errNotFound
	}
	return c, nil
}

// Package localassembly is barness-ai's minimal local assembly (spec I3, User
// Story 2): a local program uses the same Client as a cloud host and states
// explicitly what a cloud host would resolve per tenant — one tenant, one
// binding and one API key.
//
// The division of work it shows:
//
//   - The program, not barness-ai, reads the key, from the environment
//     variable or secret file it names (KeySource). barness-ai and the
//     provider SDKs never read OPENAI_API_KEY, ANTHROPIC_API_KEY, endpoint
//     variables or credential files themselves, and nothing falls back to
//     them when the named source is missing: assembly fails instead.
//   - The program declares the binding the key belongs to: provider,
//     protocol, endpoint, account and the models it may call.
//   - The program supplies a finite resource policy (LocalPolicy).
//   - Each call gets a trusted scope with a fresh RequestID (Scope).
package localassembly

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
)

// KeySource names where the program reads its API key. Exactly one field is
// set.
type KeySource struct {
	// EnvVar is an environment variable the program chose, such as
	// BARNESS_OPENAI_KEY. It is read once, by Open.
	EnvVar string
	// File is a secret file holding only the key. It is read once, by Open.
	// From either source one trailing line ending is ignored.
	File string
}

// Config is what a local program decides.
type Config struct {
	// TenantID is the local tenant; the program owns every call. It names
	// the attribution of results and observations.
	TenantID string
	// Binding is the one service configuration of the tenant. Its TenantID,
	// Version, Enabled, AuthKind and CredentialRef are filled in by Open.
	Binding ai.Binding
	Key     KeySource
	// Policy is the resource policy; nil uses LocalPolicy.
	Policy *ai.ResourcePolicy
	// Transport and AllowLoopbackHTTP are the Client's network assembly
	// (ai.Config). A program leaves both unset; tests point the Client at a
	// local controlled Provider with them.
	Transport         http.RoundTripper
	AllowLoopbackHTTP bool
}

// Assembly is the assembled local Client with its tenant and binding.
type Assembly struct {
	Client  *ai.Client
	tenant  string
	binding string
}

// ErrKeySource reports a key source that is not set up; Open never falls
// back to another source.
var ErrKeySource = errors.New("localassembly: key source")

// Open reads the key from cfg.Key and assembles the Client. It sends nothing.
func Open(cfg Config) (*Assembly, error) {
	if cfg.TenantID == "" || cfg.Binding.BindingID == "" {
		return nil, errors.New("localassembly: TenantID and Binding.BindingID are required")
	}
	key, err := readKey(cfg.Key)
	if err != nil {
		return nil, err
	}
	policy := cfg.Policy
	if policy == nil {
		policy = LocalPolicy()
	}
	b := cfg.Binding
	b.TenantID, b.Version, b.Enabled, b.AuthKind = cfg.TenantID, "local", true, ai.AuthAPIKey
	b.CredentialRef = "local-key"
	b.AllowedModels = slices.Clone(b.AllowedModels)
	store := &localResolver{binding: b, credential: ai.Credential{
		OwnerTenantID: cfg.TenantID, CredentialID: b.CredentialRef, Version: "local", BindingVersion: b.Version,
		AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(key),
	}}
	client, err := ai.NewClient(ai.Config{
		Policy: policy, Bindings: store, Credentials: store,
		Transport: cfg.Transport, AllowLoopbackHTTP: cfg.AllowLoopbackHTTP,
	})
	if err != nil {
		return nil, err
	}
	return &Assembly{Client: client, tenant: cfg.TenantID, binding: b.BindingID}, nil
}

// Scope is the trusted scope of one new logical call.
func (a *Assembly) Scope() ai.CallScope {
	return ai.CallScope{TenantID: a.tenant, RequestID: requestid.New()}
}

// Target names model on the assembly's binding.
func (a *Assembly) Target(model string) ai.Target {
	return ai.Target{BindingID: a.binding, ModelID: model}
}

// readKey reads the one source the program named. An unset variable, an
// empty value or an unreadable file is an error, never a reason to try
// another source.
func readKey(src KeySource) (string, error) {
	switch {
	case (src.EnvVar == "") == (src.File == ""):
		return "", fmt.Errorf("%w: set exactly one of EnvVar and File", ErrKeySource)
	case src.EnvVar != "":
		value, ok := os.LookupEnv(src.EnvVar)
		key := trimLineEnding(value)
		if !ok || strings.TrimSpace(key) == "" {
			return "", fmt.Errorf("%w: environment variable %s is not set", ErrKeySource, src.EnvVar)
		}
		return key, nil
	}
	data, err := os.ReadFile(src.File)
	if err != nil {
		// The path is the program's own configuration; the error never holds
		// the key.
		return "", fmt.Errorf("%w: %v", ErrKeySource, err)
	}
	key := trimLineEnding(string(data))
	if strings.TrimSpace(key) == "" {
		return "", fmt.Errorf("%w: secret file %s is empty", ErrKeySource, src.File)
	}
	return key, nil
}

// trimLineEnding drops one trailing LF or CRLF, as `echo` and editors add.
func trimLineEnding(s string) string {
	return strings.TrimSuffix(strings.TrimSuffix(s, "\n"), "\r")
}

// localResolver resolves the local tenant's one binding and its key. Anything else
// is not found: a scope for another tenant or another binding name gets
// nothing. It does not cache across tenants because there is only one.
type localResolver struct {
	binding    ai.Binding
	credential ai.Credential
}

var errNotFound = errors.New("localassembly: not found")

func (s *localResolver) ResolveBinding(_ context.Context, scope ai.CallScope, bindingID string) (ai.Binding, error) {
	if scope.TenantID != s.binding.TenantID || bindingID != s.binding.BindingID {
		return ai.Binding{}, errNotFound
	}
	b := s.binding
	b.AllowedModels = slices.Clone(b.AllowedModels)
	return b, nil
}

func (s *localResolver) ResolveCredential(_ context.Context, scope ai.CallScope, b ai.Binding) (ai.Credential, error) {
	if scope.TenantID != s.credential.OwnerTenantID || b.CredentialRef != s.credential.CredentialID {
		return ai.Credential{}, errNotFound
	}
	return s.credential, nil
}

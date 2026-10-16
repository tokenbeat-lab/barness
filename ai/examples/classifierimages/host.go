package main

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

// Live config is supplied by a trusted Host, never decoded from end-user input.
// It intentionally contains no credential; -key-env names its sole key source.
type hostConfig struct {
	Binding ai.Binding         `json:"binding"`
	Target  ai.Target          `json:"target"`
	Catalog ai.Catalog         `json:"catalog"`
	Policy  *ai.ResourcePolicy `json:"policy"`
}
type resolver struct {
	mu              sync.Mutex
	binding         ai.Binding
	key             string
	credentialReads int
}

func (r *resolver) ResolveBinding(_ context.Context, s ai.CallScope, id string) (ai.Binding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.TenantID != r.binding.TenantID || id != r.binding.BindingID {
		return ai.Binding{}, errors.New("Host tenant/binding denied")
	}
	b := r.binding
	b.AllowedModels = slices.Clone(b.AllowedModels)
	return b, nil
}
func (r *resolver) ResolveCredential(_ context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.TenantID != r.binding.TenantID || b.BindingID != r.binding.BindingID || b.CredentialRef != r.binding.CredentialRef {
		return ai.Credential{}, errors.New("Host credential denied")
	}
	r.credentialReads++
	return ai.Credential{OwnerTenantID: b.TenantID, CredentialID: b.CredentialRef, Version: "example-key-v1", BindingVersion: b.Version, AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(r.key)}, nil
}
func (r *resolver) revoke(imagesOnly bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if imagesOnly {
		r.binding.AllowClassifierImages = false
	} else {
		r.binding.Enabled = false
	}
	r.binding.Version = "revoked-v2"
}
func (r *resolver) reads() int { r.mu.Lock(); defer r.mu.Unlock(); return r.credentialReads }

type observer struct {
	mu       sync.Mutex
	records  []ai.Observation
	finished chan struct{}
}

func (o *observer) Observe(v ai.Observation) error {
	o.mu.Lock()
	o.records = append(o.records, v)
	o.mu.Unlock()
	if v.Kind == ai.ObservationCallFinished {
		close(o.finished)
	}
	return nil
}
func (o *observer) await(ctx context.Context) ([]ai.Observation, error) {
	select {
	case <-o.finished:
	case <-ctx.Done():
		return nil, errors.New("observer evidence incomplete")
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	return slices.Clone(o.records), nil
}
func controlledConfig(route, endpoint string) hostConfig {
	catalog := ai.BuiltinCatalog()
	p := localassembly.MixedPolicy()
	// One serial synthetic call; these bounds describe this example, not defaults.
	p.Image = nil
	p.MaxConcurrentPerTenant = 1
	p.MaxConcurrentProcess = 1
	p.MaxRequestBytes = 1 << 20
	p.MaxOutputBytes = 64 << 10
	p.MaxToolJSONBytes = 32 << 10
	p.MaxFrameBytes = 32 << 10
	p.MaxImageBytes = 64 << 10
	p.MaxQueuedObservations = 16
	p.CallTimeout = 30 * time.Second
	p.ConnectTimeout = 5 * time.Second
	p.ResponseHeaderTimeout = 10 * time.Second
	p.ReadIdleTimeout = 10 * time.Second
	p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 4, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 2, MaxTotalImageBytes: 128 << 10, References: &ai.ClassifierImageReferencePolicy{MaxReferences: 8, MaxExpandedImageBytes: 512 << 10}}}
	b := ai.Binding{TenantID: "synthetic-host", BindingID: route, Version: "example-v1", Enabled: true, Operation: ai.OperationClassifier, ProviderID: ai.ProviderTypeSafe, API: ai.APITypeSafeSystemOne, Endpoint: endpoint, AccountScopeID: "synthetic-" + route, CredentialRef: "example-" + route, AuthKind: ai.AuthAPIKey, AllowedModels: []string{"jev-1.14.0"}, AllowClassifierImages: true}
	if route == "decisions" {
		b.ProviderID = ai.ProviderOpenAI
		b.API = ai.APIOpenAIDecisions
		b.AllowedModels = []string{"gpt-6-luna"}
	}
	return hostConfig{b, ai.Target{BindingID: b.BindingID, ModelID: b.AllowedModels[0]}, catalog, p}
}
func newClient(cfg hostConfig, r *resolver, o *observer, transport http.RoundTripper, controlled bool, gauges *probe.Probe) (*ai.Client, error) {
	return ai.NewClient(ai.Config{Probe: gauges, Catalog: &cfg.Catalog, Policy: cfg.Policy, Bindings: r, Credentials: r, Observer: o, Transport: transport, AllowLoopbackHTTP: controlled})
}

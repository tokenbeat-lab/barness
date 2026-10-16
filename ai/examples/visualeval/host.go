package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
)

type host struct {
	binding ai.Binding
	key     string
	reads   int
}

func (h *host) ResolveBinding(_ context.Context, s ai.CallScope, id string) (ai.Binding, error) {
	if s.TenantID != h.binding.TenantID || id != h.binding.BindingID {
		return ai.Binding{}, errors.New("Host binding denied")
	}
	return h.binding, nil
}
func (h *host) ResolveCredential(_ context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
	if s.TenantID != h.binding.TenantID || b.CredentialRef != h.binding.CredentialRef {
		return ai.Credential{}, errors.New("Host credential denied")
	}
	h.reads++
	return ai.Credential{OwnerTenantID: b.TenantID, CredentialID: b.CredentialRef, Version: "visual-evaluation-v1", BindingVersion: b.Version, AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(h.key)}, nil
}
func (p *plan) binding(endpoint, alias string) ai.Binding {
	return ai.Binding{TenantID: "visual-evaluation-host", BindingID: p.route, Version: "visual-evaluation-v1", Enabled: true, Operation: ai.OperationClassifier, ProviderID: p.model.Provider, API: p.model.API, Endpoint: endpoint, AccountScopeID: alias, CredentialRef: "isolated-visual-evaluation", AuthKind: ai.AuthAPIKey, AllowedModels: []string{p.model.ID}, AllowClassifierImages: true}
}
func policy(call time.Duration) *ai.ResourcePolicy {
	// Design load: one serial call, up to three questions/references, one small
	// 320x320 image. Cumulative experiment budgets are checked before each call.
	p := localassembly.MixedPolicy()
	p.Image = nil
	p.MaxConcurrentProcess = 1
	p.MaxConcurrentPerTenant = 1
	p.MaxQueuedObservations = 16
	p.MaxRequestBytes = 1 << 20
	p.MaxOutputBytes = 64 << 10
	p.MaxImageBytes = 64 << 10
	p.MaxFrameBytes = 32 << 10
	p.MaxToolJSONBytes = 32 << 10
	p.AdmissionWait = min(call, p.AdmissionWait)
	p.CallTimeout = call
	p.ConnectTimeout = min(call, 5*time.Second)
	p.ResponseHeaderTimeout = call
	p.ReadIdleTimeout = call
	p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 3, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 1, MaxTotalImageBytes: 64 << 10, References: &ai.ClassifierImageReferencePolicy{MaxReferences: 3, MaxExpandedImageBytes: 192 << 10}}}
	return p
}

type observations struct {
	mu      sync.Mutex
	records []ai.Observation
	done    chan struct{}
}

func (o *observations) Observe(v ai.Observation) error {
	o.mu.Lock()
	o.records = append(o.records, v)
	o.mu.Unlock()
	if v.Kind == ai.ObservationCallFinished {
		close(o.done)
	}
	return nil
}
func (o *observations) wait(ctx context.Context) []ai.Observation {
	select {
	case <-o.done:
	case <-ctx.Done():
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]ai.Observation{}, o.records...)
}

type wire struct {
	Method          string            `json:"method"`
	Path            string            `json:"path"`
	Headers         map[string]string `json:"headers"`
	Request         []byte            `json:"request"`
	Status          int               `json:"status"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	Response        []byte            `json:"response"`
}

// Capture is local to this example's per-sample evidence schema. It records
// only consumed bytes and whitelisted headers; redirects and proxies are off.
type capture struct {
	next http.RoundTripper
	mu   sync.Mutex
	wire wire
}

func (c *capture) RoundTrip(r *http.Request) (*http.Response, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, (1<<20)+1))
	r.Body.Close()
	if err != nil || len(raw) > 1<<20 {
		return nil, errors.New("bounded request capture failed")
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	c.mu.Lock()
	c.wire = wire{Method: r.Method, Path: r.URL.Path, Request: raw, Headers: map[string]string{"Content-Type": r.Header.Get("Content-Type"), "Typesafe-Danger-Zone-Images": r.Header.Get("Typesafe-Danger-Zone-Images")}, ResponseHeaders: map[string]string{}}
	c.mu.Unlock()
	res, err := c.next.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.wire.Status = res.StatusCode
	for _, k := range []string{"Content-Type", "X-Request-ID", "X-Typesafe-Request-ID"} {
		if v := res.Header.Get(k); v != "" {
			c.wire.ResponseHeaders[k] = v
		}
	}
	c.mu.Unlock()
	res.Body = &capturedBody{res.Body, c}
	return res, nil
}

type capturedBody struct {
	io.ReadCloser
	c *capture
}

func (b *capturedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.c.mu.Lock()
	keep := min(n, ((64<<10)+1)-len(b.c.wire.Response))
	b.c.wire.Response = append(b.c.wire.Response, p[:keep]...)
	b.c.mu.Unlock()
	return n, err
}
func (c *capture) snapshot() wire { c.mu.Lock(); defer c.mu.Unlock(); return c.wire }

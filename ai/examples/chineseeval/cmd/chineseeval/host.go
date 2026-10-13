package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

const tenant = "chinese-evaluation"
const binding = "typesafe-chinese-evaluation"

type resolver struct {
	cancelBeforeBinding context.CancelFunc
	binding             ai.Binding
	credential          ai.Credential
}

func (r *resolver) ResolveBinding(_ context.Context, s ai.CallScope, id string) (ai.Binding, error) {
	if s.TenantID != tenant || id != binding {
		return ai.Binding{}, errors.New("evaluation binding denied")
	}
	if r.cancelBeforeBinding != nil {
		r.cancelBeforeBinding()
	}
	return r.binding, nil
}
func (r *resolver) ResolveCredential(_ context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
	if s.TenantID != tenant || b.CredentialRef != r.credential.CredentialID || b.BindingID != binding {
		return ai.Credential{}, errors.New("evaluation credential denied")
	}
	return r.credential, nil
}

func newClient(p *chineseeval.Plan, key string, o *observer, injected http.RoundTripper, capture *captureTransport, proxy *url.URL) (*ai.Client, *http.Transport, error) {
	c := p.Config()
	route := p.Route()
	r := &resolver{
		binding:    ai.Binding{TenantID: tenant, BindingID: binding, Version: "zh-eval-v1", Enabled: true, Operation: ai.OperationClassifier, ProviderID: route.Provider, API: route.API, Endpoint: route.Endpoint(), AuthKind: ai.AuthAPIKey, AccountScopeID: "evaluation-" + string(route.Provider), CredentialRef: "evaluation-key", AllowedModels: []string{c.Model}},
		credential: ai.Credential{OwnerTenantID: tenant, CredentialID: "evaluation-key", Version: "evaluation-v1", BindingVersion: "zh-eval-v1", AccountScopeID: "evaluation-" + string(route.Provider), Active: true, APIKey: ai.NewSecret(key)},
	}
	if f, ok := injected.(*fixtureTransport); ok && f.profile == "pre-attempt-cancel" {
		r.cancelBeforeBinding = f.cancel
	}
	// Zero retries at the binding; one serial request per sample. Transport has
	// no environment credential/proxy source and no alternate provider route.
	transport := &http.Transport{TLSHandshakeTimeout: 5 * time.Second, MaxIdleConns: 1, MaxIdleConnsPerHost: 1, IdleConnTimeout: 30 * time.Second}
	if proxy != nil {
		transport.Proxy = http.ProxyURL(proxy)
	}
	policy := &ai.ResourcePolicy{
		Classifier:      &ai.ClassifierPolicy{MaxQuestions: 1, MaxStateBytes: 4096, MaxQuestionBytes: 4096},
		MaxRequestBytes: 16384, MaxImageBytes: 1024, MaxFrameBytes: 16384, MaxToolJSONBytes: 1024, MaxErrorBodyBytes: 4096, MaxOutputBytes: 16384, MaxQueuedEvents: 8, MaxQueuedEventBytes: 16384,
		MaxQueuedObservations: 128, MaxConcurrentPerTenant: 1, MaxConcurrentProcess: 1, MaxAdmissionWaiters: 1, AdmissionWait: time.Second,
		CallTimeout: time.Duration(c.CallTimeoutSeconds) * time.Second, ConnectTimeout: 5 * time.Second, ResponseHeaderTimeout: 10 * time.Second, ReadIdleTimeout: 10 * time.Second,
	}
	var rt http.RoundTripper = transport
	if injected != nil {
		rt = injected
	}
	if capture != nil {
		capture.requestIDHeader = "x-request-id"
		if route.API == ai.APITypeSafeSystemOne {
			capture.requestIDHeader = "x-typesafe-request-id"
		}
		capture.inner = rt
		rt = capture
	}
	catalog := p.Catalog()
	client, err := ai.NewClient(ai.Config{Bindings: r, Credentials: r, Catalog: &catalog, Policy: policy, Transport: rt, Observer: o})
	return client, transport, err
}

// A proxy is explicit trusted Host transport configuration; the official
// endpoint and ordinary TLS verification still apply. Keep this independent
// from live-smoke configuration: that runner has a separate evidence contract.
func evaluationProxy(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid proxy")
	}
	ip := net.ParseIP(u.Hostname())
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 || ip == nil || !ip.IsLoopback() || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("explicit loopback HTTP proxy required")
	}
	return u, nil
}

type observer struct {
	mu       sync.Mutex
	records  []ai.Observation
	terminal int
	changed  chan struct{}
}

func newObserver() *observer { return &observer{changed: make(chan struct{}, 1)} }
func (o *observer) Observe(r ai.Observation) error {
	o.mu.Lock()
	o.records = append(o.records, r)
	if r.Kind == ai.ObservationCallFinished {
		o.terminal++
	}
	o.mu.Unlock()
	select {
	case o.changed <- struct{}{}:
	default:
	}
	return nil
}
func (o *observer) finished(calls int) ([]ai.Observation, error) {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		o.mu.Lock()
		done := o.terminal == calls
		records := append([]ai.Observation(nil), o.records...)
		o.mu.Unlock()
		if done {
			return records, nil
		}
		select {
		case <-o.changed:
		case <-timer.C:
			return nil, errors.New("observer terminal evidence missing")
		}
	}
}

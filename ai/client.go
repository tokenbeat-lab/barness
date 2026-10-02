package ai

import (
	"errors"
	"net/http"

	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

// ErrInvalidConfig matches every Client construction error.
var ErrInvalidConfig = errors.New("ai: invalid client config")

// ConfigError reports which Config field made construction fail.
type ConfigError struct {
	Field   string
	Problem string
}

func (e *ConfigError) Error() string { return "ai: invalid config " + e.Field + ": " + e.Problem }

// Is makes errors.Is(err, ErrInvalidConfig) hold.
func (e *ConfigError) Is(target error) bool { return target == ErrInvalidConfig }

// Config assembles a Client. It is copied at construction; later changes to
// the value the caller passed have no effect.
type Config struct {
	// Policy is required (D1, ADR-0002).
	Policy *ResourcePolicy
	// Bindings and Credentials are the trusted host's resolvers; both required.
	Bindings    BindingResolver
	Credentials CredentialResolver
	// Catalog defaults to BuiltinCatalog when nil. Its prices must be
	// finite and non-negative.
	Catalog *Catalog
	// Transport is the only way to customize networking and is fixed for the
	// Client's lifetime; there is no per-call override. Nil selects a transport
	// that ignores proxy environment variables. Tests inject a loopback-only one.
	// The policy's ConnectTimeout ends when the transport reports a
	// connection through net/http/httptrace (GotConn), as *http.Transport
	// does; with a transport that never reports one, it bounds the wait for
	// the response headers instead.
	Transport http.RoundTripper
	// Admission is the trusted host's admission, consulted for every attempt
	// after the built-in per-tenant and process limits granted one; nil uses
	// the built-in limits alone. It is where per-account or distributed
	// admission plugs in.
	Admission Admission
	// Observer receives the Client's call and attempt records asynchronously
	// (see Observer); nil records nothing. It requires a positive
	// ResourcePolicy.MaxQueuedObservations.
	Observer Observer
	// AllowLoopbackHTTP is for test assembly only: it lets bindings target a
	// plain-http loopback endpoint such as a local controlled Provider. Without
	// it every binding endpoint must be https (spec I3).
	AllowLoopbackHTTP bool
	// Probe is for barness-ai's own acceptance tests, which observe internal
	// resource gauges through it. Its type is internal, so hosts cannot
	// construct one and leave it nil.
	Probe *probe.Probe
	// Clock is for barness-ai's own acceptance tests, which replace the time
	// and jitter source of retry backoff with a fake one. Its type is
	// internal, so hosts cannot construct one and leave it nil, the system
	// clock.
	Clock *clock.Clock
}

// Client is the single trusted entry point for credential-dependent provider
// operations. Its configuration is read-only after construction and it is
// safe for concurrent use.
type Client struct {
	// policy is validated at construction (D1) and enforced per call.
	policy ResourcePolicy
	// admission is shared by all calls, so its limits hold across them.
	admission   admitter
	bindings    BindingResolver
	credentials CredentialResolver
	catalog     Catalog
	catalogHash string
	adapters    map[API]adapter
	http        *http.Client
	loopback    bool
	// observations is nil without an Observer.
	observations *observations
	probe        *probe.Probe // nil outside acceptance tests
	clock        *clock.Clock // nil is the system clock
}

// NewClient validates cfg and builds a Client.
func NewClient(cfg Config) (*Client, error) {
	if err := cfg.Policy.validate(); err != nil {
		return nil, err
	}
	if cfg.Bindings == nil {
		return nil, &ConfigError{Field: "Bindings", Problem: "a binding resolver is required"}
	}
	if cfg.Credentials == nil {
		return nil, &ConfigError{Field: "Credentials", Problem: "a credential resolver is required"}
	}
	if cfg.Observer != nil && cfg.Policy.MaxQueuedObservations == 0 {
		return nil, policyError("MaxQueuedObservations", "must be positive when an Observer is configured")
	}
	catalog := BuiltinCatalog()
	if cfg.Catalog != nil {
		catalog = cfg.Catalog.clone()
	}
	catalogHash, err := catalog.Hash()
	if err != nil {
		return nil, err
	}
	return &Client{
		policy:       *cfg.Policy,
		bindings:     cfg.Bindings,
		credentials:  cfg.Credentials,
		catalog:      catalog,
		catalogHash:  catalogHash,
		adapters:     registry(),
		admission:    admitter{local: newLimiter(cfg.Policy, cfg.Probe), host: cfg.Admission},
		http:         newHTTPClient(cfg.Transport, cfg.Policy),
		loopback:     cfg.AllowLoopbackHTTP,
		observations: newObservations(cfg.Observer, cfg.Policy.MaxQueuedObservations),
		probe:        cfg.Probe,
		clock:        cfg.Clock,
	}, nil
}

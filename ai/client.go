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
	// Catalog defaults to BuiltinCatalog when nil.
	Catalog *Catalog
	// Transport is the only way to customize networking and is fixed for the
	// Client's lifetime; there is no per-call override. Nil selects a transport
	// that ignores proxy environment variables. Tests inject a loopback-only one.
	Transport http.RoundTripper
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
	// policy is validated at construction (D1). Its byte and event queue
	// limits are enforced per call; admission and time limits are ticket 13's.
	policy      ResourcePolicy
	bindings    BindingResolver
	credentials CredentialResolver
	catalog     Catalog
	adapters    map[API]adapter
	http        *http.Client
	loopback    bool
	probe       *probe.Probe // nil outside acceptance tests
	clock       *clock.Clock // nil is the system clock
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
	catalog := BuiltinCatalog()
	if cfg.Catalog != nil {
		catalog = cfg.Catalog.clone()
	}
	return &Client{
		policy:      *cfg.Policy,
		bindings:    cfg.Bindings,
		credentials: cfg.Credentials,
		catalog:     catalog,
		adapters:    registry(),
		http:        newHTTPClient(cfg.Transport),
		loopback:    cfg.AllowLoopbackHTTP,
		probe:       cfg.Probe,
		clock:       cfg.Clock,
	}, nil
}

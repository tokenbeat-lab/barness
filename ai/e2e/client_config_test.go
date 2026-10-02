package e2e

import (
	"errors"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

// TestClientConstruction covers D1 / ADR-0002's construction check: a Client
// only exists with an explicit, finite and internally consistent resource
// policy and both trusted resolvers.
func TestClientConstruction(t *testing.T) {
	h := host.New()
	baseConfig := func() ai.Config {
		return ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h}
	}

	// Every way construction must fail, written before the implementation.
	failures := []struct {
		id     string
		field  string
		mutate func(*ai.Config)
	}{
		{"missing-policy", "Policy", func(c *ai.Config) { c.Policy = nil }},
		{"missing-binding-resolver", "Bindings", func(c *ai.Config) { c.Bindings = nil }},
		{"missing-credential-resolver", "Credentials", func(c *ai.Config) { c.Credentials = nil }},

		{"zero-request-bytes", "Policy.MaxRequestBytes", func(c *ai.Config) { c.Policy.MaxRequestBytes = 0 }},
		{"zero-image-bytes", "Policy.MaxImageBytes", func(c *ai.Config) { c.Policy.MaxImageBytes = 0 }},
		{"zero-frame-bytes", "Policy.MaxFrameBytes", func(c *ai.Config) { c.Policy.MaxFrameBytes = 0 }},
		{"zero-tool-json-bytes", "Policy.MaxToolJSONBytes", func(c *ai.Config) { c.Policy.MaxToolJSONBytes = 0 }},
		{"zero-error-body-bytes", "Policy.MaxErrorBodyBytes", func(c *ai.Config) { c.Policy.MaxErrorBodyBytes = 0 }},
		{"zero-output-bytes", "Policy.MaxOutputBytes", func(c *ai.Config) { c.Policy.MaxOutputBytes = 0 }},
		{"zero-queued-events", "Policy.MaxQueuedEvents", func(c *ai.Config) { c.Policy.MaxQueuedEvents = 0 }},
		{"zero-queued-event-bytes", "Policy.MaxQueuedEventBytes", func(c *ai.Config) { c.Policy.MaxQueuedEventBytes = 0 }},
		{"zero-tenant-concurrency", "Policy.MaxConcurrentPerTenant", func(c *ai.Config) { c.Policy.MaxConcurrentPerTenant = 0 }},
		{"zero-process-concurrency", "Policy.MaxConcurrentProcess", func(c *ai.Config) { c.Policy.MaxConcurrentProcess = 0 }},
		{"zero-admission-waiters", "Policy.MaxAdmissionWaiters", func(c *ai.Config) { c.Policy.MaxAdmissionWaiters = 0 }},
		{"zero-call-timeout", "Policy.CallTimeout", func(c *ai.Config) { c.Policy.CallTimeout = 0 }},
		{"zero-connect-timeout", "Policy.ConnectTimeout", func(c *ai.Config) { c.Policy.ConnectTimeout = 0 }},
		{"zero-response-header-timeout", "Policy.ResponseHeaderTimeout", func(c *ai.Config) { c.Policy.ResponseHeaderTimeout = 0 }},
		{"zero-read-idle-timeout", "Policy.ReadIdleTimeout", func(c *ai.Config) { c.Policy.ReadIdleTimeout = 0 }},

		{"negative-request-bytes", "Policy.MaxRequestBytes", func(c *ai.Config) { c.Policy.MaxRequestBytes = -1 }},
		{"negative-output-bytes", "Policy.MaxOutputBytes", func(c *ai.Config) { c.Policy.MaxOutputBytes = -1 }},
		{"negative-queued-events", "Policy.MaxQueuedEvents", func(c *ai.Config) { c.Policy.MaxQueuedEvents = -1 }},
		{"negative-process-concurrency", "Policy.MaxConcurrentProcess", func(c *ai.Config) { c.Policy.MaxConcurrentProcess = -1 }},
		{"negative-admission-waiters", "Policy.MaxAdmissionWaiters", func(c *ai.Config) { c.Policy.MaxAdmissionWaiters = -1 }},
		{"negative-call-timeout", "Policy.CallTimeout", func(c *ai.Config) { c.Policy.CallTimeout = -time.Second }},
		{"negative-admission-wait", "Policy.AdmissionWait", func(c *ai.Config) { c.Policy.AdmissionWait = -time.Millisecond }},
		{"negative-queued-observations", "Policy.MaxQueuedObservations", func(c *ai.Config) { c.Policy.MaxQueuedObservations = -1 }},
		// An Observer needs a bounded queue; without one 0 is valid.
		{"observer-without-queue", "Policy.MaxQueuedObservations", func(c *ai.Config) { c.Observer = newRecorder() }},

		// Field relationships.
		{"image-exceeds-request", "Policy.MaxImageBytes", func(c *ai.Config) { c.Policy.MaxImageBytes = c.Policy.MaxRequestBytes + 1 }},
		{"tool-json-exceeds-output", "Policy.MaxToolJSONBytes", func(c *ai.Config) { c.Policy.MaxToolJSONBytes = c.Policy.MaxOutputBytes + 1 }},
		{"tenant-exceeds-process-concurrency", "Policy.MaxConcurrentPerTenant", func(c *ai.Config) {
			c.Policy.MaxConcurrentPerTenant = c.Policy.MaxConcurrentProcess + 1
		}},
		{"admission-wait-exceeds-call", "Policy.AdmissionWait", func(c *ai.Config) { c.Policy.AdmissionWait = c.Policy.CallTimeout + 1 }},
		{"connect-exceeds-call", "Policy.ConnectTimeout", func(c *ai.Config) { c.Policy.ConnectTimeout = c.Policy.CallTimeout + 1 }},
		{"response-header-exceeds-call", "Policy.ResponseHeaderTimeout", func(c *ai.Config) {
			c.Policy.ResponseHeaderTimeout = c.Policy.CallTimeout + 1
		}},
		{"read-idle-exceeds-call", "Policy.ReadIdleTimeout", func(c *ai.Config) { c.Policy.ReadIdleTimeout = c.Policy.CallTimeout + 1 }},
	}
	for _, f := range failures {
		t.Run(f.id, func(t *testing.T) {
			ev := run.Case(t, "D1-construct-reject-"+f.id)
			cfg := baseConfig()
			f.mutate(&cfg)
			client, err := ai.NewClient(cfg)
			ev.Record("outcome", map[string]any{"error": errString(err), "client_nil": client == nil})
			ev.Check("construction fails", err != nil && client == nil, "client=%v err=%v", client != nil, err)
			ev.Check("error is ErrInvalidConfig", errors.Is(err, ai.ErrInvalidConfig), "err=%v", err)
			var cfgErr *ai.ConfigError
			ev.Check("error names the field", errors.As(err, &cfgErr) && cfgErr.Field == f.field,
				"want field %q, got %+v", f.field, cfgErr)
		})
	}

	t.Run("boundaries-accepted", func(t *testing.T) {
		ev := run.Case(t, "D1-construct-accept-boundaries")
		cfg := baseConfig()
		// Equal values are valid relationships; a zero admission wait means
		// "reject immediately when at capacity", which is still bounded.
		cfg.Policy.MaxImageBytes = cfg.Policy.MaxRequestBytes
		cfg.Policy.MaxToolJSONBytes = cfg.Policy.MaxOutputBytes
		cfg.Policy.MaxConcurrentPerTenant = cfg.Policy.MaxConcurrentProcess
		cfg.Policy.AdmissionWait = 0
		cfg.Policy.ConnectTimeout = cfg.Policy.CallTimeout
		cfg.Policy.ResponseHeaderTimeout = cfg.Policy.CallTimeout
		cfg.Policy.ReadIdleTimeout = cfg.Policy.CallTimeout
		cfg.Policy.MaxQueuedEvents = 1
		client, err := ai.NewClient(cfg)
		ev.Record("outcome", map[string]any{"error": errString(err)})
		ev.Check("construction succeeds", err == nil && client != nil, "err=%v", err)
	})
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

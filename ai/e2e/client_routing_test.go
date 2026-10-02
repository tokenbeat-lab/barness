package e2e

import (
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestClientRouting covers how a call reaches its adapter: endpoints obey the
// assembly's transport rules and the Client's configuration is fixed at
// construction. Sharing one Client across tenants is E06
// (TestTenantIsolation). Model authorization (AllowedModels ∩ catalog) is
// part of the preflight matrix in TestPreflightRejections.
func TestClientRouting(t *testing.T) {
	t.Run("loopback-http-refused-without-test-assembly", func(t *testing.T) {
		ev, _, f := startHardeningCase(t, "P01-E07-loopback-http-refused-without-test-assembly")
		// Same binding, but the Client is assembled as production would be: no
		// test opt-in for plain-http loopback endpoints.
		w := newWorldWith(t, func(c *ai.Config) { c.AllowLoopbackHTTP = false }, tenantA)
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))
		res, err := w.client.Complete(ctxFor(t), textScope("req-loopback"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		checkFailed(ev, res, err)
		ev.Check("error code", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest}), "got %v", err)
	})

	t.Run("config-read-only-after-construction", func(t *testing.T) {
		ev, _, f := startHardeningCase(t, "P01-E01-config-read-only-after-construction")
		catalog := ai.BuiltinCatalog()
		policy := validPolicy()
		w := newWorldWith(t, func(c *ai.Config) { c.Catalog, c.Policy = &catalog, policy }, tenantA)
		// The caller changes everything it handed over; the Client must not notice.
		for i := range catalog.Models {
			catalog.Models[i].ID = "mutated"
		}
		catalog.Models = nil
		*policy = ai.ResourcePolicy{}
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))

		res, err := w.client.Complete(ctxFor(t), textScope("req-read-only"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call still resolves the model it was built with", err == nil, "err=%v", err)
		checkFinalMessage(ev, f, res.Message)
		checkTextRequest(ev, w, f, tenantA, false)
	})

}

func textOf(m ai.AssistantMessage) string {
	var out string
	for _, c := range m.Content {
		if t, ok := c.(ai.Text); ok {
			out += t.Text
		}
	}
	return out
}

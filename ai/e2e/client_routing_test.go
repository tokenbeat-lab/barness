package e2e

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestClientRouting covers how a call reaches its adapter: endpoints obey the
// assembly's transport rules, the Client is safe to share across tenants, and
// its configuration is fixed at construction. Model authorization (AllowedModels
// ∩ catalog) is part of the preflight matrix in TestPreflightRejections.
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

	t.Run("shared-client-concurrent-tenants", func(t *testing.T) {
		ev := run.Case(t, "P01-E06-shared-client-concurrent-tenants")
		f, raw := loadTextFixture(t, "text-basic.json")
		ev.Fixture("fixture.json", raw)
		w := newWorld(t, tenantA, tenantB)
		const calls = 16
		replies := make([]provider.Reply, calls)
		for i := range replies {
			replies[i] = sseReply(t, f, provider.FramingBytewise)
		}
		enqueue(ev, w, replies...)
		tenantOf := func(i int) tenantKey { return []tenantKey{tenantA, tenantB}[i%2] }

		// Both tenants use the same binding and model names on one Client.
		results := make([]ai.Result, calls)
		errs := make([]error, calls)
		var wg sync.WaitGroup
		for i := range calls {
			scope := ai.CallScope{TenantID: tenantOf(i).tenant, RequestID: fmt.Sprintf("req-concurrent-%02d", i)}
			wg.Go(func() {
				if i%4 < 2 {
					results[i], errs[i] = w.client.Complete(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
					return
				}
				s := w.client.Stream(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
				defer s.Close()
				for s.Next() {
				}
				results[i], errs[i] = s.Result()
			})
		}
		wg.Wait()

		ev.Record("results", results)
		perAlias := map[string]int{}
		for _, r := range w.provider.Requests() {
			perAlias[r.KeyAlias]++
		}
		ev.Record("requests_per_key", perAlias)
		ev.Check("each tenant's own key used for each of its calls",
			perAlias[tenantA.alias] == calls/2 && perAlias[tenantB.alias] == calls/2 && len(perAlias) == 2, "got %v", perAlias)
		for i := range calls {
			want := tenantOf(i).tenant
			ev.Check(fmt.Sprintf("call %02d succeeds for its tenant", i),
				errs[i] == nil && results[i].Metadata.TenantID == want && results[i].Metadata.RequestID == fmt.Sprintf("req-concurrent-%02d", i) &&
					textOf(results[i].Message) == f.Expect.Text,
				"err=%v metadata=%+v", errs[i], results[i].Metadata)
		}
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
		checkTextRequest(ev, w, f, tenantA)
	})

	t.Run("request-copied-on-receipt", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "P01-E01-request-copied-on-receipt")
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))
		release := w.host.HoldBindings()
		req := textRequest(f)
		s := w.client.Stream(ctxFor(t), textScope("req-copy"), textTarget(f), req, nil)
		// Mutate the caller's request while the call is still resolving.
		req.Messages[0].(ai.UserMessage).Content[0] = ai.Text{Text: "mutated after hand-over"}
		req.Messages[0] = ai.UserText("replaced after hand-over")
		release()
		checkStreamedText(ev, f, s)
		checkTextRequest(ev, w, f, tenantA)
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

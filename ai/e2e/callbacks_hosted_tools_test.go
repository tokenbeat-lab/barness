package e2e

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestTrustedCallbackHostedTools covers the binding's hosted tool allowance
// (ADR-0005): a payload callback may declare a provider-hosted tool only when
// the call's binding names its type, and the allowance relaxes nothing else.
func TestTrustedCallbackHostedTools(t *testing.T) {
	webSearch := map[string]any{"type": "web_search"}
	addTools := func(tools ...map[string]any) func(map[string]any) {
		return func(b map[string]any) {
			list, _ := b["tools"].([]any)
			for _, tool := range tools {
				list = append(list, tool)
			}
			b["tools"] = list
		}
	}
	hooksFor := func(change func(map[string]any)) ai.Hooks {
		return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			change(p.Body)
			return ai.KeepPayload(), nil
		}}
	}
	allowWebSearch := func(b *ai.Binding) { b.AllowedHostedTools = []string{"web_search"} }

	t.Run("allowed-type-is-sent", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-hosted-allowed")
		w.updateBinding(tenantA, allowWebSearch)
		res, err := w.client.WithHooks(hooksFor(addTools(webSearch))).Complete(ctxFor(t), textScope("req-hosted-allowed"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if ev.Check("exactly one request", len(reqs) == 1, "got %d", len(reqs)) {
			var body struct {
				Tools []map[string]any `json:"tools"`
			}
			mustUnmarshal(t, reqs[0].Body, &body)
			ev.Check("hosted tool sent as declared", len(body.Tools) == 1 && body.Tools[0]["type"] == "web_search", "tools=%v", body.Tools)
		}
	})

	refusedUnderAllowance := []struct {
		name   string
		change func(map[string]any)
	}{
		{"other-hosted-type", addTools(webSearch, map[string]any{"type": "file_search", "vector_store_ids": []any{"vs_other_tenant"}})},
		{"mcp-not-named", addTools(map[string]any{"type": "mcp", "server_label": "x", "server_url": "https://attacker.example/mcp"})},
		{"vendor-state-still-refused", func(b map[string]any) {
			addTools(webSearch)(b)
			b["previous_response_id"] = "resp_other_tenant"
		}},
		{"storage-still-refused", func(b map[string]any) {
			addTools(webSearch)(b)
			b["store"] = true
		}},
		{"untyped-tool", addTools(map[string]any{"name": "x"})},
	}
	for _, c := range refusedUnderAllowance {
		t.Run("refused-"+c.name, func(t *testing.T) {
			ev, w, f := startCallbackCase(t, "P01-E04-callbacks-hosted-refused-"+c.name)
			w.updateBinding(tenantA, allowWebSearch)
			res, err := w.client.WithHooks(hooksFor(c.change)).Complete(ctxFor(t), textScope("req-hosted-"+c.name), textTarget(f), textRequest(f), nil)
			checkRefused(ev, w, res, err)
		})
	}

	// The allowance is the binding's: the same callback is honored for the
	// tenant whose binding names the type and refused for the one whose
	// binding does not.
	t.Run("allowance-is-per-binding", func(t *testing.T) {
		ev := run.Case(t, "P01-E04-callbacks-hosted-per-binding")
		f, raw := loadTextFixture(t, "text-basic.json")
		ev.Fixture("fixture.json", raw)
		w := newWorld(t, tenantA, tenantB)
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))
		w.updateBinding(tenantA, allowWebSearch)
		hc := w.client.WithHooks(hooksFor(addTools(webSearch)))

		_, errA := hc.Complete(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-hosted-a"}, textTarget(f), textRequest(f), nil)
		resB, errB := hc.Complete(ctxFor(t), ai.CallScope{TenantID: tenantB.tenant, RequestID: "req-hosted-b"}, textTarget(f), textRequest(f), nil)
		ev.Record("result-b", resB)
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		ev.Check("allowed tenant's call succeeds", errA == nil, "err=%v", errA)
		var aiErr *ai.Error
		ev.Check("other tenant refused as tenant_denied in the request phase",
			errors.As(errB, &aiErr) && aiErr.Code == ai.CodeTenantDenied && aiErr.Phase == ai.PhaseRequest, "err=%v", errB)
		ev.Check("only the allowed tenant's request was sent", len(reqs) == 1 && reqs[0].KeyAlias == tenantA.alias, "got %d", len(reqs))
	})

	// A malformed allowance is a host configuration fault: the binding is
	// refused before any callback runs or anything is sent.
	for _, bad := range []struct {
		name  string
		types []string
	}{
		{"empty-type", []string{""}},
		{"function-type", []string{"web_search", "function"}},
	} {
		t.Run("invalid-allowance-"+bad.name, func(t *testing.T) {
			ev, w, f := startCallbackCase(t, "P01-E04-callbacks-hosted-invalid-"+bad.name)
			w.updateBinding(tenantA, func(b *ai.Binding) { b.AllowedHostedTools = slices.Clone(bad.types) })
			called := false
			hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				called = true
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-hosted-invalid-"+bad.name), textTarget(f), textRequest(f), nil)
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())
			checkFailed(ev, res, err)
			var aiErr *ai.Error
			ev.Check("refused as invalid_request in the binding phase",
				errors.As(err, &aiErr) && aiErr.Code == ai.CodeInvalidRequest && aiErr.Phase == ai.PhaseBinding, "err=%v", err)
			ev.Check("callback never ran", !called, "called")
			ev.Check("nothing sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		})
	}
}

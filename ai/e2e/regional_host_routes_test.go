package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestRegionalProviderHostRoutes(t *testing.T) {
	for _, service := range []ai.ProviderID{"zhipu", "zai", "minimax", "minimax-cn"} {
		for route := 0; route < 3; route++ {
			models := []string{"MiniMax-M3"}
			if service == "zhipu" || service == "zai" {
				models = []string{"glm-5.3", "glm-5.2"}
			}
			for _, model := range models {
				for _, entry := range []string{"complete", "stream", "simple-complete", "simple-stream"} {
					for _, hooked := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%d/%s/%s/hooks-%t", service, route, model, entry, hooked), func(t *testing.T) {
							w := newRegionalHostWorld(t)
							b := w.bindings[service][route]
							ev := run.Case(t, regionalHostCaseID(b, model, entry, hooked))
							reply, raw := regionalHostScript(t, b, model)
							ev.Fixture("synthetic-script.json", raw)
							w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
							var c regionalCaller = w.client
							if hooked {
								c = w.client.WithHooks(ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
									h.Set("X-Host-Trace", "regional-safe")
									return nil
								}})
							}
							out := regionalHostCall(t, c, b, model, entry, "route-call")
							out.record(ev)
							ev.Record("bindings", w.bindings)
							ev.Record("requests", w.provider.Requests())
							ev.Record("secret-read-count", w.secretReads)
							meta := out.result.Metadata
							ev.Check("four public entries preserve exact service/protocol/model snapshot", out.err == nil && meta.ProviderID == service && meta.API == b.API && meta.ModelID == model && meta.AccountScopeID == "same-account" && meta.CredentialVersion == "key-v1" && meta.BindingVersion == "v1" && len(meta.Attempts) == 1, "err=%v metadata=%+v", out.err, meta)
							ev.Check("one shared credential record is valid across authorized protocols", w.secretReads[service] == 1, "reads=%v", w.secretReads)
							requests := w.provider.Requests()
							suffix := "/chat/completions"
							if route == 1 {
								suffix = "/responses"
							}
							if route == 2 {
								suffix = "/v1/messages"
							}
							ev.Check("one request to explicitly chosen gateway and wire model", len(requests) == 1 && requests[0].Path == strings.TrimPrefix(b.Endpoint, w.provider.URL())+suffix, "requests=%+v", requests)
							if len(requests) == 1 {
								fields := bodyFields(t, requests[0].Body)
								ev.Check("wire model has no region suffix", string(fields["model"]) == fmt.Sprintf("%q", model), "model=%s", fields["model"])
								if hooked {
									ev.Check("trusted hooks preserve route", requests[0].Header["X-Host-Trace"] == "regional-safe", "header missing")
								}
							}
							fixed, ok := w.client.Catalog().Lookup(service, b.API, model)
							ev.Check("Client owns catalog snapshot after caller mutation", ok && fixed.Input[0] == ai.ModalityText && fixed.Cost.Input != 999 && meta.CatalogHash != "", "model=%+v", fixed)
							ev.Record("resources", map[string]int64{"calls": w.gauges.ActiveCalls(), "permits": w.gauges.Permits(), "queued": w.gauges.QueuedEvents(), "openBodies": w.bodies.Open()})
							ev.Check("all resources released", w.gauges.ActiveCalls() == 0 && w.gauges.Permits() == 0 && w.gauges.QueuedEvents() == 0 && w.bodies.Open() == 0, "resource leak")
						})
					}
				}
			}
		}
	}
}

func TestRegionalProviderHostRevocationRollback(t *testing.T) {
	ev := run.Case(t, "regional-host-revocation-rollback")
	w := newRegionalHostWorld(t)
	for _, service := range []ai.ProviderID{"zai", "minimax-cn"} {
		b := w.bindings[service][0]
		b.Enabled = false
		w.host.PutBinding(b)
		out := regionalHostCall(t, w.client, b, b.AllowedModels[0], "complete", "disabled-"+string(service))
		failure, ok := out.err.(*ai.Error)
		ev.Record("disabled-"+string(service), out.result)
		ev.Check("disable refuses new calls before secrets or attempts", ok && failure.Code == ai.CodeTenantDenied && len(out.result.Metadata.Attempts) == 0 && w.secretReads[service] == 0, "err=%v", out.err)
	}
	ev.Check("no cross-market fallback while new services disabled", len(w.provider.Requests()) == 0, "requests=%+v", w.provider.Requests())
	// Existing services keep working when only the new markets are disabled.
	for _, service := range []ai.ProviderID{"zhipu", "minimax"} {
		b := w.bindings[service][0]
		reply, raw := regionalHostScript(t, b, b.AllowedModels[0])
		ev.Fixture(string(service)+"-script.json", raw)
		w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
		out := regionalHostCall(t, w.client, b, b.AllowedModels[0], "complete", "surviving-"+string(service))
		ev.Record(string(service)+"-result", out.result)
		ev.Check("other market remains independently authorized", out.err == nil && out.result.Metadata.ProviderID == service, "err=%v", out.err)
	}
	// Rollback order: bindings disabled first, remove new identities before the
	// old Client is loaded. Re-enabling a stale binding cannot borrow a sibling.
	old, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, service := range []ai.ProviderID{"zai", "minimax-cn"} {
		b := w.bindings[service][0]
		b.Enabled = true
		w.host.PutBinding(b)
		out := regionalHostCall(t, old, b, b.AllowedModels[0], "complete", "rollback-"+string(service))
		failure, ok := out.err.(*ai.Error)
		ev.Record("rollback-"+string(service), out.result)
		ev.Check("old catalog rejects new identity before secrets", ok && failure.Code == ai.CodeInvalidRequest && failure.Phase == ai.PhaseCapability && w.secretReads[service] == 0 && len(out.result.Metadata.Attempts) == 0, "err=%v", out.err)
	}
	ev.Record("requests", w.provider.Requests())
	ev.Record("secret-read-count", w.secretReads)
	ev.Check("rollback sends only the two explicit old-market requests", len(w.provider.Requests()) == 2, "requests=%+v", w.provider.Requests())
}

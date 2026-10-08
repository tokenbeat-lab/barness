package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMixedLocalGuardrails(t *testing.T) {
	for _, mode := range []string{"missing-policy", "missing-key", "duplicate-binding", "shared-account-mismatch", "shared-provider-mismatch", "disabled", "foreign-tenant"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "E10-mixed-local-guard-"+mode)
			t.Setenv("BARNESS_MIXED_KEY", tenantA.secret)
			srv := localProvider(t)
			cfg := localassembly.OperationsConfig{TenantID: tenantA.tenant, Policy: localassembly.MixedPolicy(), Catalog: ptrMixedCatalog(), Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true, Keys: map[string]localassembly.KeySource{"openai": {EnvVar: "BARNESS_MIXED_KEY"}}}
			for _, r := range mixedRoutes[:2] {
				cfg.Bindings = append(cfg.Bindings, r.binding(tenantA, srv.URL()))
			}
			switch mode {
			case "missing-policy":
				cfg.Policy = nil
			case "missing-key":
				cfg.Keys = nil
			case "duplicate-binding":
				cfg.Bindings[1].BindingID = cfg.Bindings[0].BindingID
			case "shared-account-mismatch":
				cfg.Bindings[1].AccountScopeID = "other"
			case "shared-provider-mismatch":
				cfg.Bindings[1].ProviderID = ai.ProviderGoogle
			case "disabled":
				cfg.Bindings[1].Enabled = false
			}
			a, err := localassembly.OpenOperations(cfg)
			if mode == "disabled" || mode == "foreign-tenant" {
				if err != nil {
					t.Fatal(err)
				}
				scope := a.Scope()
				if mode == "foreign-tenant" {
					scope.TenantID = tenantB.tenant
				}
				res, e := a.Client.GenerateImages(ctxFor(t), scope, ai.Target{BindingID: "openai-image", ModelID: mixedModel}, ai.ImagesRequest{Prompt: "synthetic"}, nil)
				ev.Record("result", res)
				ev.Record("error", errString(e))
				ev.Check("local boundary fails closed", e != nil && len(res.Metadata.Attempts) == 0, "got %v", e)
			} else {
				ev.Record("error", errString(err))
				ev.Check("invalid assembly refused", err != nil && a == nil, "got %v", err)
			}
			ev.Check("guard sends no Provider request", len(srv.Requests()) == 0, "sent %d", len(srv.Requests()))
		})
	}
}

package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

func TestMiniMaxHostLocalAssembly(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			t.Run(r.name+"/"+string(purpose), func(t *testing.T) {
				ev := run.Case(t, "MM08-local-assembly-"+r.name+"-"+string(purpose))
				w := newMiniMaxHostWorld(t)
				cat := minimax.CandidateCatalog()
				key := miniMaxHostKey(tenantA, purpose)
				variable := "BARNESS_MM08_STANDARD_KEY"
				if purpose == minimaxhost.MPlan {
					variable = "BARNESS_MM08_QUALIFIED_PLAN_KEY"
				}
				t.Setenv(variable, key.secret)
				t.Setenv("MINIMAX_API_KEY", tenantB.secret)
				b := w.host.Binding(tenantA.tenant, miniMaxHostTarget(purpose, r).BindingID)
				assembly, err := localassembly.Open(localassembly.Config{TenantID: tenantA.tenant, Binding: b, Key: localassembly.KeySource{EnvVar: variable}, Catalog: &cat, Policy: validPolicy(), Transport: w.bodies, AllowLoopbackHTTP: true})
				if !ev.Check("local Host explicitly accepts its reviewed candidate catalog", err == nil, "%v", err) {
					return
				}
				cat.Models = nil
				w.provider.EnqueueAt(miniMaxHostPrefix(tenantA, purpose, r), miniMaxHostReply(t, r, miniMaxHostModel(purpose)))
				res, callErr := assembly.Client.Complete(ctxFor(t), assembly.Scope(), assembly.Target(miniMaxHostModel(purpose)), miniMaxMessagesRequest(), nil)
				ev.Record("result", res)
				ev.Record("requests", w.provider.Requests())
				ev.Check("local Client owns catalog, selected key and tenant attribution", callErr == nil && res.Metadata.TenantID == tenantA.tenant && res.Metadata.ProviderID == ai.ProviderMiniMax && res.Metadata.API == r.api && res.Metadata.CredentialVersion == "local" && len(w.provider.Requests()) == 1, "%v %+v", callErr, res.Metadata)
				if len(w.provider.Requests()) == 1 {
					ev.Check("ambient vendor key never overrides the Host source", w.provider.Requests()[0].KeyAlias == key.alias, "%+v", w.provider.Requests()[0])
				}
			})
		}
	}
}

package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

// A host aggregates real service/account identity, independently of protocol
// and tenant. Missing ProviderID would conflate the four same-account aliases.
func TestRegionalProviderAdmission(t *testing.T) {
	for _, p := range regionalRuntimeProviders {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			t.Run(string(p)+"/"+string(api), func(t *testing.T) {
				ev := run.Case(t, "RP14-admission-"+string(p)+"-"+string(api))
				w := newRegionalRuntimeWorld(t, nil)
				for _, tenant := range []string{"regional-a", "regional-b"} {
					reply := regionalRuntimeReply(t, api)
					ev.Record(tenant+"-script", reply)
					w.provider.EnqueueAt(regionalRuntimePrefix(tenant, p, api), reply)
					id := "regional-admission-" + tenant + "-" + string(p) + "-" + string(api)
					o := drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: tenant, RequestID: id}, regionalRuntimeTarget(p, api), ai.Request{Messages: []ai.Message{ai.UserText("Synthetic regional input.")}}, nil))
					regionalRuntimeRecord(ev, w, o)
					ev.Check("resolved actual service identity", o.err == nil && o.result.Metadata.ProviderID == p && o.result.Metadata.API == api && o.result.Metadata.AccountScopeID == "same-account" && o.result.Metadata.TenantID == tenant, "%v %+v", o.err, o.result.Metadata)
					regionalRuntimeAdmitted(ev, w, tenant, p, id, 1)
				}
				w.released(ev)
			})
		}
	}
}

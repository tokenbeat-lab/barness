package e2e

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/glmhost"
)

func TestGLMHostDisable(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			t.Run(r.name+"/"+model, func(t *testing.T) {
				ev := run.Case(t, "P0-GLM13-disable-"+r.name+"-"+model)
				w := newGLMSharedWorld(t, func(c *ai.Config) { c.Catalog = nil })
				req, _, answer, raw := r.script(t, model)
				req.Tools = nil
				ev.Fixture("own-wire.json", raw)
				invoke := func(id string, k tenantKey) (glmhost.ObservedTurn, error) {
					return glmhost.Turn(ctxFor(t), w.client, ai.CallScope{TenantID: k.tenant, RequestID: id}, r.target(model), req, r.api, glmhost.FullComplete, ai.Hooks{})
				}
				w.provider.EnqueueAt(r.prefix(glmA), answer)
				prior, err := invoke("glm13-prior", glmA)
				priorBytes := mustMarshal(t, prior.Result)
				ev.Record("prior-result", prior.Result)
				ev.Check("prior consumption retained", err == nil && prior.Result.Message.Usage.Cost.Total > 0, "err=%v", err)
				arrived, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				unblock := func() { once.Do(func() { close(release) }) }
				defer unblock()
				activeReply := answer
				activeReply.OnReceive = func() { close(arrived); <-release }
				w.provider.EnqueueAt(r.prefix(glmA), activeReply)
				done := make(chan outcome, 1)
				go func() { v, e := invoke("glm13-active", glmA); done <- outcome{result: v.Result, err: e} }()
				waitFor(ev, "active request already sent", arrived)
				b := w.host.Binding(glmA.tenant, r.target(model).BindingID)
				beforeBinding := b
				b.Enabled = false
				b.Version = "b2"
				w.host.PutBinding(b)
				before := len(w.provider.Requests())
				denied, deniedErr := invoke("glm13-denied", glmA)
				ev.Record("disabled-binding", b)
				ev.Record("denied-result", denied.Result)
				ev.Record("denied-error", errString(deniedErr))
				ev.Record("denied-reads", w.host.ReadsFor("glm13-denied"))
				ev.Check("disabled binding refuses before credentials or attempts", errors.Is(deniedErr, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseBinding}) && w.host.ReadsFor("glm13-denied").Credentials == 0 && len(denied.Result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == before, "err=%v", deniedErr)
				for _, other := range glmSharedRoutes {
					for _, k := range []tenantKey{glmA, glmB} {
						if other == r && k == glmA {
							continue
						}
						otherReq, _, otherAnswer, _ := other.script(t, model)
						otherReq.Tools = nil
						w.provider.EnqueueAt(other.prefix(k), otherAnswer)
						id := fmt.Sprintf("glm13-unaffected-%s-%s", other.name, k.tenant)
						v, e := glmhost.Turn(ctxFor(t), w.client, ai.CallScope{TenantID: k.tenant, RequestID: id}, other.target(model), otherReq, other.api, glmhost.FullComplete, ai.Hooks{})
						ev.Record(id, v.Result)
						ev.Check("other route/tenant still independently authorized", e == nil, "%s err=%v", id, e)
					}
				}
				unblock()
				active := <-done
				active.record(ev)
				ev.Check("active snapshot finishes with original version and known consumption", active.err == nil && active.result.Metadata.BindingVersion == beforeBinding.Version && active.result.Metadata.CredentialVersion == "v1" && active.result.Message.Usage.Cost.Total > 0 && len(active.result.Metadata.Attempts) == 1, "err=%v metadata=%+v", active.err, active.result.Metadata)
				ev.Check("prior result unchanged", reflect.DeepEqual(priorBytes, mustMarshal(t, prior.Result)), "prior mutated")
				b.Enabled = true
				b.Version = "b3"
				w.host.PutBinding(b)
				w.provider.EnqueueAt(r.prefix(glmA), answer)
				restored, restoredErr := invoke("glm13-restored", glmA)
				ev.Record("restored-result", restored.Result)
				ev.Check("fresh call observes restored binding without fallback", restoredErr == nil && restored.Result.Metadata.BindingVersion == "b3" && restored.Result.Metadata.API == r.api && restored.Result.Metadata.ProviderID == ai.ProviderZhipu && restored.Result.Metadata.AccountScopeID == beforeBinding.AccountScopeID, "err=%v", restoredErr)
				w.record(ev, mixedOutcome{metadata: restored.Result.Metadata, usage: restored.Result.Message.Usage, chat: restored.Result, err: restoredErr})
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}

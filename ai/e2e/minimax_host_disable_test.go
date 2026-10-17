package e2e

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMiniMaxHostDisable(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			t.Run(r.name+"/"+string(purpose), func(t *testing.T) {
				ev := run.Case(t, "MM08-config-rollback-"+r.name+"-"+string(purpose))
				w := newMiniMaxHostWorld(t)
				model := miniMaxHostModel(purpose)
				prefix := miniMaxHostPrefix(tenantA, purpose, r)
				invoke := func(k tenantKey, p minimaxhost.Purpose, route miniMaxHostRoute, id string) outcome {
					return miniMaxHostInvoke(ctxFor(t), w, k, p, route, "complete", miniMaxMessagesRequest(), ai.Hooks{}, miniMaxHostScope(k, id))
				}
				w.provider.EnqueueAt(prefix, miniMaxHostReply(t, r, model))
				prior := invoke(tenantA, purpose, r, "mm08-prior")
				priorBytes := mustMarshal(t, prior.result)
				prior.record(ev)
				if !ev.Check("initial authorized route records consumption", prior.err == nil && prior.result.Message.Usage.Cost.Total > 0, "%v", prior.err) {
					return
				}
				arrived, release := make(chan struct{}), make(chan struct{})
				var once sync.Once
				unblock := func() { once.Do(func() { close(release) }) }
				defer unblock()
				held := miniMaxHostReply(t, r, model)
				held.OnReceive = func() { close(arrived); <-release }
				w.provider.EnqueueAt(prefix, held)
				done := make(chan outcome, 1)
				go func() { done <- invoke(tenantA, purpose, r, "mm08-active") }()
				waitFor(ev, "active request already sent", arrived)
				before := w.host.Binding(tenantA.tenant, miniMaxHostTarget(purpose, r).BindingID)
				disabled := before
				disabled.Enabled = false
				disabled.Version = "mm08-disabled"
				w.host.PutBinding(disabled)
				requestsBefore := len(w.provider.Requests())
				scope := miniMaxHostScope(tenantA, "mm08-disabled")
				denied := miniMaxHostInvoke(ctxFor(t), w, tenantA, purpose, r, "stream", miniMaxMessagesRequest(), ai.Hooks{}, scope)
				ev.Record("disabled-result", denied.result)
				ev.Record("disabled-error", errString(denied.err))
				ev.Record("disabled-reads", w.host.ReadsFor(scope.RequestID))
				checkRejected(ev, denied, ai.CodeTenantDenied, ai.PhaseBinding)
				ev.Check("configuration rollback makes zero credential reads and sends", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(denied.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == requestsBefore, "%+v", denied.result.Metadata)
				otherRoute := miniMaxHostRoutes[0]
				if otherRoute == r {
					otherRoute = miniMaxHostRoutes[1]
				}
				otherPurpose := minimaxhost.Standard
				if purpose == minimaxhost.Standard {
					otherPurpose = minimaxhost.MPlan
				}
				for _, other := range []struct {
					k tenantKey
					p minimaxhost.Purpose
					r miniMaxHostRoute
				}{{tenantB, purpose, r}, {tenantA, otherPurpose, r}, {tenantA, purpose, otherRoute}} {
					if other.k == tenantA && other.p == purpose && other.r == r {
						continue
					}
					w.provider.EnqueueAt(miniMaxHostPrefix(other.k, other.p, other.r), miniMaxHostReply(t, other.r, miniMaxHostModel(other.p)))
					o := invoke(other.k, other.p, other.r, "mm08-unaffected")
					ev.Record("unaffected-"+other.k.tenant+"-"+string(other.p)+"-"+other.r.name, o.result)
					ev.Check("another tenant, purpose or protocol remains independently usable", o.err == nil, "%v", o.err)
				}
				native, _ := loadTextFixture(t, "text-basic.json")
				w.provider.Enqueue(sseEvents(t, native.Events, provider.FramingLF))
				oldScope := miniMaxHostScope(tenantA, "mm08-native")
				old, oldErr := w.client.Complete(ctxFor(t), oldScope, ai.Target{BindingID: "primary", ModelID: native.Model}, ai.Request{SystemPrompt: native.SystemPrompt, Messages: []ai.Message{ai.UserText(native.User)}}, nil)
				ev.Record("native-route-result", old)
				ev.Check("old provider works during MiniMax revocation", oldErr == nil && old.Metadata.ProviderID == ai.ProviderOpenAI, "%v", oldErr)
				unblock()
				active := <-done
				ev.Record("active-result", active.result)
				ev.Check("in-flight call retains its original binding and credential snapshot", active.err == nil && active.result.Metadata.BindingVersion == before.Version && active.result.Metadata.CredentialVersion == "mm08-key-v1" && active.result.Message.Usage.Cost.Total > 0, "%v %+v", active.err, active.result.Metadata)
				ev.Check("prior result retains its usage and price snapshot", reflect.DeepEqual(priorBytes, mustMarshal(t, prior.result)), "prior result mutated")
				restored := before
				restored.Version = "mm08-restored"
				w.host.PutBinding(restored)
				w.provider.EnqueueAt(prefix, miniMaxHostReply(t, r, model))
				o := invoke(tenantA, purpose, r, "mm08-restored")
				ev.Record("restored-result", o.result)
				ev.Check("restoring authorization is a fresh call on the same route", o.err == nil && o.result.Metadata.BindingVersion == restored.Version && o.result.Metadata.API == r.api && o.result.Metadata.ModelID == model && o.result.Metadata.AccountScopeID == before.AccountScopeID, "%v", o.err)
				ev.Record("requests", w.provider.Requests())
				w.rec.awaitCall(ev, scope.RequestID)
				checkReleased(ev, w.limitWorld)
			})
		}
	}
}

func TestMiniMaxHostPurposeDenial(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			t.Run(r.name+"/"+string(purpose), func(t *testing.T) {
				ev := run.Case(t, "MM08-purpose-denied-"+r.name+"-"+string(purpose))
				w := newMiniMaxHostWorld(t)
				target := miniMaxHostTarget(purpose, r)
				target.ModelID = "MiniMax-M3.1-Flash-Preview"
				if purpose == minimaxhost.MPlan {
					target.ModelID = "MiniMax-M3"
				}
				scope := miniMaxHostScope(tenantA, "mm08-purpose-denied")
				o := drain(w.client.Stream(ctxFor(t), scope, target, miniMaxMessagesRequest(), nil))
				o.record(ev)
				ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
				ev.Record("requests", w.provider.Requests())
				ev.Check("a different model entitlement never causes credential-purpose fallback", errors.Is(o.err, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseCapability}) && w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(o.result.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "%v", o.err)
			})
		}
	}
}

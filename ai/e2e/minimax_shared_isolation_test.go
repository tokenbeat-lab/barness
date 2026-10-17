package e2e

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestMiniMaxSharedTenantRoundTrips(t *testing.T) {
	for _, model := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "MM05-tenant-tools-"+model)
			w := newMiniMaxSharedWorld(t, nil)
			out := make([]outcome, 6)
			var wg sync.WaitGroup
			for i, r := range miniMaxSharedRoutes {
				for j, k := range []tenantKey{miniMaxSharedA, miniMaxSharedB} {
					req, first, last, raw := r.script(t, model)
					ev.Fixture(r.name+".json", raw)
					w.provider.EnqueueAt(r.prefix(k), first, last)
					wg.Go(func() {
						scope := scopeFor(k, fmt.Sprintf("mm05-tools-%s-%s-%s", r.name, model, k.tenant))
						a, e := w.client.Complete(ctxFor(t), scope, r.target(model), req, r.options())
						ev.Record(scope.RequestID+"-first", a)
						out[i*2+j] = outcome{result: a, err: e}
						if e != nil {
							return
						}
						scope.RequestID += "-continue"
						out[i*2+j] = drain(w.client.Stream(ctxFor(t), scope, r.target(model), miniMaxSharedContinue(req, a.Message), r.options()))
					})
				}
			}
			wg.Wait()
			for i, o := range out {
				r, k := miniMaxSharedRoutes[i/2], []tenantKey{miniMaxSharedA, miniMaxSharedB}[i%2]
				miniMaxSharedRecord(ev, w, o)
				ev.Check("own account/protocol/tenant provenance", o.err == nil && o.result.Metadata.API == r.api && o.result.Metadata.TenantID == k.tenant && o.result.Metadata.ModelID == model && o.result.Metadata.AccountScopeID == "acct-"+k.alias && o.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "%v %+v", o.err, o.result.Metadata)
			}
			ev.Record("requests", w.provider.Requests())
			keys := map[string]string{}
			for _, q := range w.provider.Requests() {
				k := miniMaxSharedA
				if strings.HasPrefix(q.Path, "/tenant-b/") {
					k = miniMaxSharedB
				}
				ev.Check("same names select only own credential", q.KeyAlias == k.alias, "%s %s", q.Path, q.KeyAlias)
				for _, r := range miniMaxSharedRoutes {
					if !strings.HasPrefix(q.Path, r.prefix(k)) {
						continue
					}
					native, calls, results := protocolWirePairs(t, r.name, q.Body)
					if len(results) > 0 {
						ev.Check("complete native reasoning and paired tools", native > 0 && reflect.DeepEqual(calls, results), "%d %v %v", native, calls, results)
					}
					if r.name == "responses" {
						var key string
						mustUnmarshal(t, bodyFields(t, q.Body)["prompt_cache_key"], &key)
						if prev := keys[k.tenant]; prev != "" {
							ev.Check("session stable within account", prev == key, "cache key changed")
						}
						keys[k.tenant] = key
					} else {
						ev.Check("other protocols do not leak session hint", !strings.Contains(string(q.Body), "same-session") && len(bodyFields(t, q.Body)["prompt_cache_key"]) == 0, "raw session appeared")
					}
				}
			}
			ev.Check("same session isolated across tenant/purpose/account", keys["tenant-a"] != "" && keys["tenant-b"] != "" && keys["tenant-a"] != keys["tenant-b"], "cache scopes collided")
			w.released(ev)
		})
	}
}

func TestMiniMaxSharedNativeBoundaries(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, mode := range []string{"restored", "untrusted", "foreign-tenant", "account", "model-account", "api", "provider"} {
			t.Run(r.name+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM05-native-"+r.name+mode)
				w := newMiniMaxSharedWorld(t, nil)
				req, first, last, raw := r.script(t, "MiniMax-M3")
				ev.Fixture("source.json", raw)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedA), first)
				scope := scopeFor(miniMaxSharedA, "mm05-native-"+r.name+mode)
				a, e := w.client.Complete(ctxFor(t), scope, r.target("MiniMax-M3"), req, r.options())
				ev.Record("first", a)
				if !ev.Check("source tools generated", e == nil && a.Message.StopReason == ai.StopReasonToolUse, "%v", e) {
					return
				}
				m, env, _ := restore(t, persist(t, a.Message))
				if env == nil {
					t.Fatal("missing source envelope")
				}
				want := ai.NativeStateDowngrades{}
				target := r.target("MiniMax-M3")
				nextTenant := miniMaxSharedA
				switch mode {
				case "untrusted":
					want.NoEnvelope = 1
				case "foreign-tenant":
					nextTenant = miniMaxSharedB
					scope.TenantID = nextTenant.tenant
					want.NoEnvelope = 1
				case "account":
					env.AccountScopeID = "different-account"
					want.AccountMismatch = 1
				case "model-account":
					env.AccountScopeID = "different-account"
					target.ModelID = "MiniMax-M3.1-Flash-Preview"
					want.CrossModel = 1
				case "api":
					env.API = ai.APIGoogleGenerativeAI
					want.CrossModel = 1
				case "provider":
					env.ProviderID = ai.ProviderOpenAI
					want.CrossModel = 1
				}
				if mode != "untrusted" {
					attest := scope
					attest.TenantID = env.TenantID
					m.NativeState, e = ai.TrustNativeState(attest, *env)
					ev.Check("Host attests stored origin", e == nil, "%v", e)
				}
				before := persist(t, m)
				w.provider.EnqueueAt(r.prefix(nextTenant), last)
				scope.RequestID += "-continue"
				o := drain(w.client.Stream(ctxFor(t), scope, target, miniMaxSharedContinue(req, m), r.options()))
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				ev.Check("authorized downgrade remains usable and immutable", o.err == nil && o.result.Metadata.NativeStateDowngrades == want && string(before) == string(persist(t, m)), "%v %+v", o.err, o.result.Metadata)
				if len(w.provider.Requests()) == 2 {
					native, calls, results := protocolWirePairs(t, r.name, w.provider.Requests()[1].Body)
					ev.Check("only matching trusted state replays", (want == (ai.NativeStateDowngrades{}) && native > 0) || (want != (ai.NativeStateDowngrades{}) && native == 0), "native=%d", native)
					ev.Check("tool identities remain paired", len(calls) > 0 && reflect.DeepEqual(calls, results), "%v %v", calls, results)
				}
				w.released(ev)
			})
		}
	}
}

func TestMiniMaxSharedCrossProtocol(t *testing.T) {
	for _, source := range miniMaxSharedRoutes {
		for _, target := range miniMaxSharedRoutes {
			if source.name == target.name {
				continue
			}
			t.Run(source.name+"/"+target.name, func(t *testing.T) {
				ev := run.Case(t, "MM05-cross-"+source.name+target.name)
				w := newMiniMaxSharedWorld(t, nil)
				req, first, _, raw := source.script(t, "MiniMax-M3")
				ev.Fixture("source.json", raw)
				_, _, last, targetRaw := target.script(t, "MiniMax-M3")
				ev.Fixture("target.json", targetRaw)
				w.provider.EnqueueAt(source.prefix(miniMaxSharedA), first)
				w.provider.EnqueueAt(target.prefix(miniMaxSharedA), last)
				a, e := w.client.Complete(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-cross-first-"+source.name+target.name), source.target("MiniMax-M3"), req, source.options())
				ev.Record("first", a)
				if !ev.Check("source generation succeeds", e == nil, "%v", e) {
					return
				}
				o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-cross-last-"+source.name+target.name), target.target("MiniMax-M3"), miniMaxSharedContinue(req, a.Message), target.options()))
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				ev.Check("cross-protocol history downgrades transparently", o.err == nil && o.result.Metadata.NativeStateDowngrades.CrossModel == 1, "%v %+v", o.err, o.result.Metadata)
				if len(w.provider.Requests()) == 2 {
					native, calls, results := protocolWirePairs(t, target.name, w.provider.Requests()[1].Body)
					ev.Check("native stripped and tool pairing preserved", native == 0 && len(calls) > 0 && reflect.DeepEqual(calls, results), "%d %v %v", native, calls, results)
				}
				w.released(ev)
			})
		}
	}
}

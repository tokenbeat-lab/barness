package e2e

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMSharedTenantTools(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM09-shared-tenant-tools-"+model)
			w := newGLMSharedWorld(t, nil)
			var wg sync.WaitGroup
			out := make([]outcome, 6)
			for i, r := range glmSharedRoutes {
				for j, k := range []tenantKey{glmA, glmB} {
					req, first, last, raw := r.script(t, model)
					ev.Fixture(r.name+".json", raw)
					ev.Record("responses-"+r.name, []any{first, last})
					w.provider.EnqueueAt(r.prefix(k), first, last)
					wg.Go(func() {
						scope := scopeFor(k, fmt.Sprintf("glm09-%s-%s-%s-first", r.name, model, k.tenant))
						a, err := w.client.Complete(ctxFor(t), scope, r.target(model), req, r.options())
						out[i*2+j] = outcome{result: a, err: err}
						if err != nil {
							return
						}
						scope.RequestID += "-continue"
						out[i*2+j] = drain(w.client.Stream(ctxFor(t), scope, r.target(model), glmSharedContinue(req, a.Message), r.options()))
					})
				}
			}
			wg.Wait()
			for i, o := range out {
				r, k := glmSharedRoutes[i/2], []tenantKey{glmA, glmB}[i%2]
				glmSharedRecord(ev, w, o)
				ev.Check("each own continuation succeeds with preserved provenance", o.err == nil && o.result.Metadata.NativeStateDowngrades == (ai.NativeStateDowngrades{}), "%v %+v", o.err, o.result.Metadata)
				m := o.result.Metadata
				ev.Check("result and Attempt retain complete actual attribution", m.ProviderID == ai.ProviderZhipu && m.API == r.api && m.ModelID == model && m.TenantID == k.tenant && m.AccountScopeID == "acct-"+k.tenant && m.Operation == ai.OperationChat && len(m.Attempts) == 1, "%+v", m)
			}
			keys := map[string]string{}
			ev.Record("requests", w.provider.Requests())
			for _, q := range w.provider.Requests() {
				tenant, want := "tenant-a", glmA.alias
				if strings.HasPrefix(q.Path, "/tenant-b/") {
					tenant, want = "tenant-b", glmB.alias
				}
				ev.Check("same names select only own tenant credential", q.KeyAlias == want, "path %s key %s", q.Path, q.KeyAlias)
				var route glmSharedRoute
				for _, r := range glmSharedRoutes {
					if strings.Contains(q.Path, "glm-"+r.name) {
						route = r
					}
				}
				n, calls, results := protocolWirePairs(t, route.name, q.Body)
				if len(results) > 0 {
					ev.Check("own native state and valid pairs survive shared Client", (n > 0 || (route.name == "messages" && model == "glm-5.2")) && len(calls) == len(results) && reflect.DeepEqual(calls, results), "%d %v %v", n, calls, results)
				}
				if route.name == "responses" {
					var key string
					mustUnmarshal(t, bodyFields(t, q.Body)["prompt_cache_key"], &key)
					if prior := keys[tenant]; prior != "" {
						ev.Check("session key stable within tenant", prior == key, "%s", tenant)
					}
					keys[tenant] = key
				}
			}
			ev.Check("same session is isolated across tenant/account", keys["tenant-a"] != "" && keys["tenant-b"] != "" && keys["tenant-a"] != keys["tenant-b"], "%+v", keys)
			w.released(ev)
		})
	}
}

func TestGLMSharedNativeBoundaries(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, origin := range []string{"trusted", "restored", "untrusted", "tenant", "foreign-tenant", "account", "model-account", "api", "provider"} {
				t.Run(r.name+"/"+model+"/"+origin, func(t *testing.T) {
					ev := run.Case(t, "GLM09-native-"+t.Name())
					w := newGLMSharedWorld(t, nil)
					req, first, last, raw := r.script(t, model)
					ev.Fixture(r.name+".json", raw)
					ev.Record("responses-"+r.name, []any{first, last})
					w.provider.EnqueueAt(r.prefix(glmA), first)
					scope := scopeFor(glmA, "glm09-native-"+r.name+model+origin)
					a, err := w.client.Complete(ctxFor(t), scope, r.target(model), req, r.options())
					if !ev.Check("own protocol source tool history", err == nil && a.Message.StopReason == ai.StopReasonToolUse, "%v", err) {
						return
					}
					m := a.Message
					// Own 5.2 tool capture has no thinking. Obtain its independent native
					// thinking capture as another explicit generation; retain the actual
					// tool pairs in the history instead of fabricating a signed tool turn.
					if r.name == "messages" && model == "glm-5.2" {
						c := glmMessagesCaptureByName(t, model+"-thinking-enabled")
						ev.Record("independent-thinking-response", c.reply(t))
						w.provider.EnqueueAt(r.prefix(glmA), c.reply(t))
						native, err := w.client.Complete(ctxFor(t), scopeFor(glmA, scope.RequestID+"-thinking"), r.target(model), glmMessagesCapturedRequest(t, c), glmMessagesEnabled())
						if !ev.Check("own independent thinking generation", err == nil, "%v", err) {
							return
						}
						req = glmSharedContinue(req, a.Message)
						m = native.Message
					}
					nextTenant := glmA
					if origin == "foreign-tenant" {
						nextTenant = glmB
					}
					w.provider.EnqueueAt(r.prefix(nextTenant), last)
					target := r.target(model)
					want := ai.NativeStateDowngrades{}
					if origin != "trusted" && origin != "foreign-tenant" {
						var env *ai.NativeStateEnvelope
						m, env = glmSharedRestore(t, m)
						if !ev.Check("storage strips trust", env != nil, "missing envelope") {
							return
						}
						switch origin {
						case "untrusted":
							want.NoEnvelope = 1
						case "tenant":
							env.TenantID = glmB.tenant
							want.NoEnvelope = 1
						case "account":
							env.AccountScopeID = "different-authorized-account"
							want.AccountMismatch = 1
						case "model-account":
							env.AccountScopeID = "different-authorized-account"
							target.ModelID = map[string]string{"glm-5.3": "glm-5.2", "glm-5.2": "glm-5.3"}[model]
							want.CrossModel = 1
						case "api":
							env.API = ai.APIGoogleGenerativeAI
							want.CrossModel = 1
						case "provider":
							env.ProviderID = ai.ProviderOpenAI
							want.CrossModel = 1
						}
						if origin != "untrusted" {
							s := scope
							s.TenantID = env.TenantID
							m.NativeState, err = ai.TrustNativeState(s, *env)
							ev.Check("Host explicitly attests storage identity", err == nil, "%v", err)
						}
					}
					if origin == "foreign-tenant" {
						scope.TenantID = glmB.tenant
						want.NoEnvelope = 1
					}
					before := persist(t, m)
					scope.RequestID += "-continue"
					o := drain(w.client.Stream(ctxFor(t), scope, target, glmSharedContinue(req, m), r.options()))
					glmSharedRecord(ev, w, o)
					ev.Record("requests", w.provider.Requests())
					ev.Check("authorized conversion remains usable and records downgrade", o.err == nil && o.result.Metadata.NativeStateDowngrades == want, "%v %+v", o.err, o.result.Metadata)
					ev.Check("caller history stays immutable", string(before) == string(persist(t, m)), "history changed")
					qs := w.provider.Requests()
					if len(qs) >= 2 {
						n, calls, results := protocolWirePairs(t, r.name, qs[len(qs)-1].Body)
						ev.Check("only trusted matching state replays natively", (want == (ai.NativeStateDowngrades{}) && n > 0) || (want != (ai.NativeStateDowngrades{}) && n == 0), "native=%d", n)
						ev.Check("downgrade preserves effective tool pairs", len(calls) > 0 && reflect.DeepEqual(calls, results), "%v %v", calls, results)
					}
					w.released(ev)
				})
			}
		}
	}
}

package e2e

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/examples/toolloop"
)

func TestMiniMaxHostExample(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			for _, entry := range []string{"stream", "complete", "simple-stream", "simple-complete"} {
				for _, hooked := range []bool{false, true} {
					name := r.name + "/" + string(purpose) + "/" + entry + "/plain"
					if hooked {
						name = strings.TrimSuffix(name, "plain") + "hooks"
					}
					t.Run(name, func(t *testing.T) {
						ev := run.Case(t, "MM08-host-"+strings.ReplaceAll(name, "/", "-"))
						w := newMiniMaxHostWorld(t)
						model := miniMaxHostModel(purpose)
						w.provider.EnqueueAt(miniMaxHostPrefix(tenantA, purpose, r), miniMaxHostReply(t, r, model))
						hooks := ai.Hooks{}
						phases := []string{}
						if hooked {
							hooks = ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
								phases = append(phases, "headers")
								h.Set("X-Host-Trace", "safe-example")
								return nil
							}, OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
								phases = append(phases, "payload")
								return ai.KeepPayload(), nil
							}, OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
								phases = append(phases, "response")
								return nil
							}}
						}
						scope := miniMaxHostScope(tenantA, "mm08-example")
						o := miniMaxHostInvoke(ctxFor(t), w, tenantA, purpose, r, entry, miniMaxMessagesRequest(), hooks, scope)
						o.record(ev)
						ev.Record("requests", w.provider.Requests())
						ev.Record("catalog-kind", "controlled candidate with synthetic prices; no formal support")
						ev.Check("own model, purpose, account and fixed credential snapshot", o.err == nil && o.result.Metadata.ProviderID == ai.ProviderMiniMaxCN && o.result.Metadata.API == r.api && o.result.Metadata.ModelID == model && o.result.Metadata.AccountScopeID == "acct-mm08-"+tenantA.tenant+"-"+string(purpose) && o.result.Metadata.CredentialVersion == "mm08-key-v1" && o.result.Metadata.BindingVersion == "mm08-v1", "%v %+v", o.err, o.result.Metadata)
						if len(w.provider.Requests()) == 1 {
							req := w.provider.Requests()[0]
							ev.Check("one authorized protocol and credential purpose", req.Path == miniMaxHostPrefix(tenantA, purpose, r)+r.suffix && req.KeyAlias == miniMaxHostKey(tenantA, purpose).alias, "%+v", req)
							f := bodyFields(t, req.Body)
							ev.Check("omission keeps provider thinking default", len(f["thinking"]) == 0 && len(f["reasoning_effort"]) == 0 && len(f["reasoning"]) == 0, "%s", req.Body)
						}
						if hooked {
							ev.Check("trusted hook order remains fixed", reflect.DeepEqual(phases, []string{"headers", "payload", "response"}), "%v", phases)
						}
						checkCallRecords(ev, w.rec.awaitCall(ev, scope.RequestID), scope, miniMaxHostTarget(purpose, r).BindingID, o, 1)
						checkReleased(ev, w.limitWorld)
					})
				}
			}
		}
	}
}

func TestMiniMaxHostTools(t *testing.T) {
	for _, r := range miniMaxHostRoutes {
		for _, purpose := range []minimaxhost.Purpose{minimaxhost.Standard, minimaxhost.MPlan} {
			t.Run(r.name+"/"+string(purpose), func(t *testing.T) {
				ev := run.Case(t, "MM08-host-tools-"+r.name+"-"+string(purpose))
				w := newMiniMaxHostWorld(t)
				model := miniMaxHostModel(purpose)
				w.provider.EnqueueAt(miniMaxHostPrefix(tenantA, purpose, r), miniMaxHostToolReply(t, r, model), miniMaxHostReply(t, r, model))
				req := miniMaxToolRequest()
				called := 0
				ids := []string{}
				loop := toolloop.Loop{Client: w.client, Target: miniMaxHostTarget(purpose, r), MaxTurns: 2, Scope: func() ai.CallScope {
					scope := miniMaxHostScope(tenantA, "mm08-tool")
					ids = append(ids, scope.RequestID)
					return scope
				}}
				turns, err := loop.Run(ctxFor(t), req, []toolloop.Tool{{Declaration: req.Tools[0], Run: func(_ context.Context, call ai.ValidToolCall) (string, error) {
					called++
					var args struct{ Label string }
					if err := call.Decode(&args); err != nil {
						return "", err
					}
					return "Synthetic result for " + args.Label, nil
				}}})
				ev.Record("turns", turns)
				ev.Record("requests", w.provider.Requests())
				ev.Check("Host validates finished tools and bounds the round trip", err == nil && len(turns) == 2 && called > 0 && len(ids) == 2 && ids[0] != ids[1], "%v calls=%d turns=%d", err, called, len(turns))
				for _, id := range ids {
					w.rec.awaitCall(ev, id)
				}
				checkReleased(ev, w.limitWorld)
			})
		}
	}
}

package e2e

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestDecisionsRuntimeRollback(t *testing.T) {
	for _, mode := range []string{"disabled-binding", "removed-model", "revoked-images"} {
		for _, entry := range []string{"initial-images", "hook-images"} {
			t.Run(mode+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD58-runtime-"+mode+"-"+entry)
				w := newRuntimeMixedWorld(t, true, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) })
				r := runtimeMixedRoutes[3]
				b := w.host.Binding(tenantA.tenant, r.id)
				b.Retry.MaxRetries = 1
				w.host.PutBinding(b)
				recordRuntimeMixedConfig(ev, w)
				recordClassifierAssets(ev, classifierAssets(t))
				runtimeMixedEnqueue(ev, w, r, tenantA, runtimeMixedReply(t, r, tenantA))
				prior := runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-prior", ai.Hooks{}, true)
				priorBytes := mustMarshal(t, prior.classifier)
				w.record(ev, prior)
				prior.recordResult(ev, "prior-result")
				ev.Check("reported cost exists before revocation", prior.err == nil && prior.usage.Cost.Total > 0, "got %v", prior.err)
				arrived, release := make(chan struct{}), make(chan struct{})
				// Cleanup opens the barrier even if an assertion aborts this case.
				var once sync.Once
				unblock := func() { once.Do(func() { close(release) }) }
				defer unblock()
				fail := provider.Reply{Status: 429, Header: map[string]string{"retry-after-ms": "1", "x-request-id": "up-before-revoke"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic limited"}}`)}, OnReceive: func() { close(arrived); <-release }}
				fail.Script = joinChunks(fail.Chunks)
				runtimeMixedEnqueue(ev, w, r, tenantA, fail, runtimeMixedReply(t, r, tenantA))
				hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
					h.Set("X-Barness-Test-Call", "active")
					return nil
				}}
				if entry == "hook-images" {
					hooks.OnPayload = func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
						p.Body["input"] = classifierImageMessages(classifierAssets(t)[:1])
						return ai.KeepPayload(), nil
					}
				}
				done := make(chan mixedOutcome, 1)
				go func() {
					done <- runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-active", hooks, entry == "initial-images")
				}()
				waitFor(ev, "active request is frozen and sent", arrived)
				changed := b
				changed.Version = "b2"
				code, phase, credentials := ai.CodeTenantDenied, ai.PhaseCapability, 0
				switch mode {
				case "disabled-binding":
					changed.Enabled = false
					phase = ai.PhaseBinding
				case "removed-model":
					changed.AllowedModels = nil
				case "revoked-images":
					changed.AllowClassifierImages = false
					if entry == "hook-images" {
						phase = ai.PhaseRequest
						credentials = 1
					}
				}
				w.host.PutBinding(changed)
				ev.Record("revoked-binding", changed)
				before := len(w.provider.Requests())
				denied := runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-after-revoke", hooks, entry == "initial-images")
				denied.recordResult(ev, "denied-result")
				w.record(ev, denied)
				reads := w.host.ReadsFor(denied.metadata.RequestID)
				ev.Record("denied-reads", reads)
				ev.Check("new authorization rejected without fallback", errors.Is(denied.err, &ai.Error{Code: code, Phase: phase}) && len(denied.metadata.Attempts) == 0 && reads.Credentials == credentials && len(w.provider.Requests()) == before, "got %v / %+v", denied.err, reads)
				for _, k := range []tenantKey{tenantA, tenantB} {
					for _, other := range runtimeMixedRoutes {
						if k == tenantA && other == r {
							continue
						}
						runtimeMixedEnqueue(ev, w, other, k, runtimeMixedReply(t, other, k))
						o := runtimeMixedInvoke(t, ctxFor(t), w, other, k, "req-available-"+k.tenant+"-"+other.id, ai.Hooks{}, true)
						o.recordResult(ev, o.metadata.RequestID+"-result")
						w.record(ev, o)
						ev.Check("same account key does not merge authorization", o.err == nil, "got %s %s: %v", k.tenant, other.id, o.err)
					}
				}
				if mode == "revoked-images" {
					runtimeMixedEnqueue(ev, w, r, tenantA, runtimeMixedReply(t, r, tenantA))
					o := runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-text-still-allowed", ai.Hooks{}, false)
					w.record(ev, o)
					o.recordResult(ev, "text-result")
					ev.Check("text authorization remains usable", o.err == nil, "got %v", o.err)
				}
				unblock()
				var active mixedOutcome
				select {
				case active = <-done:
				case <-time.After(long):
					t.Fatal("active call did not complete")
				}
				active.recordResult(ev, "active-result")
				w.record(ev, active)
				ev.Check("in-flight retry retains authorization and cost", active.err == nil && active.metadata.BindingVersion == "b1" && len(active.metadata.Attempts) == 2 && active.usage.Cost.Total > 0, "got %v / %+v", active.err, active.metadata)
				var captures []provider.Request
				for _, req := range w.provider.Requests() {
					if req.Header["X-Barness-Test-Call"] == "active" {
						captures = append(captures, req)
					}
				}
				ev.Check("active attempts identical after revocation", len(captures) == 2 && reflect.DeepEqual(captures[0], captures[1]), "got %d captures", len(captures))
				ev.Check("prior results cannot lose reported consumption", reflect.DeepEqual(priorBytes, mustMarshal(t, prior.classifier)), "prior result mutated")
				// Restoring Host authorization is a fresh call, never a hidden route.
				b.Version = "b3"
				w.host.PutBinding(b)
				runtimeMixedEnqueue(ev, w, r, tenantA, runtimeMixedReply(t, r, tenantA))
				restored := runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-restored", ai.Hooks{}, true)
				w.record(ev, restored)
				restored.recordResult(ev, "restored-result")
				ev.Check("restored authorization observed on fresh call", restored.err == nil && restored.metadata.BindingVersion == "b3" && restored.metadata.API == ai.APIOpenAIDecisions, "got %v", restored.err)
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}

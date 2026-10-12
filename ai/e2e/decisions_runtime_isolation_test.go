package e2e

import (
	"math"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsRuntimeMixedIsolation(t *testing.T) {
	for _, shared := range []bool{false, true} {
		name := "independent-accounts"
		if shared {
			name = "shared-account-and-key"
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD44-runtime-"+name)
			w := newRuntimeMixedWorld(t, shared, nil)
			recordRuntimeMixedConfig(ev, w)
			recordClassifierAssets(ev, classifierAssets(t))
			var arrived sync.WaitGroup
			arrived.Add(8)
			ready := doneWhen(arrived.Wait)
			results := make([]mixedOutcome, 8)
			var wg sync.WaitGroup
			for n, k := range []tenantKey{tenantA, tenantB} {
				for i, r := range runtimeMixedRoutes {
					reply := runtimeMixedReply(t, r, k)
					reply.OnReceive = func() { arrived.Done(); waitFor(ev, "all eight calls in flight", ready) }
					runtimeMixedEnqueue(ev, w, r, k, reply)
					wg.Go(func() {
						results[n*4+i] = runtimeMixedInvoke(t, ctxFor(t), w, r, k, "req-runtime-"+k.tenant+"-"+r.id, ai.Hooks{}, true)
					})
				}
			}
			wg.Wait()
			for n, k := range []tenantKey{tenantA, tenantB} {
				for i, r := range runtimeMixedRoutes {
					o := results[n*4+i]
					o.recordResult(ev, o.metadata.RequestID+"-result")
					b := w.host.Binding(k.tenant, r.id)
					ev.Check("trusted tenant and full operation identity", o.err == nil && o.metadata.TenantID == k.tenant && o.metadata.ActorID == "actor-"+k.tenant && o.metadata.JobID == "job-"+k.tenant && o.metadata.Operation == r.op && o.metadata.ProviderID == r.provider && o.metadata.API == r.api && o.metadata.ModelID == mixedModel && o.metadata.AccountScopeID == b.AccountScopeID, "got %+v / %v", o.metadata, o.err)
					input := int64(31)
					if n == 1 {
						input = 47
					}
					ev.Check("own usage and unique attempt identity", o.usage.Input == input && len(o.metadata.Attempts) == 1 && o.metadata.Attempts[0].AttemptID == o.metadata.RequestID+"#1" && o.metadata.Attempts[0].ProviderRequestID == "up-"+k.tenant+"-"+r.id && o.metadata.Attempts[0].Usage.Input == input, "got %+v", o.metadata)
					if r.op == ai.OperationChat {
						ev.Check("own chat content", len(o.chat.Message.Content) == 1 && o.chat.Message.Content[0].(ai.Text).Text == "mixed-output-"+k.tenant, "wrong chat")
					}
					if r.op == ai.OperationImage {
						ev.Check("own image output", len(o.images.Content) == 1 && o.images.Content[0].(ai.ImageOutputImage).Data == mixedTenantImage(k), "wrong image")
					}
					if r.op == ai.OperationClassifier {
						choice := "track"
						if n == 1 {
							choice = "refund"
							if r.api == ai.APIOpenAIDecisions {
								choice = "other"
							}
						}
						a, ok := o.classifier.Answers["intent"].(ai.ChoiceAnswer)
						ev.Check("own classifier answer", ok && a.Choice == choice, "got %+v", a)
						if r.api == ai.APIOpenAIDecisions {
							want := .0000031
							if n == 1 {
								want = .0000047
							}
							ev.Check("Luna price independent of same-name operations", math.Abs(o.usage.Cost.Total-want) < 1e-15 && o.usage.Cost.Output == 0, "got %+v", o.usage.Cost)
						}
					}
					reads := w.host.ReadsFor(o.metadata.RequestID)
					ev.Record(o.metadata.RequestID+"-reads", reads)
					ev.Check("one own resolver snapshot", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)
					w.record(ev, o)
				}
			}
			reqs := w.provider.Requests()
			ev.Record("requests", reqs)
			ev.Check("eight actual requests", len(reqs) == 8, "got %d", len(reqs))
			for _, req := range reqs {
				k := tenantA
				if strings.HasPrefix(req.Path, "/"+tenantB.tenant+"/") {
					k = tenantB
				}
				alias := k.alias
				if shared {
					alias = tenantA.alias
				}
				ev.Check("authorized credential and private input", req.KeyAlias == alias && strings.Contains(string(req.Body), k.tenant), "got %s %s", req.Path, req.KeyAlias)
				var body map[string]any
				mustUnmarshal(t, req.Body, &body)
				ev.Check("same model stays on selected protocol", body["model"] == mixedModel, "got %v", body["model"])
			}
			w.released(ev)
		})
	}
}

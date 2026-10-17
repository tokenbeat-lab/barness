package e2e

import (
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMSharedCrossProtocolHistory(t *testing.T) {
	for _, source := range glmSharedRoutes {
		for _, target := range glmSharedRoutes {
			if source.name == target.name {
				continue
			}
			for _, model := range []string{"glm-5.3", "glm-5.2"} {
				t.Run(source.name+"/"+target.name+"/"+model, func(t *testing.T) {
					ev := run.Case(t, "GLM09-cross-protocol-"+t.Name())
					w := newGLMSharedWorld(t, nil)
					req, first, _, raw := source.script(t, model)
					ev.Fixture("source.json", raw)
					_, _, last, targetRaw := target.script(t, model)
					ev.Fixture("target.json", targetRaw)
					ev.Record("responses", []any{first, last})
					w.provider.EnqueueAt(source.prefix(glmA), first)
					w.provider.EnqueueAt(target.prefix(glmA), last)
					scope := scopeFor(glmA, "glm09-cross-"+source.name+target.name+model)
					a, err := w.client.Complete(ctxFor(t), scope, source.target(model), req, source.options())
					ev.Record("source-result", a)
					if !ev.Check("own actual source generation exposes tools", err == nil && a.Message.StopReason == ai.StopReasonToolUse, "%v", err) {
						return
					}
					scope.RequestID += "-continue"
					o := drain(w.client.Stream(ctxFor(t), scope, target.target(model), glmSharedContinue(req, a.Message), target.options()))
					glmSharedRecord(ev, w, o)
					want := ai.NativeStateDowngrades{CrossModel: 1}
					if source.name == "messages" && model == "glm-5.2" {
						want = ai.NativeStateDowngrades{}
					}
					ev.Check("authorized cross-protocol conversion succeeds with scoped downgrade", o.err == nil && o.result.Metadata.API == target.api && o.result.Metadata.NativeStateDowngrades == want, "%v %+v", o.err, o.result.Metadata)
					qs := w.provider.Requests()
					ev.Record("requests", qs)
					if len(qs) == 2 {
						native, calls, results := glmSharedWirePairs(t, target, qs[1].Body)
						ev.Check("source opaque state does not enter another protocol", native == 0, "native=%d", native)
						ev.Check("converted call IDs still pair each tool result", len(calls) > 0 && reflect.DeepEqual(calls, results), "%v %v", calls, results)
					}
					w.released(ev)
				})
			}
		}
	}
}

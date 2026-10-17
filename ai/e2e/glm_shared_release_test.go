package e2e

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Repeated failures exercise the same shared Client without Close being the
// mechanism that stops a call. Tenant A never consumes until queue overflow;
// tenant B's Provider holds an oversized unfinished frame. Both must settle
// their Result, return only their own lease, and leave bounded goroutines.
func TestGLMSharedConcurrentRelease(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		t.Run(model, func(t *testing.T) {
			ev := run.Case(t, "GLM09-concurrent-release-"+model)
			w := newGLMSharedWorld(t, func(c *ai.Config) { c.Policy.MaxQueuedEvents = 2; c.Policy.MaxFrameBytes = 1024 })
			baseline := runtime.NumGoroutine()
			const rounds = 4
			for round := range rounds {
				streams := make([]*ai.Stream, 6)
				settled := make([]callResult, 6)
				var wg sync.WaitGroup
				for i, r := range glmSharedRoutes {
					for j, k := range []tenantKey{glmA, glmB} {
						req, reply, _, raw := r.script(t, model)
						ev.Fixture(r.name+".json", raw)
						if k.tenant == glmB.tenant {
							reply = heldReply(200, "text/event-stream", "data: "+strings.Repeat("x", 1100))
						}
						ev.Record("response-"+r.name+k.tenant, reply)
						w.provider.EnqueueAt(r.prefix(k), reply)
						index := i*2 + j
						streams[index] = w.client.Stream(ctxFor(t), scopeFor(k, fmt.Sprintf("glm09-release-%s-%d-%s-%s", model, round, r.name, k.tenant)), r.target(model), req, r.options())
						wg.Go(func() { res, err := streams[index].Result(); settled[index] = callResult{res, err} })
					}
				}
				waitFor(ev, "all independent Results settle without a consumer", doneWhen(wg.Wait))
				for i, s := range streams {
					o := drain(s)
					glmSharedRecord(ev, w, o)
					phase := ai.PhaseEventQueue
					if i%2 == 1 {
						phase = ai.PhaseStream
					}
					ev.Check("own resource failure and actual Attempt retained", errors.Is(o.err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}) && len(o.result.Metadata.Attempts) == 1 && settled[i].err == o.err, "%v %+v", o.err, o.result.Metadata)
					glmSharedOneTerminal(ev, o)
				}
				w.released(ev)
			}
			ev.Record("load", map[string]int{"rounds": rounds, "concurrentCallsPerRound": 6, "failedLogicalCalls": rounds * 6, "queueLimit": 2, "frameLimit": 1024})
			ev.Record("requests", w.provider.Requests())
			checkGoroutinesSettle(ev, baseline)
		})
	}
}

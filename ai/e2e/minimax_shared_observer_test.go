package e2e

import (
	"errors"
	"github.com/tokenbeat-lab/barness/ai"
	"testing"
)

func TestMiniMaxSharedObserverFailures(t *testing.T) {
	for _, r := range miniMaxSharedRoutes {
		for _, mode := range []string{"error", "panic"} {
			t.Run(r.name+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "MM05-observer-"+r.name+mode)
				rec := newRecorder()
				if mode == "error" {
					rec.fail = errors.New("synthetic-private-observer-error")
				} else {
					rec.panics = true
				}
				w := newMiniMaxSharedWorld(t, func(c *ai.Config) { c.Observer = rec })
				w.rec = rec
				req, reply, _, raw := r.script(t, "MiniMax-M3")
				ev.Fixture("source.json", raw)
				w.provider.EnqueueAt(r.prefix(miniMaxSharedA), reply)
				o := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-observer-"+r.name+mode), r.target("MiniMax-M3"), req, r.options()))
				miniMaxSharedRecord(ev, w, o)
				ev.Record("requests", w.provider.Requests())
				ev.Check("Observer cannot change successful model call", o.err == nil && len(o.result.Metadata.Attempts) == 1, "%v", o.err)
				eventually(ev, "all failed observations accounted", func() bool { return w.client.ObserverStats().Failed == 4 })
				ev.Record("observer-stats", w.client.ObserverStats())
				miniMaxSharedOneTerminal(ev, o)
				w.released(ev)
			})
		}
	}
}

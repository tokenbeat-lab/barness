package e2e

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestMixedObserverFailures(t *testing.T) {
	for _, mode := range []string{"slow", "congested", "error", "panic"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "E09-mixed-observer-"+mode)
			rec := newRecorder()
			var release sync.Once
			if mode == "slow" || mode == "congested" {
				rec.gate = make(chan struct{})
				defer release.Do(func() { close(rec.gate) })
			}
			if mode == "error" {
				rec.fail = errors.New("synthetic observer error")
			}
			if mode == "panic" {
				rec.panics = true
			}
			w := newMixedWorld(t, func(c *ai.Config) {
				c.Observer = rec
				if mode == "congested" {
					c.Policy.MaxQueuedObservations = 1
				}
			}, tenantA)
			var wg sync.WaitGroup
			results := make([]mixedOutcome, 4)
			for i, r := range mixedRoutes {
				w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
				wg.Go(func() {
					results[i] = r.invoke(ctxFor(t), w.client, textScope("req-observer-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
				})
			}
			wg.Wait()
			for _, o := range results {
				ev.Check("Observer cannot change mixed success", o.err == nil, "got %v", o.err)
				o.recordResult(ev, o.metadata.RequestID+"-result")
			}
			failedRoute := mixedRoutes[3]
			w.provider.EnqueueAt(failedRoute.prefix(tenantA), provider.Reply{Status: 422, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(`{"error":"synthetic private error body"}`)}})
			failed := failedRoute.invoke(ctxFor(t), w.client, textScope("req-observer-failure"), ai.Target{BindingID: failedRoute.id, ModelID: mixedModel})
			ev.Check("model failure preserved with failing Observer", failed.err != nil, "missing model error")
			failed.recordResult(ev, "failed-result")
			if rec.gate != nil {
				release.Do(func() { close(rec.gate) })
			}
			eventually(ev, "all records accounted", func() bool { s := w.client.ObserverStats(); return s.Delivered+s.Failed+s.Dropped == 20 })
			stats := w.client.ObserverStats()
			ev.Record("observer-stats", stats)
			if mode == "congested" {
				ev.Check("congestion observable", stats.Dropped > 0, "got %+v", stats)
			}
			if mode == "error" || mode == "panic" {
				ev.Check("failure deliveries observable", stats.Failed == 20, "got %+v", stats)
			}
			if mode == "slow" {
				ev.Check("slow delivery preserves records", stats.Delivered == 20 && stats.Dropped == 0, "got %+v", stats)
			}
			obs := rec.all()
			ev.Record("observations-mixed", obs)
			wire := observationText(t, obs)
			for _, forbidden := range []string{"synthetic private error body", "synthetic mixed chat", "synthetic mixed image", "ticket", "Choose intent", "b64_json", "probabilities", "confidence", "answers", "Authorization", tenantA.secret} {
				ev.Check("metadata excludes "+forbidden, !strings.Contains(wire, forbidden), "leaked content")
			}
			w.released(ev)
		})
	}
}

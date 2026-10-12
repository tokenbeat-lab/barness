package e2e

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestDecisionsRuntimeObserver(t *testing.T) {
	for _, mode := range []string{"slow", "congested", "error", "panic"} {
		for _, failed := range []bool{false, true} {
			name := mode + "-success"
			if failed {
				name = mode + "-failure"
			}
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P10-OD46-runtime-"+name)
				rec := newRecorder()
				if mode == "slow" || mode == "congested" {
					rec.gate = make(chan struct{})
				}
				if mode == "error" {
					rec.fail = errors.New("synthetic observer failure")
				}
				if mode == "panic" {
					rec.panics = true
				}
				var once sync.Once
				unblock := func() {
					if rec.gate != nil {
						once.Do(func() { close(rec.gate) })
					}
				}
				defer unblock()
				w := newRuntimeMixedWorld(t, true, func(c *ai.Config) {
					c.Observer = rec
					if mode == "congested" {
						c.Policy.MaxQueuedObservations = 1
					}
				})
				w.rec = rec
				recordRuntimeMixedConfig(ev, w)
				var outcomes []mixedOutcome
				for i, r := range runtimeMixedRoutes {
					reply := runtimeMixedReply(t, r, tenantA)
					if failed {
						reply = provider.Reply{Status: 429, Header: map[string]string{"x-request-id": "up-observer", "x-typesafe-request-id": "up-observer"}, Chunks: [][]byte{[]byte(`{"error":{"message":"synthetic private error ` + tenantA.secret + `"}}`)}}
						reply.Script = joinChunks(reply.Chunks)
					}
					runtimeMixedEnqueue(ev, w, r, tenantA, reply)
					hooks := ai.Hooks{}
					if i == 0 {
						hooks.OnPayload = func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
							select {
							case <-rec.entered:
							case <-time.After(long):
								t.Error("observer never entered")
							}
							return ai.KeepPayload(), nil
						}
					}
					o := runtimeMixedInvoke(t, ctxFor(t), w, r, tenantA, "req-observer-"+r.id, hooks, true)
					outcomes = append(outcomes, o)
					o.recordResult(ev, o.metadata.RequestID+"-result")
					if failed {
						ev.Check("Observer cannot change failure", errors.Is(o.err, &ai.Error{Code: ai.CodeRateLimited, Phase: ai.PhaseRequest}) && o.usage == (ai.Usage{}) && len(o.metadata.Attempts) == 1, "got %v", o.err)
					} else {
						ev.Check("Observer cannot change success", o.err == nil && o.usage.Input == 31, "got %v", o.err)
					}
				}
				unblock()
				eventually(ev, "bounded observations accounted", func() bool { s := w.client.ObserverStats(); return s.Delivered+s.Failed+s.Dropped == 16 })
				stats := w.client.ObserverStats()
				ev.Record("observer-stats", stats)
				obs := rec.all()
				ev.Record("observer-records", obs)
				switch mode {
				case "congested":
					ev.Check("queue bounded to one waiting record", stats.Dropped == 14 && stats.Delivered == 2 && stats.Failed == 0, "got %+v", stats)
				case "error", "panic":
					ev.Check("Observer failures counted only", stats.Failed == 16 && stats.Delivered == 0 && stats.Dropped == 0, "got %+v", stats)
				case "slow":
					ev.Check("slow delivery retains records", stats.Delivered == 16 && stats.Dropped == 0, "got %+v", stats)
				}
				if mode != "congested" {
					for _, o := range outcomes {
						w.record(ev, o)
					}
				}
				wire := observationText(t, obs)
				for _, marker := range []string{"synthetic private state", "synthetic private error", "mixed-output-", "probabilities", "confidence", "image_url", classifierAssets(t)[0].Data, tenantA.secret, "Authorization"} {
					ev.Check("metadata excludes content and credentials", !strings.Contains(wire, marker), "Observer leaked content")
				}
				ev.Record("requests", w.provider.Requests())
				w.released(ev)
			})
		}
	}
}

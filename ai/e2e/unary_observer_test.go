package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestUnaryObserverFailures(t *testing.T) {
	for _, mode := range []string{"slow", "congested", "error", "panic"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "P07-E09-unary-observer-"+mode)
			u := newUnaryWorld(t, nil)
			rec := newRecorder()
			if mode == "slow" || mode == "congested" {
				rec.gate = make(chan struct{})
			}
			if mode == "error" {
				rec.fail = errors.New("synthetic observer error")
			}
			if mode == "panic" {
				rec.panics = true
			}
			// Reassemble through the public constructor to fix this Observer before use.
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
			p.MaxQueuedObservations = 64
			if mode == "congested" {
				p.MaxQueuedObservations = 1
			}
			catalog := u.client.Catalog()
			var err error
			u.client, err = ai.NewClient(ai.Config{Policy: p, Catalog: &catalog, Bindings: u.host, Credentials: u.host, Admission: u.adm, Transport: u.bodies, Probe: u.gauges, Observer: rec, AllowLoopbackHTTP: true})
			if err != nil {
				t.Fatal(err)
			}
			u.rec = rec
			enqueue(ev, u.world, u.sc.replies(t)[0])
			req := classifierInput(t, u.sc)
			req.State = json.RawMessage(`{"private":"synthetic private state marker"}`)
			res, err := u.client.WithHooks(ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				select {
				case <-rec.entered:
				case <-time.After(long):
					t.Error("observer never entered")
				}
				return ai.KeepPayload(), nil
			}}).Classify(ctxFor(t), textScope("req-observer-"+mode), u.sc.target(), req, nil)
			ev.Check("observer never changes public success", err == nil && len(res.Answers) == 1, "got %v", err)
			ev.Record("result", res)
			if rec.gate != nil {
				close(rec.gate)
			}
			if mode == "congested" {
				eventually(ev, "bounded observer counts lost observations", func() bool { s := u.client.ObserverStats(); return s.Dropped > 0 && s.Delivered+s.Dropped == 4 })
			} else {
				u.records(ev, res, err)
				eventually(ev, "observer outcomes accounted", func() bool { s := u.client.ObserverStats(); return s.Delivered+s.Failed == 4 })
			}
			stats := u.client.ObserverStats()
			ev.Record("observer-stats", stats)
			if mode == "error" || mode == "panic" {
				ev.Check("observer failures counted", stats.Failed == 4 && stats.Delivered == 0, "got %+v", stats)
			}
			if mode == "slow" {
				ev.Check("slow observer preserves records", stats.Delivered == 4 && stats.Dropped == 0, "got %+v", stats)
			}
			obs := rec.all()
			ev.Record("observer-records", obs)
			wire := observationText(t, obs)
			for _, forbidden := range []string{"synthetic private state marker", "Choose intent", "probabilities", "confidence", tenantA.secret, "Authorization", "fixture failure"} {
				ev.Check("no content or secret: "+forbidden, !strings.Contains(wire, forbidden), "observer content leaked")
			}
			u.released(ev)
		})
	}
}

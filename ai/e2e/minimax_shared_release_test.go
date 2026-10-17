package e2e

import (
	"context"
	"errors"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMiniMaxSharedConcurrentRelease(t *testing.T) {
	ev := run.Case(t, "MM05-concurrent-release")
	w := newMiniMaxSharedWorld(t, func(c *ai.Config) { c.Policy.MaxQueuedEvents = 2; c.Policy.MaxFrameBytes = 1024 })
	baseline := runtime.NumGoroutine()
	for round := range 4 {
		streams := make([]*ai.Stream, 6)
		settled := make([]callResult, 6)
		var wg sync.WaitGroup
		for i, r := range miniMaxSharedRoutes {
			for j, k := range []tenantKey{miniMaxSharedA, miniMaxSharedB} {
				req, reply, _, raw := r.script(t, "MiniMax-M3")
				ev.Fixture(r.name+".json", raw)
				if j == 1 {
					reply = heldReply(200, "text/event-stream", "data: "+strings.Repeat("x", 1100))
				}
				ev.Record(fmt.Sprintf("response-%s-%s-%d", r.name, k.tenant, round), reply)
				w.provider.EnqueueAt(r.prefix(k), reply)
				index := i*2 + j
				streams[index] = w.client.Stream(ctxFor(t), scopeFor(k, fmt.Sprintf("mm05-release-%d-%s-%s", round, r.name, k.tenant)), r.target("MiniMax-M3"), req, r.options())
				wg.Go(func() { res, e := streams[index].Result(); settled[index] = callResult{res, e} })
			}
		}
		waitFor(ev, "all independent Results settle unread", doneWhen(wg.Wait))
		for i, s := range streams {
			o := drain(s)
			miniMaxSharedRecord(ev, w, o)
			phase := ai.PhaseEventQueue
			if i%2 == 1 {
				phase = ai.PhaseStream
			}
			ev.Check("own bounded failure without replay", errors.Is(o.err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}) && len(o.result.Metadata.Attempts) == 1 && settled[i].err == o.err, "%v %+v", o.err, o.result.Metadata)
			miniMaxSharedOneTerminal(ev, o)
		}
		w.released(ev)
	}
	ev.Record("requests", w.provider.Requests())
	checkGoroutinesSettle(ev, baseline)
}

func TestMiniMaxSharedAdmissionIsolation(t *testing.T) {
	for phase, subject := range miniMaxSharedRoutes {
		t.Run(subject.name, func(t *testing.T) {
			ev := run.Case(t, "MM05-admission-"+subject.name)
			w := newMiniMaxSharedWorld(t, func(c *ai.Config) {
				c.Policy.MaxConcurrentPerTenant = 1
				c.Policy.MaxConcurrentProcess = 2
				c.Policy.MaxAdmissionWaiters = 1
				c.Policy.AdmissionWait = 300 * time.Millisecond
			})
			held := make(chan struct{})
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			req, reply, _, raw := subject.script(t, "MiniMax-M3")
			ev.Fixture("source.json", raw)
			w.provider.EnqueueAt(subject.prefix(miniMaxSharedA), reply)
			s := w.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				close(held)
				<-ctx.Done()
				return ctx.Err()
			}}).Stream(ctx, scopeFor(miniMaxSharedA, "mm05-admission-held-"+subject.name), subject.target("MiniMax-M3"), req, subject.options())
			defer s.Close()
			waitFor(ev, "subject owns tenant lease", held)
			waiter, overflow := miniMaxSharedRoutes[(phase+1)%3], miniMaxSharedRoutes[(phase+2)%3]
			a := w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-admission-wait-"+subject.name), waiter.target("MiniMax-M3"), miniMaxMessagesRequest(), waiter.options())
			defer a.Close()
			eventually(ev, "one bounded admission waiter", func() bool { return w.gauges.AdmissionWaiters() == 1 })
			b := drain(w.client.Stream(ctxFor(t), scopeFor(miniMaxSharedA, "mm05-admission-overflow-"+subject.name), overflow.target("MiniMax-M3"), miniMaxMessagesRequest(), overflow.options()))
			miniMaxSharedRecord(ev, w, b)
			ev.Check("wait queue is bounded before send", errors.Is(b.err, &ai.Error{Code: ai.CodeAdmissionDenied}) && len(b.result.Metadata.Attempts) == 0, "%v", b.err)
			w.provider.EnqueueAt(subject.prefix(miniMaxSharedB), reply)
			sibling, e := w.client.Complete(ctxFor(t), scopeFor(miniMaxSharedB, "mm05-admission-sibling-"+subject.name), subject.target("MiniMax-M3"), req, subject.options())
			ev.Record("sibling", sibling)
			ev.Check("other tenant retains independent allowance", e == nil, "%v", e)
			resultWithin(ev, "waiter ends at bounded deadline", a)
			o := drain(a)
			miniMaxSharedRecord(ev, w, o)
			ev.Check("bounded wait sends no request", errors.Is(o.err, &ai.Error{Code: ai.CodeAdmissionDenied}) && len(o.result.Metadata.Attempts) == 0, "%v", o.err)
			cancel()
			heldOut := drain(s)
			miniMaxSharedRecord(ev, w, heldOut)
			miniMaxSharedOneTerminal(ev, o)
			miniMaxSharedOneTerminal(ev, b)
			miniMaxSharedOneTerminal(ev, heldOut)
			ev.Record("requests", w.provider.Requests())
			w.released(ev)
		})
	}
}

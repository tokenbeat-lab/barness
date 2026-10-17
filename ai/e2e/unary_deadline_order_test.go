package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// A Host context may dispatch its AfterFunc asynchronously. Delaying that
// propagation makes the observed scheduler inversion deterministic at the
// public Client boundary, without replacing timers or touching private errors.
// The inverse control must still report a genuine attempt idle timeout.
func TestUnaryDeadlineOrder(t *testing.T) {
	for _, callFirst := range []bool{true, false} {
		for _, hooked := range []bool{false, true} {
			name := fmt.Sprintf("call-first-%t-hooked-%t", callFirst, hooked)
			t.Run(name, func(t *testing.T) {
				ev := run.Case(t, "P07-E08-unary-deadline-order-"+name)
				u := newUnaryWorld(t, func(c *ai.Config) { c.Policy.ReadIdleTimeout = short })
				reply := u.sc.replies(t)[0]
				reply.End = provider.EndHold
				enqueue(ev, u.world, reply)
				duration, stop := 10*short, ai.StopReasonError
				if callFirst {
					duration, stop = short, ai.StopReasonAborted
				}
				parent, cancel := context.WithTimeout(ctxFor(t), duration)
				defer cancel()
				ctx := deferredDeadlineContext{Context: parent, lag: 2 * short}
				started := time.Now()
				var res ai.ClassifierResult
				var err error
				if hooked {
					res, err = u.call(t, ctx, "req-order-"+name, ai.Hooks{})
				} else {
					res, err = u.client.Classify(ctx, textScope("req-order-"+name), u.sc.target(), classifierInput(t, u.sc), nil)
				}
				ev.Record("timing", map[string]any{"hostDeadlineAfter": duration.String(), "propagationDelay": ctx.lag.String(), "elapsed": time.Since(started).String()})
				checkUnaryFailure(ev, res, err, ai.CodeDeadlineExceeded, ai.PhaseResponse, stop)
				ev.Check("one attempt without replay", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "metadata %+v", res.Metadata)
				u.records(ev, res, err)
				u.released(ev)
				enqueue(ev, u.world, u.sc.replies(t)[0])
				next, e := u.call(t, ctxFor(t), "req-after-order-"+name, ai.Hooks{})
				ev.Check("subsequent call recovers", e == nil && len(next.Answers) == 1, "err=%v", e)
				u.released(ev)
			})
		}
	}
}

// This synthetic Host has no context values and owns cancellation delivery.
// Context's optional AfterFunc contract permits callbacks in their own
// goroutine; the bounded lag models a runnable callback awaiting scheduling.
type deferredDeadlineContext struct {
	context.Context
	lag time.Duration
}

func (deferredDeadlineContext) Value(any) any { return nil }

func (c deferredDeadlineContext) AfterFunc(f func()) func() bool {
	return context.AfterFunc(c.Context, func() {
		time.Sleep(c.lag)
		f()
	})
}

package e2e

import (
	"errors"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// goroutineSlack bounds how many goroutines may remain above the baseline
// once calls ended: idle pooled connections and the runtime's own helpers
// account for a few, independent of how many calls ran. It is a bound, not
// an exact count (spec E08: 资源探针证明，不断言精确 goroutine 数).
const goroutineSlack = 10

// TestResourceRelease covers E08's release after a limit (spec I8, I9): many
// concurrent Streams, never read or closed, that each end at a limit — a
// queue overflow, an oversized frame or error body with the provider still
// holding its connection — leave no open response body and no running call,
// and the process's goroutines return to a bound that does not grow with the
// number of calls.
func TestResourceRelease(t *testing.T) {
	ev := run.Case(t, "E08-resource-release")
	events, _ := longOutput(t, longOutputDeltas)
	lw := newLimitWorld(t, func(p *ai.ResourcePolicy) {
		p.MaxQueuedEvents = 20
		setFrameBytes(p, 2048)
		setErrorBodyBytes(p, 1024)
	})
	baseline := runtime.NumGoroutine()

	const rounds = 8
	type call struct {
		name  string
		reply provider.Reply
	}
	calls := []call{
		{"queue-unread", sseEvents(t, events, provider.FramingLF)},
		{"frame-held", heldReply(200, "text/event-stream", "event: x\ndata: "+strings.Repeat("x", 2100))},
		{"error-body-held", heldReply(500, "application/json", strings.Repeat("e", 1100))},
	}
	var all []call
	for range rounds {
		all = append(all, calls...)
	}
	for _, c := range all {
		lw.provider.Enqueue(c.reply)
	}
	ev.Record("calls", len(all))

	target := ai.Target{BindingID: "primary", ModelID: "gpt-4.1-mini"}
	req := ai.Request{Messages: []ai.Message{ai.UserText("hi")}}
	outcomes := make([]outcome, len(all))
	var wg sync.WaitGroup
	for i := range all {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Only the Result: the Stream is never read or closed, so
			// nothing but the call's own end releases its resources.
			s := lw.client.Stream(ctxFor(t), textScope("req-e08-release-"+all[i].name), target, req, nil)
			res, err := s.Result()
			outcomes[i] = outcome{result: res, err: err}
		}()
	}
	waitFor(ev, "every call ends", doneWhen(wg.Wait))

	// The provider serves replies in arrival order, so which call got which
	// reply varies; the phases they ended in do not.
	limited := map[ai.Phase]int{}
	for _, o := range outcomes {
		var ae *ai.Error
		if errors.As(o.err, &ae) && ae.Code == ai.CodeResourceLimit {
			limited[ae.Phase]++
		}
	}
	ev.Record("limited_by_phase", limited)
	ev.Check("every call ended at a resource limit", limited[ai.PhaseEventQueue] == rounds && limited[ai.PhaseStream] == rounds &&
		limited[ai.PhaseRequest] == rounds, "got %v", limited)
	checkReleased(ev, lw)

	deadline := time.Now().Add(waitDeadline)
	now := runtime.NumGoroutine()
	for now > baseline+goroutineSlack && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
		now = runtime.NumGoroutine()
	}
	ev.Record("goroutines", map[string]int{"baseline": baseline, "after": now, "slack": goroutineSlack})
	ev.Check("goroutines converge to a bound independent of the number of calls", now <= baseline+goroutineSlack,
		"baseline %d, after %d", baseline, now)
}

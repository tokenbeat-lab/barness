package e2e

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Failure inventory precedes the boundary fix: ENOBUFS can hide bytes already
// enqueued by Darwin, so even a transient error cannot safely resume. Clean
// writes preserve exact bytes; every socket fault/cancellation still terminates
// without replaying either bytes or the logical call.
func TestPressureFixtureWriteBackpressure(t *testing.T) {
	for _, mode := range []string{"clean", "transient", "partial", "persistent", "other", "closed", "deadline", "canceled", "unreported", "partial-unreported"} {
		t.Run(mode, func(t *testing.T) {
			ev := pressureCase(t, "E08-pressure-write-"+mode)
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			fault := &pressureFaultListener{Listener: ln, mode: mode, hit: make(chan struct{}), release: make(chan struct{})}
			pw := newPressureWorldWith(t, *validPolicy(), 1, func(aliases map[string]string) *provider.Server {
				return provider.NewWithListener(aliases, fault)
			})
			f, raw := loadTextFixture(t, "text-basic.json")
			ev.Fixture("fixture.json", raw)
			reply := sseReply(t, f, provider.FramingLF)
			reply.TraceWrites = true
			reply.OnRequest = func(r provider.Request) {
				pw.failures.expect(r.Header[http.CanonicalHeaderKey(pressureCallHeader)], reply.Chunks)
			}
			finished := make(chan provider.ReplyOutcome, 1)
			reply.OnFinish = func(o provider.ReplyOutcome) { finished <- o }
			pw.provider.Enqueue(reply)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			if mode == "canceled" {
				go func() {
					defer close(fault.release)
					select {
					case <-fault.hit:
						cancel()
					case <-ctx.Done():
					}
				}()
			}
			started := time.Now()
			res, err := pw.client.WithHooks(pressureDiagnosticHooks()).Complete(ctx, scopeFor(pw.tenants[0], "req-write-"+mode), textTarget(f), textRequest(f), nil)
			ev.Record("result", res)
			traces := pw.failures.integrity()
			ev.Record("stream-integrity", traces)
			ev.Check("no socket fault shifts the scripted byte or frame sequence", len(traces) == 1 && traces[0].FirstDifference == nil, "trace %+v", traces)
			var text strings.Builder
			for _, c := range res.Message.Content {
				if c, ok := c.(ai.Text); ok {
					text.WriteString(c.Text)
				}
			}
			wantFaults := int64(1)
			if mode == "clean" {
				wantFaults = 0
			}
			ev.Check("a socket error is never retried", fault.faults.Load() == wantFaults, "fault count %d", fault.faults.Load())
			if mode == "clean" {
				ev.Check("a clean socket preserves the complete turn", err == nil && res.Message.ErrorMessage == "" && text.String() == "Hello, world!" && traces[0].Complete, "err %v, text %q", err, text.String())
			} else {
				var ae *ai.Error
				want := ai.CodeTransport
				if mode == "canceled" {
					want = ai.CodeCanceled
				}
				ev.Check("unrecoverable socket failure keeps the public failure", errors.As(err, &ae) && ae.Code == want, "err %v", err)
			}
			select {
			case o := <-finished:
				ev.Record("provider-outcome", o)
				ev.Check("fixture completion is bounded", time.Since(started) < 2*time.Second, "elapsed %s", time.Since(started))
				if mode != "clean" {
					ev.Check("the failed socket boundary is recorded once", o.ConnectionID > 0 && len(o.Writes) == 1 && o.WriteError != "", "outcome %+v", o)
				}
				if mode == "persistent" || mode == "transient" || mode == "partial" || mode == "unreported" || mode == "partial-unreported" {
					ev.Check("buffer exhaustion remains an explicit fixture failure", strings.Contains(o.WriteError, "no buffer space available"), "outcome %+v", o)
				}
				if mode == "deadline" {
					ev.Check("write deadline is honored during buffer wait", strings.Contains(o.WriteError, "i/o timeout"), "outcome %+v", o)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("fixture did not terminate")
			}
			ev.Check("socket waiting never replays an AI request", len(pw.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "requests %d, attempts %d", len(pw.provider.Requests()), len(res.Metadata.Attempts))
			checkPressureReleased(ev, pw)
		})
	}
}

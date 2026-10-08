package e2e

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Failure modes are fixed before implementing socket backpressure: temporary
// and partial ENOBUFS must preserve exact bytes; persistent ENOBUFS, another
// errno, close, write deadline and caller cancellation must still terminate.
func TestPressureFixtureWriteBackpressure(t *testing.T) {
	for _, mode := range []string{"transient", "partial", "persistent", "other", "closed", "deadline", "canceled"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "E08-pressure-write-"+mode)
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			fault := &pressureFaultListener{Listener: ln, mode: mode, hit: make(chan struct{})}
			pw := newPressureWorldWith(t, *validPolicy(), 1, func(aliases map[string]string) *provider.Server {
				return provider.NewWithListener(aliases, fault)
			})
			f, raw := loadTextFixture(t, "text-basic.json")
			ev.Fixture("fixture.json", raw)
			reply := sseReply(t, f, provider.FramingLF)
			finished := make(chan provider.ReplyOutcome, 1)
			reply.OnFinish = func(o provider.ReplyOutcome) { finished <- o }
			pw.provider.Enqueue(reply)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			if mode == "canceled" {
				go func() {
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
			var text strings.Builder
			for _, c := range res.Message.Content {
				if c, ok := c.(ai.Text); ok {
					text.WriteString(c.Text)
				}
			}
			ev.Check("the socket fault was exercised", fault.faults.Load() > 0, "fault count %d", fault.faults.Load())
			if mode == "transient" || mode == "partial" {
				ev.Check("temporary socket pressure preserves the complete turn", err == nil && res.Message.ErrorMessage == "" && text.String() == "Hello, world!", "err %v, text %q", err, text.String())
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
				if mode == "transient" || mode == "partial" {
					ev.Check("recovered socket pressure is observable", o.WriteBufferWaits > 0 && o.WriteError == "", "outcome %+v", o)
				}
				if mode == "other" {
					ev.Check("other errors are never retried", o.WriteBufferWaits == 0 && o.WriteError != "", "outcome %+v", o)
				}
				if mode == "persistent" {
					ev.Check("persistent buffer exhaustion still fails", o.WriteBufferWaits > 0 && strings.Contains(o.WriteError, "no buffer space available"), "outcome %+v", o)
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

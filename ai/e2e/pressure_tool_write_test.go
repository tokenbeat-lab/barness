package e2e

import (
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Failure inventory: the naturally captured final tool write can enqueue an
// unreported prefix before ENOBUFS. Replaying 256 KiB damages the next JSON
// envelope; three 128 KiB prefixes inflate valid argument JSON beyond 512 KiB.
// Neither may become a successful fixture, a parser error or a limit error
// after a socket failure. The full 32-call design load remains a separate gate.
func TestPressureToolWriteIntegrity(t *testing.T) {
	for _, pattern := range []struct {
		name   string
		prefix int
		faults int64
	}{
		{"clean", 0, 0},
		{"json-envelope", 256 << 10, 1},
		{"argument-limit", 128 << 10, 3},
	} {
		t.Run(pattern.name, func(t *testing.T) {
			ev := run.Case(t, "E08-pressure-tool-write-"+pattern.name)
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			fault := &toolWriteListener{Listener: ln, prefix: pattern.prefix, limit: pattern.faults}
			pw := newPressureWorldWith(t, *hostintegration.CloudInteractivePolicy(), 1, func(aliases map[string]string) *provider.Server {
				return provider.NewWithListener(aliases, fault)
			})
			load := cloudInteractiveDesignLoad()
			reply := pressureReply(coalesce(pressureToolCall(t, load.ToolJSONBytes), 64))
			reply.TraceWrites = true
			reply.OnRequest = func(r provider.Request) {
				pw.failures.expect(r.Header[http.CanonicalHeaderKey(pressureCallHeader)], reply.Chunks)
			}
			finished := make(chan provider.ReplyOutcome, 1)
			reply.OnFinish = func(o provider.ReplyOutcome) { finished <- o }
			pw.provider.Enqueue(reply)
			res, err := pw.client.WithHooks(pressureDiagnosticHooks()).Complete(ctxFor(t), scopeFor(pw.tenants[0], "req-tool-"+pattern.name), ai.Target{BindingID: "primary", ModelID: pressureModel}, pressureRequest(load), nil)
			ev.Record("result", res)
			traces := pw.failures.integrity()
			ev.Record("stream-integrity", traces)
			ev.Check("a socket failure never shifts the tool byte sequence", len(traces) == 1 && traces[0].FirstDifference == nil, "%+v", traces)
			if pattern.faults == 0 {
				var args string
				for _, c := range res.Message.Content {
					if c, ok := c.(ai.ToolCall); ok {
						args = c.RawArguments
					}
				}
				ev.Check("clean final tool frames remain complete", err == nil && len(args) == load.ToolJSONBytes && traces[0].Complete, "err %v, args %d", err, len(args))
			} else {
				var ae *ai.Error
				ev.Check("the native write failure stays a transport failure", errors.As(err, &ae) && ae.Code == ai.CodeTransport, "err %v", err)
			}
			wantFaults := min(pattern.faults, 1)
			ev.Check("the unreported suffix is never resumed", fault.faults.Load() == wantFaults, "%d faults", fault.faults.Load())
			ev.Check("the logical tool call is never replayed", len(pw.provider.Requests()) == 1 && len(res.Metadata.Attempts) == 1, "requests %d, attempts %d", len(pw.provider.Requests()), len(res.Metadata.Attempts))
			select {
			case o := <-finished:
				ev.Record("provider-outcome", o)
				if pattern.faults != 0 {
					ev.Check("the terminal write error remains visible", o.ConnectionID > 0 && len(o.Writes) == 1 && o.WriteError != "" && o.Chunks < len(reply.Chunks), "%+v", o)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("tool fixture did not finish")
			}
			checkPressureReleased(ev, pw)
		})
	}
}

// This fixture deliberately uses the recorded large terminal-write geometry;
// the existing text fault fixture targets small early writes independently.
type toolWriteListener struct {
	net.Listener
	prefix int
	limit  int64
	faults atomic.Int64
}

func (l *toolWriteListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &toolWriteConn{Conn: c, listener: l}, nil
}

type toolWriteConn struct {
	net.Conn
	listener *toolWriteListener
}

func (c *toolWriteConn) Write(p []byte) (int, error) {
	l := c.listener
	if len(p) <= 512<<10 || l.faults.Load() >= l.limit {
		return c.Conn.Write(p)
	}
	l.faults.Add(1)
	if _, err := c.Conn.Write(p[:min(l.prefix, len(p))]); err != nil {
		return 0, err
	}
	return 0, syscall.ENOBUFS
}

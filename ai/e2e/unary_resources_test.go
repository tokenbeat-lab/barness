package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Counts bytes delivered by the real HTTP body, before the library's limits.
// The close gauge verifies the lease still belongs to the parsed response.
type unaryMeter struct {
	next        http.RoundTripper
	read        atomic.Int64
	heldAtClose atomic.Bool
	held        func() bool
}

func (m *unaryMeter) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := m.next.RoundTrip(req)
	if res != nil && res.Body != nil {
		res.Body = &unaryMeterBody{ReadCloser: res.Body, m: m}
	}
	return res, err
}

type unaryMeterBody struct {
	io.ReadCloser
	m *unaryMeter
}

func (b *unaryMeterBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.m.read.Add(int64(n))
	return n, err
}
func (b *unaryMeterBody) Close() error {
	b.m.heldAtClose.Store(b.m.held())
	return b.ReadCloser.Close()
}

func TestUnaryExactBodyBudgets(t *testing.T) {
	for _, name := range []string{"output-exact", "output-over", "error-exact", "error-over", "bad-json", "missing-answer", "invalid-distribution", "read-broken"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E08-unary-budget-"+name)
			var meter *unaryMeter
			u := newUnaryWorld(t, func(c *ai.Config) {
				meter = &unaryMeter{next: c.Transport, held: func() bool { return c.Probe.Permits() == 1 }}
				c.Transport = meter
				c.Policy.MaxFrameBytes = 16
				c.Policy.MaxOutputBytes = 1024
				c.Policy.MaxToolJSONBytes = 16
				c.Policy.MaxErrorBodyBytes = 32
			})
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
			reply := u.sc.replies(t)[0]
			budget := int64(1024)
			want := ai.Code("")
			phase := ai.PhaseResponse
			switch name {
			case "output-exact", "output-over":
				padding := 1024 - len(reply.Chunks[0])
				if name == "output-over" {
					padding += 10000
					want = ai.CodeResourceLimit
				}
				reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", padding))...)
			case "error-exact", "error-over":
				budget = 32
				phase = ai.PhaseRequest
				want = ai.CodeUpstreamError
				n := 32
				if name == "error-over" {
					n = 10000
					want = ai.CodeResourceLimit
				}
				reply = fixtureReply{Status: 529, Header: map[string]string{"x-should-retry": "false", "retry-after": "1"}, Body: strings.Repeat("x", n)}.script(t, classifierProtocol, "")
				if name == "error-over" {
					delete(reply.Header, "x-should-retry")
				}
			case "bad-json":
				reply.Chunks = [][]byte{[]byte("{")}
				want = ai.CodeProtocol
			case "missing-answer":
				reply.Chunks = [][]byte{[]byte(`{"answers":{},"usage":{"input_tokens":296,"output_tokens":20}}`)}
				want = ai.CodeProtocol
			case "invalid-distribution":
				reply.Chunks = [][]byte{[]byte(strings.Replace(reply.Script, `"track":0.9`, `"track":0.1`, 1))}
				want = ai.CodeProtocol
			case "read-broken":
				reply.End = provider.EndAbort
				want = ai.CodeTransport
			}
			reply.Script = joinChunks(reply.Chunks)
			enqueue(ev, u.world, reply)
			res, err := u.call(t, ctxFor(t), "req-budget-"+name, ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				ev.Check("permit held through response metadata", u.gauges.Permits() == 1 && u.adm.Held() == 1 && u.bodies.Open() == 1, "permits=%d bodies=%d", u.gauges.Permits(), u.bodies.Open())
				return nil
			}})
			if want == "" {
				ev.Check("exact unary output above frame succeeds", err == nil && len(res.Answers) == 1, "got %v", err)
				ev.Record("result", res)
			} else {
				checkUnaryFailure(ev, res, err, want, phase, ai.StopReasonError)
			}
			if name == "missing-answer" || name == "invalid-distribution" {
				ev.Check("invalid answers retain usage", res.Usage.Input == 296 && res.Metadata.Attempts[0].UsageReporting == ai.UsageComplete, "got %+v", res)
			}
			ev.Record("body-byte-count", meter.read.Load())
			if strings.HasSuffix(name, "-over") {
				ev.Check("stop reading at one probe byte", meter.read.Load() == budget+1, "got %d want %d", meter.read.Load(), budget+1)
			}
			if strings.HasSuffix(name, "-exact") {
				ev.Check("inclusive byte budget", meter.read.Load() == budget, "got %d want %d", meter.read.Load(), budget)
			}
			ev.Check("lease held until body closed after parse", meter.heldAtClose.Load(), "lease released early")
			ev.Check("no replay after success or error limit", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
			enqueue(ev, u.world, u.sc.replies(t)[0])
			next, e := u.call(t, ctxFor(t), "req-budget-next-"+name, ai.Hooks{})
			ev.Check("next call gets permit", e == nil && len(next.Answers) == 1, "got %v", e)
			u.released(ev)
		})
	}
}

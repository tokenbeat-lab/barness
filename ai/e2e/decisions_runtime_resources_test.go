package e2e

import (
	"context"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsRuntimeBodyBudgets(t *testing.T) {
	for _, name := range []string{"output-exact", "output-over", "error-exact", "error-over", "deep-json", "all-question-probabilities"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD43-runtime-"+name)
			var meter *unaryMeter
			u := runtimeDecisionsWorld(t, func(c *ai.Config) {
				meter = &unaryMeter{next: c.Transport, held: func() bool { return c.Probe.Permits() == 1 }}
				c.Transport = meter
				c.Policy.MaxFrameBytes, c.Policy.MaxToolJSONBytes, c.Policy.MaxQueuedEventBytes = 16, 16, 16
				c.Policy.MaxOutputBytes, c.Policy.MaxErrorBodyBytes = 1024, 32
				if name == "deep-json" {
					c.Policy.MaxOutputBytes = 32768
				}
			})
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
			reply := runtimeDecisionsReply(t)
			budget := int64(1024)
			if name == "deep-json" {
				budget = 32768
			}
			code, phase := ai.Code(""), ai.PhaseResponse
			switch name {
			case "output-exact", "output-over":
				padding := 1024 - len(reply.Chunks[0])
				if name == "output-over" {
					padding += 10000
					code = ai.CodeResourceLimit
				}
				reply.Chunks[0] = append(reply.Chunks[0], []byte(strings.Repeat(" ", padding))...)
			case "error-exact", "error-over":
				budget, code, phase = 32, ai.CodeUpstreamError, ai.PhaseRequest
				n := 32
				if name == "error-over" {
					n = 10000
					code = ai.CodeResourceLimit
				}
				reply = fixtureReply{Status: 503, Header: map[string]string{"x-request-id": "up-budget", "x-should-retry": "false"}, Body: strings.Repeat("x", n)}.script(t, decisionsProtocol, "")
				// An oversized retryable error must terminate on its own budget.
				if name == "error-over" {
					delete(reply.Header, "x-should-retry")
				}
			case "deep-json":
				reply.Chunks = [][]byte{[]byte(`{"answers":[` + decisionsAnswer + `],"vendor":` + strings.Repeat("[", 10001) + "0" + strings.Repeat("]", 10001) + `, ` + decisionsFullUsage + `}`)}
				code = ai.CodeProtocol
			case "all-question-probabilities":
				// The full envelope includes three requested answers and additive
				// per-probability metadata; it cannot bypass the unary byte limit.
				answer := strings.Replace(decisionsAnswer, `"probability":0.9`, `"probability":0.9,"metadata":"`+strings.Repeat("x", 10000)+`"`, 1)
				reply.Chunks = [][]byte{[]byte(`{"answers":[` + answer + `,` + decisionsScore + `,` + decisionsPredicate + `],` + decisionsFullUsage + `}`)}
				code = ai.CodeResourceLimit
			}
			reply.Script = joinChunks(reply.Chunks)
			enqueue(ev, u.world, reply)
			req := decisionsInput()
			if name == "all-question-probabilities" {
				req = decisionsMixedInput()
			}
			res, err := runtimeDecisionsCall(ctxFor(t), u, "req-budget-"+name, req, ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				ev.Check("response retains body and permit", u.bodies.Open() == 1 && u.adm.Held() == 1 && u.gauges.Permits() == 1, "lease lost")
				return nil
			}})
			ev.Record("result", res)
			recordRuntimeConfig(ev, u)
			ev.Record("limits", map[string]any{"output": budget, "error": 32, "frame": 16, "eventBytes": 16, "jsonDepth": 10000, "deliveredBodyBytes": meter.read.Load()})
			if code == "" {
				ev.Check("exact unary body exceeds SSE allowance", err == nil && len(res.Answers) == 1, "got %v", err)
			} else {
				checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
			}
			if strings.HasSuffix(name, "-over") || name == "all-question-probabilities" {
				ev.Check("read only one probe byte", meter.read.Load() == budget+1, "got %d", meter.read.Load())
			}
			if strings.HasSuffix(name, "-exact") {
				ev.Check("inclusive budget", meter.read.Load() == budget, "got %d", meter.read.Load())
			}
			ev.Check("permit spans body closure", meter.heldAtClose.Load(), "permit released early")
			ev.Check("limits cannot replay", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
			enqueue(ev, u.world, runtimeDecisionsReply(t))
			next, e := runtimeDecisionsCall(ctxFor(t), u, "req-budget-recovery-"+name, decisionsInput(), ai.Hooks{})
			ev.Check("next call can acquire", e == nil && len(next.Answers) == 1, "got %v", e)
			u.records(ev, next, e)
			u.released(ev)
		})
	}
}

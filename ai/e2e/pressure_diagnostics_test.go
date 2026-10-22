package e2e

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// The diagnostic boundary must see body errors hidden by the public, sanitized
// transport/stream terminal. Only this synthetic loopback world records them.
func TestLoopbackTransportDiagnostics(t *testing.T) {
	ev := run.Case(t, "E08-pressure-body-error-diagnostics")
	pw := newPressureWorld(t, *validPolicy(), 1)
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("fixture.json", raw)
	reply := sseReply(t, f, provider.FramingLF)
	reply.Chunks = reply.Chunks[:len(reply.Chunks)-1]
	reply.End = provider.EndAbort
	reply.OnRequest = func(r provider.Request) {
		pw.failures.expect(r.Header[http.CanonicalHeaderKey(pressureCallHeader)], lfChunks(t, f.Events))
	}
	finished := make(chan provider.ReplyOutcome, 1)
	reply.OnFinish = func(o provider.ReplyOutcome) { finished <- o }
	pw.provider.Enqueue(reply)
	res, err := pw.client.WithHooks(pressureDiagnosticHooks()).Complete(ctxFor(t), scopeFor(pw.tenants[0], "req-pressure-diagnostic"), textTarget(f), textRequest(f), nil)
	var ae *ai.Error
	ev.Check("the public terminal reports the same stream interruption", errors.As(err, &ae) && ae.Code == ai.CodeTransport && ae.Phase == ai.PhaseStream, "err %v", err)
	ev.Record("result", res)
	errs := pw.failures.list()
	ev.Record("transport-errors", errs)
	diagnosticJSON, _ := json.Marshal(errs)
	ev.Check("read diagnostics identify the synthetic call", strings.Contains(string(diagnosticJSON), "req-pressure-diagnostic"), "call identity missing")
	ev.Check("the synthetic evidence retains the underlying read error", len(errs) == 1 && errs[0].Error == "unexpected EOF" && errs[0].Phase == "body-read", "diagnostic count %d", len(errs))
	ev.Check("the read was interrupted before cancellation or body close", len(errs) == 1 && errs[0].ContextError == "" && !errs[0].BodyClosed, "diagnostic count %d", len(errs))
	traces := pw.failures.integrity()
	ev.Record("stream-integrity", traces)
	ev.Check("an interrupted exact prefix identifies the first missing byte and frame", len(traces) == 1 && traces[0].RequestID == "req-pressure-diagnostic" && traces[0].FirstDifference == nil && traces[0].ReadBytes == int64(len(reply.Script)-len(lfChunks(t, f.Events)[len(f.Events)-1])) && traces[0].NextFrame == len(f.Events)-1 && traces[0].NextFrameOffset == 0 && !traces[0].Complete, "traces %+v", traces)
	select {
	case outcome := <-finished:
		ev.Record("provider-outcome", outcome)
		outcomeJSON, _ := json.Marshal(outcome)
		ev.Check("fixture outcome identifies the same synthetic call", strings.Contains(string(outcomeJSON), "req-pressure-diagnostic"), "call identity missing")
		ev.Check("fixture lifecycle records the deliberate abort", outcome.End == provider.EndAbort && outcome.WrittenBytes > 0, "outcome %+v", outcome)
	case <-ctxFor(t).Done():
		t.Fatal("fixture did not finish")
	}
	checkPressureReleased(ev, pw)
}

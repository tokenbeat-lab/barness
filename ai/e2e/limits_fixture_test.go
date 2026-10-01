package e2e

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// limitWorld is tenant A's world under validPolicy as changed by a scenario,
// with a transport that counts unclosed response bodies and a resource probe.
type limitWorld struct {
	*world
	bodies *provider.BodyTracker
	probe  *probe.Probe
}

func newLimitWorld(t *testing.T, policy func(*ai.ResourcePolicy)) limitWorld {
	t.Helper()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	p := probe.New()
	w := newWorldWith(t, func(c *ai.Config) {
		c.Transport, c.Probe = bodies, p
		if policy != nil {
			policy(c.Policy)
		}
	}, tenantA)
	return limitWorld{world: w, bodies: bodies, probe: p}
}

// checkLimited asserts a call ended at a resource limit: StopReason error,
// resource_limit in phase, an error naming the policy field, and on streams
// exactly one error terminal carrying the Result.
func checkLimited(ev *evidence.Case, o outcome, phase ai.Phase, field string) {
	m := o.result.Message
	ev.Check("stop reason is error", m.StopReason == ai.StopReasonError, "got %q", m.StopReason)
	ev.Check("errors.Is matches resource_limit and phase", errors.Is(o.err, &ai.Error{Code: ai.CodeResourceLimit, Phase: phase}),
		"want resource_limit/%s, got %v", phase, o.err)
	var ae *ai.Error
	if errors.As(o.err, &ae) {
		ev.Check("error names the policy limit", strings.Contains(ae.Message, field), "got %q", ae.Message)
		ev.Check("message errorMessage is the error's message", m.ErrorMessage == ae.Message, "message=%q error=%q", m.ErrorMessage, ae.Message)
	}
	ev.Check("message carries its timestamp", m.Timestamp > 0, "got %d", m.Timestamp)
	if !o.streamed {
		return
	}
	got := eventSummary(o.events)
	terminals := 0
	for _, s := range got {
		if strings.HasPrefix(s, "done:") || strings.HasPrefix(s, "error:") {
			terminals++
		}
	}
	ev.Check("stream ends with exactly one error terminal", terminals == 1 && len(got) > 0 && got[len(got)-1] == "error:error", "got %q", got)
	ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
	if e, ok := o.events[len(o.events)-1].(ai.ErrorEvent); ok {
		ev.Check("error event carries the Result's message and error", reflect.DeepEqual(e.Message, m) && error(e.Err) == o.err, "event=%+v", e)
	}
}

// checkReleased asserts that the call left nothing behind: every response
// body is closed and no call is running.
func checkReleased(ev *evidence.Case, lw limitWorld) {
	ev.Record("probe", map[string]int64{"open_bodies": lw.bodies.Open(), "active_calls": lw.probe.ActiveCalls()})
	ev.Check("every response body is closed", lw.bodies.Open() == 0, "%d open", lw.bodies.Open())
	ev.Check("no call is active", lw.probe.ActiveCalls() == 0, "%d active", lw.probe.ActiveCalls())
}

// checkTextSucceeded asserts a call within its limits ended exactly as the
// text fixture expects.
func checkTextSucceeded(ev *evidence.Case, f textFixture, o outcome) {
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	checkFinalMessage(ev, f, o.result.Message)
	if o.streamed {
		got := eventSummary(o.events)
		ev.Check("stream ends with done", len(got) > 0 && got[len(got)-1] == "done:stop", "got %q", got)
	}
}

// boundary is one byte limit run at its boundary on every entry point: set
// to exactly the size the scenario needs, the call behaves as without the
// limit; set one lower (the scenario then needs limit+1), it ends as
// resource_limit in phase.
type boundary struct {
	id    string
	field string // the ResourcePolicy field, named by the error
	phase ai.Phase
	// size is the exact size the scenario needs of the limit on entry e.
	size func(t *testing.T, e outcomeEntry) int64
	set  func(p *ai.ResourcePolicy, n int64)
	// script records the fixture and enqueues the provider's replies.
	script  func(t *testing.T, ev *evidence.Case, lw limitWorld, over bool)
	target  ai.Target
	request ai.Request
	within  func(ev *evidence.Case, lw limitWorld, o outcome)
	over    func(ev *evidence.Case, lw limitWorld, o outcome)
}

func runBoundary(t *testing.T, bc boundary) {
	t.Helper()
	for _, over := range []bool{false, true} {
		name := "at-limit"
		if over {
			name = "over-limit"
		}
		for _, e := range outcomeEntries {
			t.Run(name+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "E08-"+bc.id+"-"+name+"-"+e.name)
				size := bc.size(t, e)
				limit := size
				if over {
					limit = size - 1
				}
				lw := newLimitWorld(t, func(p *ai.ResourcePolicy) { bc.set(p, limit) })
				ev.Record("limit", map[string]any{"field": bc.field, "limit": limit, "scenario_size": size})
				bc.script(t, ev, lw, over)
				scope := textScope("req-e08-" + bc.id + "-" + name + "-" + e.name)
				o := within(ev, "call ends", func() outcome { return e.invoke(ctxFor(t), lw.world, scope, bc.target, bc.request) })
				o.record(ev)
				if over {
					checkLimited(ev, o, bc.phase, bc.field)
					bc.over(ev, lw, o)
				} else {
					bc.within(ev, lw, o)
				}
				checkReleased(ev, lw)
			})
		}
	}
}

// Setters keep the policy valid: a limit another must not exceed lowers that
// one with it.
func setRequestBytes(p *ai.ResourcePolicy, n int64) {
	p.MaxRequestBytes, p.MaxImageBytes = n, min(p.MaxImageBytes, n)
}
func setImageBytes(p *ai.ResourcePolicy, n int64)     { p.MaxImageBytes = n }
func setFrameBytes(p *ai.ResourcePolicy, n int64)     { p.MaxFrameBytes = n }
func setErrorBodyBytes(p *ai.ResourcePolicy, n int64) { p.MaxErrorBodyBytes = n }
func setToolJSONBytes(p *ai.ResourcePolicy, n int64)  { p.MaxToolJSONBytes = n }
func setOutputBytes(p *ai.ResourcePolicy, n int64) {
	p.MaxOutputBytes, p.MaxToolJSONBytes = n, min(p.MaxToolJSONBytes, n)
}

// fixedSize is a size that does not depend on the entry point.
func fixedSize(n int) func(*testing.T, outcomeEntry) int64 {
	return func(*testing.T, outcomeEntry) int64 { return int64(n) }
}

// lfChunks lays events out one SSE frame per chunk (FramingLF), so a
// chunk's length is that frame's size: its bytes through the blank line that
// ends it.
func lfChunks(t *testing.T, events []json.RawMessage) [][]byte {
	t.Helper()
	chunks, err := provider.EncodeSSE(events, provider.FramingLF)
	if err != nil {
		t.Fatal(err)
	}
	return chunks
}

// largestAndTotal are the largest chunk's and all chunks' lengths.
func largestAndTotal(chunks [][]byte) (largest, total int) {
	for _, c := range chunks {
		largest, total = max(largest, len(c)), total+len(c)
	}
	return largest, total
}

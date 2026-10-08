package e2e

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// TestHostIntegrationStreams covers the host example's downstream side (E10
// 下游取消、多流汇合; spec I2, I8): a downstream that disconnects or fails a
// send cancels only its own turn while another tenant's turn on the same
// Client streams on untouched; and streams merged into one channel are
// routed by each event's envelope, never by arrival order.
func TestHostIntegrationStreams(t *testing.T) {
	f, raw := loadIsolationFixture(t)
	for _, p := range f.Protocols {
		for _, how := range []string{"send-fails", "client-disconnects"} {
			t.Run(p.ID+"/"+how, func(t *testing.T) {
				ev := run.Case(t, protocolCase(p)+"-E10-example-host-downstream-"+how)
				ev.Fixture("interleave.json", raw)
				testDownstreamEnds(t, ev, f, p, how)
			})
		}
	}

	t.Run("merge", func(t *testing.T) {
		// Tenant A on Responses and tenant B on Anthropic Messages, their
		// frames written alternately by the Provider, consumed through one
		// merged channel.
		ev := run.Case(t, "E10-example-host-merge")
		ev.Fixture("interleave.json", raw)
		responses, anthropic := f.Protocols[0], f.Protocols[1]
		hw := newHostWorld(t)
		rA, rB := responses.success(t, tenantA), anthropic.success(t, tenantB)
		trace := &wireTrace{}
		lockstep(ev, trace, &rA, &rB)
		hw.enqueueFor(ev, tenantA, rA)
		hw.enqueueFor(ev, tenantB, rB)

		scopeA, scopeB := principalScope(tenantA), principalScope(tenantB)
		ctx := ctxFor(t)
		sA := hw.client.Stream(ctx, scopeA, responses.target(), responses.request(), nil)
		sB := hw.client.Stream(ctx, scopeB, anthropic.target(), anthropic.request(), nil)
		var merged []ai.EventEnvelope
		within(ev, "the merged channel closes", func() struct{} {
			for env := range hostintegration.Merge(ctx, sA, sB) {
				merged = append(merged, env)
			}
			return struct{}{}
		})
		resA, resB := resultWithin(ev, "A's Result", sA), resultWithin(ev, "B's Result", sB)
		ev.Record("merged", envelopesJSON(merged))
		ev.Record("wire_trace", trace.all())
		ev.Check("the Provider interleaved the two streams frame by frame",
			reflect.DeepEqual(trace.all(), alternating(len(rA.Chunks), len(rB.Chunks))), "got %q", trace.all())

		// Route by envelope only, then compare each route with that call's
		// run alone.
		routes := map[string][]ai.EventEnvelope{}
		for _, env := range merged {
			routes[env.Call.RequestID] = append(routes[env.Call.RequestID], env)
		}
		ev.Check("every event routes to one of the two calls", len(routes) == 2 && len(routes[scopeA.RequestID])+len(routes[scopeB.RequestID]) == len(merged),
			"routes %d", len(routes))
		for _, c := range []struct {
			k     tenantKey
			p     isolationProtocol
			scope ai.CallScope
			res   callResult
		}{{tenantA, responses, scopeA, resA}, {tenantB, anthropic, scopeB, resB}} {
			route := routes[c.scope.RequestID]
			want := ai.CallAttribution{Operation: ai.OperationChat, TenantID: c.k.tenant, RequestID: c.scope.RequestID, ActorID: c.scope.ActorID, BindingID: c.p.BindingID,
				Resolved: true, ProviderID: c.p.provider(), API: c.p.api(), ModelID: c.p.Model, AccountScopeID: c.p.account(c.k)}
			ev.Check(c.k.tenant+"'s call succeeds", c.res.err == nil, "err=%v", c.res.err)
			checkEnvelopes(ev, route, c.res.res, want)
			solo := soloOutcome(t, c.p, c.k)
			var events []ai.Event
			for _, env := range route {
				events = append(events, env.Event)
			}
			ev.Check(c.k.tenant+"'s routed events equal its run alone", reflect.DeepEqual(eventSummary(events), eventSummary(solo.events)),
				"got %q\nalone %q", eventSummary(events), eventSummary(solo.events))
			other := f.Markers[tenantB.tenant][0]
			if c.k == tenantB {
				other = f.Markers[tenantA.tenant][0]
			}
			for _, s := range eventSummary(events) {
				ev.Check(c.k.tenant+"'s route holds none of the other's text", !strings.Contains(s, other), "event %q", s)
			}
		}
		checkAdmissionReleased(ev, hw.admissionWorld)
	})
}

// testDownstreamEnds runs A's turn, whose downstream fails or disconnects
// at A's first text delta while A's provider still streams, during B's turn,
// which is held mid-stream until A's turn ended.
func testDownstreamEnds(t *testing.T, ev *evidence.Case, f isolationFixture, p isolationProtocol, how string) {
	hw := newHostWorld(t)
	sessA, sessB := hw.open(t, tenantA), hw.open(t, tenantB)
	rB := p.success(t, tenantB)
	trace := &wireTrace{}
	holding, resume := newCue(), newCue()
	holdAfter(ev, trace, &rB, p.firstMarkedChunk(t, f, tenantB), holding, resume)
	hw.enqueueFor(ev, tenantA, heldThroughMarker(t, f, p)...)
	hw.enqueueFor(ev, tenantB, rB)

	sinkB := &collectSink{}
	doneB := make(chan callResult, 1)
	go func() {
		res, err := hw.host.Turn(ctxFor(t), principal(tenantB), hw.turn(p, sessB, p.User), sinkB)
		doneB <- callResult{res, err}
	}()
	holding.await(ev, "B's turn is held mid-stream")

	ctxA, disconnect := context.WithCancel(ctxFor(t))
	defer disconnect()
	errSend := errors.New("downstream: write: broken pipe")
	sinkA := &collectSink{fail: func(env ai.EventEnvelope) error {
		if _, ok := env.Event.(ai.TextDeltaEvent); !ok {
			return nil
		}
		if how == "send-fails" {
			return errSend
		}
		disconnect() // the client went away; the send itself succeeded
		return nil
	}}
	a := within(ev, "A's turn ends", func() callResult {
		res, err := hw.host.Turn(ctxA, principal(tenantA), hw.turn(p, sessA, p.User), sinkA)
		return callResult{res, err}
	})
	resume.fire()
	b := waitValue(ev, "B's turn ends", doneB)
	ev.Record("outcome-tenant-a", map[string]any{"result": a.res, "error": errString(a.err), "envelopes": envelopesJSON(sinkA.all())})
	ev.Record("outcome-tenant-b", map[string]any{"result": b.res, "error": errString(b.err), "envelopes": envelopesJSON(sinkB.all())})
	ev.Record("wire_trace", trace.all())

	// A's generation is canceled, and its turn is not stored.
	ev.Check("A's message is aborted", a.res.Message.StopReason == ai.StopReasonAborted, "got %q", a.res.Message.StopReason)
	if how == "send-fails" {
		ev.Check("A's turn reports the failed send", errors.Is(a.err, hostintegration.ErrDownstream) && errors.Is(a.err, errSend), "err=%v", a.err)
	} else {
		ev.Check("A's turn reports the cancellation", errors.Is(a.err, &ai.Error{Code: ai.CodeCanceled}), "err=%v", a.err)
	}
	ev.Check("A's session stores nothing", len(hw.store.records(t, tenantA.tenant, sessA)) == 0, "got %d records", len(hw.store.records(t, tenantA.tenant, sessA)))

	// B streamed on untouched and its turn was stored.
	ev.Check("B's turn succeeds", b.err == nil, "err=%v", b.err)
	solo := soloOutcome(t, p, tenantB)
	got, want := b.res.Message, solo.result.Message
	got.Timestamp, want.Timestamp = 0, 0
	ev.Check("B's message equals its run alone", reflect.DeepEqual(got, want), "got %+v\nalone %+v", got, want)
	var eventsB []ai.Event
	for _, env := range sinkB.all() {
		eventsB = append(eventsB, env.Event)
	}
	ev.Check("B's downstream received every event of its run alone", reflect.DeepEqual(eventSummary(eventsB), eventSummary(solo.events)),
		"got %q\nalone %q", eventSummary(eventsB), eventSummary(solo.events))
	checkHostTurn(ev, p, tenantB, b.res, sinkB)
	ev.Check("B's session stores its turn", len(hw.store.records(t, tenantB.tenant, sessB)) == 2, "got %d records", len(hw.store.records(t, tenantB.tenant, sessB)))
	checkAdmissionReleased(ev, hw.admissionWorld)
}

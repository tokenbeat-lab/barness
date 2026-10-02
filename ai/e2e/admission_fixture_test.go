package e2e

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// admissionWorld is tenants A and B under validPolicy as changed by a
// scenario, with a transport that counts unclosed response bodies, a
// resource probe and, when the scenario injects one, the host's admission.
type admissionWorld struct {
	*world
	bodies    *provider.BodyTracker
	probe     *probe.Probe
	admission *host.Admission // nil when the host injects none
	text      textFixture
	transport *http.Transport // under bodies, unless a scenario replaced it
}

func newAdmissionWorld(t *testing.T, policy func(*ai.ResourcePolicy), admission *host.Admission, configure ...func(*ai.Config)) admissionWorld {
	t.Helper()
	transport := provider.LoopbackTransport()
	bodies := provider.TrackBodies(transport)
	p := probe.New()
	w := newWorldWith(t, func(c *ai.Config) {
		c.Transport, c.Probe = bodies, p
		if admission != nil {
			c.Admission = admission
		}
		if policy != nil {
			policy(c.Policy)
		}
		for _, f := range configure {
			f(c)
		}
	}, tenantA, tenantB)
	text, _ := loadTextFixture(t, "text-basic.json")
	return admissionWorld{world: w, bodies: bodies, probe: p, admission: admission, text: text, transport: transport}
}

// scopeFor is k's tenant's call scope for requestID.
func scopeFor(k tenantKey, requestID string) ai.CallScope {
	return ai.CallScope{TenantID: k.tenant, RequestID: requestID, ActorID: "actor-7"}
}

// hold starts a Stream for k whose attempt keeps its permit: the provider
// sends the first frame and holds the connection open. It returns once the
// provider holds it. Holds are started one at a time, so each gets the held
// reply it enqueued.
func (aw admissionWorld) hold(t *testing.T, ev *evidence.Case, ctx context.Context, k tenantKey, requestID string) *ai.Stream {
	t.Helper()
	held := make(chan struct{})
	r := heldReply(200, "text/event-stream", string(lfChunks(t, aw.text.Events)[0]))
	r.OnHold = func() { close(held) }
	aw.provider.Enqueue(r)
	s := aw.client.Stream(ctx, scopeFor(k, requestID), textTarget(aw.text), textRequest(aw.text), nil)
	t.Cleanup(func() { _ = s.Close() })
	waitFor(ev, "provider holds "+requestID, held)
	return s
}

// call invokes e for k on the text fixture.
func (aw admissionWorld) call(ctx context.Context, e outcomeEntry, k tenantKey, requestID string) outcome {
	return e.invoke(ctx, aw.world, scopeFor(k, requestID), textTarget(aw.text), textRequest(aw.text))
}

// eventually polls cond until it holds or the wait deadline passes. It is
// for gauges a scenario cannot be signaled about, such as a call having
// started to wait for a permit.
func eventually(ev *evidence.Case, what string, cond func() bool) {
	deadline := time.Now().Add(waitDeadline)
	for !cond() {
		if time.Now().After(deadline) {
			ev.Check(what, false, "not within %s", waitDeadline)
			ev.T().FailNow()
		}
		time.Sleep(5 * time.Millisecond)
	}
	ev.Check(what, true, "")
}

// checkAdmissionDenied asserts a call refused admission: StopReason error,
// admission_denied in the admission phase, an error naming why, no HTTP
// attempt recorded, and on streams exactly one error terminal.
func checkAdmissionDenied(ev *evidence.Case, o outcome, reason string) {
	checkRejected(ev, o, ai.CodeAdmissionDenied, ai.PhaseAdmission)
	m := o.result.Message
	ev.Check("stop reason is error", m.StopReason == ai.StopReasonError, "got %q", m.StopReason)
	ev.Check("error says why", strings.Contains(m.ErrorMessage, reason), "got %q want %q in it", m.ErrorMessage, reason)
	ev.Check("no HTTP attempt is recorded", len(o.result.Metadata.Attempts) == 0, "got %+v", o.result.Metadata.Attempts)
	ev.Check("call identity resolved", o.result.Metadata.Resolved, "metadata=%+v", o.result.Metadata)
}

// checkAdmissionReleased asserts nothing is left behind: no permit held or
// waited for, by the built-in limits or the host's admission, no open
// response body and no running call.
func checkAdmissionReleased(ev *evidence.Case, aw admissionWorld) {
	eventually(ev, "every call ended", func() bool { return aw.probe.ActiveCalls() == 0 })
	gauges := map[string]int64{"permits": aw.probe.Permits(), "waiters": aw.probe.AdmissionWaiters(),
		"open_bodies": aw.bodies.Open(), "active_calls": aw.probe.ActiveCalls()}
	if aw.admission != nil {
		gauges["host_permits"] = int64(aw.admission.Held())
	}
	ev.Record("probe", gauges)
	ev.Check("no built-in permit is held", aw.probe.Permits() == 0, "%d held", aw.probe.Permits())
	ev.Check("no attempt waits for a permit", aw.probe.AdmissionWaiters() == 0, "%d waiting", aw.probe.AdmissionWaiters())
	ev.Check("every response body is closed", aw.bodies.Open() == 0, "%d open", aw.bodies.Open())
	if aw.admission != nil {
		ev.Check("every host permit is released", aw.admission.Held() == 0, "%d held", aw.admission.Held())
	}
}

// errHostAdmission is the injected admission's own refusal.
var errHostAdmission = errors.New("host admission: account quota exhausted for acct-tenant-a")

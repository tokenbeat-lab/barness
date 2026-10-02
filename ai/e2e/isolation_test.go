package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// isolationScenario is what happens to tenant A's call while tenant B's
// same-named call on the same Client and transport is in flight.
type isolationScenario struct {
	name string
	// lockstep: both calls succeed with their frames interleaved one by one.
	// Otherwise B is held mid-stream, holding its own text, from before A's
	// call starts until after it ended.
	lockstep bool
	// arrange changes the world before any call (A's retry policy, A's
	// credential); nil changes nothing.
	arrange func(iw isolationWorld)
	// aReplies scripts A's endpoint; nil when A sends nothing.
	aReplies func(t *testing.T, f isolationFixture, p isolationProtocol) []provider.Reply
	// aTimeoutMs is A's own protocol timeout; B never sets one.
	aTimeoutMs int
	// cancelA cancels A's context once A's endpoint holds its stream.
	cancelA bool
	// cancelOnArrival cancels A's context as A's response comes back from
	// the shared transport, before the SDK sees it.
	cancelOnArrival bool
	// aAttempts is how many requests A's call sends.
	aAttempts int
	// checkA asserts how A's call ended; nil when A succeeds and must equal
	// its run alone.
	checkA func(ev *evidence.Case, p isolationProtocol, o outcome)
}

var isolationScenarios = []isolationScenario{
	{
		name: "both-succeed", lockstep: true, aAttempts: 1,
		aReplies: func(t *testing.T, _ isolationFixture, p isolationProtocol) []provider.Reply {
			return []provider.Reply{p.success(t, tenantA)}
		},
	},
	{
		name: "a-unauthorized", aAttempts: 1,
		aReplies: func(_ *testing.T, _ isolationFixture, p isolationProtocol) []provider.Reply {
			return []provider.Reply{p.script(tenantA).Unauthorized.reply()}
		},
		checkA: func(ev *evidence.Case, p isolationProtocol, o outcome) {
			checkAEnded(ev, o, ai.StopReasonError, ai.CodeUpstreamAuth, ai.PhaseRequest)
			// Gemini sends no vendor request id header.
			requestID := "req_alpha_401"
			if p.gemini() {
				requestID = ""
			}
			var e *ai.Error
			ev.Check("A's error is its own provider's 401", errors.As(o.err, &e) && e.HTTPStatus == 401 && e.ProviderRequestID == requestID,
				"got %+v", e)
		},
	},
	{
		name: "a-canceled", cancelA: true, aAttempts: 1,
		aReplies: heldThroughMarker, checkA: checkACanceled,
	},
	{
		// The SDKs drop a response unclosed when the caller's context has
		// ended by the time it arrives; the call must still close it.
		name: "a-canceled-on-arrival", cancelOnArrival: true, aAttempts: 1,
		aReplies: heldThroughMarker, checkA: checkACanceled,
	},
	{
		// A's own timeoutMs, on the shared Client, ends A's call while its
		// endpoint withholds the response; B sets none and is not cut off.
		name: "a-timed-out", aTimeoutMs: int(short.Milliseconds()), aAttempts: 1,
		aReplies: func(*testing.T, isolationFixture, isolationProtocol) []provider.Reply {
			return []provider.Reply{{End: provider.EndSilent}}
		},
		checkA: func(ev *evidence.Case, _ isolationProtocol, o outcome) {
			checkAEnded(ev, o, ai.StopReasonError, ai.CodeDeadlineExceeded, ai.PhaseRequest)
		},
	},
	{
		name: "a-retried", aAttempts: 2,
		arrange: func(iw isolationWorld) {
			for _, id := range []string{"primary", "claude", "gemini"} {
				b := iw.host.Binding(tenantA.tenant, id)
				b.Retry = ai.RetryPolicy{MaxRetries: 1}
				iw.host.PutBinding(b)
			}
		},
		aReplies: func(t *testing.T, _ isolationFixture, p isolationProtocol) []provider.Reply {
			return []provider.Reply{p.script(tenantA).Throttled.reply(), p.success(t, tenantA)}
		},
	},
	{
		// A's credential is gone; neither B's key nor the polluted
		// environment stands in for it.
		name: "a-credential-missing", aAttempts: 0,
		arrange: func(iw isolationWorld) {
			iw.host.DeleteCredential(tenantA.tenant, "cred-"+tenantA.tenant)
			iw.host.DeleteCredential(tenantA.tenant, "cred-"+tenantA.tenant+"-anthropic")
			iw.host.DeleteCredential(tenantA.tenant, "cred-"+tenantA.tenant+"-google")
		},
		checkA: func(ev *evidence.Case, _ isolationProtocol, o outcome) {
			checkRejected(ev, o, ai.CodeCredentialUnavailable, ai.PhaseCredential)
			checkUnresolved(ev, o.result, scopeFor(tenantA, o.result.Metadata.RequestID))
		},
	},
}

// TestTenantIsolation is E06 (spec T03, T05, T06, T08; User Stories 5, 30):
// tenants A and B call same-named bindings and the same model with, on
// every entry that takes one, the same session id (Anthropic's and Gemini's
// full options have none), on one Client and one transport, each binding pointing at its
// tenant's own endpoint. Provider and
// transport barriers force the calls to interleave: frame by frame when
// both succeed, or with B held mid-stream across the whole of A's call
// while A is refused (401), canceled (while streaming, or as its response
// arrives), timed out by its own timeoutMs, retried or left without a
// credential. On Responses, Anthropic Messages and the Gemini Developer API
// (which has no timeoutMs to time A out with), through every entry
// point:
//
//   - each endpoint receives only its tenant's key, exactly what that
//     tenant sends running alone;
//   - B's events, message, usage, cost, metadata and attempts equal B's run
//     alone, hold nothing of A's, and A's hold nothing of B's;
//   - Result, observations and admission name each call's own tenant,
//     request and account;
//   - nothing falls back to the other tenant or the polluted environment,
//     and every permit, body and call is released.
//
// Repeat under the race detector with
// `go test -race -count=20 -run TestTenantIsolation ./ai/e2e`.
func TestTenantIsolation(t *testing.T) {
	decoy := polluteEnvironment(t)
	f, raw := loadIsolationFixture(t)
	for _, p := range f.Protocols {
		for _, sc := range isolationScenarios {
			if sc.aTimeoutMs != 0 && !p.protocolTimeout() {
				// A's own timeoutMs cannot end A's call on a protocol
				// without one (TestGeminiTimeouts asserts it bounds
				// nothing there).
				continue
			}
			for _, e := range isolationEntries {
				t.Run(p.ID+"/"+sc.name+"/"+e.name, func(t *testing.T) {
					ev := run.Case(t, "E06-"+p.ID+"-"+sc.name+"-"+e.name)
					ev.Fixture("interleave.json", raw)
					runIsolation(t, ev, f, p, sc, e)
				})
			}
		}
	}
	ev := run.Case(t, "E06-no-environment-fallback")
	ev.Record("decoy_requests", decoy.Requests())
	ev.Check("environment endpoints and proxies received nothing", len(decoy.Requests()) == 0, "got %d", len(decoy.Requests()))
}

func runIsolation(t *testing.T, ev *evidence.Case, f isolationFixture, p isolationProtocol, sc isolationScenario, e isolationEntry) {
	// RequestIDs are globally unique (ADR-0001), so each tenant has its own.
	requestID := "req-e06-" + sc.name + "-" + e.name
	scopeA, scopeB := scopeFor(tenantA, requestID+"-a"), scopeFor(tenantB, requestID+"-b")

	// Each tenant alone, on a Client of its own, is the reference.
	soloB, soloBSent := runSolo(t, p, e, tenantB, scopeB, nil, []provider.Reply{p.success(t, tenantB)}, 0)
	var soloA outcome
	if sc.checkA == nil {
		soloA, _ = runSolo(t, p, e, tenantA, scopeA, sc.arrange, sc.aReplies(t, f, p), sc.aTimeoutMs)
	}

	iw := newIsolationWorld(t)
	if sc.arrange != nil {
		sc.arrange(iw)
	}
	trace := &wireTrace{}
	bReply := p.success(t, tenantB)
	var aReplies []provider.Reply
	if sc.aReplies != nil {
		aReplies = sc.aReplies(t, f, p)
	}
	ctxA, cancelA := context.WithCancel(ctxFor(t))
	defer cancelA()
	invoke := func(ctx context.Context, scope ai.CallScope, timeoutMs int) <-chan outcome {
		done := make(chan outcome, 1)
		go func() { done <- e.invoke(ctx, iw.world, p, scope, timeoutMs) }()
		return done
	}

	var oA, oB outcome
	if sc.lockstep {
		lockstep(ev, trace, &aReplies[0], &bReply)
		iw.enqueueFor(ev, tenantA, aReplies...)
		iw.enqueueFor(ev, tenantB, bReply)
		doneA := invoke(ctxA, scopeA, sc.aTimeoutMs)
		doneB := invoke(ctxFor(t), scopeB, 0)
		oA = waitValue(ev, "tenant A's call ends", doneA)
		oB = waitValue(ev, "tenant B's call ends", doneB)
		want := alternating(len(aReplies[0].Chunks), len(bReply.Chunks))
		ev.Check("the provider wrote the tenants' frames alternately", reflect.DeepEqual(trace.all(), want), "got %q want %q", trace.all(), want)
	} else {
		bHolding, aEnded, aHolding := newCue(), newCue(), newCue()
		heldAt := p.firstMarkedChunk(t, f, tenantB)
		holdAfter(ev, trace, &bReply, heldAt, bHolding, aEnded)
		if sc.cancelOnArrival {
			iw.arrivals.onArrival(func(r *http.Request) {
				if strings.HasPrefix(r.URL.Path, tenantPrefix(tenantA)+"/") {
					trace.add("A:arrived")
					cancelA()
				}
			})
		}
		if sc.cancelA {
			aReplies[len(aReplies)-1].OnHold = func() { trace.add("A:held"); aHolding.fire() }
		}
		iw.enqueueFor(ev, tenantB, bReply)
		iw.enqueueFor(ev, tenantA, aReplies...)
		doneB := invoke(ctxFor(t), scopeB, 0)
		waitFor(ev, "tenant B is mid-stream", bHolding.ch)
		doneA := invoke(ctxA, scopeA, sc.aTimeoutMs)
		if sc.cancelA {
			waitFor(ev, "tenant A's stream is held", aHolding.ch)
			cancelA()
		}
		oA = waitValue(ev, "tenant A's call ends", doneA)
		trace.add("A:ended")
		aEnded.fire()
		oB = waitValue(ev, "tenant B's call ends", doneB)
		steps := trace.all()
		before, _, _ := strings.Cut(strings.Join(steps, " "), "A:ended")
		ev.Check("B was held mid-stream across A's whole call", strings.Count(before, "B:") == heldAt+1 && len(steps) == len(bReply.Chunks)+strings.Count(before, "A:")+1,
			"got %q, B held after frame %d", steps, heldAt)
	}
	ev.Record("wire_trace", trace.all())
	recordTenant(ev, tenantA, oA)
	recordTenant(ev, tenantB, oB)
	ev.Record("requests", iw.provider.Requests())

	// B is untouched by whatever happened to A.
	checkMatchesSolo(ev, tenantB, oB, soloB)
	checkSentAsSolo(ev, tenantB, iw.requestsAt(tenantB), soloBSent)
	checkNoForeignContent(ev, tenantB, oB, f.Markers[tenantA.tenant])
	// A ended as it should, with nothing of B's.
	if sc.checkA == nil {
		checkMatchesSolo(ev, tenantA, oA, soloA)
	} else {
		sc.checkA(ev, p, oA)
	}
	checkNoForeignContent(ev, tenantA, oA, f.Markers[tenantB.tenant])

	checkWire(ev, iw, p, sc)
	for _, c := range []struct {
		k        tenantKey
		scope    ai.CallScope
		o        outcome
		attempts int
	}{{tenantA, scopeA, oA, sc.aAttempts}, {tenantB, scopeB, oB, 1}} {
		checkAttribution(ev, p, c.k, c.scope, c.o)
		checkObserved(ev, iw, p, c.k, c.scope, c.o, c.attempts)
		checkAdmitted(ev, iw, p, c.k, c.scope, c.attempts)
	}
	checkAdmissionReleased(ev, iw.admissionWorld)
}

// runSolo runs k's call alone on a Client of its own and returns its
// outcome and what its endpoint received.
func runSolo(t *testing.T, p isolationProtocol, e isolationEntry, k tenantKey, scope ai.CallScope,
	arrange func(isolationWorld), replies []provider.Reply, timeoutMs int) (outcome, []provider.Request) {
	t.Helper()
	iw := newIsolationWorld(t)
	if arrange != nil {
		arrange(iw)
	}
	iw.provider.EnqueueAt(tenantPrefix(k)+"/", replies...)
	o := e.invoke(ctxFor(t), iw.world, p, scope, timeoutMs)
	if o.err != nil {
		t.Fatalf("%s alone: %v", k.tenant, o.err)
	}
	return o, iw.requestsAt(k)
}

// checkWire asserts which requests reached the Provider: each tenant's at
// its own endpoint path with its own key, as many as its call attempts,
// and none anywhere else. On Responses, the session id both tenants pass
// yields a different cache key and affinity header for each.
func checkWire(ev *evidence.Case, iw isolationWorld, p isolationProtocol, sc isolationScenario) {
	a, b := iw.requestsAt(tenantA), iw.requestsAt(tenantB)
	all := iw.provider.Requests()
	ev.Check("every request reached one tenant's endpoint", len(all) == len(a)+len(b), "got %d, A %d, B %d", len(all), len(a), len(b))
	ev.Check("tenant A's call sent its attempts", len(a) == sc.aAttempts, "got %d want %d", len(a), sc.aAttempts)
	ev.Check("tenant B's call sent one request", len(b) == 1, "got %d", len(b))
	for _, c := range []struct {
		k    tenantKey
		reqs []provider.Request
	}{{tenantA, a}, {tenantB, b}} {
		for _, r := range c.reqs {
			ev.Check(c.k.tenant+"'s endpoint received its own key on its own path",
				r.KeyAlias == p.key(c.k).alias && r.Path == p.path(c.k), "got %s with %q", r.Path, r.KeyAlias)
		}
	}
	if p.anthropic() || p.gemini() || len(a) == 0 {
		return
	}
	keyA, keyB := promptCacheKey(ev, a[0]), promptCacheKey(ev, b[0])
	ev.Check("the shared session id yields each tenant its own cache key", keyA != "" && keyB != "" && keyA != keyB,
		"A %q B %q", keyA, keyB)
	ev.Check("the shared session id yields each tenant its own affinity header",
		a[0].Header["Session_id"] != "" && a[0].Header["Session_id"] != b[0].Header["Session_id"],
		"A %q B %q", a[0].Header["Session_id"], b[0].Header["Session_id"])
}

func promptCacheKey(ev *evidence.Case, r provider.Request) string {
	var body struct {
		Key string `json:"prompt_cache_key"`
	}
	if err := json.Unmarshal(r.Body, &body); err != nil {
		ev.Check("request body decodes", false, "%v", err)
	}
	return body.Key
}

// heldThroughMarker scripts A's reply as its frames through the first one
// carrying A's text, then holding the connection open.
func heldThroughMarker(t *testing.T, f isolationFixture, p isolationProtocol) []provider.Reply {
	chunks := p.success(t, tenantA).Chunks[:p.firstMarkedChunk(t, f, tenantA)+1]
	var parts []string
	for _, c := range chunks {
		parts = append(parts, string(c))
	}
	return []provider.Reply{heldReply(200, "text/event-stream", parts...)}
}

func checkACanceled(ev *evidence.Case, _ isolationProtocol, o outcome) {
	m := o.result.Message
	ev.Check("A's call is aborted", m.StopReason == ai.StopReasonAborted, "got %q", m.StopReason)
	ev.Check("A's error is canceled", errors.Is(o.err, &ai.Error{Code: ai.CodeCanceled}), "got %v", o.err)
}

// checkAEnded asserts A's call ended with stop and the classified error.
func checkAEnded(ev *evidence.Case, o outcome, stop ai.StopReason, code ai.Code, phase ai.Phase) {
	m := o.result.Message
	ev.Check("A's stop reason", m.StopReason == stop, "got %q want %q", m.StopReason, stop)
	ev.Check("A's error code and phase", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}), "want %s/%s, got %v", code, phase, o.err)
	ev.Check("A's message carries its error text", m.ErrorMessage != "", "empty")
}

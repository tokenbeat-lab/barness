package e2e

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestProtocolObservability is TestObservability on each scenario protocol
// (P02–P04/E09): success, upstream failure, retry and interruption each
// produce the call/attempt lifecycle attributed to the trusted scope and the
// resolved snapshot of the protocol's binding, with no attempt invented.
func TestProtocolObservability(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolObservability)
}

func testProtocolObservability(t *testing.T, s *protocolSuite) {
	a := loadScenarioText(t, s.proto)
	rf, rraw := loadFixture(t, s.proto, "retry.json")
	id := s.requestID("e09-")
	binding := s.proto.binding

	t.Run("success", func(t *testing.T) {
		for _, e := range s.entries() {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E09-observe-success-"+e.name))
				ev.Fixture("text.json", a.raw)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				enqueue(ev, w, a.reply(t))
				scope := textScope(id + "ok-" + e.name)
				o := e.invoke(ctxFor(t), w, scope, a.target(), a.request(t))
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				checkCallRecords(ev, rec.awaitCall(ev, scope.RequestID), scope, binding, o, 1)
				m := o.result.Metadata
				ev.Check("resolved snapshot is attributed", m.Resolved && m.AccountScopeID == s.proto.account &&
					m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.ProviderID == s.proto.provider &&
					m.API == s.proto.api && m.ModelID == a.sc.Model, "metadata=%+v", m)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		ev := run.Case(t, s.caseID("E09-observe-failure"))
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		enqueue(ev, w, s.authFailure)
		scope := textScope(id + "failure")
		res, err := w.client.Complete(ctxFor(t), scope, a.target(), a.request(t), s.options)
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("upstream_auth", errors.Is(err, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, binding, o, 1)
		at := obs[2].Attempt
		ev.Check("attempt_finished classifies the failed attempt", at != nil && at.HTTPStatus == s.authFailure.Status && at.Code == ai.CodeUpstreamAuth &&
			at.ProviderRequestID == s.authFailureRequestID, "attempt=%+v", at)
	})

	t.Run("retry", func(t *testing.T) {
		ev := run.Case(t, s.caseID("E09-observe-retry"))
		sc := scenarioByID(t, rf, "retry-429-then-success")
		ev.Fixture("retry.json", rraw)
		rec := newRecorder()
		w := observedWorld(t, rec, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) }, tenantA)
		w.updateBindingOf(tenantA, binding, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: sc.MaxRetries} })
		enqueue(ev, w, sc.replies(t)...)
		scope := textScope(id + "retry")
		res, err := w.client.Complete(ctxFor(t), scope, sc.target(), sc.request(t), s.options)
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, binding, o, 2)
		first := obs[2].Attempt
		ev.Check("the retried attempt records its classification", first != nil && first.HTTPStatus == 429 && first.Code == ai.CodeRateLimited,
			"attempt=%+v", first)
	})

	t.Run("interrupted", func(t *testing.T) {
		for _, how := range []string{"cancel", "close"} {
			t.Run(how, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E09-observe-interrupted-"+how))
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				held := make(chan struct{})
				r := heldReply(200, "text/event-stream", string(a.chunks(t)[0]))
				r.OnHold = func() { close(held) }
				enqueue(ev, w, r)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				scope := textScope(id + "interrupted-" + how)
				st := w.client.Stream(ctx, scope, a.target(), a.request(t), nil)
				waitFor(ev, "provider holds the stream", held)
				if how == "cancel" {
					cancel()
				} else {
					within(ev, "Close returns", func() error { return st.Close() })
				}
				cr := resultWithin(ev, "call ends", st)
				o := outcome{result: cr.res, err: cr.err}
				o.record(ev)
				ev.Check("call ends canceled", errors.Is(cr.err, &ai.Error{Code: ai.CodeCanceled}), "err=%v", cr.err)
				obs := rec.awaitCall(ev, scope.RequestID)
				checkCallRecords(ev, obs, scope, binding, o, 1)
				ev.Check("call_finished reports the abort", obs[len(obs)-1].StopReason == ai.StopReasonAborted, "got %q", obs[len(obs)-1].StopReason)
				_ = st.Close()
			})
		}
	})
}

// TestProtocolObservabilityRedaction is TestObservabilityRedaction on each
// scenario protocol (P02–P04/E09): with the key sent in the protocol's key
// header, content, a tool call and signed or reasoning blocks in play, and
// a provider error body echoing the key, nothing sensitive reaches an
// observation, the logs or the error text, the echoed key is redacted from
// ErrorMessage, and neither the key nor the raw TenantID reaches the
// provider anywhere but the key header.
func TestProtocolObservabilityRedaction(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolObservabilityRedaction)
}

func testProtocolObservabilityRedaction(t *testing.T, s *protocolSuite) {
	ev := run.Case(t, s.caseID("E09-redaction"))
	key := s.proto.key(tenantA)
	rd := s.redaction
	tf, traw := loadFixture(t, s.proto, "text.json")
	ev.Fixture("text.json", traw)
	sc := scenarioByID(t, tf, "interleaved")
	logs := captureLogs(t)
	rec := newRecorder()
	w := observedWorld(t, rec, nil, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	okID := s.requestID("e09-redaction")
	turn, err := w.client.Complete(ctxFor(t), textScope(okID), sc.target(), sc.request(t), s.options)
	ev.Record("turn", turn)
	ev.Check("the turn succeeds", err == nil, "err=%v", err)
	for _, k := range rd.kept {
		ev.Check("the authorized Result keeps its signed and reasoning content", strings.Contains(string(mustMarshal(t, turn.Message)), k), "missing %q", k)
	}

	status, echo := rd.echo(key.secret)
	enqueue(ev, w, provider.Reply{Status: status, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(echo)}, Script: echo})
	failedID := fmt.Sprintf("%s-%d", okID, status)
	failed, failedErr := w.client.Complete(ctxFor(t), textScope(failedID), sc.target(), sc.request(t), s.options)
	ev.Record("failed", failed)
	ev.Check("the echoing error is classified", errors.Is(failedErr, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", failedErr)

	var obs []ai.Observation
	for _, id := range []string{okID, failedID} {
		obs = append(obs, rec.awaitCall(ev, id)...)
	}
	sensitive := append([]string{key.secret}, rd.sensitive...)
	for name, text := range map[string]string{"observations": observationText(t, obs), "logs": logs.String()} {
		for _, v := range sensitive {
			ev.Check("no sensitive value in "+name, !strings.Contains(text, v), "found %q", v)
		}
	}
	// The vendor's own echo of the prompt stays in ErrorMessage as in pi;
	// only key-shaped text is redacted (ADR-0009).
	errText := errString(failedErr) + failed.Message.ErrorMessage
	for _, v := range append([]string{key.secret}, rd.notInError...) {
		ev.Check("no secret or library-added content in the error text", !strings.Contains(errText, v), "found %q in %q", v, errText)
	}
	ev.Check("the echoed key is redacted", strings.Contains(failed.Message.ErrorMessage, "[REDACTED]"), "got %q", failed.Message.ErrorMessage)
	ev.Record("echoed-error-message", failed.Message.ErrorMessage)
	for _, r := range w.provider.Requests() {
		// The capture shows the key as its alias, which names the tenant;
		// the key itself is what was sent there.
		header := maps.Clone(r.Header)
		delete(header, s.keyHeader)
		sent := string(r.Body) + fmt.Sprint(header) + r.Query
		ev.Check("the raw TenantID never reaches the provider", !strings.Contains(sent, tenantA.tenant), "request %s %s", r.Method, r.Path)
		others := ""
		for _, other := range protocolSuites {
			if other.keyHeader != s.keyHeader {
				others += r.Header[other.keyHeader]
			}
		}
		query, _ := url.ParseQuery(r.Query)
		ev.Check("the key travels only as "+strings.ToLower(s.keyHeader)+", never in the URL", r.KeyAlias == key.alias && others == "" && !query.Has("key"),
			"got %q / other key headers %q / query %q", r.KeyAlias, others, r.Query)
	}
}

// TestProtocolPreflightRejections is E07 on each scenario protocol's
// binding (P02–P04/E07): each refusal ends before any inference request
// through the full and simple entries, with an unresolved identity.
func TestProtocolPreflightRejections(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolPreflightRejections)
}

func testProtocolPreflightRejections(t *testing.T, s *protocolSuite) {
	a := loadScenarioText(t, s.proto)
	binding := s.proto.binding
	cases := append([]rejectCase{
		{id: "options-of-another-api", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(s.foreignOptions, s.foreignOptions)},
		{id: "actor-not-permitted", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.host.RestrictActors(tenantA.tenant, binding, "actor-admin") }},
		{id: "binding-disabled", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.updateBindingOf(tenantA, binding, func(b *ai.Binding) { b.Enabled = false }) }},
		{id: "credential-missing", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) { w.host.DeleteCredential(tenantA.tenant, s.credential(tenantA).CredentialID) }},
		{id: "credential-revoked", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) {
				c := s.credential(tenantA)
				c.Active = false
				w.host.PutCredential(c)
			}},
	}, s.rejections...)
	for _, c := range cases {
		entries := c.entries
		if entries == nil {
			entries = s.entries()
		}
		for _, e := range entries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, s.caseID("E07-reject-"+c.id+"-"+e.name))
				decoy := polluteEnvironment(t)
				w := newWorld(t, tenantA)
				if c.arrange != nil {
					c.arrange(w)
				}
				target := a.target()
				if c.target != (ai.Target{}) {
					target = c.target
				}
				scope := textScope(s.requestID("e07-" + c.id + "-" + e.name))
				o := within(ev, "call ends", func() outcome { return e.invoke(ctxFor(t), w, scope, target, a.request(t)) })
				o.record(ev)
				checkRejected(ev, o, c.code, c.phase)
				checkUnresolved(ev, o.result, scope)
				checkNoSecrets(ev, o)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				ev.Check("the environment received nothing", len(decoy.Requests()) == 0, "got %d", len(decoy.Requests()))
			})
		}
	}
}

// TestProtocolResourceRelease is TestResourceRelease on each scenario
// protocol (P02–P04/E08): concurrent Streams never read or closed, each
// ending at a limit while the provider still holds its connection, leave no
// open body, running call or growing goroutines.
func TestProtocolResourceRelease(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolResourceRelease)
}

func testProtocolResourceRelease(t *testing.T, s *protocolSuite) {
	ev := run.Case(t, s.caseID("E08-resource-release"))
	long, _ := s.longOutput(t, longOutputDeltas)
	target := loadScenarioText(t, s.proto).target()
	lw := newLimitWorld(t, func(p *ai.ResourcePolicy) {
		p.MaxQueuedEvents = 20
		setFrameBytes(p, 2048)
		setErrorBodyBytes(p, 1024)
	})
	baseline := runtime.NumGoroutine()
	const rounds = 8
	replies := []provider.Reply{
		long,
		heldReply(200, "text/event-stream", s.held.frameOpen+strings.Repeat("x", 2100)),
		heldReply(500, "application/json", strings.Repeat("e", 1100)),
	}
	n := 0
	for range rounds {
		lw.provider.Enqueue(replies...)
		n += len(replies)
	}
	outcomes := make([]outcome, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			st := lw.client.Stream(ctxFor(t), textScope(s.requestID(fmt.Sprintf("release-%d", i))), target,
				ai.Request{Messages: []ai.Message{ai.UserText("hi")}}, nil)
			res, err := st.Result()
			outcomes[i] = outcome{result: res, err: err}
		})
	}
	waitFor(ev, "every call ends", doneWhen(wg.Wait))
	limited := map[ai.Phase]int{}
	for _, o := range outcomes {
		var ae *ai.Error
		if errors.As(o.err, &ae) && ae.Code == ai.CodeResourceLimit {
			limited[ae.Phase]++
		}
	}
	ev.Record("limited_by_phase", limited)
	ev.Check("every call ended at a resource limit", limited[ai.PhaseEventQueue] == rounds && limited[ai.PhaseStream] == rounds &&
		limited[ai.PhaseRequest] == rounds, "got %v", limited)
	checkReleased(ev, lw)
	checkGoroutinesSettle(ev, baseline)
}

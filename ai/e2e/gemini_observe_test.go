package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Gemini's data and its deliberate
// differences (ADR-0012); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestGeminiObservability is TestObservability on the Gemini Developer API
// (P03/E09): success, upstream failure, retry and interruption each produce
// the call/attempt lifecycle attributed to the trusted scope and the
// resolved Google snapshot, with no attempt invented.
func TestGeminiObservability(t *testing.T) {
	a := loadScenarioText(t, geminiProtocol)
	rf, rraw := loadFixture(t, geminiProtocol, "retry.json")

	t.Run("success", func(t *testing.T) {
		for _, e := range geminiEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E09-observe-success-"+e.name)
				ev.Fixture("text.json", a.raw)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				enqueue(ev, w, a.reply(t))
				scope := textScope("req-p03-e09-ok-" + e.name)
				o := e.invoke(ctxFor(t), w, scope, a.target(), a.request(t))
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				checkCallRecords(ev, rec.awaitCall(ev, scope.RequestID), scope, "gemini", o, 1)
				m := o.result.Metadata
				ev.Check("resolved Google snapshot is attributed", m.Resolved && m.AccountScopeID == "acct-tenant-a-google" &&
					m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.ProviderID == ai.ProviderGoogle &&
					m.API == ai.APIGoogleGenerativeAI && m.ModelID == a.sc.Model, "metadata=%+v", m)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		ev := run.Case(t, "P03-E09-observe-failure")
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		body := `{"error":{"code":403,"message":"Permission denied.","status":"PERMISSION_DENIED"}}`
		enqueue(ev, w, provider.Reply{Status: 403, Header: map[string]string{"Content-Type": "application/json"},
			Chunks: [][]byte{[]byte(body)}, Script: body})
		scope := textScope("req-p03-e09-failure")
		res, err := w.client.Complete(ctxFor(t), scope, a.target(), a.request(t), ai.GeminiOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("upstream_auth", errors.Is(err, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "gemini", o, 1)
		at := obs[2].Attempt
		ev.Check("attempt_finished classifies the failed attempt", at != nil && at.HTTPStatus == 403 && at.Code == ai.CodeUpstreamAuth,
			"attempt=%+v", at)
	})

	t.Run("retry", func(t *testing.T) {
		ev := run.Case(t, "P03-E09-observe-retry")
		sc := scenarioByID(t, rf, "retry-429-then-success")
		ev.Fixture("retry.json", rraw)
		rec := newRecorder()
		w := observedWorld(t, rec, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) }, tenantA)
		w.updateBindingOf(tenantA, "gemini", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: sc.MaxRetries} })
		enqueue(ev, w, sc.replies(t)...)
		scope := textScope("req-p03-e09-retry")
		res, err := w.client.Complete(ctxFor(t), scope, sc.target(), sc.request(t), ai.GeminiOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "gemini", o, 2)
		first := obs[2].Attempt
		ev.Check("the retried attempt records its classification", first != nil && first.HTTPStatus == 429 && first.Code == ai.CodeRateLimited,
			"attempt=%+v", first)
	})

	t.Run("interrupted", func(t *testing.T) {
		for _, how := range []string{"cancel", "close"} {
			t.Run(how, func(t *testing.T) {
				ev := run.Case(t, "P03-E09-observe-interrupted-"+how)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				held := make(chan struct{})
				r := heldReply(200, "text/event-stream", string(a.chunks(t)[0]))
				r.OnHold = func() { close(held) }
				enqueue(ev, w, r)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				scope := textScope("req-p03-e09-interrupted-" + how)
				s := w.client.Stream(ctx, scope, a.target(), a.request(t), nil)
				waitFor(ev, "provider holds the stream", held)
				if how == "cancel" {
					cancel()
				} else {
					within(ev, "Close returns", func() error { return s.Close() })
				}
				cr := resultWithin(ev, "call ends", s)
				o := outcome{result: cr.res, err: cr.err}
				o.record(ev)
				ev.Check("call ends canceled", errors.Is(cr.err, &ai.Error{Code: ai.CodeCanceled}), "err=%v", cr.err)
				obs := rec.awaitCall(ev, scope.RequestID)
				checkCallRecords(ev, obs, scope, "gemini", o, 1)
				ev.Check("call_finished reports the abort", obs[len(obs)-1].StopReason == ai.StopReasonAborted, "got %q", obs[len(obs)-1].StopReason)
				_ = s.Close()
			})
		}
	})
}

// TestGeminiObservabilityRedaction is TestObservabilityRedaction on the
// Gemini Developer API (P03/E09): with the key sent as x-goog-api-key,
// content, a function call and thought signatures in play, and a provider
// error body echoing the key, nothing sensitive reaches an observation, the
// logs or the error text, and neither the key nor the raw TenantID reaches
// the provider anywhere but the key header.
func TestGeminiObservabilityRedaction(t *testing.T) {
	ev := run.Case(t, "P03-E09-redaction")
	tf, traw := loadFixture(t, geminiProtocol, "text.json")
	ev.Fixture("text.json", traw)
	sc := scenarioByID(t, tf, "interleaved")
	logs := captureLogs(t)
	rec := newRecorder()
	w := observedWorld(t, rec, nil, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	turn, err := w.client.Complete(ctxFor(t), textScope("req-p03-e09-redaction"), sc.target(), sc.request(t), ai.GeminiOptions{})
	ev.Record("turn", turn)
	ev.Check("the turn succeeds", err == nil, "err=%v", err)
	ev.Check("the authorized Result keeps its signatures", strings.Contains(string(mustMarshal(t, turn.Message)), "c2lnLWNhbGw="), "")

	echo := fmt.Sprintf(`{"error":{"code":403,"message":"API key %s cannot answer %q","status":"PERMISSION_DENIED"}}`, googleA.secret, "Look up x.")
	enqueue(ev, w, provider.Reply{Status: 403, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(echo)}, Script: echo})
	failed, failedErr := w.client.Complete(ctxFor(t), textScope("req-p03-e09-redaction-403"), sc.target(), sc.request(t), ai.GeminiOptions{})
	ev.Record("failed", failed)
	ev.Check("the echoing error is classified", errors.Is(failedErr, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", failedErr)

	var obs []ai.Observation
	for _, id := range []string{"req-p03-e09-redaction", "req-p03-e09-redaction-403"} {
		obs = append(obs, rec.awaitCall(ev, id)...)
	}
	sensitive := []string{googleA.secret, "Look up x.", `{"q":"x","n":2}`, "c2lnLXRoaW5r", "c2lnLWNhbGw=", "Let me think", "Checking.", "cannot answer"}
	for name, text := range map[string]string{"observations": observationText(t, obs), "logs": logs.String()} {
		for _, s := range sensitive {
			ev.Check("no sensitive value in "+name, !strings.Contains(text, s), "found %q", s)
		}
	}
	errText := errString(failedErr) + failed.Message.ErrorMessage
	for _, s := range []string{googleA.secret, `{"q":"x","n":2}`, "c2lnLWNhbGw="} {
		ev.Check("no secret or library-added content in the error text", !strings.Contains(errText, s), "found %q in %q", s, errText)
	}
	ev.Record("echoed-error-message", failed.Message.ErrorMessage)
	for _, r := range w.provider.Requests() {
		header := maps.Clone(r.Header)
		delete(header, "X-Goog-Api-Key")
		sent := string(r.Body) + fmt.Sprint(header) + r.Query
		ev.Check("the raw TenantID never reaches the provider", !strings.Contains(sent, tenantA.tenant), "request %s %s", r.Method, r.Path)
		ev.Check("the key travels only as x-goog-api-key, never in the URL", r.KeyAlias == googleA.alias && r.Header["Authorization"] == "" &&
			r.Query == "alt=sse", "got %q / %q / %q", r.KeyAlias, r.Header["Authorization"], r.Query)
	}
}

// TestGeminiPreflightRejections is E07 on the gemini binding (P03/E07):
// each refusal ends before any inference request through the full and
// simple entries, with an unresolved identity.
func TestGeminiPreflightRejections(t *testing.T) {
	a := loadScenarioText(t, geminiProtocol)
	cases := []struct {
		id      string
		code    ai.Code
		phase   ai.Phase
		target  ai.Target
		arrange func(w *world)
		entries []outcomeEntry
	}{
		{id: "options-of-another-api", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: []outcomeEntry{
			{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				return drain(w.client.Stream(ctx, scope, target, req, ai.AnthropicOptions{}))
			}},
			{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				res, err := w.client.Complete(ctx, scope, target, req, ai.AnthropicOptions{})
				return outcome{result: res, err: err}
			}},
		}},
		{id: "invalid-options", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: []outcomeEntry{
			{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				return drain(w.client.Stream(ctx, scope, target, req, ai.GeminiOptions{ToolChoice: ai.Value(ai.GeminiToolChoice("required"))}))
			}},
			{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				res, err := w.client.Complete(ctx, scope, target, req, ai.GeminiOptions{Thinking: ai.Value(ai.GeminiThinking{Enabled: true, BudgetTokens: ai.Value(-2)})})
				return outcome{result: res, err: err}
			}},
		}},
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "gemini", ModelID: "gemini-3.1-flash-live-preview"}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "gemini", ModelID: "claude-haiku-4-5"}},
		{id: "actor-not-permitted", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.host.RestrictActors(tenantA.tenant, "gemini", "actor-admin") }},
		{id: "binding-disabled", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.updateBindingOf(tenantA, "gemini", func(b *ai.Binding) { b.Enabled = false }) }},
		{id: "credential-missing", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) { w.host.DeleteCredential(tenantA.tenant, "cred-tenant-a-google") }},
		{id: "credential-revoked", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) {
				c := geminiCredential(tenantA)
				c.Active = false
				w.host.PutCredential(c)
			}},
	}
	for _, c := range cases {
		entries := c.entries
		if entries == nil {
			entries = geminiEntries
		}
		for _, e := range entries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P03-E07-reject-"+c.id+"-"+e.name)
				decoy := polluteEnvironment(t)
				w := newWorld(t, tenantA)
				if c.arrange != nil {
					c.arrange(w)
				}
				target := a.target()
				if c.target != (ai.Target{}) {
					target = c.target
				}
				scope := textScope("req-p03-e07-" + c.id + "-" + e.name)
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

// TestGeminiResourceRelease is TestResourceRelease on the Gemini Developer
// API (P03/E08): concurrent Streams never read or closed, each ending at a
// limit while the provider still holds its connection, leave no open body,
// running call or growing goroutines.
func TestGeminiResourceRelease(t *testing.T) {
	ev := run.Case(t, "P03-E08-resource-release")
	long, _ := geminiLongOutput(t, longOutputDeltas)
	lw := newLimitWorld(t, func(p *ai.ResourcePolicy) {
		p.MaxQueuedEvents = 20
		setFrameBytes(p, 2048)
		setErrorBodyBytes(p, 1024)
	})
	baseline := runtime.NumGoroutine()
	const rounds = 8
	replies := []provider.Reply{
		long[0],
		heldReply(200, "text/event-stream", "data: "+strings.Repeat("x", 2100)),
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
			s := lw.client.Stream(ctxFor(t), textScope(fmt.Sprintf("req-p03-release-%d", i)), ai.Target{BindingID: "gemini", ModelID: "gemini-2.5-flash"},
				ai.Request{Messages: []ai.Message{ai.UserText("hi")}}, nil)
			res, err := s.Result()
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

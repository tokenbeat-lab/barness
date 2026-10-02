package e2e

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

// TestAnthropicObservability is TestObservability on Anthropic Messages
// (P02/E09): success, upstream failure, retry and interruption each produce
// the call/attempt lifecycle attributed to the trusted scope and the
// resolved Anthropic snapshot, with no attempt invented.
func TestAnthropicObservability(t *testing.T) {
	a := loadScenarioText(t, anthropicProtocol)
	rf, rraw := loadFixture(t, anthropicProtocol, "retry.json")

	t.Run("success", func(t *testing.T) {
		for _, e := range anthropicEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P02-E09-observe-success-"+e.name)
				ev.Fixture("text.json", a.raw)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				enqueue(ev, w, a.reply(t))
				scope := textScope("req-p02-e09-ok-" + e.name)
				o := e.invoke(ctxFor(t), w, scope, a.target(), a.request(t))
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				checkCallRecords(ev, rec.awaitCall(ev, scope.RequestID), scope, "claude", o, 1)
				m := o.result.Metadata
				ev.Check("resolved Anthropic snapshot is attributed", m.Resolved && m.AccountScopeID == "acct-tenant-a-anthropic" &&
					m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.ProviderID == ai.ProviderAnthropic &&
					m.API == ai.APIAnthropicMessages && m.ModelID == a.sc.Model, "metadata=%+v", m)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		ev := run.Case(t, "P02-E09-observe-failure")
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		body := `{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`
		enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "request-id": "req_vendor_401"},
			Chunks: [][]byte{[]byte(body)}, Script: body})
		scope := textScope("req-p02-e09-failure")
		res, err := w.client.Complete(ctxFor(t), scope, a.target(), a.request(t), ai.AnthropicOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("upstream_auth", errors.Is(err, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "claude", o, 1)
		at := obs[2].Attempt
		ev.Check("attempt_finished classifies the failed attempt", at != nil && at.HTTPStatus == 401 && at.Code == ai.CodeUpstreamAuth &&
			at.ProviderRequestID == "req_vendor_401", "attempt=%+v", at)
	})

	t.Run("retry", func(t *testing.T) {
		ev := run.Case(t, "P02-E09-observe-retry")
		sc := scenarioByID(t, rf, "retry-429-then-success")
		ev.Fixture("retry.json", rraw)
		rec := newRecorder()
		w := observedWorld(t, rec, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) }, tenantA)
		w.updateBindingOf(tenantA, "claude", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: sc.MaxRetries} })
		enqueue(ev, w, sc.replies(t)...)
		scope := textScope("req-p02-e09-retry")
		res, err := w.client.Complete(ctxFor(t), scope, sc.target(), sc.request(t), ai.AnthropicOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "claude", o, 2)
		first := obs[2].Attempt
		ev.Check("the retried attempt records its classification", first != nil && first.HTTPStatus == 429 && first.Code == ai.CodeRateLimited,
			"attempt=%+v", first)
	})

	t.Run("interrupted", func(t *testing.T) {
		for _, how := range []string{"cancel", "close"} {
			t.Run(how, func(t *testing.T) {
				ev := run.Case(t, "P02-E09-observe-interrupted-"+how)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				held := make(chan struct{})
				r := heldReply(200, "text/event-stream", string(a.chunks(t)[0]))
				r.OnHold = func() { close(held) }
				enqueue(ev, w, r)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				scope := textScope("req-p02-e09-interrupted-" + how)
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
				checkCallRecords(ev, obs, scope, "claude", o, 1)
				ev.Check("call_finished reports the abort", obs[len(obs)-1].StopReason == ai.StopReasonAborted, "got %q", obs[len(obs)-1].StopReason)
				_ = s.Close()
			})
		}
	})
}

// TestAnthropicObservabilityRedaction is TestObservabilityRedaction on
// Anthropic Messages (P02/E09): with the key sent as x-api-key, content, a
// tool call and a signed thinking block in play, and a provider error body
// echoing the key, nothing sensitive reaches an observation, the logs or the
// error text, and the raw TenantID never reaches the provider.
func TestAnthropicObservabilityRedaction(t *testing.T) {
	ev := run.Case(t, "P02-E09-redaction")
	tf, traw := loadFixture(t, anthropicProtocol, "text.json")
	ev.Fixture("text.json", traw)
	sc := scenarioByID(t, tf, "interleaved")
	logs := captureLogs(t)
	rec := newRecorder()
	w := observedWorld(t, rec, nil, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	turn, err := w.client.Complete(ctxFor(t), textScope("req-p02-e09-redaction"), sc.target(), sc.request(t), ai.AnthropicOptions{})
	ev.Record("turn", turn)
	ev.Check("the turn succeeds", err == nil, "err=%v", err)
	ev.Check("the authorized Result keeps its signature", strings.Contains(string(mustMarshal(t, turn.Message)), "sig-abc"), "")

	echo := fmt.Sprintf(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key: %s for %q"}}`, anthropicA.secret, "Look up x.")
	enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(echo)}, Script: echo})
	failed, failedErr := w.client.Complete(ctxFor(t), textScope("req-p02-e09-redaction-401"), sc.target(), sc.request(t), ai.AnthropicOptions{})
	ev.Record("failed", failed)
	ev.Check("the echoing error is classified", errors.Is(failedErr, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", failedErr)

	var obs []ai.Observation
	for _, id := range []string{"req-p02-e09-redaction", "req-p02-e09-redaction-401"} {
		obs = append(obs, rec.awaitCall(ev, id)...)
	}
	sensitive := []string{anthropicA.secret, "Look up x.", `{"q":"x"}`, "sig-abc", "Let me think", "Hi there", "invalid x-api-key"}
	for name, text := range map[string]string{"observations": observationText(t, obs), "logs": logs.String()} {
		for _, s := range sensitive {
			ev.Check("no sensitive value in "+name, !strings.Contains(text, s), "found %q", s)
		}
	}
	errText := errString(failedErr) + failed.Message.ErrorMessage
	for _, s := range []string{anthropicA.secret, `{"q":"x"}`, "sig-abc"} {
		ev.Check("no secret or library-added content in the error text", !strings.Contains(errText, s), "found %q in %q", s, errText)
	}
	ev.Record("echoed-error-message", failed.Message.ErrorMessage)
	for _, r := range w.provider.Requests() {
		header := maps.Clone(r.Header)
		delete(header, "X-Api-Key")
		sent := string(r.Body) + fmt.Sprint(header)
		ev.Check("the raw TenantID never reaches the provider", !strings.Contains(sent, tenantA.tenant), "request %s %s", r.Method, r.Path)
		ev.Check("the key travels only as x-api-key", r.KeyAlias == anthropicA.alias && r.Header["Authorization"] == "", "got %q / %q", r.KeyAlias, r.Header["Authorization"])
	}
}

// TestAnthropicPreflightRejections is E07 on the claude binding (P02/E07):
// each refusal ends before any inference request through the full and simple
// entries, with an unresolved identity.
func TestAnthropicPreflightRejections(t *testing.T) {
	a := loadScenarioText(t, anthropicProtocol)
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
				return drain(w.client.Stream(ctx, scope, target, req, ai.ResponsesOptions{}))
			}},
			{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				res, err := w.client.Complete(ctx, scope, target, req, ai.ResponsesOptions{})
				return outcome{result: res, err: err}
			}},
		}},
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "claude", ModelID: "claude-fable-5"}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "claude", ModelID: "gpt-4.1-mini"}},
		{id: "actor-not-permitted", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.host.RestrictActors(tenantA.tenant, "claude", "actor-admin") }},
		{id: "binding-disabled", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.updateBindingOf(tenantA, "claude", func(b *ai.Binding) { b.Enabled = false }) }},
		{id: "credential-missing", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) { w.host.DeleteCredential(tenantA.tenant, "cred-tenant-a-anthropic") }},
		{id: "credential-revoked", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) {
				c := anthropicCredential(tenantA)
				c.Active = false
				w.host.PutCredential(c)
			}},
	}
	for _, c := range cases {
		entries := c.entries
		if entries == nil {
			entries = anthropicEntries
		}
		for _, e := range entries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P02-E07-reject-"+c.id+"-"+e.name)
				decoy := polluteEnvironment(t)
				w := newWorld(t, tenantA)
				if c.arrange != nil {
					c.arrange(w)
				}
				target := a.target()
				if c.target != (ai.Target{}) {
					target = c.target
				}
				scope := textScope("req-p02-e07-" + c.id + "-" + e.name)
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

// TestAnthropicResourceRelease is TestResourceRelease on Anthropic Messages
// (P02/E08): concurrent Streams never read or closed, each ending at a
// limit while the provider still holds its connection, leave no open body,
// running call or growing goroutines.
func TestAnthropicResourceRelease(t *testing.T) {
	ev := run.Case(t, "P02-E08-resource-release")
	events, _ := anthropicLongOutput(t, longOutputDeltas)
	lw := newLimitWorld(t, func(p *ai.ResourcePolicy) {
		p.MaxQueuedEvents = 20
		setFrameBytes(p, 2048)
		setErrorBodyBytes(p, 1024)
	})
	baseline := runtime.NumGoroutine()
	const rounds = 8
	replies := []provider.Reply{
		sseEvents(t, events, provider.FramingLF),
		heldReply(200, "text/event-stream", "event: content_block_delta\ndata: "+strings.Repeat("x", 2100)),
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
			s := lw.client.Stream(ctxFor(t), textScope(fmt.Sprintf("req-p02-release-%d", i)), ai.Target{BindingID: "claude", ModelID: "claude-haiku-4-5"},
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

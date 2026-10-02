package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Chat's data and its differences
// (ADR-0013); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestChatObservability is TestObservability on OpenAI Chat Completions
// (P04/E09): success, upstream failure, retry and interruption each produce
// the call/attempt lifecycle attributed to the trusted scope and the
// resolved OpenAI snapshot, with no attempt invented.
func TestChatObservability(t *testing.T) {
	a := loadScenarioText(t, chatProtocol)
	rf, rraw := loadFixture(t, chatProtocol, "retry.json")

	t.Run("success", func(t *testing.T) {
		for _, e := range chatEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P04-E09-observe-success-"+e.name)
				ev.Fixture("text.json", a.raw)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				enqueue(ev, w, a.reply(t))
				scope := textScope("req-p04-e09-ok-" + e.name)
				o := e.invoke(ctxFor(t), w, scope, a.target(), a.request(t))
				o.record(ev)
				checkScenarioTextSucceeded(ev, a, o)
				checkCallRecords(ev, rec.awaitCall(ev, scope.RequestID), scope, "chat", o, 1)
				m := o.result.Metadata
				ev.Check("resolved OpenAI Chat snapshot is attributed", m.Resolved && m.AccountScopeID == "acct-tenant-a" &&
					m.BindingVersion == "b1" && m.CredentialVersion == "v1" && m.ProviderID == ai.ProviderOpenAI &&
					m.API == ai.APIOpenAICompletions && m.ModelID == a.sc.Model, "metadata=%+v", m)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		ev := run.Case(t, "P04-E09-observe-failure")
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		body := `{"error":{"message":"Incorrect API key provided.","type":"invalid_request_error","param":null,"code":"invalid_api_key"}}`
		enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "req_vendor_401"},
			Chunks: [][]byte{[]byte(body)}, Script: body})
		scope := textScope("req-p04-e09-failure")
		res, err := w.client.Complete(ctxFor(t), scope, a.target(), a.request(t), ai.ChatOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("upstream_auth", errors.Is(err, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "chat", o, 1)
		at := obs[2].Attempt
		ev.Check("attempt_finished classifies the failed attempt", at != nil && at.HTTPStatus == 401 && at.Code == ai.CodeUpstreamAuth &&
			at.ProviderRequestID == "req_vendor_401", "attempt=%+v", at)
	})

	t.Run("retry", func(t *testing.T) {
		ev := run.Case(t, "P04-E09-observe-retry")
		sc := scenarioByID(t, rf, "retry-429-then-success")
		ev.Fixture("retry.json", rraw)
		rec := newRecorder()
		w := observedWorld(t, rec, func(c *ai.Config) { c.Clock = clock.NewFake(time.Unix(1790000000, 0), 0) }, tenantA)
		w.updateBindingOf(tenantA, "chat", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: sc.MaxRetries} })
		enqueue(ev, w, sc.replies(t)...)
		scope := textScope("req-p04-e09-retry")
		res, err := w.client.Complete(ctxFor(t), scope, sc.target(), sc.request(t), ai.ChatOptions{})
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "chat", o, 2)
		first := obs[2].Attempt
		ev.Check("the retried attempt records its classification", first != nil && first.HTTPStatus == 429 && first.Code == ai.CodeRateLimited,
			"attempt=%+v", first)
	})

	t.Run("interrupted", func(t *testing.T) {
		for _, how := range []string{"cancel", "close"} {
			t.Run(how, func(t *testing.T) {
				ev := run.Case(t, "P04-E09-observe-interrupted-"+how)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				held := make(chan struct{})
				r := heldReply(200, "text/event-stream", string(a.chunks(t)[0]))
				r.OnHold = func() { close(held) }
				enqueue(ev, w, r)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				scope := textScope("req-p04-e09-interrupted-" + how)
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
				checkCallRecords(ev, obs, scope, "chat", o, 1)
				ev.Check("call_finished reports the abort", obs[len(obs)-1].StopReason == ai.StopReasonAborted, "got %q", obs[len(obs)-1].StopReason)
				_ = s.Close()
			})
		}
	})
}

// TestChatObservabilityRedaction is TestObservabilityRedaction on OpenAI
// Chat Completions (P04/E09): with the key sent as a bearer Authorization,
// content, reasoning and a tool call in play, and a provider error body
// echoing the key and its header, nothing sensitive reaches an
// observation, the logs or the error text, and the raw TenantID never
// reaches the provider.
func TestChatObservabilityRedaction(t *testing.T) {
	ev := run.Case(t, "P04-E09-redaction")
	tf, traw := loadFixture(t, chatProtocol, "text.json")
	ev.Fixture("text.json", traw)
	sc := scenarioByID(t, tf, "interleaved")
	logs := captureLogs(t)
	rec := newRecorder()
	w := observedWorld(t, rec, nil, tenantA)
	enqueue(ev, w, sc.replies(t)...)
	turn, err := w.client.Complete(ctxFor(t), textScope("req-p04-e09-redaction"), sc.target(), sc.request(t), ai.ChatOptions{})
	ev.Record("turn", turn)
	ev.Check("the turn succeeds", err == nil, "err=%v", err)
	ev.Check("the authorized Result keeps its reasoning and call", strings.Contains(string(mustMarshal(t, turn.Message)), "Let me think.") &&
		strings.Contains(string(mustMarshal(t, turn.Message)), "call_1"), "")

	echo := fmt.Sprintf(`{"error":{"message":"Incorrect API key provided: %s. Received header Authorization: Bearer %s for input %q","type":"invalid_request_error","code":"invalid_api_key"}}`,
		tenantA.secret, tenantA.secret, "Look up x.")
	enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(echo)}, Script: echo})
	failed, failedErr := w.client.Complete(ctxFor(t), textScope("req-p04-e09-redaction-401"), sc.target(), sc.request(t), ai.ChatOptions{})
	ev.Record("failed", failed)
	ev.Check("the echoing error is classified", errors.Is(failedErr, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", failedErr)

	var obs []ai.Observation
	for _, id := range []string{"req-p04-e09-redaction", "req-p04-e09-redaction-401"} {
		obs = append(obs, rec.awaitCall(ev, id)...)
	}
	sensitive := []string{tenantA.secret, "Bearer", "Look up x.", `{"q":"x","n":2}`, `\"q\"`, "Let me think", "Still sure", "Checking.", "Incorrect API key"}
	for name, text := range map[string]string{"observations": observationText(t, obs), "logs": logs.String()} {
		for _, s := range sensitive {
			ev.Check("no sensitive value in "+name, !strings.Contains(text, s), "found %q", s)
		}
	}
	// As on Responses, the vendor's own echo of the prompt stays in
	// ErrorMessage as in pi; only key-shaped text is redacted (ADR-0009).
	errText := errString(failedErr) + failed.Message.ErrorMessage
	for _, s := range []string{tenantA.secret, "Bearer sk-", `{"q":"x","n":2}`, "Let me think"} {
		ev.Check("no secret or library-added content in the error text", !strings.Contains(errText, s), "found %q in %q", s, errText)
	}
	ev.Check("the echoed key is redacted", strings.Contains(failed.Message.ErrorMessage, "[REDACTED]"), "got %q", failed.Message.ErrorMessage)
	ev.Record("echoed-error-message", failed.Message.ErrorMessage)
	for _, r := range w.provider.Requests() {
		// The capture shows the bearer key as its alias, which names the
		// tenant; the key itself is what was sent there.
		header := maps.Clone(r.Header)
		delete(header, "Authorization")
		sent := string(r.Body) + fmt.Sprint(header)
		ev.Check("the raw TenantID never reaches the provider", !strings.Contains(sent, tenantA.tenant), "request %s %s", r.Method, r.Path)
		ev.Check("the key travels only as the bearer Authorization", r.KeyAlias == tenantA.alias && r.Header["X-Api-Key"] == "",
			"got %q / %q", r.KeyAlias, r.Header["X-Api-Key"])
	}
}

// TestChatPreflightRejections is E07 on the chat binding (P04/E07): each
// refusal ends before any inference request through the full and simple
// entries, with an unresolved identity.
func TestChatPreflightRejections(t *testing.T) {
	a := loadScenarioText(t, chatProtocol)
	fullEntries := func(first, second ai.ChatOptions) []outcomeEntry {
		return []outcomeEntry{
			{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				return drain(w.client.Stream(ctx, scope, target, req, first))
			}},
			{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
				res, err := w.client.Complete(ctx, scope, target, req, second)
				return outcome{result: res, err: err}
			}},
		}
	}
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
				res, err := w.client.Complete(ctx, scope, target, req, ai.AnthropicOptions{})
				return outcome{result: res, err: err}
			}},
		}},
		{id: "invalid-options", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(
			ai.ChatOptions{ReasoningEffort: ai.ThinkingLevel("extreme")},
			ai.ChatOptions{ToolChoice: ai.Value(ai.ChatToolChoice{Mode: "sometimes"})})},
		{id: "reserved-sampling-params", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability, entries: fullEntries(
			ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"store": json.RawMessage(`true`)}},
			ai.ChatOptions{SamplingParams: map[string]json.RawMessage{"web_search_options": json.RawMessage(`{}`)}})},
		{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			arrange: func(w *world) {
				w.updateBindingOf(tenantA, "chat", func(b *ai.Binding) {
					b.AllowedModels = slices.DeleteFunc(slices.Clone(b.AllowedModels), func(id string) bool { return id == "gpt-4.1-mini" })
				})
			}},
		{id: "model-of-another-api", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "chat", ModelID: "gpt-5-pro"}},
		{id: "model-of-another-provider", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
			target: ai.Target{BindingID: "chat", ModelID: "claude-haiku-4-5"}},
		{id: "actor-not-permitted", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.host.RestrictActors(tenantA.tenant, "chat", "actor-admin") }},
		{id: "binding-disabled", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
			arrange: func(w *world) { w.updateBindingOf(tenantA, "chat", func(b *ai.Binding) { b.Enabled = false }) }},
		{id: "credential-missing", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) { w.host.DeleteCredential(tenantA.tenant, "cred-tenant-a") }},
		{id: "credential-revoked", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
			arrange: func(w *world) {
				c := primaryCredential(tenantA, "v1")
				c.Active = false
				w.host.PutCredential(c)
			}},
	}
	for _, c := range cases {
		entries := c.entries
		if entries == nil {
			entries = chatEntries
		}
		for _, e := range entries {
			t.Run(c.id+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P04-E07-reject-"+c.id+"-"+e.name)
				decoy := polluteEnvironment(t)
				w := newWorld(t, tenantA)
				if c.arrange != nil {
					c.arrange(w)
				}
				target := a.target()
				if c.target != (ai.Target{}) {
					target = c.target
				}
				scope := textScope("req-p04-e07-" + c.id + "-" + e.name)
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

// TestChatResourceRelease is TestResourceRelease on OpenAI Chat Completions
// (P04/E08): concurrent Streams never read or closed, each ending at a
// limit while the provider still holds its connection, leave no open body,
// running call or growing goroutines.
func TestChatResourceRelease(t *testing.T) {
	ev := run.Case(t, "P04-E08-resource-release")
	events, _ := chatLongOutput(t, longOutputDeltas)
	lw := newLimitWorld(t, func(p *ai.ResourcePolicy) {
		p.MaxQueuedEvents = 20
		setFrameBytes(p, 2048)
		setErrorBodyBytes(p, 1024)
	})
	baseline := runtime.NumGoroutine()
	const rounds = 8
	replies := []provider.Reply{
		chatEvents(t, events, provider.FramingLF),
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
			s := lw.client.Stream(ctxFor(t), textScope(fmt.Sprintf("req-p04-release-%d", i)), chatTarget,
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

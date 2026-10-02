package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestObservability covers E09's tracing half (spec I9, User Stories 37–38):
// an Observer receives CallStarted, AttemptStarted, AttemptFinished and
// CallFinished for every path — success, preflight refusal, admission
// refusal, failure, retry, interruption and Close — attributed to the
// trusted scope and the resolved snapshot, with no attempt invented for a
// call that sent nothing. Delivery is asynchronous and bounded: a slow,
// failing or panicking Observer neither blocks nor changes a call, and what
// it lost is countable.
func TestObservability(t *testing.T) {
	text, textRaw := loadTextFixture(t, "text-basic.json")

	t.Run("success", func(t *testing.T) {
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E09-observe-success-"+e.name)
				ev.Fixture("text-basic.json", textRaw)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				enqueue(ev, w, sseReply(t, text, provider.FramingLF))
				scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-e09-ok-" + e.name, ActorID: "actor-7", JobID: "job-42"}
				o := e.invoke(ctxFor(t), w, scope, textTarget(text), textRequest(text))
				o.record(ev)
				ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
				obs := rec.awaitCall(ev, scope.RequestID)
				checkCallRecords(ev, obs, scope, "primary", o, 1)
				m := o.result.Metadata
				ev.Check("resolved snapshot is attributed", m.Resolved && m.AccountScopeID == "acct-tenant-a" && m.BindingVersion == "b1" &&
					m.CredentialVersion == "v1" && m.ProviderID == ai.ProviderOpenAI && m.API == ai.APIOpenAIResponses && m.ModelID == text.Model,
					"metadata=%+v", m)
				ev.Check("the streamed attempt finished with its response", len(m.Attempts) == 1 && m.Attempts[0].HTTPStatus == 200 && m.Attempts[0].Code == "",
					"attempts=%+v", m.Attempts)
				ev.Check("usage is reported", o.result.Message.Usage == text.Expect.Usage, "got %+v", o.result.Message.Usage)
			})
		}
	})

	t.Run("preflight-refusal", func(t *testing.T) {
		cases := []struct {
			name  string
			scope ai.CallScope
			code  ai.Code
		}{
			// No identity at all: the record names no tenant.
			{"no-tenant", ai.CallScope{RequestID: "req-e09-no-tenant"}, ai.CodeInvalidRequest},
			// Tenant B has no "primary" binding; tenant A's is not B's.
			{"foreign-binding", ai.CallScope{TenantID: tenantB.tenant, RequestID: "req-e09-foreign"}, ai.CodeBindingNotFound},
			{"denied-actor", ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-e09-denied", ActorID: "intruder"}, ai.CodeTenantDenied},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				ev := run.Case(t, "E09-observe-preflight-"+c.name)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				w.host.RestrictActors(tenantA.tenant, "primary", "actor-7")
				res, err := w.client.Complete(ctxFor(t), c.scope, textTarget(text), textRequest(text), nil)
				o := outcome{result: res, err: err}
				o.record(ev)
				ev.Check("refused", errors.Is(err, &ai.Error{Code: c.code}), "err=%v", err)
				obs := rec.awaitCall(ev, c.scope.RequestID)
				checkCallRecords(ev, obs, c.scope, "primary", o, 0)
				for _, x := range obs {
					ev.Check("no record is attributed to an authorized tenant's snapshot", !x.Call.Resolved && x.Call.AccountScopeID == "" &&
						x.Call.ProviderID == "" && x.Call.ModelID == "" && x.Call.BindingVersion == "" && x.Call.TenantID == c.scope.TenantID, "record=%+v", x)
				}
				ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
			})
		}
	})

	t.Run("admission-refusal", func(t *testing.T) {
		ev := run.Case(t, "E09-observe-admission-refusal")
		rec := newRecorder()
		adm := host.NewAdmission()
		adm.Deny(errHostAdmission)
		w := observedWorld(t, rec, func(c *ai.Config) { c.Admission = adm }, tenantA)
		scope := textScope("req-e09-admission")
		res, err := w.client.Complete(ctxFor(t), scope, textTarget(text), textRequest(text), nil)
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("admission denied", errors.Is(err, &ai.Error{Code: ai.CodeAdmissionDenied}), "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "primary", o, 0)
		ev.Check("the refused call is resolved but sent nothing", obs[len(obs)-1].Call.Resolved && len(w.provider.Requests()) == 0,
			"call=%+v requests=%d", obs[len(obs)-1].Call, len(w.provider.Requests()))
		ev.Check("the host's refusal text stays out of the record", !strings.Contains(observationText(t, obs), errHostAdmission.Error()), "")
	})

	t.Run("failure", func(t *testing.T) {
		ev := run.Case(t, "P01-E09-observe-failure")
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		body := `{"error":{"message":"The server had an error"}}`
		enqueue(ev, w, provider.Reply{Status: 500, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "req_vendor_500"},
			Chunks: [][]byte{[]byte(body)}, Script: body})
		scope := textScope("req-e09-failure")
		res, err := w.client.Complete(ctxFor(t), scope, textTarget(text), textRequest(text), nil)
		o := outcome{result: res, err: err}
		o.record(ev)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "primary", o, 1)
		a := obs[2].Attempt
		ev.Check("attempt_finished classifies the failed attempt", a != nil && a.HTTPStatus == 500 && a.Code == ai.CodeUpstreamError &&
			a.ProviderRequestID == "req_vendor_500", "attempt=%+v", a)
		ev.Check("call_finished names the vendor request id", obs[3].Error != nil && obs[3].Error.ProviderRequestID == "req_vendor_500", "error=%+v", obs[3].Error)
	})

	t.Run("retry", func(t *testing.T) {
		ev := run.Case(t, "P01-E09-observe-retry")
		f, raw := loadRetryFixture(t)
		ev.Fixture("retry.json", raw)
		rec := newRecorder()
		w := observedWorld(t, rec, func(c *ai.Config) { c.Clock = f.clock() }, tenantA)
		w.updateBinding(tenantA, func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: 2} })
		failing := provider.Reply{Status: 503, Header: map[string]string{"Content-Type": "application/json", "retry-after-ms": "10"},
			Chunks: [][]byte{[]byte("{}")}, Script: "{}"}
		enqueue(ev, w, failing, sseReply(t, text, provider.FramingLF))
		scope := textScope("req-e09-retry")
		res, err := w.client.Complete(ctxFor(t), scope, textTarget(text), textRequest(text), nil)
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("retried call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "primary", o, 2)
		first := obs[2].Attempt
		ev.Check("the retried attempt records its classification and delay", first != nil && first.HTTPStatus == 503 &&
			first.Code == ai.CodeUpstreamError && first.RetryDelay == 10*time.Millisecond, "attempt=%+v", first)
	})

	t.Run("interrupted", func(t *testing.T) {
		for _, how := range []string{"cancel", "close"} {
			t.Run(how, func(t *testing.T) {
				ev := run.Case(t, "P01-E09-observe-interrupted-"+how)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				held := make(chan struct{})
				r := heldReply(200, "text/event-stream", string(lfChunks(t, text.Events)[0]))
				r.OnHold = func() { close(held) }
				enqueue(ev, w, r)
				ctx, cancel := context.WithCancel(ctxFor(t))
				defer cancel()
				scope := textScope("req-e09-interrupted-" + how)
				s := w.client.Stream(ctx, scope, textTarget(text), textRequest(text), nil)
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
				checkCallRecords(ev, obs, scope, "primary", o, 1)
				ev.Check("call_finished reports the abort", obs[len(obs)-1].StopReason == ai.StopReasonAborted, "got %q", obs[len(obs)-1].StopReason)
				_ = s.Close()
			})
		}
	})

	t.Run("native-state-downgrade", func(t *testing.T) {
		ev := run.Case(t, "P01-E09-observe-native-downgrade")
		f, raw := loadNativeFixture(t)
		ev.Fixture("native-state.json", raw)
		rec := newRecorder()
		w := observedWorld(t, rec, nil, tenantA)
		// Restored from storage without TrustNativeState: no envelope.
		restored, _, _ := restore(t, persist(t, f.round1(t, w, tenantA, f.Model).Message))
		w.provider.Enqueue(f.round2Reply(t))
		scope := textScope("req-e09-downgrade")
		res, err := w.client.Complete(ctxFor(t), scope, f.target(f.Model), f.round2Request(restored), nil)
		o := outcome{result: res, err: err}
		o.record(ev)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		obs := rec.awaitCall(ev, scope.RequestID)
		checkCallRecords(ev, obs, scope, "primary", o, 1)
		want := ai.NativeStateDowngrades{NoEnvelope: 1}
		ev.Check("the Result counts the downgrade", res.Metadata.NativeStateDowngrades == want, "got %+v", res.Metadata.NativeStateDowngrades)
		for _, x := range obs[1:] {
			ev.Check("the observer sees the downgrade count and reason", x.Call.NativeStateDowngrades == want, "%s: got %+v", x.Kind, x.Call.NativeStateDowngrades)
		}
	})

	t.Run("slow-observer", func(t *testing.T) {
		// The observer blocks on its first record; the queue holds 2. Every
		// call still completes at once, the overflow is dropped and counted,
		// and once the observer resumes every record is accounted for.
		ev := run.Case(t, "E09-observe-slow-observer")
		rec := newRecorder()
		rec.gate = make(chan struct{})
		w := observedWorld(t, rec, func(c *ai.Config) { c.Policy.MaxQueuedObservations = 2 }, tenantA)
		const calls = 5
		var reference ai.AssistantMessage
		for i := range calls {
			enqueue(ev, w, sseReply(t, text, provider.FramingLF))
			res := within(ev, fmt.Sprintf("call %d completes while the observer blocks", i), func() callResult {
				res, err := w.client.Complete(ctxFor(t), textScope(fmt.Sprintf("req-e09-slow-%d", i)), textTarget(text), textRequest(text), nil)
				return callResult{res, err}
			})
			ev.Check("call succeeds", res.err == nil, "err=%v", res.err)
			if i == 0 {
				waitFor(ev, "observer is blocked on its first record", rec.entered)
				reference = res.res.Message
			}
			checkFinalMessage(ev, text, res.res.Message)
			ev.Check("the result does not depend on the observer", res.res.Message.StopReason == reference.StopReason, "")
		}
		stats := w.client.ObserverStats()
		ev.Record("stats-while-blocked", stats)
		ev.Check("records beyond the queue are dropped and counted", stats.Dropped >= calls*4-3, "stats=%+v", stats)
		close(rec.gate)
		eventually(ev, "every record is delivered or dropped", func() bool {
			s := w.client.ObserverStats()
			return s.Delivered+s.Dropped == calls*4
		})
		final := w.client.ObserverStats()
		ev.Record("stats", final)
		ev.Check("delivered records are those not dropped", final.Delivered == uint64(len(rec.all())) && final.Failed == 0, "stats=%+v got %d", final, len(rec.all()))
	})

	t.Run("failing-observer", func(t *testing.T) {
		for _, mode := range []string{"error", "panic"} {
			t.Run(mode, func(t *testing.T) {
				ev := run.Case(t, "E09-observe-failing-observer-"+mode)
				rec := newRecorder()
				if mode == "error" {
					rec.fail = errors.New("observer backend down")
				} else {
					rec.panics = true
				}
				w := observedWorld(t, rec, nil, tenantA)
				for _, e := range outcomeEntries {
					enqueue(ev, w, sseReply(t, text, provider.FramingLF))
					o := e.invoke(ctxFor(t), w, textScope("req-e09-failing-"+mode+"-"+e.name), textTarget(text), textRequest(text))
					ev.Check(e.name+" succeeds", o.err == nil, "err=%v", o.err)
					checkFinalMessage(ev, text, o.result.Message)
				}
				want := uint64(len(outcomeEntries) * 4)
				eventually(ev, "every record is handed over", func() bool { return w.client.ObserverStats().Failed == want })
				stats := w.client.ObserverStats()
				ev.Record("stats", stats)
				ev.Check("failures are counted, nothing dropped", stats == ai.ObserverStats{Failed: want}, "stats=%+v", stats)
			})
		}
	})
}

// TestObservabilityRedaction covers E09's secret half (spec I9, User Story
// 39): with a key, its Authorization header, request content, tool
// arguments and native reasoning ciphertext in play, and a provider error
// body echoing the key, none of them reaches an observation, the error text
// or anything logged, while the authorized Result keeps its native state
// whole; the raw TenantID never reaches the provider.
func TestObservabilityRedaction(t *testing.T) {
	ev := run.Case(t, "P01-E09-redaction")
	f, raw := loadNativeFixture(t)
	ev.Fixture("native-state.json", raw)
	logs := captureLogs(t)
	rec := newRecorder()
	w := observedWorld(t, rec, nil, tenantA)

	round1 := f.round1(t, w, tenantA, f.Model)
	ev.Record("round1", round1)
	env, trusted := round1.Message.NativeState.Envelope()
	ev.Check("the authorized Result keeps its native state", trusted && env.TenantID == tenantA.tenant &&
		strings.Contains(string(mustMarshal(t, round1.Message)), "gAAAA-n1-backfilled") &&
		strings.Contains(string(mustMarshal(t, round1.Message)), "gAAAA-n2"), "envelope=%+v trusted=%v", env, trusted)

	w.provider.Enqueue(f.round2Reply(t))
	round2 := drain(w.client.Stream(ctxFor(t), textScope("req-e09-redaction-round2"), f.target(f.Model), f.round2Request(round1.Message), nil))
	round2.record(ev)
	ev.Check("round 2 replays natively and succeeds", round2.err == nil && round2.result.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
		"err=%v downgrades=%+v", round2.err, round2.result.Metadata.NativeStateDowngrades)

	// The vendor echoes the key, its Authorization header and the prompt.
	echo := fmt.Sprintf(`{"error":{"message":"Incorrect API key provided: %s. Received header Authorization: Bearer %s for input %q","type":"invalid_request_error","code":"invalid_api_key"}}`,
		tenantA.secret, tenantA.secret, f.User)
	enqueue(ev, w, provider.Reply{Status: 401, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(echo)}, Script: echo})
	failed, failedErr := w.client.Complete(ctxFor(t), textScope("req-e09-redaction-401"), f.target(f.Model), f.round2Request(round1.Message), nil)
	ev.Record("failed", failed)
	ev.Check("the echoing error is classified", errors.Is(failedErr, &ai.Error{Code: ai.CodeUpstreamAuth}), "err=%v", failedErr)

	var obs []ai.Observation
	for _, id := range []string{"req-native-round1-" + f.Model, "req-e09-redaction-round2", "req-e09-redaction-401"} {
		obs = append(obs, rec.awaitCall(ev, id)...)
	}
	sensitive := []string{tenantA.secret, "Bearer", f.User, f.SystemPrompt, `{"city":"Zürich"}`, "Zürich", f.ToolResult,
		"gAAAA", "Checking now.", "18°C and clear.", "Incorrect API key"}
	surface := map[string]string{
		"observations": observationText(t, obs),
		"logs":         logs.String(),
	}
	for name, text := range surface {
		for _, s := range sensitive {
			ev.Check("no sensitive value in "+name, !strings.Contains(text, s), "found %q", s)
		}
	}
	// The library adds no content of its own to error text. A provider's
	// own echo of the prompt stays in ErrorMessage as in pi (only key-shaped
	// text is redacted); that is recorded in ADR-0009 for the maintainer.
	errText := errString(failedErr) + errString(round2.err) + failed.Message.ErrorMessage
	for _, s := range []string{tenantA.secret, "Bearer sk-", f.SystemPrompt, `{"city":"Zürich"}`, f.ToolResult, "gAAAA"} {
		ev.Check("no secret or library-added content in the error text", !strings.Contains(errText, s), "found %q in %q", s, errText)
	}
	ev.Record("echoed-error-message", failed.Message.ErrorMessage)
	ev.Record("logs", logs.String())

	for _, r := range w.provider.Requests() {
		// The capture shows the bearer key as its alias, which names the
		// tenant; the key itself is what was sent there.
		header := maps.Clone(r.Header)
		delete(header, "Authorization")
		sent := string(r.Body) + fmt.Sprint(header)
		ev.Check("the raw TenantID never reaches the provider", !strings.Contains(sent, tenantA.tenant), "request %s %s", r.Method, r.Path)
	}
}

// observationText is everything an Observer could write out: each record as
// JSON and as formatted by fmt.
func observationText(t *testing.T, obs []ai.Observation) string {
	t.Helper()
	data, err := json.Marshal(obs)
	if err != nil {
		t.Fatal(err)
	}
	return string(data) + fmt.Sprintf("%+v %#v", obs, obs)
}

// captureLogs collects whatever the standard and structured loggers write
// during the test.
func captureLogs(t *testing.T) *lockedBuffer {
	t.Helper()
	buf := &lockedBuffer{}
	prevSlog, prevOut, prevFlags := slog.Default(), log.Writer(), log.Flags()
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	log.SetOutput(buf)
	t.Cleanup(func() {
		slog.SetDefault(prevSlog)
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	})
	return buf
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

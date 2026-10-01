package e2e

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// TestNativeStateReplay is the native state part of E03 (spec I6 last three
// paragraphs, User Stories 18–20, P01 "reasoning replay"). Round 1 produces
// reasoning items, a commentary message and a tool call; round 2 replays them
// with the tool's result. Native state — reasoning items with their encrypted
// content, redacted blocks, message and function call item ids — is replayed
// only when the message is from the target's provider, API and model and its
// envelope is trusted for the calling tenant and the binding's account.
// Otherwise the history is downgraded by pi's cross-model rules and the call
// still succeeds; the reason is counted in the Result metadata only.
func TestNativeStateReplay(t *testing.T) {
	f, raw := loadNativeFixture(t)
	native, downgraded := f.Round2.Expect.NativeBody, f.Round2.Expect.DowngradedBody
	envA := ai.NativeStateEnvelope{TenantID: tenantA.tenant, AccountScopeID: "acct-" + tenantA.tenant,
		ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ModelID: f.Model}

	t.Run("round1", func(t *testing.T) {
		ev := run.Case(t, "P01-E03-native-round1")
		ev.Fixture("native-state.json", raw)
		w := newWorld(t, tenantA)
		res := f.round1(t, w, tenantA, f.Model)
		ev.Record("result", res)
		m := res.Message
		ev.Check("reasoning signatures are JSON.stringify of the item, encrypted content backfilled",
			reflect.DeepEqual(contentSummary(m.Content), f.Round1.Expect.Content), "got %q\nwant %q", contentSummary(m.Content), f.Round1.Expect.Content)
		env, ok := m.NativeState.Envelope()
		ev.Check("the result's native state carries its envelope", ok && env == envA, "got %+v ok=%v", env, ok)
		ev.Check("no downgrade on a call without history", res.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{}, "got %+v", res.Metadata.NativeStateDowngrades)
		ev.Check("the envelope stays out of the message JSON", !jsonHasKey(t, mustMarshal(t, m), "nativeState"), "")
	})

	type scenario struct {
		name string
		// arrange runs round 1 (or not) and returns round 2's history message
		// and the tenant making round 2.
		arrange func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey)
		body    []byte
		want    ai.NativeStateDowngrades
		key     tenantKey
		model   string
	}
	inProcess := func(t *testing.T, _ *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
		return f.round1(t, w, tenantA, f.Model).Message, tenantA
	}
	restoredTrusted := func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
		stored := persist(t, f.round1(t, w, tenantA, f.Model).Message)
		ev.Record("stored-record", json.RawMessage(stored))
		m, env, _ := restore(t, stored)
		if !ev.Check("the stored record holds the envelope", env != nil && *env == envA, "got %+v", env) {
			t.FailNow()
		}
		// The host checked the record's tenant and integrity in its own store.
		trusted, err := ai.TrustNativeState(ai.CallScope{TenantID: tenantA.tenant, RequestID: "req-native-round2"}, *env)
		if !ev.Check("the host vouches for its own record", err == nil, "err=%v", err) {
			t.FailNow()
		}
		m.NativeState = trusted
		return m, tenantA
	}
	scenarios := []scenario{
		{name: "trusted-in-process", arrange: inProcess, body: native, key: tenantA},
		{name: "trusted-restored", arrange: restoredTrusted, body: native, key: tenantA},
		{name: "key-rotated", body: native, key: tenantA2, arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
			m, _ := restoredTrusted(t, ev, w)
			// K1→K2 on the same vendor account: the state stays replayable.
			w.host.PutCredential(primaryCredential(tenantA2, "v2"))
			return m, tenantA2
		}},
		{name: "no-envelope", body: downgraded, key: tenantA, want: ai.NativeStateDowngrades{NoEnvelope: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				m, _, _ := restore(t, persist(t, f.round1(t, w, tenantA, f.Model).Message))
				return m, tenantA
			}},
		{name: "forged-record", body: downgraded, key: tenantA, want: ai.NativeStateDowngrades{NoEnvelope: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				forged := forge(t, persist(t, f.round1(t, w, tenantA, f.Model).Message), envA)
				ev.Record("forged-record", json.RawMessage(forged))
				m, _, rec := restore(t, forged)
				_, ok := m.NativeState.Envelope()
				ev.Check("forged JSON decodes to untrusted state", rec.Trusted && !ok, "trusted=%v envelope ok=%v", rec.Trusted, ok)
				var direct ai.AssistantMessage
				err := json.Unmarshal([]byte(`{"content":[],"nativeState":`+string(rec.NativeState)+`,"trusted":true}`), &direct)
				_, ok = direct.NativeState.Envelope()
				ev.Check("a message decoded from JSON is never trusted", err == nil && !ok, "err=%v ok=%v", err, ok)
				return m, tenantA
			}},
		{name: "account-moved", body: downgraded, key: tenantA, want: ai.NativeStateDowngrades{AccountMismatch: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				m, k := restoredTrusted(t, ev, w)
				// The binding now serves another vendor account of the tenant.
				w.updateBinding(tenantA, func(b *ai.Binding) { b.AccountScopeID = "acct-tenant-a-eu" })
				w.replaceCredential(tenantA, func(c *ai.Credential) { c.AccountScopeID, c.Version = "acct-tenant-a-eu", "v2" })
				return m, k
			}},
		{name: "other-tenant", body: downgraded, key: tenantB, want: ai.NativeStateDowngrades{NoEnvelope: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				// Tenant A's own, in-process result leaks into tenant B's call.
				return f.round1(t, w, tenantA, f.Model).Message, tenantB
			}},
		{name: "relabeled-model", body: downgraded, key: tenantA, want: ai.NativeStateDowngrades{CrossModel: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				// The record claims the target model, but the envelope the
				// host vouches for says the state came from another one.
				m, env, _ := restore(t, persist(t, f.round1(t, w, tenantA, f.OtherModel).Message))
				m.Model = f.Model
				trusted, err := ai.TrustNativeState(ai.CallScope{TenantID: tenantA.tenant}, *env)
				ev.Check("the host vouches for the envelope", err == nil, "err=%v", err)
				m.NativeState = trusted
				return m, tenantA
			}},
		{name: "cross-model", body: downgraded, key: tenantA, want: ai.NativeStateDowngrades{CrossModel: 1},
			arrange: func(t *testing.T, ev *evidence.Case, w *world) (ai.AssistantMessage, tenantKey) {
				return f.round1(t, w, tenantA, f.OtherModel).Message, tenantA
			}},
	}
	for _, sc := range scenarios {
		entries := outcomeEntries[2:3] // complete-full; the trusted path covers all four
		if sc.name == "trusted-in-process" || sc.name == "no-envelope" {
			entries = outcomeEntries
		}
		for _, e := range entries {
			t.Run(sc.name+"/"+e.name, func(t *testing.T) {
				ev := run.Case(t, "P01-E03-native-"+sc.name+"-"+e.name)
				ev.Fixture("native-state.json", raw)
				w := newWorld(t, tenantA, tenantB)
				assistant, k := sc.arrange(t, ev, w)
				req := f.round2Request(assistant)
				before := persist(t, req.Messages[1].(ai.AssistantMessage))

				enqueue(ev, w, f.round2Reply(t))
				scope := ai.CallScope{TenantID: k.tenant, RequestID: "req-native-round2"}
				o := within(ev, "round 2 ends", func() outcome { return e.invoke(ctxFor(t), w, scope, f.target(f.Model), req) })
				o.record(ev)

				m := o.result.Message
				ev.Check("round 2 succeeds: a downgrade is never invalid_request", o.err == nil, "err=%v", o.err)
				ev.Check("stop reason", m.StopReason == f.Round2.Expect.StopReason, "got %q", m.StopReason)
				ev.Check("content", reflect.DeepEqual(contentSummary(m.Content), f.Round2.Expect.Content), "got %q", contentSummary(m.Content))
				ev.Check("downgrade counts and reasons in the Result metadata", o.result.Metadata.NativeStateDowngrades == sc.want,
					"got %+v want %+v", o.result.Metadata.NativeStateDowngrades, sc.want)
				reqs := w.provider.Requests()
				last := reqs[len(reqs)-1]
				ev.Check("round 2 wire body", jsonEqual(last.Body, entryBody(t, sc.body, strings.Contains(e.name, "simple"), f.SimpleMaxOutputTokens)), "got %s", last.Body)
				ev.Check("round 2 uses the calling tenant's current key", last.KeyAlias == k.alias, "got %q want %q", last.KeyAlias, k.alias)
				ev.Check("the caller's history is unchanged", string(persist(t, req.Messages[1].(ai.AssistantMessage))) == string(before), "")
				ev.Check("downgrade counts stay out of the pi-compatible message", !jsonHasKey(t, mustMarshal(t, m), "nativeStateDowngrades"), "")
			})
		}
	}
}

// TestTrustNativeState: the only way to trusted native state from stored data
// is the host's explicit TrustNativeState, and it refuses an envelope that
// cannot be checked or names another tenant than the scope (spec I6, E10
// "JSON 反序列化无法得到可信原生状态").
//
// Ways it could fail, listed before the implementation:
//
//	T1 zero TrustedNativeState (or one decoded from JSON) reports an envelope
//	T2 an envelope for another tenant than the scope is accepted
//	T3 a scope without TenantID is accepted
//	T4 an envelope missing account, provider, API or model is accepted, so
//	   a partial record could later match by accident
//	T5 a decoded {"trusted":true,…} TrustedNativeState is trusted
//	T6 the accepted envelope is not returned unchanged by Envelope
func TestTrustNativeState(t *testing.T) {
	ev := run.Case(t, "E10-native-trust-constructor")
	env := ai.NativeStateEnvelope{TenantID: "tenant-a", AccountScopeID: "acct-tenant-a",
		ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ModelID: "gpt-4.1-mini"}
	scope := ai.CallScope{TenantID: "tenant-a", RequestID: "req-trust"}

	var zero ai.TrustedNativeState
	_, ok := zero.Envelope()
	ev.Check("T1 zero value is untrusted", !ok, "")

	var decoded ai.TrustedNativeState
	err := json.Unmarshal([]byte(`{"trusted":true,"env":{"tenantId":"tenant-a"},"TenantID":"tenant-a"}`), &decoded)
	_, ok = decoded.Envelope()
	ev.Check("T5 JSON cannot make trusted state", !ok, "err=%v", err)

	refusals := map[string]func() (ai.CallScope, ai.NativeStateEnvelope){
		"T2 other tenant":   func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.TenantID = "tenant-b"; return scope, e },
		"T3 no scope":       func() (ai.CallScope, ai.NativeStateEnvelope) { return ai.CallScope{}, env },
		"T4 no account":     func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.AccountScopeID = ""; return scope, e },
		"T4 no provider":    func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.ProviderID = ""; return scope, e },
		"T4 no API":         func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.API = ""; return scope, e },
		"T4 no model":       func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.ModelID = ""; return scope, e },
		"T4 no tenant":      func() (ai.CallScope, ai.NativeStateEnvelope) { e := env; e.TenantID = ""; return scope, e },
		"T2 empty envelope": func() (ai.CallScope, ai.NativeStateEnvelope) { return scope, ai.NativeStateEnvelope{} },
	}
	for name, arrange := range refusals {
		s, e := arrange()
		trusted, err := ai.TrustNativeState(s, e)
		_, ok := trusted.Envelope()
		ev.Check(name+" refused", errors.Is(err, ai.ErrInvalidNativeStateEnvelope) && !ok, "err=%v ok=%v", err, ok)
	}

	trusted, err := ai.TrustNativeState(scope, env)
	got, ok := trusted.Envelope()
	ev.Check("T6 accepted envelope round-trips", err == nil && ok && got == env, "err=%v got=%+v", err, got)
}

// jsonHasKey reports whether key appears as an object key anywhere in data.
func jsonHasKey(t *testing.T, data []byte, key string) bool {
	t.Helper()
	var walk func(v any) bool
	walk = func(v any) bool {
		switch v := v.(type) {
		case map[string]any:
			for k, x := range v {
				if k == key || walk(x) {
					return true
				}
			}
		case []any:
			for _, x := range v {
				if walk(x) {
					return true
				}
			}
		}
		return false
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatal(err)
	}
	return walk(v)
}

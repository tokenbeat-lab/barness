package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/requestid"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestHostIntegrationExample runs the host integration example (E10 宿主接入,
// Testing Decisions §1, User Story 49) against the local controlled Provider
// on one Client shared by tenants A and B, under the example's
// CloudInteractivePolicy. It asserts the two responsibilities separately:
// what the host boundary refuses before barness-ai is called (a claimed
// tenant, another tenant's session, a forged history record's trust), and
// what the library refuses whatever the host does (an invalid scope, a
// binding or model the tenant may not use). Downstream failures and merged
// streams are TestHostIntegrationStreams.
func TestHostIntegrationExample(t *testing.T) {
	f, raw := loadIsolationFixture(t)
	for _, p := range f.Protocols {
		t.Run(p.ID+"/session-turns", func(t *testing.T) {
			// Turn 1 is stored by the host; turn 2 reads it back from
			// storage, vouches for its native state, and replays it exactly
			// as the in-process message would be.
			ev := run.Case(t, protocolCase(p)+"-E10-example-host-session-turns")
			ev.Fixture("interleave.json", raw)
			hw := newHostWorld(t)
			sess := hw.open(t, tenantA)
			hw.enqueueFor(ev, tenantA, nativeReply(t, p), p.success(t, tenantA))

			sink1 := &collectSink{}
			res1, err := hw.host.Turn(ctxFor(t), principal(tenantA), hw.turn(p, sess, p.User), sink1)
			ev.Record("turn-1", map[string]any{"result": res1, "envelopes": envelopesJSON(sink1.all())})
			ev.Check("turn 1 succeeds", err == nil, "err=%v", err)
			checkHostTurn(ev, p, tenantA, res1, sink1)

			sink2 := &collectSink{}
			res2, err := hw.host.Turn(ctxFor(t), principal(tenantA), hw.turn(p, sess, "And once more?"), sink2)
			ev.Record("turn-2", map[string]any{"result": res2, "envelopes": envelopesJSON(sink2.all())})
			ev.Check("turn 2 succeeds", err == nil, "err=%v", err)
			checkHostTurn(ev, p, tenantA, res2, sink2)
			ev.Check("each turn is its own logical call", res1.Metadata.RequestID != res2.Metadata.RequestID, "both %q", res1.Metadata.RequestID)
			ev.Check("stored native state replays natively", res2.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{},
				"got %+v", res2.Metadata.NativeStateDowngrades)

			sent := hw.requestsAt(tenantA)
			want := inProcessSecondTurn(t, p, res1.Message)
			ev.Record("requests", sent)
			if ev.Check("two inference requests", len(sent) == 2, "got %d", len(sent)) {
				ev.Check("turn 2 sends what the in-process message would", jsonEqual(sent[1].Body, want), "got %s\nwant %s", sent[1].Body, want)
			}
			records := hw.store.records(t, tenantA.tenant, sess)
			ev.Record("stored-records", records)
			ev.Check("the session holds both turns", len(records) == 4, "got %d", len(records))
			ev.Check("no key is stored", !strings.Contains(strings.Join(records, ""), p.key(tenantA).secret), "a key was stored")
		})
	}

	for _, p := range f.Protocols {
		t.Run(p.ID+"/forged-history", func(t *testing.T) { testForgedHistory(t, p, raw) })
	}

	p := f.Protocols[0]
	t.Run("host-boundary", func(t *testing.T) {
		cases := []struct {
			name string
			// request builds the turn A's principal submits; sessB is B's.
			request func(hw hostWorld, sessA, sessB string) hostintegration.TurnRequest
			who     hostintegration.Principal
			want    error
		}{
			{"claimed-tenant", func(hw hostWorld, sessA, _ string) hostintegration.TurnRequest {
				r := hw.turn(p, sessA, p.User)
				r.TenantID = tenantB.tenant
				return r
			}, principal(tenantA), hostintegration.ErrForbidden},
			{"foreign-session", func(hw hostWorld, _, sessB string) hostintegration.TurnRequest {
				return hw.turn(p, sessB, p.User)
			}, principal(tenantA), hostintegration.ErrForbidden},
			{"unknown-session", func(hw hostWorld, _, _ string) hostintegration.TurnRequest {
				return hw.turn(p, "sess-made-up", p.User)
			}, principal(tenantA), hostintegration.ErrForbidden},
			{"unauthenticated", func(hw hostWorld, sessA, _ string) hostintegration.TurnRequest {
				r := hw.turn(p, sessA, p.User)
				r.TenantID = tenantA.tenant
				return r
			}, hostintegration.Principal{}, hostintegration.ErrForbidden},
			{"missing-prompt", func(hw hostWorld, sessA, _ string) hostintegration.TurnRequest {
				return hw.turn(p, sessA, "")
			}, principal(tenantA), hostintegration.ErrBadRequest},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				ev := run.Case(t, "E10-example-host-refuses-"+c.name)
				hw := newHostWorld(t)
				sessA, sessB := hw.open(t, tenantA), hw.open(t, tenantB)
				req := c.request(hw, sessA, sessB)
				ev.Record("turn-request", req)
				sink := &collectSink{}
				res, err := hw.host.Turn(ctxFor(t), c.who, req, sink)
				ev.Record("error", errString(err))
				ev.Check("the host refuses", errors.Is(err, c.want), "err=%v", err)
				ev.Check("no Result is made up", reflect.DeepEqual(res, ai.Result{}), "got %+v", res)
				checkLibraryUntouched(ev, hw, sink)
			})
		}
	})

	t.Run("library-boundary", func(t *testing.T) {
		// A host that skips its own checks still cannot get past the
		// library: these are refused by barness-ai itself, before anything
		// is sent.
		cases := []struct {
			name   string
			scope  ai.CallScope
			target ai.Target
			code   ai.Code
			phase  ai.Phase
		}{
			{"scope-without-tenant", ai.CallScope{RequestID: requestid.New()}, p.target(), ai.CodeInvalidRequest, ai.PhaseScope},
			{"scope-without-request-id", ai.CallScope{TenantID: tenantA.tenant}, p.target(), ai.CodeInvalidRequest, ai.PhaseScope},
			{"unknown-binding", ai.CallScope{TenantID: tenantA.tenant, RequestID: requestid.New()},
				ai.Target{BindingID: "not-a-binding-of-tenant-a", ModelID: p.Model}, ai.CodeBindingNotFound, ai.PhaseBinding},
			{"model-not-allowed", ai.CallScope{TenantID: tenantA.tenant, RequestID: requestid.New()},
				ai.Target{BindingID: p.BindingID, ModelID: "gpt-4.1"}, ai.CodeTenantDenied, ai.PhaseCapability},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				ev := run.Case(t, "E10-library-refuses-"+c.name)
				hw := newHostWorld(t)
				o := drain(hw.client.Stream(ctxFor(t), c.scope, c.target, p.request(), nil))
				o.record(ev)
				checkRejected(ev, o, c.code, c.phase)
				ev.Check("the refusal claims no resolved identity", !o.result.Metadata.Resolved && o.result.Metadata.ProviderID == "" &&
					o.result.Metadata.ModelID == "" && o.result.Metadata.AccountScopeID == "", "metadata %+v", o.result.Metadata)
				ev.Check("nothing was sent", len(hw.provider.Requests()) == 0, "got %d", len(hw.provider.Requests()))
			})
		}

		t.Run("native-state-trust", func(t *testing.T) {
			ev := run.Case(t, "E10-library-refuses-native-state-trust")
			env := ai.NativeStateEnvelope{TenantID: tenantB.tenant, AccountScopeID: "acct-tenant-b", ProviderID: ai.ProviderOpenAI,
				API: ai.APIOpenAIResponses, ModelID: p.Model}
			_, err := ai.TrustNativeState(principalScope(tenantA), env)
			ev.Check("another tenant's envelope cannot be vouched for", errors.Is(err, ai.ErrInvalidNativeStateEnvelope), "err=%v", err)
			var decoded ai.AssistantMessage
			forged := `{"content":[],"api":"openai-responses","provider":"openai","model":"gpt-4.1-mini","stopReason":"stop","timestamp":1,` +
				`"NativeState":{"env":{"TenantID":"tenant-a"}},"nativeState":{"trusted":true}}`
			_ = json.Unmarshal([]byte(forged), &decoded)
			var trusted ai.TrustedNativeState
			_ = json.Unmarshal([]byte(`{"env":{"TenantID":"tenant-a","AccountScopeID":"acct-tenant-a"},"trusted":true}`), &trusted)
			_, fromMessage := decoded.NativeState.Envelope()
			_, fromValue := trusted.Envelope()
			ev.Check("JSON never decodes into trusted native state", !fromMessage && !fromValue, "message %v value %v", fromMessage, fromValue)
		})
	})
}

// testForgedHistory: a stored record whose provenance was rewritten —
// naming another tenant, or dropped and replaced by a client's trust claim —
// is not vouched for by the host; the turn goes ahead with the state
// downgraded rather than replayed.
func testForgedHistory(t *testing.T, p isolationProtocol, raw []byte) {
	forgeries := []struct {
		name   string
		change func(rec map[string]any)
	}{
		{"envelope-names-other-tenant", func(rec map[string]any) {
			rec["envelope"].(map[string]any)["tenantId"] = tenantB.tenant
		}},
		{"trust-claimed-without-envelope", func(rec map[string]any) {
			delete(rec, "envelope")
			rec["trusted"] = true
			rec["nativeState"] = map[string]any{"trusted": true, "tenantId": tenantA.tenant}
		}},
	}
	for _, fg := range forgeries {
		t.Run(fg.name, func(t *testing.T) {
			ev := run.Case(t, protocolCase(p)+"-E10-example-host-forged-history-"+fg.name)
			ev.Fixture("interleave.json", raw)
			hw := newHostWorld(t)
			sess := hw.open(t, tenantA)
			hw.enqueueFor(ev, tenantA, nativeReply(t, p), p.success(t, tenantA))
			_, err := hw.host.Turn(ctxFor(t), principal(tenantA), hw.turn(p, sess, p.User), &collectSink{})
			ev.Check("turn 1 succeeds", err == nil, "err=%v", err)
			hw.store.forgeLast(fg.change)
			ev.Record("stored-records", hw.store.records(t, tenantA.tenant, sess))

			res, err := hw.host.Turn(ctxFor(t), principal(tenantA), hw.turn(p, sess, "And once more?"), &collectSink{})
			ev.Record("result", res)
			ev.Check("the turn still runs", err == nil, "err=%v", err)
			ev.Check("the forged state is downgraded as unvouched", res.Metadata.NativeStateDowngrades == ai.NativeStateDowngrades{NoEnvelope: 1},
				"got %+v", res.Metadata.NativeStateDowngrades)
		})
	}
}

// nativeReply is a first turn whose message carries native state for p:
// Responses' text keeps its output item id; on Anthropic, a signed thinking
// block (text.json "block-start-content"); on Gemini, text with a thought
// signature (text.json "signature-on-empty-text"); on Chat, thinking signed
// with reasoning details (text.json "reasoning-details-only"); on DeepSeek
// Responses, a reasoning item (deepseek-responses/text.json "reasoning");
// on DeepSeek Chat, thinking signed reasoning_content
// (deepseek-chat/text.json "reasoning").
func nativeReply(t *testing.T, p isolationProtocol) provider.Reply {
	t.Helper()
	switch {
	case p.anthropic():
		f, _ := loadFixture(t, anthropicProtocol, "text.json")
		return scenarioByID(t, f, "block-start-content").replies(t)[0]
	case p.gemini():
		f, _ := loadFixture(t, geminiProtocol, "text.json")
		return scenarioByID(t, f, "signature-on-empty-text").replies(t)[0]
	case p.chat():
		f, _ := loadFixture(t, chatProtocol, "text.json")
		return scenarioByID(t, f, "reasoning-details-only").replies(t)[0]
	case p.deepseek():
		f, _ := loadFixture(t, deepseekResponsesProtocol, "text.json")
		return scenarioByID(t, f, "reasoning").replies(t)[0]
	case p.deepseekChat():
		f, _ := loadFixture(t, deepseekChatProtocol, "text.json")
		return scenarioByID(t, f, "reasoning").replies(t)[0]
	}
	return p.success(t, tenantA)
}

// hostWorld is the host example on an isolation world: tenants A and B on
// one Client under CloudInteractivePolicy, each binding at its tenant's own
// endpoint, with an observer, the host's admission and resource gauges.
type hostWorld struct {
	isolationWorld
	store *forgeableStore
	host  *hostintegration.Host
}

func newHostWorld(t *testing.T) hostWorld {
	t.Helper()
	iw := newIsolationWorld(t, func(c *ai.Config) { c.Policy = hostintegration.CloudInteractivePolicy() })
	store := &forgeableStore{MemoryStore: hostintegration.NewMemoryStore()}
	return hostWorld{isolationWorld: iw, store: store, host: hostintegration.New(iw.client, store, requestid.New)}
}

func (hw hostWorld) open(t *testing.T, k tenantKey) string {
	t.Helper()
	sess, err := hw.host.OpenSession(ctxFor(t), principal(k))
	if err != nil {
		t.Fatal(err)
	}
	return sess
}

// turn is a client's request on p's binding and model.
func (hostWorld) turn(p isolationProtocol, session, prompt string) hostintegration.TurnRequest {
	return hostintegration.TurnRequest{SessionID: session, BindingID: p.BindingID, ModelID: p.Model, Prompt: prompt}
}

// principal is k's tenant's authentication result.
func principal(k tenantKey) hostintegration.Principal {
	return hostintegration.Principal{TenantID: k.tenant, ActorID: "user-of-" + k.tenant}
}

func principalScope(k tenantKey) ai.CallScope {
	return ai.CallScope{TenantID: k.tenant, ActorID: "user-of-" + k.tenant, RequestID: requestid.New()}
}

// collectSink is a downstream connection that keeps every envelope and,
// when fail is set, lets it decide whether a send fails.
type collectSink struct {
	mu   sync.Mutex
	envs []ai.EventEnvelope
	fail func(ai.EventEnvelope) error
}

func (s *collectSink) Send(env ai.EventEnvelope) error {
	s.mu.Lock()
	s.envs = append(s.envs, env)
	fail := s.fail
	s.mu.Unlock()
	if fail != nil {
		return fail(env)
	}
	return nil
}

func (s *collectSink) all() []ai.EventEnvelope {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]ai.EventEnvelope(nil), s.envs...)
}

// forgeableStore is the example's MemoryStore whose reads can return a
// rewritten last record, standing in for a record forged in, or corrupted
// by, storage.
type forgeableStore struct {
	*hostintegration.MemoryStore
	mu    sync.Mutex
	forge func(map[string]any)
}

func (s *forgeableStore) Load(ctx context.Context, tenant, sessionID string) ([][]byte, error) {
	recs, err := s.MemoryStore.Load(ctx, tenant, sessionID)
	s.mu.Lock()
	forge := s.forge
	s.mu.Unlock()
	if err != nil || forge == nil || len(recs) == 0 {
		return recs, err
	}
	var rec map[string]any
	if err := json.Unmarshal(recs[len(recs)-1], &rec); err != nil {
		return nil, err
	}
	forge(rec)
	last, err := json.Marshal(rec)
	recs[len(recs)-1] = last
	return recs, err
}

// forgeLast makes every later read rewrite the last record with change.
func (s *forgeableStore) forgeLast(change func(map[string]any)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.forge = change
}

func (s *forgeableStore) records(t *testing.T, tenant, session string) []string {
	t.Helper()
	recs, err := s.Load(context.Background(), tenant, session)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(recs))
	for i, r := range recs {
		out[i] = string(r)
	}
	return out
}

// checkHostTurn asserts every envelope sent downstream is attributed to k's
// resolved call, as its Result is.
func checkHostTurn(ev *evidence.Case, p isolationProtocol, k tenantKey, res ai.Result, sink *collectSink) {
	want := ai.CallAttribution{Operation: ai.OperationChat, TenantID: k.tenant, RequestID: res.Metadata.RequestID, ActorID: "user-of-" + k.tenant,
		BindingID: p.BindingID, Resolved: true, ProviderID: p.provider(), API: p.api(), ModelID: p.Model, AccountScopeID: p.account(k)}
	ev.Check("the RequestID is minted by the host", strings.HasPrefix(res.Metadata.RequestID, "req-"), "got %q", res.Metadata.RequestID)
	checkEnvelopes(ev, sink.all(), res, want)
}

// checkLibraryUntouched asserts a request refused at the host boundary
// never reached barness-ai: nothing sent, observed, admitted or streamed.
func checkLibraryUntouched(ev *evidence.Case, hw hostWorld, sink *collectSink) {
	ev.Check("nothing was sent", len(hw.provider.Requests()) == 0, "got %d", len(hw.provider.Requests()))
	ev.Check("barness-ai observed no call", len(hw.rec.all()) == 0, "got %d records", len(hw.rec.all()))
	ev.Check("no admission was asked", len(hw.admission.Requests()) == 0, "got %d", len(hw.admission.Requests()))
	ev.Check("nothing went downstream", len(sink.all()) == 0, "got %d", len(sink.all()))
}

// inProcessSecondTurn is the body of the host's second turn when the first
// turn's message is passed on in process, trusted by construction: the
// reference for a turn replayed from storage.
func inProcessSecondTurn(t *testing.T, p isolationProtocol, first ai.AssistantMessage) []byte {
	t.Helper()
	iw := newIsolationWorld(t)
	iw.provider.EnqueueAt(tenantPrefix(tenantA)+"/", p.success(t, tenantA))
	req := ai.Request{Messages: []ai.Message{ai.UserText(p.User), first, ai.UserText("And once more?")}}
	if _, err := iw.client.Complete(ctxFor(t), principalScope(tenantA), p.target(), req, nil); err != nil {
		t.Fatalf("reference turn: %v", err)
	}
	return iw.requestsAt(tenantA)[0].Body
}

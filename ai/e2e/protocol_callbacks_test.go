package e2e

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestProtocolCallbacks covers the trusted callbacks on each scenario
// protocol (spec I7, E04): the order header transform → payload → response →
// start — on Gemini the response callback never runs, as in pi, and the
// payload is the REST body (ADR-0012) — a replaced or changed payload sent,
// a failing callback ending the call, the authorization a callback cannot
// widen, and the header transform's guarded credential.
func TestProtocolCallbacks(t *testing.T) {
	forEachSuite(t, protocolSuites, testProtocolCallbacks)
}

func testProtocolCallbacks(t *testing.T, s *protocolSuite) {
	text, raw := loadFixture(t, s.proto, "text.json")
	plain := scenarioByID(t, text, "text")
	cb := s.callbacks
	start := func(t *testing.T, name string) (*evidence.Case, *world) {
		return startScenarioCallbackCase(t, s.caseID("E04-callbacks-"+name), plain, raw)
	}

	t.Run("order-and-inputs", func(t *testing.T) {
		ev, w := start(t, "order-and-inputs")
		var log callLog
		var seenHeader http.Header
		var seenPayload ai.Payload
		var seenResponse ai.ResponseInfo
		hooks := ai.Hooks{
			TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
				log.add("headers")
				seenHeader = h.Clone()
				return nil
			},
			OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				log.add("payload")
				seenPayload = *p
				return ai.KeepPayload(), nil
			},
			OnResponse: func(_ context.Context, _ ai.CallScope, r ai.ResponseInfo) error {
				log.add("response")
				seenResponse = r
				return nil
			},
		}
		st := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("req-cb-order"), plain.target(), plain.request(t), nil)
		for st.Next() {
			log.add("event:" + string(st.Event().Type()))
		}
		_, err := st.Result()
		ev.Record("callback-log", log.entries())
		ev.Check("call succeeds", err == nil, "err=%v", err)
		got := log.entries()
		want := []string{"headers", "payload", "response", "event:start"}
		if !cb.onResponse {
			want = []string{"headers", "payload", "event:start"}
		}
		ev.Check("headers, then payload, then response if the protocol has one, then start", len(got) >= len(want) &&
			reflect.DeepEqual(got[:len(want)], want), "got %q", got)
		responses := 0
		for _, e := range got {
			if e == "response" {
				responses++
			}
		}
		if cb.onResponse {
			ev.Check("the response callback runs once", responses == 1, "got %q", got)
			ev.Check("response callback reads status and the vendor's headers", seenResponse.Status == http.StatusOK &&
				seenResponse.Header.Get("Content-Type") == "text/event-stream" && seenResponse.ProviderID == s.proto.provider &&
				seenResponse.API == s.proto.api, "got %+v", seenResponse)
		} else {
			ev.Check("the response callback never runs, as in pi", responses == 0, "got %q", got)
		}
		ev.Check("transform sees the key, redacted", seenHeader.Get(s.keyHeader) == "[REDACTED]", "got %q", seenHeader.Get(s.keyHeader))
		ev.Check("payload identifies the model", seenPayload.ProviderID == s.proto.provider &&
			seenPayload.API == s.proto.api && seenPayload.ModelID == plain.Model, "got %+v", seenPayload)
		ev.Check("payload body is the request body sent", jsonEqual(mustMarshal(t, seenPayload.Body), plain.Expect.Request.Body),
			"got %s", mustMarshal(t, seenPayload.Body))
	})

	if cb.onResponse {
		t.Run("response-callback-failure-before-start", func(t *testing.T) {
			ev, w := start(t, "response-callback-failure-before-start")
			hostErr := errors.New("host rejected the response")
			hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { return hostErr }}
			st := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("req-cb-response"), plain.target(), plain.request(t), nil)
			var events []ai.Event
			for st.Next() {
				events = append(events, st.Event())
			}
			res, err := st.Result()
			ev.Record("result", res)
			ev.Check("only the error terminal: the callback ran before start", reflect.DeepEqual(eventSummary(events), []string{"error:error"}),
				"got %q", eventSummary(events))
			ev.Check("callback_failed, the host's error reachable", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed}) && errors.Is(err, hostErr),
				"got %v", err)
		})
	} else {
		t.Run("response-callback-error-ignored", func(t *testing.T) {
			ev, w := start(t, "response-callback-error-ignored")
			hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				return errors.New("never called")
			}}
			res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-response"), plain.target(), plain.request(t), nil)
			ev.Record("result", res)
			ev.Check("call succeeds: the response callback is not part of this protocol", err == nil && res.Message.StopReason == ai.StopReasonStop,
				"err=%v", err)
		})
	}

	t.Run(cb.replaced.name, func(t *testing.T) {
		ev, w := start(t, cb.replaced.name)
		calls := 0
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			calls++
			return ai.ReplacePayload(cb.replaced.body(p)), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-replace"), plain.target(), plain.request(t), s.options)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("payload callback ran once", calls == 1, "got %d", calls)
		r := lastRequest(ev, w)
		ev.Check(cb.replaced.about, jsonEqual(r.Body, []byte(cb.replaced.sent)), "got %s", r.Body)
		want := plain.Expect.Request
		ev.Check("to the authorized model's path", r.Path == want.Path && r.Query == want.Query, "got %s?%s", r.Path, r.Query)
		if cb.replaced.check != nil {
			cb.replaced.check(ev, r)
		}
	})

	if c := cb.changedInPlace; c != nil {
		t.Run("changed-in-place", func(t *testing.T) {
			ev, w := start(t, "changed-in-place")
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				c.change(p.Body)
				return ai.KeepPayload(), nil
			}}
			_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-in-place"), plain.target(), plain.request(t), nil)
			ev.Check("call succeeds", err == nil, "err=%v", err)
			var body map[string]any
			mustUnmarshal(t, lastRequest(ev, w).Body, &body)
			ev.Check(c.about, c.sent(body), "got %v", body)
		})
	}

	// Every refusal holds on the full and the simple entry alike (spec: no
	// lower-level bypass).
	for _, c := range cb.widen {
		for _, simple := range []bool{false, true} {
			entry := map[bool]string{false: "full", true: "simple"}[simple]
			t.Run("cannot-widen-"+c.name+"/"+entry, func(t *testing.T) {
				ev, w := start(t, "cannot-widen-"+c.name+"-"+entry)
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					c.change(p.Body)
					return ai.KeepPayload(), nil
				}}
				hooked := w.client.WithHooks(hooks)
				var res ai.Result
				var err error
				if simple {
					res, err = hooked.CompleteSimple(ctxFor(t), textScope("req-cb-widen"), plain.target(), plain.request(t), ai.SimpleOptions{})
				} else {
					res, err = hooked.Complete(ctxFor(t), textScope("req-cb-widen"), plain.target(), plain.request(t), nil)
				}
				ev.Record("result", res)
				ev.Check("tenant_denied in the request phase", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied, Phase: ai.PhaseRequest}), "got %v", err)
				ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
			})
		}
	}

	t.Run("hosted-tool-the-binding-allows", func(t *testing.T) {
		ev, w := start(t, "hosted-tool-the-binding-allows")
		w.updateBindingOf(tenantA, s.proto.binding, func(b *ai.Binding) { b.AllowedHostedTools = []string{cb.hostedTool.name} })
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			cb.hostedTool.add(p.Body)
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-hosted"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check(cb.hostedTool.about, cb.hostedTool.sent(body), "got %v", body)
	})

	t.Run("payload-callback-failure", func(t *testing.T) {
		ev, w := start(t, "payload-callback-failure")
		hostErr := errors.New("host rejected the payload")
		hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
			return ai.KeepPayload(), hostErr
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-payload-failure"), plain.target(), plain.request(t), nil)
		ev.Record("result", res)
		ev.Check("callback_failed, the host's error reachable", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed}) && errors.Is(err, hostErr),
			"got %v", err)
		ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})

	t.Run("header-transform", func(t *testing.T) {
		ev, w := start(t, "header-transform")
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("X-Request-Tag", "host-1")
			h.Del("User-Agent")
			if cb.headerTransform.extra != nil {
				cb.headerTransform.extra(h)
			}
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-headers"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		r := lastRequest(ev, w)
		ev.Check("an added header is sent", r.Header["X-Request-Tag"] == "host-1", "got %q", r.Header["X-Request-Tag"])
		_, ua := r.Header["User-Agent"]
		ev.Check("a removed header is not sent, not even the SDK's or HTTP stack's default", !ua, "got %q", r.Header["User-Agent"])
		ev.Check("the tenant's key still authenticates", r.KeyAlias == s.proto.key(tenantA).alias, "got %q", r.KeyAlias)
		if cb.headerTransform.check != nil {
			cb.headerTransform.check(ev, r)
		}
	})

	t.Run("header-transform-cannot-change-the-key", func(t *testing.T) {
		ev, w := start(t, "header-transform-cannot-change-the-key")
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set(s.keyHeader, s.keyValue(s.proto.key(tenantB)))
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-key"), plain.target(), plain.request(t), nil)
		ev.Check("tenant_denied", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}), "got %v", err)
		ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})

	if cb.more != nil {
		cb.more(t, s)
	}
}

func startScenarioCallbackCase(t *testing.T, id string, sc fixtureScenario, raw []byte) (*evidence.Case, *world) {
	t.Helper()
	ev := run.Case(t, id)
	ev.Fixture("fixture.json", raw)
	w := scenarioWorld(t, sc)
	enqueue(ev, w, sc.replies(t)...)
	return ev, w
}

func lastRequest(ev *evidence.Case, w *world) provider.Request {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if len(reqs) == 0 {
		ev.Check("a request was sent", false, "none")
		ev.T().FailNow()
	}
	return reqs[len(reqs)-1]
}

// TestProtocolNoEnvironmentFallback is H1 on each scenario protocol:
// everything the protocol's SDKs read from the environment (keys, base URLs,
// and the protocol's own: Anthropic's profile and custom headers, Google's
// Vertex switch and project, OpenAI's organization and project) points at a
// decoy, and the call still uses only the binding's endpoint and the
// tenant's key.
func TestProtocolNoEnvironmentFallback(t *testing.T) {
	forEachSuite(t, protocolSuites, func(t *testing.T, s *protocolSuite) {
		f, raw := loadFixture(t, s.proto, "text.json")
		sc := scenarioByID(t, f, "text")
		ev := run.Case(t, s.caseID("H1-no-environment-fallback"))
		decoy := polluteEnvironment(t)
		for k, v := range s.env.vars {
			t.Setenv(k, v)
		}
		w, o := runScenario(t, ev, sc, raw, modeComplete)
		ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
		ev.Check("environment endpoint not used", len(decoy.Requests()) == 0, "decoy got %d", len(decoy.Requests()))
		r := lastRequest(ev, w)
		ev.Check("the tenant's key, not the environment's", r.KeyAlias == s.proto.key(tenantA).alias, "got %q", r.KeyAlias)
		for _, name := range s.env.leaked {
			_, leaked := r.Header[http.CanonicalHeaderKey(name)]
			ev.Check("no environment "+name+" header", !leaked, "got %q", r.Header[http.CanonicalHeaderKey(name)])
		}
	})
}

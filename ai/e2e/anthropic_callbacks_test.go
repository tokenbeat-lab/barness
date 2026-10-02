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

// TestAnthropicCallbacks covers the trusted callbacks on the Anthropic path
// (spec I7 Anthropic row, E04): the order header transform → payload →
// response → start, a replaced payload still streaming, the beta features as
// the betas parameter a payload callback may change, the authorization a
// callback cannot widen, and the header transform's guarded credential.
func TestAnthropicCallbacks(t *testing.T) {
	text, raw := loadFixture(t, anthropicProtocol, "text.json")
	plain := scenarioByID(t, text, "text")
	options, oraw := loadFixture(t, anthropicProtocol, "options.json")
	budget := scenarioByID(t, options, "simple-budget-reasoning")

	t.Run("order-and-inputs", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-order-and-inputs", plain, raw)
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
		s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("req-cb-order"), plain.target(), plain.request(t), nil)
		for s.Next() {
			log.add("event:" + string(s.Event().Type()))
		}
		_, err := s.Result()
		ev.Record("callback-log", log.entries())
		ev.Check("call succeeds", err == nil, "err=%v", err)
		got := log.entries()
		want := []string{"headers", "payload", "response", "event:start"}
		ev.Check("headers, then payload, then response, then start", len(got) >= 4 && reflect.DeepEqual(got[:4], want), "got %q", got)
		ev.Check("transform sees the API key, redacted", seenHeader.Get("X-Api-Key") == "[REDACTED]", "got %q", seenHeader.Get("X-Api-Key"))
		ev.Check("payload identifies the model", seenPayload.ProviderID == ai.ProviderAnthropic &&
			seenPayload.API == ai.APIAnthropicMessages && seenPayload.ModelID == plain.Model, "got %+v", seenPayload)
		ev.Check("payload body is the native request body", jsonEqual(mustMarshal(t, seenPayload.Body), plain.Expect.Request.Body),
			"got %s", mustMarshal(t, seenPayload.Body))
		ev.Check("response callback reads status and the vendor's headers", seenResponse.Status == http.StatusOK &&
			seenResponse.Header.Get("Content-Type") == "text/event-stream" && seenResponse.ProviderID == ai.ProviderAnthropic,
			"got %+v", seenResponse)
	})

	t.Run("replaced-payload-still-streams", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-replaced-payload-still-streams", plain, raw)
		calls := 0
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			calls++
			return ai.ReplacePayload(map[string]any{
				"model":      p.ModelID,
				"messages":   []any{map[string]any{"role": "user", "content": "Replaced."}},
				"max_tokens": 10,
				"stream":     false,
				"betas":      []any{"custom-beta-1", "custom-beta-1", "b2"},
			}), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-replace"), plain.target(), plain.request(t), ai.AnthropicOptions{})
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("payload callback ran once", calls == 1, "got %d", calls)
		r := lastRequest(ev, w)
		ev.Check("the replacement is sent with stream forced back on and without betas", jsonEqual(r.Body,
			[]byte(`{"model":"claude-haiku-4-5","messages":[{"role":"user","content":"Replaced."}],"max_tokens":10,"stream":true}`)),
			"got %s", r.Body)
		ev.Check("the betas parameter is sent as the anthropic-beta header as given, as pi's SDK does",
			r.Header["Anthropic-Beta"] == "custom-beta-1,custom-beta-1,b2", "got %q", r.Header["Anthropic-Beta"])
	})

	t.Run("payload-sees-betas", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-payload-sees-betas", budget, oraw)
		var betas any
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			betas = p.Body["betas"]
			delete(p.Body, "betas")
			p.Body["top_k"] = 3
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).CompleteSimple(ctxFor(t), textScope("req-cb-betas"), budget.target(), budget.request(t), budget.simpleOptions(t))
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("the callback sees the beta features as pi's betas parameter",
			reflect.DeepEqual(betas, []any{"interleaved-thinking-2025-05-14"}), "got %#v", betas)
		r := lastRequest(ev, w)
		_, beta := r.Header["Anthropic-Beta"]
		ev.Check("removing them in place removes the header", !beta, "got %q", r.Header["Anthropic-Beta"])
		var body map[string]any
		mustUnmarshal(t, r.Body, &body)
		_, hasBetas := body["betas"]
		ev.Check("an in-place change is sent; betas never travel in the body", body["top_k"] == 3.0 && !hasBetas, "got %s", r.Body)
	})

	widen := []struct {
		name   string
		change func(map[string]any)
	}{
		{"model", func(b map[string]any) { b["model"] = "claude-opus-4-5" }},
		{"mcp-servers", func(b map[string]any) {
			b["mcp_servers"] = []any{map[string]any{"type": "url", "url": "https://mcp.example"}}
		}},
		{"container", func(b map[string]any) { b["container"] = "container_123" }},
		{"hosted-tool", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "web_search_20250305", "name": "web_search"}}
		}},
		{"file-source", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "document",
				"source": map[string]any{"type": "file", "file_id": "file_123"}}}}}
		}},
	}
	// Every refusal holds on the full and the simple entry alike (spec: no
	// lower-level bypass).
	for _, c := range widen {
		for _, simple := range []bool{false, true} {
			entry := map[bool]string{false: "full", true: "simple"}[simple]
			t.Run("cannot-widen-"+c.name+"/"+entry, func(t *testing.T) {
				ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-cannot-widen-"+c.name+"-"+entry, plain, raw)
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
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-hosted-tool-the-binding-allows", plain, raw)
		b := w.host.Binding(tenantA.tenant, "claude")
		b.AllowedHostedTools = []string{"web_search_20250305"}
		w.host.PutBindingAt(tenantA.tenant, "claude", b)
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["tools"] = []any{map[string]any{"type": "web_search_20250305", "name": "web_search"}}
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-hosted"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check("the allowed hosted tool is sent", len(body["tools"].([]any)) == 1, "got %v", body["tools"])
	})

	t.Run("response-callback-failure-before-start", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-response-callback-failure-before-start", plain, raw)
		hostErr := errors.New("host rejected the response")
		hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { return hostErr }}
		s := w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("req-cb-response"), plain.target(), plain.request(t), nil)
		var events []ai.Event
		for s.Next() {
			events = append(events, s.Event())
		}
		res, err := s.Result()
		ev.Record("result", res)
		ev.Check("only the error terminal: the callback ran before start", reflect.DeepEqual(eventSummary(events), []string{"error:error"}),
			"got %q", eventSummary(events))
		ev.Check("callback_failed, the host's error reachable", errors.Is(err, &ai.Error{Code: ai.CodeCallbackFailed}) && errors.Is(err, hostErr),
			"got %v", err)
	})

	t.Run("header-transform", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-header-transform", plain, raw)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Anthropic-Beta", "x-beta, x-beta,")
			h.Del("User-Agent")
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-headers"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		_, ua := lastRequest(ev, w).Header["User-Agent"]
		ev.Check("a removed header is not sent, not even the SDK's default", !ua, "got %q", lastRequest(ev, w).Header["User-Agent"])
		ev.Check("a configured anthropic-beta replaces the features, normalized as pi does",
			lastRequest(ev, w).Header["Anthropic-Beta"] == "x-beta", "got %q", lastRequest(ev, w).Header["Anthropic-Beta"])
	})

	t.Run("header-transform-cannot-change-the-key", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P02-E04-callbacks-header-transform-cannot-change-the-key", plain, raw)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("X-Api-Key", anthropicB.secret)
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-key"), plain.target(), plain.request(t), nil)
		ev.Check("tenant_denied", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}), "got %v", err)
		ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})
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

// TestAnthropicNoEnvironmentFallback is H1 on the Anthropic path: everything
// the Anthropic SDK's default client options read from the environment
// (key, auth token, base URL, profile, custom headers) points at a decoy,
// and the call still uses only the binding's endpoint and the tenant's key.
func TestAnthropicNoEnvironmentFallback(t *testing.T) {
	f, raw := loadFixture(t, anthropicProtocol, "text.json")
	sc := scenarioByID(t, f, "text")
	ev := run.Case(t, "P02-H1-no-environment-fallback")
	decoy := polluteEnvironment(t)
	t.Setenv("ANTHROPIC_PROFILE", "env-leak")
	t.Setenv("ANTHROPIC_CUSTOM_HEADERS", "X-Env-Leak: 1")
	w, o := runScenario(t, ev, sc, raw, modeComplete)
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	ev.Check("environment endpoint not used", len(decoy.Requests()) == 0, "decoy got %d", len(decoy.Requests()))
	_, leaked := lastRequest(ev, w).Header["X-Env-Leak"]
	ev.Check("no environment header", !leaked, "X-Env-Leak sent")
}

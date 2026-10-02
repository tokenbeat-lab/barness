package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Chat's data and its differences
// (ADR-0013); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestChatCallbacks covers the trusted callbacks on the Chat path (spec I7
// Chat row, E04): the order header transform → payload → response → start,
// a replaced payload still streaming, an in-place change sent, the
// authorization a callback cannot widen, and the header transform's guarded
// credential.
func TestChatCallbacks(t *testing.T) {
	text, raw := loadFixture(t, chatProtocol, "text.json")
	plain := scenarioByID(t, text, "text")

	t.Run("order-and-inputs", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-order-and-inputs", plain, raw)
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
		responses := 0
		for _, e := range got {
			if e == "response" {
				responses++
			}
		}
		ev.Check("the response callback runs once", responses == 1, "got %q", got)
		ev.Check("transform sees the bearer key, redacted", seenHeader.Get("Authorization") == "[REDACTED]", "got %q", seenHeader.Get("Authorization"))
		ev.Check("payload identifies the model", seenPayload.ProviderID == ai.ProviderOpenAI &&
			seenPayload.API == ai.APIOpenAICompletions && seenPayload.ModelID == plain.Model, "got %+v", seenPayload)
		ev.Check("payload body is the native request body", jsonEqual(mustMarshal(t, seenPayload.Body), plain.Expect.Request.Body),
			"got %s", mustMarshal(t, seenPayload.Body))
		ev.Check("response callback reads status and the vendor's headers", seenResponse.Status == http.StatusOK &&
			seenResponse.Header.Get("Content-Type") == "text/event-stream" && seenResponse.ProviderID == ai.ProviderOpenAI &&
			seenResponse.API == ai.APIOpenAICompletions, "got %+v", seenResponse)
	})

	t.Run("replaced-payload-still-streams", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-replaced-payload-still-streams", plain, raw)
		calls := 0
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			calls++
			return ai.ReplacePayload(map[string]any{
				"model":                 p.ModelID,
				"messages":              []any{map[string]any{"role": "user", "content": "Replaced."}},
				"max_completion_tokens": 10,
				"stream":                false,
			}), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-replace"), plain.target(), plain.request(t), ai.ChatOptions{})
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("payload callback ran once", calls == 1, "got %d", calls)
		r := lastRequest(ev, w)
		ev.Check("the replacement is sent with stream forced back on", jsonEqual(r.Body,
			[]byte(`{"model":"gpt-4.1-mini","messages":[{"role":"user","content":"Replaced."}],"max_completion_tokens":10,"stream":true}`)),
			"got %s", r.Body)
		ev.Check("to the Chat Completions path", r.Path == "/v1/chat/completions", "got %s", r.Path)
	})

	t.Run("changed-in-place", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-changed-in-place", plain, raw)
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["top_p"] = 0.5
			p.Body["functions"] = []any{map[string]any{"name": "legacy", "parameters": map[string]any{"type": "object"}}}
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-in-place"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check("the in-place changes are sent, legacy function tools included", body["top_p"] == 0.5 && body["functions"] != nil &&
			body["store"] == false, "got %v", body)
	})

	widen := []struct {
		name   string
		change func(map[string]any)
	}{
		{"model", func(b map[string]any) { b["model"] = "gpt-4o-mini" }},
		{"model-removed", func(b map[string]any) { delete(b, "model") }},
		{"cache-key", func(b map[string]any) { b["prompt_cache_key"] = "other-tenant-cache" }},
		{"stored", func(b map[string]any) { b["store"] = true }},
		{"hosted-tool", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "custom", "custom": map[string]any{"name": "grammar"}}}
		}},
		{"untyped-tool", func(b map[string]any) { b["tools"] = []any{map[string]any{"name": "x"}} }},
		{"web-search", func(b map[string]any) { b["web_search_options"] = map[string]any{} }},
		{"file-id", func(b map[string]any) {
			b["messages"] = []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "file",
				"file": map[string]any{"file_id": "file_123"}}}}}
		}},
		{"stored-audio", func(b map[string]any) {
			b["messages"] = append(b["messages"].([]any), map[string]any{"role": "assistant", "audio": map[string]any{"id": "audio_123"}})
		}},
	}
	// Every refusal holds on the full and the simple entry alike (spec: no
	// lower-level bypass).
	for _, c := range widen {
		for _, simple := range []bool{false, true} {
			entry := map[bool]string{false: "full", true: "simple"}[simple]
			t.Run("cannot-widen-"+c.name+"/"+entry, func(t *testing.T) {
				ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-cannot-widen-"+c.name+"-"+entry, plain, raw)
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
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-hosted-tool-the-binding-allows", plain, raw)
		w.updateBindingOf(tenantA, "chat", func(b *ai.Binding) { b.AllowedHostedTools = []string{"web_search_options"} })
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["web_search_options"] = map[string]any{"search_context_size": "low"}
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-hosted"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check("the allowed hosted web search is sent", body["web_search_options"] != nil, "got %v", body)
	})

	t.Run("response-callback-failure-before-start", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-response-callback-failure-before-start", plain, raw)
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

	t.Run("payload-callback-failure", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-payload-callback-failure", plain, raw)
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
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-header-transform", plain, raw)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("X-Request-Tag", "host-1")
			h.Del("User-Agent")
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-headers"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		r := lastRequest(ev, w)
		ev.Check("an added header is sent", r.Header["X-Request-Tag"] == "host-1", "got %q", r.Header["X-Request-Tag"])
		_, ua := r.Header["User-Agent"]
		ev.Check("a removed header is not sent, not even the SDK's default", !ua, "got %q", r.Header["User-Agent"])
		ev.Check("the tenant's key still authenticates", r.KeyAlias == tenantA.alias, "got %q", r.KeyAlias)
	})

	t.Run("header-transform-cannot-change-the-key", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P04-E04-callbacks-header-transform-cannot-change-the-key", plain, raw)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Authorization", "Bearer "+tenantB.secret)
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-key"), plain.target(), plain.request(t), nil)
		ev.Check("tenant_denied", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}), "got %v", err)
		ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})
}

// TestChatNoEnvironmentFallback is H1 on the Chat path: everything the
// OpenAI SDK's default client options read from the environment (key, base
// URL, organization, project) points at a decoy, and the call still uses
// only the binding's endpoint and the tenant's key.
func TestChatNoEnvironmentFallback(t *testing.T) {
	f, raw := loadFixture(t, chatProtocol, "text.json")
	sc := scenarioByID(t, f, "text")
	ev := run.Case(t, "P04-H1-no-environment-fallback")
	decoy := polluteEnvironment(t)
	t.Setenv("OPENAI_ORG_ID", "org-env-leak")
	t.Setenv("OPENAI_PROJECT_ID", "proj-env-leak")
	w, o := runScenario(t, ev, sc, raw, modeComplete)
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	ev.Check("environment endpoint not used", len(decoy.Requests()) == 0, "decoy got %d", len(decoy.Requests()))
	r := lastRequest(ev, w)
	ev.Check("the tenant's key, not the environment's", r.KeyAlias == tenantA.alias, "got %q", r.KeyAlias)
	for _, name := range []string{"OpenAI-Organization", "OpenAI-Project"} {
		_, leaked := r.Header[http.CanonicalHeaderKey(name)]
		ev.Check("no environment "+name+" header", !leaked, "got %q", r.Header[http.CanonicalHeaderKey(name)])
	}
}

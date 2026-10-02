package e2e

// The suites in this file parallel their Anthropic counterparts case for
// case on the shared helpers, with Gemini's data and its deliberate
// differences (ADR-0012); sharing one protocol-parametric suite is issue 30.

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// TestGeminiCallbacks covers the trusted callbacks on the Gemini path (spec
// I7 Gemini row, E04): the header transform, then the payload callback over
// the REST body (ADR-0012), then start — the response callback never runs,
// as in pi — a replaced or changed payload sent as given, the authorization
// a callback cannot widen, and the header transform's guarded credential.
func TestGeminiCallbacks(t *testing.T) {
	text, raw := loadFixture(t, geminiProtocol, "text.json")
	plain := scenarioByID(t, text, "text")

	t.Run("order-and-inputs", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-order-and-inputs", plain, raw)
		var log callLog
		var seenHeader http.Header
		var seenPayload ai.Payload
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
			OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
				log.add("response")
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
		want := []string{"headers", "payload", "event:start"}
		ev.Check("headers, then payload, then start", len(got) >= 3 && reflect.DeepEqual(got[:3], want), "got %q", got)
		for _, e := range got {
			ev.Check("the response callback never runs, as in pi", e != "response", "got %q", got)
		}
		ev.Check("transform sees the API key, redacted", seenHeader.Get("X-Goog-Api-Key") == "[REDACTED]", "got %q", seenHeader.Get("X-Goog-Api-Key"))
		ev.Check("payload identifies the model", seenPayload.ProviderID == ai.ProviderGoogle &&
			seenPayload.API == ai.APIGoogleGenerativeAI && seenPayload.ModelID == plain.Model, "got %+v", seenPayload)
		ev.Check("payload body is the REST request body", jsonEqual(mustMarshal(t, seenPayload.Body), plain.Expect.Request.Body),
			"got %s", mustMarshal(t, seenPayload.Body))
	})

	t.Run("response-callback-error-ignored", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-response-callback-error-ignored", plain, raw)
		hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
			return errors.New("never called")
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-response"), plain.target(), plain.request(t), nil)
		ev.Record("result", res)
		ev.Check("call succeeds: the response callback is not part of this protocol", err == nil && res.Message.StopReason == ai.StopReasonStop,
			"err=%v", err)
	})

	t.Run("replaced-payload", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-replaced-payload", plain, raw)
		calls := 0
		hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
			calls++
			return ai.ReplacePayload(map[string]any{
				"contents":         []any{map[string]any{"role": "user", "parts": []any{map[string]any{"text": "Replaced."}}}},
				"generationConfig": map[string]any{"maxOutputTokens": 10},
			}), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-replace"), plain.target(), plain.request(t), ai.GeminiOptions{})
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("payload callback ran once", calls == 1, "got %d", calls)
		r := lastRequest(ev, w)
		ev.Check("the replacement is sent as given", jsonEqual(r.Body,
			[]byte(`{"contents":[{"role":"user","parts":[{"text":"Replaced."}]}],"generationConfig":{"maxOutputTokens":10}}`)), "got %s", r.Body)
		ev.Check("to the authorized model's URL", r.Path == "/v1beta/models/gemini-2.5-flash:streamGenerateContent" && r.Query == "alt=sse",
			"got %s?%s", r.Path, r.Query)
	})

	t.Run("changed-in-place", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-changed-in-place", plain, raw)
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["generationConfig"].(map[string]any)["topK"] = 3
			p.Body["safetySettings"] = []any{map[string]any{"category": "HARM_CATEGORY_HARASSMENT", "threshold": "BLOCK_NONE"}}
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-in-place"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check("the in-place changes are sent", body["generationConfig"].(map[string]any)["topK"] == 3.0 && body["safetySettings"] != nil,
			"got %v", body)
	})

	widen := []struct {
		name   string
		change func(map[string]any)
	}{
		{"model", func(b map[string]any) { b["model"] = "models/gemini-2.5-pro" }},
		{"cached-content", func(b map[string]any) { b["cachedContent"] = "cachedContents/abc123" }},
		{"mcp-servers", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"mcpServers": []any{map[string]any{"name": "x", "streamableHttpTransport": map[string]any{"url": "https://mcp.example"}}}}}
		}},
		{"file-search", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"fileSearch": map[string]any{"fileSearchStoreNames": []any{"fileSearchStores/s1"}}}}
		}},
		{"hosted-tool", func(b map[string]any) { b["tools"] = []any{map[string]any{"googleSearch": map[string]any{}}} }},
		{"file-data", func(b map[string]any) {
			b["contents"] = []any{map[string]any{"role": "user", "parts": []any{map[string]any{
				"fileData": map[string]any{"mimeType": "application/pdf", "fileUri": "https://generativelanguage.googleapis.com/v1beta/files/abc"}}}}}
		}},
	}
	// Every refusal holds on the full and the simple entry alike (spec: no
	// lower-level bypass).
	for _, c := range widen {
		for _, simple := range []bool{false, true} {
			entry := map[bool]string{false: "full", true: "simple"}[simple]
			t.Run("cannot-widen-"+c.name+"/"+entry, func(t *testing.T) {
				ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-cannot-widen-"+c.name+"-"+entry, plain, raw)
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
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-hosted-tool-the-binding-allows", plain, raw)
		b := w.host.Binding(tenantA.tenant, "gemini")
		b.AllowedHostedTools = []string{"googleSearch"}
		w.host.PutBindingAt(tenantA.tenant, "gemini", b)
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["tools"] = []any{map[string]any{"googleSearch": map[string]any{}}}
			return ai.KeepPayload(), nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-hosted"), plain.target(), plain.request(t), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		var body map[string]any
		mustUnmarshal(t, lastRequest(ev, w).Body, &body)
		ev.Check("the allowed hosted tool is sent", len(body["tools"].([]any)) == 1, "got %v", body["tools"])
	})

	t.Run("payload-callback-failure", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-payload-callback-failure", plain, raw)
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
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-header-transform", plain, raw)
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
		ev.Check("a removed header is not sent, not even the HTTP stack's default", !ua, "got %q", r.Header["User-Agent"])
		ev.Check("the tenant's key still authenticates", r.KeyAlias == googleA.alias, "got %q", r.KeyAlias)
	})

	t.Run("header-transform-cannot-change-the-key", func(t *testing.T) {
		ev, w := startScenarioCallbackCase(t, "P03-E04-callbacks-header-transform-cannot-change-the-key", plain, raw)
		hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("X-Goog-Api-Key", googleB.secret)
			return nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-cb-key"), plain.target(), plain.request(t), nil)
		ev.Check("tenant_denied", errors.Is(err, &ai.Error{Code: ai.CodeTenantDenied}), "got %v", err)
		ev.Check("nothing was sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	})
}

// TestGeminiNoEnvironmentFallback is H1 on the Gemini path: everything the
// Google SDKs read from the environment (GOOGLE_API_KEY, GEMINI_API_KEY,
// GOOGLE_GEMINI_BASE_URL, the Vertex switch and project) points at a decoy,
// and the call still uses only the binding's endpoint and the tenant's key.
func TestGeminiNoEnvironmentFallback(t *testing.T) {
	f, raw := loadFixture(t, geminiProtocol, "text.json")
	sc := scenarioByID(t, f, "text")
	ev := run.Case(t, "P03-H1-no-environment-fallback")
	decoy := polluteEnvironment(t)
	t.Setenv("GOOGLE_GENAI_USE_VERTEXAI", "true")
	t.Setenv("GOOGLE_CLOUD_PROJECT", "env-leak")
	w, o := runScenario(t, ev, sc, raw, modeComplete)
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	ev.Check("environment endpoint not used", len(decoy.Requests()) == 0, "decoy got %d", len(decoy.Requests()))
	ev.Check("the tenant's key, not the environment's", lastRequest(ev, w).KeyAlias == googleA.alias, "got %q", lastRequest(ev, w).KeyAlias)
}

package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// errHostCallback is a trusted host callback's own failure.
var errHostCallback = errors.New("host callback: internal detail db=10.0.0.7")

// TestTrustedCallbackFailures covers callbacks that fail or outlive the
// call's context: each ends the call in the phase it runs in, and nothing
// is sent when a pre-send callback fails (spec I7).
func TestTrustedCallbackFailures(t *testing.T) {
	failing := []struct {
		name     string
		hooks    func(cancel context.CancelFunc) ai.Hooks
		requests int // provider requests expected
		code     ai.Code
		stop     ai.StopReason
	}{
		{"headers-error", func(context.CancelFunc) ai.Hooks {
			return ai.Hooks{TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { return errHostCallback }}
		}, 0, ai.CodeInvalidRequest, ai.StopReasonError},
		{"payload-error", func(context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				return ai.KeepPayload(), errHostCallback
			}}
		}, 0, ai.CodeInvalidRequest, ai.StopReasonError},
		{"response-error", func(context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { return errHostCallback }}
		}, 1, ai.CodeInvalidRequest, ai.StopReasonError},
		{"headers-canceled", func(cancel context.CancelFunc) ai.Hooks {
			return ai.Hooks{TransformHeaders: func(ctx context.Context, _ ai.CallScope, _ http.Header) error {
				cancel()
				<-ctx.Done()
				return ctx.Err()
			}}
		}, 0, ai.CodeCanceled, ai.StopReasonAborted},
		{"payload-canceled", func(cancel context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnPayload: func(ctx context.Context, _ ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
				cancel()
				<-ctx.Done()
				// A callback that reports success after the context ended
				// still does not get the request sent.
				return ai.KeepPayload(), nil
			}}
		}, 0, ai.CodeCanceled, ai.StopReasonAborted},
		{"response-canceled", func(cancel context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				cancel()
				<-ctx.Done()
				return ctx.Err()
			}}
		}, 1, ai.CodeCanceled, ai.StopReasonAborted},
		{"payload-replaced-with-nothing", func(context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				return ai.ReplacePayload(nil), nil
			}}
		}, 0, ai.CodeInvalidRequest, ai.StopReasonError},
		{"payload-not-encodable", func(context.CancelFunc) ai.Hooks {
			return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["temperature"] = make(chan int)
				return ai.KeepPayload(), nil
			}}
		}, 0, ai.CodeInvalidRequest, ai.StopReasonError},
	}
	for _, c := range failing {
		t.Run(c.name, func(t *testing.T) {
			ev, w, f := startCallbackCase(t, "P01-E04-callbacks-fail-"+c.name)
			ctx, cancel := context.WithCancel(ctxFor(t))
			defer cancel()
			s := w.client.WithHooks(c.hooks(cancel)).Stream(ctx, textScope("req-fail-"+c.name), textTarget(f), textRequest(f), nil)
			var events []ai.Event
			for s.Next() {
				events = append(events, s.Event())
			}
			res, err := s.Result()
			ev.Record("events", eventsJSON(events))
			ev.Record("result", res)
			ev.Record("requests", w.provider.Requests())

			var aiErr *ai.Error
			ev.Check("classified error", errors.As(err, &aiErr), "err=%v", err)
			if aiErr != nil {
				ev.Check("code", aiErr.Code == c.code, "got %q want %q", aiErr.Code, c.code)
				ev.Check("phase is the request phase", aiErr.Phase == ai.PhaseRequest, "got %q", aiErr.Phase)
			}
			ev.Check("stop reason", res.Message.StopReason == c.stop, "got %q want %q", res.Message.StopReason, c.stop)
			ev.Check("error message set", res.Message.ErrorMessage != "", "empty")
			if c.code == ai.CodeInvalidRequest && c.name != "payload-replaced-with-nothing" && c.name != "payload-not-encodable" {
				ev.Check("host error reachable for diagnosis", errors.Is(err, errHostCallback), "err=%v", err)
			}
			ev.Check("host error text not surfaced", !strings.Contains(res.Message.ErrorMessage, "10.0.0.7"), "got %q", res.Message.ErrorMessage)
			ev.Check("provider requests", len(w.provider.Requests()) == c.requests, "got %d want %d", len(w.provider.Requests()), c.requests)
			ev.Check("no start before the callbacks succeeded", !hasEvent(events, "start"), "events=%q", eventSummary(events))
			ev.Check("one terminal event", len(events) == 1 && string(events[0].Type()) == "error", "events=%q", eventSummary(events))
		})
	}
}

// TestTrustedCallbackAuthorization covers callbacks trying to change the
// binding's authentication, target or model, or to reference native state
// the library cannot authorize: each is refused and nothing is sent.
func TestTrustedCallbackAuthorization(t *testing.T) {
	headerAttempts := []struct {
		name   string
		change func(h http.Header)
	}{
		{"authorization-rewritten", func(h http.Header) { h.Set("Authorization", "Bearer "+tenantB.secret) }},
		{"authorization-removed", func(h http.Header) { h.Del("Authorization") }},
		{"authorization-added-twice", func(h http.Header) { h.Add("Authorization", "Bearer "+tenantB.secret) }},
		{"host-overridden", func(h http.Header) { h.Set("Host", "attacker.example") }},
		{"alternate-key-header", func(h http.Header) { h.Set("X-Api-Key", tenantB.secret) }},
		{"proxy-authorization", func(h http.Header) { h.Set("Proxy-Authorization", "Basic eDp5") }},
		{"openai-project", func(h http.Header) { h.Set("OpenAI-Project", "proj-other") }},
		{"openai-organization", func(h http.Header) { h.Set("OpenAI-Organization", "org-other") }},
		{"cache-session-added", func(h http.Header) { h.Set("session_id", "other-tenant-session") }},
	}
	for _, a := range headerAttempts {
		t.Run("headers-"+a.name, func(t *testing.T) {
			ev, w, f := startCallbackCase(t, "P01-E04-callbacks-deny-headers-"+a.name)
			hooks := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
				a.change(h)
				return nil
			}}
			res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-deny-"+a.name), textTarget(f), textRequest(f), nil)
			checkRefused(ev, w, res, err)
		})
	}

	payloadAttempts := []struct {
		name   string
		change func(body map[string]any)
	}{
		{"model-switched", func(b map[string]any) { b["model"] = "gpt-4o-mini" }},
		{"model-unauthorized", func(b map[string]any) { b["model"] = "gpt-4.1" }},
		{"model-removed", func(b map[string]any) { delete(b, "model") }},
		{"previous-response", func(b map[string]any) { b["previous_response_id"] = "resp_other_tenant" }},
		{"conversation", func(b map[string]any) { b["conversation"] = "conv_other_tenant" }},
		{"stored-prompt", func(b map[string]any) { b["prompt"] = map[string]any{"id": "pmpt_other_tenant"} }},
		{"background", func(b map[string]any) { b["background"] = true }},
		{"stored", func(b map[string]any) { b["store"] = true }},
		{"cache-key-added", func(b map[string]any) { b["prompt_cache_key"] = "other-tenant-cache" }},
		{"hosted-tool", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "file_search", "vector_store_ids": []any{"vs_other_tenant"}}}
		}},
		{"mcp-tool", func(b map[string]any) {
			b["tools"] = []any{map[string]any{"type": "mcp", "server_label": "x", "server_url": "https://attacker.example/mcp"}}
		}},
		{"file-id", func(b map[string]any) {
			b["input"] = append(b["input"].([]any), map[string]any{"role": "user", "content": []any{
				map[string]any{"type": "input_file", "file_id": "file_other_tenant"}}})
		}},
		{"item-reference", func(b map[string]any) {
			b["input"] = append(b["input"].([]any), map[string]any{"type": "item_reference", "id": "rs_other_tenant"})
		}},
	}
	for _, a := range payloadAttempts {
		t.Run("payload-"+a.name, func(t *testing.T) {
			ev, w, f := startCallbackCase(t, "P01-E04-callbacks-deny-payload-"+a.name)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				a.change(p.Body)
				return ai.KeepPayload(), nil
			}}
			res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-deny-"+a.name), textTarget(f), textRequest(f), nil)
			checkRefused(ev, w, res, err)
		})
	}
}

// checkRefused asserts an authorization refusal that sent nothing.
func checkRefused(ev *evidence.Case, w *world, res ai.Result, err error) {
	ev.Record("result", res)
	ev.Record("requests", w.provider.Requests())
	checkFailed(ev, res, err)
	var aiErr *ai.Error
	if errors.As(err, &aiErr) {
		ev.Check("refused as tenant_denied in the request phase",
			aiErr.Code == ai.CodeTenantDenied && aiErr.Phase == ai.PhaseRequest, "got %q/%q", aiErr.Code, aiErr.Phase)
	}
	ev.Check("nothing sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
	ev.Check("no secret in the message", !strings.Contains(res.Message.ErrorMessage, tenantB.secret), "got %q", res.Message.ErrorMessage)
}

// TestTrustedCallbackTenantIsolation runs two tenants' calls concurrently,
// each with its own callbacks: every callback sees only its own call's scope
// and only its own changes reach the provider (spec I7, E04 cross-tenant).
func TestTrustedCallbackTenantIsolation(t *testing.T) {
	ev := run.Case(t, "P01-E04-callbacks-tenant-isolation")
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("fixture.json", raw)
	w := newWorld(t, tenantA, tenantB)
	const perTenant = 8
	for range 2 * perTenant {
		w.provider.Enqueue(sseReply(t, f, provider.FramingLF))
	}

	var mu sync.Mutex
	var mismatches []string
	hooksFor := func(k tenantKey) ai.Hooks {
		check := func(where string, s ai.CallScope) {
			if s.TenantID != k.tenant {
				mu.Lock()
				mismatches = append(mismatches, fmt.Sprintf("%s hook of %s saw %s", where, k.tenant, s.TenantID))
				mu.Unlock()
			}
		}
		return ai.Hooks{
			TransformHeaders: func(_ context.Context, s ai.CallScope, h http.Header) error {
				check("headers", s)
				h.Set("X-Hook-Tenant", k.tenant)
				return nil
			},
			OnPayload: func(_ context.Context, s ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				check("payload", s)
				p.Body["metadata"] = map[string]any{"hook_tenant": k.tenant}
				return ai.KeepPayload(), nil
			},
			OnResponse: func(_ context.Context, s ai.CallScope, _ ai.ResponseInfo) error {
				check("response", s)
				return nil
			},
		}
	}

	var wg sync.WaitGroup
	errs := make(chan error, 2*perTenant)
	for _, k := range []tenantKey{tenantA, tenantB} {
		hc := w.client.WithHooks(hooksFor(k))
		for i := range perTenant {
			wg.Go(func() {
				scope := ai.CallScope{TenantID: k.tenant, RequestID: fmt.Sprintf("req-%s-%d", k.tenant, i)}
				_, err := hc.Complete(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
				errs <- err
			})
		}
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		ev.Check("call succeeds", err == nil, "err=%v", err)
	}
	ev.Record("scope-mismatches", mismatches)
	ev.Check("every callback saw only its own tenant", len(mismatches) == 0, "got %q", mismatches)

	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	ev.Check("every call reached the provider", len(reqs) == 2*perTenant, "got %d", len(reqs))
	for _, r := range reqs {
		var body struct {
			Metadata map[string]string `json:"metadata"`
		}
		mustUnmarshal(t, r.Body, &body)
		header, payload := r.Header["X-Hook-Tenant"], body.Metadata["hook_tenant"]
		want := tenantA.tenant
		if r.KeyAlias == tenantB.alias {
			want = tenantB.tenant
		}
		ev.Check("header change belongs to the key's tenant", header == want, "key=%s header=%s", r.KeyAlias, header)
		ev.Check("payload change belongs to the key's tenant", payload == want, "key=%s payload=%s", r.KeyAlias, payload)
	}
}

func hasEvent(events []ai.Event, typ string) bool {
	for _, e := range events {
		if string(e.Type()) == typ {
			return true
		}
	}
	return false
}

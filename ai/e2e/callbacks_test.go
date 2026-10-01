package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestTrustedCallbacks covers a trusted host's per-call header transform,
// payload callback and response callback: their order, what each sees and
// what each may change (spec I7, E04 callbacks, User Story 35).
func TestTrustedCallbacks(t *testing.T) {
	t.Run("order-and-inputs", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-order-and-inputs")
		var log callLog
		var seenHeader http.Header
		var seenPayload ai.Payload
		var seenResponse ai.ResponseInfo
		scope := textScope("req-callbacks-order")
		hooks := ai.Hooks{
			TransformHeaders: func(_ context.Context, s ai.CallScope, h http.Header) error {
				log.add("headers:" + s.TenantID + "/" + s.RequestID)
				seenHeader = h.Clone()
				h.Set("X-Host-Trace", "trace-1")
				return nil
			},
			OnPayload: func(_ context.Context, s ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				log.add("payload:" + s.TenantID + "/" + s.RequestID)
				seenPayload = *p
				return ai.KeepPayload(), nil
			},
			OnResponse: func(_ context.Context, s ai.CallScope, r ai.ResponseInfo) error {
				log.add("response:" + s.TenantID + "/" + s.RequestID)
				seenResponse = r
				return nil
			},
		}
		s := w.client.WithHooks(hooks).Stream(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
		for s.Next() {
			log.add("event:" + string(s.Event().Type()))
		}
		res, err := s.Result()
		ev.Record("result", res)
		ev.Record("callback-log", log.entries())
		ev.Check("call succeeds", err == nil, "err=%v", err)

		want := []string{
			"headers:tenant-a/req-callbacks-order",
			"payload:tenant-a/req-callbacks-order",
			"response:tenant-a/req-callbacks-order",
			"event:start",
		}
		got := log.entries()
		ev.Check("headers, then payload, then response, then start",
			len(got) >= len(want) && reflect.DeepEqual(got[:len(want)], want), "got %q", got)

		// The header transform sees the merged authentication and request
		// headers, with the credential redacted: the secret never reaches it.
		ev.Record("transform-headers-input", seenHeader)
		ev.Check("transform sees the authentication header, redacted",
			seenHeader.Get("Authorization") == "[REDACTED]", "got %q", seenHeader.Get("Authorization"))
		ev.Check("transform sees the request headers", seenHeader.Get("User-Agent") == "barness-ai", "got %q", seenHeader.Get("User-Agent"))

		// The payload callback sees the native body the adapter built.
		ev.Check("payload identifies the model",
			seenPayload.ProviderID == ai.ProviderOpenAI && seenPayload.API == ai.APIOpenAIResponses && seenPayload.ModelID == f.Model,
			"got %+v", seenPayload)
		ev.Check("payload body is the native request body",
			jsonEqual(mustMarshal(t, seenPayload.Body), f.Expect.Request.Body), "got %s", mustMarshal(t, seenPayload.Body))

		ev.Record("response-info", seenResponse)
		ev.Check("response callback reads the HTTP status", seenResponse.Status == http.StatusOK, "got %d", seenResponse.Status)
		ev.Check("response callback reads the HTTP headers",
			seenResponse.Header.Get("X-Request-Id") == "req_vendor_42", "got %q", seenResponse.Header.Get("X-Request-Id"))
		ev.Check("response callback identifies the model",
			seenResponse.ProviderID == ai.ProviderOpenAI && seenResponse.API == ai.APIOpenAIResponses && seenResponse.ModelID == f.Model,
			"got %+v", seenResponse)

		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if ev.Check("exactly one request", len(reqs) == 1, "got %d", len(reqs)) {
			ev.Check("transformed header sent", reqs[0].Header["X-Host-Trace"] == "trace-1", "got %q", reqs[0].Header["X-Host-Trace"])
			ev.Check("tenant key still used", reqs[0].KeyAlias == tenantA.alias, "got %q", reqs[0].KeyAlias)
			ev.Check("observed payload sent unchanged", jsonEqual(reqs[0].Body, f.Expect.Request.Body), "got %s", reqs[0].Body)
		}
		checkFinalMessage(ev, f, res.Message)
	})

	t.Run("payload-in-place-kept", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-payload-in-place")
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["temperature"] = json.Number("0.25")
			p.Body["metadata"] = map[string]any{"host": "x"}
			// Not replacing keeps the in-place changes.
			return ai.KeepPayload(), nil
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-in-place"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		checkSentBody(ev, w, withFields(t, f.Expect.Request.Body, map[string]any{"temperature": 0.25, "metadata": map[string]any{"host": "x"}}))
	})

	t.Run("payload-zero-decision-keeps", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-payload-zero-decision")
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			p.Body["temperature"] = json.Number("1")
			return ai.PayloadDecision{}, nil
		}}
		_, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-zero-decision"), textTarget(f), textRequest(f), nil)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		checkSentBody(ev, w, withFields(t, f.Expect.Request.Body, map[string]any{"temperature": 1}))
	})

	t.Run("payload-replaced", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-payload-replaced")
		var replacement map[string]any
		mustUnmarshal(t, f.Expect.Request.Body, &replacement)
		replacement["max_output_tokens"] = 77
		hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
			// The replacement wins over in-place changes to the original.
			p.Body["temperature"] = 2
			return ai.ReplacePayload(replacement), nil
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-replaced"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		checkSentBody(ev, w, withFields(t, f.Expect.Request.Body, map[string]any{"max_output_tokens": 77}))
	})

	t.Run("response-once-message-untouched", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-response-once")
		calls := 0
		hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
			calls++
			return nil
		}}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-response-once"), textTarget(f), textRequest(f), nil)
		ev.Record("result", res)
		ev.Check("call succeeds", err == nil, "err=%v", err)
		ev.Check("response callback runs once per call, not per token", calls == 1, "got %d", calls)
		checkFinalMessage(ev, f, res.Message)
	})

	t.Run("simple-entries-run-callbacks", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-simple-entries")
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))
		var log callLog
		hooks := ai.Hooks{
			TransformHeaders: func(_ context.Context, s ai.CallScope, _ http.Header) error {
				log.add("headers:" + s.RequestID)
				return nil
			},
			OnPayload: func(_ context.Context, s ai.CallScope, _ *ai.Payload) (ai.PayloadDecision, error) {
				log.add("payload:" + s.RequestID)
				return ai.KeepPayload(), nil
			},
			OnResponse: func(_ context.Context, s ai.CallScope, _ ai.ResponseInfo) error {
				log.add("response:" + s.RequestID)
				return nil
			},
		}
		hc := w.client.WithHooks(hooks)
		_, err := hc.CompleteSimple(ctxFor(t), textScope("req-cs"), textTarget(f), textRequest(f), ai.SimpleOptions{})
		ev.Check("CompleteSimple succeeds", err == nil, "err=%v", err)
		s := hc.StreamSimple(ctxFor(t), textScope("req-ss"), textTarget(f), textRequest(f), ai.SimpleOptions{})
		_, err = s.Result()
		ev.Check("StreamSimple succeeds", err == nil, "err=%v", err)
		want := []string{"headers:req-cs", "payload:req-cs", "response:req-cs", "headers:req-ss", "payload:req-ss", "response:req-ss"}
		ev.Record("callback-log", log.entries())
		ev.Check("every callback ran once per call", reflect.DeepEqual(log.entries(), want), "got %q", log.entries())
	})

	t.Run("session-cache-identity-kept", func(t *testing.T) {
		ev, w, f := startCallbackCase(t, "P01-E04-callbacks-session-cache-identity")
		hooks := ai.Hooks{
			TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
				h.Set("X-Host-Trace", "trace-2")
				return nil
			},
			OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				p.Body["temperature"] = json.Number("0.5")
				return ai.KeepPayload(), nil
			},
		}
		opts := ai.ResponsesOptions{SessionID: "session-1"}
		res, err := w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("req-session"), textTarget(f), textRequest(f), opts)
		ev.Record("result", res)
		ev.Check("call keeping the derived cache identity succeeds", err == nil, "err=%v", err)
		reqs := w.provider.Requests()
		ev.Record("requests", reqs)
		if ev.Check("exactly one request", len(reqs) == 1, "got %d", len(reqs)) {
			var body map[string]any
			mustUnmarshal(t, reqs[0].Body, &body)
			key, _ := body["prompt_cache_key"].(string)
			ev.Check("derived cache key sent in body and headers", key != "" && reqs[0].Header["Session_id"] == key && reqs[0].Header["X-Client-Request-Id"] == key,
				"body=%q headers=%q/%q", key, reqs[0].Header["Session_id"], reqs[0].Header["X-Client-Request-Id"])
			ev.Check("transformed header sent", reqs[0].Header["X-Host-Trace"] == "trace-2", "got %q", reqs[0].Header["X-Host-Trace"])
		}
	})

	t.Run("no-per-call-transport", func(t *testing.T) {
		ev := run.Case(t, "P01-E04-callbacks-no-per-call-transport")
		// Networking is assembled only in ai.Config; nothing a call carries
		// can hold a transport, HTTP client or endpoint.
		var offending []string
		for _, typ := range []reflect.Type{reflect.TypeFor[ai.Hooks](), reflect.TypeFor[ai.Request](), reflect.TypeFor[ai.CallScope](),
			reflect.TypeFor[ai.Target](), reflect.TypeFor[ai.ResponsesOptions](), reflect.TypeFor[ai.SimpleOptions]()} {
			for i := range typ.NumField() {
				field := typ.Field(i)
				if carriesNetworking(field.Type) {
					offending = append(offending, typ.Name()+"."+field.Name)
				}
			}
		}
		ev.Record("offending-fields", offending)
		ev.Check("no per-call transport entry", len(offending) == 0, "got %q", offending)
	})
}

// carriesNetworking reports whether a field type can hold a transport or an
// HTTP client.
func carriesNetworking(t reflect.Type) bool {
	roundTripper := reflect.TypeFor[http.RoundTripper]()
	for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
		t = t.Elem()
	}
	return t == reflect.TypeFor[http.Client]() || t.Implements(roundTripper) || (t.Kind() == reflect.Interface && roundTripper.Implements(t) && t.NumMethod() > 0)
}

// startCallbackCase is startTextCase with a vendor request id on the reply.
func startCallbackCase(t *testing.T, caseID string) (*evidence.Case, *world, textFixture) {
	t.Helper()
	ev, w, f := startHardeningCase(t, caseID)
	reply := sseReply(t, f, provider.FramingLF)
	reply.Header["X-Request-Id"] = "req_vendor_42"
	enqueue(ev, w, reply)
	return ev, w, f
}

// checkSentBody asserts the single request the provider received carried want.
func checkSentBody(ev *evidence.Case, w *world, want []byte) {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if ev.Check("exactly one request", len(reqs) == 1, "got %d", len(reqs)) {
		ev.Check("sent body", jsonEqual(reqs[0].Body, want), "got %s\nwant %s", reqs[0].Body, want)
	}
}

// withFields is body with fields set.
func withFields(t *testing.T, body json.RawMessage, fields map[string]any) []byte {
	t.Helper()
	var m map[string]any
	mustUnmarshal(t, body, &m)
	for k, v := range fields {
		m[k] = v
	}
	return mustMarshal(t, m)
}

// callLog is an ordered, concurrency-safe record of what callbacks and the
// consumer observed.
type callLog struct {
	mu  sync.Mutex
	log []string
}

func (l *callLog) add(s string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.log = append(l.log, s)
}

func (l *callLog) entries() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string(nil), l.log...)
}

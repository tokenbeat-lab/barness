package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

func TestXAIRuntimeHTTPFailures(t *testing.T) {
	for _, status := range []int{401, 403, 429, 408, 409, 500, 503} {
		for _, retries := range []int{0, 1} {
			t.Run(fmt.Sprintf("%d/retries-%d", status, retries), func(t *testing.T) {
				ev := run.Case(t, "XAI-E14-http-"+t.Name())
				rec := newRecorder()
				w := newXAIRuntimeWorld(t, nil, func(c *ai.Config) { c.Observer = rec; c.Policy.MaxQueuedObservations = 64 })
				b := xaiBinding()
				b.Retry.MaxRetries = retries
				w.host.PutBinding(b)
				enqueue(ev, w.world, xaiHTTPFailure(status), xaiReply(t))
				scope := textScope(strings.ReplaceAll(t.Name(), "/", "-"))
				o := drain(w.client.Stream(ctxFor(t), scope, xaiTarget(), xaiRequest(), nil))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				wantRetry := retries == 1 && status != 401 && status != 403
				wantAttempts := 1
				if wantRetry {
					wantAttempts = 2
				}
				ev.Check("default and explicit retries are finite", len(o.result.Metadata.Attempts) == wantAttempts && len(w.provider.Requests()) == wantAttempts && (o.err == nil) == wantRetry, "err=%v metadata=%+v", o.err, o.result.Metadata)
				code := ai.CodeUpstreamError
				if status == 401 || status == 403 {
					code = ai.CodeUpstreamAuth
				}
				if status == 429 {
					code = ai.CodeRateLimited
				}
				if !wantRetry {
					var ae *ai.Error
					ev.Check("precise HTTP classification", errors.Is(o.err, &ai.Error{Code: code, Phase: ai.PhaseRequest}) && errors.As(o.err, &ae) && ae.HTTPStatus == status && ae.ProviderRequestID == "", "%v", o.err)
					ev.Check("upstream echo is absent from diagnostics", !strings.Contains(errString(o.err), "private-"), "%v", o.err)
				}
				if len(o.result.Metadata.Attempts) > 0 {
					a := o.result.Metadata.Attempts[0]
					ev.Check("attempt preserves status and safe classification", a.HTTPStatus == status && a.Code == code && a.ProviderRequestID == "", "%+v", a)
				}
				obs := rec.awaitCall(ev, scope.RequestID)
				checkCallRecords(ev, obs, scope, xaiTarget().BindingID, o, wantAttempts)
				ev.Check("Observer has no body or fabricated ID", !strings.Contains(string(mustMarshal(t, obs)), "private-") && !strings.Contains(string(mustMarshal(t, obs)), "unverified-"), "%+v", obs)
				w.released(ev)
			})
		}
	}
}

func TestXAIRuntimeRetrySnapshot(t *testing.T) {
	ev := run.Case(t, "XAI-E14-pinned-retry-callbacks")
	w := newXAIRuntimeWorld(t, nil)
	b := xaiBinding()
	b.Retry.MaxRetries = 1
	w.host.PutBinding(b)
	first := xaiHTTPFailure(503)
	first.OnReceive = func() {
		b.Version, b.Retry.MaxRetries = "v2", 0
		w.host.PutBinding(b)
		w.host.PutCredential(ai.Credential{OwnerTenantID: b.TenantID, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v2", Active: true, APIKey: ai.NewSecret(tenantA2.secret)})
	}
	enqueue(ev, w.world, first, xaiReply(t), xaiReply(t))
	var order []string
	h := ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, header http.Header) error {
		order = append(order, "headers")
		header.Set("X-Synthetic-Trace", "pinned")
		return nil
	}, OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		order = append(order, "payload")
		p.Body["temperature"] = 0.4
		return ai.KeepPayload(), nil
	}, OnResponse: func(_ context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
		order = append(order, "response")
		return nil
	}}
	scope := textScope("xai-pinned-call")
	s := w.client.WithHooks(h).Stream(ctxFor(t), scope, xaiTarget(), xaiRequest(), ai.ResponsesOptions{SessionID: "same-session"})
	var o outcome
	o.streamed = true
	for s.Next() {
		o.events = append(o.events, s.Event())
		ev.Check("domain events follow response policy callback", strings.Join(order, ",") == "headers,payload,response", "%v", order)
	}
	o.result, o.err = s.Result()
	o.streamErr = s.Err()
	o.record(ev)
	ev.Record("callback-order", order)
	ev.Check("snapshot completes with one callback sequence", o.err == nil && strings.Join(order, ",") == "headers,payload,response" && o.result.Metadata.BindingVersion == "v1" && o.result.Metadata.CredentialVersion == "v1", "%v %+v", o.err, o.result.Metadata)
	ev.Check("resolvers read once per logical call", w.host.ReadsFor(scope.RequestID) == (host.Reads{Bindings: 1, Credentials: 1}), "%+v", w.host.ReadsFor(scope.RequestID))
	future, err := w.client.Complete(ctxFor(t), textScope("xai-after-rotation"), xaiTarget(), xaiRequest(), ai.ResponsesOptions{SessionID: "same-session"})
	ev.Record("future-result", future)
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	if ev.Check("two pinned attempts and one new logical call", err == nil && len(reqs) == 3, "%v count=%d", err, len(reqs)) {
		ev.Check("retry bytes and bearer remain pinned", bytes.Equal(reqs[0].Body, reqs[1].Body) && reqs[0].KeyAlias == xaiKey.alias && reqs[1].KeyAlias == xaiKey.alias && reqs[2].KeyAlias == tenantA2.alias, "%+v", reqs)
		a, z := bodyFields(t, reqs[0].Body), bodyFields(t, reqs[2].Body)
		ev.Check("credential rotation preserves cache scope", bytes.Equal(a["prompt_cache_key"], z["prompt_cache_key"]), "%s %s", a["prompt_cache_key"], z["prompt_cache_key"])
		ev.Check("new call resolves rotated versions", future.Metadata.BindingVersion == "v2" && future.Metadata.CredentialVersion == "v2", "%+v", future.Metadata)
	}
	w.released(ev)
}

func TestXAIRuntimeRetryBounds(t *testing.T) {
	for _, mode := range []string{"exhausted", "delay-cap"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "XAI-E14-retry-"+mode)
			w := newXAIRuntimeWorld(t, nil)
			b := xaiBinding()
			b.Retry.MaxRetries = 2
			r := xaiHTTPFailure(503)
			want := 3
			if mode == "delay-cap" {
				cap := time.Millisecond
				b.Retry.MaxRetryDelay = &cap
				r.Header["Retry-After-Ms"] = "60000"
				want = 1
			}
			w.host.PutBinding(b)
			enqueue(ev, w.world, r, r, r)
			o := drain(w.client.Stream(ctxFor(t), textScope(t.Name()), xaiTarget(), xaiRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			ev.Check("retry cannot escape attempt or delay policy", errors.Is(o.err, &ai.Error{Code: ai.CodeUpstreamError, Phase: ai.PhaseRequest}) && len(o.result.Metadata.Attempts) == want && len(w.provider.Requests()) == want, "%v %+v", o.err, o.result.Metadata)
			ev.Check("delay refusal does not expose provider echo", !strings.Contains(errString(o.err), "private-"), "%v", o.err)
			w.released(ev)
		})
	}
}

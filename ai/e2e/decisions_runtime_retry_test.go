package e2e

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/clock"
)

func TestDecisionsRuntimeRetry(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		status, retries, attempts int
		header                    map[string]string
		delays                    []time.Duration
		cap                       *time.Duration
	}{
		{"default-429", 429, 0, 1, nil, nil, nil},
		{"default-503", 503, 0, 1, nil, nil, nil},
		{"default-504", 504, 0, 1, nil, nil, nil},
		{"429-ms", 429, 1, 2, map[string]string{"retry-after-ms": "10", "retry-after": "2"}, []time.Duration{10 * time.Millisecond}, nil},
		{"503-seconds", 503, 1, 2, map[string]string{"retry-after": "0.02"}, []time.Duration{20 * time.Millisecond}, nil},
		{"504-extension", 504, 1, 2, nil, []time.Duration{500 * time.Millisecond}, nil},
		{"header-false", 429, 2, 1, map[string]string{"x-should-retry": "false"}, nil, nil},
		{"header-true", 400, 1, 2, map[string]string{"x-should-retry": "true", "retry-after-ms": "1"}, []time.Duration{time.Millisecond}, nil},
		{"invalid-ms-falls-to-seconds", 503, 1, 2, map[string]string{"retry-after-ms": "Infinity", "retry-after": "0.02"}, []time.Duration{20 * time.Millisecond}, nil},
		{"date", 503, 1, 2, map[string]string{"retry-after": time.Unix(1790000000, 0).Add(time.Second).UTC().Format(http.TimeFormat)}, []time.Duration{time.Second}, nil},
		{"backoff", 503, 1, 2, map[string]string{"retry-after": "invalid"}, []time.Duration{500 * time.Millisecond}, nil},
		{"delay-over-cap", 429, 2, 1, map[string]string{"retry-after-ms": "21"}, nil, ptrDuration(20 * time.Millisecond)},
		{"delay-exact-cap", 429, 1, 2, map[string]string{"retry-after-ms": "20"}, []time.Duration{20 * time.Millisecond}, ptrDuration(20 * time.Millisecond)},
		{"exhausted", 503, 2, 3, nil, []time.Duration{500 * time.Millisecond, time.Second}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD40-runtime-"+tc.name)
			clk := clock.NewFake(time.Unix(1790000000, 0), 0)
			u := runtimeDecisionsWorld(t, func(c *ai.Config) { c.Clock = clk })
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry = ai.RetryPolicy{MaxRetries: tc.retries, MaxRetryDelay: tc.cap} })
			header := map[string]string{"x-request-id": "up-failed"}
			for k, v := range tc.header {
				header[k] = v
			}
			body := `{"error":{"message":"synthetic limited ` + tenantA.secret + `"}}`
			fail := fixtureReply{Status: tc.status, Header: header, Body: body}.script(t, decisionsProtocol, "")
			success := tc.attempts > 1 && tc.name != "exhausted"
			for i := 0; i < tc.attempts; i++ {
				reply := fail
				if success && i == tc.attempts-1 {
					reply = runtimeDecisionsReply(t)
				}
				enqueue(ev, u.world, reply)
			}
			counts := [3]int{}
			res, err := runtimeDecisionsCall(ctxFor(t), u, "req-retry-"+tc.name, decisionsInput(), ai.Hooks{
				TransformHeaders: func(context.Context, ai.CallScope, http.Header) error { counts[0]++; return nil },
				OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
					counts[1]++
					return ai.KeepPayload(), nil
				},
				OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error { counts[2]++; return nil },
			})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("expected", map[string]any{"status": tc.status, "retries": tc.retries, "attempts": tc.attempts, "header": tc.header, "delays": tc.delays, "maxRetryDelay": tc.cap})
			ev.Record("sleeps", clk.Sleeps())
			recordRuntimeConfig(ev, u)
			ev.Check("bounded explicit attempts", len(res.Metadata.Attempts) == tc.attempts && len(u.provider.Requests()) == tc.attempts, "got %+v", res.Metadata)
			ev.Check("exact delay policy", reflect.DeepEqual(clk.Sleeps(), tc.delays), "got %v want %v", clk.Sleeps(), tc.delays)
			wantResponse := 0
			if success {
				wantResponse = 1
				ev.Check("final success", err == nil && len(res.Answers) == 1, "got %v", err)
			} else {
				code := ai.CodeUpstreamError
				if tc.status == 429 {
					code = ai.CodeRateLimited
				}
				checkUnaryFailure(ev, res, err, code, ai.PhaseRequest, ai.StopReasonError)
				var ae *ai.Error
				if errors.As(err, &ae) {
					ev.Check("diagnostic and vendor identity survive redacted", ae.HTTPStatus == tc.status && ae.ProviderRequestID == "up-failed" && strings.Contains(ae.Message, "synthetic limited") && !strings.Contains(ae.Message, tenantA.secret), "diagnostic mismatch")
				}
			}
			ev.Check("callbacks once per logical call", counts == [3]int{1, 1, wantResponse}, "got %v", counts)
			for i, a := range res.Metadata.Attempts {
				if i < tc.attempts-1 || !success {
					ev.Check("failed attempts never claim usage", a.HTTPStatus == tc.status && a.ProviderRequestID == "up-failed" && a.UsageReporting == ai.UsageUnreported && a.Usage == (ai.Usage{}), "got %+v", a)
				}
				if i < len(tc.delays) {
					ev.Check("planned delay matches wait", a.RetryDelay == tc.delays[i], "got %+v", a)
				}
			}
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

func ptrDuration(d time.Duration) *time.Duration { return &d }

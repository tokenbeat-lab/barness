package e2e

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// OD41 uses Decisions itself with retries explicitly enabled. A failed read
// or metadata callback has not delivered a validated usage report; a damaged
// answer in a complete envelope retains the independently valid report.
func TestDecisionsUsageAfter2xx(t *testing.T) {
	for _, name := range []string{"malformed-json", "bad-answer", "refusal", "read-failure", "response-hook-error"} {
		for _, entry := range []string{"plain", "hooked"} {
			if name == "response-hook-error" && entry == "plain" {
				continue
			}
			t.Run(name+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "P10-OD41-usage-"+name+"-"+entry)
				u := newDecisionsWorld(t, nil)
				u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Retry.MaxRetries = 2 })
				body := `{"answers":[` + decisionsAnswer + `],` + decisionsFullUsage + `}`
				want := decisionsKnownUsage()
				want.Cost = ai.UsageCost{Input: .000296, Total: .000296}
				reporting := ai.UsageComplete
				code, phase := ai.CodeProtocol, ai.PhaseResponse
				switch name {
				case "malformed-json":
					body = "{"
					want, reporting = ai.Usage{}, ai.UsageUnreported
				case "bad-answer":
					body = `{"answers":[],` + decisionsFullUsage + `}`
				case "refusal":
					body = `{"answers":[{"type":"refusal","name":"intent"}],` + decisionsFullUsage + `}`
					code = ""
				case "read-failure":
					want, reporting = ai.Usage{}, ai.UsageUnreported
					code = ai.CodeTransport
				case "response-hook-error":
					want, reporting = ai.Usage{}, ai.UsageUnreported
					code, phase = ai.CodeCallbackFailed, ai.PhaseRequest
				}
				fixtureName := "response.json"
				if name == "malformed-json" {
					fixtureName = "response.txt"
				}
				ev.Fixture(fixtureName, []byte(body))
				reply := fixtureReply{Status: 200, Header: map[string]string{"x-request-id": "req-2xx"}, Body: body}.script(t, decisionsProtocol, "")
				if name == "read-failure" {
					reply.End = provider.EndAbort
				}
				enqueue(ev, u.world, reply)
				sentinel := &unaryHostError{}
				callbacks := 0
				hooks := ai.Hooks{OnResponse: func(context.Context, ai.CallScope, ai.ResponseInfo) error {
					callbacks++
					ev.Check("lease held at metadata boundary", u.gauges.Permits() == 1 && u.adm.Held() == 1 && u.bodies.Open() == 1, "lease missing")
					if name == "response-hook-error" {
						return sentinel
					}
					return nil
				}}
				res, err := decisionsPublicCall(t, u, entry, "req-2xx-"+name+"-"+entry, decisionsInput(), hooks)
				ev.Record("result", res)
				ev.Record("expected", map[string]any{"usage": want, "reporting": reporting, "code": code, "phase": phase, "maxRetries": 2})
				if code == "" {
					_, refusal := res.Answers["intent"].(ai.RefusalAnswer)
					ev.Check("valid refusal keeps reported consumption", err == nil && res.StopReason == ai.StopReasonStop && refusal, "got %v / %+v", err, res)
				} else {
					checkUnaryFailure(ev, res, err, code, phase, ai.StopReasonError)
				}
				if name == "response-hook-error" {
					ev.Check("trusted hook cause survives", errors.Is(err, sentinel), "cause missing")
				}
				ev.Check("validated consumption only", reflect.DeepEqual(res.Usage, want), "got %+v want %+v", res.Usage, want)
				if ev.Check("explicit retry cannot replay 2xx", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata) {
					a := res.Metadata.Attempts[0]
					ev.Check("attempt usage and HTTP identity", a.HTTPStatus == 200 && a.ProviderRequestID == "req-2xx" && a.RetryDelay == 0 && a.UsageReporting == reporting && reflect.DeepEqual(a.Usage, want), "got %+v", a)
				}
				wantCallbacks := 0
				if entry == "hooked" {
					wantCallbacks = 1
				}
				ev.Check("metadata hook executed once", callbacks == wantCallbacks, "got %d want %d", callbacks, wantCallbacks)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

package e2e

import (
	"errors"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsRuntimeGatewayDiagnostics(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"html", `<html>synthetic private gateway</html>`},
		{"malformed", `{"error":{"message":"synthetic private gateway"}`},
		{"duplicate-error", `{"error":{"message":"synthetic private gateway"},"error":{"message":"synthetic private gateway"}}`},
		{"duplicate-message", `{"error":{"message":"synthetic private gateway","message":"synthetic private gateway"}}`},
		{"non-string", `{"error":{"message":{"detail":"synthetic private gateway"}}}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD40-runtime-gateway-"+tc.name)
			u := runtimeDecisionsWorld(t, nil)
			enqueue(ev, u.world, fixtureReply{Status: 504, Header: map[string]string{"x-request-id": "up-gateway"}, Body: tc.body}.script(t, decisionsProtocol, ""))
			res, err := runtimeDecisionsCall(ctxFor(t), u, "req-gateway-"+tc.name, decisionsInput(), ai.Hooks{})
			checkUnaryFailure(ev, res, err, ai.CodeUpstreamError, ai.PhaseRequest, ai.StopReasonError)
			var ae *ai.Error
			if errors.As(err, &ae) {
				ev.Check("gateway status and identity with fixed diagnostic", ae.HTTPStatus == 504 && ae.ProviderRequestID == "up-gateway" && ae.Message == "504 Gateway Timeout: Decisions request failed" && !strings.Contains(ae.Message, "synthetic private gateway"), "unexpected diagnostic")
			}
			ev.Check("default gateway failure sends once", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "replayed")
			recordRuntimeConfig(ev, u)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

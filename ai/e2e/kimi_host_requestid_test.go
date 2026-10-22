package e2e

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiHostUnverifiedRequestID(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for route := 0; route < 2; route++ {
			t.Run(fmt.Sprintf("%s/%d", service, route), func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("KIMI-%s-K14-unverified-request-id-%d", service, route))
				w := newKimiHostWorld(t)
				b := w.bindings[service][route]
				reply, raw := kimiHostScript(t, b.API)
				ev.Fixture("synthetic-script.json", raw)
				reply.Header["x-request-id"] = "unverified-openai-header"
				reply.Header["request-id"] = "unverified-anthropic-header"
				w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
				result, err := w.client.Complete(ctxFor(t), textScope("unverified-header"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
				ev.Record("result", result)
				if ev.Check("controlled call succeeds", err == nil && len(result.Metadata.Attempts) == 1, "err=%v result=%+v", err, result) {
					ev.Check("unverified response headers never become attributed provider IDs", result.Metadata.Attempts[0].ProviderRequestID == "", "attempt=%+v", result.Metadata.Attempts[0])
				}
			})
		}
	}
}

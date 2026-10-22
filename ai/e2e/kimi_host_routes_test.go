package e2e

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
)

func TestKimiHostRoutes(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for route := 0; route < 3; route++ {
			for _, entry := range []kimihost.Entry{kimihost.FullComplete, kimihost.FullStream, kimihost.SimpleComplete, kimihost.SimpleStream} {
				t.Run(fmt.Sprintf("%s/%d/%s", service, route, entry), func(t *testing.T) {
					ev := run.Case(t, fmt.Sprintf("KIMI-%s-host-K01-K02-K14-%d-%s", service, route, entry))
					w := newKimiHostWorld(t)
					b := w.bindings[service][route]
					reply, raw := kimiHostScript(t, b.API)
					ev.Fixture("synthetic-kimi-host-script.json", raw)
					w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
					out, err := kimihost.Turn(ctxFor(t), w.client, textScope("kimi-host-turn"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, b.API, entry, ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
						h.Set("X-Host-Trace", "kimi-host-safe")
						return nil
					}})
					ev.Record("result", out.Result)
					ev.Record("events", out.Events)
					ev.Record("envelopes", out.Envelopes)
					ev.Record("requests", w.provider.Requests())
					ev.Record("secret-reads", w.reads)
					meta := out.Result.Metadata
					ev.Check("Host wrapper chooses exact market protocol account and snapshot", err == nil && meta.ProviderID == service && meta.API == b.API && meta.ModelID == "kimi-k3" && meta.AccountScopeID == "same-account-label" && meta.CredentialVersion == "key-v1" && meta.BindingVersion == "binding-v1" && len(meta.Attempts) == 1, "err=%v metadata=%+v", err, meta)
					ev.Check("one selected secret read", w.reads[service] == 1, "reads=%v", w.reads)
					requests := w.provider.Requests()
					suffix := "/chat/completions"
					if b.API == ai.APIOpenAIResponses {
						suffix = "/responses"
					}
					if b.API == ai.APIAnthropicMessages {
						suffix = "/v1/messages"
					}
					ev.Check("single explicit target with trusted trace hook", len(requests) == 1 && requests[0].Path == strings.TrimPrefix(b.Endpoint, w.provider.URL())+suffix && requests[0].Query == "" && requests[0].Header["X-Host-Trace"] == "kimi-host-safe", "requests=%+v", requests)
					if len(meta.Attempts) == 1 {
						ev.Check("Host receives cost certainty independently of token completeness", meta.Attempts[0].UsageReporting == ai.UsageComplete && meta.Attempts[0].CostEstimate.Status == ai.CostResolved, "attempt=%+v", meta.Attempts[0])
					}
					if entry == kimihost.FullStream || entry == kimihost.SimpleStream {
						ev.Check("stream retains attributed events", len(out.Events) > 0 && len(out.Events) == len(out.Envelopes), "events=%d envelopes=%d", len(out.Events), len(out.Envelopes))
					}
					ev.Record("resources", map[string]int64{"calls": w.gauges.ActiveCalls(), "permits": w.gauges.Permits(), "queued": w.gauges.QueuedEvents(), "bodies": w.bodies.Open()})
					ev.Check("wrapper releases resources", w.gauges.ActiveCalls() == 0 && w.gauges.Permits() == 0 && w.gauges.QueuedEvents() == 0 && w.bodies.Open() == 0, "resource leak")
				})
			}
		}
	}
}

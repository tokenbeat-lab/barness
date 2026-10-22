package e2e

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiUsagePricingBoundaries(t *testing.T) {
	tests := []struct {
		name, raw        string
		status           ai.CostEstimateStatus
		code             ai.Code
		highWrite        bool
		minimum, maximum float64
	}{
		{name: "empty", raw: `{}`, status: ai.CostUnreported},
		{name: "all-null", raw: `{"prompt_tokens":null,"completion_tokens":null,"prompt_tokens_details":{"cached_tokens":null,"cache_write_tokens":null}}`, status: ai.CostUnreported},
		{name: "total-below-known-output", raw: `{"completion_tokens":50,"total_tokens":20}`, code: ai.CodeProtocol},
		{name: "details-array", raw: `{"prompt_tokens":100,"prompt_tokens_details":[]}`, code: ai.CodeProtocol},
		{name: "details-string", raw: `{"completion_tokens_details":"invalid"}`, code: ai.CodeProtocol},
		// Deliberately synthetic trusted Host rates: 5m write=10,1h=6.
		// For input=100/output=50/read=20/write=80, the true lower is1h.
		{name: "reversed-write-rates", raw: `{"prompt_tokens":200,"completion_tokens":50,"prompt_tokens_details":{"cached_tokens":20,"cache_write_tokens":80},"completion_tokens_details":{"reasoning_tokens":10}}`, status: ai.CostBounded, highWrite: true, minimum: .001536, maximum: .001856},
	}
	for _, sc := range tests {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "KIMI-INTL-CHAT-K10-K11-foundation-"+sc.name)
			model := kimiFoundationModel()
			if sc.highWrite {
				model.Cost.CacheWrite = 10
			}
			cat := ai.Catalog{Version: "kimi-cost-boundaries-synthetic", Models: []ai.Model{model}}
			w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
			b := chatBinding(tenantA, w.provider.URL())
			b.ProviderID = ai.ProviderMoonshot
			b.AllowedModels = []string{"kimi-k3"}
			w.host.PutBinding(b)
			events := []json.RawMessage{json.RawMessage(`{"choices":[{"delta":{"content":"Ready."},"finish_reason":"stop"}]}`), json.RawMessage(`{"choices":[],"usage":` + sc.raw + `}`)}
			reply := chatEvents(t, events, "")
			enqueue(ev, w, reply)
			result, err := w.client.Complete(ctxFor(t), textScope(sc.name), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, ai.ChatOptions{})
			ev.Record("input-usage", json.RawMessage(sc.raw))
			ev.Record("result", result)
			if sc.code != "" {
				e, ok := err.(*ai.Error)
				ev.Check("contradictory or invalid counts reject", ok && e.Code == sc.code, "err=%v", err)
				return
			}
			if !ev.Check("supported call retains usage report", err == nil && len(result.Metadata.Attempts) == 1, "err=%v result=%+v", err, result) {
				return
			}
			estimate := result.Metadata.Attempts[0].CostEstimate
			ev.Check("cost reliability follows available billable facts", estimate.Status == sc.status, "got=%+v", estimate)
			low, lok := estimate.Minimum.Get()
			high, hok := estimate.Maximum.Get()
			if sc.highWrite {
				ev.Check("Host prices produce ordered real bounds and point", lok && hok && math.Abs(low.Total-sc.minimum) < 1e-12 && math.Abs(high.Total-sc.maximum) < 1e-12 && result.Message.Usage.Cost == low, "estimate=%+v usage=%+v", estimate, result.Message.Usage)
			} else {
				ev.Check("unreported has no fabricated amount bounds", !lok && !hok, "estimate=%+v", estimate)
			}
		})
	}
}

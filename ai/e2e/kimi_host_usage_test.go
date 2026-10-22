package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestKimiHostReasoningTotalContainment(t *testing.T) {
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		t.Run(string(service), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("KIMI-%s-MESSAGES-K10-reasoning-exceeds-reported-total", service))
			w := newKimiHostWorld(t)
			b := w.bindings[service][2]
			// Missing output is unknown, but reasoning50 cannot fit reported total20.
			frames := []json.RawMessage{
				json.RawMessage(`{"type":"message_start","message":{"id":"kimi-total","model":"kimi-k3","type":"message","role":"assistant","content":[],"usage":{"total_tokens":20,"output_tokens_details":{"thinking_tokens":50}}}}`),
				json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":"Ready."}}`),
				json.RawMessage(`{"type":"content_block_stop","index":0}`),
				json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"}}`),
				json.RawMessage(`{"type":"message_stop"}`),
			}
			ev.Fixture("synthetic-kimi-total-contradiction.json", mustMarshal(t, frames))
			w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), sseEvents(t, frames, provider.FramingLF))
			result, err := w.client.Complete(ctxFor(t), textScope("reasoning-total"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
			ev.Record("result", result)
			e, ok := err.(*ai.Error)
			ev.Check("reasoning subset must fit total even without output counter", ok && e.Code == ai.CodeProtocol, "err=%v usage=%+v", err, result.Message.Usage)
		})
	}
}

func TestKimiHostCostOverflowPreservesLegalUsage(t *testing.T) {
	ev := run.Case(t, "KIMI-INTL-CHAT-K10-K11-cost-overflow-last-legal-snapshot")
	model := kimiFoundationModel()
	model.Cost.Input = 1e300
	model.Cost.CacheWrite = 1e300
	cat := ai.Catalog{Version: "kimi-synthetic-finite-extreme-prices", Models: []ai.Model{model}}
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &cat }, tenantA)
	b := chatBinding(tenantA, w.provider.URL())
	b.ProviderID = ai.ProviderMoonshot
	b.AllowedModels = []string{"kimi-k3"}
	w.host.PutBinding(b)
	frames := []json.RawMessage{
		json.RawMessage(`{"choices":[{"index":0,"delta":{"content":"Ready."},"finish_reason":null}],"usage":{"prompt_tokens":1,"completion_tokens":1,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}}`),
		json.RawMessage(`{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":9223372036854775806,"completion_tokens":1,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}}`),
	}
	ev.Fixture("synthetic-overflow-script.json", mustMarshal(t, frames))
	enqueue(ev, w, chatEvents(t, frames, ""))
	result, err := w.client.Complete(ctxFor(t), textScope("overflow"), ai.Target{BindingID: b.BindingID, ModelID: "kimi-k3"}, ai.Request{Messages: []ai.Message{ai.UserText("Ready?")}}, nil)
	// Record only finite facts before the expected fix; the failed raw stdout
	// retains the contradictory result without forging a valid JSON amount.
	ev.Record("error-code", fmt.Sprintf("%T", err))
	ev.Record("attempt-count", len(result.Metadata.Attempts))
	e, ok := err.(*ai.Error)
	ev.Check("non-finite cost fails with protocol classification", ok && e.Code == ai.CodeProtocol && e.Phase == ai.PhaseStream, "err=%v", err)
	if len(result.Metadata.Attempts) == 1 {
		a := result.Metadata.Attempts[0]
		ev.Check("last legal report survives arithmetic overflow", a.Usage.Input == 1 && a.Usage.Output == 1 && a.Usage.TotalTokens == 2 && a.CostEstimate.Status == ai.CostResolved && result.Message.Usage == a.Usage, "usage=%+v estimate=%+v", a.Usage, a.CostEstimate)
	}
}

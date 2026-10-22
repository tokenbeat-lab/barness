package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiMessagesNamedGuards(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, c := range []struct {
			name string
			opts ai.AnthropicOptions
		}{
			{"medium", ai.AnthropicOptions{Effort: ai.AnthropicEffortMedium}},
			{"xhigh", ai.AnthropicOptions{Effort: ai.AnthropicEffortXHigh}},
			{"enabled", ai.AnthropicOptions{ThinkingEnabled: ai.Value(true)}},
			{"disabled", ai.AnthropicOptions{ThinkingEnabled: ai.Value(false)}},
			{"enabled-null", ai.AnthropicOptions{ThinkingEnabled: ai.Null[bool]()}},
			{"budget", ai.AnthropicOptions{ThinkingBudgetTokens: 1024}},
			{"display", ai.AnthropicOptions{ThinkingDisplay: "summarized"}},
			{"interleaved", ai.AnthropicOptions{InterleavedThinking: ai.Value(false)}},
			{"temperature", ai.AnthropicOptions{Temperature: ai.Value(1.)}},
			{"cache-none", ai.AnthropicOptions{CacheRetention: ai.CacheRetentionNone}},
			{"zero-output", ai.AnthropicOptions{MaxTokens: ai.Value(0)}},
			{"null-output", ai.AnthropicOptions{MaxTokens: ai.Null[int]()}},
			{"large-output", ai.AnthropicOptions{MaxTokens: ai.Value(131073)}},
			{"named-tool", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Tool: "inspect"})}},
		} {
			t.Run(string(market)+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K04-K05-K09-"+c.name)
				w := newKimiMessagesWorld(t, market, nil)
				scope := textScope("kimi-named-" + c.name)
				o := drain(w.client.Stream(ctxFor(t), scope, kimiMessagesTarget, kimiMessagesRequest(), c.opts))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
				ev.Check("no credential read or HTTP attempt", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", w.host.ReadsFor(scope.RequestID))
			})
		}
	}
}

func TestKimiMessagesFinalGuards(t *testing.T) {
	for _, market := range []ai.ProviderID{"moonshot", "moonshot-cn"} {
		for _, c := range []struct {
			name string
			edit func(*ai.Payload)
		}{
			{"thinking", func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "adaptive"} }},
			{"temperature", func(p *ai.Payload) { p.Body["temperature"] = 1 }},
			{"summary", func(p *ai.Payload) {
				p.Body["output_config"] = map[string]any{"format": map[string]any{"type": "json_schema"}}
			}},
			{"effort", func(p *ai.Payload) { p.Body["output_config"] = map[string]any{"effort": "medium"} }},
			{"named-tool", func(p *ai.Payload) { p.Body["tool_choice"] = map[string]any{"type": "tool", "name": "inspect"} }},
			{"fraction-output", func(p *ai.Payload) { p.Body["max_tokens"] = json.Number("1.00000000000000001") }},
			{"cache-null", func(p *ai.Payload) { p.Body["cache_control"] = nil }},
			{"cache-24h", func(p *ai.Payload) { p.Body["cache_control"] = map[string]any{"type": "ephemeral", "ttl": "24h"} }},
			{"cache-block", func(p *ai.Payload) {
				p.Body["messages"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["cache_control"] = map[string]any{"type": "ephemeral"}
			}},
			{"image", func(p *ai.Payload) {
				p.Body["messages"].([]any)[0].(map[string]any)["content"] = []any{map[string]any{"type": "image", "source": map[string]any{"type": "base64", "data": "AA=="}}}
			}},
			{"beta", func(p *ai.Payload) { p.Body["betas"] = []any{"thinking-beta"} }},
			{"unknown-private-field", func(p *ai.Payload) { p.Body["sensitive-unknown-value"] = "secret-test-marker" }},
		} {
			t.Run(string(market)+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "KIMI-"+string(market)+"-MESSAGES-K03-K14-"+c.name)
				w := newKimiMessagesWorld(t, market, nil)
				hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
					c.edit(p)
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("kimi-final"), kimiMessagesTarget, kimiMessagesRequest(), nil))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("final invalid payload never reaches HTTP", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

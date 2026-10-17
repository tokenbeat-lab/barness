package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestMiniMaxMessagesNamedGuards(t *testing.T) {
	for _, c := range []struct {
		name, model string
		opts        ai.AnthropicOptions
	}{
		{"m3-depth", "MiniMax-M3", ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}},
		{"m31-off", "MiniMax-M3.1-Flash-Preview", ai.AnthropicOptions{ThinkingEnabled: ai.Value(false)}},
		{"budget", "MiniMax-M3", ai.AnthropicOptions{ThinkingBudgetTokens: 1024}},
		{"display", "MiniMax-M3", ai.AnthropicOptions{ThinkingDisplay: "omitted"}},
		{"interleaved-false", "MiniMax-M3", ai.AnthropicOptions{InterleavedThinking: ai.Value(false)}},
		{"budget-zero", "MiniMax-M3", ai.AnthropicOptions{MaxTokens: ai.Value(0)}},
		{"budget-high", "MiniMax-M3", ai.AnthropicOptions{MaxTokens: ai.Value(131073)}},
		{"temperature-high", "MiniMax-M3", ai.AnthropicOptions{Temperature: ai.Value(2.1)}},
		{"forced-any", "MiniMax-M3", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "any"})}},
		{"forced-name", "MiniMax-M3", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Tool: "lookup"})}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-named-"+c.name)
			w := newMiniMaxMessagesWorld(t, nil)
			scope := textScope("mm04-" + c.name)
			callbacks := 0
			hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
				callbacks++
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, miniMaxMessagesTarget(c.model), miniMaxMessagesRequest(), c.opts))
			o.record(ev)
			ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
			ev.Check("invalid named control rejected before credential and callback", callbacks == 0 && w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

func TestMiniMaxMessagesFinalGuards(t *testing.T) {
	for _, c := range []struct {
		name, model string
		edit        func(*ai.Payload)
	}{
		{"m3-effort", "MiniMax-M3", func(p *ai.Payload) { p.Body["output_config"] = map[string]any{"effort": "high"} }},
		{"m31-off", "MiniMax-M3.1-Flash-Preview", func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "disabled"} }},
		{"m31-none", "MiniMax-M3.1-Flash-Preview", func(p *ai.Payload) { p.Body["output_config"] = map[string]any{"effort": "none"} }},
		{"m31-minimal", "MiniMax-M3.1-Flash-Preview", func(p *ai.Payload) { p.Body["output_config"] = map[string]any{"effort": "minimal"} }},
		{"budget-fraction", "MiniMax-M3", func(p *ai.Payload) { p.Body["max_tokens"] = json.Number("1.000000000000000000000000001") }},
		{"budget-null", "MiniMax-M3", func(p *ai.Payload) { p.Body["max_tokens"] = nil }},
		{"budget-zero", "MiniMax-M3", func(p *ai.Payload) { p.Body["max_tokens"] = 0 }},
		{"budget-high", "MiniMax-M3", func(p *ai.Payload) { p.Body["max_tokens"] = 131073 }},
		{"temperature-high", "MiniMax-M3", func(p *ai.Payload) { p.Body["temperature"] = 2.1 }},
		{"top-p-high", "MiniMax-M3", func(p *ai.Payload) { p.Body["top_p"] = 1.1 }},
		{"top-p-null", "MiniMax-M3", func(p *ai.Payload) { p.Body["top_p"] = nil }},
		{"thinking-budget", "MiniMax-M3", func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "adaptive", "budget_tokens": 1024} }},
		{"thinking-display", "MiniMax-M3", func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "adaptive", "display": "omitted"} }},
		{"thinking-binding", "MiniMax-M3", func(p *ai.Payload) {
			p.Body["thinking"] = map[string]any{"type": "adaptive", "block_binding": map[string]any{}}
		}},
		{"stop", "MiniMax-M3", func(p *ai.Payload) { p.Body["stop_sequences"] = []any{"stop"} }},
		{"top-k", "MiniMax-M3", func(p *ai.Payload) { p.Body["top_k"] = 1 }},
		{"priority", "MiniMax-M3", func(p *ai.Payload) { p.Body["service_tier"] = "priority" }},
		{"cache", "MiniMax-M3", func(p *ai.Payload) {
			p.Body["messages"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["cache_control"] = map[string]any{"type": "ephemeral"}
		}},
		{"parallel", "MiniMax-M3", func(p *ai.Payload) {
			p.Body["tool_choice"] = map[string]any{"type": "auto", "disable_parallel_tool_use": true}
		}},
		{"forced", "MiniMax-M3", func(p *ai.Payload) { p.Body["tool_choice"] = map[string]any{"type": "any"} }},
		{"redacted", "MiniMax-M3", func(p *ai.Payload) {
			p.Body["messages"].([]any)[0].(map[string]any)["content"] = []any{map[string]any{"type": "redacted_thinking", "data": "opaque"}}
		}},
		{"image", "MiniMax-M3", func(p *ai.Payload) {
			p.Body["messages"].([]any)[0].(map[string]any)["content"] = []any{map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "AA=="}}}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "MM04-final-"+c.name)
			w := newMiniMaxMessagesWorld(t, nil)
			hooks := ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
				c.edit(p)
				return ai.KeepPayload(), nil
			}}
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("mm04-final-"+c.name), miniMaxMessagesTarget(c.model), miniMaxMessagesRequest(), nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("invalid final shape never reaches Provider", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// These are unverified Messages controls, regardless of another protocol's
// acceptance. A named control cannot be repaired by a later payload callback.
func TestGLMMessagesUnverifiedOptions(t *testing.T) {
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, c := range []struct {
			name string
			opts ai.AnthropicOptions
		}{
			{"thinking-enabled", ai.AnthropicOptions{ThinkingEnabled: ai.Value(true)}},
			{"thinking-disabled", ai.AnthropicOptions{ThinkingEnabled: ai.Value(false)}},
			{"thinking-budget", ai.AnthropicOptions{ThinkingBudgetTokens: 1024}},
			{"effort", ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}},
			{"display", ai.AnthropicOptions{ThinkingDisplay: "summarized"}},
			{"interleaved", ai.AnthropicOptions{InterleavedThinking: ai.Value(false)}},
			{"cache-short", ai.AnthropicOptions{CacheRetention: ai.CacheRetentionShort}},
			{"cache-long", ai.AnthropicOptions{CacheRetention: ai.CacheRetentionLong}},
			{"temperature", ai.AnthropicOptions{Temperature: ai.Value(0.2)}},
			{"metadata", ai.AnthropicOptions{Metadata: map[string]json.RawMessage{"user_id": json.RawMessage(`"synthetic-user"`)}}},
			{"forced-tool", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Tool: "lookup_probe_code"})}},
		} {
			t.Run(model+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "GLM06-options-"+model+"-"+c.name)
				w := newGLMMessagesWorld(t, nil)
				payloads := 0
				hooks := ai.Hooks{OnPayload: func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
					payloads++
					return ai.KeepPayload(), nil
				}}
				o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm-message-option-"+c.name), ai.Target{BindingID: "glm-messages", ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input.")}}, c.opts))
				o.record(ev)
				ev.Record("requests", w.provider.Requests())
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("unverified option rejected before callback or send", payloads == 0 && len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMMessagesFinalGuards(t *testing.T) {
	for _, c := range []struct {
		name  string
		hooks ai.Hooks
	}{
		{"thinking", glmMessagesPayload(func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "disabled"} })},
		{"output-config", glmMessagesPayload(func(p *ai.Payload) { p.Body["output_config"] = map[string]any{"effort": "high"} })},
		{"cache", glmMessagesPayload(func(p *ai.Payload) { p.Body["cache_control"] = map[string]any{"type": "ephemeral"} })},
		{"nested-cache", glmMessagesPayload(func(p *ai.Payload) {
			p.Body["messages"].([]any)[0].(map[string]any)["content"].([]any)[0].(map[string]any)["cache_control"] = map[string]any{"type": "ephemeral"}
		})},
		{"beta-body", glmMessagesPayload(func(p *ai.Payload) { p.Body["betas"] = []any{"interleaved-thinking-2025-05-14"} })},
		{"beta-header", ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Anthropic-Beta", "interleaved-thinking-2025-05-14")
			return nil
		}}},
		{"browser", ai.Hooks{TransformHeaders: func(_ context.Context, _ ai.CallScope, h http.Header) error {
			h.Set("Anthropic-Dangerous-Direct-Browser-Access", "true")
			return nil
		}}},
		{"temperature", glmMessagesPayload(func(p *ai.Payload) { p.Body["temperature"] = 0.2 })},
		{"budget-zero", glmMessagesPayload(func(p *ai.Payload) { p.Body["max_tokens"] = 0 })},
		{"budget-fraction", glmMessagesPayload(func(p *ai.Payload) { p.Body["max_tokens"] = 1.5 })},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM06-hook-guard-"+c.name)
			w := newGLMMessagesWorld(t, nil)
			o := drain(w.client.WithHooks(c.hooks).Stream(ctxFor(t), textScope("glm-message-hook-"+c.name), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.3"}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input.")}}, nil))
			o.record(ev)
			ev.Record("requests", w.provider.Requests())
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("unsupported final wire field never sent", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

func glmMessagesPayload(edit func(*ai.Payload)) ai.Hooks {
	return ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, p *ai.Payload) (ai.PayloadDecision, error) {
		edit(p)
		return ai.KeepPayload(), nil
	}}
}

func TestGLMMessagesSamplingIgnored(t *testing.T) {
	scenarios, prompt, raw := loadGLMMessages(t)
	for _, sc := range scenarios {
		t.Run(sc.Model, func(t *testing.T) {
			ev := run.Case(t, "GLM06-sampling-ignored-"+sc.Model)
			params := map[string]json.RawMessage{"thinking": json.RawMessage(`{"type":"disabled"}`), "max_tokens": json.RawMessage(`1`)}
			w := newGLMMessagesWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					cat.Models[i].SamplingParams = params
					cat.Models[i].SamplingParamsByThinkingLevel = map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingLow: params}
				}
			})
			ev.Fixture("text.json", raw)
			enqueue(ev, w, glmMessagesReply(t, sc))
			res, err := w.client.CompleteSimple(ctxFor(t), textScope("glm-messages-sampling-"+sc.Model), ai.Target{BindingID: "glm-messages", ModelID: sc.Model}, ai.Request{Messages: []ai.Message{ai.UserText(prompt)}}, ai.SimpleOptions{MaxTokens: ai.Value(4096), SamplingParams: params})
			ev.Record("result", res)
			ev.Record("error", errString(err))
			ev.Record("requests", w.provider.Requests())
			ev.Check("Messages ignores all sampling maps", err == nil && len(w.provider.Requests()) == 1, "%v", err)
			if len(w.provider.Requests()) == 1 {
				fields := bodyFields(t, w.provider.Requests()[0].Body)
				_, thinking := fields["thinking"]
				ev.Check("maps did not alter omission or output cap", !thinking && string(fields["max_tokens"]) == "4096", "%s", w.provider.Requests()[0].Body)
			}
		})
	}
}

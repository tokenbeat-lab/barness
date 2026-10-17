package e2e

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestGLMMessagesThinkingBoundary(t *testing.T) {
	for _, model := range []string{"glm-5.2", "glm-5.3"} {
		for _, c := range []struct {
			name string
			opts ai.AnthropicOptions
		}{
			{"budget-512", ai.AnthropicOptions{MaxTokens: ai.Value(4096), ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 512}},
			{"budget-2048", ai.AnthropicOptions{MaxTokens: ai.Value(4096), ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 2048}},
			{"output-2048", ai.AnthropicOptions{MaxTokens: ai.Value(2048), ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 1024}},
			{"output-default", ai.AnthropicOptions{ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 1024}},
			{"budget-without-enabled", ai.AnthropicOptions{MaxTokens: ai.Value(4096), ThinkingBudgetTokens: 1024}},
			{"any", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "any"})}},
			{"none", ai.AnthropicOptions{ToolChoice: ai.Value(ai.AnthropicToolChoice{Mode: "none"})}},
		} {
			t.Run(model+"/"+c.name, func(t *testing.T) {
				ev := run.Case(t, "GLM07-thinking-boundary-"+model+"-"+c.name)
				w := newGLMMessagesWorld(t, nil)
				o := drain(w.client.Stream(ctxFor(t), textScope("glm07-thinking-boundary"), ai.Target{BindingID: "glm-messages", ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, c.opts))
				o.record(ev)
				checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
				ev.Check("unsupported named shape never attempted", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
		for _, level := range []ai.ThinkingLevel{"off", ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax} {
			t.Run(model+"/simple-"+string(level), func(t *testing.T) {
				ev := run.Case(t, "GLM07-simple-canonical-"+model+"-"+string(level))
				w := newGLMMessagesWorld(t, nil)
				o := drain(w.client.StreamSimple(ctxFor(t), textScope("glm07-simple-canonical"), ai.Target{BindingID: "glm-messages", ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, ai.SimpleOptions{MaxTokens: ai.Value(4096), Reasoning: level, ThinkingBudgets: ai.ThinkingBudgets{Low: ai.Value(1024)}}))
				o.record(ev)
				phase := ai.PhaseRequest
				if level == "off" {
					phase = ai.PhaseCapability
				}
				checkRejected(ev, o, ai.CodeInvalidRequest, phase)
				ev.Check("exact budget evidence never enables guessed canonical intensity", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
			})
		}
	}
}

func TestGLMMessagesFinalThinkingBoundary(t *testing.T) {
	for _, c := range []struct {
		name string
		edit func(*ai.Payload)
	}{
		{"fractional-output", func(p *ai.Payload) { p.Body["max_tokens"] = json.Number("4096.0000000000000001") }},
		{"fractional-budget", func(p *ai.Payload) {
			p.Body["thinking"].(map[string]any)["budget_tokens"] = json.Number("1024.0000000000000001")
		}},
		{"other-budget", func(p *ai.Payload) { p.Body["thinking"].(map[string]any)["budget_tokens"] = 2048 }},
		{"other-output", func(p *ai.Payload) { p.Body["max_tokens"] = 8192 }},
		{"display", func(p *ai.Payload) { p.Body["thinking"].(map[string]any)["display"] = "summarized" }},
		{"adaptive", func(p *ai.Payload) { p.Body["thinking"] = map[string]any{"type": "adaptive"} }},
		{"any-choice", func(p *ai.Payload) { p.Body["tool_choice"] = map[string]any{"type": "any"} }},
		{"none-choice", func(p *ai.Payload) { p.Body["tool_choice"] = map[string]any{"type": "none"} }},
		{"named-choice", func(p *ai.Payload) {
			p.Body["tool_choice"] = map[string]any{"type": "tool", "name": "lookup_probe_code"}
		}},
		{"empty-signature", func(p *ai.Payload) {
			p.Body["messages"] = []any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "thinking", "thinking": "synthetic", "signature": ""}}}}
		}},
		{"redacted", func(p *ai.Payload) {
			p.Body["messages"] = []any{map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "redacted_thinking", "data": "synthetic-unproven-redacted"}}}}
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			ev := run.Case(t, "GLM07-final-thinking-boundary-"+c.name)
			w := newGLMMessagesWorld(t, nil)
			o := drain(w.client.WithHooks(glmMessagesPayload(c.edit)).Stream(ctxFor(t), textScope("glm07-final-boundary"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.2"}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, glmMessagesEnabled()))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("unsupported final shape never attempted", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

func TestGLMMessagesBudgetMetadata(t *testing.T) {
	for _, budget := range []int{0, 2048} {
		t.Run(strconv.Itoa(budget), func(t *testing.T) {
			ev := run.Case(t, "GLM07-model-budget-"+strconv.Itoa(budget))
			w := newGLMMessagesWorld(t, func(cat *ai.Catalog) {
				for i := range cat.Models {
					cat.Models[i].Compat.MessagesThinkingBudgetTokens = budget
				}
			})
			opts := glmMessagesEnabled()
			if budget != 0 {
				opts.ThinkingBudgetTokens = budget
			}
			o := drain(w.client.Stream(ctxFor(t), textScope("glm07-model-budget"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.2"}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, opts))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("metadata narrows provider support without enabling arbitrary budgets", len(w.provider.Requests()) == 0, "%+v", w.provider.Requests())
		})
	}
}

func TestGLMMessagesCanonicalCannotGuessBudget(t *testing.T) {
	ev := run.Case(t, "GLM07-canonical-cannot-guess-budget")
	w := newGLMMessagesWorld(t, func(cat *ai.Catalog) {
		for i := range cat.Models {
			cat.Models[i].ThinkingLevelMap.Low = ai.Value("low")
		}
	})
	o := drain(w.client.StreamSimple(ctxFor(t), textScope("glm07-canonical-guess"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.2"}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input")}}, ai.SimpleOptions{MaxTokens: ai.Value(3072), Reasoning: ai.ThinkingLow, ThinkingBudgets: ai.ThinkingBudgets{Low: ai.Value(1024)}}))
	o.record(ev)
	checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
	ev.Check("canonical budget arithmetic cannot opt into one observed full wire shape", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
}

func TestGLMMessagesUnverifiedToolFields(t *testing.T) {
	for _, field := range []string{"strict", "defer_loading", "type"} {
		t.Run(field, func(t *testing.T) {
			ev := run.Case(t, "GLM07-tool-field-"+field)
			w := newGLMMessagesWorld(t, nil)
			req := glmMessagesCapturedRequest(t, glmMessagesCaptureByName(t, "glm-5.3-tool"))
			hooks := glmMessagesPayload(func(p *ai.Payload) {
				t := p.Body["tools"].([]any)[0].(map[string]any)
				t[field] = true
				if field == "type" {
					t[field] = "custom"
				}
			})
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), textScope("glm07-tool-field"), ai.Target{BindingID: "glm-messages", ModelID: "glm-5.3"}, req, nil))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseRequest)
			ev.Check("unverified Claude tool extension never attempted", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "%+v", o.result.Metadata)
		})
	}
}

package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func TestRegionalZAIThinkingBoundaries(t *testing.T) {
	for _, mode := range []string{"budget512", "budget2048", "output2048", "output-default", "disabled", "missing-enabled", "metadata-zero", "metadata2048", "model53", "hook-budget", "hook-output", "hook-fraction", "hook-extra", "hook-adaptive"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "RP13-zai-thinking-boundary-"+mode)
			cat, err := regional.CandidateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			opts, model, hooks := glmMessagesEnabled(), "glm-5.2", ai.Hooks{}
			reads, phase := 0, ai.PhaseCapability
			switch mode {
			case "budget512":
				opts.ThinkingBudgetTokens = 512
			case "budget2048":
				opts.ThinkingBudgetTokens = 2048
			case "output2048":
				opts.MaxTokens = ai.Value(2048)
			case "output-default":
				opts.MaxTokens = ai.Nullable[int]{}
			case "disabled":
				opts.ThinkingEnabled = ai.Value(false)
			case "missing-enabled":
				opts.ThinkingEnabled = ai.Nullable[bool]{}
			case "model53":
				model = "glm-5.3"
			case "metadata-zero", "metadata2048":
				for i := range cat.Models {
					if cat.Models[i].Provider == ai.ProviderZAI && cat.Models[i].API == ai.APIAnthropicMessages && cat.Models[i].ID == model {
						cat.Models[i].Compat.MessagesThinkingBudgetTokens = 0
						if mode == "metadata2048" {
							cat.Models[i].Compat.MessagesThinkingBudgetTokens, opts.ThinkingBudgetTokens = 2048, 2048
						}
					}
				}
			default:
				reads, phase = 1, ai.PhaseRequest
				hooks = glmMessagesPayload(func(p *ai.Payload) {
					switch mode {
					case "hook-budget":
						p.Body["thinking"].(map[string]any)["budget_tokens"] = 2048
					case "hook-output":
						p.Body["max_tokens"] = 8192
					case "hook-fraction":
						p.Body["thinking"].(map[string]any)["budget_tokens"] = json.Number("1024.0000000000000001")
					case "hook-extra":
						p.Body["thinking"].(map[string]any)["clear_thinking"] = false
					case "hook-adaptive":
						p.Body["thinking"] = map[string]any{"type": "adaptive"}
					}
				})
			}
			w, target := regionalWorld(t, ai.ProviderZAI, ai.APIAnthropicMessages, func(c *ai.Config) { c.Catalog = &cat })
			target.ModelID = model
			scope := textScope("zai-thinking-boundary")
			o := drain(w.client.WithHooks(hooks).Stream(ctxFor(t), scope, target, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic input.")}}, opts))
			o.record(ev)
			ev.Record("host-reads", w.host.ReadsFor(scope.RequestID))
			checkRejected(ev, o, ai.CodeInvalidRequest, phase)
			ev.Check("unsupported shape cannot cross appropriate secret/attempt boundary", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0 && w.host.ReadsFor(scope.RequestID).Credentials == reads, "reads=%+v", w.host.ReadsFor(scope.RequestID))
		})
	}
}

func TestRegionalZAICanonicalCannotGuessBudget(t *testing.T) {
	for _, level := range []ai.ThinkingLevel{"off", ai.ThinkingMinimal, ai.ThinkingLow, ai.ThinkingMedium, ai.ThinkingHigh, ai.ThinkingXHigh, ai.ThinkingMax} {
		t.Run(string(level), func(t *testing.T) {
			ev := run.Case(t, "RP13-zai-canonical-no-budget-"+string(level))
			cat, err := regional.CandidateCatalog()
			if err != nil {
				t.Fatal(err)
			}
			w, target := regionalWorld(t, ai.ProviderZAI, ai.APIAnthropicMessages, func(c *ai.Config) { c.Catalog = &cat })
			target.ModelID = "glm-5.2"
			scope := textScope("zai-canonical-budget")
			o := drain(w.client.StreamSimple(ctxFor(t), scope, target, ai.Request{}, ai.SimpleOptions{MaxTokens: ai.Value(4096), Reasoning: level, ThinkingBudgets: ai.ThinkingBudgets{Low: ai.Value(1024)}}))
			o.record(ev)
			checkRejected(ev, o, ai.CodeInvalidRequest, ai.PhaseCapability)
			ev.Check("exact full shape grants no canonical effort before secret", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0 && w.host.ReadsFor(scope.RequestID).Credentials == 0, "reads=%+v", w.host.ReadsFor(scope.RequestID))
		})
	}
}

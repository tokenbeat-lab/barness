//go:build live

package live

import (
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
)

// effortChanges settles what issue 27 leaves to the live smoke: Anthropic
// accepts a managed-effort model's conversation whose effort changes
// between turns (ADR-0019). The first turn runs at high and its thinking is
// replayed into a second turn at low, preceded by the system message naming
// high, with low trailing. A refusal takes the model out of the catalog
// again; the conclusion goes to ADR-0019 and the support matrix.
func effortChanges(model string) scenario {
	return scenario{id: "effort-changes-" + model, model: model, run: func(s *session) {
		if !s.spend() {
			return
		}
		req := ai.Request{SystemPrompt: "You are a concise assistant.", Messages: []ai.Message{ai.UserText(shortPrompt)}}
		opts := ai.AnthropicOptions{MaxTokens: ai.Value(2048), Effort: ai.AnthropicEffortHigh}
		res, err := s.env.client.Complete(s.ctx, newScope(), s.target(), req, opts)
		ex := s.observe("high", res, err)
		if len(ex) > 0 {
			// The recorder keys request headers in lower case.
			betas := ex[len(ex)-1].RequestHeaders["anthropic-beta"]
			s.check("high: managed-effort betas sent", strings.Contains(betas, "mid-conversation-output-config-2026-07-01") &&
				strings.Contains(betas, "thinking-binding-controls-2026-08-01"), "Anthropic-Beta %q", betas)
		}
		if !s.ok("high", err) || !s.turn("high", res, ai.StopReasonStop) {
			return
		}
		s.check("high: the turn's effort is recorded", res.Message.ProviderThinkingLevel == "high", "got %q", res.Message.ProviderThinkingLevel)
		if !s.spend() {
			return
		}
		req.Messages = append(req.Messages, res.Message, ai.UserText("Now say it in three words."))
		opts.Effort = ai.AnthropicEffortLow
		res, err = s.env.client.Complete(s.ctx, newScope(), s.target(), req, opts)
		ex = s.observe("low", res, err)
		messages, _ := requestBody(ex)["messages"].([]any)
		var efforts []string
		for _, m := range messages {
			msg, _ := m.(map[string]any)
			if cfg, ok := msg["output_config"].(map[string]any); ok && msg["role"] == "system" {
				effort, _ := cfg["effort"].(string)
				efforts = append(efforts, effort)
			}
		}
		s.check("low: the first turn's effort precedes it and the new one trails",
			strings.Join(efforts, ",") == "high,low", "efforts %q", efforts)
		if !s.ok("low", err) || !s.turn("low", res, ai.StopReasonStop) {
			return
		}
		s.check("low: the turn's effort is recorded", res.Message.ProviderThinkingLevel == "low", "got %q", res.Message.ProviderThinkingLevel)
		s.note("the vendor accepted the effort change from high to low")
	}}
}

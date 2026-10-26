package supportmatrix

import "github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"

// GenericExpected is the existing fixed complete smoke contract plus the
// default-model xAI-origin history scene. Extra-model capabilities stay intact.
func GenericExpected(combo string) []string {
	base := []string{"text-stream", "text-complete", "tool-round-trip", "cancel-after-first-frame", "reasoning-history", "image"}
	var extras []string
	switch combo {
	case "openai-responses":
		extras = []string{"tool-changes-gpt-5.4", "tool-changes-gpt-6-sol"}
	case "openai-chat":
		extras = []string{"usage-position", "tool-changes-gpt-5.4", "tool-changes-gpt-6-sol", "pro-model-unavailable"}
	case "deepseek-responses":
		extras = []string{"reasoning-level-high", "reasoning-level-max"}
	case "deepseek-chat":
		extras = []string{"usage-position", "forced-tool-with-thinking", "prompt-cache-long-retention", "reasoning-level-high", "reasoning-level-max"}
	case "anthropic-messages":
		extras = []string{"tool-changes-claude-opus-4-8", "tool-changes-claude-fable-5", "effort-changes-claude-fable-5-1", "effort-changes-claude-opus-5", "effort-changes-claude-opus-5-5"}
	case "google-gemini":
	default:
		return nil
	}
	return append(append(base, extras...), grokhistory.Scenario)
}

// Existing optional capabilities keep their historical reporting semantics.
// The additive foreign-history probe is mandatory: UNSUPPORTED cannot be
// merged as a successful current capability. FAIL/NOT_RUN remain partial facts.
func checkGenericHistory(r Report) error {
	if len(GenericExpected(r.Combo)) == 0 {
		return nil
	}
	for _, s := range r.Scenarios {
		if s.ID == grokhistory.Scenario && s.Outcome == Unsupported {
			return refuse("mandatory xAI history cannot be unsupported")
		}
	}
	return nil
}

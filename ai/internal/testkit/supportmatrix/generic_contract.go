package supportmatrix

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
)

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

// OpenAIChatNoReplay is the pre-existing native noReplay declaration. It says
// nothing about model reasoning usage: Chat exposes no native state to replay.
const OpenAIChatNoReplay = "OpenAI Chat Completions returns no reasoning content or signature to replay; reasoning is only counted in usage"

// declaredGenericUnsupported is intentionally fixed to the one reachable
// native declaration for the six current routes/models. Dynamic catalog skips
// are not authority to excuse a required capability in a formal report.
func declaredGenericUnsupported(r Report, s ScenarioResult) bool {
	return r.Combo == "openai-chat" && r.Operation == "chat" && r.Provider == "openai" && r.API == "openai-completions" &&
		r.Endpoint == "https://api.openai.com/v1" && r.Model == "gpt-5-mini" &&
		s.ID == "reasoning-history" && s.Model == r.Model && s.Outcome == Unsupported && s.Note == OpenAIChatNoReplay &&
		s.Calls == 0 && s.Attempts == 0 && s.Retries == 0 && len(s.ProviderRequestIDs) == 0 &&
		s.Images == 0 && s.Questions == 0 && s.StateBytes == 0 && s.DecodedImageBytes == 0 && len(s.ImageCalls) == 0 &&
		s.ErrorCategory == "" && !s.Environment && s.DurationMS >= 0
}

// VerifyGenericUnsupportedCase checks the declaration after the caller has
// authenticated case bytes and inventory. A no-request capability has only its
// matching unsupported assertion record, never HTTP/results/Observer captures.
func VerifyGenericUnsupportedCase(r Report, s ScenarioResult, artifacts map[string]string, reason string, assertions int) error {
	if !declaredGenericUnsupported(r, s) || len(artifacts) != 1 || artifacts["assertions.json"] == "" || reason != s.Note || assertions != 0 {
		return fmt.Errorf("unsupported generic case differs from native no-request declaration")
	}
	return nil
}

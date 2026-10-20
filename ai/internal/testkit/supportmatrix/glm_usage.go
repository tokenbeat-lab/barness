package supportmatrix

import (
	"github.com/tokenbeat-lab/barness/ai"
	"path/filepath"
)

// Attempt usage is authoritative. These sums retain only known reported
// counters: cancellation and partial/missing reporting remain unknown, and
// reasoning is already included in Output, never charged a second time.
func glmKnownUsage(dir string, r *Report) error {
	for _, s := range r.Scenarios {
		if s.Attempts == 0 {
			continue
		}
		var outcome struct {
			Result struct{ Metadata ai.CallMetadata }
		}
		if readGLMJSON(filepath.Join(dir, s.CaseID, "attempt-1-result.json"), &outcome) != nil {
			return glmBad("known usage capture unavailable")
		}
		for _, attempt := range outcome.Result.Metadata.Attempts {
			u := attempt.Usage
			r.Budget.InputTokensUsed += int(u.Input + u.CacheRead + u.CacheWrite)
			r.Budget.OutputTokensUsed += int(u.Output)
			if attempt.UsageReporting != ai.UsageComplete {
				r.GLM.UnknownUsageAttempts++
			}
		}
	}
	return nil
}

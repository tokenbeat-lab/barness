package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// This is an assertion on paired public runs, not a projection filter: every
// duration field still participates in the unmodified whole-observation diff.
func checkObservationDuration(t *testing.T, ev *evidence.Case, side string, obs pioracle.Observation) {
	t.Helper()
	var result struct {
		Duration *int64 `json:"durationMs"`
	}
	err := json.Unmarshal(obs.Result, &result)
	if !ev.Check(side+" final duration is a nonnegative integer", err == nil && result.Duration != nil && *result.Duration >= 0, "result=%s", obs.Result) {
		return
	}
	terminals := 0
	for _, raw := range obs.Events {
		var e struct {
			Type    string `json:"type"`
			Message struct {
				Duration *int64 `json:"durationMs"`
			} `json:"message"`
			Error struct {
				Duration *int64 `json:"durationMs"`
			} `json:"error"`
		}
		if json.Unmarshal(raw, &e) != nil {
			continue
		}
		var duration *int64
		switch e.Type {
		case "done":
			duration = e.Message.Duration
		case "error":
			duration = e.Error.Duration
		default:
			continue
		}
		terminals++
		ev.Check(side+" terminal/result duration agree", duration != nil && *duration == *result.Duration, "event=%s result=%s", raw, obs.Result)
	}
	ev.Check(side+" has one terminal", terminals == 1, "terminals=%d", terminals)
}

package release

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The live stage records the new thinking-on process. Its consumer may additionally
// reference the independently preserved old thirteen; neither old commit nor old
// calls are rewritten to look like part of the new execution.
func verifySelectedMiniMaxThinking(root, newBundle, composition, route string, stage StageRecord, current, measured ExecutionSource) error {
	q, err := supportmatrix.VerifyMiniMaxThinkingComposition(composition, true)
	if err != nil {
		return err
	}
	if q.Combo != route || len(q.Measurements) != 2 || len(stage.Consumers) != 1 {
		return fmt.Errorf("MiniMax thinking-on selected composition route differs")
	}
	supplement := q.Measurements[1]
	bundle, err := recordPath(root, supplement.Evidence.Bundle)
	if err != nil || filepath.Clean(bundle) != filepath.Clean(newBundle) || supplement.Original.GitCommit != measured.Commit || supplement.Evidence.Producer.Path != stage.Receipt.Path || supplement.Evidence.Producer.SHA256 != stage.Receipt.SHA256 {
		return fmt.Errorf("MiniMax thinking-on stage does not identify the actual new measurement")
	}
	// Stage receipt verification already bound measured source and command. The
	// descriptor additionally pins that exact receipt, not just a matching label.
	if len(stage.RequiredCases) != len(q.Expected) {
		return fmt.Errorf("MiniMax thinking-on release requires all fourteen mapped capabilities")
	}
	seen := map[string]bool{}
	for _, id := range stage.RequiredCases {
		if id == "" || seen[id] || !slices.ContainsFunc(q.Expected, func(expected string) bool { return q.Coverage[expected].CaseID == id }) {
			return fmt.Errorf("MiniMax thinking-on required provenance missing or duplicated")
		}
		seen[id] = true
	}
	return verifyLiveConsumer(root, composition, stage.Consumers[0], current, liveReportHash(q))
}

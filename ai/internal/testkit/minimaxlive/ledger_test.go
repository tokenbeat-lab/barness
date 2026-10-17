package minimaxlive_test

import (
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// Failure inventory is recorded before this auxiliary isolated slice in
// .scratch/barness-ai-minimax/evidence/ticket-07/failure-inventory.md.
func TestStageLedgerSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.jsonl")
	id, err := minimaxlive.Select("chat", "MiniMax-M3", "standard")
	if err != nil {
		t.Fatal(err)
	}
	for n := 0; n < 20; n++ {
		ledger := minimaxlive.Ledger{Path: path, Stage: "candidate", Identity: id}
		if _, err := ledger.Reserve("call-"+string(rune('a'+n)), 1000, 2048); err != nil {
			t.Fatal(err)
		}
	}
	ledger := minimaxlive.Ledger{Path: path, Stage: "candidate", Identity: id}
	if _, err := ledger.Reserve("restart-bypass", 1000, 2048); err == nil {
		t.Fatal("restart bypassed cumulative identity call ceiling")
	}
}

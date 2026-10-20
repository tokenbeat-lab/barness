package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Explicit diagnostic only: adopt the two immutable original executions,
// preserve their failures, and never attribute their commands to current Go.
func TestReleaseRecordedBundleDiagnostic(t *testing.T) {
	raw := os.Getenv("BARNESS_RELEASE_DIAGNOSTIC_BUNDLES")
	if raw == "" {
		t.Skip("optional read-only recorded-bundle diagnosis; no new suite certification")
	}
	ev := run.Case(t, "OD57-release-recorded-bundle-diagnostic")
	var dirs []string
	if err := json.Unmarshal([]byte(raw), &dirs); err != nil || len(dirs) != 2 {
		t.Fatal("diagnostic requires exactly two recorded bundle paths")
	}
	wantHashes := []string{
		"b4dcf3e8772285e0cd06e03dc627be0dad33315199d8e9965c734187407af21e",
		"14b1f2b38e69e62ac628554c8db436390aeeae682086f0dd9291ab3865a2c907",
	}
	wantFails := []string{"GLM12-formal-import-chat", "GLM12-formal-import-messages", "GLM12-formal-import-responses", "GLM12-reported-known-usage-chat", "GLM12-reported-known-usage-messages", "GLM12-reported-known-usage-responses"}
	for i, dir := range dirs {
		before, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		if err != nil {
			t.Fatal("recorded manifest unavailable")
		}
		if !ev.Check("literal original manifest", pioracle.SHA256(before) == wantHashes[i], "wrong historical input") {
			continue
		}
		b, err := release.LoadBundle(dir)
		if !ev.Check("recorded recursive inventory imports", err == nil && len(b.Problems) == 0, "import error=%t inventory problems=%d", err != nil, len(b.Problems)) {
			continue
		}
		var failed []string
		passed := 0
		for _, c := range b.Cases {
			if c.Status == release.Fail {
				failed = append(failed, c.ID)
			}
			if c.Status == release.Pass {
				passed++
			}
		}
		slices.Sort(failed)
		ev.Check("original outcomes preserved", len(b.Cases) == 12391 && passed == 12385 && slices.Equal(failed, wantFails), "cases=%d pass=%d fail=%d", len(b.Cases), passed, len(failed))
		for _, id := range releaseEmptyEffortIDs() {
			ev.Check("raw original identity "+id, slices.Contains(b.Cases, release.Case{ID: id, Status: release.Pass}), "identity changed")
		}
		after, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
		ev.Check("original manifest byte immutable", err == nil && pioracle.SHA256(after) == wantHashes[i], "historical input was modified")
		ev.Record([]string{"ordinary", "race"}[i], map[string]any{"diagnosticOnly": true, "sourceCommit": b.Commit, "manifestSHA256": wantHashes[i], "cases": len(b.Cases), "passed": passed, "failed": failed, "inventoryProblems": b.Problems, "newCommandsExecuted": false})
	}
}

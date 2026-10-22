package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// These exact empty-effort identifiers are emitted by the actual public E2E.
// Empty effort means omission; the trailing separator is part of the raw ID.
func releaseEmptyEffortIDs() []string {
	var ids []string
	for _, model := range []string{"MiniMax-M3.1-Flash-Preview", "MiniMax-M3"} {
		for _, source := range []string{"full", "hook", "sampling", "simple"} {
			ids = append(ids, "MM03-effort-TestMiniMaxResponsesEffort/"+model+"/"+source+"/")
		}
	}
	return ids
}

func TestReleaseProducerNestedEmptyEffortDirectories(t *testing.T) {
	ev := run.Case(t, "OD57-release-producer-empty-effort-directories")
	t.Setenv(evidence.DirEnv, t.TempDir())
	producer, err := evidence.NewRun(release.OfflinePackage)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range releaseEmptyEffortIDs() {
		for _, rawID := range []string{id, id + "high"} {
			t.Run(rawID, func(t *testing.T) {
				c := producer.Case(t, rawID)
				c.Record("result", map[string]any{"fixtureOnly": true, "visibleText": "ok"})
				c.Check("controlled producer case", true, "no model call")
			})
		}
	}
	if err := producer.Finish(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(producer.Dir(), "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("producer-manifest.json", raw)
	b, err := release.LoadBundle(producer.Dir())
	ev.Record("import", map[string]any{"error": err != nil, "cases": b.Cases, "problems": b.Problems})
	if !ev.Check("recursive producer inventory imports", err == nil && len(b.Problems) == 0, "error=%v problems=%v", err, b.Problems) {
		return
	}
	ev.Check("all raw identities preserved", len(b.Cases) == 16, "cases=%d", len(b.Cases))
	for _, id := range releaseEmptyEffortIDs() {
		ev.Check("empty identity preserved "+id, slices.Contains(b.Cases, release.Case{ID: id, Status: release.Pass}), "raw ID changed")
		assertions, e := os.ReadFile(filepath.Join(producer.Dir(), id, "assertions.json"))
		var root struct {
			CaseID string `json:"case_id"`
		}
		e2 := json.Unmarshal(assertions, &root)
		ev.Check("parent assertion attribution "+id, e == nil && e2 == nil && root.CaseID == id, "parent assertion replaced by child")
	}
}

func TestReleaseBundleRootSelection(t *testing.T) {
	for _, mode := range []string{"directory", "symlink", "symlink-trailing-slash"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-root-selection-"+mode)
			dir := releaseRollbackBundle(t)
			if mode != "directory" {
				outside := filepath.Join(t.TempDir(), "original")
				if err := os.Rename(dir, outside); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, dir); err != nil {
					t.Fatal(err)
				}
				if mode == "symlink-trailing-slash" {
					dir += "/"
				}
			}
			b, err := release.LoadBundle(dir)
			accepted := err == nil && len(b.Problems) == 0
			ev.Record("selection", map[string]any{"accepted": accepted, "error": err != nil, "problems": len(b.Problems)})
			ev.Check("only a real selected bundle directory imports", accepted == (mode == "directory"), "accepted=%t", accepted)
		})
	}
}

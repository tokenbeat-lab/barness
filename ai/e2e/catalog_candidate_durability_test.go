package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogCandidateApplyDirectorySyncFailure(t *testing.T) {
	ev := run.Case(t, "CATALOG-T06-post-rename-sync-failure")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	root := catalogSnapshotRoot(t)
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	candidate := freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "directory durability failure retains baseline")
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	json.Unmarshal(out, &checked)
	if !ev.Check("candidate has actual contract proof", exit == 0 && checked.Status == "applicable", "exit=%d output=%s", exit, out) {
		return
	}
	before, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	if !ev.Check("baseline available", exit == 0, "exit=%d", exit) {
		return
	}
	for _, target := range []string{"output", "current"} {
		for _, existed := range []bool{true, false} {
			outputDir := t.TempDir()
			output := filepath.Join(outputDir, "runtime.json")
			old := []byte("previous runtime bytes\n")
			if existed {
				os.WriteFile(output, old, 0600)
			}
			parent := outputDir
			if target == "current" {
				parent = store
			}
			t.Cleanup(func() { os.Chmod(parent, 0700) })
			if err := os.Chmod(parent, 0300); err != nil {
				t.Fatal(err)
			}
			// This real filesystem state allows create/rename while denying the open
			// used for directory fsync. No production hook or injected error is used.
			handle, openErr := os.Open(parent)
			if handle != nil {
				handle.Close()
			}
			if openErr == nil {
				os.Chmod(parent, 0700)
				t.Fatal("filesystem did not deny directory open")
			}
			out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", checked.Reference, "--out", output, "--confirm")
			if err := os.Chmod(parent, 0700); err != nil {
				t.Fatal(err)
			}
			ev.Check("durability failure has nonzero structured exit", exit != 0 && bytes.Contains(out, []byte(`"category":"storage_failed"`)), "target=%s existed=%v exit=%d output=%s", target, existed, exit, out)
			after, readErr := os.ReadFile(output)
			restored := (existed && readErr == nil && bytes.Equal(after, old)) || (!existed && os.IsNotExist(readErr))
			ev.Check("failed apply restores prior output existence and bytes", restored, "target=%s existed=%v err=%v bytes=%d", target, existed, readErr, len(after))
			current, currentExit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
			ev.Check("failed apply retains exact effective baseline", currentExit == 0 && bytes.Equal(current, before), "target=%s existed=%v exit=%d", target, existed, currentExit)
		}
	}
}

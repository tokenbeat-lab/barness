package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogMaintenanceCLIRevisionIntegrity(t *testing.T) {
	ev := run.Case(t, "CATALOG-T06-cli-current-revision-integrity")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	candidate := freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "revision integrity fixture")
	revision := checkApplyCatalogAction(t, ev, binary, store, candidate)
	path := ""
	filepath.WalkDir(store, func(name string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		raw, err := os.ReadFile(name)
		if err != nil {
			return err
		}
		var got maintenance.Revision
		if json.Unmarshal(raw, &got) == nil && got.ID == revision.ID {
			path = name
		}
		return nil
	})
	if path == "" {
		t.Fatal("applied revision object not retained")
	}
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("approved-revision.json", original)
	revision.Snapshot = []byte(`{"hash":"unapproved","catalog":{"version":"replacement","models":[]}}`)
	corrupt, _ := json.Marshal(revision)
	os.WriteFile(path, corrupt, 0600)
	ev.Fixture("corrupt-revision.json", corrupt)
	for _, command := range []string{"export", "versions", "history"} {
		out, exit := invokeCatalogCLI(t, ev, binary, command, "--store", store)
		ev.Check("tampered current snapshot is refused rather than exported", exit != 0, "command=%s exit=%d output=%s", command, exit, out)
	}
}

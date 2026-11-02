package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// A fabricated source and model exercise the real CLI without claiming a
// vendor offering. Its candidate and output remain in temporary directories.
func TestCatalogFutureModelCLIJSONOnly(t *testing.T) {
	ev := run.Case(t, "FUTURE-MODEL-cli-json-only")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	source := filepath.Join(dir, "synthetic.txt")
	definition := filepath.Join(dir, "definition.json")
	model := futureCatalogCases(t)[0].cat.Models[0]
	os.WriteFile(source, []byte("Synthetic source for controlled contract only.\n"), 0600)
	os.WriteFile(definition, mustMarshal(t, model), 0600)
	args := []string{"add", "--store", store, "--operation", "chat", "--provider", string(model.Provider), "--api", string(model.API), "--model", model.ID, "--source-kind", "official", "--source", source, "--source-url", "https://synthetic.invalid/catalog", "--source-time", "2026-11-02T00:00:00Z", "--definition", definition, "--reason", "synthetic public JSON-only test"}
	preview, exit := invokeCatalogCLI(t, ev, binary, append(args, "--dry-run")...)
	ev.Check("preview succeeds without mutation", exit == 0, "exit=%d output=%s", exit, preview)
	out, exit := invokeCatalogCLI(t, ev, binary, append(args, "--confirm")...)
	var frozen struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	_ = json.Unmarshal(out, &frozen)
	if !ev.Check("synthetic candidate freezes", exit == 0 && frozen.Candidate.ID != "", "exit=%d output=%s", exit, out) {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", catalogSnapshotRoot(t), "--evidence-feature", "catalog-model-id-policy", "--confirm")
	var check maintenance.CheckResult
	_ = json.Unmarshal(out, &check)
	if !ev.Check("fixed public contracts accept synthetic model", exit == 0 && check.Status == "applicable", "exit=%d output=%s", exit, out) {
		return
	}
	output := filepath.Join(dir, "runtime.json")
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", frozen.Candidate.ID, "--repo", catalogSnapshotRoot(t), "--check", check.Reference, "--out", output, "--confirm")
	raw, err := os.ReadFile(output)
	cat, report, loadErr := catalog.LoadFile(output)
	got, found := cat.Lookup(model.Provider, model.API, model.ID)
	ev.Check("apply installs exact tested JSON and identity", exit == 0 && err == nil && bytes.Equal(raw, frozen.Candidate.Snapshot) && loadErr == nil && report.Failed == 0 && found && got.ID == model.ID, "exit=%d err=%v load=%v output=%s", exit, err, loadErr, out)
}

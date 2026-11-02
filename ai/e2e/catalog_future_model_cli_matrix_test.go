package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Each artificial identity has a separate store and source. The source is a
// test fixture and is not evidence that a vendor offers any such model.
func TestCatalogFutureModelCLIAllIdentities(t *testing.T) {
	previous := ai.BuiltinCatalog()
	buildEvidence := run.Case(t, "FUTURE-MODEL-cli-matrix-build")
	binary := buildCatalogCLI(t, buildEvidence)
	if binary == "" {
		t.FailNow()
	}
	for _, tc := range futureCatalogCases(t) {
		t.Run(tc.name, func(t *testing.T) {
			ev := run.Case(t, "FUTURE-MODEL-cli-"+tc.name)
			dir := t.TempDir()
			store := filepath.Join(dir, "store")
			source := filepath.Join(dir, "synthetic.txt")
			definition := filepath.Join(dir, "definition.json")
			var model any
			switch tc.identity.Operation {
			case ai.OperationChat:
				model = tc.cat.Models[0]
			case ai.OperationImage:
				model = tc.cat.ImageModels[0]
			case ai.OperationClassifier:
				model = tc.cat.ClassifierModels[0]
			}
			capture := []byte("Synthetic test source; no vendor offering or qualification.\n")
			if err := os.WriteFile(source, capture, 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(definition, mustMarshal(t, model), 0600); err != nil {
				t.Fatal(err)
			}
			ev.Fixture("synthetic-source.txt", capture)
			ev.Record("synthetic-definition", model)
			args := []string{"add", "--store", store, "--operation", string(tc.identity.Operation), "--provider", string(tc.identity.Provider), "--api", string(tc.identity.API), "--model", tc.identity.ID, "--source-kind", "official", "--source", source, "--source-url", "https://synthetic.invalid/catalog", "--source-time", "2026-11-02T00:00:00Z", "--definition", definition, "--reason", "synthetic independent JSON-only maintenance test"}
			preview, exit := invokeCatalogCLI(t, ev, binary, append(append([]string{}, args...), "--dry-run")...)
			_, statErr := os.Stat(store)
			if !ev.Check("preview is successful and read-only", exit == 0 && os.IsNotExist(statErr), "exit=%d err=%v output=%s", exit, statErr, preview) {
				t.FailNow()
			}
			out, exit := invokeCatalogCLI(t, ev, binary, append(append([]string{}, args...), "--confirm")...)
			var frozen struct {
				Candidate maintenance.Candidate `json:"candidate"`
			}
			_ = json.Unmarshal(out, &frozen)
			if !ev.Check("synthetic definition freezes", exit == 0 && frozen.Candidate.ID != "", "exit=%d output=%s", exit, out) {
				t.FailNow()
			}
			out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", catalogSnapshotRoot(t), "--evidence-feature", "catalog-model-id-policy", "--confirm")
			var check maintenance.CheckResult
			_ = json.Unmarshal(out, &check)
			if !ev.Check("fixed public contracts validate this candidate", exit == 0 && check.Status == "applicable" && len(check.RequiredCases) > 0, "exit=%d output=%s", exit, out) {
				t.FailNow()
			}
			ev.Record("check-reference", check)
			output := filepath.Join(dir, "runtime.json")
			out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", frozen.Candidate.ID, "--repo", catalogSnapshotRoot(t), "--check", check.Reference, "--out", output, "--confirm")
			raw, readErr := os.ReadFile(output)
			if !ev.Check("apply installs exact tested bytes", exit == 0 && readErr == nil && bytes.Equal(raw, frozen.Candidate.Snapshot), "exit=%d err=%v output=%s", exit, readErr, out) {
				t.FailNow()
			}
			loaded, report, loadErr := catalog.LoadFile(output)
			e := historicalCatalogInCurrentOrder(loaded, previous)
			ev.Check("all prior builtin identities and order remain unchanged", loadErr == nil && report.Failed == 0 && reflect.DeepEqual(e, previous), "err=%v report=%+v", loadErr, report)
			var found bool
			switch tc.identity.Operation {
			case ai.OperationChat:
				_, found = loaded.Lookup(tc.identity.Provider, tc.identity.API, tc.identity.ID)
			case ai.OperationImage:
				_, found = loaded.LookupImage(tc.identity.Provider, tc.identity.API, tc.identity.ID)
			case ai.OperationClassifier:
				_, found = loaded.LookupClassifier(tc.identity.Provider, tc.identity.API, tc.identity.ID)
			}
			total := len(loaded.Models) + len(loaded.ImageModels) + len(loaded.ClassifierModels)
			ev.Check("only the selected identity is added", found && total == len(previous.Models)+len(previous.ImageModels)+len(previous.ClassifierModels)+1, "identity=%+v total=%d", tc.identity, total)
			ev.Fixture("applied-runtime.json", raw)
			ev.Record("identity", fmt.Sprint(tc.identity))
		})
	}
}

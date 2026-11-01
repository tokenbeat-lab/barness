package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogCandidateQueryStatus(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-read-only-candidate-status")
	binary, store, candidate, check := frozenCheckedCatalog(t, ev)
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference)
	var response struct {
		Status string                  `json:"status"`
		Check  maintenance.CheckResult `json:"check"`
	}
	_ = json.Unmarshal(out, &response)
	ev.Check("query exposes independently verified applicable state", exit == 0 && response.Status == "applicable" && response.Check.Conclusion == "code_compatible", "%s", out)
	fake := check
	fake.Reference = filepath.ToSlash(filepath.Join(filepath.Dir(check.Reference), "self-report-query.json"))
	fake.Prepared = ""
	fake.Receipt = ""
	fake.ReceiptSHA256 = ""
	if err := os.WriteFile(filepath.Join(root, fake.Reference), mustMarshal(t, fake), 0600); err != nil {
		t.Fatal(err)
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", candidate.ID, "--repo", root, "--check", fake.Reference)
	ev.Check("query never promotes a self-reported compatible state", exit != 0, "%s", out)
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "catalog-query-dry-run-absent", "--dry-run")
	var preview maintenance.CheckResult
	_ = json.Unmarshal(out, &preview)
	ev.Check("check preview is read-only machine state", exit == 0 && preview.Status == "awaiting_check", "%s", out)
	_, err := os.Stat(filepath.Join(root, ".evidence", "catalog-query-dry-run-absent"))
	ev.Check("dry-run creates no evidence root", os.IsNotExist(err), "%v", err)
	_, err = os.Stat(filepath.Join(store, "current.json"))
	ev.Check("query never advances effective version", os.IsNotExist(err), "%v", err)
}

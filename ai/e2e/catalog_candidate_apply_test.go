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

func TestCatalogCandidateCheckApply(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-public-check-apply")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	root := catalogSnapshotRoot(t)
	store := filepath.Join(t.TempDir(), "store")
	out, exit := invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "controlled withdrawal", "--confirm")
	var frozen struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	_ = json.Unmarshal(out, &frozen)
	ev.Check("candidate frozen", exit == 0 && frozen.Candidate.ID != "", "%s", out)
	if exit != 0 {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	_ = json.Unmarshal(out, &checked)
	ev.Check("actual controlled check passes", exit == 0 && checked.Status == "applicable" && len(checked.RequiredCases) > 0, "exit=%d %s", exit, out)
	if exit != 0 {
		return
	}
	_, err := os.Stat(filepath.Join(store, "current.json"))
	ev.Check("check leaves effective revision untouched", os.IsNotExist(err), "%v", err)
	output := filepath.Join(t.TempDir(), "runtime.json")
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--check", checked.Reference, "--out", output, "--confirm")
	ev.Check("checked candidate applies", exit == 0, "exit=%d %s", exit, out)
	if exit != 0 {
		return
	}
	raw, err := os.ReadFile(output)
	ev.Check("apply writes exact tested bytes", err == nil && bytes.Equal(raw, frozen.Candidate.Snapshot), "%v", err)
	cat, report, err := catalog.Load(raw)
	_, found := cat.Lookup("openai", "openai-responses", "gpt-4")
	ev.Check("withdrawn identity not restored from builtin", err == nil && report.Failed == 0 && !found, "%v %+v", err, report)
	out, exit = invokeCatalogCLI(t, ev, binary, "history", "--store", store)
	var history struct {
		History []maintenance.Revision `json:"history"`
	}
	_ = json.Unmarshal(out, &history)
	ev.Check("append-only history retains new and prior revision", exit == 0 && len(history.History) == 2, "%s", out)
}

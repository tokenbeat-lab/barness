package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogCandidateDataOnlyAndCodeGap(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-data-only-code-gap")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	root := catalogSnapshotRoot(t)
	for _, api := range []string{"openai-completions", "future-catalog-api"} {
		directory := t.TempDir()
		store := filepath.Join(directory, "store")
		source := filepath.Join(directory, "source.json")
		if err := os.WriteFile(source, []byte(maintenanceSourceFixture), 0600); err != nil {
			t.Fatal(err)
		}
		ev.Fixture("synthetic-models-dev.json", []byte(maintenanceSourceFixture))
		args := []string{"add", "--store", store, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", api, "--model", "fixture-chat", "--reason", "controlled data-only model", "--confirm"}
		out, exit := invokeCatalogCLI(t, ev, binary, args...)
		var frozen struct {
			Candidate maintenance.Candidate `json:"candidate"`
		}
		mustUnmarshal(t, out, &frozen)
		if !ev.Check("explicit source/API candidate freezes", exit == 0, "%s", out) {
			continue
		}
		out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
		if api == "future-catalog-api" {
			var failed struct {
				Category string                  `json:"category"`
				Result   maintenance.CheckResult `json:"result"`
			}
			_ = json.Unmarshal(out, &failed)
			ev.Check("missing route is code gap with retained candidate", exit != 0 && failed.Category == "code_gap" && failed.Result.Reference != "", "%s", out)
			_, err := os.Stat(filepath.Join(store, "current.json"))
			ev.Check("code gap never changes effective directory", os.IsNotExist(err), "%v", err)
			query, queryExit := invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--check", failed.Result.Reference)
			var state struct {
				Status string `json:"status"`
			}
			_ = json.Unmarshal(query, &state)
			ev.Check("query distinguishes retained code gap", queryExit == 0 && state.Status == "code_gap", "%s", query)
			continue
		}
		var check maintenance.CheckResult
		_ = json.Unmarshal(out, &check)
		if !ev.Check("new identity uses current generic contract", exit == 0 && check.Status == "applicable", "%s", out) {
			continue
		}
		if err := os.Remove(source); err != nil {
			t.Fatal(err)
		}
		output := filepath.Join(directory, "runtime.json")
		out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--check", check.Reference, "--out", output, "--confirm")
		ev.Check("apply uses frozen bytes without reacquiring deleted source", exit == 0, "%s", out)
		raw, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		cat, report, err := catalog.Load(raw)
		model, found := cat.Lookup("openai", "openai-completions", "fixture-chat")
		ev.Check("data-only accepted model remains queryable and priced", err == nil && report.Failed == 0 && found && model.Cost.Input == 1 && model.Cost.Output == 2, "%v %+v", err, model)
	}
}

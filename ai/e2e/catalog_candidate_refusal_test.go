package e2e

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func frozenCheckedCatalog(t *testing.T, ev *evidence.Case) (string, string, maintenance.Candidate, maintenance.CheckResult) {
	t.Helper()
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		t.FailNow()
	}
	store := filepath.Join(t.TempDir(), "store")
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "refusal fixture", "--confirm")
	var response struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	mustUnmarshal(t, out, &response)
	if !ev.Check("freeze refusal candidate", exit == 0, "%s", out) {
		t.FailNow()
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", response.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var check maintenance.CheckResult
	_ = json.Unmarshal(out, &check)
	if !ev.Check("real controlled fixture check", exit == 0 && check.Status == "applicable", "%s", out) {
		t.FailNow()
	}
	return binary, store, response.Candidate, check
}

func TestCatalogCandidateApplyRefusals(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-apply-refusals")
	binary, store, candidate, check := frozenCheckedCatalog(t, ev)
	root := catalogSnapshotRoot(t)
	for _, location := range []string{"candidate", "current", "history", "evidence"} {
		clone := filepath.Join(t.TempDir(), "store")
		catalogCopyTree(t, store, clone)
		output := filepath.Join(clone, "current.json")
		switch location {
		case "candidate":
			output = filepath.Join(clone, "candidates", strings.TrimPrefix(candidate.ID, "sha256:")+".json")
		case "history":
			output = filepath.Join(clone, "revisions", "new.json")
		case "evidence":
			output = filepath.Join(root, filepath.Dir(check.Reference), "unsafe-output.json")
		}
		out, exit := invokeCatalogCLI(t, ev, binary, "apply", "--store", clone, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--out", output, "--confirm")
		ev.Check("protected output path refused "+location, exit != 0, "exit=%d %s", exit, out)
		_, err := os.Stat(filepath.Join(clone, "current.json"))
		ev.Check("protected output never commits "+location, os.IsNotExist(err), "%v", err)
	}
	// A caller may report PASS, but no completed native source/scenario evidence
	// exists for this new reference. The CLI must refuse it before any effect.
	fake := filepath.Join(root, filepath.Dir(check.Reference), "self-reported.json")
	raw := []byte(`{"status":"applicable","pass":true}`)
	if err := os.WriteFile(fake, raw, 0600); err != nil {
		t.Fatal(err)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", fake, "--confirm")
	ev.Check("self-reported pass refused", exit != 0, "%s", out)
	// Two independently confirmed candidates may have the same runtime bytes;
	// a check for one must not be rebound by editing its summary for the other.
	out, exit = invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "different reviewed reason", "--confirm")
	var second struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	mustUnmarshal(t, out, &second)
	ev.Check("separate candidate identity", exit == 0 && second.Candidate.ID != candidate.ID, "%s", out)
	forged := check
	forged.CandidateID = second.Candidate.ID
	forged.Reference = filepath.ToSlash(filepath.Join(filepath.Dir(check.Reference), "rebound.json"))
	raw = mustMarshal(t, forged)
	if err := os.WriteFile(filepath.Join(root, forged.Reference), raw, 0600); err != nil {
		t.Fatal(err)
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", second.Candidate.ID, "--repo", root, "--check", forged.Reference, "--dry-run")
	ev.Check("evidence cannot be rebound to a different candidate", exit != 0, "%s", out)
	// File-set drift matters even when every previously measured file is intact.
	source := filepath.Join(root, "catalog", "candidate_stale_scope_fixture.go")
	if err := os.WriteFile(source, []byte("package catalog\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--dry-run")
	if err := os.Remove(source); err != nil {
		t.Fatal(err)
	}
	ev.Check("added implementation source invalidates check", exit != 0, "%s", out)
	// A pre-existing source changed in place is independently rejected.
	source = filepath.Join(root, "catalog", "report.go")
	before, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, append(append([]byte(nil), before...), '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--dry-run")
	if err := os.WriteFile(source, before, 0600); err != nil {
		t.Fatal(err)
	}
	ev.Check("changed actual source invalidates check", exit != 0, "%s", out)
	badOutput := filepath.Join(t.TempDir(), "missing", "runtime.json")
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", check.Reference, "--out", badOutput, "--confirm")
	ev.Check("output failure reported", exit != 0, "%s", out)
	_, err = os.Stat(filepath.Join(store, "current.json"))
	ev.Check("output failure leaves effective revision untouched", os.IsNotExist(err), "%v", err)
}

func catalogCopyTree(t *testing.T, from, to string) {
	t.Helper()
	err := filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		target := filepath.Join(to, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

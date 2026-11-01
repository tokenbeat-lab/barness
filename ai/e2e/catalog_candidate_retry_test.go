package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// An output failure may retain immutable orphan revisions. A new valid check
// of the same state must recover through normal apply without deleting or
// rewriting that failed attempt's application metadata.
func TestCatalogCandidateApplyRetry(t *testing.T) {
	for _, mode := range []string{"new-check", "new-candidate"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "CATALOG-CHECK-apply-retry-"+mode)
			binary := buildSingleModelCatalogCLI(t, ev)
			if binary == "" {
				return
			}
			directory := t.TempDir()
			store := filepath.Join(directory, "store")
			freeze := func(reason string) maintenance.Candidate {
				return freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", reason)
			}
			first := freeze("first failed-output attempt")
			check := catalogRetryCheck(t, ev, binary, store, first)
			root := catalogSnapshotRoot(t)
			out, exit := invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", first.ID, "--repo", root, "--check", check.Reference, "--out", filepath.Join(directory, "absent-parent", "runtime.json"), "--confirm")
			var failure struct{ Category string }
			mustUnmarshal(t, out, &failure)
			ev.Check("output save failure is reported", exit != 0 && failure.Category == "storage_failed", "exit=%d %s", exit, out)
			_, err := os.Stat(filepath.Join(store, "current.json"))
			ev.Check("failed output does not advance current", os.IsNotExist(err), "%v", err)
			orphans := catalogRetryOrphans(t, ev, store)
			orphanID := ""
			for _, raw := range orphans {
				var revision maintenance.Revision
				mustUnmarshal(t, raw, &revision)
				if revision.CheckReference == check.Reference {
					orphanID = revision.ID
				}
			}
			if !ev.Check("failed apply retains its original orphan application", len(orphans) == 2 && orphanID != "", "objects=%d orphan=%s", len(orphans), orphanID) {
				return
			}
			out, exit = invokeCatalogCLI(t, ev, binary, "history", "--store", store)
			var history struct {
				History []maintenance.Revision `json:"history"`
			}
			mustUnmarshal(t, out, &history)
			ev.Check("orphan is absent from effective history", exit == 0 && len(history.History) == 1 && history.History[0].ID == first.BaseDigest, "exit=%d %s", exit, out)
			candidate := first
			if mode == "new-candidate" {
				candidate = freeze("second independently reviewed reason")
				ev.Check("different reasons retain distinct candidates for identical state", candidate.ID != first.ID && bytes.Equal(mustMarshal(t, candidate.State), mustMarshal(t, first.State)) && bytes.Equal(candidate.Snapshot, first.Snapshot), "first=%s second=%s", first.ID, candidate.ID)
			}
			fresh := catalogRetryCheck(t, ev, binary, store, candidate)
			ev.Check("retry has fresh proof for the same baseline", fresh.Reference != check.Reference && fresh.BaseDigest == check.BaseDigest, "old=%+v fresh=%+v", check, fresh)
			output := filepath.Join(directory, "runtime.json")
			out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", fresh.Reference, "--out", output, "--confirm")
			for name, original := range orphans {
				raw, err := os.ReadFile(filepath.Join(store, "revisions", name))
				ev.Check("retry preserves failed original bytes", err == nil && bytes.Equal(raw, original), "object=%s err=%v", name, err)
			}
			if !ev.Check("fresh proof recovers from failed output without immutable collision", exit == 0, "exit=%d %s", exit, out) {
				return
			}
			var applied struct {
				Revision maintenance.Revision `json:"revision"`
			}
			mustUnmarshal(t, out, &applied)
			revision := applied.Revision
			ev.Check("new application identity binds its exact candidate and check", revision.ID != orphanID && revision.CandidateID == candidate.ID && revision.CheckReference == fresh.Reference && revision.Parent == candidate.BaseDigest, "%+v", revision)
			raw, err := os.ReadFile(output)
			ev.Check("retry writes the exact tested snapshot", err == nil && bytes.Equal(raw, candidate.Snapshot), "%v", err)
			out, exit = invokeCatalogCLI(t, ev, binary, "history", "--store", store)
			mustUnmarshal(t, out, &history)
			ev.Check("successful retry adds only its committed application to history", exit == 0 && len(history.History) == 2 && history.History[0].ID == revision.ID && history.History[1].ID == candidate.BaseDigest, "exit=%d %s", exit, out)
			// A revision's check reference is immutable identity, as are its state,
			// snapshot, parent and candidate. Reading must reject an altered body.
			path := filepath.Join(store, "revisions", strings.TrimPrefix(revision.ID, "sha256:")+".json")
			original, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("approved-retry-revision.json", original)
			revision.CheckReference = "different-check-reference"
			corrupt := mustMarshal(t, revision)
			ev.Fixture("altered-retry-revision.json", corrupt)
			if err := os.WriteFile(path, corrupt, 0600); err != nil {
				t.Fatal(err)
			}
			out, exit = invokeCatalogCLI(t, ev, binary, "export", "--store", store)
			_ = json.Unmarshal(out, &failure)
			ev.Check("application metadata tamper is refused during public read", exit != 0 && failure.Category == "revision_changed", "exit=%d %s", exit, out)
		})
	}
}

func catalogRetryCheck(t *testing.T, ev *evidence.Case, binary, store string, candidate maintenance.Candidate) maintenance.CheckResult {
	t.Helper()
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", catalogSnapshotRoot(t), "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	mustUnmarshal(t, out, &checked)
	if !ev.Check("retry candidate passes actual fixed contracts", exit == 0 && checked.Status == "applicable", "exit=%d %s", exit, out) {
		t.FailNow()
	}
	return checked
}

func catalogRetryOrphans(t *testing.T, ev *evidence.Case, store string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(store, "revisions"))
	if err != nil {
		t.Fatal(err)
	}
	result := map[string][]byte{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(store, "revisions", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		result[entry.Name()] = raw
		ev.Fixture("failed-original-"+entry.Name(), raw)
	}
	return result
}

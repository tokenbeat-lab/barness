package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Failure inventory: recover only the catalog pinned by the measured commit,
// version and hash. Missing objects, substituted data, mutable worktree data and
// changed current model facts must never silently qualify an old measurement.
func TestHistoricalCatalogFromMeasuredCommit(t *testing.T) {
	for _, mode := range []string{"valid", "dirty-worktree", "missing-commit", "invalid-commit", "missing-snapshot", "malformed", "wrong-hash", "wrong-version", "model-drift"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "PD-catalog-"+mode)
			repo := t.TempDir()
			git := func(args ...string) string {
				t.Helper()
				cmd := exec.CommandContext(ctxFor(t), "git", append([]string{"-C", repo}, args...)...)
				for _, value := range xaiProcessEnv(t) {
					if !strings.HasPrefix(value, "GIT_") {
						cmd.Env = append(cmd.Env, value)
					}
				}
				raw, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("private Git fixture: %v %s", err, raw)
				}
				return strings.TrimSpace(string(raw))
			}
			git("init", "--quiet")
			catalog := ai.BuiltinCatalog()
			catalog.Version = "TEST_ONLY_historical-catalog"
			hash, err := catalog.Hash()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(repo, "ai/release"), 0700); err != nil {
				t.Fatal(err)
			}
			writeReleaseJSON(t, repo, "ai/release/catalog-snapshot.json", map[string]any{"hash": hash, "catalog": catalog})
			if mode == "missing-snapshot" {
				if err := os.Remove(filepath.Join(repo, "ai/release/catalog-snapshot.json")); err != nil {
					t.Fatal(err)
				}
			} else if mode == "malformed" {
				mustWriteXAINative(t, filepath.Join(repo, "ai/release/catalog-snapshot.json"), []byte("{"))
			}
			git("add", ".")
			git("-c", "user.name=Evidence fixture", "-c", "user.email=evidence@example.invalid", "commit", "--quiet", "--allow-empty", "-m", "fixed catalog")
			commit := git("rev-parse", "HEAD")
			if mode == "missing-commit" {
				commit = strings.Repeat("a", 40)
			} else if mode == "invalid-commit" {
				commit = "HEAD"
			}
			bundle := filepath.Join(repo, "bundle")
			if err := os.Mkdir(bundle, 0700); err != nil {
				t.Fatal(err)
			}
			writeReleaseJSON(t, bundle, "live-report.json", map[string]any{"gitCommit": commit})
			if mode == "dirty-worktree" {
				mustWriteXAINative(t, filepath.Join(repo, "ai/release/catalog-snapshot.json"), []byte("{}"))
			} else if mode == "wrong-hash" {
				hash = "sha256:" + strings.Repeat("a", 64)
			} else if mode == "wrong-version" {
				catalog.Version = "substituted-version"
			}
			measured, err := evidence.MeasuredCatalog(bundle, catalog.Version, hash, ai.BuiltinCatalog())
			want := mode == "valid" || mode == "dirty-worktree" || mode == "model-drift"
			ev.Check("only fixed matching original catalog is recovered", (err == nil) == want, "%v", err)
			if want && err == nil {
				current := ai.BuiltinCatalog()
				if mode == "model-drift" {
					current.ImageModels[0].Capabilities.MaxOutputImages++
				}
				model := catalog.ImageModels[0]
				err = evidence.VerifyRouteCatalog(measured, current, ai.OperationImage, model.Provider, model.API, []string{model.ID})
				ev.Check("recovery does not waive current model comparison", (err == nil) == (mode != "model-drift"), "%v", err)
			}
			ev.Record("recovery", map[string]any{"mode": mode, "commit": commit, "error": errString(err), "vendorCalls": 0})
		})
	}
}

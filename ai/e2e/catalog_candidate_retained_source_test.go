package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogCandidateCheckRetainedSource(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-recompile-retained-embedded-source")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	candidate := freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "prove complete compiler inputs are retained")
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	json.Unmarshal(out, &checked)
	if !ev.Check("actual check completes", exit == 0 && checked.Status == "applicable", "exit=%d output=%s", exit, out) {
		return
	}
	var prepared struct {
		Extra map[string]string `json:"extra"`
	}
	raw, err := os.ReadFile(filepath.Join(root, checked.Prepared))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &prepared); err != nil {
		t.Fatal(err)
	}
	marker := "catalog/data/baseline/README.md"
	actual, err := os.ReadFile(filepath.Join(root, marker))
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(actual)
	ev.Check("prepared inventory binds the compiler-required baseline marker", prepared.Extra[marker] == hex.EncodeToString(sum[:]), "extra digest=%s", prepared.Extra[marker])
	retainedRoot := filepath.Join(root, filepath.Dir(checked.Prepared), "source")
	retained, readErr := os.ReadFile(filepath.Join(retainedRoot, marker))
	ev.Check("source archive retains exact embedded marker bytes", readErr == nil && bytes.Equal(retained, actual), "err=%v", readErr)
	replay := filepath.Join(t.TempDir(), "catalogctl-from-retained-source")
	cmd := exec.CommandContext(ctxFor(t), "go", catalogBuildArgs(replay, "./cmd/catalogctl")...)
	cmd.Dir = retainedRoot
	build, buildErr := cmd.CombinedOutput()
	ev.Record("retained-source-build", map[string]any{"argv": cmd.Args, "directory": retainedRoot, "exitCode": commandExit(buildErr), "output": string(build)})
	if ev.Check("retained actual source recompiles the public maintenance CLI", buildErr == nil, "err=%v output=%s", buildErr, build) {
		raw, err := os.ReadFile(replay)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(raw)
		ev.Record("retained-source-binary", map[string]any{"sha256": hex.EncodeToString(sum[:]), "bytes": len(raw)})
	}
}

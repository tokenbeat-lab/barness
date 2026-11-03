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
	candidate := freezeCatalogAction(t, ev, binary, store, "withdraw", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "prove complete compiler inputs are retained")
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
	baseline := filepath.Join(root, "catalog", "data", "baseline")
	retainedRoot := filepath.Join(root, filepath.Dir(checked.Prepared), "source")
	count := 0
	err = filepath.WalkDir(baseline, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		actual, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(actual)
		retained, readErr := os.ReadFile(filepath.Join(retainedRoot, filepath.FromSlash(name)))
		ev.Check("prepared inventory and archive retain embedded baseline bytes", prepared.Extra[name] == hex.EncodeToString(sum[:]) && readErr == nil && bytes.Equal(retained, actual), "path=%s err=%v", name, readErr)
		count++
		return nil
	})
	ev.Check("complete adopted baseline inventory was inspected", err == nil && count > 0, "count=%d err=%v", count, err)
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

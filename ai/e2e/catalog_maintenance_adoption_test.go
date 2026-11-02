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

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func copyAdoptedCatalogReview(t *testing.T, source, target string) {
	t.Helper()
	// An adopted review replaces the complete inventory; merging leaves stale blobs.
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, _ := filepath.Rel(source, path)
		dest := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0700)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, raw, 0600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func verifyAdoptedCatalogCycle(t *testing.T, ev *evidence.Case, binary, fresh, dir string, capture, manual []byte, oldReferences []maintenance.ReviewReference) {
	t.Helper()
	indexBytes, _ := os.ReadFile(filepath.Join(fresh, "catalog", "data", "baseline", "review.json"))
	indexSum := sha256.Sum256(indexBytes)
	origin := "sha256:" + hex.EncodeToString(indexSum[:])
	store := filepath.Join(dir, "fresh-maintenance-store")
	overlay := filepath.Join(dir, "next-overlay.json")
	os.WriteFile(overlay, []byte(`{"name":"Review model second cycle"}`), 0600)
	candidate := freezeCatalogAction(t, ev, binary, store, "revise", "--operation", "chat", "--provider", "openai", "--api", "openai-completions", "--model", "review-model", "--definition", overlay, "--reason", "maintain adopted default without refetching sources")
	checkApplyCatalogAction(t, ev, binary, store, candidate)
	second := filepath.Join(dir, "second-review")
	out, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--repo", catalogSnapshotRoot(t), "--review-dir", second, "--confirm")
	if ev.Check("adopted default supports a second maintenance and review export", exit == 0, "exit=%d output=%s", exit, out) {
		var result struct {
			Inputs     []maintenance.ReviewInput                            `json:"inputs"`
			References []struct{ Scope, Path, Digest, OriginReview string } `json:"references"`
		}
		if err := json.Unmarshal(out, &result); err != nil {
			t.Fatal(err)
		}
		sourceFound, manualFound := false, false
		for _, input := range result.Inputs {
			raw, err := os.ReadFile(filepath.Join(second, input.Path))
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(raw, capture) {
				sourceFound = input.Kind == "official" && input.Role == "source" && input.URL == "https://official.example.test/review" && input.FetchedAt == "2026-11-01T00:00:00Z"
			}
			manualFound = manualFound || bytes.Equal(raw, manual)
		}
		ev.Check("second cycle preserves exact official provenance and manual bytes", sourceFound && manualFound, "inputs=%+v", result.Inputs)
		for _, previous := range oldReferences {
			expectedOrigin := previous.OriginReview
			if expectedOrigin == "" {
				expectedOrigin = origin
			}
			found := false
			for _, current := range result.References {
				found = found || (current.Scope == previous.Scope && current.Path == previous.Path && current.Digest == previous.Digest && current.OriginReview == expectedOrigin)
			}
			ev.Check("inherited reference records origin without claiming original availability", found, "reference=%+v", previous)
		}
		ev.Fixture("second-review.json", out)
	}
	baseline := filepath.Join(fresh, "catalog", "data", "baseline")
	indexPath := filepath.Join(baseline, "review.json")
	index, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest maintenance.ReviewExport
	if err := json.Unmarshal(index, &manifest); err != nil {
		t.Fatal(err)
	}
	blob := filepath.Join(baseline, manifest.Inputs[0].Path)
	original, err := os.ReadFile(blob)
	if err != nil {
		t.Fatal(err)
	}
	providerPath := filepath.Join(fresh, "catalog", "data", "providers", "openai.json")
	providerBytes, _ := os.ReadFile(providerPath)
	runtimePath := filepath.Join(fresh, "ai", "catalog_builtin.json")
	runtimeBytes, _ := os.ReadFile(runtimePath)
	originalRuntime, _ := os.ReadFile(filepath.Join(catalogSnapshotRoot(t), "ai", "catalog_builtin.json"))
	for _, corruption := range []string{"blob", "missing-blob", "index", "extra", "metadata", "provider", "runtime"} {
		switch corruption {
		case "blob":
			os.WriteFile(blob, []byte("changed original"), 0600)
		case "missing-blob":
			os.Remove(blob)
		case "index":
			os.Remove(indexPath)
		case "metadata":
			changed := manifest
			changed.Inputs = append([]maintenance.ReviewInput(nil), manifest.Inputs...)
			for i := range changed.Inputs {
				if changed.Inputs[i].Role == "source" {
					changed.Inputs[i].URL = "https://different.example.test/source"
				}
			}
			raw, _ := json.Marshal(changed)
			os.WriteFile(indexPath, raw, 0600)
		case "provider":
			os.WriteFile(providerPath, bytes.Replace(providerBytes, []byte("Review model"), []byte("Wrong source"), 1), 0600)
		case "runtime":
			os.WriteFile(runtimePath, originalRuntime, 0600)
		case "extra":
			os.WriteFile(filepath.Join(baseline, "unexpected.json"), []byte(`{}`), 0600)
		}
		corrupted := filepath.Join(dir, "corrupted-"+corruption)
		cmd := exec.CommandContext(ctxFor(t), "go", catalogBuildArgs(corrupted, "./cmd/catalogctl")...)
		cmd.Dir = fresh
		build, err := cmd.CombinedOutput()
		ev.Record("corrupted-build-"+corruption, map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(build)})
		if ev.Check("corrupted baseline binary builds for runtime refusal", err == nil, "kind=%s err=%v output=%s", corruption, err, build) {
			binaryBytes, _ := os.ReadFile(corrupted)
			binarySum := sha256.Sum256(binaryBytes)
			ev.Record("corrupted-binary-"+corruption, map[string]any{"sha256": hex.EncodeToString(binarySum[:]), "bytes": len(binaryBytes)})
			absent := filepath.Join(dir, "corrupt-store-"+corruption)
			out, exit := invokeCatalogCLI(t, ev, corrupted, "versions", "--store", absent)
			ev.Check("corrupt or missing adopted baseline refuses bootstrap", exit != 0, "kind=%s exit=%d output=%s", corruption, exit, out)
			_, err := os.Stat(absent)
			ev.Check("baseline refusal is read-only", os.IsNotExist(err), "kind=%s err=%v", corruption, err)
		}
		os.WriteFile(providerPath, providerBytes, 0600)
		os.WriteFile(runtimePath, runtimeBytes, 0600)
		os.WriteFile(blob, original, 0600)
		os.WriteFile(indexPath, index, 0600)
		os.Remove(filepath.Join(baseline, "unexpected.json"))
	}
}

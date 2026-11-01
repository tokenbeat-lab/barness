package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCatalogMaintenanceCLIReviewExport(t *testing.T) {
	ev := run.Case(t, "CATALOG-T01-cli-review-source-bootstrap")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	source := filepath.Join(dir, "official.txt")
	definition := filepath.Join(dir, "definition.json")
	capture := []byte("Synthetic official captured source, distinct from models.dev, for review provenance.\n")
	manual := []byte(`{"provider":"openai","api":"openai-completions","id":"review-model","name":"Review model","reasoning":false,"input":["text"],"contextWindow":4096,"maxTokens":128,"cost":{"input":1,"output":2,"cacheRead":0,"cacheWrite":0}}`)
	os.WriteFile(source, capture, 0600)
	os.WriteFile(definition, manual, 0600)
	candidate := freezeCatalogAction(t, ev, binary, store, "add", "--operation", "chat", "--provider", "openai", "--api", "openai-completions", "--model", "review-model", "--source-kind", "official", "--source", source, "--source-url", "https://official.example.test/review", "--source-time", "2026-11-01T00:00:00Z", "--definition", definition, "--reason", "export coherent approved review source")
	revision := checkApplyCatalogAction(t, ev, binary, store, candidate)
	review := filepath.Join(dir, "new-review")
	out, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--review-dir", review, "--repo", catalogSnapshotRoot(t), "--dry-run")
	_, err := os.Stat(review)
	ev.Check("review preview creates no directory", exit == 0 && os.IsNotExist(err), "exit=%d err=%v output=%s", exit, err, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--review-dir", review, "--repo", catalogSnapshotRoot(t), "--confirm")
	if !ev.Check("approved review bundle exported to new directory", exit == 0, "exit=%d output=%s", exit, out) {
		return
	}
	state, err := os.ReadFile(filepath.Join(review, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("export-state.json", state)
	snapshot, err := os.ReadFile(filepath.Join(review, "snapshot.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("review uses exact checked runtime bytes", bytes.Equal(snapshot, revision.Snapshot), "snapshot changed")
	index, err := os.ReadFile(filepath.Join(review, "review.json"))
	if err != nil {
		t.Fatal(err)
	}
	ev.Fixture("review.json", index)
	var manifest struct {
		Files      map[string]string                      `json:"files"`
		Revision   string                                 `json:"revision"`
		Inputs     []struct{ Kind, Path string }          `json:"inputs"`
		References []struct{ Scope, Path, Digest string } `json:"references"`
	}
	_ = json.Unmarshal(index, &manifest)
	retainedReview := []map[string]string{}
	for name, expected := range manifest.Files {
		raw, err := os.ReadFile(filepath.Join(review, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		artifact := "review-file-" + strings.ReplaceAll(name, "/", "--")
		ev.Fixture(artifact, raw)
		retainedReview = append(retainedReview, map[string]string{"path": name, "artifact": artifact, "digest": expected})
		sum := sha256.Sum256(raw)
		ev.Check("review inventory binds actual retained file bytes", "sha256:"+hex.EncodeToString(sum[:]) == expected, "file=%s", name)
	}
	ev.Record("review-retention", retainedReview)
	foundSource, foundLocal, foundHistorical := false, false, false
	for _, input := range manifest.Inputs {
		raw, err := os.ReadFile(filepath.Join(review, input.Path))
		if err != nil {
			t.Fatal(err)
		}
		foundSource = foundSource || bytes.Equal(raw, capture)
		foundLocal = foundLocal || bytes.Equal(raw, manual)
		foundHistorical = foundHistorical || input.Kind == "historical"
	}
	ev.Check("captured source and manual originals remain self-contained", foundSource && foundLocal && foundHistorical && manifest.Revision == revision.ID, "inputs=%+v", manifest.Inputs)
	retainedReferences := []map[string]string{}
	for index, reference := range manifest.References {
		base := catalogSnapshotRoot(t)
		if reference.Scope == "store" {
			base = store
		}
		raw, err := os.ReadFile(filepath.Join(base, reference.Path))
		if err == nil {
			artifact := fmt.Sprintf("reference-%03d.json", index)
			ev.Fixture(artifact, raw)
			retainedReferences = append(retainedReferences, map[string]string{"scope": reference.Scope, "originalPath": reference.Path, "artifact": artifact, "digest": reference.Digest})
		}
		ev.Check("exact external evidence/store reference remains available", err == nil && len(raw) > 0 && func() bool {
			sum := sha256.Sum256(raw)
			return "sha256:"+hex.EncodeToString(sum[:]) == reference.Digest
		}(), "reference=%+v err=%v", reference, err)
	}
	ev.Record("reference-retention", retainedReferences)
	fresh := copyCatalogBootstrapSource(t)
	providers, err := os.ReadDir(filepath.Join(review, "providers"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range providers {
		raw, err := os.ReadFile(filepath.Join(review, "providers", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fresh, "catalog", "data", "providers", entry.Name()), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(fresh, "ai", "catalog_builtin.json"), snapshot, 0600)
	copyAdoptedCatalogReview(t, review, filepath.Join(fresh, "catalog", "data", "baseline"))
	freshCLI := filepath.Join(dir, "fresh-catalogctl")
	cmd := exec.CommandContext(ctxFor(t), "go", "build", "-o", freshCLI, "./cmd/catalogctl")
	cmd.Dir = fresh
	build, err := cmd.CombinedOutput()
	ev.Record("fresh-build", map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(build)})
	if !ev.Check("fresh checkout with reviewed Provider source builds", err == nil, "%v %s", err, build) {
		return
	}
	freshBinary, _ := os.ReadFile(freshCLI)
	freshSum := sha256.Sum256(freshBinary)
	ev.Record("fresh-binary", map[string]any{"sha256": hex.EncodeToString(freshSum[:]), "bytes": len(freshBinary)})
	out, exit = invokeCatalogCLI(t, ev, freshCLI, "export", "--store", filepath.Join(dir, "absent-fresh-store"))
	ev.Check("absent-store bootstrap matches newly adopted default", exit == 0 && bytes.Equal(out, snapshot), "exit=%d size=%d", exit, len(out))
	probeDir := filepath.Join(fresh, "cmd", "observebuiltin")
	os.MkdirAll(probeDir, 0700)
	probe := []byte("package main\nimport(\"encoding/json\";\"os\";\"github.com/tokenbeat-lab/barness/ai\")\nfunc main(){ c:=ai.BuiltinCatalog();h,e:=c.Hash();if e!=nil{panic(e)};json.NewEncoder(os.Stdout).Encode(struct{Hash string `json:\"hash\"`;Catalog ai.Catalog `json:\"catalog\"`}{h,c}) }\n")
	os.WriteFile(filepath.Join(probeDir, "main.go"), probe, 0600)
	ev.Fixture("fresh-builtin-consumer.go", probe)
	observer := filepath.Join(dir, "observebuiltin")
	cmd = exec.CommandContext(ctxFor(t), "go", "build", "-o", observer, "./cmd/observebuiltin")
	cmd.Dir = fresh
	build, err = cmd.CombinedOutput()
	ev.Record("fresh-builtin-build", map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(build)})
	if !ev.Check("fresh public builtin consumer builds", err == nil, "err=%v output=%s", err, build) {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, observer)
	ev.Check("fresh public BuiltinCatalog matches approved source and runtime artifact", exit == 0 && bytes.Equal(out, snapshot), "exit=%d size=%d", exit, len(out))
	verifyAdoptedCatalogCycle(t, ev, freshCLI, fresh, dir, capture, manual, manifest.References)
	before, _ := os.ReadFile(filepath.Join(review, "review.json"))
	out, exit = invokeCatalogCLI(t, ev, binary, "export", "--store", store, "--review-dir", review, "--repo", catalogSnapshotRoot(t), "--confirm")
	after, _ := os.ReadFile(filepath.Join(review, "review.json"))
	ev.Check("existing review bundle is never overwritten", exit != 0 && bytes.Equal(before, after), "exit=%d output=%s", exit, out)
}

func copyCatalogBootstrapSource(t *testing.T) string {
	t.Helper()
	root := catalogSnapshotRoot(t)
	destination := t.TempDir()
	for _, scope := range []string{"ai", "ai/internal/clock", "ai/internal/probe", "catalog", "cmd/catalogctl", "cmd/cataloggen"} {
		base := filepath.Join(root, scope)
		err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if scope == "ai" && path != base {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, "_test.go") || !(strings.HasSuffix(path, ".go") || strings.HasSuffix(path, ".json")) {
				return nil
			}
			relative, _ := filepath.Rel(root, path)
			target := filepath.Join(destination, relative)
			if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
				return err
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
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(destination, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return destination
}

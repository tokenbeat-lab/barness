package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// The public consumer must reject malformed directories and artifact trees,
// while retaining the producer's recursive ownership. No provider is invoked.
func TestReleaseBundleDirectoryIsolation(t *testing.T) {
	modes := []string{"flat", "nested", "nested-verdict", "alias", "duplicate-id", "traversal", "absolute", "dot", "internal-dot", "repeated", "double-trailing", "backslash", "case-symlink", "ancestor-symlink", "nested-symlink", "file-symlink", "missing", "extra", "omitted", "hash", "artifact-traversal", "artifact-alias", "artifact-absolute", "empty-extra-dir"}
	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-directory-isolation-"+mode)
			dir, id := t.TempDir(), "case"
			files := map[string]string{"result.json": `{"parent":true}`}
			if strings.HasPrefix(mode, "nested") || strings.HasPrefix(mode, "artifact-") || mode == "missing" || mode == "extra" || mode == "omitted" || mode == "hash" {
				files["high/result.json"] = `{"child":true}`
				files["high/pidiff.json"] = `{"notTheParentVerdict":true}`
			}
			if mode == "ancestor-symlink" {
				id = "branch/case"
			}
			artifacts := releaseBoundaryFiles(t, filepath.Join(dir, id), files)
			c := map[string]any{"id": "parent", "dir": id, "status": "PASS", "artifacts": artifacts}
			cases := []any{c}
			switch mode {
			case "nested-verdict":
				c["id"] = "PIDIFF-P01-parent"
			case "alias", "duplicate-id":
				other := "other"
				if mode == "duplicate-id" {
					other = "parent"
				}
				cases = append(cases, map[string]any{"id": other, "dir": id + "/", "status": "PASS", "artifacts": artifacts})
			case "traversal", "absolute", "dot", "internal-dot", "repeated", "double-trailing", "backslash":
				c["dir"] = map[string]string{"traversal": "case/../case", "absolute": "/case", "dot": "./case", "internal-dot": "case/./high", "repeated": "case//high", "double-trailing": "case//", "backslash": `case\high`}[mode]
			case "case-symlink", "ancestor-symlink", "nested-symlink", "file-symlink":
				path := filepath.Join(dir, id)
				switch mode {
				case "ancestor-symlink":
					path = filepath.Join(dir, "branch")
				case "nested-symlink":
					path = filepath.Join(path, "high")
				case "file-symlink":
					path = filepath.Join(path, "result.json")
				}
				moved := filepath.Join(t.TempDir(), "original")
				if err := os.Rename(path, moved); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(moved, path); err != nil {
					t.Fatal(err)
				}
			case "missing":
				if err := os.Remove(filepath.Join(dir, id, "high/result.json")); err != nil {
					t.Fatal(err)
				}
			case "extra":
				releaseBoundaryFiles(t, filepath.Join(dir, id), map[string]string{"high/extra.json": "{}"})
			case "omitted":
				delete(artifacts, "high/result.json")
			case "hash":
				artifacts["high/result.json"] = strings.Repeat("0", 64)
			case "artifact-traversal", "artifact-alias", "artifact-absolute":
				key := map[string]string{"artifact-traversal": "high/../result.json", "artifact-alias": "high//result.json", "artifact-absolute": "/result.json"}[mode]
				artifacts[key] = artifacts["result.json"]
			case "empty-extra-dir":
				if err := os.MkdirAll(filepath.Join(dir, id, "extra"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			manifest := map[string]any{"package": release.OfflinePackage, "cases": cases}
			writeReleaseJSON(t, dir, "manifest.json", manifest)
			ev.Record("manifest", manifest)
			b, err := release.LoadBundle(dir)
			accepted := err == nil && len(b.Problems) == 0
			ev.Record("import", map[string]any{"accepted": accepted, "error": err != nil, "problems": b.Problems, "cases": b.Cases, "diffs": b.Diffs})
			ev.Check("only intact canonical trees import", accepted == (mode == "flat" || strings.HasPrefix(mode, "nested") && mode != "nested-symlink"), "accepted=%t error=%v problems=%v", accepted, err, b.Problems)
			if accepted {
				ev.Check("nested artifacts do not invent differential cases", len(b.Cases) == 1 && len(b.Diffs) == 0 && b.Cases[0].ID == c["id"], "unexpected case attribution")
			}
		})
	}
}

func releaseBoundaryFiles(t *testing.T, dir string, files map[string]string) map[string]string {
	t.Helper()
	artifacts := map[string]string{}
	for name, raw := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		artifacts[name] = pioracle.SHA256([]byte(raw))
	}
	return artifacts
}

func TestReleaseBundleInventoryFiniteBounds(t *testing.T) {
	for _, mode := range []string{"files", "entries", "depth", "file-bytes", "total-bytes"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-inventory-bound-"+mode)
			dir := t.TempDir()
			files := map[string]string{"result.json": "{}"}
			switch mode {
			case "files":
				for i := 0; i < 512; i++ {
					files[fmt.Sprintf("extra-%03d.json", i)] = "{}"
				}
			case "entries":
				files = map[string]string{}
				for i := 0; i < 512; i++ {
					files[fmt.Sprintf("part-%03d/a/b/c/result.json", i)] = "{}"
				}
			case "depth":
				files[strings.Repeat("child/", 24)+"result.json"] = "{}"
			}
			artifacts := releaseBoundaryFiles(t, filepath.Join(dir, "case"), files)
			if mode == "file-bytes" || mode == "total-bytes" {
				count, size := 1, int64(64<<20)+1
				if mode == "total-bytes" {
					count, size = 5, 64<<20
				}
				for i := 0; i < count; i++ {
					name := fmt.Sprintf("bounded-%d.bin", i)
					artifacts[name] = releaseSparseArtifact(t, filepath.Join(dir, "case", name), size)
				}
			}
			manifest := map[string]any{"package": release.OfflinePackage, "cases": []any{map[string]any{"id": "case", "dir": "case", "status": "PASS", "artifacts": artifacts}}}
			writeReleaseJSON(t, dir, "manifest.json", manifest)
			ev.Record("manifest", manifest)
			b, err := release.LoadBundle(dir)
			ev.Record("import", map[string]any{"error": err != nil, "problems": b.Problems})
			ev.Check("finite inventory ceiling enforced", err != nil || len(b.Problems) != 0, "over-bound tree accepted")
		})
	}
}

func releaseSparseArtifact(t *testing.T, path string, size int64) string {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := f.Truncate(size); err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	if n, err := io.Copy(h, io.LimitReader(f, size)); err != nil || n != size {
		t.Fatalf("controlled sparse artifact failed: n=%d err=%v", n, err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

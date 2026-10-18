package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestReleaseBundlePostLoadIsolation(t *testing.T) {
	for _, mode := range []string{"unchanged", "case-symlink", "ancestor-symlink", "bundle-symlink", "file-symlink", "file-tamper"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD58-release-post-load-"+mode)
			dir := releaseRollbackBundle(t)
			id := "P10-OD58-runtime-disabled-binding-initial-images"
			physical := id
			if mode == "ancestor-symlink" {
				physical = "ancestor/" + id
				if err := os.Mkdir(filepath.Join(dir, "ancestor"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(filepath.Join(dir, id), filepath.Join(dir, physical)); err != nil {
					t.Fatal(err)
				}
				raw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
				var m map[string]any
				if err := json.Unmarshal(raw, &m); err != nil {
					t.Fatal(err)
				}
				for _, entry := range m["cases"].([]any) {
					c := entry.(map[string]any)
					if c["id"] == id {
						c["dir"] = physical
					}
				}
				writeReleaseJSON(t, dir, "manifest.json", m)
			}
			b, err := release.LoadBundle(dir)
			if !ev.Check("intact public import", err == nil && len(b.Problems) == 0, "error=%t problems=%d", err != nil, len(b.Problems)) {
				return
			}
			before := release.RevocationEvidence(b)
			if !ev.Check("genuine intact revocation PASS", before.Status == release.Pass, "%s", before.Detail) {
				return
			}
			path := filepath.Join(dir, physical)
			switch mode {
			case "bundle-symlink":
				path = dir
			case "ancestor-symlink":
				path = filepath.Join(dir, "ancestor")
			case "file-symlink", "file-tamper":
				path = filepath.Join(path, "resources.json")
			}
			if mode == "file-tamper" {
				f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
				if err != nil {
					t.Fatal(err)
				}
				_, err = f.WriteString("\n")
				f.Close()
				if err != nil {
					t.Fatal(err)
				}
			} else if mode != "unchanged" {
				outside := filepath.Join(t.TempDir(), "original")
				if err := os.Rename(path, outside); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			}
			after := release.RevocationEvidence(b)
			ev.Record("consumption", map[string]any{"before": before.Status, "after": after.Status, "detail": after.Detail})
			ev.Check("only unchanged authenticated bytes qualify", (after.Status == release.Pass) == (mode == "unchanged"), "modified input qualified: %s", after.Status)
		})
	}
}

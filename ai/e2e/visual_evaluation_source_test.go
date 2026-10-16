package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVisualEvaluationSourceSet(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "VQ12-source-set-"+route)
			out := filepath.Join(run.Dir(), "visual-source-set-"+route)
			if _, err := visualCommand(t, "-route", route, "-out", out); err != nil {
				t.Fatal("controlled workflow failed")
			}
			if _, err := visualCommand(t, "-verify", out); err != nil {
				t.Fatal("original source snapshot invalid")
			}
			hashes := visualRead(t, out, "source-hashes.json")
			original := "ai/examples/visualeval/verify.go"
			substitute := "unknown-empty.go"
			delete(hashes, original)
			hashes[substitute] = visualHash(nil)
			visualWrite(t, filepath.Join(out, "source-hashes.json"), hashes)
			if err := os.Remove(filepath.Join(out, "source", original)); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(out, "source", substitute), nil, 0600); err != nil {
				t.Fatal(err)
			}
			encoded, _ := json.MarshalIndent(hashes, "", "  ")
			encoded = append(encoded, '\n')
			report := visualRead(t, out, "report.json")
			report["codeHash"] = visualHash(encoded)
			visualPatch(t, filepath.Join(out, "report.json"), report)
			manifest := visualRead(t, out, "manifest.json")
			files := manifest["artifacts"].(map[string]any)
			delete(files, "source/"+original)
			files["source/"+substitute] = visualHash(nil)
			visualWrite(t, filepath.Join(out, "manifest.json"), manifest)
			visualReseal(t, out)
			_, err := visualCommand(t, "-verify", out)
			ev.Check("same-length source map cannot substitute unknown empty source", err != nil, "accepted missing runtime source")
		})
	}
}

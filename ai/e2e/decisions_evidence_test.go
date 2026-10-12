package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecisionsEvidence(t *testing.T) {
	base := filepath.Join("..", "..", ".scratch", "barness-ai-openai-decisions", "text-choice-evidence")
	for _, name := range []string{"intact", "missing-archive", "changed-archive", "rewritten-manifest"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD32-evidence-"+name)
			if _, err := os.Stat(filepath.Join(base, "verify.py")); err != nil {
				ev.Check("verifier available", false, "text evidence verifier absent")
				return
			}
			dir := base
			if name != "intact" {
				dir = t.TempDir()
				for _, file := range []string{"manifest.json", "text-choice.zip"} {
					raw, err := os.ReadFile(filepath.Join(base, file))
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(dir, file), raw, 0o600); err != nil {
						t.Fatal(err)
					}
				}
				switch name {
				case "missing-archive":
					if err := os.Remove(filepath.Join(dir, "text-choice.zip")); err != nil {
						t.Fatal(err)
					}
				case "changed-archive":
					if err := os.WriteFile(filepath.Join(dir, "text-choice.zip"), []byte("changed evidence"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "rewritten-manifest":
					if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"status":"PASS"}`), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			out, err := exec.CommandContext(ctxFor(t), "python3", filepath.Join(base, "verify.py"), "--bundle", dir).CombinedOutput()
			ev.Record("verification", map[string]any{"output": string(out), "succeeded": err == nil})
			if name == "intact" {
				ev.Check("recorded evidence intact", err == nil && strings.Contains(string(out), "no live execution"), "got %s", out)
			} else {
				ev.Check("missing or rewritten evidence rejected", err != nil && strings.Contains(string(out), "FAIL:"), "got %s", out)
			}
		})
	}
}

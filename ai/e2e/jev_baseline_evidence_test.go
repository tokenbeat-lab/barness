package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The agreed artifact seam must reject missing captures, changed bytes and a
// rewritten inventory passed off as the original. Verification success means
// recorded evidence is intact; it cannot turn historical release FAIL into PASS.
func TestJevBaselineEvidence(t *testing.T) {
	base := filepath.Join("testdata", "retained", "jev-baseline")
	verifier := filepath.Join("..", "..", ".scratch", "barness-ai-openai-decisions", "jev-baseline", "verify.py")
	for _, name := range []string{"intact", "missing-archive", "changed-archive", "rewritten-manifest"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-OD50-jev-baseline-"+name)
			if _, err := os.Stat(verifier); err != nil {
				ev.Check("verifier available", false, "baseline verifier is absent")
				return
			}
			dir := base
			if name != "intact" {
				dir = t.TempDir()
				for _, file := range []string{"manifest.json", "baseline.zip"} {
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
					if err := os.Remove(filepath.Join(dir, "baseline.zip")); err != nil {
						t.Fatal(err)
					}
				case "changed-archive":
					if err := os.WriteFile(filepath.Join(dir, "baseline.zip"), []byte("rewritten captures"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "rewritten-manifest":
					if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"status":"PASS"}`), 0o600); err != nil {
						t.Fatal(err)
					}
				}
			}
			cmd := exec.CommandContext(ctxFor(t), "python3", verifier, "--bundle", dir)
			out, err := cmd.CombinedOutput()
			root, rootErr := filepath.Abs(filepath.Join("..", ".."))
			if rootErr != nil {
				t.Fatal(rootErr)
			}
			out = []byte(strings.ReplaceAll(string(out), root, "$REPO"))
			ev.Record("verification", map[string]any{"output": string(out), "succeeded": err == nil})
			if name == "intact" {
				ev.Check("fixed evidence verifies and preserves release failure", err == nil && strings.Contains(string(out), "recorded release FAIL"), "got %s", out)
			} else {
				ev.Check("rewritten baseline refused", err != nil && strings.Contains(string(out), "FAIL:"), "got %s", out)
			}
		})
	}
}

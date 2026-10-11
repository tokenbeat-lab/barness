package e2e

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// B01 keeps the two actually executed source-equivalence proofs in the release
// evidence. Hashes protect their full inventories, commands and fault checks;
// this verification is offline and never substitutes for a new rebuild.
func TestThinkingSamplingBaseline(t *testing.T) {
	base := filepath.Join("..", "..", ".scratch", "barness-ai-thinking-sampling", "release-verification")
	for _, proof := range []string{"evidence-b", "evidence-c"} {
		t.Run(proof, func(t *testing.T) {
			ev := run.Case(t, "E03-thinking-sampling-baseline-"+proof)
			dir := filepath.Join(base, proof)
			sums, err := os.ReadFile(filepath.Join(dir, "SHA256SUMS"))
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("proof.sha256", sums)
			for _, line := range strings.Split(strings.TrimSpace(string(sums)), "\n") {
				parts := strings.Fields(line)
				if len(parts) != 2 || filepath.Base(parts[1]) != parts[1] {
					t.Fatal("invalid proof hash entry")
				}
				raw, err := os.ReadFile(filepath.Join(dir, parts[1]))
				ev.Check("proof file hash: "+parts[1], err == nil && fmt.Sprintf("%x", sha256.Sum256(raw)) == parts[0], "proof file differs")
				if strings.HasSuffix(parts[1], ".json") || strings.HasSuffix(parts[1], ".sha256") {
					ev.Fixture(parts[1], raw)
				}
			}
			raw, err := os.ReadFile(filepath.Join(dir, "report.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report struct {
				Status string `json:"status"`
				Source struct {
					Target string `json:"target_commit"`
					Tag    string `json:"tag_commit"`
				} `json:"source"`
				Data struct {
					Hash string `json:"sha256"`
				} `json:"model_data"`
				Comparisons []struct {
					Files       int `json:"compared_files"`
					Differences int `json:"differences"`
				} `json:"comparisons"`
			}
			if err := json.Unmarshal(raw, &report); err != nil {
				t.Fatal(err)
			}
			ev.Check("fixed source and released data verified", report.Status == "PASS" && report.Source.Target == "abe508e1b89912adde45528136c3221eb69acdd7" && report.Source.Tag == report.Source.Target && report.Data.Hash == "8a06cf635be3392b977f7bb328e7ba5065cbe55825d997ca7d00309fedcdced8", "identity differs")
			ev.Check("complete dist equivalence", len(report.Comparisons) == 2 && report.Comparisons[0].Files == 827 && report.Comparisons[1].Files == 24 && report.Comparisons[0].Differences == 0 && report.Comparisons[1].Differences == 0, "incomplete proof")
		})
	}
	t.Run("oracle-guards", func(t *testing.T) {
		ev := run.Case(t, "E03-thinking-sampling-baseline-oracle-guards")
		ev.ReplayEnv(pioracle.EnableEnv + "=1")
		if !pioracle.Enabled() {
			t.Skip("oracle guard execution requires the installed frozen oracle")
		}
		cmd := exec.CommandContext(ctxFor(t), "node", "--test", "../internal/testkit/pioracle/node/runner.test.mjs")
		cmd.Env = []string{}
		output, err := cmd.CombinedOutput()
		ev.Record("runner-guard-results", string(output))
		ev.Check("version/commit/data mismatches rejected", err == nil, "guard command failed: %v", err)
	})
}

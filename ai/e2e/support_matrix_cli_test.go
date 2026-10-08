package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The existing command is the public file boundary. The prewritten failure
// list includes a rejected later report leaving an earlier partial write.
func TestSupportMatrixCLIAtomic(t *testing.T) {
	for _, reject := range []bool{false, true} {
		name := "own-row"
		if reject {
			name = "atomic-refusal"
		}
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-E09-matrix-cli-"+name)
			dir := t.TempDir()
			matrixPath := filepath.Join(dir, "matrix.json")
			m := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{
				{Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses"},
				{Combo: "deepseek-responses", Operation: "chat", Provider: "deepseek", API: "openai-responses"},
			}}
			before, _ := supportmatrix.Encode(m)
			if err := os.WriteFile(matrixPath, before, 0600); err != nil {
				t.Fatal(err)
			}
			at := time.Date(2026, 10, 8, 3, 0, 0, 0, time.UTC)
			r := supportmatrix.Report{Schema: supportmatrix.SchemaVersion, Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses", Model: "synthetic", SDK: "synthetic", AccountAlias: "test@unknown", FinishedAt: at, Expected: []string{"text-stream"}, Scenarios: []supportmatrix.ScenarioResult{{ID: "text-stream", Model: "synthetic", Outcome: supportmatrix.Pass}}}
			good := filepath.Join(dir, "good.json")
			data, _ := json.Marshal(r)
			os.WriteFile(good, data, 0600)
			args := []string{"run", "./ai/live/cmd/supportmatrix", "-matrix", matrixPath, good}
			if reject {
				r.Schema = 1
				r.FinishedAt = at.Add(time.Second)
				bad := filepath.Join(dir, "bad.json")
				data, _ = json.Marshal(r)
				os.WriteFile(bad, data, 0600)
				args = append(args, bad)
			}
			cmd := exec.CommandContext(ctxFor(t), "go", args...)
			cmd.Dir = "../.."
			output, err := cmd.CombinedOutput()
			ev.Record("command", map[string]any{"output": string(output), "refused": err != nil})
			after, readErr := os.ReadFile(matrixPath)
			if readErr != nil {
				t.Fatal(readErr)
			}
			ev.Record("matrix-before", m)
			if reject {
				ev.Check("batch refuses without partial write", err != nil && bytes.Equal(before, after), "exit=%v unchanged=%v", err, bytes.Equal(before, after))
			} else {
				var got supportmatrix.Matrix
				mustUnmarshal(t, after, &got)
				ev.Record("matrix-after", got)
				ev.Check("only report's route updated", err == nil && got.Rows[0].AllPassedAt != nil && got.Rows[1].AllPassedAt == nil && got.Rows[1].LastRunAt == nil, "exit=%v", err)
			}
		})
	}
}

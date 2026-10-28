package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func verifyXAINativeMatrixFixtures(t *testing.T, f xaiFormalNativeFixture) {
	t.Helper()
	for _, complete := range []bool{true, false} {
		name := "complete-synthetic"
		if !complete {
			name = "generic-only"
		}
		t.Run("matrix-"+name, func(t *testing.T) {
			ev := run.Case(t, "XAI-E17-formal-matrix-"+name)
			dir := f.copy(t, "matrix-"+name)
			if !complete {
				mustRemoveXAINative(t, filepath.Join(dir, "xai-formal-report.json"))
			}
			m := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{{Combo: "xai-responses", Operation: "chat", Provider: "xai", API: "openai-responses"}}}
			p := filepath.Join(t.TempDir(), "matrix.json")
			writeXAINativeJSON(t, p, m)
			before, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(f.ctx, f.matrixBinary, "-matrix", p, dir)
			cmd.Dir, cmd.Env = f.root, xaiProcessEnv(t)
			raw, callErr := cmd.CombinedOutput()
			exit := 0
			if callErr != nil {
				exit = -1
				if e, ok := callErr.(*exec.ExitError); ok {
					exit = e.ExitCode()
				}
			}
			ev.Record("synthetic-matrix-process", map[string]any{"kind": "synthetic-native-contract-fixture", "notActualLive": true, "argv": cmd.Args, "exitCode": exit, "output": string(raw)})
			after, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("synthetic-matrix.json", after)
			var got supportmatrix.Matrix
			if err := json.Unmarshal(after, &got); err != nil {
				t.Fatal(err)
			}
			ev.Check("independent xAI proof required by matrix", (exit == 0) == complete, "exit=%d complete=%t", exit, complete)
			if complete {
				ev.Check("all fixed capabilities projected", len(got.Rows) == 1 && len(got.Rows[0].Capabilities) == 18 && got.Rows[0].AllPassedAt != nil, "missing fixed matrix row")
			} else {
				ev.Check("refused matrix bytes remain unchanged", string(before) == string(after), "matrix changed on refusal")
			}
		})
	}
}

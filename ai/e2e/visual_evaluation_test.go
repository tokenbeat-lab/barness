package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func visualCommand(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	return visualCommandEnv(t, nil, args...)
}
func visualCommandEnv(t *testing.T, extra []string, args ...string) ([]byte, error) {
	t.Helper()
	argv := []string{"run"}
	if os.Getenv("BARNESS_EXAMPLE_RACE") == "1" {
		argv = append(argv, "-race")
	}
	argv = append(argv, "./ai/examples/visualeval")
	cmd := exec.CommandContext(ctxFor(t), "go", append(argv, args...)...)
	cmd.Dir = "../.."
	// No ambient credentials reach the Host evaluation process.
	for _, k := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(k); ok {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, extra...)
	return cmd.CombinedOutput()
}
func visualRead(t *testing.T, dir, name string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	var v map[string]any
	if err = json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestVisualEvaluationFixedWorkflow(t *testing.T) {
	for _, route := range []string{"jev", "decisions"} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(route+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, "VQ12-fixed-"+route+"-"+entry)
				out := filepath.Join(run.Dir(), "visual-"+route+"-"+entry)
				_, err := visualCommand(t, "-route", route, "-entry", entry, "-mode", "controlled", "-out", out)
				if !ev.Check("executable crosses public Classify", err == nil, "command unavailable or replay failed") {
					return
				}
				r := visualRead(t, out, "report.json")
				ev.Record("report", r)
				ev.Check("controlled completion never asserts real quality", r["status"] == "COMPLETE" && r["realQuality"] == "NOT_RUN", "wrong claims")
				ev.Check("all 12 retained", len(r["samples"].([]any)) == 12, "shortened denominator")
				m := r["metrics"].(map[string]any)
				common := m["common"].(map[string]any)
				ev.Check("common task hand scored", common["total"] == float64(8) && common["classCorrect"] == float64(4) && common["existCorrect"] == float64(4) && common["falsePositive"] == float64(4), "wrong scores: %v", common)
				b := r["budget"].(map[string]any)
				ev.Check("no retry and exact fixed calls", b["calls"] == float64(12) && b["attempts"] == float64(12) && b["retries"] == float64(0) && b["inputTokens"] == float64(1200), "wrong budget: %v", b)
				if route == "jev" {
					box := m["box"].(map[string]any)
					ev.Check("IoU independent from existence", box["present"] == float64(2) && box["iouCount"] == float64(2) && box["meanIoU"] == float64(1) && box["consumed"] == float64(1) && box["absentFalsePositive"] == float64(1) && box["absentRawBoxes"] == float64(2), "wrong Box scores: %v", box)
				} else {
					ev.Check("Decisions has no Box result", m["box"] == nil, "invented Box")
				}
				_, err = visualCommand(t, "-verify", out)
				ev.Check("complete package independently verifies", err == nil, "verification failed")
				_, err = visualCommand(t, "-verify", out, "-require-live")
				ev.Check("controlled never satisfies live evidence", err != nil, "controlled promoted")
			})
		}
	}
}

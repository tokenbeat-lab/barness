package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGrokHistoryGLMControlledFullCapture(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-GLM-CONTROLLED-FULL")
	base := t.TempDir()
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-tags", "live", "./ai/live", "-run", "^TestGLM(CandidateProcess|FormalProcess|FormalCaptureValidation)$", "-count=1", "-timeout=180s")
	cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), "BARNESS_AI_EVIDENCE_DIR="+base)
	raw, e := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "output": regionalSafeOutput(raw), "success": e == nil, "vendorCalls": 0, "origin": "controlled"})
	ev.Check("all full GLM producer and independent controlled-native checks pass", e == nil, "%s", raw)
	err := filepath.WalkDir(base, func(path string, d os.DirEntry, e error) error {
		if e != nil || d.IsDir() {
			return e
		}
		rel, e := filepath.Rel(base, path)
		if e != nil {
			return e
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		if e = os.MkdirAll(filepath.Join(run.Dir(), ev.ID(), "controlled-native", filepath.Dir(rel)), 0700); e != nil {
			return e
		}
		ev.Fixture("controlled-native/"+filepath.ToSlash(rel), raw)
		return nil
	})
	ev.Check("complete repeatable controlled raw artifacts retained", err == nil, "%v", err)
}

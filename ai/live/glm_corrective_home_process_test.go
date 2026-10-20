//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise the executable/audit seam when the checkout is inside the auditing
// process's HOME. The home belongs only to an isolated child environment.
func TestGLMCorrectiveDiagnosticHomeProcess(t *testing.T) {
	cs := run.Case(t, "GLM-corrective-1-diagnostic-home")
	base := t.TempDir()
	home, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMCorrectiveDiagnosticHomeChild$", "-test.timeout=140s")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "BARNESS_GLM_DIAGNOSTIC_HOME_TEST=1", "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, err := cmd.CombinedOutput()
	cs.Record("isolated-home-exit", map[string]any{"success": err == nil, "output": string(output)})
	cs.Check("replay diagnostic passes with checkout under controlled HOME", err == nil, "%v %s", err, output)
	files, _ := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
	if !cs.Check("isolated HOME bundle retained", len(files) == 1, "bundles=%d", len(files)) {
		return
	}
	dir := filepath.Dir(files[0])
	cs.Check("full isolated diagnostic evidence retained", glmRetainCorrectiveChild(cs, dir, "home-child") == nil, "copy failed")
	var audit struct{ Findings []struct{ Rule string } }
	raw, _ := os.ReadFile(filepath.Join(dir, "audit.json"))
	json.Unmarshal(raw, &audit)
	cs.Check("unchanged environment audit finds no HOME leak", audit.Findings != nil && len(audit.Findings) == 0, "findings=%d", len(audit.Findings))
	var diagnostic struct{ Output string }
	raw, _ = os.ReadFile(filepath.Join(dir, "GLM-corrective-home-child", "new-capture-replay-exit.json"))
	json.Unmarshal(raw, &diagnostic)
	cs.Check("diagnostic preserves replay path suffix through alias", strings.Contains(diagnostic.Output, "$GLM_REPLAY_ROOT/replayed/") && !strings.Contains(diagnostic.Output, home), "alias and path context required")
}

func TestGLMCorrectiveDiagnosticHomeChild(t *testing.T) {
	if os.Getenv("BARNESS_GLM_DIAGNOSTIC_HOME_TEST") != "1" {
		t.Skip("isolated HOME diagnostic child only")
	}
	cs := run.Case(t, "GLM-corrective-home-child")
	base := t.TempDir()
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMFormal$", "-test.timeout=110s")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=responses", "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_GLM_CONTROLLED=success-responses", "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, err := cmd.CombinedOutput()
	cs.Record("controlled-source-exit", map[string]any{"success": err == nil, "output": string(output)})
	if !cs.Check("small socket-free source capture", err == nil, "%v %s", err, output) {
		return
	}
	files, _ := filepath.Glob(filepath.Join(base, "*", "formal-glm-report.json"))
	if !cs.Check("own source report", len(files) == 1, "reports=%d", len(files)) {
		return
	}
	glmCorrectiveReplay(t, cs, filepath.Dir(files[0]))
}

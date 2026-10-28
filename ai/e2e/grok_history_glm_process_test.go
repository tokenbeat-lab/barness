package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func TestGrokHistoryGLMControlledFullCapture(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-GLM-CONTROLLED-FULL")
	// Keep the child's native case boundaries. The collector discovers this
	// sibling auxiliary bundle and the record consumer verifies every case;
	// copying all 32 cases into this one case exceeded its 512-file ceiling.
	base := filepath.Dir(run.Dir())
	before, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, entry := range before {
		seen[entry.Name()] = true
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "glm-controlled.test")
	build := exec.Command("go", "test", "-c", "-tags", "live", "-o", binary, "./ai/live")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("native build: %v %s", err, out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-test.run=^TestGLM(CandidateProcess|FormalProcess|FormalCaptureValidation)$", "-test.count=1", "-test.timeout=180s")
	cmd.Dir, cmd.Env = filepath.Join(root, "ai/live"), append(xaiProcessEnv(t), "BARNESS_AI_EVIDENCE_DIR="+base)
	raw, processErr := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "output": regionalSafeOutput(raw), "success": processErr == nil})
	ev.Check("full controlled native checks complete", processErr == nil, "%v %s", processErr, regionalSafeOutput(raw))
	after, err := os.ReadDir(base)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, entry := range after {
		if seen[entry.Name()] || !entry.IsDir() {
			continue
		}
		path := filepath.Join(base, entry.Name(), "manifest.json")
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var manifest struct {
			Package string
			Cases   []struct{ ID string }
		}
		if json.Unmarshal(raw, &manifest) != nil || manifest.Package != "-tags live ./ai/live" {
			continue
		}
		count++
		ev.Record("controlled-native-bundle", map[string]any{"sibling": entry.Name(), "manifestSHA256": evidence.Digest(raw), "package": manifest.Package, "nativeCases": len(manifest.Cases)})
	}
	ev.Check("complete original GLM auxiliary bundle retained independently", count == 1, "bundles=%d", count)
}

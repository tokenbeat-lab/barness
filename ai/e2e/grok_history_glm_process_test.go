package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

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
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "-tags", "live", "./ai/live", "-run", "^TestGLM(CandidateProcess|FormalProcess|FormalCaptureValidation)$", "-count=1", "-timeout=180s")
	cmd.Dir, cmd.Env = "../..", append(xaiProcessEnv(t), "BARNESS_AI_EVIDENCE_DIR="+base)
	raw, e := cmd.CombinedOutput()
	ev.Record("process", map[string]any{"argv": cmd.Args, "output": regionalSafeOutput(raw), "success": e == nil, "vendorCalls": 0, "origin": "controlled"})
	ev.Check("all full GLM producer and independent controlled-native checks pass", e == nil, "%s", raw)
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

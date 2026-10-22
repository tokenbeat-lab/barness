package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

// Failure inventory: an unconfigured manual diagnostic must never register a
// mandatory NOT_RUN case; explicit invalid inputs must still register a failed
// diagnostic and preserve the literal original-manifest guard. Child commands
// inherit no credentials and never execute a captured request.
func TestReleaseRecordedBundleDiagnosticProcess(t *testing.T) {
	for _, mode := range []string{"default", "malformed", "wrong-count", "wrong-historical-hash"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "OD57-release-diagnostic-process-"+mode)
			base := t.TempDir()
			cmd := exec.CommandContext(ctxFor(t), "go", "test", "./ai/e2e", "-count=1", "-run", "^TestReleaseRecordedBundleDiagnostic$")
			cmd.Dir = "../.."
			for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
				if value, ok := os.LookupEnv(key); ok {
					cmd.Env = append(cmd.Env, key+"="+value)
				}
			}
			cmd.Env = append(cmd.Env, "BARNESS_AI_EVIDENCE_DIR="+base)
			raw := ""
			switch mode {
			case "malformed":
				raw = "not-json"
			case "wrong-count":
				raw = "[]"
			case "wrong-historical-hash":
				var dirs []string
				for range 2 {
					dir := t.TempDir()
					writeReleaseJSON(t, dir, "manifest.json", map[string]any{"cases": []any{}})
					dirs = append(dirs, dir)
				}
				data, err := json.Marshal(dirs)
				if err != nil {
					t.Fatal(err)
				}
				raw = string(data)
			}
			if raw != "" {
				cmd.Env = append(cmd.Env, "BARNESS_RELEASE_DIAGNOSTIC_BUNDLES="+raw)
			}
			output, processErr := cmd.CombinedOutput()
			dir, err := release.FindBundle(base)
			if err != nil {
				t.Fatal("public diagnostic child evidence unavailable")
			}
			bundle, err := release.LoadBundle(dir)
			if err != nil {
				t.Fatal(err)
			}
			ev.Record("process", map[string]any{"mode": mode, "failed": processErr != nil, "outputSHA256": evidence.Digest(output), "cases": bundle.Cases, "problems": bundle.Problems, "vendorCalls": 0})
			ev.Check("child inventory remains complete", len(bundle.Problems) == 0, "inventory problems=%d", len(bundle.Problems))
			if mode == "default" {
				ev.Check("unconfigured diagnostic skips without evidence case", processErr == nil && len(bundle.Cases) == 0, "failed=%t cases=%v", processErr != nil, bundle.Cases)
				return
			}
			ev.Check("explicit invalid diagnostic remains registered and refused", processErr != nil && len(bundle.Cases) == 1 && slices.Contains(bundle.Cases, release.Case{ID: "OD57-release-recorded-bundle-diagnostic", Status: release.Fail}), "failed=%t cases=%v", processErr != nil, bundle.Cases)
			if mode == "wrong-historical-hash" {
				data, err := os.ReadFile(filepath.Join(dir, "OD57-release-recorded-bundle-diagnostic", "assertions.json"))
				var assertions struct {
					Assertions []struct {
						Name string `json:"name"`
						Pass bool   `json:"pass"`
					}
				}
				if err != nil || json.Unmarshal(data, &assertions) != nil {
					t.Fatal("literal diagnostic guard evidence unavailable")
				}
				ev.Check("both wrong original manifests fail literal guard", len(assertions.Assertions) == 2 && slices.IndexFunc(assertions.Assertions, func(a struct {
					Name string `json:"name"`
					Pass bool   `json:"pass"`
				}) bool {
					return a.Name != "literal original manifest" || a.Pass
				}) == -1, "literal assertions=%d", len(assertions.Assertions))
			}
		})
	}
}

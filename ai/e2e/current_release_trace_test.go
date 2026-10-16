package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestCurrentReleaseIncompleteTrace(t *testing.T) {
	ev := run.Case(t, "UC14-incomplete-pointer-trace")
	base := t.TempDir()
	cmd := exec.CommandContext(ctxFor(t), "go", "test", "./ai/e2e", "-run", "^TestDecisionsReferencesValues$/^escaped-keys$", "-count=1")
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "BARNESS_AI_EVIDENCE_DIR="+base)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("public pointer call: %s", output)
	}
	dir, err := release.FindBundle(base)
	if err != nil {
		t.Fatal(err)
	}
	report := releaseProcess(t, ev, "-bundle", dir)
	for _, id := range []string{"UC-R01", "UC-R03", "UC-T03", "UC-T04", "UC-US01"} {
		i := slices.IndexFunc(report.Trace, func(r release.TraceResult) bool { return r.ID == id })
		ev.Check("partial evidence never proves "+id, i >= 0 && report.Trace[i].Status != release.Pass, "partial trace PASS")
	}
}

func TestCurrentReleaseExtensionPlan(t *testing.T) {
	for _, mode := range []string{"intact", "removed-case", "widened-path", "wrong-baseline"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "UC14-extension-plan-"+mode)
			raw, err := os.ReadFile("../release/classifier-extensions.json")
			if err != nil {
				t.Fatal(err)
			}
			var plan map[string]any
			mustUnmarshal(t, raw, &plan)
			switch mode {
			case "removed-case":
				plan["items"].([]any)[0].(map[string]any)["caseIDs"] = []any{}
			case "widened-path":
				plan["items"].([]any)[0].(map[string]any)["paths"] = []any{"*"}
			case "wrong-baseline":
				plan["piVersion"] = "1.0.0"
			}
			path := t.TempDir()
			writeReleaseJSON(t, path, "extensions.json", plan)
			report := releaseProcess(t, ev, "-extensions", filepath.Join(path, "extensions.json"))
			i := slices.IndexFunc(report.Snapshots, func(s release.Snapshot) bool { return s.Name == "classifier-extension-plan" })
			ev.Check("finite extension contract pinned", i >= 0 && report.Snapshots[i].Current == (mode == "intact"), "extension plan accepted wrong scope")
		})
	}
}

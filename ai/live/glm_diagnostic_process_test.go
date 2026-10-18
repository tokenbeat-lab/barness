//go:build live

package live

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
)

// The public child binary uses only the existing synthetic transport: no socket
// or vendor credential. Refused formal subsetting remains independently checked.
func TestGLMMessagesCancelDiagnosticProcess(t *testing.T) {
	selectors := map[string]string{}
	for _, name := range []string{"live-captures.json", "live-captures.sha256", "known-usage.json", "support-views.json"} {
		data, err := readDecisionsFile("testdata/glm-formal/"+name, 2<<20)
		if err != nil {
			t.Fatal(err)
		}
		selectors[name] = sha256Text(data)
	}
	defer func() {
		for name, want := range selectors {
			data, err := readDecisionsFile("testdata/glm-formal/"+name, 2<<20)
			if err != nil || sha256Text(data) != want {
				t.Errorf("diagnostic changed selector %s", name)
			}
		}
	}()

	for _, name := range []string{"success", "narrowed", "wrong-route", "wrong-model", "formal-mixed", "formal-narrowed", "proxy", "missing-key", "429"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "GLM18-diagnostic-process-"+name)
			root := t.TempDir()
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=messages", "BARNESS_GLM_DIAGNOSTIC=glm-5.3/cancel", "BARNESS_GLM_KEY=synthetic-glm-diagnostic", "BARNESS_GLM_CONTROLLED=success-messages", "BARNESS_AI_EVIDENCE_DIR=" + root}
			entry := "^TestGLMMessagesCancelDiagnostic$"
			switch name {
			case "narrowed":
				entry += "/cancel"
			case "wrong-route":
				env[1] = "BARNESS_GLM_ROUTE=responses"
			case "wrong-model":
				env[2] = "BARNESS_GLM_DIAGNOSTIC=glm-5.2/cancel"
			case "formal-mixed":
				env = append(env, "BARNESS_GLM_FORMAL=1")
			case "formal-narrowed":
				env[2] = "BARNESS_GLM_DIAGNOSTIC="
				env = append(env, "BARNESS_GLM_FORMAL=1")
				entry = "^TestGLMFormal$/^glm-5.3$/^cancel$"
			case "proxy":
				env = append(env, "BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890")
			case "missing-key":
				env[3] = "BARNESS_GLM_KEY="
			case "429":
				env[4] = "BARNESS_GLM_CONTROLLED=429"
			}
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=25s")
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("command", map[string]any{"argv": cmd.Args, "environment": env, "success": err == nil, "output": string(output)})
			wantOK := name == "success" || name == "missing-key"
			cs.Check("exact public selection exit", (err == nil) == wantOK, "%v %s", err, output)
			reportName := "diagnostic-glm-messages-cancel-report.json"
			if name == "formal-narrowed" {
				reportName = "formal-glm-report.json"
			}
			paths, _ := filepath.Glob(filepath.Join(root, "*", reportName))
			if !cs.Check("distinct bounded diagnostic report retained", len(paths) == 1, "files=%v", paths) {
				entries, _ := os.ReadDir(root)
				for _, entry := range entries {
					if entry.IsDir() {
						cs.Check("refused child retained", glmRetainCorrectiveChild(cs, filepath.Join(root, entry.Name()), "refused-child") == nil, "copy failed")
					}
				}
				return
			}
			dir := filepath.Dir(paths[0])
			cs.Check("complete raw controlled child retained", glmRetainCorrectiveChild(cs, dir, "child") == nil, "copy failed")
			data, readErr := readDecisionsFile(paths[0], 1<<20)
			var report glmReport
			if !cs.Check("report decodes", readErr == nil && json.Unmarshal(data, &report) == nil, "%v", readErr) {
				return
			}
			cs.Record("detail", json.RawMessage(data))
			_, liveErr := os.Stat(filepath.Join(dir, "live-report.json"))
			cs.Check("never publishes support", os.IsNotExist(liveErr), "live support exists")
			if name == "formal-narrowed" {
				cs.Check("formal22 subsetting still refused before calls", report.Kind == "glm-formal" && len(report.Expected) == 22 && report.Budget.CallsUsed == 0 && report.Budget.HTTPAttempts == 0 && !report.Complete, "%s", data)
				return
			}
			cs.Check("diagnostic scope only", report.Kind == "glm-single-case-diagnostic-only" && report.Budget.MaxCalls == 1 && report.Budget.MaxAttempts == 1 && report.Budget.MaxOutputTokens == 4096 && report.Budget.CallSeconds == 120 && report.Budget.ProcessSeconds == 180 && report.Budget.FinalizationReserveSeconds == 20, "%s", data)
			cs.Check("no absent model certification", len(report.Models) <= 1 && report.Models["glm-5.2"] == "", "models=%v", report.Models)
			for _, leaf := range []string{"formal-glm-report.json", "formal-glm-5.3-report.json", "candidate-glm-5.2-report.json", "candidate-glm-5.3-report.json"} {
				_, e := os.Stat(filepath.Join(dir, leaf))
				cs.Check("no unrelated report "+leaf, os.IsNotExist(e), "unexpected report")
			}
			if name == "success" || name == "429" {
				hash, _ := ai.BuiltinCatalog().Hash()
				cs.Check("fixed builtin one-case source", report.CatalogHash == hash && len(report.Expected) == 1 && report.Expected[0] == "glm-5.3/cancel" && len(report.Scenarios) == 1, "wrong selected identity")
				cs.Check("exact finite actual accounting", report.Budget.CallsUsed == 1 && report.Budget.AttemptsReserved == 1 && report.Budget.HTTPAttempts == 1 && report.Budget.Retries == 0 && report.Scenarios[0].Calls == 1 && report.Scenarios[0].HTTPAttempts == 1, "%+v", report.Budget)
				cs.Check("truthful diagnostic result", (report.Outcome == "PASS" && report.Complete) == (name == "success"), "%s", data)
			} else {
				cs.Check("refusal or not-run spends zero", report.Budget.CallsUsed == 0 && report.Budget.HTTPAttempts == 0 && !report.Complete, "%s", data)
			}
		})
	}
}

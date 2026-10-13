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
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// OD54 public process seam: own key, official catalog, full reserved pipeline,
// foreign-key refusal, unexecuted/malformed/fault reports and final audit status.
func TestDecisionsFormalProcess(t *testing.T) {
	if os.Getenv("BARNESS_DECISIONS_FORMAL_CONTROLLED") != "" {
		return
	}
	for _, name := range []string{"success", "not-run", "foreign-key", "endpoint-override", "source-changed", "narrowed", "429", "504", "questions-budget", "audit-failed", "publish-crash"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "P10-OD54-formal-process-"+name)
			base := t.TempDir()
			env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "TMPDIR=" + os.TempDir(), "BARNESS_DECISIONS_FORMAL_CONTROLLED=" + name, envLive + "=1", envCombo + "=openai-decisions", keyVar("openai-decisions") + "=key:controlled-decisions", "BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1", "BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890", "BARNESS_AI_EVIDENCE_DIR=" + base}
			if name == "not-run" {
				env = append(env, keyVar("openai-decisions")+"=")
			}
			if name == "foreign-key" {
				env = append(env, keyVar("openai-decisions-candidate")+"=key:foreign")
			}
			if name == "endpoint-override" {
				env = append(env, "OPENAI_BASE_URL=https://invalid.example")
			}
			dir, _ := os.Getwd()
			if name == "source-changed" {
				dir = filepath.Join(base, "source")
				os.MkdirAll(filepath.Join(dir, "testdata", "decisions"), 0700)
				files, _ := os.ReadDir("testdata/decisions")
				for _, f := range files {
					if !f.IsDir() {
						raw, _ := os.ReadFile(filepath.Join("testdata/decisions", f.Name()))
						os.WriteFile(filepath.Join(dir, "testdata", "decisions", f.Name()), raw, 0600)
					}
				}
				os.WriteFile(filepath.Join(dir, "testdata", "decisions", "guide.md"), []byte("changed"), 0600)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestDecisionsFormalControlledChild$", "-test.v")
			cmd.Dir = dir
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			paths, _ := filepath.Glob(filepath.Join(base, "*", supportmatrix.ReportFile))
			if len(paths) != 1 {
				t.Fatalf("formal report missing: %v %s", err, output)
			}
			raw, readErr := os.ReadFile(paths[0])
			if readErr != nil {
				t.Fatal(readErr)
			}
			var r supportmatrix.Report
			if json.Unmarshal(raw, &r) != nil {
				t.Fatal("invalid report")
			}
			cs.Record("formal-report", r)
			cs.Record("process", map[string]any{"exitFailed": err != nil, "output": string(output)})
			cat := ai.BuiltinCatalog()
			hash, _ := cat.Hash()
			cs.Check("formal route and price identity", r.Schema == 3 && r.Combo == "openai-decisions" && r.Operation == "classifier" && r.Provider == "openai" && r.API == "openai-decisions" && r.Model == "gpt-6-luna" && r.CatalogVersion == cat.Version && r.CatalogHash == hash, "wrong formal identity")
			cs.Check("all required outcomes emitted", len(r.Expected) == 9 && len(r.Scenarios) >= 9, "narrowed report")
			if name == "success" {
				cs.Check("known consumption reports actual successful counters", r.Budget.InputTokensUsed == 800 && r.Budget.OutputTokensUsed == 0, "input=%d output=%d", r.Budget.InputTokensUsed, r.Budget.OutputTokensUsed)
				cs.Check("controlled full formal pipeline", err == nil && r.Decisions != nil && r.Decisions.Complete && r.Decisions.AuditPassed && r.Budget.CallsUsed == 9 && r.Budget.HTTPAttempts == 9 && r.Budget.QuestionsUsed == 13 && r.Budget.ImagesUsed == 3, "incomplete controlled smoke")
			}
			if name == "not-run" {
				cs.Check("no credential no spend", err == nil && !r.Decisions.Complete && r.Budget.CallsUsed == 0, "wrong NOT_RUN")
			}
			if name != "success" && name != "not-run" && name != "narrowed" {
				cs.Check("failure never formal PASS", err != nil && r.Decisions != nil && !r.Decisions.Complete, "failure promoted")
			}
			_, mergeErr := supportmatrix.Merge(supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "openai-decisions", Operation: "classifier", Provider: "openai", API: "openai-decisions"}}}, r)
			cs.Check("controlled report cannot become real support", mergeErr != nil, "synthetic report accepted")
			if name == "success" || name == "not-run" {
				r.Decisions.Mode = "live"
				_, mergeErr = supportmatrix.Merge(supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "openai-decisions", Operation: "classifier", Provider: "openai", API: "openai-decisions"}}}, r)
				cs.Check("full producer matches formal consumer contract", mergeErr == nil, "err=%v", mergeErr)
			}
			root := filepath.Dir(paths[0])
			if name == "success" {
				a := release.AuditBundle(root, nil)
				cs.Record("consumer-audit", a)
				cs.Check("schema3 public evidence consumer accepts own full captures", a.Err == "" && len(a.Findings) == 0, "err=%s", a.Err)
			}
			for _, file := range []string{"manifest.json", "audit.json", "decisions-integrity.json"} {
				data, e := os.ReadFile(filepath.Join(root, file))
				cs.Check("final "+file, e == nil || (name == "audit-failed" && file == "decisions-integrity.json"), "missing")
				if e == nil {
					cs.Fixture(file, data)
				}
			}
		})
	}
}
func TestDecisionsFormalControlledChild(t *testing.T) {
	name := os.Getenv("BARNESS_DECISIONS_FORMAL_CONTROLLED")
	if name == "" {
		t.Skip("controlled child only")
	}
	if name == "publish-crash" {
		beforeDecisionsPublication = func() { os.Exit(91) }
	}
	driveDecisions(t, controlledDecisionsRecorder(t, name), func(t *testing.T) []decisionsScenario { return controlledCandidateBudgetScenarios(t, name) }, "controlled")
	if name == "audit-failed" {
		os.WriteFile(filepath.Join(run.Dir(), "unredacted.txt"), []byte(cfg.key), 0600)
	}
}

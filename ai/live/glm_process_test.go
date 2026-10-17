//go:build live

package live

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Public executable seam. Children use one synthetic credential and a transport
// that cannot open a socket; intentional failures remain in their own bundles.
func TestGLMCandidateProcess(t *testing.T) {
	for _, name := range []string{"not-run", "switch-off", "narrowed", "foreign-key", "mixed-route", "ambient-key", "endpoint-override", "source-changed", "source-missing", "catalog-changed", "plan-changed", "wrong-route", "success-chat", "success-responses", "success-messages", "429", "missing-text", "calls-budget", "attempts-budget", "actual-budget", "deadline", "inventory-failure"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "GLM11-process-"+name)
			base := t.TempDir()
			cwd, _ := os.Getwd()
			route := "chat"
			if strings.HasPrefix(name, "success-") {
				route = strings.TrimPrefix(name, "success-")
			}
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=" + route, "BARNESS_GLM_CONTROLLED=" + name, "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_AI_EVIDENCE_DIR=" + base}
			switch name {
			case "not-run":
				env[3] = "BARNESS_GLM_KEY="
			case "switch-off":
				env[0] = "BARNESS_AI_LIVE=0"
			case "foreign-key":
				env = append(env, keyVar("openai-chat")+"=synthetic-foreign")
			case "mixed-route":
				env = append(env, "BARNESS_MM_ROUTE=jev")
			case "ambient-key":
				env = append(env, "OPENAI_API_KEY=synthetic-foreign")
			case "endpoint-override":
				env = append(env, "OPENAI_BASE_URL=https://invalid.example")
			}
			if name == "wrong-route" {
				env[1] = "BARNESS_GLM_ROUTE=invalid"
			}
			if strings.HasPrefix(name, "source-") || name == "catalog-changed" || name == "plan-changed" {
				input := filepath.Join(base, "inputs", "testdata", "glm-candidate")
				if e := os.MkdirAll(input, 0755); e != nil {
					t.Fatal(e)
				}
				files, _ := os.ReadDir("testdata/glm-candidate")
				for _, f := range files {
					raw, e := os.ReadFile(filepath.Join("testdata/glm-candidate", f.Name()))
					if e != nil {
						t.Fatal(e)
					}
					if e = os.WriteFile(filepath.Join(input, f.Name()), raw, 0600); e != nil {
						t.Fatal(e)
					}
				}
				target := "pricing.txt"
				if name == "catalog-changed" {
					target = "candidate.json"
				}
				if name == "plan-changed" {
					target = "plan.json"
				}
				path := filepath.Join(input, target)
				if name == "source-missing" {
					os.Remove(path)
				} else {
					f, e := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
					if e != nil {
						t.Fatal(e)
					}
					f.WriteString("\nchanged")
					f.Close()
				}
				cwd = filepath.Join(base, "inputs")
			}
			entry := "^TestGLMCandidate$"
			if name == "narrowed" {
				entry = "^TestGLMCandidate$/glm-5.3"
			}
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=40s")
			cmd.Dir = cwd
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
			wantOK := strings.HasPrefix(name, "success-") || name == "not-run" || name == "switch-off"
			cs.Check("child exit matches contract", (err == nil) == wantOK, "child exit %v: %s", err, output)
			matches, _ := filepath.Glob(filepath.Join(base, "*", "candidate-glm-report.json"))
			if !cs.Check("candidate report retained", len(matches) == 1, "reports=%d", len(matches)) {
				return
			}
			childDir := filepath.Dir(matches[0])
			copyErr := filepath.WalkDir(childDir, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					return nil
				}
				rel, e := filepath.Rel(childDir, path)
				if e != nil {
					return e
				}
				target := filepath.Join("child-evidence", rel)
				if e = os.MkdirAll(filepath.Join(run.Dir(), cs.ID(), filepath.Dir(target)), 0755); e != nil {
					return e
				}
				data, e := readDecisionsFile(path, 16<<20)
				if e != nil {
					return e
				}
				cs.Fixture(filepath.ToSlash(target), data)
				return nil
			})
			cs.Check("raw controlled child evidence retained", copyErr == nil, "%v", copyErr)
			raw, _ := os.ReadFile(matches[0])
			var r struct {
				Kind, Outcome         string
				Complete, AuditPassed bool
				Budget                struct{ CallsUsed, AttemptsReserved, HTTPAttempts int }
				Expected              []string
			}
			json.Unmarshal(raw, &r)
			cs.Record("child-report", json.RawMessage(raw))
			cs.Check("never formal", r.Kind == "glm-candidate-only", "kind=%s", r.Kind)
			formal, _ := filepath.Glob(filepath.Join(base, "*", "live-report.json"))
			cs.Check("no formal publication", len(formal) == 0, "formal=%v", formal)
			foreign, _ := filepath.Glob(filepath.Join(base, "*", "multimodal-report.json"))
			cs.Check("no foreign route report", len(foreign) == 0, "foreign report was initialized")
			if strings.HasPrefix(name, "success-") {
				cs.Check("full own route", r.Outcome == "PASS" && r.Complete && r.AuditPassed && r.Budget.CallsUsed == 22 && r.Budget.HTTPAttempts == 22 && len(r.Expected) == 22, "%s", raw)
				dir := filepath.Dir(matches[0])
				inventoryRaw, e := os.ReadFile(filepath.Join(dir, "glm-integrity.json"))
				var inventory struct{ Files map[string]string }
				if !cs.Check("sealed inventory available", e == nil && json.Unmarshal(inventoryRaw, &inventory) == nil, "%v", e) {
					return
				}
				for name, want := range inventory.Files {
					data, e := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
					sum := sha256.Sum256(data)
					cs.Check("inventory hash "+name, e == nil && hex.EncodeToString(sum[:]) == want, "mismatch")
				}
				manifestRaw, e := os.ReadFile(filepath.Join(dir, "manifest.json"))
				var manifest struct{ Versions map[string]string }
				json.Unmarshal(manifestRaw, &manifest)
				cs.Check("manifest uses actual candidate snapshot", e == nil && manifest.Versions["model_catalog_version"] == "glm-six-candidate-2026-10-17.1-sourced-usd" && manifest.Versions["glm_source_manifest_hash"] == glmSourcePin, "candidate provenance changed")
				if name == "success-messages" {
					for _, kind := range []string{"native-history", "control-unsigned"} {
						data, e := os.ReadFile(filepath.Join(dir, "GLM11-glm-5.2-"+kind, "request.json"))
						var req struct {
							Max      int `json:"max_tokens"`
							Thinking struct {
								Type   string
								Budget int `json:"budget_tokens"`
							}
						}
						json.Unmarshal(data, &req)
						cs.Check("5.2 native keeps known thinking shape "+kind, e == nil && req.Max == 4096 && req.Thinking.Type == "enabled" && req.Thinking.Budget == 1024, "unexpected native thinking shape")
					}
				}
				for _, model := range []string{"glm-5.3", "glm-5.2"} {
					data, e := os.ReadFile(filepath.Join(dir, "candidate-"+model+"-report.json"))
					var detail struct {
						Model, Outcome      string
						Complete            bool
						Calls, HTTPAttempts int
						Scenarios           []struct{ ID, Model string }
					}
					json.Unmarshal(data, &detail)
					cs.Check("independent model report "+model, e == nil && detail.Complete && detail.Outcome == "PASS" && detail.Model == model && detail.Calls == 11 && detail.HTTPAttempts == 11 && len(detail.Scenarios) == 11, "incomplete own model report")
					for _, sc := range detail.Scenarios {
						cs.Check("model-qualified capability "+sc.ID, sc.Model == model && strings.HasPrefix(sc.ID, model+"/"), "cross-model certification")
					}
				}

			}
			if name == "not-run" || name == "switch-off" {
				cs.Check("not run spends zero", r.Outcome == "NOT_RUN" && !r.Complete && r.Budget.CallsUsed == 0 && r.Budget.HTTPAttempts == 0, "%s", raw)
			}
			if !wantOK {
				cs.Check("failure is retained", r.Outcome == "FAIL" && !r.Complete, "%s", raw)
			}
		})
	}
}

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

func TestGLMCaptureReplayProcess(t *testing.T) {
	var index struct {
		Routes map[string]struct{ Archive string }
	}
	raw, e := os.ReadFile("testdata/glm-candidate/live-captures.json")
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("fixed capture index unavailable")
	}
	for _, name := range []string{"chat", "responses", "messages", "missing", "mutated", "wrong-route", "credential", "narrowed", "provider-id", "native-signature"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "GLM11-replay-process-"+name)
			route := name
			if _, ok := index.Routes[route]; !ok {
				route = "responses"
			}
			archive := filepath.Join("../..", index.Routes[route].Archive)
			archive, _ = filepath.Abs(archive)
			base := t.TempDir()
			if name == "missing" {
				archive = filepath.Join(base, "missing.zip")
			}
			if name == "mutated" {
				raw, e := os.ReadFile(archive)
				if e != nil {
					t.Fatal(e)
				}
				archive = filepath.Join(base, "changed.zip")
				raw = append(raw, byte('x'))
				os.WriteFile(archive, raw, 0600)
			}
			if name == "wrong-route" {
				route = "chat"
			}
			env := []string{"BARNESS_AI_LIVE=0", "BARNESS_GLM_ROUTE=" + route, "BARNESS_GLM_REPLAY_ARCHIVE=" + archive, "BARNESS_AI_EVIDENCE_DIR=" + base, "BARNESS_GLM_REPLAY_TESTCASE=" + name}
			if name == "credential" {
				env = append(env, "BARNESS_GLM_KEY=forbidden-synthetic-replay-credential")
			}
			if name == "provider-id" || name == "native-signature" {
				env = append(env, "BARNESS_GLM_REPLAY_FAULT="+name)
			}
			entry := "^TestGLMCaptureReplay$"
			if name == "narrowed" {
				entry += "/glm-5.3"
			}
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=50s")
			cwd, _ := os.Getwd()
			cmd.Dir = cwd
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
			success := name == "chat" || name == "responses" || name == "messages"
			cs.Check("public replay exit", (err == nil) == success, "%v: %s", err, output)
			matches, _ := filepath.Glob(filepath.Join(base, "*", "capture-glm-replay-report.json"))
			if !cs.Check("separate replay report", len(matches) == 1, "reports=%d", len(matches)) {
				return
			}
			data, _ := os.ReadFile(matches[0])
			cs.Record("replay-report", json.RawMessage(data))
			var report glmReport
			json.Unmarshal(data, &report)
			cs.Check("never live", report.Mode == "replay" && report.Kind == "glm-capture-replay-only", "mode=%s kind=%s", report.Mode, report.Kind)
			forbidden, _ := filepath.Glob(filepath.Join(base, "*", "candidate-glm-report.json"))
			cs.Check("original candidate marker absent", len(forbidden) == 0, "mislabelled report")
			if success {
				cs.Check("exact own22 replay", report.Complete && report.Outcome == "PASS" && report.Budget.CallsUsed == 22 && report.Budget.HTTPAttempts == 0 && report.Budget.ReplayAttempts == 22, "%s", data)
			} else {
				cs.Check("failure preserved", !report.Complete && report.Outcome == "FAIL", "%s", data)
			}
			child := filepath.Dir(matches[0])
			if name == "provider-id" || name == "native-signature" {
				data, e := os.ReadFile(filepath.Join(child, "GLM11-glm-5.3-full-stream", "capture-result-comparison.json"))
				var comparison struct{ Equal bool }
				json.Unmarshal(data, &comparison)
				cs.Check("specific upstream field comparison rejects mutation", e == nil && !comparison.Equal, "comparison did not reject changed upstream field")
			}

			e = filepath.WalkDir(child, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(child, path)
				target := filepath.Join("child-evidence", rel)
				if e = os.MkdirAll(filepath.Join(run.Dir(), cs.ID(), filepath.Dir(target)), 0755); e != nil {
					return e
				}
				raw, e := readDecisionsFile(path, 16<<20)
				if e != nil {
					return e
				}
				cs.Fixture(filepath.ToSlash(target), raw)
				return nil
			})
			cs.Check("raw replay child retained", e == nil, "%v", e)
			if strings.Contains(string(output), "normalized provider") {
				cs.Check("upstream IDs never normalized", false, "illegal normalization")
			}
		})
	}
}

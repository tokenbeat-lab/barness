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

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestGLMCaptureReplayProcess(t *testing.T)       { driveGLMCaptureReplayProcess(t, false) }
func TestGLMFormalCaptureReplayProcess(t *testing.T) { driveGLMCaptureReplayProcess(t, true) }
func driveGLMCaptureReplayProcess(t *testing.T, formal bool) {
	var index struct {
		Routes map[string]struct{ Archive string }
	}
	indexPath := "testdata/glm-candidate/live-captures.json"
	if formal {
		indexPath = "testdata/glm-formal/live-captures.json"
	}
	raw, e := os.ReadFile(indexPath)
	if e != nil || json.Unmarshal(raw, &index) != nil {
		t.Fatal("fixed capture index unavailable")
	}
	names := []string{"chat", "responses", "messages", "missing", "mutated", "wrong-route", "credential", "narrowed", "provider-id", "native-signature"}
	if formal {
		names = append(names, "index-mutated", "snapshot-unavailable", "snapshot-hash", "snapshot-overflow")
	}
	for _, name := range names {
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
			if formal {
				env = append(env, "BARNESS_GLM_FORMAL=1")
			}
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
			if name == "index-mutated" || name == "snapshot-unavailable" {
				inputs := filepath.Join(base, "inputs")
				for _, source := range []string{"glm-candidate", "glm-formal"} {
					if e := os.MkdirAll(filepath.Join(inputs, "testdata", source), 0755); e != nil {
						t.Fatal(e)
					}
					if e := os.CopyFS(filepath.Join(inputs, "testdata", source), os.DirFS(filepath.Join("testdata", source))); e != nil {
						t.Fatal(e)
					}
				}
				p := filepath.Join(inputs, "testdata/glm-formal/live-captures.json")
				raw, e := os.ReadFile(p)
				if e != nil {
					t.Fatal(e)
				}
				if name == "index-mutated" {
					os.WriteFile(p, append(raw, ' '), 0600)
				}
				cwd = inputs
			}
			cmd.Dir = cwd
			env = append(env, "PATH="+os.Getenv("PATH"))
			if name == "snapshot-hash" || name == "snapshot-overflow" {
				realGit, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				payload := []byte(`{"hash":"sha256:wrong","catalog":{"version":"wrong"}}`)
				if name == "snapshot-overflow" {
					payload = []byte(strings.Repeat(" ", (2<<20)+1))
				}
				file := filepath.Join(base, "snapshot.json")
				os.WriteFile(file, payload, 0600)
				bin := filepath.Join(base, "bin")
				os.Mkdir(bin, 0700)
				script := "#!/bin/sh\ncase \"$*\" in *show*) exec /bin/cat '" + file + "';; *) exec '" + realGit + "' \"$@\";; esac\n"
				os.WriteFile(filepath.Join(bin, "git"), []byte(script), 0700)
				env = append(env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))
			}
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
			if formal && success {
				_, certificationError := supportmatrix.NewGLMReport(child)
				published, _ := filepath.Glob(filepath.Join(child, "live-report.json"))
				cs.Check("historical replay cannot certify current live support", certificationError != nil && len(published) == 0, "replay certified current support")
				raw, err := os.ReadFile(filepath.Join(child, "capture-replay-catalog.json"))
				cs.Check("original catalog authority retained", err == nil && strings.Contains(string(raw), report.CatalogHash), "historical catalog provenance missing")
			}
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

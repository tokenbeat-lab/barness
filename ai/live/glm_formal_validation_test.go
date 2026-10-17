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

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// No-network captured-process seam. These are explicitly transformed synthetic
// fixtures for import validation, never vendor evidence or support publication.
func TestGLMFormalCaptureValidation(t *testing.T) {
	cs := run.Case(t, "GLM12-capture-validation")
	base := t.TempDir()
	exe, _ := os.Executable()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-test.run=^TestGLMFormal$", "-test.timeout=50s")
	cmd.Env = []string{"BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=chat", "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_GLM_CONTROLLED=success-chat", "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, e := cmd.CombinedOutput()
	if !cs.Check("synthetic source process", e == nil, "%v %s", e, output) {
		return
	}
	files, _ := filepath.Glob(filepath.Join(base, "*", "formal-glm-report.json"))
	if len(files) != 1 {
		t.Fatal("missing controlled capture")
	}
	original := filepath.Dir(files[0])
	for _, name := range []string{"baseline", "honest-no-request-id", "fabricated-request-id", "source-plan", "source-candidate", "source-price", "foreign-operation", "stale-result-version", "execution-add", "execution-delete", "execution-change"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if e := os.CopyFS(dir, os.DirFS(original)); e != nil {
				t.Fatal(e)
			}
			edit := func(path string, f func(map[string]any)) {
				raw, e := os.ReadFile(filepath.Join(dir, path))
				if e != nil {
					t.Fatal(e)
				}
				var v map[string]any
				if json.Unmarshal(raw, &v) != nil {
					t.Fatal("fixture JSON")
				}
				f(v)
				b, _ := json.MarshalIndent(v, "", "  ")
				os.WriteFile(filepath.Join(dir, path), append(b, '\n'), 0600)
			}
			edit("formal-glm-report.json", func(v map[string]any) { v["mode"] = "live" })
			edit("manifest.json", func(v map[string]any) { v["versions"].(map[string]any)["glm_mode"] = "live" })
			if name == "honest-no-request-id" || name == "fabricated-request-id" {
				id := ""
				if name == "fabricated-request-id" {
					id = "invented-local-id"
				}
				edit("formal-glm-report.json", func(v map[string]any) {
					s := v["scenarios"].([]any)[0].(map[string]any)
					s["providerRequestIds"] = []string{id}
					s["requestIdSources"] = []string{"unreported"}
				})
				p := filepath.Join(dir, "GLM11-glm-5.3-full-stream")
				raw, _ := os.ReadFile(filepath.Join(p, "attempt-1-exchanges.json"))
				var ex []map[string]any
				json.Unmarshal(raw, &ex)
				ex[0]["responseHeaders"] = map[string]any{"Content-Type": "text/event-stream"}
				raw, _ = json.Marshal(ex)
				os.WriteFile(filepath.Join(p, "attempt-1-exchanges.json"), raw, 0600)
				edit("GLM11-glm-5.3-full-stream/attempt-1-result.json", func(v map[string]any) {
					v["result"].(map[string]any)["metadata"].(map[string]any)["attempts"].([]any)[0].(map[string]any)["providerRequestId"] = id
				})
				raw, _ = os.ReadFile(filepath.Join(p, "observations-1.json"))
				var obs []map[string]any
				json.Unmarshal(raw, &obs)
				for _, o := range obs {
					if c, ok := o["call"].(map[string]any); ok {
						if attempts, ok := c["attempts"].([]any); ok {
							for _, a := range attempts {
								a.(map[string]any)["providerRequestId"] = id
							}
						}
					}
					if a, ok := o["attempt"].(map[string]any); ok {
						a["providerRequestId"] = id
					}
				}
				raw, _ = json.Marshal(obs)
				os.WriteFile(filepath.Join(p, "observations-1.json"), raw, 0600)
			}
			if name == "source-plan" || name == "source-candidate" || name == "source-price" {
				target := "plan.json"
				if name == "source-candidate" {
					target = "candidate.json"
				}
				if name == "source-price" {
					target = "pricing.txt"
				}
				p := filepath.Join(dir, "GLM11-sources", target)
				raw, e := os.ReadFile(p)
				if e != nil {
					t.Fatal("source snapshot was not captured:", target)
				}
				os.WriteFile(p, append(raw, []byte("\nchanged snapshot")...), 0600)
			}
			if name == "foreign-operation" || name == "stale-result-version" {
				field, value := "operation", "classifier"
				if name == "stale-result-version" {
					field, value = "catalogVersion", "old-catalog"
				}
				edit("GLM11-glm-5.3-full-stream/attempt-1-result.json", func(v map[string]any) { v["result"].(map[string]any)["metadata"].(map[string]any)[field] = value })
				p := filepath.Join(dir, "GLM11-glm-5.3-full-stream/observations-1.json")
				raw, _ := os.ReadFile(p)
				var obs []map[string]any
				json.Unmarshal(raw, &obs)
				for _, o := range obs {
					if call, ok := o["call"].(map[string]any); ok {
						call[field] = value
					}
				}
				raw, _ = json.Marshal(obs)
				os.WriteFile(p, raw, 0600)
			}
			if name == "execution-add" || name == "execution-delete" || name == "execution-change" {
				cwd, _ := os.Getwd()
				root := filepath.Clean(filepath.Join(cwd, "../.."))
				sourceRoot := t.TempDir()
				raw, _ := os.ReadFile(filepath.Join(dir, "glm-execution-source.json"))
				var sources map[string]string
				json.Unmarshal(raw, &sources)
				for file := range sources {
					p := filepath.Join(sourceRoot, file)
					os.MkdirAll(filepath.Dir(p), 0755)
					raw, e := os.ReadFile(filepath.Join(root, file))
					if e != nil {
						t.Fatal(e)
					}
					os.WriteFile(p, raw, 0600)
				}
				target := filepath.Join(sourceRoot, "ai/examples/requestid/requestid.go")
				if name == "execution-add" {
					target = filepath.Join(sourceRoot, "ai/runtime_added.go")
					os.WriteFile(target, []byte("package ai\n"), 0600)
				}
				if name == "execution-delete" {
					os.Remove(target)
				}
				if name == "execution-change" {
					os.WriteFile(target, []byte("package requestid\n"), 0600)
				}
				t.Chdir(sourceRoot)
			}
			// Coherent local rehashing is insufficient authority to alter pinned source
			// facts. This deliberately gives the verifier every editable inventory hash.
			edit("manifest.json", func(v map[string]any) {
				for _, x := range v["cases"].([]any) {
					c := x.(map[string]any)
					entries, _ := os.ReadDir(filepath.Join(dir, c["dir"].(string)))
					hashes := map[string]string{}
					for _, f := range entries {
						raw, _ := os.ReadFile(filepath.Join(dir, c["dir"].(string), f.Name()))
						hashes[f.Name()] = sha256Text(raw)
					}
					c["artifacts"] = hashes
				}
			})
			hashes := map[string]string{}
			filepath.WalkDir(dir, func(p string, d os.DirEntry, e error) error {
				if e != nil || d.IsDir() {
					return e
				}
				rel, _ := filepath.Rel(dir, p)
				if rel != "glm-integrity.json" {
					raw, _ := os.ReadFile(p)
					hashes[filepath.ToSlash(rel)] = sha256Text(raw)
				}
				return nil
			})
			raw, _ := json.Marshal(map[string]any{"schema": 1, "files": hashes})
			os.WriteFile(filepath.Join(dir, "glm-integrity.json"), raw, 0600)
			_, err := supportmatrix.NewGLMReport(dir)
			want := name == "baseline" || name == "honest-no-request-id"
			cs.Check("captured import "+name, (err == nil) == want, "%v", err)
		})
	}
	cs.Record("scope", "Synthetic controlled captures transformed solely to exercise public import validation; zero vendor HTTP calls; never published support")
}

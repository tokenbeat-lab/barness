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

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure-first executable boundary: formal catalog selection is explicit,
// controlled transports never certify live support and narrow entry fails closed.
func TestGLMFormalProcess(t *testing.T) {
	for _, name := range []string{"chat", "responses", "messages", "missing-key", "narrowed", "429", "inventory-failure", "candidate-entry", "proxy-override"} {
		t.Run(name, func(t *testing.T) {
			cs := run.Case(t, "GLM12-formal-process-"+name)
			base := t.TempDir()
			route := name
			if name != "chat" && name != "responses" && name != "messages" {
				route = "chat"
			}
			env := []string{"BARNESS_AI_LIVE=1", "BARNESS_GLM_ROUTE=" + route, "BARNESS_GLM_FORMAL=1", "BARNESS_GLM_KEY=synthetic-glm", "BARNESS_GLM_CONTROLLED=success-" + route, "BARNESS_AI_EVIDENCE_DIR=" + base}
			if name == "proxy-override" {
				env = append(env, "BARNESS_AI_LIVE_PROXY=http://127.0.0.1:7890")
			}
			if name == "missing-key" {
				env[3] = "BARNESS_GLM_KEY="
			}
			if name == "429" || name == "inventory-failure" {
				env[4] = "BARNESS_GLM_CONTROLLED=" + name
			}
			entry := "^TestGLMFormal$"
			if name == "narrowed" {
				entry += "/glm-5.3"
			}
			if name == "candidate-entry" {
				entry = glmEntry
			}
			exe, _ := os.Executable()
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, exe, "-test.run="+entry, "-test.timeout=50s")
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("exit", map[string]any{"success": err == nil, "output": string(output)})
			success := name == route || name == "missing-key"
			cs.Check("bounded formal entry", (err == nil) == success, "%v %s", err, output)
			files, _ := filepath.Glob(filepath.Join(base, "*", "formal-glm-report.json"))
			if !cs.Check("distinct formal detail", len(files) == 1, "files=%v", files) {
				return
			}
			raw, _ := os.ReadFile(files[0])
			var r glmReport
			json.Unmarshal(raw, &r)
			cs.Record("formal-detail", json.RawMessage(raw))
			hash, _ := ai.BuiltinCatalog().Hash()
			cs.Check("actual complete builtin hash", r.CatalogHash == hash && r.CatalogVersion == ai.BuiltinCatalog().Version, "catalog=%s", r.CatalogHash)
			cs.Check("formal identity separate", r.Kind == "glm-formal" && strings.HasPrefix(r.Route, route), "kind=%s", r.Kind)
			formal, _ := filepath.Glob(filepath.Join(base, "*", "live-report.json"))
			cs.Check("controlled and absent credentials never publish support", len(formal) == 0, "published support")
			// A caller cannot import a controlled detail by pointing LoadReport at it.
			_, e := supportmatrix.LoadReport(files[0])
			cs.Check("detail is not a support report", e != nil, "accepted candidate schema")
			copyErr := filepath.WalkDir(filepath.Dir(files[0]), func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(filepath.Dir(files[0]), path)
				target := filepath.Join("child", rel)
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
			cs.Check("raw child retained", copyErr == nil, "%v", copyErr)
		})
	}
}

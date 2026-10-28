//go:build live

package live

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGLMFinalBudgetSourceProcess(t *testing.T) {
	cs := run.Case(t, "REGIONAL-final-round-glm-source-closure")
	binary := filepath.Join(t.TempDir(), "glmformal")
	build := exec.Command("go", "build", "-o", binary, "./cmd/processgrant")
	output, err := build.CombinedOutput()
	cs.Record("build", map[string]any{"argv": build.Args, "success": err == nil, "output": string(output)})
	if !cs.Check("keyless public source exporter builds", err == nil, "%v %s", err, output) {
		return
	}
	root := t.TempDir()
	directory := filepath.Join(root, "ai/internal/testkit/finalbudget")
	os.MkdirAll(directory, 0700)
	os.WriteFile(filepath.Join(root, "go.mod"), []byte("module controlled\n"), 0600)
	os.WriteFile(filepath.Join(root, "go.sum"), nil, 0600)
	os.WriteFile(filepath.Join(directory, "guard.go"), []byte("package finalbudget\n// controlled original\n"), 0600)
	read := func(label string) map[string]string {
		cmd := exec.Command(binary, "-source", "glm")
		cmd.Dir = root
		for _, key := range []string{"PATH", "HOME", "TMPDIR"} {
			if value := os.Getenv(key); value != "" {
				cmd.Env = append(cmd.Env, key+"="+value)
			}
		}
		raw, err := cmd.CombinedOutput()
		cs.Record(label, map[string]any{"argv": cmd.Args, "success": err == nil, "output": string(raw)})
		if !cs.Check("public exporter accepts keyless source-only mode", err == nil, "%v %s", err, raw) {
			return nil
		}
		var value struct {
			Files map[string]string `json:"files"`
		}
		json.Unmarshal(raw, &value)
		return value.Files
	}
	initial := read("initial-source")
	name := "ai/internal/testkit/finalbudget/guard.go"
	cs.Check("actual budget helper belongs to complete GLM source", initial[name] != "", "file missing")
	os.WriteFile(filepath.Join(directory, "guard.go"), []byte("package finalbudget\n// changed request budget\n"), 0600)
	changed := read("changed-source")
	cs.Check("changed final USD helper expires old source", changed[name] != "" && changed[name] != initial[name], "helper drift omitted")
	os.WriteFile(filepath.Join(directory, "next.go"), []byte("package finalbudget\n"), 0600)
	added := read("added-source")
	cs.Check("new consumed budget files extend exact source set", added["ai/internal/testkit/finalbudget/next.go"] != "", "addition omitted")
	os.Remove(filepath.Join(directory, "guard.go"))
	deleted := read("deleted-source")
	_, present := deleted[name]
	cs.Check("deleted helper expires captured source set", !present, "deletion omitted")
}

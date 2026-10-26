package supportmatrix

import (
	"os"
	"path/filepath"
	"strings"
)

// GLMExecutionSource binds evidence to the code that executes and captures the
// public calls. Documentation, other live suites and release plumbing may evolve
// without pretending to rerun the measured model calls.
func GLMExecutionSource() (map[string]string, error) {
	root, e := os.Getwd()
	if e != nil {
		return nil, e
	}
	for {
		if _, e = os.Stat(filepath.Join(root, "go.mod")); e == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil, glmBad("source root missing")
		}
		root = parent
	}
	out := map[string]string{}
	e = filepath.WalkDir(filepath.Join(root, "ai"), func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".evidence" {
				return filepath.SkipDir
			}
			return nil
		}
		name, _ := filepath.Rel(root, path)
		name = filepath.ToSlash(name)
		runtime := filepath.Dir(name) == "ai" && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		capture := strings.HasPrefix(name, "ai/live/glm_") && strings.HasSuffix(name, ".go")
		for _, file := range []string{"ai/live/grok_process_grant_test.go", "ai/live/main_test.go", "ai/live/recorder_test.go", "ai/live/proxy_test.go", "ai/live/combos_test.go", "ai/live/decisions_config_test.go", "ai/live/decisions_sources_test.go", "ai/live/session_test.go"} {
			capture = capture || name == file
		}
		for _, prefix := range []string{"ai/examples/requestid/", "ai/internal/testkit/evidence/", "ai/internal/testkit/audit/", "ai/internal/testkit/glm/", "ai/internal/testkit/finalbudget/", "ai/internal/testkit/regional/", "ai/internal/testkit/supportmatrix/", "ai/internal/testkit/grokhistory/", "ai/internal/testkit/livegrant/"} {
			if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") {
				capture = true
			}
		}
		// The fixed round guard loads the independently pinned regional candidate
		// catalog and its embedded source bytes before pricing a request.
		capture = capture || strings.HasPrefix(name, "ai/internal/testkit/regional/")
		if !runtime && !capture {
			return nil
		}
		raw, e := glmRead(path)
		if e != nil {
			return e
		}
		out[name] = glmHash(raw)
		return nil
	})
	if e != nil {
		return nil, e
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, e := glmRead(filepath.Join(root, name))
		if e != nil {
			return nil, e
		}
		out[name] = glmHash(raw)
	}
	return out, nil
}

package release

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func sourceRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			return root, nil
		}
		parent := filepath.Dir(root)
		if parent == root {
			return "", errors.New("source root missing")
		}
		root = parent
	}
}

// A new gate/test does not change the model execution that an earlier report
// measured. Compare the complete execution-source set, including additions and
// deletions, while retaining the original full code snapshot/hash untouched.
// Runtime, evaluation CLI, its request IDs and testkit dependencies are included;
// unrelated example programs, live smoke, E2E and release tooling are excluded.
func evaluationSource(name string) bool {
	if name == "go.mod" || name == "go.sum" || name == "ai/examples/chineseeval/testdata/recorded-jev-20261008.json" {
		return true
	}
	if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false
	}
	for _, prefix := range []string{"ai/release/", "ai/e2e/", "ai/live/", "ai/internal/testkit/release/"} {
		if strings.HasPrefix(name, prefix) {
			return false
		}
	}
	if strings.HasPrefix(name, "ai/examples/") {
		return strings.HasPrefix(name, "ai/examples/chineseeval/") || strings.HasPrefix(name, "ai/examples/requestid/")
	}
	return strings.HasPrefix(name, "ai/")
}
func currentEvaluationSource(root, dir, want, provider, api string) error {
	bad := errors.New("execution source drift")
	raw, err := readArtifact(filepath.Join(dir, "code-snapshot.json"), 1<<20)
	if err != nil || chineseeval.Hash(raw) != want {
		return bad
	}
	var recorded map[string]string
	if json.Unmarshal(raw, &recorded) != nil {
		return bad
	}
	current := map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "ai"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if !evaluationSource(name) {
			return nil
		}
		raw, err := readArtifact(path, 2<<20)
		if err != nil {
			return bad
		}
		current[name] = chineseeval.Hash(raw)
		return nil
	})
	if err != nil {
		return bad
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, err := readArtifact(filepath.Join(root, name), 2<<20)
		if err != nil {
			return err
		}
		current[name] = chineseeval.Hash(raw)
	}
	for _, source := range []map[string]string{recorded, current} {
		for name := range source {
			if strings.HasPrefix(name, "ai/examples/chineseeval/") || strings.HasPrefix(name, "ai/examples/requestid/") {
				if recorded[name] != current[name] {
					return bad
				}
			}
		}
	}
	return evidence.VerifyRouteSource(recorded, current, provider, api)
}

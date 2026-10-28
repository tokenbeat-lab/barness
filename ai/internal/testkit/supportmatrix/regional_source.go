package supportmatrix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// RegionalExecutionSource freezes the complete runtime, independent regional
// facts, capture harness and its native consumers. Additions and removals are
// material; neither a commit label nor a refreshed timestamp repairs drift.
func RegionalExecutionSource() (map[string]string, error) {
	root, err := minimaxlive.Root()
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "ai"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".evidence" {
				return filepath.SkipDir
			}
			return nil
		}
		name, _ := filepath.Rel(root, path)
		name = filepath.ToSlash(name)
		// Unknown product subpackages must reach the conservative route guard.
		runtime := strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, "ai/internal/testkit/") && !strings.HasPrefix(name, "ai/examples/") && !strings.HasPrefix(name, "ai/e2e/") && !strings.HasPrefix(name, "ai/live/") && !strings.HasPrefix(name, "ai/release/")
		capture := filepath.Dir(name) == "ai/live" && strings.HasSuffix(name, ".go") || strings.HasPrefix(name, "ai/internal/testkit/regional/") || strings.HasPrefix(name, "ai/internal/testkit/kimi/") || strings.HasPrefix(name, "ai/internal/") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
		for _, exact := range []string{"ai/live/main_test.go", "ai/live/recorder_test.go", "ai/live/proxy_test.go", "ai/live/combos_test.go", "ai/live/stream_test.go", "ai/live/session_test.go"} {
			capture = capture || name == exact
		}
		for _, prefix := range []string{"ai/examples/requestid/", "ai/internal/testkit/evidence/", "ai/internal/testkit/audit/", "ai/internal/testkit/supportmatrix/"} {
			capture = capture || strings.HasPrefix(name, prefix) && !strings.HasSuffix(name, "_test.go")
		}
		capture = capture || name == "ai/catalog_kimi.json"
		if !runtime && !capture {
			return nil
		}
		raw, err := glmRead(path)
		if err != nil {
			return err
		}
		out[name] = glmHash(raw)
		return nil
	})
	if err != nil {
		return nil, err
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

func RegionalSourceHash(source map[string]string) string {
	raw, _ := json.Marshal(source)
	return glmHash(raw)
}

func RegionalSDK(combo string) (string, error) {
	route := "chat"
	if strings.HasSuffix(combo, "-messages") {
		route = "messages"
	}
	return minimaxlive.SDKVersion(minimaxlive.Identity{Route: route})
}

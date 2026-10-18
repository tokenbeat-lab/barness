package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExecutionSource recomputes the complete selected runtime and capture source
// set. Additions/deletions are material, independently of Git revision.
func ExecutionSource() (map[string]string, error) {
	root, e := Root()
	if e != nil {
		return nil, e
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
		capture := strings.HasPrefix(name, "ai/live/minimax_") && strings.HasSuffix(name, ".go")
		for _, file := range []string{"ai/live/main_test.go", "ai/live/recorder_test.go", "ai/live/proxy_test.go", "ai/live/combos_test.go", "ai/live/decisions_config_test.go", "ai/live/decisions_sources_test.go", "ai/live/session_test.go"} {
			capture = capture || name == file
		}
		for _, prefix := range []string{"ai/examples/requestid/", "ai/internal/testkit/minimax/", "ai/internal/testkit/minimaxlive/", "ai/internal/testkit/evidence/", "ai/internal/testkit/audit/", "ai/internal/testkit/supportmatrix/", "ai/live/testdata/minimax/"} {
			if strings.HasPrefix(name, prefix) && !strings.HasSuffix(name, "_test.go") {
				capture = true
			}
		}
		if !runtime && !capture {
			return nil
		}
		raw, e := ReadFile(path, 2<<20)
		if e != nil {
			return e
		}
		out[name] = Hash(raw)
		return nil
	})
	if e != nil {
		return nil, e
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, e := ReadFile(filepath.Join(root, name), 2<<20)
		if e != nil {
			return nil, e
		}
		out[name] = Hash(raw)
	}
	return out, nil
}
func Root() (string, error) {
	dir, e := os.Getwd()
	if e != nil {
		return "", e
	}
	for {
		if _, e = os.Stat(filepath.Join(dir, "go.mod")); e == nil {
			return dir, nil
		}
		next := filepath.Dir(dir)
		if next == dir {
			return "", fmt.Errorf("source root missing")
		}
		dir = next
	}
}

type Source struct {
	File   string `json:"file"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// SourcePin binds complete official snapshots, not a mutable URL or excerpt.
const SourcePin = "41dbf84331b516c45e180128634146aefd54b6daaedf1118f0628b355d8e0684"

func Sources() (map[string][]byte, error) {
	root, e := Root()
	if e != nil {
		return nil, e
	}
	base := filepath.Join(root, "ai/live/testdata/minimax")
	raw, e := ReadFile(filepath.Join(base, "sources.json"), 1<<20)
	if e != nil || Hash(raw) != SourcePin {
		return nil, fmt.Errorf("fixed MiniMax source manifest missing or changed")
	}
	var sources []Source
	if e = json.Unmarshal(raw, &sources); e != nil {
		return nil, e
	}
	out := map[string][]byte{"sources.json": raw}
	entries, e := os.ReadDir(base)
	if e != nil || len(entries) != len(sources)+1 {
		return nil, fmt.Errorf("unexpected fixed source set")
	}
	for _, s := range sources {
		if filepath.Base(s.File) != s.File {
			return nil, fmt.Errorf("unsafe fixed source path")
		}
		raw, e = ReadFile(filepath.Join(base, s.File), 2<<20)
		if e != nil || Hash(raw) != s.SHA256 {
			return nil, fmt.Errorf("fixed MiniMax snapshot missing or changed")
		}
		out[s.File] = raw
	}
	return out, nil
}

// SDKVersion is derived from the pinned dependency rather than report claims.
func SDKVersion(id Identity) (string, error) {
	root, e := Root()
	if e != nil {
		return "", e
	}
	raw, e := ReadFile(filepath.Join(root, "go.mod"), 2<<20)
	if e != nil {
		return "", e
	}
	module := "github.com/openai/openai-go/v3"
	if id.Route == "messages" {
		module = "github.com/anthropics/anthropic-sdk-go"
	}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "require "))
		if len(fields) >= 2 && fields[0] == module {
			suffix := ""
			if id.Route == "messages" {
				suffix = " (HTTP; SSE decoded by barness-ai, ADR-0011)"
			}
			return module + " " + fields[1] + suffix, nil
		}
	}
	return "", fmt.Errorf("MiniMax SDK dependency unavailable")
}

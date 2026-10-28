package supportmatrix

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func kimiRepositoryRoot() (string, error) { return minimaxlive.Root() }
func kimiLocalPath(root, name string) (string, error) {
	if !fs.ValidPath(name) || !filepath.IsLocal(name) {
		return "", refuse("Kimi input must be repository local")
	}
	p := root
	for _, part := range strings.Split(name, "/") {
		p = filepath.Join(p, part)
		info, e := os.Lstat(p)
		if e != nil || info.Mode()&os.ModeSymlink != 0 {
			return "", refuse("Kimi input missing or symlinked")
		}
	}
	return p, nil
}

type kimiProducer struct {
	Status   string   `json:"status"`
	Stage    string   `json:"stage"`
	ExitCode *int     `json:"exitCode"`
	Argv     []string `json:"argv"`
	Source   struct {
		Commit string            `json:"commit"`
		Files  map[string]string `json:"files"`
	} `json:"source"`
	StartedAt      time.Time                               `json:"startedAt"`
	FinishedAt     time.Time                               `json:"finishedAt"`
	Files          map[string]string                       `json:"files"`
	SelectedInputs map[string]string                       `json:"selectedInputs"`
	Bundles        []struct{ Path, ManifestSHA256 string } `json:"bundles"`
}

// Bind actual process exit/logs, retained executable and original execution
// source. A native report's own PASS flag is never an exit-code receipt.
func verifyKimiProducer(root, dir string, ref KimiFileRef, r Report) (map[string]string, error) {
	inputs := map[string]string{}
	path, e := kimiLocalPath(root, ref.Path)
	if e != nil {
		return nil, e
	}
	opened, e := os.OpenRoot(filepath.Dir(path))
	if e != nil {
		return nil, e
	}
	defer opened.Close()
	raw, e := regionalRead(opened, filepath.Base(path))
	if e != nil || !validManifestHash(ref.SHA256) || glmHash(raw) != ref.SHA256 {
		return nil, refuse("Kimi producer receipt changed")
	}
	var p kimiProducer
	// Producer receipts have tool-specific metadata; their owned fields are
	// decoded separately, unlike the strict two-reference composition descriptor.
	if regionalDecode(raw, &p) != nil || p.Status != "PASS" || p.ExitCode == nil || *p.ExitCode != 0 || len(p.Argv) < 2 || p.Source.Commit != r.GitCommit || p.StartedAt.IsZero() || p.StartedAt.After(r.StartedAt) || p.FinishedAt.Before(r.FinishedAt) {
		return nil, refuse("Kimi actual producer exit/source/time unavailable")
	}
	if r.Regional.Mode == "live" && p.Stage != "live" || r.Regional.Mode == "controlled" && p.Stage != "e2e" {
		return nil, refuse("Kimi producer stage contradicts actual mode")
	}
	if !kimiProducerArgs(p.Argv, r) {
		return nil, refuse("Kimi producer entry, count or timeout differs")
	}
	inputs[ref.Path] = ref.SHA256
	for _, name := range []string{"stdout.jsonl", "stderr.log", "exit-code"} {
		h := p.Files[name]
		b, e := regionalRead(opened, name)
		if e != nil || !validManifestHash(h) || glmHash(b) != h || name == "exit-code" && strings.TrimSpace(string(b)) != "0" {
			return nil, refuse("Kimi producer streams or numeric exit changed")
		}
		rel, _ := filepath.Rel(root, filepath.Join(filepath.Dir(path), name))
		inputs[filepath.ToSlash(rel)] = h
	}
	matched := false
	for _, b := range p.Bundles {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(b.Path)))
		if candidate == filepath.Clean(dir) && b.ManifestSHA256 != "" {
			m, e := regionalReadFile(filepath.Join(dir, "manifest.json"))
			if e == nil && glmHash(m) == b.ManifestSHA256 {
				matched = true
			}
		}
	}
	if !matched {
		return nil, refuse("Kimi producer did not capture selected bundle")
	}
	var execution map[string]string
	if readJSON(filepath.Join(dir, "regional-execution-source.json"), &execution) != nil {
		return nil, refuse("Kimi producer source map unavailable")
	}
	for n, h := range execution {
		compiled, inSource := p.Source.Files[n]
		selected, inInputs := p.SelectedInputs[n]
		// run.py records Markdown and other explicit inputs separately from
		// source.files. Both domains must agree when an input appears in both.
		if !inSource && !inInputs || inSource && compiled != h || inInputs && selected != h {
			return nil, refuse("Kimi captured execution source disagrees with producer")
		}
	}
	binary := p.Argv[0]
	if filepath.IsAbs(binary) {
		binary, e = filepath.Rel(root, binary)
		if e != nil {
			return nil, e
		}
	}
	binary = filepath.ToSlash(binary)
	binaryPath, e := kimiLocalPath(root, binary)
	if e != nil {
		return nil, e
	}
	h, e := livegrant.FileHash(binaryPath)
	if e != nil || h != r.Regional.BinarySHA256 {
		return nil, refuse("Kimi retained producer executable differs")
	}
	inputs[binary] = h
	return inputs, nil
}

func kimiProducerArgs(argv []string, r Report) bool {
	// Preserve the real old-nine outer command (31m); its native report still
	// enforces the independently measured 1800s process budget.
	entry, timeout := "^TestRegionalCandidate$", 31*time.Minute
	if r.Regional.Purpose == KimiHistoryPurpose {
		entry, timeout = "^TestKimiHistorySupplement$", 240*time.Second
	} else if r.Regional.Mode == "controlled" {
		entry, timeout = "^TestKimiBaselineControlled$", 1800*time.Second
	}
	seen := map[string]bool{}
	for _, arg := range argv[1:] {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || seen[key] {
			return false
		}
		seen[key] = true
		switch key {
		case "-test.run":
			if value != entry {
				return false
			}
		case "-test.timeout":
			d, e := time.ParseDuration(value)
			if e != nil || d != timeout {
				return false
			}
		case "-test.count":
			if value != "1" {
				return false
			}
		case "-test.v":
			if value != "true" && value != "false" {
				return false
			}
		default:
			return false
		}
	}
	return seen["-test.run"] && seen["-test.timeout"]
}
func regionalReadFile(path string) ([]byte, error) {
	r, e := os.OpenRoot(filepath.Dir(path))
	if e != nil {
		return nil, e
	}
	defer r.Close()
	return regionalRead(r, filepath.Base(path))
}

package minimaxlive

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
)

type ThinkingFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type thinkingProducer struct {
	Schema        int
	Route         string
	CommandSHA256 string
	Status, Stage string
	ExitCode      *int
	Argv          []string
	Source        struct {
		Commit string
		Files  map[string]string
	}
	SelectedInputs        map[string]string
	Files                 map[string]string
	StartedAt, FinishedAt time.Time
	Bundles               []struct{ Path, ManifestSHA256 string }
}

// LocalEvidencePath is shared file admission, not a cross-provider composition
// engine. All ancestors are checked; descriptors cannot select authority paths.
func LocalEvidencePath(root, name string) (string, error) {
	if !fs.ValidPath(name) || !filepath.IsLocal(name) {
		return "", fmt.Errorf("MiniMax evidence must be repository local")
	}
	p := root
	for _, part := range strings.Split(name, "/") {
		p = filepath.Join(p, part)
		s, e := os.Lstat(p)
		if e != nil || s.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("MiniMax original missing or symlinked")
		}
	}
	return p, nil
}

// VerifyThinkingProducer requires the actual failed baseline exit, not a
// fabricated partial-PASS measurement. Its returned map includes every retained
// byte (or exact original Git blob) from both source and selected input domains.
func VerifyThinkingProducer(dir string, ref ThinkingFileRef, r Report, baseline bool) (map[string]string, error) {
	dir, e := filepath.Abs(dir)
	if e != nil {
		return nil, e
	}
	root, e := Root()
	if e != nil {
		return nil, e
	}
	path, e := LocalEvidencePath(root, ref.Path)
	if e != nil {
		return nil, e
	}
	raw, e := ReadFile(path, 4<<20)
	if e != nil || Hash(raw) != ref.SHA256 {
		return nil, fmt.Errorf("MiniMax original producer receipt changed")
	}
	var p thinkingProducer
	if json.Unmarshal(raw, &p) != nil {
		return nil, fmt.Errorf("MiniMax producer receipt invalid")
	}
	exit, status := 0, "PASS"
	if baseline {
		exit, status = 1, "FAIL"
	}
	if p.Schema != 1 || p.Route != "minimax-messages" || p.ExitCode == nil || *p.ExitCode != exit || p.Status != status || p.Source.Commit != r.GitCommit || p.StartedAt.IsZero() || p.StartedAt.After(r.StartedAt) || p.FinishedAt.Before(r.FinishedAt) || r.Mode == "live" && p.Stage != "live" || r.Mode == "controlled" && p.Stage != "e2e" {
		return nil, fmt.Errorf("MiniMax actual producer exit/source/time differs")
	}
	binary, e := thinkingProducerArgs(root, p.Argv, r, baseline)
	if e != nil {
		return nil, e
	}
	inputs := map[string]string{ref.Path: ref.SHA256}
	for _, name := range []string{"stdout.jsonl", "stderr.log", "exit-code", "command.json"} {
		full, e := LocalEvidencePath(root, filepath.ToSlash(filepath.Join(filepath.Dir(ref.Path), name)))
		if e != nil {
			return nil, e
		}
		b, e := ReadFile(full, 64<<20)
		h := p.Files[name]
		if e != nil || h == "" || Hash(b) != h || name == "exit-code" && strings.TrimSpace(string(b)) != strconv.Itoa(exit) {
			return nil, fmt.Errorf("MiniMax producer streams/numeric exit changed")
		}
		rel, _ := filepath.Rel(root, full)
		inputs[filepath.ToSlash(rel)] = h
	}

	if p.CommandSHA256 == "" || p.Files["command.json"] != p.CommandSHA256 {
		return nil, fmt.Errorf("MiniMax producer command pin absent")
	}
	commandRaw, e := ReadFile(filepath.Join(filepath.Dir(path), "command.json"), 1<<20)
	var command struct {
		Schema       int
		Argv         []string
		Stage, Route string
	}
	if e != nil || json.Unmarshal(commandRaw, &command) != nil || command.Schema != 1 || !reflect.DeepEqual(command.Argv, p.Argv) || command.Stage != p.Stage || command.Route != p.Route {
		return nil, fmt.Errorf("MiniMax producer command contradicts receipt")
	}
	matched := 0
	for _, b := range p.Bundles {
		candidate := filepath.Clean(filepath.Join(filepath.Dir(path), filepath.FromSlash(b.Path)))
		if candidate == filepath.Clean(dir) {
			raw, e := ReadFile(filepath.Join(dir, "manifest.json"), 2<<20)
			if e != nil || Hash(raw) != b.ManifestSHA256 {
				return nil, fmt.Errorf("MiniMax producer manifest differs")
			}
			matched++
		}
	}
	// run.py intentionally inventories only successful native bundles. A
	// failed baseline therefore has zero bundle rows, but its real native TestMain
	// writes the exact retained directory to stderr. Bind that original layout,
	// grant/session, argv and sealed descriptor instead of inventing a PASS row.
	if matched == 0 && baseline && len(p.Bundles) == 0 {
		stderr, e := ReadFile(filepath.Join(filepath.Dir(path), "stderr.log"), 64<<20)
		expectedA := filepath.Join(filepath.Dir(path), "bundles", r.Session)
		expectedB := filepath.Join(filepath.Dir(path), r.Session)
		if e == nil && (dir == expectedA || dir == expectedB) && filepath.Base(dir) == r.Session {
			marker := "barness-ai evidence bundle: " + dir
			for _, line := range strings.Split(string(stderr), "\n") {
				if line == marker {
					matched++
				}
			}
		}
	}
	if matched != 1 {
		return nil, fmt.Errorf("MiniMax producer selected bundle absent or duplicated")
	}
	for n, h := range p.SelectedInputs {
		if old, ok := p.Source.Files[n]; ok && old != h {
			return nil, fmt.Errorf("MiniMax producer source and selectedInputs conflict")
		}
	}
	for n, h := range r.ExecutionSource {
		a, as := p.Source.Files[n]
		b, bs := p.SelectedInputs[n]
		if !as && !bs || as && a != h || bs && b != h {
			return nil, fmt.Errorf("MiniMax producer source/selectedInputs mismatch: %s", n)
		}
	}
	// Prove the complete original domains. A current filename with different
	// bytes cannot stand in for a missing historical source or input.
	for _, group := range []map[string]string{p.Source.Files, p.SelectedInputs} {
		for n, h := range group {
			key, e := thinkingOriginalInput(root, p.Source.Commit, n, h)
			if e != nil {
				return nil, e
			}
			if old, ok := inputs[key]; ok && old != h {
				return nil, fmt.Errorf("MiniMax producer input conflicts")
			}
			inputs[key] = h
		}
	}
	g, _, _, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return nil, e
	}
	h, e := livegrant.FileHash(binary)
	if e != nil || h != g.Scope.BinarySHA256 {
		return nil, fmt.Errorf("MiniMax retained producer executable differs")
	}
	rel, _ := filepath.Rel(root, binary)
	inputs[filepath.ToSlash(rel)] = h
	return inputs, nil
}
func thinkingOriginalInput(root, commit, name, hash string) (string, error) {
	if !fs.ValidPath(name) || !filepath.IsLocal(name) || len(hash) != 64 || len(commit) != 40 {
		return "", fmt.Errorf("MiniMax original input identity invalid")
	}
	p, e := LocalEvidencePath(root, name)
	if e == nil {
		b, e := ReadFile(p, 64<<20)
		if e == nil && Hash(b) == hash {
			return "@retained/" + name, nil
		}
	}
	cmd := exec.Command("git", "show", commit+":"+name)
	cmd.Dir = root
	b, e := cmd.Output()
	if e != nil || Hash(b) != hash {
		return "", fmt.Errorf("MiniMax original input bytes unavailable: %s", name)
	}
	return "@git/" + commit + "/" + name, nil
}
func thinkingProducerArgs(root string, argv []string, r Report, baseline bool) (string, error) {
	if len(argv) < 2 {
		return "", fmt.Errorf("MiniMax producer argv missing")
	}
	if argv[0] == "/usr/bin/env" {
		if len(argv) < 5 || argv[1] != "-C" || !filepath.IsAbs(argv[2]) {
			return "", fmt.Errorf("MiniMax producer launcher differs")
		}
		argv = argv[3:]
	}
	binary := argv[0]
	if filepath.IsAbs(binary) {
		p, e := filepath.Rel(root, binary)
		if e != nil {
			return "", e
		}
		binary = filepath.ToSlash(p)
	}
	full, e := LocalEvidencePath(root, binary)
	if e != nil {
		return "", e
	}
	entry, timeout := "^TestMiniMaxThinkingOnSupplement$", 180*time.Second
	if baseline {
		entry, timeout = "^TestMiniMaxCandidate$", 1800*time.Second
		if r.Stage == "formal" {
			entry = "^TestMiniMaxFormal$"
		}
	}
	seen := map[string]bool{}
	for _, arg := range argv[1:] {
		k, v, ok := strings.Cut(arg, "=")
		if !ok || seen[k] {
			return "", fmt.Errorf("MiniMax duplicate/unknown process flag")
		}
		seen[k] = true
		switch k {
		case "-test.run":
			if v != entry {
				return "", fmt.Errorf("MiniMax producer entry differs")
			}
		case "-test.timeout":
			d, e := time.ParseDuration(v)
			if e != nil || d != timeout {
				return "", fmt.Errorf("MiniMax producer timeout differs")
			}
		case "-test.count":
			if v != "1" {
				return "", fmt.Errorf("MiniMax producer count differs")
			}
		case "-test.v":
			if v != "true" && v != "false" {
				return "", fmt.Errorf("MiniMax producer verbosity invalid")
			}
		default:
			return "", fmt.Errorf("MiniMax producer flags differ")
		}
	}
	if !seen["-test.run"] || !seen["-test.timeout"] || !baseline && !seen["-test.count"] {
		return "", fmt.Errorf("MiniMax producer exact selection absent")
	}
	return full, nil
}

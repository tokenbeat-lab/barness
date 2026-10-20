package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
)

const checkpointLimit = 16 << 20
const verificationFileLimit = 256 << 20

type frozenInputs struct {
	Sources map[string]string `json:"sources"`
	Files   map[string]string `json:"files"`
}
type checkState struct {
	Schema     int               `json:"schema"`
	RootSHA256 string            `json:"rootSHA256"`
	Inputs     options           `json:"inputs"`
	Frozen     frozenInputs      `json:"frozen"`
	Bundle     string            `json:"bundle"`
	Metadata   map[string]string `json:"metadata"`
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// Checkpoint metadata and source files have their own finite regular-file
// boundary. Native bundles continue to use their original stricter consumers.
func openRegular(path string, limit int64) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil || !before.Mode().IsRegular() || before.Size() > limit {
		return nil, errors.New("verification file missing, non-regular or oversized")
	}
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errors.New("verification file unavailable")
	}
	opened, e1 := f.Stat()
	after, e2 := os.Lstat(path)
	if e1 != nil || e2 != nil || !opened.Mode().IsRegular() || !after.Mode().IsRegular() || opened.Size() > limit || !os.SameFile(before, opened) || !os.SameFile(before, after) {
		f.Close()
		return nil, errors.New("verification file changed")
	}
	return f, nil
}
func readRegular(path string, limit int64) ([]byte, error) {
	f, err := openRegular(path, limit)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err != nil || int64(len(raw)) > limit {
		return nil, errors.New("verification file unreadable or oversized")
	}
	return raw, nil
}
func fileHash(path string) (string, int64, error) {
	f, err := openRegular(path, verificationFileLimit)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	before, _ := f.Stat()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, verificationFileLimit+1))
	after, e := f.Stat()
	if err != nil || e != nil || n > verificationFileLimit || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "", 0, errors.New("verification file changed or exceeded limit")
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}
func captureCheckpointInputs(root string, o options) (frozenInputs, error) {
	f := frozenInputs{Sources: map[string]string{}, Files: map[string]string{}}
	addTree := func(dir string, selected func(string) bool, dst map[string]string, sourceTree bool) error {
		var total int64
		count := 0
		return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return errors.New("verification source/input unavailable")
			}
			if d.IsDir() {
				if sourceTree && (d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "__pycache__" || d.Name() == "session-analysis") {
					return filepath.SkipDir
				}
				return nil
			}
			if !selected(path) {
				return nil
			}
			count++
			if count > 250000 {
				return errors.New("verification inventory over file limit")
			}
			hash, n, err := fileHash(path)
			if err != nil {
				return err
			}
			total += n
			if total > 8<<30 {
				return errors.New("verification inventory over byte limit")
			}
			name, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			dst[filepath.ToSlash(name)] = hash
			return nil
		})
	}
	isReadme := func(path string) bool {
		name := filepath.Base(path)
		return name == "README.md" || name == "README.zh-CN.md"
	}
	source := func(path string) bool { return !isReadme(path) }
	if err := addTree(filepath.Join(root, "ai"), source, f.Sources, true); err != nil {
		return f, err
	}
	for _, name := range []string{"go.mod", "go.sum", "go.work", "go.work.sum"} {
		path := filepath.Join(root, name)
		if _, err := os.Lstat(path); os.IsNotExist(err) && name != "go.mod" && name != "go.sum" {
			continue
		}
		h, _, err := fileHash(path)
		if err != nil {
			return f, err
		}
		f.Sources[name] = h
	}
	scratch := filepath.Join(root, ".scratch")
	if _, err := os.Stat(scratch); err == nil {
		scripts := func(path string) bool {
			ext := filepath.Ext(path)
			return ext == ".go" || ext == ".py" || ext == ".sh" || filepath.Base(path) == "rollback-probe.txt" || strings.Contains(filepath.ToSlash(path), "/tools/") && !isReadme(path)
		}
		if err := addTree(scratch, scripts, f.Sources, true); err != nil {
			return f, err
		}
	}
	for _, name := range []string{"go", "node", "python3"} {
		tool, err := exec.LookPath(name)
		if err != nil {
			if name == "go" {
				return f, errors.New("Go tool unavailable")
			}
			f.Sources["@tool-"+name] = "unavailable"
			continue
		}
		tool, err = filepath.EvalSymlinks(tool)
		if err != nil {
			return f, errors.New("tool identity unavailable")
		}
		h, _, err := fileHash(tool)
		if err != nil {
			return f, err
		}
		f.Sources["@tool-"+name] = h
	}
	f.Sources["@runtime"] = digest([]byte(runtime.Version() + "/" + runtime.GOOS + "/" + runtime.GOARCH))
	f.Sources["@build-environment"] = digest([]byte(strings.Join(buildEnvironment(), "\n")))
	for _, path := range []string{o.Matrix, o.Extensions, o.Trace, o.Ledger, o.Snapshot} {
		if path != "" {
			if !filepath.IsAbs(path) {
				path = filepath.Join(root, path)
			}
			h, _, err := fileHash(path)
			if err != nil {
				return f, err
			}
			f.Files[path] = h
		}
	}
	for _, dir := range append(append([]string{}, o.Live...), o.Jev, o.Luna, o.VisualJev, o.VisualDecisions) {
		if dir == "" {
			continue
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(root, dir)
		}
		files := map[string]string{}
		if err := addTree(dir, func(string) bool { return true }, files, false); err != nil {
			return f, err
		}
		for name, h := range files {
			f.Files[filepath.Join(root, filepath.FromSlash(name))] = h
		}
	}
	return f, nil
}
func sameFrozen(a, b frozenInputs) bool {
	return maps.Equal(a.Sources, b.Sources) && maps.Equal(a.Files, b.Files)
}
func verifyFrozen(root string, o options, want frozenInputs) error {
	now, err := captureCheckpointInputs(root, o)
	if err != nil || !sameFrozen(want, now) {
		return errors.New("source, build tool, tests or selected inputs changed")
	}
	return nil
}
func rootIdentity(root string) string { p, _ := filepath.Abs(root); return digest([]byte(p)) }
func saveCheckpoint(root, out string, o options, f frozenInputs, bundle string) error {
	rel, err := filepath.Rel(out, bundle)
	if err != nil || !fs.ValidPath(filepath.ToSlash(rel)) || !strings.HasPrefix(filepath.ToSlash(rel), "offline/") {
		return errors.New("ordinary bundle is outside check output")
	}
	o.Out = ""
	o.From = ""
	c := checkState{Schema: 1, RootSHA256: rootIdentity(root), Inputs: o, Frozen: f, Bundle: filepath.ToSlash(rel), Metadata: map[string]string{}}
	names := checkpointMetadataNames(c.Bundle)
	for _, name := range names {
		h, _, err := fileHash(filepath.Join(out, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		c.Metadata[name] = h
	}
	return writeJSON(out, "check-state.json", c)
}
func loadCheckpoint(root, from string) (checkState, executionReport, error) {
	var c checkState
	var r executionReport
	bad := errors.New("check checkpoint or actual command receipts invalid")
	raw, err := readRegular(filepath.Join(from, "check-state.json"), checkpointLimit)
	if err != nil || json.Unmarshal(raw, &c) != nil {
		return c, r, bad
	}
	if c.Schema != 1 || c.RootSHA256 != rootIdentity(root) || c.Inputs.Stage != "check" || c.Inputs.Preflight || c.Inputs.WriteSnapshot || c.Inputs.Bundle != "" || c.Inputs.RaceBundle != "" || c.Inputs.From != "" || c.Inputs.Out != "" || len(c.Frozen.Sources) == 0 || len(c.Frozen.Files) == 0 || !fs.ValidPath(c.Bundle) || !strings.HasPrefix(c.Bundle, "offline/") {
		return c, r, bad
	}
	names := checkpointMetadataNames(c.Bundle)
	if len(c.Metadata) != len(names) {
		return c, r, bad
	}
	for _, name := range names {
		h, _, err := fileHash(filepath.Join(from, filepath.FromSlash(name)))
		if err != nil || c.Metadata[name] != h {
			return c, r, bad
		}
	}
	raw, err = readRegular(filepath.Join(from, "execution-report.json"), checkpointLimit)
	if err != nil || json.Unmarshal(raw, &r) != nil {
		return c, r, bad
	}
	if r.Schema != 1 || r.Stage != "check" || r.Status != "READY_FOR_RACE" || r.Race != "NOT_RUN" || r.ReleasePass || r.BlockedBy != "" || r.Error != "" || len(r.Steps) != 5 {
		return c, r, bad
	}
	for i, id := range []string{"go-vet", "go-vet-live", "go-test", "ordinary-evidence", "go-test-race"} {
		s := r.Steps[i]
		if s.ID != id {
			return c, r, bad
		}
		if i == 4 {
			if s.Status != "NOT_RUN" || s.ExitCode != nil || len(s.Argv) != 0 {
				return c, r, bad
			}
			continue
		}
		if s.Status != "PASS" || s.Error != "" || s.StartedAt.IsZero() || s.FinishedAt.Before(s.StartedAt) || s.Seconds < 0 {
			return c, r, bad
		}
		if i < 3 && (!slices.Equal(s.Argv, commandArgs(id)) || s.ExitCode == nil || *s.ExitCode != 0 || s.Log != id+".log") {
			return c, r, bad
		}
		if i == 2 && s.Events != "go-test.events.jsonl" {
			return c, r, bad
		}
	}
	if err := verifyFrozen(root, c.Inputs, c.Frozen); err != nil {
		return c, r, err
	}
	return c, r, nil
}

func checkpointMetadataNames(bundle string) []string {
	return []string{"execution-report.json", "go-vet.log", "go-vet-live.log", "go-test.log", "go-test.events.jsonl", bundle + "/manifest.json", bundle + "/audit.json"}
}

func checkpointsOverlap(a, b string) bool {
	a, _ = filepath.Abs(a)
	b, _ = filepath.Abs(b)
	rel, err := filepath.Rel(a, b)
	return err == nil && (rel == "." || rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func checkpointError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("check continuation refused: %w", err)
}

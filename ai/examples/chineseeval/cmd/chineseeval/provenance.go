package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
)

func readBounded(path string, max int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, max+1))
	if err != nil || int64(len(raw)) > max {
		return nil, errors.New("oversized input")
	}
	return raw, nil
}

// Hash the source contents independently of Git HEAD, which may have local
// changes. Report dates, commit, compiler and this hash have distinct meanings.
func codeSnapshot() (map[string]string, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil, errors.New("source root missing")
		}
		root = parent
	}
	sources := map[string]string{}
	err = filepath.WalkDir(filepath.Join(root, "ai"), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".go" && filepath.ToSlash(path) != filepath.ToSlash(filepath.Join(root, "ai/examples/chineseeval/testdata/recorded-jev-20261008.json")) {
			return nil
		}
		raw, err := readBounded(path, 2<<20)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sources[filepath.ToSlash(rel)] = chineseeval.Hash(raw)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, err := readBounded(filepath.Join(root, name), 2<<20)
		if err != nil {
			return nil, err
		}
		sources[name] = chineseeval.Hash(raw)
	}
	return sources, nil
}
func freezeFiles(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	files := map[string]string{}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "files.json" {
			return errors.New("unexpected bundle entry")
		}
		raw, err := readBounded(filepath.Join(dir, entry.Name()), 16<<20)
		if err != nil {
			return err
		}
		files[entry.Name()] = chineseeval.Hash(raw)
	}
	raw, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "files.json"), raw, 0600)
}
func verifyFiles(dir string) error {
	raw, err := readBounded(filepath.Join(dir, "files.json"), 1<<20)
	if err != nil {
		return err
	}
	var files map[string]string
	if json.Unmarshal(raw, &files) != nil || len(files) < 7 || len(files) > 16 {
		return errors.New("invalid file inventory")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != len(files)+1 {
		return errors.New("incomplete inventory")
	}
	for _, e := range entries {
		if e.Name() == "files.json" {
			continue
		}
		if e.IsDir() {
			return errors.New("unexpected bundle directory")
		}
		raw, err := readBounded(filepath.Join(dir, e.Name()), 16<<20)
		if err != nil || chineseeval.Hash(raw) != files[e.Name()] {
			return errors.New("artifact hash mismatch")
		}
	}
	return nil
}
func verifyCode(dir, want string) error {
	raw, err := readBounded(filepath.Join(dir, "code-snapshot.json"), 1<<20)
	if err != nil || chineseeval.Hash(raw) != want {
		return errors.New("source snapshot hash mismatch")
	}
	var sources map[string]string
	if json.Unmarshal(raw, &sources) != nil || len(sources) < 3 || sources["go.mod"] == "" || sources["go.sum"] == "" {
		return errors.New("incomplete source snapshot")
	}
	hash := regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
	for name, value := range sources {
		if filepath.IsAbs(name) || !filepath.IsLocal(name) || !hash.MatchString(value) {
			return errors.New("invalid source snapshot")
		}
	}
	return nil
}

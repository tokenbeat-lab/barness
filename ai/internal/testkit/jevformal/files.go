// Package jevformal verifies the fixed Jev formal evidence boundary. Protocol
// and quality keep their own budgets, captures and independent public verifiers.
package jevformal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

func Hash(raw []byte) string { sum := sha256.Sum256(raw); return hex.EncodeToString(sum[:]) }

// Read bounds every artifact and rejects symlinks in its complete path.
func Read(path string) ([]byte, error) {
	clean, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	for p := clean; ; p = filepath.Dir(p) {
		info, e := os.Lstat(p)
		if e != nil {
			return nil, e
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("symlink in evidence path")
		}
		if p == filepath.Dir(p) {
			break
		}
	}
	info, err := os.Lstat(clean)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > 16<<20 {
		return nil, errors.New("nonregular or oversized artifact")
	}
	f, err := os.Open(clean)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, (16<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 16<<20 {
		return nil, errors.New("artifact exceeds finite limit")
	}
	return raw, nil
}

func inventory(root string) (map[string]string, error) {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("invalid bundle root")
	}
	files := map[string]string{}
	total := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return errors.New("symlink in bundle")
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "formal-integrity.json" {
			return nil
		}
		raw, err := Read(path)
		if err != nil {
			return err
		}
		total += len(raw)
		if len(files) >= 4096 || total > 64<<20 {
			return errors.New("evidence inventory exceeds finite limit")
		}
		files[filepath.ToSlash(rel)] = Hash(raw)
		return nil
	})
	return files, err
}

// Copy retains raw subpackages unchanged. The destination must be new.
func Copy(source, dest string) error {
	files, err := inventory(source)
	if err != nil {
		return err
	}
	if err = os.Mkdir(dest, 0700); err != nil {
		return err
	}
	for name := range files {
		raw, err := Read(filepath.Join(source, filepath.FromSlash(name)))
		if err != nil {
			return err
		}
		path := filepath.Join(dest, filepath.FromSlash(name))
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			return err
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			return err
		}
	}
	return nil
}

func write(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0600)
}

// Seal is called only after both original subpackages and the final report exist.
func Seal(root string) error {
	if err := write(filepath.Join(root, "manifest.json"), map[string]any{"schema": 1, "kind": "jev-formal-support"}); err != nil {
		return err
	}
	findings, err := audit.Bundle(root, audit.Environment())
	if err != nil || len(findings) > 0 {
		return errors.New("formal evidence redaction audit failed")
	}
	if err = write(filepath.Join(root, "audit.json"), map[string]any{"findings": []audit.Finding{}}); err != nil {
		return err
	}
	files, err := inventory(root)
	if err != nil {
		return err
	}
	return write(filepath.Join(root, "formal-integrity.json"), files)
}

func CheckInventory(root string) error {
	raw, err := Read(filepath.Join(root, "formal-integrity.json"))
	if err != nil {
		return err
	}
	var want map[string]string
	if json.Unmarshal(raw, &want) != nil {
		return errors.New("invalid formal inventory")
	}
	got, err := inventory(root)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(want, got) {
		return errors.New("formal inventory mismatch")
	}
	raw, err = Read(filepath.Join(root, "manifest.json"))
	if err != nil {
		return err
	}
	var manifest struct {
		Schema int    `json:"schema"`
		Kind   string `json:"kind"`
	}
	if json.Unmarshal(raw, &manifest) != nil || manifest.Schema != 1 || manifest.Kind != "jev-formal-support" {
		return errors.New("invalid formal manifest identity")
	}
	for name := range got {
		if name != "live-report.json" && name != "audit.json" && name != "manifest.json" && !startsWithPackage(name) {
			return errors.New("unrecognized formal artifact")
		}
	}
	raw, err = Read(filepath.Join(root, "audit.json"))
	if err != nil {
		return err
	}
	var receipt struct {
		Findings []audit.Finding `json:"findings"`
	}
	if json.Unmarshal(raw, &receipt) != nil || receipt.Findings == nil || len(receipt.Findings) > 0 {
		return errors.New("invalid formal audit receipt")
	}
	findings, err := audit.Bundle(root, audit.Environment())
	if err != nil || len(findings) > 0 {
		return errors.New("formal independent audit failed")
	}
	return nil
}

func startsWithPackage(name string) bool {
	return strings.HasPrefix(name, "protocol/") || strings.HasPrefix(name, "effects/")
}

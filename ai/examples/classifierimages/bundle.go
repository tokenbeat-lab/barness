package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

type bundle struct {
	dir    string
	secret string
	hashes map[string]string
	cases  []string
	mode   string
}

func newBundle(dir, secret, mode string) (*bundle, error) {
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return nil, err
	}
	// Never overwrite an earlier run, including a partial/failed one.
	if err := os.Mkdir(dir, 0700); err != nil {
		return nil, errors.New("output directory must be new")
	}
	return &bundle{dir, secret, map[string]string{}, nil, mode}, nil
}
func (b *bundle) file(name string, raw []byte) error {
	if b.secret != "" {
		raw = bytes.ReplaceAll(raw, []byte(b.secret), []byte("[REDACTED]"))
	}
	path := filepath.Join(b.dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		return err
	}
	sum := sha256.Sum256(raw)
	b.hashes[filepath.ToSlash(name)] = hex.EncodeToString(sum[:])
	return nil
}
func (b *bundle) record(name string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return b.file(name, append(raw, '\n'))
}
func (b *bundle) finish() error {
	real := "NOT_RUN"
	if b.mode == "live" {
		real = "OBSERVED_NOT_EVALUATED"
	}
	manifest := func() map[string]any {
		// manifest.json is a root receipt, not its own hash entry.
		artifacts := map[string]string{}
		for name, hash := range b.hashes {
			if name != "manifest.json" {
				artifacts[name] = hash
			}
		}
		sort.Strings(b.cases)
		return map[string]any{"schema": 1, "mode": b.mode, "realService": real, "visualQuality": "NOT_RUN", "goVersion": runtime.Version(), "createdAt": time.Now().UTC(), "cases": b.cases, "artifacts": artifacts, "replay": "go run ./ai/examples/classifierimages (see each case/config.json for explicit input and Host configuration)"}
	}
	if err := b.record("manifest.json", manifest()); err != nil {
		return err
	}
	forbidden := append(audit.Environment(), audit.Forbidden{Value: b.secret, Label: "explicit example credential"})
	findings, err := audit.Bundle(b.dir, forbidden)
	if err != nil {
		return err
	}
	if findings == nil {
		findings = []audit.Finding{}
	}
	if err := b.record("audit.json", map[string]any{"findings": findings}); err != nil {
		return err
	}
	if err := b.record("manifest.json", manifest()); err != nil {
		return err
	}
	if len(findings) > 0 {
		return errors.New("evidence redaction audit failed")
	}
	return nil
}

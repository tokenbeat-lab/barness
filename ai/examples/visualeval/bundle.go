package main

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

type bundle struct {
	dir, secret string
	files       map[string]string
}
type manifest struct {
	Schema    int               `json:"schema"`
	GoVersion string            `json:"goVersion"`
	Artifacts map[string]string `json:"artifacts"`
}

func newBundle(dir, secret string) (*bundle, error) {
	if err := os.MkdirAll(filepath.Dir(dir), 0700); err != nil {
		return nil, err
	}
	if err := os.Mkdir(dir, 0700); err != nil {
		return nil, errors.New("output directory must be new")
	}
	return &bundle{dir, secret, map[string]string{}}, nil
}
func readBounded(path string, limit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("nonregular or oversized evidence input")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, limit+1))
	if int64(len(raw)) > limit {
		return nil, errors.New("evidence input exceeds limit")
	}
	return raw, err
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
	b.files[name] = digest(raw)
	return nil
}
func (b *bundle) record(name string, v any) error {
	if w, ok := v.(wire); ok && b.secret != "" {
		w.Request = bytes.ReplaceAll(w.Request, []byte(b.secret), []byte("[REDACTED]"))
		w.Response = bytes.ReplaceAll(w.Response, []byte(b.secret), []byte("[REDACTED]"))
		v = w
	}
	return b.file(name, encoded(v))
}
func sourceFiles() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir("ai", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			raw, e := readBounded(path, 1<<20)
			if e != nil {
				return e
			}
			files[filepath.ToSlash(path)] = raw
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		raw, e := readBounded(name, 1<<20)
		if e != nil {
			return nil, e
		}
		files[name] = raw
	}
	return files, nil
}
func (b *bundle) inputs(p *plan, a arguments, auth *authorization, r *report) error {
	for name := range p.pins {
		raw, _ := data.ReadFile("testdata/" + name)
		if err := b.file("inputs/"+name, raw); err != nil {
			return err
		}
	}
	raw, _ := data.ReadFile("testdata/pins.json")
	if err := b.file("inputs/pins.json", raw); err != nil {
		return err
	}
	if p.formal {
		if err := b.record("formal-catalog.json", p.catalog); err != nil {
			return err
		}
	}
	if err := b.record("run-config.json", map[string]any{"mode": a.mode, "route": a.route, "entry": a.entry, "formal": p.formal, "proxy": a.proxy, "maxCalls": a.calls, "workMS": a.workMS, "callMS": a.callMS, "policy": policy(time.Duration(a.callMS) * time.Millisecond), "authorization": auth}); err != nil {
		return err
	}
	files, err := sourceFiles()
	if err != nil {
		return err
	}
	hashes := map[string]string{}
	for name, raw := range files {
		if err := b.file("source/"+name, raw); err != nil {
			return err
		}
		hashes[name] = digest(raw)
	}
	r.CodeHash = digest(encoded(hashes))
	if err := b.record("source-hashes.json", hashes); err != nil {
		return err
	}
	// Protocol evidence remains independently NOT_RUN. This receipt binds each
	// route's candidate hash to ticket 11 without borrowing its controlled PASS.
	receipt, err := readBounded(p.experiment.ProtocolEvidence[p.route], 1<<20)
	if err != nil {
		return errors.New("ticket 11 protocol receipt unavailable")
	}
	r.ProtocolEvidenceHash = digest(receipt)
	return b.file("protocol-evidence.json", receipt)
}
func (b *bundle) finish(r report) error {
	if err := b.record("report.json", r); err != nil {
		return err
	}
	writeManifest := func() error {
		hashes := map[string]string{}
		for k, v := range b.files {
			if k != "manifest.json" {
				hashes[k] = v
			}
		}
		return b.record("manifest.json", manifest{1, runtime.Version(), hashes})
	}
	if err := writeManifest(); err != nil {
		return err
	}
	findings, err := audit.Bundle(b.dir, append(audit.Environment(), audit.Forbidden{Value: b.secret, Label: "isolated evaluation credential"}))
	if err != nil {
		return err
	}
	if findings == nil {
		findings = []audit.Finding{}
	}
	if err := b.record("audit.json", map[string]any{"findings": findings}); err != nil {
		return err
	}
	if err := writeManifest(); err != nil {
		return err
	}
	if len(findings) > 0 {
		return errors.New("evidence redaction audit failed")
	}
	return nil
}

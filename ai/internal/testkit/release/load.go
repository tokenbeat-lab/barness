package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// OfflinePackage is the package argument the offline E2E bundle records.
const OfflinePackage = "./ai/e2e"

// LoadBundle reads an offline evidence bundle: every case's status from
// manifest.json and every differential case's verdict from its pidiff.json.
func LoadBundle(dir string) (Bundle, error) {
	var m struct {
		Package   string `json:"package"`
		GitCommit string `json:"git_commit"`
		Cases     []struct {
			ID     string `json:"id"`
			Status Status `json:"status"`
			Dir    string `json:"dir"`
		} `json:"cases"`
	}
	if err := readJSON(filepath.Join(dir, "manifest.json"), &m); err != nil {
		return Bundle{}, err
	}
	if m.Package != OfflinePackage {
		return Bundle{}, fmt.Errorf("release: %s is a bundle of %q, not the offline E2E (%s)", dir, m.Package, OfflinePackage)
	}
	b := Bundle{Dir: dir, Commit: m.GitCommit}
	for _, c := range m.Cases {
		b.Cases = append(b.Cases, Case{ID: c.ID, Status: c.Status})
		if kindOf(c.ID) != kindDifferential || c.Status == NotRun {
			continue
		}
		var v map[string]json.RawMessage
		path := filepath.Join(dir, c.Dir, "pidiff.json")
		if _, err := os.Stat(path); err != nil {
			continue // the gate reports the missing verdict
		}
		if err := readJSON(path, &v); err != nil {
			return Bundle{}, err
		}
		var pending *int
		if err := json.Unmarshal(v["pending"], &pending); err != nil || pending == nil {
			return Bundle{}, fmt.Errorf("release: %s has no pending count", path)
		}
		d := DiffCase{ID: c.ID, Status: c.Status, Pending: *pending}
		for _, field := range diffRecordFields {
			if raw, ok := v[field]; !ok || string(raw) == "null" || string(raw) == `""` {
				d.Missing = append(d.Missing, field)
			}
		}
		b.Diffs = append(b.Diffs, d)
	}
	return b, nil
}

// diffRecordFields are the fields spec Testing Decisions §6 requires of every
// differential record (the first differing position is first_diff, null
// when nothing differs, so it is not required to be set).
var diffRecordFields = []string{"case_id", "pi_commit", "sdk_versions", "model_catalog_hash", "request_sha256", "frame_sha256", "findings"}

// FindBundle returns the single offline E2E bundle below base, where a
// `go test ./...` run with BARNESS_AI_EVIDENCE_DIR=base wrote it.
func FindBundle(base string) (string, error) {
	entries, err := os.ReadDir(base)
	if err != nil {
		return "", fmt.Errorf("release: %w", err)
	}
	var found []string
	for _, e := range entries {
		var m struct {
			Package string `json:"package"`
		}
		if e.IsDir() && readJSON(filepath.Join(base, e.Name(), "manifest.json"), &m) == nil && m.Package == OfflinePackage {
			found = append(found, filepath.Join(base, e.Name()))
		}
	}
	if len(found) != 1 {
		return "", fmt.Errorf("release: want exactly one offline E2E bundle in %s, found %d", base, len(found))
	}
	return found[0], nil
}

// AuditBundle re-audits a bundle with the gate's own environment and checks
// the audit the run recorded when it finished (audit.json, evidence.Run.Audit),
// which alone knew the run's registered test secrets: a bundle without it,
// or whose recorded audit has findings, fails too.
func AuditBundle(dir string, forbidden []audit.Forbidden) AuditResult {
	a := AuditResult{Bundle: dir}
	findings, err := audit.Bundle(dir, forbidden)
	if err != nil {
		a.Err = err.Error()
		return a
	}
	a.Findings = findings
	var live struct {
		Combo string `json:"combo"`
	}
	if readJSON(filepath.Join(dir, "live-report.json"), &live) == nil {
		a.Combo = live.Combo
	}
	var recorded struct {
		Findings *[]audit.Finding `json:"findings"`
	}
	switch err := readJSON(filepath.Join(dir, "audit.json"), &recorded); {
	case err != nil:
		a.Err = "no audit recorded by the run: " + err.Error()
	case recorded.Findings == nil:
		a.Err = "audit.json has no findings list"
	default:
		for _, f := range *recorded.Findings {
			f.Detail = "recorded by the run: " + f.Detail
			a.Findings = append(a.Findings, f)
		}
	}
	return a
}

// GoModules returns the module versions go.mod in dir pins.
func GoModules(dir string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("release: %w", err)
	}
	out := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(strings.TrimPrefix(strings.TrimSpace(line), "require "))
		if len(fields) >= 2 && strings.Contains(fields[0], ".") && strings.HasPrefix(fields[1], "v") {
			out[fields[0]] = fields[1]
		}
	}
	return out, nil
}

// LoadLedger reads and validates the difference ledger.
func LoadLedger(path string) (Ledger, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Ledger{}, fmt.Errorf("release: %w", err)
	}
	l, err := pioracle.ParseLedger(data)
	if err != nil {
		return Ledger{}, err
	}
	out := Ledger{Fixed: l.Count(pioracle.Fixed), Extension: l.Count(pioracle.Extension), Pending: l.Count(pioracle.Pending)}
	for _, r := range l.Routes() {
		out.Routes = append(out.Routes, Route{Provider: r.Provider, API: r.API})
	}
	return out, nil
}

// LoadTraceMap reads and validates the traceability map.
func LoadTraceMap(path string) (TraceMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TraceMap{}, fmt.Errorf("release: %w", err)
	}
	return ParseTraceMap(data)
}

// CatalogSnapshot is the built-in model catalog with its hash: the model and
// price snapshot a release publishes (spec I10). The checked-in copy must
// equal it.
func CatalogSnapshot() ([]byte, error) {
	c := ai.BuiltinCatalog()
	hash, err := c.Hash()
	if err != nil {
		return nil, err
	}
	return Encode(struct {
		Note    string     `json:"note"`
		Hash    string     `json:"hash"`
		Catalog ai.Catalog `json:"catalog"`
	}{
		Note: "barness-ai built-in model catalog and price snapshot (ai.BuiltinCatalog). Generated by `go run ./ai/release/cmd/releasegate -write-snapshot`; " +
			"prices are estimates, not billing. The release gate fails when this file differs from the code.",
		Hash: hash, Catalog: c,
	})
}

// CheckSnapshot compares the checked-in snapshot at path with the code's.
func CheckSnapshot(name, path string, want []byte) Snapshot {
	got, err := os.ReadFile(path)
	switch {
	case err != nil:
		return Snapshot{Name: name, Detail: err.Error()}
	case !bytes.Equal(got, want):
		return Snapshot{Name: name, Detail: path + " differs from the code; regenerate it with -write-snapshot and review the diff"}
	}
	return Snapshot{Name: name, Current: true}
}

// Fixture is one checked-in offline fixture.
type Fixture struct {
	Path   string `json:"path"`
	Bytes  int    `json:"bytes"`
	SHA256 string `json:"sha256"`
}

// Fixtures indexes every file below root (the offline fixtures), with
// paths relative to base.
func Fixtures(base, root string) ([]Fixture, error) {
	var out []Fixture
	err := filepath.WalkDir(filepath.Join(base, root), func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		out = append(out, Fixture{Path: filepath.ToSlash(rel), Bytes: len(data), SHA256: hex.EncodeToString(sum[:])})
		return nil
	})
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, err
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("release: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("release: %s: %w", path, err)
	}
	return nil
}

// Encode is the JSON encoding release artifacts are written with.
func Encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

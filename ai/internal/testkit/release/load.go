package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// OfflinePackage is the package argument the offline E2E bundle records.
const OfflinePackage = "./ai/e2e"

// LoadBundle verifies native offline case inventories before exposing consumers.
func LoadBundle(dir string) (Bundle, error) {
	source, root, err := openArtifactBundle(dir)
	if err != nil {
		return Bundle{}, err
	}
	defer root.Close()
	var m struct {
		Package   string `json:"package"`
		GitCommit string `json:"git_commit"`
		Cases     []struct {
			ID        string            `json:"id"`
			Status    Status            `json:"status"`
			Dir       string            `json:"dir"`
			Artifacts map[string]string `json:"artifacts"`
		} `json:"cases"`
	}
	if err := readBundleManifest(root, &m); err != nil {
		return Bundle{}, err
	}
	if m.Package != OfflinePackage {
		return Bundle{}, fmt.Errorf("release: %s is a bundle of %q, not the offline E2E (%s)", dir, m.Package, OfflinePackage)
	}
	b := Bundle{Dir: dir, Commit: m.GitCommit, caseFiles: map[string]verifiedCase{}}
	seenDirs := map[string]bool{}
	for _, c := range m.Cases {
		// Manifest identifiers become report fields. Reject sensitive names
		// before conversion rather than copy them into a failed release report.
		if len(audit.File("identifiers.txt", []byte(c.ID+"\n"+c.Dir), audit.Environment())) != 0 {
			return Bundle{}, fmt.Errorf("release: evidence case identifier failed redaction audit")
		}
		// Run.Finish preserves the raw ID, including the meaningful empty-effort
		// leaf. Only that single trailing separator may differ from its physical
		// directory; never clean an otherwise unsafe path into an accepted one.
		physical := strings.TrimSuffix(c.Dir, "/")
		if c.ID == "" || !canonicalArtifactPath(physical) || b.caseFiles[c.ID].bundle != nil || seenDirs[physical] {
			return Bundle{}, fmt.Errorf("release: invalid or duplicate evidence case directory")
		}
		seenDirs[physical] = true
		caseRoot, inventoryErr := openInventoryCase(root, physical)
		if inventoryErr != nil && !errors.Is(inventoryErr, fs.ErrNotExist) {
			return Bundle{}, fmt.Errorf("release: unsafe evidence case directory")
		}
		if inventoryErr == nil {
			inventoryErr = verifyCaseArtifactInventory(caseRoot, c.Artifacts)
			caseRoot.Close()
		}
		if inventoryErr != nil {
			b.Problems = append(b.Problems, inventoryErr.Error())
		}
		files := verifiedCase{bundle: source, dir: physical, artifacts: c.Artifacts, verified: inventoryErr == nil}
		b.caseFiles[c.ID] = files
		b.Cases = append(b.Cases, Case{ID: c.ID, Status: c.Status})
	}
	return b, nil
}

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
	return auditBundle(dir, forbidden, true)
}

// Native qualification has already authenticated its capture. Re-audit the
// bytes and recorded redaction findings without recursively qualifying again.
func auditBundle(dir string, forbidden []audit.Forbidden, checkLive bool, supplement ...string) AuditResult {
	a := AuditResult{Bundle: dir}
	findings, err := audit.Bundle(dir, forbidden)
	if err != nil {
		a.Err = err.Error()
		return a
	}
	a.Findings = safeAuditFindings(findings, forbidden)
	if live, err := supportmatrix.LoadReport(filepath.Join(dir, supportmatrix.ReportFile)); err == nil {
		a.Live = &live
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
		for _, f := range safeAuditFindings(*recorded.Findings, forbidden) {
			f.Detail = "recorded by the run: " + f.Detail
			a.Findings = append(a.Findings, f)
		}
	}
	if len(a.Findings) != 0 {
		a.Err = "evidence bundle failed redaction audit"
		return a
	}
	if checkLive && a.Live != nil && a.Err == "" && len(a.Findings) == 0 {
		r := a.Live
		if err := verifyLiveManifest(dir, *r, supplement...); err != nil {
			a.Err = "live artifact integrity verification failed"
			return a
		}
		if (supportmatrix.Row{Combo: r.Combo, Operation: r.Operation, Provider: r.Provider, API: r.API}).ValidIdentity() {
			a.Combo = r.Combo
		}
	}
	return a
}

// An audit can itself name an unsafe artifact or JSON field. Keep the rule
// and a stable path digest for diagnosis without republishing that value.
func safeAuditFindings(findings []audit.Finding, forbidden []audit.Forbidden) []audit.Finding {
	out := make([]audit.Finding, 0, len(findings))
	for _, f := range findings {
		switch f.Rule {
		case audit.RuleSecret, audit.RuleKeyShaped, audit.RuleCredentialHeader, audit.RuleBearer, audit.RuleCredentialQuery,
			audit.RuleObservationField, audit.RuleResponseHeader, audit.RuleEnvironment, audit.RuleUnparseable:
		default:
			f.Rule, f.Detail = audit.RuleUnparseable, "recorded audit contains an invalid rule"
		}
		if len(audit.File("audit-diagnostic.txt", []byte(f.File+"\n"+f.Detail), forbidden)) != 0 {
			f.File = "artifact-sha256:" + evidence.Digest([]byte(f.File))
			f.Detail = "artifact content rejected by redaction audit"
		}
		out = append(out, f)
	}
	return out
}

// GoModules returns the module versions go.mod in dir pins.
func GoModules(dir string) (map[string]string, error) {
	data, err := readArtifact(filepath.Join(dir, "go.mod"), 1<<20)
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
	data, err := readArtifact(path, maxInventoryFile)
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

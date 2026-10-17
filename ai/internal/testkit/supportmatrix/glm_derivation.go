package supportmatrix

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
)

// GLMDerivation identifies the consumer that corrected the generic known-usage
// aggregates while retaining the original measured report/capture bytes.
type GLMDerivation struct {
	Schema                      int       `json:"schema"`
	Kind                        string    `json:"kind"`
	Route                       string    `json:"route"`
	RawCaptureIndexSHA256       string    `json:"rawCaptureIndexSHA256"`
	RawArchiveSHA256            string    `json:"rawArchiveSHA256"`
	OriginalFormalReportSHA256  string    `json:"originalFormalReportSHA256"`
	OriginalSupportReportSHA256 string    `json:"originalSupportReportSHA256"`
	OriginalGitCommit           string    `json:"originalGitCommit"`
	ImporterGitCommit           string    `json:"importerGitCommit"`
	ImporterSourceSHA256        string    `json:"importerSourceSHA256"`
	DerivedAt                   time.Time `json:"derivedAt"`
}

// CreateGLMSupportView atomically writes a separate, audited no-paid view.
// The target must not exist. The trusted raw ZIP and measured Go source are
// never edited, and replay continues to use only those original paid captures.
func CreateGLMSupportView(route, destination string) (Report, error) {
	var empty Report
	if len(GLMExpected("zhipu-"+route)) == 0 {
		return empty, glmBad("unknown support view route")
	}
	if _, e := os.Lstat(destination); !os.IsNotExist(e) {
		return empty, glmBad("support view target already exists")
	}
	raw, e := loadGLMRawCapture(route)
	if e != nil {
		return empty, e
	}
	stage, e := os.MkdirTemp(filepath.Dir(destination), ".glm-support-view-")
	if e != nil {
		return empty, e
	}
	defer os.RemoveAll(stage)
	original := filepath.Join(stage, "original")
	for name, data := range raw.files {
		p := filepath.Join(original, filepath.FromSlash(name))
		if e = os.MkdirAll(filepath.Dir(p), 0755); e != nil {
			return empty, e
		}
		if e = os.WriteFile(p, data, 0600); e != nil {
			return empty, e
		}
	}
	r, e := NewGLMReport(original)
	if e != nil {
		return empty, e
	}
	commit, source, e := glmImporterIdentity()
	if e != nil {
		return empty, e
	}
	committed, err := glmImporterCommitHash(commit)
	if err != nil || committed != source {
		return empty, glmBad("consumer source must be committed before derivation")
	}
	proof := &GLMDerivation{Schema: 1, Kind: "glm-derived-live-support", Route: route, RawCaptureIndexSHA256: raw.indexHash, RawArchiveSHA256: raw.archiveHash, OriginalFormalReportSHA256: raw.reportHash, OriginalSupportReportSHA256: raw.supportHash, OriginalGitCommit: raw.commit, ImporterGitCommit: commit, ImporterSourceSHA256: source, DerivedAt: time.Now().UTC()}
	r.GLM.Derivation = proof
	r.evidenceDir = stage
	write := func(name string, v any) error {
		b, e := Encode(v)
		if e != nil {
			return e
		}
		return os.WriteFile(filepath.Join(stage, name), b, 0600)
	}
	if e = write("manifest.json", glmViewManifest{Schema: 1, Kind: "glm-derived-live-support", Mode: "no-paid-derivation", NetworkRequests: 0, Derivation: *proof}); e != nil {
		return empty, e
	}
	if e = write("glm-support-derivation.json", proof); e != nil {
		return empty, e
	}
	if e = write(ReportFile, r); e != nil {
		return empty, e
	}
	findings, e := audit.Bundle(stage, nil)
	if e != nil || len(findings) > 0 {
		return empty, glmBad("derived support view failed audit")
	}
	if findings == nil {
		findings = []audit.Finding{}
	}
	if e = write("audit.json", map[string]any{"findings": findings}); e != nil {
		return empty, e
	}
	inventory, e := glmViewInventory(stage, raw.files)
	if e != nil {
		return empty, e
	}
	if e = write("glm-support-integrity.json", map[string]any{"schema": 1, "files": inventory}); e != nil {
		return empty, e
	}
	if e = VerifyGLMReport(r); e != nil {
		return empty, e
	}
	if e = os.Rename(stage, destination); e != nil {
		return empty, e
	}
	r.evidenceDir = destination
	return r, nil
}

func NewGLMSupportViewReport(dir string) (Report, error) {
	var empty Report
	var proof GLMDerivation
	if e := readJSON(filepath.Join(dir, "glm-support-derivation.json"), &proof); e != nil {
		return empty, glmBad("support derivation missing or invalid")
	}
	if proof.Schema != 1 || proof.Kind != "glm-derived-live-support" || proof.DerivedAt.IsZero() || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(proof.ImporterGitCommit) {
		return empty, glmBad("support derivation identity invalid")
	}
	raw, e := loadGLMRawCapture(proof.Route)
	if e != nil {
		return empty, e
	}
	if proof.RawCaptureIndexSHA256 != raw.indexHash || proof.RawArchiveSHA256 != raw.archiveHash || proof.OriginalFormalReportSHA256 != raw.reportHash || proof.OriginalSupportReportSHA256 != raw.supportHash || proof.OriginalGitCommit != raw.commit {
		return empty, glmBad("support derivation substituted raw capture/source identity")
	}
	// Reject unknown paths/counts/bytes before hashing view payloads, including
	// stale views: source errors must not hide unbounded shape traversal.
	paths, e := glmViewPaths(dir, raw.files)
	if e != nil {
		return empty, e
	}
	_, current, e := glmImporterIdentity()
	if e != nil || proof.ImporterSourceSHA256 != current {
		return empty, glmBad("consumer/importer source drift")
	}
	committed, commitError := glmImporterCommitHash(proof.ImporterGitCommit)
	if commitError != nil || committed != proof.ImporterSourceSHA256 {
		return empty, glmBad("claimed importer commit does not contain its source identity")
	}
	var inventory struct {
		Schema int
		Files  map[string]string
	}
	if readGLMJSON(filepath.Join(dir, "glm-support-integrity.json"), &inventory) != nil || inventory.Schema != 1 {
		return empty, glmBad("support view inventory missing")
	}
	actual, e := glmHashViewPaths(paths)
	if e != nil || !reflect.DeepEqual(actual, inventory.Files) {
		return empty, glmBad("support view inventory changed")
	}
	var receipt struct{ Findings []audit.Finding }
	if readGLMJSON(filepath.Join(dir, "audit.json"), &receipt) != nil || receipt.Findings == nil || len(receipt.Findings) > 0 {
		return empty, glmBad("support view audit missing or failed")
	}
	if findings, e := audit.Bundle(dir, nil); e != nil || len(findings) > 0 {
		return empty, glmBad("support view independent audit failed")
	}
	original := filepath.Join(dir, "original")
	if e = glmOriginalMatches(original, raw); e != nil {
		return empty, e
	}
	r, e := NewGLMReport(original)
	if e != nil {
		return empty, e
	}
	if r.GitCommit != proof.OriginalGitCommit {
		return empty, glmBad("support view measured source differs")
	}
	r.GLM.Derivation = &proof
	r.evidenceDir = dir
	return r, nil
}

type glmViewManifest struct {
	Schema          int           `json:"schema"`
	Kind            string        `json:"kind"`
	Mode            string        `json:"mode"`
	NetworkRequests int           `json:"networkRequests"`
	Derivation      GLMDerivation `json:"derivation"`
}

package supportmatrix

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"time"
)

const KimiCompositionFile = "kimi-composition.json"

type KimiFileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type KimiEvidenceRef struct {
	Bundle         string      `json:"bundle"`
	ManifestSHA256 string      `json:"manifestSHA256"`
	ReportSHA256   string      `json:"reportSHA256"`
	SealSHA256     string      `json:"sealSHA256"`
	Producer       KimiFileRef `json:"producer"`
}
type KimiComposition struct {
	Schema     int             `json:"schema"`
	Kind       string          `json:"kind"`
	Combo      string          `json:"combo"`
	Model      string          `json:"model"`
	Baseline   KimiEvidenceRef `json:"baseline"`
	Supplement KimiEvidenceRef `json:"supplement"`
}
type KimiMeasurement struct {
	Kind     string          `json:"kind"`
	Evidence KimiEvidenceRef `json:"evidence"`
	Original Report          `json:"original"`
}
type KimiCaseEvidence struct {
	Measurement    int               `json:"measurement"`
	CaseID         string            `json:"caseId"`
	ManifestSHA256 string            `json:"manifestSHA256"`
	Artifacts      map[string]string `json:"artifacts"`
}

// KimiQualification is a current decision over two real measurements. It is
// deliberately not Report: no invented shared commit/time/grant/budget exists.
// Decision time belongs to the enclosing inspection/consumer receipt.
type KimiQualification struct {
	Schema        int                         `json:"schema"`
	Combo         string                      `json:"combo"`
	Model         string                      `json:"model"`
	Mode          string                      `json:"mode"`
	Qualification string                      `json:"qualification"`
	Expected      []string                    `json:"expected"`
	Measurements  []KimiMeasurement           `json:"measurements"`
	Coverage      map[string]KimiCaseEvidence `json:"coverage"`
}

func IsKimiComposition(dir string) bool {
	_, e := os.Lstat(filepath.Join(dir, KimiCompositionFile))
	return e == nil
}

// VerifyKimiComposition is the only composition seam: exactly old-nine plus
// history-one, no recursive combination or caller-selected capability set.
func VerifyKimiComposition(dir string, formal bool) (KimiQualification, error) {
	var q KimiQualification
	root, e := kimiRepositoryRoot()
	if e != nil {
		return q, e
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 || entries[0].Name() != KimiCompositionFile {
		return q, refuse("Kimi composition must contain one descriptor")
	}
	var c KimiComposition
	r, e := os.OpenRoot(dir)
	if e != nil {
		return q, e
	}
	_, e = regionalRead(r, KimiCompositionFile)
	r.Close()
	if e != nil || readJSON(filepath.Join(dir, KimiCompositionFile), &c) != nil || c.Schema != 1 || c.Kind != "kimi-history-composition" || !RegionalIsKimi(c.Combo) || c.Model != "kimi-k3" || c.Baseline.Bundle == c.Supplement.Bundle {
		return q, refuse("invalid fixed Kimi composition")
	}
	refs := []KimiEvidenceRef{c.Baseline, c.Supplement}
	var original []Report
	var manifests []struct{ Cases []regionalManifestCase }
	mode := ""
	for i, ref := range refs {
		p, e := kimiCheckedEvidence(root, ref)
		if e != nil {
			return q, e
		}
		var header Report
		if readJSON(filepath.Join(p, RegionalReportFile), &header) != nil || header.Regional == nil {
			return q, refuse("Kimi measurement missing")
		}
		if i == 0 {
			mode = header.Regional.Mode
		}
		selection := kimiBaselineNine
		if i == 1 {
			selection = kimiHistoryOne
		}
		measured, e := regionalSelectedReport(p, false, mode, selection)
		if e != nil {
			return q, e
		}
		if measured.Combo != c.Combo || measured.Model != c.Model {
			return q, refuse("Kimi composition crosses route or model")
		}
		if _, e := verifyKimiProducer(root, p, ref.Producer, measured); e != nil {
			return q, e
		}
		var m struct{ Cases []regionalManifestCase }
		manifestRaw, manifestErr := regionalReadFile(filepath.Join(p, "manifest.json"))
		if manifestErr != nil || regionalDecode(manifestRaw, &m) != nil {
			return q, refuse("Kimi original manifest missing")
		}
		original = append(original, measured)
		manifests = append(manifests, m)
	}
	a, b := original[0], original[1]
	if a.Provider != b.Provider || a.API != b.API || a.Endpoint != b.Endpoint || a.AccountAlias != b.AccountAlias || a.Regional.RegionAlias != b.Regional.RegionAlias || a.SDK != b.SDK {
		return q, refuse("Kimi measurements cross service, account, region or SDK")
	}
	var grant RegionalGrant
	if readJSON(filepath.Join(root, filepath.FromSlash(c.Supplement.Bundle), "regional-grant.json"), &grant) != nil || grant.BaselineManifestSHA256 != c.Baseline.ManifestSHA256 {
		return q, refuse("Kimi supplement approved a different baseline")
	}
	q = KimiQualification{Schema: 1, Combo: c.Combo, Model: c.Model, Mode: mode, Expected: RegionalExpected(c.Combo), Coverage: map[string]KimiCaseEvidence{}, Qualification: "CANDIDATE_COMPLETE"}
	if mode == "controlled" {
		q.Qualification = "CONTROLLED_COMPLETE"
	}
	for i, r := range original {
		kind := "baseline-nine"
		if i == 1 {
			kind = "history-one"
		}
		q.Measurements = append(q.Measurements, KimiMeasurement{Kind: kind, Evidence: refs[i], Original: r})
		for _, s := range r.Scenarios {
			if _, dup := q.Coverage[s.ID]; dup || !slices.Contains(q.Expected, s.ID) {
				return KimiQualification{}, refuse("Kimi duplicate or substituted capability")
			}
			found := false
			for _, c := range manifests[i].Cases {
				if c.ID == s.CaseID {
					q.Coverage[s.ID] = KimiCaseEvidence{Measurement: i, CaseID: c.ID, ManifestSHA256: refs[i].ManifestSHA256, Artifacts: c.Artifacts}
					found = true
				}
			}
			if !found {
				return KimiQualification{}, refuse("Kimi original capability provenance absent")
			}
		}
	}
	if len(q.Coverage) != len(q.Expected) {
		return KimiQualification{}, refuse("Kimi composition lacks current required capability")
	}
	// Both independent native prefixes already enforce account/tenant ownership.
	// Also bind the two producer tenants without exposing native trust handles.
	var tenants []string
	for _, r := range original {
		p := filepath.Join(root, r.evidenceDir)
		if filepath.IsAbs(r.evidenceDir) {
			p = r.evidenceDir
		}
		var result regionalResult
		resultRaw, resultErr := regionalReadFile(filepath.Join(p, r.Scenarios[0].CaseID, "result.json"))
		if resultErr != nil || regionalDecode(resultRaw, &result) != nil {
			return KimiQualification{}, refuse("Kimi measurement tenant absent")
		}
		tenants = append(tenants, result.Metadata.TenantID)
	}
	if !reflect.DeepEqual(tenants, []string{tenants[0], tenants[0]}) {
		return KimiQualification{}, refuse("Kimi measurements cross tenant")
	}
	if formal {
		if mode != "live" {
			return KimiQualification{}, refuse("controlled composition never grants actual formal support")
		}
		if _, e := RegionalFormalCatalog(c.Combo); e != nil {
			return KimiQualification{}, e
		}
		q.Qualification = "FORMAL_APPLICABLE"
	}
	return q, nil
}

// Inspection time is distinct from every measurement's original timestamps.
type KimiInspection struct {
	KimiQualification
	QualificationAt time.Time `json:"qualificationAt"`
}

func kimiCheckedEvidence(root string, ref KimiEvidenceRef) (string, error) {
	dir, e := kimiLocalPath(root, ref.Bundle)
	if e != nil {
		return "", e
	}
	if IsKimiComposition(dir) {
		return "", refuse("nested Kimi composition forbidden")
	}
	r, e := os.OpenRoot(dir)
	if e != nil {
		return "", e
	}
	defer r.Close()
	for name, h := range map[string]string{"manifest.json": ref.ManifestSHA256, RegionalReportFile: ref.ReportSHA256, regionalSealFile: ref.SealSHA256} {
		raw, e := regionalRead(r, name)
		if e != nil || !validManifestHash(h) || glmHash(raw) != h {
			return "", fmt.Errorf("%w: Kimi original %s changed", ErrInvalidReport, name)
		}
	}
	return dir, nil
}

package supportmatrix

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

const RegionalReportFile = "regional-report.json"

// RegionalEvidence retains the exact own qualification contract. Candidate
// success remains qualification evidence, never a formal support projection.
type RegionalEvidence struct {
	Schema             int    `json:"schema"`
	Mode               string `json:"mode"`
	Stage              string `json:"stage"`
	Purpose            string `json:"purpose"`
	CredentialProvider string `json:"credentialProvider"`
	RegionAlias        string `json:"regionAlias"`
	Contract           string `json:"contract"`
	GrantID            string `json:"grantId"`
	GrantSHA256        string `json:"grantSha256"`
	SourceSHA256       string `json:"sourceSha256"`
	BuildSourceSHA256  string `json:"buildSourceSha256"`
	BinarySHA256       string `json:"binarySha256"`
	Complete           bool   `json:"complete"`
	AuditPassed        bool   `json:"auditPassed"`
	// NativeSourceCases is predeclared per identity, not a fallback lookup.
	NativeSourceCases map[string]string `json:"nativeSourceCases"`
}

func VerifyRegionalQualification(dir string) (Report, error) {
	return regionalReport(dir, false, "live")
}
func NewRegionalReport(dir string) (Report, error) { return regionalReport(dir, true, "live") }

// VerifyRegionalControlledCapture verifies the complete keyless native bundle.
// Its controlled origin is retained and cannot satisfy qualification or support.
func VerifyRegionalControlledCapture(dir string) (Report, error) {
	return regionalReport(dir, false, "controlled")
}

func regionalReport(dir string, formal bool, mode string) (Report, error) {
	var r Report
	if err := readJSON(filepath.Join(dir, RegionalReportFile), &r); err != nil {
		return Report{}, err
	}
	r.evidenceDir = dir
	route, ok := RegionalRouteOf(r.Combo)
	e := r.Regional
	alias := regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	if !ok || r.Schema != SchemaVersion || r.Operation != "chat" || r.Provider != route.Provider || r.API != route.API || r.Endpoint != route.Endpoint || r.Model != RegionalModels(r.Combo)[0] || e == nil || e.Schema != 1 || e.Mode != mode || !slices.Contains([]string{"candidate", "formal"}, e.Stage) || formal && e.Stage != "formal" || e.Purpose != "standard" || e.CredentialProvider != r.Provider || e.Contract != RegionalAuthorization(r.Combo) || !alias.MatchString(e.GrantID) || len(e.GrantSHA256) != 64 || !alias.MatchString(r.AccountAlias) || !alias.MatchString(e.RegionAlias) || mode == "live" && e.BuildSourceSHA256 != e.SourceSHA256 || len(e.BinarySHA256) != 64 || !e.Complete || !e.AuditPassed || r.GLM != nil || r.MiniMax != nil {
		return Report{}, refuse("regional own service, grant, stage or mode invalid")
	}
	if r.GitCommit == "" || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) || r.FinishedAt.Sub(r.StartedAt) > 1800*time.Second {
		return Report{}, refuse("regional execution time/source missing")
	}
	if !reflect.DeepEqual(e.NativeSourceCases, regionalNativeSources(r.Combo)) {
		return Report{}, refuse("regional native source case selection differs")
	}
	sdk, err := RegionalSDK(r.Combo)
	if err != nil || r.SDK != sdk {
		return Report{}, refuse("regional SDK differs from selected dependency")
	}
	cat, err := regional.CandidateCatalog()
	if e.Stage == "formal" {
		cat, err = RegionalFormalCatalog(r.Combo)
	}
	if err != nil {
		return Report{}, err
	}
	hash, err := cat.Hash()
	if err != nil || r.CatalogVersion != cat.Version || r.CatalogHash != hash {
		return Report{}, refuse("regional fixed catalog/price snapshot differs")
	}
	for _, model := range RegionalModels(r.Combo) {
		if _, ok := cat.Lookup(ai.ProviderID(r.Provider), ai.API(r.API), model); !ok {
			return Report{}, refuse("regional full model identity absent from selected catalog")
		}
	}
	b := r.Budget
	if b.MaxCalls != 24 || b.MaxAttempts != 24 || b.MaxOutputTokens != 4096 || b.MaxRequestedOutputTokens != 98304 || b.MaxReservedInputTokens != 120000 || b.MaxRetries != 0 || b.CallSeconds != 120 || b.ProcessSeconds != 1800 || b.CallsUsed < 0 || b.CallsUsed > 24 || b.AttemptsReserved != b.CallsUsed || b.HTTPAttempts != b.CallsUsed || b.RequestedOutputTokens < b.CallsUsed || b.RequestedOutputTokens > b.CallsUsed*4096 || b.ReservedInputTokens < b.CallsUsed*2048 || b.ReservedInputTokens > b.MaxReservedInputTokens {
		return Report{}, refuse("regional own fixed budget differs")
	}
	expected := RegionalExpected(r.Combo)
	if !slices.Equal(r.Expected, expected) || len(r.Scenarios) != len(expected) || b.CallsUsed != len(expected) {
		return Report{}, refuse("regional complete model-qualified set required")
	}
	for i, s := range r.Scenarios {
		model, _, _ := strings.Cut(s.ID, "/")
		if s.ID != expected[i] || s.Model != model || s.Outcome != Pass || s.CaseID != RegionalCaseID(r.Combo, s.ID) || s.Calls != 1 || s.Attempts != 1 || s.Retries != 0 || s.DurationMS < 0 || s.DurationMS > 120000 {
			return Report{}, refuse("regional own scenario incomplete or substituted")
		}
	}
	if err := verifyRegionalBundle(dir, r); err != nil {
		return Report{}, err
	}
	return r, nil
}

// Formal promotion is a reviewed data change followed by its own fresh live
// execution. Catalog existence alone cannot replace the accepted regional facts.
func RegionalFormalCatalog(combo string) (ai.Catalog, error) {
	route, ok := RegionalRouteOf(combo)
	if !ok {
		return ai.Catalog{}, refuse("unknown regional formal route")
	}
	formal := ai.BuiltinCatalog()
	candidate, err := regional.CandidateCatalog()
	if err != nil {
		return ai.Catalog{}, err
	}
	for _, id := range RegionalModels(combo) {
		want, exists := candidate.Lookup(ai.ProviderID(route.Provider), ai.API(route.API), id)
		got, promoted := formal.Lookup(ai.ProviderID(route.Provider), ai.API(route.API), id)
		want.Name = got.Name
		if !exists || !promoted || !reflect.DeepEqual(want, got) {
			return ai.Catalog{}, refuse("regional formal model unpromoted or differs from own accepted facts")
		}
	}
	return formal, nil
}

func RegionalCaseID(combo, id string) string {
	return "REGIONAL-" + combo + "-" + strings.ReplaceAll(id, "/", "-")
}

func VerifyRegionalReport(r Report) error {
	if r.evidenceDir == "" || r.Regional == nil {
		return refuse("regional support needs an owned sealed formal capture")
	}
	want, err := NewRegionalReport(r.evidenceDir)
	if err != nil {
		return err
	}
	gotRaw, _ := json.Marshal(r)
	wantRaw, _ := json.Marshal(want)
	if !reflect.DeepEqual(gotRaw, wantRaw) {
		return refuse("regional support projection differs from own capture")
	}
	return nil
}

func checkRegional(r Report) error {
	if _, ok := RegionalRouteOf(r.Combo); ok {
		return VerifyRegionalReport(r)
	}
	if r.Regional != nil {
		return fmt.Errorf("%w: foreign regional proof", ErrInvalidReport)
	}
	return nil
}

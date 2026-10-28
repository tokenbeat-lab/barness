package supportmatrix

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

const MiniMaxThinkingCompositionFile = "minimax-thinking-composition.json"

type MiniMaxThinkingRef struct {
	Bundle         string                      `json:"bundle"`
	ManifestSHA256 string                      `json:"manifestSHA256"`
	ReportSHA256   string                      `json:"reportSHA256"`
	SealSHA256     string                      `json:"sealSHA256"`
	Producer       minimaxlive.ThinkingFileRef `json:"producer"`
}
type MiniMaxThinkingComposition struct {
	Schema     int                `json:"schema"`
	Kind       string             `json:"kind"`
	Combo      string             `json:"combo"`
	Model      string             `json:"model"`
	Baseline   MiniMaxThinkingRef `json:"baseline"`
	Supplement MiniMaxThinkingRef `json:"supplement"`
}
type MiniMaxThinkingMeasurement struct {
	Kind     string             `json:"kind"`
	Evidence MiniMaxThinkingRef `json:"evidence"`
	Original minimaxlive.Report `json:"original"`
}
type MiniMaxThinkingCase struct {
	Measurement    int               `json:"measurement"`
	CaseID         string            `json:"caseId"`
	ManifestSHA256 string            `json:"manifestSHA256"`
	Artifacts      map[string]string `json:"artifacts"`
}
type MiniMaxThinkingQualification struct {
	SupplementBilling minimaxlive.ThinkingBilling    `json:"supplementBilling"`
	Schema            int                            `json:"schema"`
	Combo             string                         `json:"combo"`
	Model             string                         `json:"model"`
	Mode              string                         `json:"mode"`
	Qualification     string                         `json:"qualification"`
	Expected          []string                       `json:"expected"`
	Measurements      []MiniMaxThinkingMeasurement   `json:"measurements"`
	Coverage          map[string]MiniMaxThinkingCase `json:"coverage"`
	ExcludedFailures  []MiniMaxThinkingCase          `json:"excludedFailures"`
}

func IsMiniMaxThinkingComposition(dir string) bool {
	_, e := os.Lstat(filepath.Join(dir, MiniMaxThinkingCompositionFile))
	return e == nil
}

// Fixed thirteen + one; original FAIL and budgets remain measurement facts.
// No caller-selected scene subsets, recursive references or unified Report.
func VerifyMiniMaxThinkingComposition(dir string, formal bool) (MiniMaxThinkingQualification, error) {
	var q MiniMaxThinkingQualification
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 || entries[0].Name() != MiniMaxThinkingCompositionFile {
		return q, fmt.Errorf("MiniMax composition requires exactly one descriptor")
	}
	raw, e := minimaxlive.ReadFile(filepath.Join(dir, MiniMaxThinkingCompositionFile), 1<<20)
	var c MiniMaxThinkingComposition
	if e != nil || minimaxlive.StrictJSON(raw, &c) != nil || c.Schema != 1 || c.Kind != "minimax-messages-thinking-composition" || c.Combo != "minimax-messages" || c.Model != minimaxlive.M3 || c.Baseline.Bundle == c.Supplement.Bundle {
		return q, fmt.Errorf("invalid fixed MiniMax thirteen plus one composition")
	}
	root, e := minimaxlive.Root()
	if e != nil {
		return q, e
	}
	refs := []MiniMaxThinkingRef{c.Baseline, c.Supplement}
	q = MiniMaxThinkingQualification{Schema: 1, Combo: c.Combo, Model: c.Model, Coverage: map[string]MiniMaxThinkingCase{}}
	tenant, account := "", ""
	for i, ref := range refs {
		native, e := minimaxlive.LocalEvidencePath(root, ref.Bundle)
		if e != nil {
			return q, e
		}
		for n, h := range map[string]string{"manifest.json": ref.ManifestSHA256, minimaxlive.ReportFile: ref.ReportSHA256, minimaxlive.SealFile: ref.SealSHA256} {
			raw, e := minimaxlive.ReadFile(filepath.Join(native, n), 4<<20)
			if e != nil || !validManifestHash(h) || minimaxlive.Hash(raw) != h {
				return q, fmt.Errorf("MiniMax original reference changed: %s", n)
			}
		}
		var r minimaxlive.Report
		if i == 0 {
			r, e = minimaxlive.VerifyThinkingBaseline(native, formal)
		} else {
			r, e = minimaxlive.VerifyThinkingSupplement(native, formal)
		}
		if e != nil {
			return q, e
		}
		if _, e = minimaxlive.VerifyThinkingProducer(native, ref.Producer, r, i == 0); e != nil {
			return q, e
		}
		if i == 0 {
			q.Mode = r.Mode
		} else {
			q.SupplementBilling, e = minimaxlive.AssessThinkingBilling(native, r)
			if e != nil {
				return q, e
			}
			a := q.Measurements[0].Original
			g, _, _, err := livegrant.ReadEvidence(native)
			if err != nil || g.Scope.BaselineManifestSHA256 != c.Baseline.ManifestSHA256 || r.Identity != a.Identity || r.Mode != a.Mode || r.SDK != a.SDK || r.AccountAlias != a.AccountAlias || r.RegionAlias != a.RegionAlias {
				return q, fmt.Errorf("MiniMax supplement crosses baseline, mode, service or account")
			}
		}
		kind := "baseline-thirteen"
		if i == 1 {
			kind = "thinking-on-one"
		}
		q.Measurements = append(q.Measurements, MiniMaxThinkingMeasurement{kind, ref, r})
		var manifest struct {
			Cases []struct {
				ID        string            `json:"id"`
				Artifacts map[string]string `json:"artifacts"`
			}
		}
		raw, e := minimaxlive.ReadFile(filepath.Join(native, "manifest.json"), 4<<20)
		if e != nil || json.Unmarshal(raw, &manifest) != nil {
			return q, fmt.Errorf("MiniMax manifest absent")
		}
		for _, call := range r.Calls {
			var saved minimaxlive.Capture
			raw, e := minimaxlive.ReadFile(filepath.Join(native, call.CaseID, "result.json"), 2<<20)
			if e != nil || json.Unmarshal(raw, &saved) != nil {
				return q, fmt.Errorf("MiniMax contribution scope missing")
			}
			m := saved.Result.Metadata
			if tenant == "" {
				tenant, account = m.TenantID, m.AccountScopeID
			}
			if tenant == "" || account == "" || m.TenantID != tenant || (m.AccountScopeID != account && !(call.ID == "preflight-error" && m.AccountScopeID == "" && len(m.Attempts) == 0)) {
				return q, fmt.Errorf("MiniMax contributions cross tenant or account scope")
			}
			var fact MiniMaxThinkingCase
			found := false
			for _, mc := range manifest.Cases {
				if mc.ID == call.CaseID {
					fact = MiniMaxThinkingCase{i, call.CaseID, ref.ManifestSHA256, mc.Artifacts}
					found = true
				}
			}
			if !found {
				return q, fmt.Errorf("MiniMax contribution artifact attribution absent")
			}
			if call.Outcome != "PASS" {
				if call.ID == "thinking-on" {
					q.ExcludedFailures = append(q.ExcludedFailures, fact)
				}
				continue
			}
			if i == 0 && call.ID == "thinking-on" || i == 1 && call.ID != "thinking-on" {
				return q, fmt.Errorf("MiniMax contribution outside fixed partition")
			}
			if _, dup := q.Coverage[call.ID]; dup {
				return q, fmt.Errorf("MiniMax duplicate contribution")
			}
			q.Coverage[call.ID] = fact
		}
	}
	for _, p := range minimaxlive.Plan(q.Measurements[0].Original.Identity) {
		q.Expected = append(q.Expected, p.ID)
		if _, ok := q.Coverage[p.ID]; !ok {
			return q, fmt.Errorf("MiniMax missing required capability")
		}
	}
	if len(q.Expected) != 14 || len(q.Coverage) != 14 || len(q.ExcludedFailures) != 1 {
		return q, fmt.Errorf("MiniMax composition coverage differs")
	}
	q.Qualification = "CANDIDATE_COMPLETE"
	if q.Mode == "controlled" {
		q.Qualification = "CONTROLLED_COMPLETE"
	}
	if formal {
		if q.Mode != "live" {
			return q, fmt.Errorf("controlled MiniMax composition cannot certify real support")
		}
		q.Qualification = "FORMAL_APPLICABLE"
	}
	return q, nil
}

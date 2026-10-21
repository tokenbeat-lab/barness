package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// Recovery7RenewalID is the fixed continuation of the closed sixth campaign.
// It does not make arbitrary rounds selectable or recycle consumed rights.
const Recovery7RenewalID = "provider-sites-final-acceptance-7"
const sixthCohortSeal = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-6/seal.json"
const sixthStopIndex = ".evidence/provider-sites/final-acceptance-20261020/recovery6-stopped/index.json"
const seventhCollection = ".evidence/provider-sites/final-acceptance-20261020/recovery7/"
const seventhActualApproval = ".evidence/provider-sites/final-acceptance-20261020/recovery7-authorization.json"

func sixthCohortClosed() bool {
	owner, e := evidenceOwner()
	if e != nil {
		return true
	}
	_, e = os.Lstat(filepath.Join(owner, sixthCohortSeal))
	return !os.IsNotExist(e)
}

// Recovery7Facts validates all closed cohorts without reading a Key or issuing
// a grant. Recovery6Facts independently binds the current spending guard bytes.
func Recovery7Facts() (RenewalSourceFacts, error) {
	f, e := Recovery6Facts()
	f.RenewalID = Recovery7RenewalID
	if e != nil {
		return f, e
	}
	f.SixthCohort, e = closedSixthCohort(f)
	if e != nil {
		return f, e
	}
	f.SixthCohortHistorySHA256 = sixthHistoryPin(f.SixthCohort)
	return f, nil
}
func sixthHistoryPin(h *SixthCohortHistory) string { raw, _ := json.Marshal(h); return Hash(raw) }

func approvedRecovery7(active bool) (RenewalGrant, string, error) {
	g, pin, base, e := loadCanonicalRenewalGrant(Recovery7RenewalID)
	if e != nil {
		return g, "", e
	}
	f, e := Recovery7Facts()
	if e != nil {
		return g, "", e
	}
	if g.Schema != 8 || g.RenewalID != Recovery7RenewalID || g.Decision != "APPROVED" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) || !sameRenewalHistory(g, f) || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !reflect.DeepEqual(g.FourthCohort, f.FourthCohort) || !reflect.DeepEqual(g.FifthCohort, f.FifthCohort) || !reflect.DeepEqual(g.SixthCohort, f.SixthCohort) || !reflect.DeepEqual(g.SpendingGuardSourceSHA256, f.SpendingGuardSourceSHA256) || !fixedRenewalCaps(g) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("fixed recovery7 source/history/guard/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !regexpHex64(p) {
			return g, "", fmt.Errorf("recovery7 authority pin missing")
		}
	}
	if e = recovery7Approval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

// Recovery7Preflight remains read-only; BeginRenewal owns the process nonce.
func Recovery7Preflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(Recovery7RenewalID, route)
}

// A closed grant remains readable at its original source. This cannot admit a
// process or qualify an expired report against a different current source.
func closedSixthAuthority() (RenewalGrant, string, error) {
	f, e := Recovery7Facts()
	if e != nil {
		return RenewalGrant{}, "", e
	}
	g, pin, _, e := loadCanonicalRenewalGrant(Recovery6RenewalID)
	if e != nil {
		return g, "", e
	}
	if pin != f.SixthCohort.NativeAuthoritySHA256["grant.json"] || g.SourceCommit != f.SixthCohort.NativeSourceCommit {
		return g, "", fmt.Errorf("closed sixth authority changed after facts")
	}
	return g, pin, nil
}

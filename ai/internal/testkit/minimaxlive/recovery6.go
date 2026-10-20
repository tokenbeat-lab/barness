package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// Recovery6RenewalID is the fixed continuation of the closed fifth campaign.
// It does not make arbitrary rounds selectable or recycle consumed rights.
const Recovery6RenewalID = "provider-sites-final-acceptance-6"
const fifthCohortSeal = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-5/seal.json"
const fifthStopIndex = ".evidence/provider-sites/final-acceptance-20261020/recovery5-stopped/index.json"
const sixthCollection = ".evidence/provider-sites/final-acceptance-20261020/recovery6/"
const sixthActualApproval = ".evidence/provider-sites/final-acceptance-20261020/recovery6-authorization.json"

// FifthCohortHistory uses the same accounting vocabulary as the fourth cohort.
// Its validator below fixes a distinct eight-process failed history.
type FifthCohortHistory FourthCohortHistory

func fifthCohortClosed() bool {
	owner, e := evidenceOwner()
	if e != nil {
		return true
	}
	_, e = os.Lstat(filepath.Join(owner, fifthCohortSeal))
	return !os.IsNotExist(e)
}

// Recovery6Facts validates all closed cohorts without reading a Key or issuing
// a grant. Recovery5Facts independently binds the current spending guard bytes.
func Recovery6Facts() (RenewalSourceFacts, error) {
	f, e := Recovery5Facts()
	f.RenewalID = Recovery6RenewalID
	if e != nil {
		return f, e
	}
	f.FifthCohort, e = closedFifthCohort(f)
	if e != nil {
		return f, e
	}
	f.FifthCohortHistorySHA256 = fifthHistoryPin(f.FifthCohort)
	return f, nil
}
func fifthHistoryPin(h *FifthCohortHistory) string { raw, _ := json.Marshal(h); return Hash(raw) }

func approvedRecovery6(active bool) (RenewalGrant, string, error) {
	g, pin, base, e := loadCanonicalRenewalGrant(Recovery6RenewalID)
	if e != nil {
		return g, "", e
	}
	f, e := Recovery6Facts()
	if e != nil {
		return g, "", e
	}
	if g.Schema != 7 || g.RenewalID != Recovery6RenewalID || g.Decision != "APPROVED" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) || !sameRenewalHistory(g, f) || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !reflect.DeepEqual(g.FourthCohort, f.FourthCohort) || !reflect.DeepEqual(g.FifthCohort, f.FifthCohort) || !reflect.DeepEqual(g.SpendingGuardSourceSHA256, f.SpendingGuardSourceSHA256) || !fixedRenewalCaps(g) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("fixed recovery6 source/history/guard/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !regexpHex64(p) {
			return g, "", fmt.Errorf("recovery6 authority pin missing")
		}
	}
	if e = recovery6Approval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

// Recovery6Preflight remains read-only; BeginRenewal owns the process nonce.
func Recovery6Preflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(Recovery6RenewalID, route)
}

// A closed grant remains readable at its original source. This cannot admit a
// process or qualify an expired report against a different current source.
func closedFifthAuthority() (RenewalGrant, string, error) {
	f, e := Recovery6Facts()
	if e != nil {
		return RenewalGrant{}, "", e
	}
	g, pin, _, e := loadCanonicalRenewalGrant(Recovery5RenewalID)
	if e != nil {
		return g, "", e
	}
	if pin != f.FifthCohort.NativeAuthoritySHA256["grant.json"] || g.SourceCommit != f.FifthCohort.NativeSourceCommit {
		return g, "", fmt.Errorf("closed fifth authority changed after facts")
	}
	return g, pin, nil
}

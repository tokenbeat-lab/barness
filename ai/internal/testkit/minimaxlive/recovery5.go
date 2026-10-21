package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// Recovery5RenewalID is the fixed continuation of the closed fourth campaign.
// It does not make arbitrary rounds selectable or recycle consumed rights.
const Recovery5RenewalID = "provider-sites-final-acceptance-5"
const fourthCohortSeal = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-4/seal.json"
const fourthStopIndex = ".evidence/provider-sites/final-acceptance-20261020/recovery4-stopped/index.json"
const fourthCollection = ".evidence/provider-sites/final-acceptance-20261020/recovery4/"
const fourthActualApproval = ".evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json"

// FourthCohortHistory keeps recorder attempts separate from transport guard
// reservations. Sorted JSON field names match the launcher's compact digest.
type FourthCohortHistory struct {
	ActualHumanApprovalSHA256      string            `json:"actualHumanApprovalSha256"`
	CohortSealSHA256               string            `json:"cohortSealSha256"`
	FailedNativeRecordedAttempts   int               `json:"failedNativeRecordedAttempts"`
	FailedNativeRecordedCalls      int               `json:"failedNativeRecordedCalls"`
	FailedUSDReservationMicrousd   int64             `json:"failedUSDReservationMicrousd"`
	FailedUSDTransportReservations int               `json:"failedUSDTransportReservations"`
	NativeAttempts                 int               `json:"nativeAttempts"`
	NativeAuthoritySHA256          map[string]string `json:"nativeAuthoritySha256"`
	NativeCalls                    int               `json:"nativeCalls"`
	NativeLedgerRows               int               `json:"nativeLedgerRows"`
	NativeLedgerTailSHA256         string            `json:"nativeLedgerTailSha256"`
	NativeSourceCommit             string            `json:"nativeSourceCommit"`
	NewActualAttempts              int               `json:"newActualAttempts"`
	NewActualCalls                 int               `json:"newActualCalls"`
	NewClaimedAttempts             int               `json:"newClaimedAttempts"`
	NewClaimedCalls                int               `json:"newClaimedCalls"`
	NewReservedMicrousd            int64             `json:"newReservedMicrousd"`
	ParentClaimsSHA256             map[string]string `json:"parentClaimsSha256"`
	RetainedActualAttempts         int               `json:"retainedActualAttempts"`
	RetainedActualCalls            int               `json:"retainedActualCalls"`
	RetainedClaimedAttempts        int               `json:"retainedClaimedAttempts"`
	RetainedClaimedCalls           int               `json:"retainedClaimedCalls"`
	RetainedReservedMicrousd       int64             `json:"retainedReservedMicrousd"`
	RetainedUnknownParentProcesses int               `json:"retainedUnknownParentProcesses"`
	StopIndexSHA256                string            `json:"stopIndexSha256"`
}

func fourthCohortClosed() bool {
	owner, e := evidenceOwner()
	if e != nil {
		return true
	}
	_, e = os.Lstat(filepath.Join(owner, fourthCohortSeal))
	return !os.IsNotExist(e)
}

// Recovery5Facts validates all closed cohorts without reading a Key or issuing
// a grant. Current spending guard bytes differ from the retained fourth source.
func Recovery5Facts() (RenewalSourceFacts, error) {
	f, e := Recovery4Facts()
	f.RenewalID = Recovery5RenewalID
	if e != nil {
		return f, e
	}
	f.FourthCohort, e = closedFourthCohort(f)
	if e != nil {
		return f, e
	}
	f.FourthCohortHistorySHA256 = fourthHistoryPin(f.FourthCohort)
	source, e := ExecutionSource()
	if e != nil {
		return f, e
	}
	f.SpendingGuardSourceSHA256 = map[string]string{}
	for _, name := range []string{"ai/internal/testkit/finalbudget/guard.go", "ai/internal/testkit/finalbudget/estimate.go"} {
		if !regexpHex64(source[name]) {
			return f, fmt.Errorf("current spending guard source missing")
		}
		f.SpendingGuardSourceSHA256[name] = source[name]
	}
	return f, nil
}
func fourthHistoryPin(h *FourthCohortHistory) string { raw, _ := json.Marshal(h); return Hash(raw) }

func approvedRecovery5(active bool) (RenewalGrant, string, error) {
	g, pin, base, e := loadCanonicalRenewalGrant(Recovery5RenewalID)
	if e != nil {
		return g, "", e
	}
	f, e := Recovery5Facts()
	if e != nil {
		return g, "", e
	}
	if g.SixthCohort != nil || g.FifthCohort != nil || g.Schema != 6 || g.RenewalID != Recovery5RenewalID || g.Decision != "APPROVED" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) || !sameRenewalHistory(g, f) || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !reflect.DeepEqual(g.FourthCohort, f.FourthCohort) || !reflect.DeepEqual(g.SpendingGuardSourceSHA256, f.SpendingGuardSourceSHA256) || !fixedRenewalCaps(g) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("fixed recovery5 source/history/guard/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !regexpHex64(p) {
			return g, "", fmt.Errorf("recovery5 authority pin missing")
		}
	}
	if e = recovery5Approval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

// Recovery5Preflight remains read-only; BeginRenewal owns the process nonce.
func Recovery5Preflight(route string) (RenewalGrant, string, error) {
	return renewalPreflight(Recovery5RenewalID, route)
}

// A closed grant remains readable at its original source. This cannot admit a
// process or qualify an expired report against a different current source.
func closedFourthAuthority() (RenewalGrant, string, error) {
	f, e := Recovery5Facts()
	if e != nil {
		return RenewalGrant{}, "", e
	}
	g, pin, _, e := loadCanonicalRenewalGrant(Recovery4RenewalID)
	if e != nil {
		return g, "", e
	}
	if pin != f.FourthCohort.NativeAuthoritySHA256["grant.json"] || g.SourceCommit != f.FourthCohort.NativeSourceCommit {
		return g, "", fmt.Errorf("closed fourth authority changed after facts")
	}
	return g, pin, nil
}

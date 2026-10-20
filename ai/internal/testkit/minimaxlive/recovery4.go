package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// Recovery4RenewalID is the single fixed continuation after closed recovery3.
// A separate authorization preserves consumed rights; it is not a round registry.
const Recovery4RenewalID = "provider-sites-final-acceptance-4"
const thirdCohortSeal = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-3/seal.json"
const thirdStopIndex = ".evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json"
const thirdCollection = ".evidence/provider-sites/final-acceptance-20261020/recovery3/"
const thirdActualApproval = ".evidence/provider-sites/final-acceptance-20261020/recovery3-authorization.json"

// ThirdCohortHistory records the complete closed third cohort independently of
// RecoveryHistory's unchanged closed-second facts. Alphabetical fields preserve
// the launcher's compact sorted JSON digest, including sorted map keys.
type ThirdCohortHistory struct {
	ActualHumanApprovalSHA256      string            `json:"actualHumanApprovalSha256"`
	CohortSealSHA256               string            `json:"cohortSealSha256"`
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

func thirdCohortClosed() bool {
	owner, e := evidenceOwner()
	if e != nil {
		return true
	}
	_, e = os.Lstat(filepath.Join(owner, thirdCohortSeal))
	return !os.IsNotExist(e)
}

// Recovery4Facts is read-only and requires no new approval while preparing a plan.
func Recovery4Facts() (RenewalSourceFacts, error) {
	f, e := RecoveryFacts()
	f.RenewalID = Recovery4RenewalID
	if e != nil {
		return f, e
	}
	f.ThirdCohort, e = closedThirdCohort(f)
	if e == nil {
		f.ThirdCohortHistorySHA256 = thirdHistoryPin(f.ThirdCohort)
	}
	return f, e
}
func thirdHistoryPin(h *ThirdCohortHistory) string { raw, _ := json.Marshal(h); return Hash(raw) }

func approvedRecovery4(active bool) (RenewalGrant, string, error) {
	if !active && fourthCohortClosed() {
		return closedFourthAuthority()
	}

	g, pin, base, e := loadCanonicalRenewalGrant(Recovery4RenewalID)
	if e != nil {
		return g, "", e
	}
	f, e := Recovery4Facts()
	if e != nil {
		return g, "", e
	}
	if g.FifthCohort != nil || g.FourthCohort != nil || g.Schema != 5 || g.RenewalID != Recovery4RenewalID || g.Decision != "APPROVED" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) || !sameRenewalHistory(g, f) || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !fixedRenewalCaps(g) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("fixed recovery4 source/history/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !regexpHex64(p) {
			return g, "", fmt.Errorf("recovery4 authority pin missing")
		}
	}
	if e = recovery4Approval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}

// Only common canonical loading is shared. Fixed historical schema validation
// stays separate from current-source qualification and cannot select arbitrary IDs.
func loadCanonicalRenewalGrant(id string) (RenewalGrant, string, string, error) {
	var g RenewalGrant
	base, e := renewalBase(id)
	if e != nil {
		return g, "", "", e
	}
	if e = ownedRenewalDirectory(base); e != nil {
		return g, "", base, e
	}
	raw, e := ReadFile(filepath.Join(base, "grant.json"), 1<<20)
	if e != nil {
		return g, "", base, fmt.Errorf("canonical recovery grant unavailable")
	}
	if e = StrictJSON(raw, &g); e != nil {
		return g, "", base, e
	}
	return g, Hash(raw), base, nil
}

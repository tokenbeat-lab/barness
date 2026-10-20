package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
)

// RecoveryRenewalID is the fixed recovery from the failed second campaign.
// It is a new authorization boundary, never a replay or a round registry.
const RecoveryRenewalID = "provider-sites-final-acceptance-3"
const secondCohortSeal = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-2/seal.json"
const secondStopIndex = ".evidence/provider-sites/final-acceptance-20261020/remaining-stopped/index.json"

// RecoveryHistory keeps failed13/12 and aggregate spending separate from the
// two successful historical39/36 sets. Alphabetical fields make its JSON hash
// identical to the compact, sorted-key representation used by the launcher.
type RecoveryHistory struct {
	RetainedActualAttempts    int    `json:"retainedActualAttempts"`
	RetainedActualCalls       int    `json:"retainedActualCalls"`
	RetainedClaimedAttempts   int    `json:"retainedClaimedAttempts"`
	RetainedClaimedCalls      int    `json:"retainedClaimedCalls"`
	RetainedReservedMicrousd  int64  `json:"retainedReservedMicrousd"`
	SecondAttempts            int    `json:"secondAttempts"`
	SecondCalls               int    `json:"secondCalls"`
	SecondChatClaimSHA256     string `json:"secondChatClaimSha256"`
	SecondCohortSealSHA256    string `json:"secondCohortSealSha256"`
	SecondGrantSHA256         string `json:"secondGrantSha256"`
	SecondHumanApprovalSHA256 string `json:"secondHumanApprovalSha256"`
	SecondLedgerRows          int    `json:"secondLedgerRows"`
	SecondLedgerSHA256        string `json:"secondLedgerSha256"`
	SecondLedgerTailSHA256    string `json:"secondLedgerTailSha256"`
	SecondSourceReviewSHA256  string `json:"secondSourceReviewSha256"`
	StopIndexSHA256           string `json:"stopIndexSha256"`
}

// RecoveryFacts creates nothing. Absence of a new human approval is expected
// while preparing a concrete plan and cannot be mistaken for approval.
func RecoveryFacts() (RenewalSourceFacts, error) {
	f, e := RenewalFacts()
	f.RenewalID = RecoveryRenewalID
	if e != nil {
		return f, e
	}
	f.Recovery, e = closedSecondCohort(f)
	if e == nil {
		f.RecoveryHistorySHA256 = recoveryHistoryPin(f.Recovery)
	}
	return f, e
}
func evidenceOwner() (string, error) {
	old, e := originalStageLedgerPath("formal")
	if e != nil {
		return "", e
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(old))), nil
}
func secondCohortClosed() bool {
	owner, e := evidenceOwner()
	if e != nil {
		return true
	}
	_, e = os.Lstat(filepath.Join(owner, secondCohortSeal))
	return !os.IsNotExist(e)
}
func approvedRecoveryRenewal(active bool) (RenewalGrant, string, error) {
	g, pin, base, e := loadCanonicalRenewalGrant(RecoveryRenewalID)
	if e != nil {
		return g, "", e
	}
	f, e := RecoveryFacts()
	if e != nil {
		return g, "", e
	}
	if g.FifthCohort != nil || g.FourthCohort != nil || g.ThirdCohort != nil || g.Schema != 4 || g.RenewalID != RecoveryRenewalID || g.Decision != "APPROVED" || !regexp.MustCompile(`^[a-f0-9]{40}$`).MatchString(g.SourceCommit) || !sameRenewalHistory(g, f) || g.ExecutionSourceSHA256 != f.ExecutionSourceSHA256 || !reflect.DeepEqual(g.Recovery, f.Recovery) || !fixedRenewalCaps(g) || active && g.SourceCommit != f.SourceCommit {
		return g, "", fmt.Errorf("fixed recovery source/history/caps differ")
	}
	if e = renewalCommittedSource(g.SourceCommit); e != nil {
		return g, "", e
	}
	hex := regexp.MustCompile(`^[a-f0-9]{64}$`)
	for _, p := range []string{g.HumanApprovalSHA256, g.SourceReviewSHA256, g.SourceManifestSHA256, g.BinariesManifestSHA256, g.PlanSHA256, g.ClosedHistorySHA256, g.FirstCohortSealSHA256} {
		if !hex.MatchString(p) {
			return g, "", fmt.Errorf("recovery authority pin missing")
		}
	}
	if e = recoveryApproval(base, g); e != nil {
		return g, "", e
	}
	return g, pin, nil
}
func sameRenewalHistory(g RenewalGrant, f RenewalSourceFacts) bool {
	return g.CommonRootSHA256 == f.CommonRootSHA256 && g.OriginalLedgerSHA256 == f.OriginalLedgerSHA256 && g.OriginalLedgerTailSHA256 == f.OriginalLedgerTailSHA256 && g.OriginalCalls == f.OriginalCalls && g.OriginalAttempts == f.OriginalAttempts && g.PreviousRenewalID == f.PreviousRenewalID && g.PreviousLedgerSHA256 == f.PreviousLedgerSHA256 && g.PreviousLedgerTailSHA256 == f.PreviousLedgerTailSHA256 && g.PreviousCalls == f.PreviousCalls && g.PreviousAttempts == f.PreviousAttempts && g.ClosedHistorySHA256 == f.ClosedHistorySHA256 && sameFirstCohort(g, f)
}
func recoveryHistoryPin(h *RecoveryHistory) string { raw, _ := json.Marshal(h); return Hash(raw) }

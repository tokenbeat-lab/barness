package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Fifth stop includes its independently checked no-key bundles and binaries.
// Only this concrete batch has the larger bounds; earlier cohorts stay fixed.
const maxFifthCohortBytes = int64(640 << 20)

type fifthProcess struct {
	Name                string          `json:"name"`
	CanonicalClaim      string          `json:"canonicalClaim"`
	State               string          `json:"state"`
	ExitCode            int             `json:"exitCode"`
	ParentBudget        json.RawMessage `json:"parentBudget"`
	NativeRecordedCalls int             `json:"nativeRecordedCalls"`
	NativeRecorderHTTP  int             `json:"nativeRecorderHTTP"`
	Raw                 string          `json:"raw"`
	ReservedMicrousd    int64           `json:"reservedMicrousd"`
	FailedScenarios     []fifthFailure  `json:"failedScenarios,omitempty"`
}
type fifthFailure struct {
	ID            string `json:"id"`
	Outcome       string `json:"outcome"`
	ErrorCategory string `json:"errorCategory"`
}
type fifthNonRegular struct {
	Kind     string `json:"kind"`
	Mode     int    `json:"mode"`
	Target   string `json:"target"`
	Followed bool   `json:"followed"`
}
type fifthStop struct {
	Schema                     int                        `json:"schema"`
	Kind                       string                     `json:"kind"`
	State                      string                     `json:"state"`
	CreatedAt                  string                     `json:"createdAt"`
	SourceCommit               string                     `json:"sourceCommit"`
	ActualHumanApprovalSHA256  string                     `json:"actualHumanApprovalSHA256"`
	NewPaidProcesses           int                        `json:"newPaidProcesses"`
	PassedProcesses            int                        `json:"passedProcesses"`
	FailedProcesses            int                        `json:"failedProcesses"`
	NotRunProcesses            []string                   `json:"notRunProcesses"`
	Processes                  []fifthProcess             `json:"processes"`
	NewNativeRecordedCalls     int                        `json:"newNativeRecordedCalls"`
	NewNativeRecorderHTTP      int                        `json:"newNativeRecorderHTTP"`
	CampaignNativeCalls        int                        `json:"campaignNativeRecordedCalls"`
	CampaignNativeHTTP         int                        `json:"campaignNativeRecorderHTTP"`
	CampaignClaimedCalls       int                        `json:"campaignClaimedCalls"`
	CampaignClaimedHTTP        int                        `json:"campaignClaimedHTTP"`
	WholeReservedMicrousd      int64                      `json:"wholeReservedMicrousd"`
	MaxMicrousd                int64                      `json:"maxMicrousd"`
	RefundMicrousd             int64                      `json:"refundMicrousd"`
	UnknownParentProcesses     int                        `json:"unknownParentProcesses"`
	FailedParentRemainsUNKNOWN bool                       `json:"failedParentRemainsUNKNOWN"`
	AutomaticReruns            bool                       `json:"automaticReruns"`
	PaidContinuationAuthorized bool                       `json:"paidContinuationAuthorized"`
	CredentialsReadByThisAudit bool                       `json:"credentialsReadByThisAudit"`
	VendorCallsByThisAudit     int                        `json:"vendorCallsByThisAudit"`
	FormalREADY                string                     `json:"formalREADY"`
	Race                       string                     `json:"race"`
	ReleasePass                bool                       `json:"releasePass"`
	PredecessorStopIndex       string                     `json:"predecessorStopIndex"`
	PredecessorStopIndexSHA256 string                     `json:"predecessorStopIndexSHA256"`
	FileCount                  int                        `json:"fileCount"`
	FileBytes                  int64                      `json:"fileBytes"`
	NonRegular                 map[string]fifthNonRegular `json:"nonRegularRejectionFixtures"`
	Files                      map[string]string          `json:"files"`
}

var fifthNames = strings.Fields("formal-zai-responses mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses")
var fifthUnstarted = strings.Fields("formal-minimax-cn-messages glm-chat glm-responses glm-messages openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions")

func closedFifthCohort(f RenewalSourceFacts) (*FifthCohortHistory, error) {
	owner, e := evidenceOwner()
	if e != nil {
		return nil, e
	}
	root, e := os.OpenRoot(owner)
	if e != nil {
		return nil, e
	}
	defer root.Close()
	var seal secondSeal
	raw, e := cohortJSON(root, fifthCohortSeal, &seal)
	if e != nil || seal.Schema != 1 || seal.Kind != "CLOSED_FIFTH_COHORT_HISTORY_SEAL" || seal.RoundID != Recovery5RenewalID || seal.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || seal.StopIndex != filepath.Join(owner, fifthStopIndex) || !seal.NoAutomaticReruns || seal.PaidContinuationAuthorized || seal.ActualCalls != 702 || seal.ActualHTTPAttempts != 692 || seal.ClaimedCalls != 1160 || seal.ClaimedHTTPAttempts != 1232 || seal.WholeReservedMicrousd != 32050000 || seal.RefundMicrousd != 0 {
		return nil, fmt.Errorf("fixed closed fifth702/692,1160/1232,32.05USD required: %v", e)
	}
	h := &FifthCohortHistory{CohortSealSHA256: Hash(raw), RetainedActualCalls: 702, RetainedActualAttempts: 692, RetainedClaimedCalls: 1160, RetainedClaimedAttempts: 1232, RetainedReservedMicrousd: 32050000, RetainedUnknownParentProcesses: 5, NewActualCalls: 111, NewActualAttempts: 108, NewClaimedCalls: 180, NewClaimedAttempts: 192, NewReservedMicrousd: 3250000, FailedNativeRecordedCalls: 9, FailedNativeRecordedAttempts: 9, FailedUSDReservationMicrousd: 54238, FailedUSDTransportReservations: 9, NativeAuthoritySHA256: map[string]string{}, ParentClaimsSHA256: map[string]string{}}
	var stop fifthStop
	raw, e = cohortJSONLimit(root, fifthStopIndex, &stop, 16<<20)
	if e != nil || Hash(raw) != seal.StopIndexSHA256 || stop.Schema != 1 || stop.Kind != "ACTUAL_FIXED5_STOP_AFTER_MINIMAX_CN_RESPONSES_NATIVE_HISTORY_FAILURE" || stop.State != "STOPPED_NO_FURTHER_PAID_LAUNCH" || stop.NewPaidProcesses != 8 || stop.PassedProcesses != 7 || stop.FailedProcesses != 1 || len(stop.Processes) != 8 || !slices.Equal(stop.NotRunProcesses, fifthUnstarted) || stop.NewNativeRecordedCalls != 111 || stop.NewNativeRecorderHTTP != 108 || stop.CampaignNativeCalls != 702 || stop.CampaignNativeHTTP != 692 || stop.CampaignClaimedCalls != 1160 || stop.CampaignClaimedHTTP != 1232 || stop.WholeReservedMicrousd != 32050000 || stop.MaxMicrousd != 36000000 || stop.RefundMicrousd != 0 || stop.UnknownParentProcesses != 5 || !stop.FailedParentRemainsUNKNOWN || stop.AutomaticReruns || stop.PaidContinuationAuthorized || stop.CredentialsReadByThisAudit || stop.VendorCallsByThisAudit != 0 || stop.FormalREADY != "NOT_RUN" || stop.Race != "NOT_RUN" || stop.ReleasePass || stop.PredecessorStopIndex != fourthStopIndex || stop.PredecessorStopIndexSHA256 != f.FourthCohort.StopIndexSHA256 || stop.FileCount != 31152 || len(stop.Files) != 31152 || stop.FileBytes < 1 || stop.FileBytes > maxFifthCohortBytes {
		return nil, fmt.Errorf("fixed closed fifth31152/8 history differs: %v", e)
	}
	h.StopIndexSHA256, h.NativeSourceCommit, h.ActualHumanApprovalSHA256 = Hash(raw), stop.SourceCommit, stop.ActualHumanApprovalSHA256
	if e = fifthPinnedSelection(root, owner, stop); e != nil {
		return nil, e
	}
	total := int64(0)
	for name, want := range stop.Files {
		got, n, e := cohortFileHash(root, name, maxFifthCohortBytes-total)
		if e != nil || got != want {
			return nil, fmt.Errorf("fifth cohort pinned file differs: %s: %v", name, e)
		}
		total += n
	}
	if total != stop.FileBytes {
		return nil, fmt.Errorf("closed fifth total bytes differ")
	}
	if e = fifthNonRegularHistory(root, owner, stop); e != nil {
		return nil, e
	}
	if e = fifthParentsSelection(root); e != nil {
		return nil, e
	}
	if e = fifthNativeHistory(root, owner, stop, f, h); e != nil {
		return nil, e
	}
	return h, nil
}

// The one intentional rejection fixture is verified as link metadata only.
// Opening or hashing its target would turn a negative fixture into trusted data.
func fifthNonRegularHistory(root *os.Root, owner string, stop fifthStop) error {
	if len(stop.NonRegular) != 1 {
		return fmt.Errorf("closed fifth exact one rejection fixture required")
	}
	for name, n := range stop.NonRegular {
		prefix := fifthCollection + "no-key-live/"
		suffix := "/formal-corrupt-symlink/formal-integrity.json"
		middle := strings.TrimSuffix(strings.TrimPrefix(name, prefix), suffix)
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) || middle == "" || strings.Contains(middle, "/") || n.Kind != "symlink" || n.Followed || n.Mode != 41453 || n.Target != filepath.Join(owner, strings.TrimSuffix(name, "formal-integrity.json")+"seal.json") || stop.Files[name] != "" {
			return fmt.Errorf("closed fifth rejection link metadata differs")
		}
		info, e := root.Lstat(name)
		if e != nil || info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("closed fifth negative link changed")
		}
		target, e := root.Readlink(name)
		if e != nil || target != n.Target {
			return fmt.Errorf("closed fifth negative link target changed")
		}
	}
	return nil
}

func fifthParentsSelection(root *os.Root) error {
	dir, e := root.Open(".evidence/minimax-launches")
	if e != nil {
		return e
	}
	defer dir.Close()
	entries, e := dir.ReadDir(4097)
	if e != nil || len(entries) > 4096 {
		return fmt.Errorf("closed fifth launch selection unavailable")
	}
	want := map[string]bool{}
	for _, n := range fifthNames {
		want[Recovery5RenewalID+"-"+n+".json"] = true
	}
	found := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), Recovery5RenewalID+"-") {
			if !want[entry.Name()] {
				return fmt.Errorf("closed fifth acquired extra canonical claim")
			}
			found++
		}
	}
	if found != 8 {
		return fmt.Errorf("closed fifth exact8 canonical claims required")
	}
	return nil
}

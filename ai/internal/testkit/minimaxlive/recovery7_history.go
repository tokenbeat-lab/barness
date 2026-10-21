package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// SixthCohortHistory keeps report-backed calls separate from the interrupted
// ledger's reservations. Recorded HTTP allocations do not prove vendor sends.
// Sorted JSON field names match the independent controller's compact digest.
type SixthCohortHistory struct {
	ActualHumanApprovalSHA256            string            `json:"actualHumanApprovalSHA256"`
	CohortSealSHA256                     string            `json:"cohortSealSha256"`
	InterruptedNativeCalls               int               `json:"interruptedNativeCalls"`
	InterruptedNativeHTTPAllocations     int               `json:"interruptedNativeHTTPAllocations"`
	InterruptedNativeLedgerRows          int               `json:"interruptedNativeLedgerRows"`
	InterruptedNativeLedgerTailSHA256    string            `json:"interruptedNativeLedgerTailSHA256"`
	InterruptedUSDReservationMicrousd    int64             `json:"interruptedUSDReservationMicrousd"`
	InterruptedUSDTransportReservations  int               `json:"interruptedUSDTransportReservations"`
	NativeAuthoritySHA256                map[string]string `json:"nativeAuthoritySHA256"`
	NativeSourceCommit                   string            `json:"nativeSourceCommit"`
	NewClaimedAttempts                   int               `json:"newClaimedAttempts"`
	NewClaimedCalls                      int               `json:"newClaimedCalls"`
	NewCompletedReportCalls              int               `json:"newCompletedReportCalls"`
	NewCompletedReportHTTP               int               `json:"newCompletedReportHTTP"`
	NewReservedMicrousd                  int64             `json:"newReservedMicrousd"`
	ParentClaimsSHA256                   map[string]string `json:"parentClaimsSHA256"`
	RetainedClaimedAttempts              int               `json:"retainedClaimedAttempts"`
	RetainedClaimedCalls                 int               `json:"retainedClaimedCalls"`
	RetainedCompletedReportKnownCalls    int               `json:"retainedCompletedReportKnownCalls"`
	RetainedCompletedReportKnownHTTP     int               `json:"retainedCompletedReportKnownHTTP"`
	RetainedKnownRecordedCalls           int               `json:"retainedKnownRecordedCalls"`
	RetainedKnownRecordedHTTPAllocations int               `json:"retainedKnownRecordedHTTPAllocations"`
	RetainedReservedMicrousd             int64             `json:"retainedReservedMicrousd"`
	RetainedUnknownParentProcesses       int               `json:"retainedUnknownParentProcesses"`
	StopIndexSHA256                      string            `json:"stopIndexSha256"`
}

type sixthSeal struct {
	Schema                                            int    `json:"schema"`
	Kind                                              string `json:"kind"`
	RoundID                                           string `json:"roundId"`
	State                                             string `json:"state"`
	StopIndex                                         string `json:"stopIndex"`
	StopIndexSHA256                                   string `json:"stopIndexSHA256"`
	KnownRecordedCalls                                int    `json:"knownRecordedCalls"`
	KnownRecordedHTTPAllocations                      int    `json:"knownRecordedHTTPAllocations"`
	CompletedReportKnownCalls                         int    `json:"completedReportKnownCalls"`
	CompletedReportKnownHTTP                          int    `json:"completedReportKnownHTTP"`
	InterruptedNativeLedgerCalls                      int    `json:"interruptedNativeLedgerCalls"`
	InterruptedNativeLedgerHTTPAllocations            int    `json:"interruptedNativeLedgerHTTPAllocations"`
	InterruptedParentBudget                           string `json:"interruptedParentBudget"`
	InterruptedFinalReportAbsent                      bool   `json:"interruptedFinalReportAbsent"`
	RecordedHTTPAllocationsAreNotConfirmedVendorSends bool   `json:"recordedHTTPAllocationsAreNotConfirmedVendorSends"`
	ClaimedCalls                                      int    `json:"claimedCalls"`
	ClaimedHTTPAttempts                               int    `json:"claimedHTTPAttempts"`
	WholeReservedMicrousd                             int64  `json:"wholeReservedMicrousd"`
	UnknownParentProcesses                            int    `json:"unknownParentProcesses"`
	RefundMicrousd                                    int64  `json:"refundMicrousd"`
	NoAutomaticReruns                                 bool   `json:"noAutomaticReruns"`
	PaidContinuationAuthorized                        bool   `json:"paidContinuationAuthorized"`
}
type sixthProcess struct {
	Name                        string          `json:"name"`
	Parent                      string          `json:"parent"`
	ParentSHA256                string          `json:"parentSHA256"`
	ParentBudget                json.RawMessage `json:"parentBudget"`
	ExitCode                    int             `json:"exitCode"`
	ReservedMicrousd            int64           `json:"reservedMicrousd"`
	NativeConsumerReceiptSHA256 *string         `json:"nativeConsumerReceiptSHA256"`
}
type sixthStop struct {
	Schema                              int    `json:"schema"`
	Kind                                string `json:"kind"`
	State                               string `json:"state"`
	SourceCommit                        string `json:"sourceCommit"`
	NewPaidProcesses                    int    `json:"newPaidProcesses"`
	PassedProcesses                     int    `json:"passedProcesses"`
	InterruptedAfterFailedCaseProcesses int    `json:"interruptedAfterFailedCaseProcesses"`
	CompletedFullM3Processes            int    `json:"completedFullM3Processes"`
	M3FinalReport                       string `json:"m3FinalReport"`
	M3Raw                               string `json:"m3Raw"`
	FailedFinalCase                     string `json:"failedFinalCase"`
	Failure                             struct {
		Case       string `json:"case"`
		Status     string `json:"status"`
		ErrorCode  string `json:"errorCode"`
		ErrorPhase string `json:"errorPhase"`
	} `json:"failure"`
	ActualHumanApprovalSHA256                               string                     `json:"actualHumanApprovalSHA256"`
	NewCompletedReportCalls                                 int                        `json:"newCompletedReportCalls"`
	NewCompletedReportHTTP                                  int                        `json:"newCompletedReportHTTP"`
	CampaignCompletedReportKnownCalls                       int                        `json:"campaignCompletedReportKnownCalls"`
	CampaignCompletedReportKnownHTTP                        int                        `json:"campaignCompletedReportKnownHTTP"`
	InterruptedM3NativeLedgerCalls                          int                        `json:"interruptedM3NativeLedgerCalls"`
	InterruptedM3NativeLedgerHTTPAllocations                int                        `json:"interruptedM3NativeLedgerHTTPAllocations"`
	InterruptedM3NativeLedgerRows                           int                        `json:"interruptedM3NativeLedgerRows"`
	InterruptedM3NativeLedgerTail                           string                     `json:"interruptedM3NativeLedgerTail"`
	InterruptedM3USDReservationMicrousd                     int64                      `json:"interruptedM3USDReservationMicrousd"`
	InterruptedM3USDTransportReservations                   int                        `json:"interruptedM3USDTransportReservations"`
	InterruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends bool                       `json:"interruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends"`
	CampaignClaimedCalls                                    int                        `json:"campaignClaimedCalls"`
	CampaignClaimedHTTPAttempts                             int                        `json:"campaignClaimedHTTPAttempts"`
	WholeReservedMicrousd                                   int64                      `json:"wholeReservedMicrousd"`
	RefundMicrousd                                          int64                      `json:"refundMicrousd"`
	UnknownParentProcesses                                  int                        `json:"unknownParentProcesses"`
	FailedParentRemainsUNKNOWN                              bool                       `json:"failedParentRemainsUNKNOWN"`
	StopSignal                                              string                     `json:"stopSignal"`
	ChildNumericExitCode                                    int                        `json:"childNumericExitCode"`
	PublicOperatorNumericExitCode                           int                        `json:"publicOperatorNumericExitCode"`
	AutomaticReruns                                         bool                       `json:"automaticReruns"`
	PaidContinuationAuthorized                              bool                       `json:"paidContinuationAuthorized"`
	FormalREADY                                             string                     `json:"formalREADY"`
	Race                                                    string                     `json:"race"`
	ReleasePass                                             bool                       `json:"releasePass"`
	Processes                                               []sixthProcess             `json:"processes"`
	NotRunProcesses                                         []string                   `json:"notRunProcesses"`
	FileCount                                               int                        `json:"fileCount"`
	FileBytes                                               int64                      `json:"fileBytes"`
	Files                                                   map[string]string          `json:"files"`
	NonRegular                                              map[string]fifthNonRegular `json:"nonRegularRejectionFixtures"`
	CreatedAt                                               string                     `json:"createdAt"`
}

var sixthNames = strings.Fields("formal-zai-responses mm-global-formal-m3-chat")
var sixthUnstarted = strings.Fields("mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages glm-chat glm-responses glm-messages openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions")

// Only the concrete sixth stop includes 566,508,450 bytes of frozen binaries
// and no-key proofs. These ceilings do not broaden any earlier JSON/tree input.
const maxSixthCohortBytes = int64(640 << 20)

func closedSixthCohort(f RenewalSourceFacts) (*SixthCohortHistory, error) {
	owner, e := evidenceOwner()
	if e != nil {
		return nil, e
	}
	root, e := os.OpenRoot(owner)
	if e != nil {
		return nil, e
	}
	defer root.Close()
	var seal sixthSeal
	raw, e := cohortJSON(root, sixthCohortSeal, &seal)
	if e != nil || seal.Schema != 1 || seal.Kind != "CLOSED_SIXTH_COHORT_HISTORY_SEAL" || seal.RoundID != Recovery6RenewalID || seal.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || seal.StopIndex != filepath.Join(owner, sixthStopIndex) || !seal.NoAutomaticReruns || seal.PaidContinuationAuthorized || seal.KnownRecordedCalls != 724 || seal.KnownRecordedHTTPAllocations != 714 || seal.CompletedReportKnownCalls != 720 || seal.CompletedReportKnownHTTP != 710 || seal.InterruptedNativeLedgerCalls != 4 || seal.InterruptedNativeLedgerHTTPAllocations != 4 || seal.InterruptedParentBudget != "UNKNOWN" || !seal.InterruptedFinalReportAbsent || !seal.RecordedHTTPAllocationsAreNotConfirmedVendorSends || seal.ClaimedCalls != 1204 || seal.ClaimedHTTPAttempts != 1280 || seal.WholeReservedMicrousd != 32900000 || seal.UnknownParentProcesses != 6 || seal.RefundMicrousd != 0 {
		return nil, fmt.Errorf("fixed closed sixth report720/710+interrupted4/4,1204/1280,32.90USD required: %v", e)
	}
	h := &SixthCohortHistory{CohortSealSHA256: Hash(raw), RetainedCompletedReportKnownCalls: 720, RetainedCompletedReportKnownHTTP: 710, RetainedKnownRecordedCalls: 724, RetainedKnownRecordedHTTPAllocations: 714, RetainedClaimedCalls: 1204, RetainedClaimedAttempts: 1280, RetainedReservedMicrousd: 32900000, RetainedUnknownParentProcesses: 6, NewCompletedReportCalls: 18, NewCompletedReportHTTP: 18, NewClaimedCalls: 44, NewClaimedAttempts: 48, NewReservedMicrousd: 850000, InterruptedNativeCalls: 4, InterruptedNativeHTTPAllocations: 4, InterruptedNativeLedgerRows: 8, InterruptedUSDReservationMicrousd: 12884, InterruptedUSDTransportReservations: 4, NativeAuthoritySHA256: map[string]string{}, ParentClaimsSHA256: map[string]string{}}
	var stop sixthStop
	raw, e = cohortJSONLimit(root, sixthStopIndex, &stop, 16<<20)
	if e != nil || Hash(raw) != seal.StopIndexSHA256 || stop.Schema != 1 || stop.Kind != "ACTUAL_FIXED6_STOP_AFTER_M3_CHAT_FINAL_REQUEST_DEADLINE_CASE_FAILURE" || stop.State != "STOPPED_NO_FURTHER_PAID_LAUNCH" || stop.NewPaidProcesses != 2 || stop.PassedProcesses != 1 || stop.InterruptedAfterFailedCaseProcesses != 1 || stop.CompletedFullM3Processes != 0 || stop.M3FinalReport != "ABSENT_INTERRUPTED_AFTER_IRREVERSIBLE_CASE_FAILURE" || stop.Failure.Case != "full-stream-rerun-1" || stop.Failure.Status != "FAIL" || stop.Failure.ErrorCode != "deadline_exceeded" || stop.Failure.ErrorPhase != "request" || stop.NewCompletedReportCalls != 18 || stop.NewCompletedReportHTTP != 18 || stop.CampaignCompletedReportKnownCalls != 720 || stop.CampaignCompletedReportKnownHTTP != 710 || stop.InterruptedM3NativeLedgerCalls != 4 || stop.InterruptedM3NativeLedgerHTTPAllocations != 4 || stop.InterruptedM3NativeLedgerRows != 8 || !regexpHex64(stop.InterruptedM3NativeLedgerTail) || stop.InterruptedM3USDReservationMicrousd != 12884 || stop.InterruptedM3USDTransportReservations != 4 || !stop.InterruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends || stop.CampaignClaimedCalls != 1204 || stop.CampaignClaimedHTTPAttempts != 1280 || stop.WholeReservedMicrousd != 32900000 || stop.RefundMicrousd != 0 || stop.UnknownParentProcesses != 6 || !stop.FailedParentRemainsUNKNOWN || stop.StopSignal != "SIGTERM" || stop.ChildNumericExitCode != -15 || stop.PublicOperatorNumericExitCode != 1 || stop.AutomaticReruns || stop.PaidContinuationAuthorized || stop.FormalREADY != "NOT_RUN" || stop.Race != "NOT_RUN" || stop.ReleasePass || len(stop.Processes) != 2 || !slices.Equal(stop.NotRunProcesses, sixthUnstarted) || stop.FileCount != 30238 || len(stop.Files) != 30238 || stop.FileBytes < 1 || stop.FileBytes > maxSixthCohortBytes {
		return nil, fmt.Errorf("fixed closed sixth30238/2 partial history differs: %v", e)
	}
	h.StopIndexSHA256, h.NativeSourceCommit, h.ActualHumanApprovalSHA256 = Hash(raw), stop.SourceCommit, stop.ActualHumanApprovalSHA256
	h.InterruptedNativeLedgerTailSHA256 = stop.InterruptedM3NativeLedgerTail
	if e = sixthPinnedSelection(root, owner, stop); e != nil {
		return nil, e
	}
	total := int64(0)
	for name, want := range stop.Files {
		got, n, e := cohortFileHash(root, name, maxSixthCohortBytes-total)
		if e != nil || got != want {
			return nil, fmt.Errorf("sixth cohort pinned file differs: %s: %v", name, e)
		}
		total += n
	}
	if total != stop.FileBytes {
		return nil, fmt.Errorf("closed sixth total bytes differ")
	}
	if e = sixthNativeHistory(root, owner, stop, f, h); e != nil {
		return nil, e
	}
	return h, nil
}

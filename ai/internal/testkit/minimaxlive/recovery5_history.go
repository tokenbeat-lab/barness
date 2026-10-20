package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
)

type fourthProcess struct {
	Name                        string          `json:"name"`
	ExitCode                    int             `json:"exitCode"`
	ParentBudget                json.RawMessage `json:"parentBudget"`
	ReservedMicrousd            int64           `json:"reservedMicrousd"`
	NativeBundle                string          `json:"nativeBundle"`
	NativeReportSHA256          string          `json:"nativeReportSHA256"`
	ActualCalls                 int             `json:"actualCalls"`
	ActualHTTP                  int             `json:"actualHTTP"`
	NetworkRoute                string          `json:"networkRoute"`
	NativeOutcome               string          `json:"nativeOutcome,omitempty"`
	NativeComplete              bool            `json:"nativeComplete,omitempty"`
	NativeAuditPassed           bool            `json:"nativeAuditPassed,omitempty"`
	NativeConsumerReceiptSHA256 string          `json:"nativeConsumerReceiptSHA256,omitempty"`
	ScenarioOutcomes            map[string]int  `json:"scenarioOutcomes,omitempty"`
	FailedScenarios             []string        `json:"failedScenarios,omitempty"`
	USDLedgerAttempts           int             `json:"usdLedgerAttempts"`
	USDReservedMicrousd         int64           `json:"usdReservedMicrousd"`
}
type fourthStop struct {
	Schema                              int               `json:"schema"`
	Kind                                string            `json:"kind"`
	SourceCommit                        string            `json:"sourceCommit"`
	CreatedAt                           string            `json:"createdAt"`
	ActualHumanApprovalSHA256           string            `json:"actualHumanApprovalSHA256"`
	SuccessfulProcesses                 int               `json:"successfulProcesses"`
	FailedProcesses                     int               `json:"failedProcesses"`
	NewActualCalls                      int               `json:"newActualCalls"`
	NewActualHTTP                       int               `json:"newActualHTTP"`
	CampaignActualCalls                 int               `json:"campaignActualCalls"`
	CampaignActualHTTP                  int               `json:"campaignActualHTTP"`
	CampaignClaimedCalls                int               `json:"campaignClaimedCalls"`
	CampaignClaimedHTTP                 int               `json:"campaignClaimedHTTP"`
	NewWholeReservedMicrousd            int64             `json:"newWholeReservedMicrousd"`
	WholeReservedMicrousd               int64             `json:"wholeReservedMicrousd"`
	RefundMicrousd                      int64             `json:"refundMicrousd"`
	UnknownParentProcesses              int               `json:"unknownParentProcesses"`
	OriginalFailedGLMUnchanged          bool              `json:"originalFailedGLMUnchanged"`
	OriginalFailedM3Unchanged           bool              `json:"originalFailedM3Unchanged"`
	AutomaticReruns                     bool              `json:"automaticReruns"`
	UnstartedProcesses                  []string          `json:"unstartedProcesses"`
	Files                               map[string]string `json:"files"`
	NewREADY                            bool              `json:"newREADY"`
	Race                                string            `json:"race"`
	ReleasePass                         bool              `json:"releasePass"`
	CredentialsReadByThisAudit          bool              `json:"credentialsReadByThisAudit"`
	Processes                           []fourthProcess   `json:"processes"`
	OriginalFailedZAIResponsesUnchanged bool              `json:"originalFailedZAIResponsesUnchanged"`
	FailedNativeRecordedCalls           int               `json:"failedNativeRecordedCalls"`
	FailedNativeRecordedHTTP            int               `json:"failedNativeRecordedHTTP"`
	FailedUSDReservationMicrousd        int64             `json:"failedUSDReservationMicrousd"`
	FailedUSDTransportReservations      int               `json:"failedUSDTransportReservations"`
	FileCount                           int               `json:"fileCount"`
	FileBytes                           int64             `json:"fileBytes"`
	HTTPCountingNote                    string            `json:"httpCountingNote"`
	PriorStopIndex                      string            `json:"priorStopIndex"`
	PriorStopIndexSHA256                string            `json:"priorStopIndexSHA256"`
}

var fourthUnstarted = strings.Fields("typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions")
var fourthNames = strings.Fields("formal-zai-responses mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions")

func closedFourthCohort(f RenewalSourceFacts) (*FourthCohortHistory, error) {
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
	raw, e := cohortJSON(root, fourthCohortSeal, &seal)
	if e != nil || seal.Schema != 1 || seal.Kind != "CLOSED_FOURTH_COHORT_HISTORY_SEAL" || seal.RoundID != Recovery4RenewalID || seal.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || seal.StopIndex != filepath.Join(owner, fourthStopIndex) || !seal.NoAutomaticReruns || seal.PaidContinuationAuthorized || seal.ActualCalls != 591 || seal.ActualHTTPAttempts != 584 || seal.ClaimedCalls != 980 || seal.ClaimedHTTPAttempts != 1040 || seal.WholeReservedMicrousd != 28800000 || seal.RefundMicrousd != 0 {
		return nil, fmt.Errorf("fixed closed fourth591/584,980/1040,28.8USD required: %v", e)
	}
	h := &FourthCohortHistory{CohortSealSHA256: Hash(raw), RetainedActualCalls: 591, RetainedActualAttempts: 584, RetainedClaimedCalls: 980, RetainedClaimedAttempts: 1040, RetainedReservedMicrousd: 28800000, RetainedUnknownParentProcesses: 4, NewActualCalls: 213, NewActualAttempts: 210, NewClaimedCalls: 416, NewClaimedAttempts: 428, NewReservedMicrousd: 16200000, FailedNativeRecordedCalls: 9, FailedNativeRecordedAttempts: 9, NativeAuthoritySHA256: map[string]string{}, ParentClaimsSHA256: map[string]string{}}
	var stop fourthStop
	raw, e = cohortJSON(root, fourthStopIndex, &stop)
	if e != nil || Hash(raw) != seal.StopIndexSHA256 || stop.Schema != 1 || stop.Kind != "ACTUAL_RECOVERY4_STOP_AFTER_OPENAI_DECISIONS_PRESEND_BUDGET_FAILURE" || stop.SuccessfulProcesses != 17 || stop.FailedProcesses != 1 || stop.NewActualCalls != 213 || stop.NewActualHTTP != 210 || stop.CampaignActualCalls != 591 || stop.CampaignActualHTTP != 584 || stop.CampaignClaimedCalls != 980 || stop.CampaignClaimedHTTP != 1040 || stop.NewWholeReservedMicrousd != 16200000 || stop.WholeReservedMicrousd != 28800000 || stop.RefundMicrousd != 0 || stop.UnknownParentProcesses != 4 || !stop.OriginalFailedGLMUnchanged || !stop.OriginalFailedM3Unchanged || !stop.OriginalFailedZAIResponsesUnchanged || stop.AutomaticReruns || stop.NewREADY || stop.ReleasePass || stop.Race != "NOT_RUN" || stop.CredentialsReadByThisAudit || len(stop.Files) != 2032 || stop.FileCount != 2032 || len(stop.Processes) != 18 || !slices.Equal(stop.UnstartedProcesses, fourthUnstarted) || stop.FailedNativeRecordedCalls != 9 || stop.FailedNativeRecordedHTTP != 9 || stop.FailedUSDReservationMicrousd != 0 || stop.FailedUSDTransportReservations != 0 || stop.HTTPCountingNote == "" || stop.PriorStopIndex != thirdStopIndex || stop.PriorStopIndexSHA256 != f.ThirdCohort.StopIndexSHA256 {
		return nil, fmt.Errorf("fixed closed fourth2032/18 history differs: %v", e)
	}
	h.StopIndexSHA256 = Hash(raw)
	h.NativeSourceCommit = stop.SourceCommit
	h.ActualHumanApprovalSHA256 = stop.ActualHumanApprovalSHA256
	total := int64(0)
	for name, want := range stop.Files {
		got, n, e := cohortFileHash(root, name, maxCohortBytes-total)
		if e != nil || got != want {
			return nil, fmt.Errorf("fourth cohort pinned file differs: %s: %v", name, e)
		}
		total += n
	}
	if total != stop.FileBytes {
		return nil, fmt.Errorf("closed fourth total bytes differ")
	}
	if e = fourthParentsSelection(root); e != nil {
		return nil, e
	}
	if e = fourthNativeHistory(root, owner, stop, f, h); e != nil {
		return nil, e
	}
	return h, nil
}
func fourthParentsSelection(root *os.Root) error {
	dir, e := root.Open(".evidence/minimax-launches")
	if e != nil {
		return e
	}
	defer dir.Close()
	entries, e := dir.ReadDir(4097)
	if e != nil || len(entries) > 4096 {
		return fmt.Errorf("closed fourth launch selection unavailable")
	}
	want := map[string]bool{}
	for _, n := range fourthNames {
		want[Recovery4RenewalID+"-"+n+".json"] = true
	}
	found := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), Recovery4RenewalID+"-") {
			if !want[entry.Name()] {
				return fmt.Errorf("closed fourth acquired extra canonical claim")
			}
			found++
		}
	}
	if found != 18 {
		return fmt.Errorf("closed fourth exact18 canonical claims required")
	}
	return nil
}
func fourthNativeHistory(root *os.Root, owner string, stop fourthStop, f RenewalSourceFacts, h *FourthCohortHistory) error {
	const base = ".evidence/minimax-renewals/" + Recovery4RenewalID + "/"
	dir, e := root.Open(strings.TrimSuffix(base, "/"))
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(10)
	dir.Close()
	if e != nil || len(entries) != 9 {
		return fmt.Errorf("closed fourth exact9 native authority required")
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if !slices.Equal(names, thirdNativeNames) {
		return fmt.Errorf("closed fourth native authority file set differs")
	}
	for _, name := range names {
		pin := stop.Files[base+name]
		if !regexpHex64(pin) {
			return fmt.Errorf("closed fourth native authority unpinned")
		}
		h.NativeAuthoritySHA256[name] = pin
	}
	var g RenewalGrant
	raw, e := cohortJSON(root, base+"grant.json", &g)
	if e != nil || g.Schema != 5 || !oldFourthGrantOnly(g) || g.RenewalID != Recovery4RenewalID || g.Decision != "APPROVED" || !fixedRenewalCaps(g) || !sameRenewalHistory(g, f) || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || g.SourceCommit != stop.SourceCommit || g.HumanApprovalSHA256 != stop.Files[base+"human-approval.json"] || g.SourceReviewSHA256 != stop.Files[base+"source-review.json"] {
		return fmt.Errorf("closed fourth original schema5 grant/history differs: %v", e)
	}
	pin := Hash(raw)
	if e = recovery4Approval(filepath.Join(owner, base), g); e != nil {
		return fmt.Errorf("closed fourth original approval: %w", e)
	}
	a, e := readActualRecoveryApproval(filepath.Join(owner, fourthActualApproval), stop.ActualHumanApprovalSHA256, Recovery4RenewalID, 1070, 1130, 17300000, 12600000)
	if e != nil || stop.Files[fourthActualApproval] != stop.ActualHumanApprovalSHA256 || a.SourceCommit != g.SourceCommit || a.SourceManifestSHA256 != g.SourceManifestSHA256 || a.BinariesManifestSHA256 != g.BinariesManifestSHA256 || a.PlanSHA256 != g.PlanSHA256 {
		return fmt.Errorf("closed fourth actual human/source differs: %v", e)
	}
	if e = closedHistoricalSource(root, stop.Files, fourthCollection, g); e != nil {
		return e
	}
	rows, e := ReadLedger(filepath.Join(owner, base, "formal.jsonl"))
	if e != nil || len(rows) != 75 {
		return fmt.Errorf("closed fourth exact75 ledger required: %v", e)
	}
	sessions, e := closedFormalSessions(rows, Recovery4RenewalID, pin)
	if e != nil {
		return e
	}
	budgets := map[string]Budget{}
	for _, r := range rows {
		b := budgets[r.Identity.Route]
		addBudget(&b, r)
		budgets[r.Identity.Route] = b
	}
	nonces := map[string]bool{}
	for i, route := range []string{"chat", "responses", "messages"} {
		id, _ := Select(route, M3, "standard")
		var claim renewalClaim
		_, e = cohortJSON(root, base+route+".json", &claim)
		if e != nil || claim.Schema != 2 || claim.RenewalID != Recovery4RenewalID || claim.GrantSHA256 != pin || claim.Identity != id || claim.Session != sessions[route] || !regexpHex64(claim.ProcessNonce) || nonces[claim.ProcessNonce] {
			return fmt.Errorf("closed fourth consumed independent nonce differs")
		}
		nonces[claim.ProcessNonce] = true
		b := budgets[route]
		if b.Calls != 13 || b.Attempts != 12 || b.ReservedAttempts != 13 || b.RequestedOutput < 1 || b.RequestedOutput > 40000 || b.ReservedInput < 1 || b.ReservedInput > 120000 {
			return fmt.Errorf("closed fourth per-route13/12 differs")
		}
		if e = closedMiniMaxProcess(root, stop.Files, fourthCollection, stop.Processes[i+1].nativeProcess(), g, pin, claim, b, rows); e != nil {
			return e
		}
	}
	h.NativeCalls = 39
	h.NativeAttempts = 36
	h.NativeLedgerRows = 75
	h.NativeLedgerTailSHA256 = rows[74].Hash
	for i, p := range stop.Processes {
		if e = fourthProcessHistory(root, stop, p, g, pin, i); e != nil {
			return e
		}
		h.ParentClaimsSHA256[p.Name] = stop.Files[".evidence/minimax-launches/"+Recovery4RenewalID+"-"+p.Name+".json"]
	}
	return nil
}

// This shape check additionally protects the old source grant from new guard
// fields; schema5 never authorized a changed spending implementation.
func oldFourthGrantOnly(g RenewalGrant) bool {
	return g.FourthCohort == nil && len(g.SpendingGuardSourceSHA256) == 0
}

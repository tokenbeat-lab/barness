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

type secondSeal struct {
	Schema                     int    `json:"schema"`
	Kind                       string `json:"kind"`
	RoundID                    string `json:"roundId"`
	State                      string `json:"state"`
	StopIndex                  string `json:"stopIndex"`
	StopIndexSHA256            string `json:"stopIndexSHA256"`
	ActualCalls                int    `json:"actualCalls"`
	ActualHTTPAttempts         int    `json:"actualHTTPAttempts"`
	ClaimedCalls               int    `json:"claimedCalls"`
	ClaimedHTTPAttempts        int    `json:"claimedHTTPAttempts"`
	WholeReservedMicrousd      int64  `json:"wholeReservedMicrousd"`
	RefundMicrousd             int64  `json:"refundMicrousd"`
	NoAutomaticReruns          bool   `json:"noAutomaticReruns"`
	PaidContinuationAuthorized bool   `json:"paidContinuationAuthorized"`
}
type secondStop struct {
	Schema                       int               `json:"schema"`
	Kind                         string            `json:"kind"`
	SourceCommit                 string            `json:"sourceCommit"`
	RegionalFormalProcesses      int               `json:"regionalFormalProcesses"`
	RegionalFormalScenarios      int               `json:"regionalFormalScenarios"`
	RegionalFormalCalls          int               `json:"regionalFormalCalls"`
	RegionalFormalHTTP           int               `json:"regionalFormalHTTP"`
	GlobalMiniMaxFailedProcesses int               `json:"globalMiniMaxFailedProcesses"`
	GlobalMiniMaxActualCalls     int               `json:"globalMiniMaxActualCalls"`
	GlobalMiniMaxActualHTTP      int               `json:"globalMiniMaxActualHTTP"`
	CampaignActualCalls          int               `json:"campaignActualCalls"`
	CampaignActualHTTP           int               `json:"campaignActualHTTP"`
	CampaignClaimedCalls         int               `json:"campaignClaimedCalls"`
	CampaignClaimedHTTP          int               `json:"campaignClaimedHTTP"`
	WholeReservedMicrousd        int64             `json:"wholeReservedMicrousd"`
	RefundMicrousd               int64             `json:"refundMicrousd"`
	UnknownParentProcesses       int               `json:"unknownParentProcesses"`
	OriginalFailedGLMUnchanged   bool              `json:"originalFailedGLMUnchanged"`
	AutomaticReruns              bool              `json:"automaticReruns"`
	UnstartedProcesses           []string          `json:"unstartedProcesses"`
	Processes                    []json.RawMessage `json:"processes"`
	Files                        map[string]string `json:"files"`
	ReleasePass                  bool              `json:"releasePass"`
	Race                         string            `json:"race"`
	NewREADY                     bool              `json:"newREADY"`
	Proxy                        string            `json:"proxy"`
	CredentialsReadByThisAudit   bool              `json:"credentialsReadByThisAudit"`
}

var secondUnstarted = strings.Fields("openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 mm-global-formal-m3-responses mm-global-formal-m3-messages text-jev text-luna visual-jev visual-decisions")

func closedSecondCohort(f RenewalSourceFacts) (*RecoveryHistory, error) {
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
	raw, e := cohortJSON(root, secondCohortSeal, &seal)
	if e != nil || seal.Schema != 1 || seal.Kind != "CLOSED_SECOND_COHORT_HISTORY_SEAL" || seal.RoundID != FinalAcceptanceRenewalID || seal.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || seal.StopIndex != secondStopIndex || !seal.NoAutomaticReruns || seal.PaidContinuationAuthorized || seal.ActualCalls != 303 || seal.ActualHTTPAttempts != 302 || seal.ClaimedCalls != 456 || seal.ClaimedHTTPAttempts != 492 || seal.WholeReservedMicrousd != 10700000 || seal.RefundMicrousd != 0 {
		return nil, fmt.Errorf("fixed closed failed second cohort303/302,456/492,10.7USD required: %v", e)
	}
	h := &RecoveryHistory{SecondCohortSealSHA256: Hash(raw), RetainedActualCalls: 303, RetainedActualAttempts: 302, RetainedClaimedCalls: 456, RetainedClaimedAttempts: 492, RetainedReservedMicrousd: 10700000}
	var stop secondStop
	raw, e = cohortJSON(root, secondStopIndex, &stop)
	if e != nil || Hash(raw) != seal.StopIndexSHA256 || stop.Schema != 1 || stop.Kind != "ACTUAL_REMAINING_STOP_AFTER_REAL_GLOBAL_MINIMAX_CHAT_THINKING_FAILURE" || stop.RegionalFormalProcesses != 6 || stop.RegionalFormalScenarios != 81 || stop.RegionalFormalCalls != 81 || stop.RegionalFormalHTTP != 81 || stop.GlobalMiniMaxFailedProcesses != 1 || stop.GlobalMiniMaxActualCalls != 13 || stop.GlobalMiniMaxActualHTTP != 12 || stop.CampaignActualCalls != 303 || stop.CampaignActualHTTP != 302 || stop.CampaignClaimedCalls != 456 || stop.CampaignClaimedHTTP != 492 || stop.WholeReservedMicrousd != 10700000 || stop.RefundMicrousd != 0 || stop.UnknownParentProcesses != 2 || !stop.OriginalFailedGLMUnchanged || stop.AutomaticReruns || stop.ReleasePass || stop.NewREADY || stop.Race != "NOT_RUN" || stop.CredentialsReadByThisAudit || len(stop.Files) != 234 || len(stop.Processes) != 7 || !slices.Equal(stop.UnstartedProcesses, secondUnstarted) {
		return nil, fmt.Errorf("fixed remaining stop index234 and failed history required: %v", e)
	}
	h.StopIndexSHA256 = Hash(raw)
	total := int64(0)
	for name, want := range stop.Files {
		got, size, e := cohortFileHash(root, name, maxCohortBytes-total)
		if e != nil || got != want {
			return nil, fmt.Errorf("second cohort pinned file differs: %s: %v", name, e)
		}
		total += size
	}
	for _, name := range secondUnstarted {
		if _, e = root.Lstat(".evidence/minimax-launches/" + FinalAcceptanceRenewalID + "-" + name + ".json"); !os.IsNotExist(e) {
			return nil, fmt.Errorf("closed second cohort acquired an unstarted process")
		}
	}
	if e = secondNativeHistory(root, stop, f, h); e != nil {
		return nil, e
	}
	return h, nil
}

func secondNativeHistory(root *os.Root, stop secondStop, f RenewalSourceFacts, h *RecoveryHistory) error {
	const base = ".evidence/minimax-renewals/" + FinalAcceptanceRenewalID + "/"
	dir, e := root.Open(strings.TrimSuffix(base, "/"))
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(8)
	dir.Close()
	if e != nil || len(entries) != 7 {
		return fmt.Errorf("closed second native authority file set differs")
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if !slices.Equal(names, []string{"chat.json", "formal.jsonl", "grant.json", "human-approval.json", "launch.lock", "operator-authority.json", "source-review.json"}) {
		return fmt.Errorf("closed second native exact consumed-chat/inactive-routes required")
	}
	for _, name := range names {
		if stop.Files[base+name] == "" {
			return fmt.Errorf("closed second native authority not pinned")
		}
	}
	var g RenewalGrant
	raw, e := cohortJSON(root, base+"grant.json", &g)
	if e != nil || g.SixthCohort != nil || g.FourthCohort != nil || g.ThirdCohort != nil || g.Schema != 3 || g.Recovery != nil || g.RenewalID != FinalAcceptanceRenewalID || g.Decision != "APPROVED" || !fixedRenewalCaps(g) || !sameRenewalHistory(g, f) || g.SourceCommit != stop.SourceCommit || g.HumanApprovalSHA256 != stop.Files[base+"human-approval.json"] || g.SourceReviewSHA256 != stop.Files[base+"source-review.json"] {
		return fmt.Errorf("closed failed ID2 original grant/history differs: %v", e)
	}
	h.SecondGrantSHA256 = Hash(raw)
	h.SecondHumanApprovalSHA256 = g.HumanApprovalSHA256
	h.SecondSourceReviewSHA256 = g.SourceReviewSHA256
	owner, e := evidenceOwner()
	if e != nil {
		return e
	}
	if e = finalRenewalApproval(filepath.Join(owner, base), g); e != nil {
		return fmt.Errorf("closed ID2 original authority: %w", e)
	}
	var claim renewalClaim
	raw, e = cohortJSON(root, base+"chat.json", &claim)
	id, _ := Select("chat", M3, "standard")
	if e != nil || claim.Schema != 2 || claim.RenewalID != FinalAcceptanceRenewalID || claim.GrantSHA256 != h.SecondGrantSHA256 || claim.Identity != id || claim.Session == "" || filepath.Base(claim.Session) != claim.Session || !regexpHex64(claim.ProcessNonce) {
		return fmt.Errorf("closed failed ID2 consumed chat nonce differs")
	}
	h.SecondChatClaimSHA256 = Hash(raw)
	rows, e := ReadLedger(filepath.Join(owner, base, "formal.jsonl"))
	if e != nil || len(rows) != 25 {
		return fmt.Errorf("closed failed ID2 exact25 ledger required: %v", e)
	}
	budget := Budget{}
	for _, r := range rows {
		if r.RenewalID != FinalAcceptanceRenewalID || r.RenewalGrantSHA256 != h.SecondGrantSHA256 || r.Identity != id || r.Stage != "formal" || r.Session != claim.Session {
			return fmt.Errorf("closed failed ID2 ledger identity differs")
		}
		addBudget(&budget, r)
	}
	if budget.Calls != 13 || budget.ReservedAttempts != 13 || budget.Attempts != 12 {
		return fmt.Errorf("closed failed ID2 exact13/12 required")
	}
	h.SecondCalls, h.SecondAttempts, h.SecondLedgerRows = 13, 12, 25
	h.SecondLedgerSHA256 = stop.Files[base+"formal.jsonl"]
	h.SecondLedgerTailSHA256 = rows[len(rows)-1].Hash
	return secondFailedProcess(root, stop, g, *h, claim, budget)
}
func regexpHex64(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			if c < 'a' || c > 'f' {
				return false
			}
		}
	}
	return true
}

// The parent receipt stays UNKNOWN. A native FAIL is not a successful formal
// session and must never pass through closedFormalSessions'39/36 contract.
func secondFailedProcess(root *os.Root, stop secondStop, g RenewalGrant, h RecoveryHistory, claim renewalClaim, budget Budget) error {
	var failed struct {
		Name                        string          `json:"name"`
		State                       string          `json:"state"`
		ExitCode                    int             `json:"exitCode"`
		ParentBudget                json.RawMessage `json:"parentBudget"`
		NativeBundle                string          `json:"nativeBundle"`
		NativeReportSHA256          string          `json:"nativeReportSHA256"`
		NativeOutcome               string          `json:"nativeOutcome"`
		NativeComplete              bool            `json:"nativeComplete"`
		NativeAuditPassed           bool            `json:"nativeAuditPassed"`
		NativeBudget                Budget          `json:"nativeBudget"`
		NativeConsumerReceiptSHA256 *string         `json:"nativeConsumerReceiptSHA256"`
		WholeReservedMicrousd       int64           `json:"wholeReservedMicrousd"`
	}
	if json.Unmarshal(stop.Processes[6], &failed) != nil || failed.Name != "mm-global-formal-m3-chat" || failed.State != "COMPLETED" || failed.ExitCode != 1 || failed.NativeOutcome != "FAIL" || failed.NativeComplete || !failed.NativeAuditPassed || failed.NativeBudget != budget || failed.WholeReservedMicrousd != 200000 || failed.NativeConsumerReceiptSHA256 != nil || !unknownParentBudget(failed.ParentBudget) {
		return fmt.Errorf("closed failed native process/parent UNKNOWN differs")
	}
	expected := []string{"formal-zai-chat", "formal-zai-responses", "formal-zai-messages", "formal-minimax-cn-chat", "formal-minimax-cn-responses", "formal-minimax-cn-messages"}
	for i, name := range expected {
		var p struct {
			Name     string `json:"name"`
			State    string `json:"state"`
			ExitCode int    `json:"exitCode"`
		}
		if json.Unmarshal(stop.Processes[i], &p) != nil || p.Name != name || p.State != "COMPLETED" || p.ExitCode != 0 {
			return fmt.Errorf("closed regional processes differ")
		}
	}
	prefix := ".evidence/provider-sites/final-acceptance-20261020/remaining/mm-global-formal-m3-chat/"
	if failed.NativeBundle != prefix+claim.Session || stop.Files[failed.NativeBundle+"/minimax-report.json"] != failed.NativeReportSHA256 {
		return fmt.Errorf("closed failed bundle is not canonical/pinned")
	}
	var report Report
	raw, e := cohortJSON(root, failed.NativeBundle+"/minimax-report.json", &report)
	source, _ := json.Marshal(report.ExecutionSource)
	if e != nil || Hash(raw) != failed.NativeReportSHA256 || report.Schema != 1 || report.Mode != "live" || report.Kind != "minimax-formal" || report.Stage != "formal" || report.Outcome != "FAIL" || report.Complete || !report.AuditPassed || report.FormalRenewalID != FinalAcceptanceRenewalID || report.FormalRenewalGrantSHA256 != h.SecondGrantSHA256 || report.Identity != claim.Identity || report.Session != claim.Session || report.GitCommit != g.SourceCommit || report.Budget != budget || report.Cumulative != budget || Hash(source) != g.ExecutionSourceSHA256 {
		return fmt.Errorf("closed actual failed native report differs: %v", e)
	}
	name := ".evidence/minimax-launches/" + FinalAcceptanceRenewalID + "-mm-global-formal-m3-chat.json"
	var parent struct {
		State              string          `json:"state"`
		Route              string          `json:"route"`
		BaseName           string          `json:"baseName"`
		ValidationRound    string          `json:"validationRound"`
		RenewalID          string          `json:"renewalId"`
		RenewalGrantSHA256 string          `json:"renewalGrantSha256"`
		SourceCommit       string          `json:"sourceCommit"`
		ExitCode           int             `json:"exitCode"`
		CredentialsRead    bool            `json:"credentialsRead"`
		Budget             json.RawMessage `json:"budget"`
	}
	// The historical parent has a larger envelope. Check duplicate fields before
	// projecting only its execution identity and unchanged UNKNOWN accounting.
	var envelope map[string]json.RawMessage
	raw, e = cohortJSON(root, name, &envelope)
	if e != nil || stop.Files[name] != Hash(raw) || json.Unmarshal(raw, &parent) != nil || parent.State != "COMPLETED" || parent.Route != FinalAcceptanceRenewalID+"-mm-global-formal-m3-chat" || parent.BaseName != "mm-global-formal-m3-chat" || parent.ValidationRound != FinalAcceptanceRenewalID || parent.RenewalID != FinalAcceptanceRenewalID || parent.RenewalGrantSHA256 != h.SecondGrantSHA256 || parent.SourceCommit != g.SourceCommit || parent.ExitCode != 1 || !parent.CredentialsRead || !unknownParentBudget(parent.Budget) {
		return fmt.Errorf("canonical closed parent UNKNOWN authority differs")
	}
	return nil
}
func unknownParentBudget(raw []byte) bool {
	var m map[string]json.RawMessage
	if StrictJSON(raw, &m) != nil || len(m) != 4 {
		return false
	}
	return reflect.DeepEqual(m["calls"], json.RawMessage("null")) && reflect.DeepEqual(m["attempts"], json.RawMessage("null")) && string(m["state"]) == `"UNKNOWN"` && string(m["status"]) == `"UNKNOWN"`
}

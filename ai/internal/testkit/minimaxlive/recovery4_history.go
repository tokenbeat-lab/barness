package minimaxlive

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
)

type thirdStop struct {
	Schema                     int               `json:"schema"`
	Kind                       string            `json:"kind"`
	SourceCommit               string            `json:"sourceCommit"`
	CreatedAt                  string            `json:"createdAt"`
	ActualHumanApprovalSHA256  string            `json:"actualHumanApprovalSHA256"`
	Processes                  []thirdProcess    `json:"processes"`
	SuccessfulProcesses        int               `json:"successfulProcesses"`
	FailedProcesses            int               `json:"failedProcesses"`
	NewActualCalls             int               `json:"newActualCalls"`
	NewActualHTTP              int               `json:"newActualHTTP"`
	CampaignActualCalls        int               `json:"campaignActualCalls"`
	CampaignActualHTTP         int               `json:"campaignActualHTTP"`
	CampaignClaimedCalls       int               `json:"campaignClaimedCalls"`
	CampaignClaimedHTTP        int               `json:"campaignClaimedHTTP"`
	NewWholeReservedMicrousd   int64             `json:"newWholeReservedMicrousd"`
	WholeReservedMicrousd      int64             `json:"wholeReservedMicrousd"`
	RefundMicrousd             int64             `json:"refundMicrousd"`
	UnknownParentProcesses     int               `json:"unknownParentProcesses"`
	OriginalFailedGLMUnchanged bool              `json:"originalFailedGLMUnchanged"`
	OriginalFailedM3Unchanged  bool              `json:"originalFailedM3Unchanged"`
	AutomaticReruns            bool              `json:"automaticReruns"`
	UnstartedProcesses         []string          `json:"unstartedProcesses"`
	Files                      map[string]string `json:"files"`
	NewREADY                   bool              `json:"newREADY"`
	Race                       string            `json:"race"`
	ReleasePass                bool              `json:"releasePass"`
	CredentialsReadByThisAudit bool              `json:"credentialsReadByThisAudit"`
}

var thirdUnstarted = strings.Fields("formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions")
var thirdNativeNames = strings.Fields("chat.json formal.jsonl grant.json human-approval.json launch.lock messages.json operator-authority.json responses.json source-review.json")

func closedThirdCohort(f RenewalSourceFacts) (*ThirdCohortHistory, error) {
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
	raw, e := cohortJSON(root, thirdCohortSeal, &seal)
	if e != nil || seal.Schema != 1 || seal.Kind != "CLOSED_THIRD_COHORT_HISTORY_SEAL" || seal.RoundID != RecoveryRenewalID || seal.State != "CLOSED_NO_FURTHER_PAID_LAUNCH" || seal.StopIndex != thirdStopIndex || !seal.NoAutomaticReruns || seal.PaidContinuationAuthorized || seal.ActualCalls != 378 || seal.ActualHTTPAttempts != 374 || seal.ClaimedCalls != 564 || seal.ClaimedHTTPAttempts != 612 || seal.WholeReservedMicrousd != 12600000 || seal.RefundMicrousd != 0 {
		return nil, fmt.Errorf("fixed closed third cohort378/374,564/612,12.6USD required: %v", e)
	}
	h := &ThirdCohortHistory{CohortSealSHA256: Hash(raw), RetainedActualCalls: 378, RetainedActualAttempts: 374, RetainedClaimedCalls: 564, RetainedClaimedAttempts: 612, RetainedReservedMicrousd: 12600000, RetainedUnknownParentProcesses: 3, NewActualCalls: 75, NewActualAttempts: 72, NewClaimedCalls: 108, NewClaimedAttempts: 120, NewReservedMicrousd: 1900000, NativeAuthoritySHA256: map[string]string{}, ParentClaimsSHA256: map[string]string{}}
	var stop thirdStop
	raw, e = cohortJSON(root, thirdStopIndex, &stop)
	if e != nil || Hash(raw) != seal.StopIndexSHA256 || stop.Schema != 1 || stop.Kind != "ACTUAL_RECOVERY3_STOP_AFTER_ZAI_RESPONSES_NATIVE_SOURCE_FAILURE" || stop.SuccessfulProcesses != 4 || stop.FailedProcesses != 1 || stop.NewActualCalls != 75 || stop.NewActualHTTP != 72 || stop.CampaignActualCalls != 378 || stop.CampaignActualHTTP != 374 || stop.CampaignClaimedCalls != 564 || stop.CampaignClaimedHTTP != 612 || stop.NewWholeReservedMicrousd != 1900000 || stop.WholeReservedMicrousd != 12600000 || stop.RefundMicrousd != 0 || stop.UnknownParentProcesses != 3 || !stop.OriginalFailedGLMUnchanged || !stop.OriginalFailedM3Unchanged || stop.AutomaticReruns || stop.NewREADY || stop.ReleasePass || stop.Race != "NOT_RUN" || stop.CredentialsReadByThisAudit || len(stop.Files) != 1012 || len(stop.Processes) != 5 || !slices.Equal(stop.UnstartedProcesses, thirdUnstarted) {
		return nil, fmt.Errorf("fixed closed third1012/5 history differs: %v", e)
	}
	h.StopIndexSHA256 = Hash(raw)
	h.NativeSourceCommit = stop.SourceCommit
	h.ActualHumanApprovalSHA256 = stop.ActualHumanApprovalSHA256
	total := int64(0)
	for name, want := range stop.Files {
		got, n, e := cohortFileHash(root, name, maxCohortBytes-total)
		if e != nil || got != want {
			return nil, fmt.Errorf("third cohort pinned file differs: %s: %v", name, e)
		}
		total += n
	}
	for _, name := range thirdUnstarted {
		if _, e = root.Lstat(".evidence/minimax-launches/" + RecoveryRenewalID + "-" + name + ".json"); !os.IsNotExist(e) {
			return nil, fmt.Errorf("closed third cohort acquired unstarted process")
		}
	}
	if e = thirdNativeHistory(root, owner, stop, f, h); e != nil {
		return nil, e
	}
	return h, nil
}

func thirdNativeHistory(root *os.Root, owner string, stop thirdStop, f RenewalSourceFacts, h *ThirdCohortHistory) error {
	const base = ".evidence/minimax-renewals/" + RecoveryRenewalID + "/"
	dir, e := root.Open(strings.TrimSuffix(base, "/"))
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(10)
	dir.Close()
	if e != nil || len(entries) != 9 {
		return fmt.Errorf("closed third native authority exact9 required")
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if !slices.Equal(names, thirdNativeNames) {
		return fmt.Errorf("closed third native authority file set differs")
	}
	for _, name := range names {
		pin := stop.Files[base+name]
		if !regexpHex64(pin) {
			return fmt.Errorf("closed third native authority unpinned")
		}
		h.NativeAuthoritySHA256[name] = pin
	}
	var g RenewalGrant
	raw, e := cohortJSON(root, base+"grant.json", &g)
	if e != nil || g.SixthCohort != nil || g.FifthCohort != nil || g.FourthCohort != nil || g.Schema != 4 || g.ThirdCohort != nil || g.RenewalID != RecoveryRenewalID || g.Decision != "APPROVED" || !fixedRenewalCaps(g) || !sameRenewalHistory(g, f) || !reflect.DeepEqual(g.Recovery, f.Recovery) || g.SourceCommit != stop.SourceCommit || g.HumanApprovalSHA256 != stop.Files[base+"human-approval.json"] || g.SourceReviewSHA256 != stop.Files[base+"source-review.json"] {
		return fmt.Errorf("closed third original schema4 grant/history differs: %v", e)
	}
	pin := Hash(raw)
	if e = recoveryApproval(filepath.Join(owner, base), g); e != nil {
		return fmt.Errorf("closed third original approval: %w", e)
	}
	a, e := readActualRecoveryApproval(filepath.Join(owner, thirdActualApproval), stop.ActualHumanApprovalSHA256, RecoveryRenewalID, 962, 1010, 19200000, 10700000)
	if e != nil || stop.Files[thirdActualApproval] != stop.ActualHumanApprovalSHA256 || a.SourceCommit != g.SourceCommit || a.SourceManifestSHA256 != g.SourceManifestSHA256 || a.BinariesManifestSHA256 != g.BinariesManifestSHA256 || a.PlanSHA256 != g.PlanSHA256 {
		return fmt.Errorf("closed third actual human/source differs: %v", e)
	}
	var human struct {
		ActualHumanApprovalSHA256 string `json:"actualHumanApprovalSHA256"`
		RequestReference          string `json:"requestReference"`
		RequestText               string `json:"requestText"`
	}
	humanRaw, _ := ReadFile(filepath.Join(owner, base, "human-approval.json"), 1<<20)
	if decodeRecoveryApproval(humanRaw, &human) != nil || human.ActualHumanApprovalSHA256 != stop.ActualHumanApprovalSHA256 || human.RequestReference != a.RequestReference || human.RequestText != a.RequestText {
		return fmt.Errorf("closed third native human and actual reply differ")
	}
	if e = thirdCanonicalParents(root, stop); e != nil {
		return e
	}
	if e = closedHistoricalSource(root, stop.Files, thirdCollection, g); e != nil {
		return e
	}
	rows, e := ReadLedger(filepath.Join(owner, base, "formal.jsonl"))
	if e != nil || len(rows) != 75 {
		return fmt.Errorf("closed third exact75 ledger required: %v", e)
	}
	sessions, e := closedFormalSessions(rows, RecoveryRenewalID, pin)
	if e != nil {
		return e
	}
	budgets := map[string]Budget{}
	for _, r := range rows {
		b := budgets[r.Identity.Route]
		addBudget(&b, r)
		budgets[r.Identity.Route] = b
	}
	for i, route := range []string{"chat", "responses", "messages"} {
		id, _ := Select(route, M3, "standard")
		var claim renewalClaim
		_, e = cohortJSON(root, base+route+".json", &claim)
		if e != nil || claim.Schema != 2 || claim.RenewalID != RecoveryRenewalID || claim.GrantSHA256 != pin || claim.Identity != id || claim.Session != sessions[route] || !regexpHex64(claim.ProcessNonce) {
			return fmt.Errorf("closed third consumed route nonce differs")
		}
		b := budgets[route]
		if b.Calls != 13 || b.Attempts != 12 || b.ReservedAttempts != 13 || b.RequestedOutput < 1 || b.RequestedOutput > 40000 || b.ReservedInput < 1 || b.ReservedInput > 120000 {
			return fmt.Errorf("closed third per-route13/12 differs")
		}
		if e = closedMiniMaxProcess(root, stop.Files, thirdCollection, stop.Processes[i], g, pin, claim, b, rows); e != nil {
			return e
		}
	}
	h.NativeCalls = 39
	h.NativeAttempts = 36
	h.NativeLedgerRows = 75
	h.NativeLedgerTailSHA256 = rows[74].Hash
	for i, p := range stop.Processes {
		if e = thirdParent(root, stop, p, g, pin, i); e != nil {
			return e
		}
		name := ".evidence/minimax-launches/" + RecoveryRenewalID + "-" + p.Name + ".json"
		h.ParentClaimsSHA256[p.Name] = stop.Files[name]
	}
	return thirdRegionalProcesses(root, stop, g)
}

func thirdCanonicalParents(root *os.Root, stop thirdStop) error {
	dir, e := root.Open(".evidence/minimax-launches")
	if e != nil {
		return e
	}
	defer dir.Close()
	entries, e := dir.ReadDir(4097)
	if e != nil || len(entries) > 4096 {
		return fmt.Errorf("closed canonical launch directory unavailable or excessive")
	}
	want := map[string]bool{}
	for _, p := range stop.Processes {
		want[RecoveryRenewalID+"-"+p.Name+".json"] = true
	}
	found := 0
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), RecoveryRenewalID+"-") {
			if !want[entry.Name()] {
				return fmt.Errorf("closed third acquired extra canonical claim")
			}
			found++
		}
	}
	if found != 5 {
		return fmt.Errorf("closed third exact five canonical claims required")
	}
	return nil
}

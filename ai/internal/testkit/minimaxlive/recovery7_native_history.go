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

func sixthNativeHistory(root *os.Root, owner string, stop sixthStop, f RenewalSourceFacts, h *SixthCohortHistory) error {
	const base = ".evidence/minimax-renewals/" + Recovery6RenewalID + "/"
	dir, e := root.Open(strings.TrimSuffix(base, "/"))
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(8)
	dir.Close()
	if e != nil || len(entries) != 7 {
		return fmt.Errorf("closed sixth exact7 partial native authority required")
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if !slices.Equal(names, []string{"chat.json", "formal.jsonl", "grant.json", "human-approval.json", "launch.lock", "operator-authority.json", "source-review.json"}) {
		return fmt.Errorf("closed sixth native authority file set differs")
	}
	for _, name := range names {
		pin := stop.Files[base+name]
		if !regexpHex64(pin) {
			return fmt.Errorf("closed sixth native authority unpinned")
		}
		h.NativeAuthoritySHA256[name] = pin
	}
	var g RenewalGrant
	raw, e := cohortJSON(root, base+"grant.json", &g)
	if e != nil || g.Schema != 7 || g.SixthCohort != nil || g.RenewalID != Recovery6RenewalID || g.Decision != "APPROVED" || !fixedRenewalCaps(g) || !sameRenewalHistory(g, f) || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !reflect.DeepEqual(g.FourthCohort, f.FourthCohort) || !reflect.DeepEqual(g.FifthCohort, f.FifthCohort) || g.SourceCommit != stop.SourceCommit || g.HumanApprovalSHA256 != stop.Files[base+"human-approval.json"] || g.SourceReviewSHA256 != stop.Files[base+"source-review.json"] {
		return fmt.Errorf("closed sixth original schema7 source/history differs: %v", e)
	}
	pin := Hash(raw)
	if e = recovery6Approval(filepath.Join(owner, base), g); e != nil {
		return fmt.Errorf("closed sixth original approval: %w", e)
	}
	if stop.Files[sixthActualApproval] != stop.ActualHumanApprovalSHA256 {
		return fmt.Errorf("closed sixth actual human unpinned")
	}
	if e = closedHistoricalSource(root, stop.Files, sixthCollection, g); e != nil {
		return e
	}
	bundle, e := sixthInterruptedBundle(owner, stop)
	if e != nil {
		return e
	}
	rows, e := ReadLedger(filepath.Join(owner, base, "formal.jsonl"))
	if e != nil || len(rows) != 8 {
		return fmt.Errorf("closed sixth exact8 partial ledger required: %v", e)
	}
	id, _ := Select("chat", M3, "standard")
	session := strings.TrimPrefix(bundle, sixthCollection+"mm-global-formal-m3-chat/")
	var claim renewalClaim
	_, e = cohortJSON(root, base+"chat.json", &claim)
	if e != nil || claim.Schema != 2 || claim.RenewalID != Recovery6RenewalID || claim.GrantSHA256 != pin || claim.Identity != id || claim.Session != session || !regexpHex64(claim.ProcessNonce) {
		return fmt.Errorf("closed sixth consumed chat nonce differs")
	}
	for i, r := range rows {
		kind, caseID, output, input := "reserve", "full-stream", 1024, 2048
		if i%2 == 1 {
			kind, output, input = "attempt", 0, 0
		}
		if i >= 4 {
			caseID = "full-complete"
		}
		if r.RenewalID != Recovery6RenewalID || r.RenewalGrantSHA256 != pin || r.Session != session || r.Sequence != i+1 || r.Stage != "formal" || r.Identity != id || r.Case != caseID || r.Kind != kind || r.Output != output || r.Input != input {
			return fmt.Errorf("closed sixth partial ledger allocation identity differs")
		}
	}
	b := Summarize(rows, id)
	if b.Calls != 4 || b.ReservedAttempts != 4 || b.Attempts != 4 || b.RequestedOutput != 4096 || b.ReservedInput != 8192 || rows[7].Hash != stop.InterruptedM3NativeLedgerTail {
		return fmt.Errorf("closed sixth partial4/4 ledger budget/tail differs")
	}
	if e = sixthInterruptedCase(root, stop, bundle); e != nil {
		return e
	}
	for i, p := range stop.Processes {
		if e = sixthProcessHistory(root, stop, p, g, pin, i); e != nil {
			return e
		}
		h.ParentClaimsSHA256[p.Name] = stop.Files[p.Parent]
	}
	return nil
}

func sixthInterruptedBundle(owner string, stop sixthStop) (string, error) {
	prefix := filepath.Join(owner, sixthCollection, "mm-global-formal-m3-chat") + "/"
	session := strings.TrimPrefix(stop.M3Raw, prefix)
	if !strings.HasPrefix(stop.M3Raw, prefix) || session == "" || strings.Contains(session, "/") || session == "." || session == ".." || stop.FailedFinalCase != stop.M3Raw+"/MM07-full-stream-rerun-1" {
		return "", fmt.Errorf("closed sixth partial bundle/case identity differs")
	}
	return sixthCollection + "mm-global-formal-m3-chat/" + session, nil
}

func sixthInterruptedCase(root *os.Root, stop sixthStop, bundle string) error {
	if _, e := root.Lstat(bundle + "/minimax-report.json"); !os.IsNotExist(e) {
		return fmt.Errorf("closed sixth interrupted final report must remain absent")
	}
	path := bundle + "/MM07-full-stream-rerun-1/"
	var result struct {
		ErrorCode  string `json:"errorCode"`
		ErrorPhase string `json:"errorPhase"`
	}
	raw, e := cohortJSONProjection(root, path+"result.json", &result, "errorCode", "errorPhase")
	if e != nil || stop.Files[path+"result.json"] != Hash(raw) || result.ErrorCode != "deadline_exceeded" || result.ErrorPhase != "request" {
		return fmt.Errorf("closed sixth original final case deadline/request differs")
	}
	var assertions struct {
		Status string `json:"status"`
		CaseID string `json:"case_id"`
		Test   string `json:"test"`
	}
	raw, e = cohortJSONProjection(root, path+"assertions.json", &assertions, "status", "case_id", "test")
	if e != nil || stop.Files[path+"assertions.json"] != Hash(raw) || assertions.Status != "FAIL" || assertions.CaseID != "MM07-full-stream-rerun-1" || assertions.Test != "TestMiniMaxFormal/full-stream-rerun-1" {
		return fmt.Errorf("closed sixth original failed assertion differs")
	}
	return nil
}

func sixthProcessHistory(root *os.Root, stop sixthStop, p sixthProcess, g RenewalGrant, pin string, i int) error {
	name := ".evidence/minimax-launches/" + Recovery6RenewalID + "-" + sixthNames[i] + ".json"
	exit, reserved := 0, int64(650000)
	if i == 1 {
		exit, reserved = -15, 200000
	}
	if p.Name != sixthNames[i] || p.Parent != name || p.ParentSHA256 != stop.Files[name] || p.ExitCode != exit || p.ReservedMicrousd != reserved {
		return fmt.Errorf("closed sixth parent index identity differs")
	}
	var parent struct {
		State                     string          `json:"state"`
		Route                     string          `json:"route"`
		BaseName                  string          `json:"baseName"`
		ValidationRound           string          `json:"validationRound"`
		NativeRenewalID           string          `json:"nativeRenewalId"`
		RenewalID                 string          `json:"renewalId"`
		RenewalGrantSHA256        string          `json:"renewalGrantSha256"`
		SourceCommit              string          `json:"sourceCommit"`
		ExitCode                  int             `json:"exitCode"`
		CredentialsRead           bool            `json:"credentialsRead"`
		Budget                    json.RawMessage `json:"budget"`
		ActualHumanApprovalSHA256 string          `json:"actualHumanApprovalSHA256"`
		ReservedMicrousd          int64           `json:"reservedMicrousd"`
		NetworkRoute              string          `json:"networkRoute"`
	}
	raw, e := cohortJSONProjection(root, name, &parent, "exitCode", "credentialsRead")
	if e != nil || Hash(raw) != p.ParentSHA256 || parent.State != "COMPLETED" || parent.Route != Recovery6RenewalID+"-"+p.Name || parent.BaseName != p.Name || parent.ValidationRound != Recovery6RenewalID || parent.NativeRenewalID != Recovery6RenewalID || parent.SourceCommit != g.SourceCommit || parent.ExitCode != exit || !parent.CredentialsRead || parent.ActualHumanApprovalSHA256 != stop.ActualHumanApprovalSHA256 || parent.ReservedMicrousd != reserved || parent.NetworkRoute != "http://127.0.0.1:7890" {
		return fmt.Errorf("closed sixth canonical parent authority differs: %s: %v", p.Name, e)
	}
	if i == 0 {
		if !reportedParentBudget(p.ParentBudget, 18, 18) || !reportedParentBudget(parent.Budget, 18, 18) || p.NativeConsumerReceiptSHA256 == nil || *p.NativeConsumerReceiptSHA256 != stop.Files[sixthCollection+p.Name+"-consume-command.json"] {
			return fmt.Errorf("closed sixth first18 consumer/accounting differs")
		}
	} else {
		if !unknownParentBudget(p.ParentBudget) || !unknownParentBudget(parent.Budget) || p.NativeConsumerReceiptSHA256 != nil || parent.RenewalID != Recovery6RenewalID || parent.RenewalGrantSHA256 != pin {
			return fmt.Errorf("closed sixth interrupted UNKNOWN/native grant differs")
		}
		var command struct {
			State           string `json:"state"`
			Name            string `json:"name"`
			ExitCode        int    `json:"exitCode"`
			CredentialsRead bool   `json:"credentialsRead"`
			ValidationRound string `json:"validationRound"`
			NativeRenewalID string `json:"nativeRenewalId"`
		}
		raw, e = cohortJSONProjection(root, sixthCollection+p.Name+"-command.json", &command, "state", "exitCode", "credentialsRead")
		if e != nil || stop.Files[sixthCollection+p.Name+"-command.json"] != Hash(raw) || command.State != "COMPLETED" || command.Name != p.Name || command.ExitCode != -15 || !command.CredentialsRead || command.ValidationRound != Recovery6RenewalID || command.NativeRenewalID != Recovery6RenewalID {
			return fmt.Errorf("closed sixth SIGTERM numeric-15 command differs")
		}
	}
	var usd struct {
		Schema           int    `json:"schema"`
		Round            string `json:"round"`
		MaxMicrousd      int64  `json:"maxMicrousd"`
		ReservedMicrousd int64  `json:"reservedMicrousd"`
		Attempts         int    `json:"attempts"`
		Basis            string `json:"basis"`
	}
	raw, e = cohortJSONProjection(root, sixthCollection+p.Name+"-usd.json", &usd, "attempts", "reservedMicrousd")
	if e != nil || stop.Files[sixthCollection+p.Name+"-usd.json"] != Hash(raw) || usd.Schema != 1 || usd.Round != FinalAcceptanceRenewalID || usd.MaxMicrousd != reserved || usd.Basis == "" || usd.ReservedMicrousd <= 0 || usd.ReservedMicrousd > reserved || i == 0 && usd.Attempts != 18 || i == 1 && (usd.ReservedMicrousd != 12884 || usd.Attempts != 4) {
		return fmt.Errorf("closed sixth USD recorded reservations differ")
	}
	return nil
}

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

func fifthNativeHistory(root *os.Root, owner string, stop fifthStop, f RenewalSourceFacts, h *FifthCohortHistory) error {
	const base = ".evidence/minimax-renewals/" + Recovery5RenewalID + "/"
	dir, e := root.Open(strings.TrimSuffix(base, "/"))
	if e != nil {
		return e
	}
	entries, e := dir.ReadDir(10)
	dir.Close()
	if e != nil || len(entries) != 9 {
		return fmt.Errorf("closed fifth exact9 native authority required")
	}
	names := []string{}
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	if !slices.Equal(names, thirdNativeNames) {
		return fmt.Errorf("closed fifth native authority file set differs")
	}
	for _, name := range names {
		pin := stop.Files[base+name]
		if !regexpHex64(pin) {
			return fmt.Errorf("closed fifth native authority unpinned")
		}
		h.NativeAuthoritySHA256[name] = pin
	}
	var g RenewalGrant
	raw, e := cohortJSON(root, base+"grant.json", &g)
	if e != nil || g.Schema != 6 || g.SixthCohort != nil || g.FifthCohort != nil || g.RenewalID != Recovery5RenewalID || g.Decision != "APPROVED" || !fixedRenewalCaps(g) || !sameRenewalHistory(g, f) || !reflect.DeepEqual(g.Recovery, f.Recovery) || !reflect.DeepEqual(g.ThirdCohort, f.ThirdCohort) || !reflect.DeepEqual(g.FourthCohort, f.FourthCohort) || g.SourceCommit != stop.SourceCommit || g.HumanApprovalSHA256 != stop.Files[base+"human-approval.json"] || g.SourceReviewSHA256 != stop.Files[base+"source-review.json"] {
		return fmt.Errorf("closed fifth original schema6 grant/history differs: %v", e)
	}
	pin := Hash(raw)
	if e = recovery5Approval(filepath.Join(owner, base), g); e != nil {
		return fmt.Errorf("closed fifth original approval: %w", e)
	}
	if stop.Files[".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json"] != stop.ActualHumanApprovalSHA256 {
		return fmt.Errorf("closed fifth actual human unpinned")
	}
	if e = closedHistoricalSource(root, stop.Files, fifthCollection, g); e != nil {
		return e
	}
	rows, e := ReadLedger(filepath.Join(owner, base, "formal.jsonl"))
	if e != nil || len(rows) != 75 {
		return fmt.Errorf("closed fifth exact75 ledger required: %v", e)
	}
	sessions, e := closedFormalSessions(rows, Recovery5RenewalID, pin)
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
		if e != nil || claim.Schema != 2 || claim.RenewalID != Recovery5RenewalID || claim.GrantSHA256 != pin || claim.Identity != id || claim.Session != sessions[route] || !regexpHex64(claim.ProcessNonce) || nonces[claim.ProcessNonce] {
			return fmt.Errorf("closed fifth consumed independent nonce differs")
		}
		nonces[claim.ProcessNonce] = true
		b := budgets[route]
		if b.Calls != 13 || b.Attempts != 12 || b.ReservedAttempts != 13 || b.RequestedOutput < 1 || b.RequestedOutput > 40000 || b.ReservedInput < 1 || b.ReservedInput > 120000 {
			return fmt.Errorf("closed fifth per-route13/12 differs")
		}
		p := stop.Processes[i+1]
		if e = closedMiniMaxProcess(root, stop.Files, fifthCollection, fifthNativeProcess(stop, p), g, pin, claim, b, rows); e != nil {
			return e
		}
	}
	h.NativeCalls, h.NativeAttempts, h.NativeLedgerRows, h.NativeLedgerTailSHA256 = 39, 36, 75, rows[74].Hash
	for i, p := range stop.Processes {
		if e = fifthProcessHistory(root, stop, p, g, pin, i); e != nil {
			return e
		}
		h.ParentClaimsSHA256[p.Name] = stop.Files[p.CanonicalClaim]
	}
	return nil
}

func fifthNativeProcess(stop fifthStop, p fifthProcess) thirdProcess {
	network := "http://127.0.0.1:7890"
	if strings.HasPrefix(p.Name, "formal-minimax-cn-") {
		network = "direct"
	}
	file := "/regional-report.json"
	if strings.HasPrefix(p.Name, "mm-") {
		file = "/minimax-report.json"
	}
	return thirdProcess{Name: p.Name, ExitCode: p.ExitCode, ParentBudget: p.ParentBudget, ReservedMicrousd: p.ReservedMicrousd, NativeBundle: p.Raw, NativeReportSHA256: stop.Files[p.Raw+file], ActualCalls: p.NativeRecordedCalls, ActualHTTP: p.NativeRecorderHTTP, NetworkRoute: network, NativeConsumerReceiptSHA256: stop.Files[fifthCollection+p.Name+"-consume-command.json"]}
}

func fifthProcessHistory(root *os.Root, stop fifthStop, p fifthProcess, g RenewalGrant, pin string, i int) error {
	calls := []int{18, 13, 13, 13, 18, 18, 9, 9}[i]
	attempts := calls
	if i >= 1 && i <= 3 {
		attempts = 12
	}
	reserved := []int64{650000, 200000, 200000, 200000, 650000, 650000, 350000, 350000}[i]
	exit := 0
	if i == 7 {
		exit = 1
	}
	name := ".evidence/minimax-launches/" + Recovery5RenewalID + "-" + fifthNames[i] + ".json"
	if p.Name != fifthNames[i] || p.CanonicalClaim != name || p.State != "COMPLETED" || p.ExitCode != exit || p.NativeRecordedCalls != calls || p.NativeRecorderHTTP != attempts || p.ReservedMicrousd != reserved {
		return fmt.Errorf("closed fifth process identity/counters differ")
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
	native := fifthNativeProcess(stop, p)
	if e != nil || stop.Files[name] != Hash(raw) || parent.State != "COMPLETED" || parent.Route != Recovery5RenewalID+"-"+p.Name || parent.BaseName != p.Name || parent.ValidationRound != Recovery5RenewalID || parent.NativeRenewalID != Recovery5RenewalID || parent.SourceCommit != g.SourceCommit || parent.ExitCode != exit || !parent.CredentialsRead || parent.ActualHumanApprovalSHA256 != stop.ActualHumanApprovalSHA256 || parent.ReservedMicrousd != reserved || parent.NetworkRoute != native.NetworkRoute {
		return fmt.Errorf("closed fifth canonical parent authority differs: %s: %v", p.Name, e)
	}
	if i >= 1 && i <= 3 && (parent.RenewalID != Recovery5RenewalID || parent.RenewalGrantSHA256 != pin) {
		return fmt.Errorf("closed fifth parent native grant differs")
	}
	if i == 7 {
		if !unknownParentBudget(parent.Budget) || !unknownParentBudget(p.ParentBudget) || !slices.Equal(p.FailedScenarios, []fifthFailure{{"MiniMax-M3/native-history", "FAIL", "assertion"}}) || native.NativeConsumerReceiptSHA256 != "" {
			return fmt.Errorf("closed fifth failed UNKNOWN/native-history changed")
		}
	} else if !reportedParentBudget(parent.Budget, calls, attempts) || !reportedParentBudget(p.ParentBudget, calls, attempts) || len(p.FailedScenarios) != 0 {
		return fmt.Errorf("closed fifth successful accounting differs")
	}
	var usd struct {
		Schema           int    `json:"schema"`
		Round            string `json:"round"`
		MaxMicrousd      int64  `json:"maxMicrousd"`
		ReservedMicrousd int64  `json:"reservedMicrousd"`
		Attempts         int    `json:"attempts"`
		Basis            string `json:"basis"`
	}
	name = fifthCollection + p.Name + "-usd.json"
	raw, e = cohortJSONProjection(root, name, &usd, "attempts", "reservedMicrousd")
	if e != nil || stop.Files[name] != Hash(raw) || usd.Schema != 1 || usd.Round != FinalAcceptanceRenewalID || usd.MaxMicrousd != reserved || usd.Attempts != attempts || usd.ReservedMicrousd <= 0 || usd.ReservedMicrousd > reserved || usd.Basis == "" || i == 7 && usd.ReservedMicrousd != 54238 {
		return fmt.Errorf("closed fifth USD reservation history differs")
	}
	if i >= 1 && i <= 3 {
		return nil
	}
	return fifthRegionalHistory(root, stop, p, g, i)
}

func fifthRegionalHistory(root *os.Root, stop fifthStop, p fifthProcess, g RenewalGrant, i int) error {
	prefix := fifthCollection + p.Name + "/"
	if !strings.HasPrefix(p.Raw, prefix) || strings.TrimPrefix(p.Raw, prefix) == "" || strings.Contains(strings.TrimPrefix(p.Raw, prefix), "/") {
		return fmt.Errorf("closed fifth native bundle identity differs")
	}
	var r struct {
		Schema    int    `json:"schema"`
		Combo     string `json:"combo"`
		Provider  string `json:"provider"`
		GitCommit string `json:"gitCommit"`
		Budget    struct {
			Calls    int `json:"callsUsed"`
			Attempts int `json:"httpAttempts"`
		} `json:"budget"`
		Scenarios []struct {
			ID       string `json:"id"`
			Outcome  string `json:"outcome"`
			Calls    int    `json:"calls"`
			Attempts int    `json:"attempts"`
		} `json:"scenarios"`
		Regional struct {
			Mode        string `json:"mode"`
			Stage       string `json:"stage"`
			Complete    bool   `json:"complete"`
			AuditPassed bool   `json:"auditPassed"`
		} `json:"regional"`
	}
	name := p.Raw + "/regional-report.json"
	raw, e := cohortJSONProjection(root, name, &r)
	combo := strings.TrimPrefix(p.Name, "formal-")
	provider := "zai"
	if i >= 6 {
		provider = "minimax-cn"
	}
	if e != nil || stop.Files[name] != Hash(raw) || r.Schema != 3 || r.Combo != combo || r.Provider != provider || r.GitCommit != g.SourceCommit || r.Budget.Calls != p.NativeRecordedCalls || r.Budget.Attempts != p.NativeRecorderHTTP || len(r.Scenarios) != p.NativeRecordedCalls || r.Regional.Mode != "live" || r.Regional.Stage != "formal" || r.Regional.Complete == (i == 7) || !r.Regional.AuditPassed {
		return fmt.Errorf("closed fifth original regional report differs: %v", e)
	}
	failures := 0
	seen := map[string]bool{}
	for _, s := range r.Scenarios {
		if s.ID == "" || seen[s.ID] || s.Calls != 1 || s.Attempts != 1 {
			return fmt.Errorf("closed fifth regional scenario identity/counters differ")
		}
		seen[s.ID] = true
		if s.Outcome == "FAIL" {
			failures++
			if i != 7 || s.ID != "MiniMax-M3/native-history" {
				return fmt.Errorf("closed fifth failure identity differs")
			}
		} else if s.Outcome != "PASS" {
			return fmt.Errorf("closed fifth regional outcome differs")
		}
	}
	if failures != map[bool]int{true: 1, false: 0}[i == 7] {
		return fmt.Errorf("closed fifth failure count differs")
	}
	if i != 7 {
		return closedConsumer(root, stop.Files, fifthCollection, fifthNativeProcess(stop, p))
	}
	return nil
}

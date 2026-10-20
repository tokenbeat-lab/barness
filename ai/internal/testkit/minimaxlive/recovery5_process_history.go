package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

func fourthProcessHistory(root *os.Root, stop fourthStop, p fourthProcess, g RenewalGrant, pin string, i int) error {
	counts := []int{18, 13, 13, 13, 18, 18, 9, 9, 9, 14, 14, 20, 8, 10, 13, 3, 2, 9}
	reserved := []int64{650000, 200000, 200000, 200000, 650000, 650000, 350000, 350000, 350000, 1500000, 1500000, 1800000, 500000, 500000, 500000, 3500000, 2600000, 200000}
	attempts := counts[i]
	if i >= 1 && i <= 3 {
		attempts = 12
	}
	network := "http://127.0.0.1:7890"
	if i >= 6 && i <= 8 || i == 13 || i == 14 {
		network = "direct"
	}
	exit := 0
	if i == 17 {
		exit = 1
	}
	if p.Name != fourthNames[i] || p.ExitCode != exit || p.ActualCalls != counts[i] || p.ActualHTTP != attempts || p.ReservedMicrousd != reserved[i] || p.NetworkRoute != network {
		return fmt.Errorf("closed fourth process identity/counters differ")
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
	name := ".evidence/minimax-launches/" + Recovery4RenewalID + "-" + p.Name + ".json"
	raw, e := cohortJSONProjection(root, name, &parent)
	if e != nil || stop.Files[name] != Hash(raw) || parent.State != "COMPLETED" || parent.Route != Recovery4RenewalID+"-"+p.Name || parent.BaseName != p.Name || parent.ValidationRound != Recovery4RenewalID || parent.NativeRenewalID != Recovery4RenewalID || parent.SourceCommit != g.SourceCommit || parent.ExitCode != exit || !parent.CredentialsRead || parent.ActualHumanApprovalSHA256 != stop.ActualHumanApprovalSHA256 || parent.ReservedMicrousd != reserved[i] || parent.NetworkRoute != network {
		return fmt.Errorf("closed fourth canonical parent authority differs: %s: %v", p.Name, e)
	}
	if i >= 1 && i <= 3 && (parent.RenewalID != Recovery4RenewalID || parent.RenewalGrantSHA256 != pin) {
		return fmt.Errorf("closed fourth parent native grant differs")
	}
	if i == 17 {
		if !unknownParentBudget(parent.Budget) || !unknownParentBudget(p.ParentBudget) || p.NativeConsumerReceiptSHA256 != "" || p.USDReservedMicrousd != 0 || p.USDLedgerAttempts != 0 {
			return fmt.Errorf("closed fourth failed UNKNOWN/guard0 changed")
		}
	} else if !reportedParentBudget(parent.Budget, counts[i], attempts) || !reportedParentBudget(p.ParentBudget, counts[i], attempts) || p.USDLedgerAttempts != attempts || p.USDReservedMicrousd <= 0 || p.USDReservedMicrousd > reserved[i] {
		return fmt.Errorf("closed fourth successful accounting differs")
	}
	var usd struct {
		Last             json.RawMessage `json:"last,omitempty"`
		Schema           int             `json:"schema"`
		Round            string          `json:"round"`
		MaxMicrousd      int64           `json:"maxMicrousd"`
		ReservedMicrousd int64           `json:"reservedMicrousd"`
		Attempts         int             `json:"attempts"`
		Basis            string          `json:"basis"`
	}
	name = fourthCollection + p.Name + "-usd.json"
	raw, e = cohortJSON(root, name, &usd)
	if e != nil || stop.Files[name] != Hash(raw) || usd.Schema != 1 || usd.Round != FinalAcceptanceRenewalID || usd.MaxMicrousd != reserved[i] || usd.ReservedMicrousd != p.USDReservedMicrousd || usd.Attempts != p.USDLedgerAttempts || usd.Basis == "" {
		return fmt.Errorf("closed fourth USD reservation history differs")
	}
	if i >= 1 && i <= 3 {
		return nil
	}
	return fourthReport(root, stop, p, g, i)
}

func fourthReport(root *os.Root, stop fourthStop, p fourthProcess, g RenewalGrant, i int) error {
	prefix := fourthCollection + p.Name + "/"
	file := "regional-report.json"
	if i >= 9 {
		prefix = fourthCollection + "live/" + p.Name + "/"
		file = "live-report.json"
	}
	if !strings.HasPrefix(p.NativeBundle, prefix) || len(p.NativeBundle) <= len(prefix) || strings.Contains(strings.TrimPrefix(p.NativeBundle, prefix), "/") {
		return fmt.Errorf("closed fourth native bundle identity differs")
	}
	var r struct {
		Schema    int    `json:"schema"`
		Combo     string `json:"combo"`
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
	name := p.NativeBundle + "/" + file
	raw, e := cohortJSONProjection(root, name, &r)
	if e != nil || stop.Files[name] != Hash(raw) || Hash(raw) != p.NativeReportSHA256 || r.Schema != 3 || r.Combo != strings.TrimPrefix(p.Name, "formal-") || r.GitCommit != g.SourceCommit || r.Budget.Calls != p.ActualCalls || r.Budget.Attempts != p.ActualHTTP {
		return fmt.Errorf("closed fourth original native report differs: %v", e)
	}
	if i < 9 && (r.Regional.Mode != "live" || r.Regional.Stage != "formal" || !r.Regional.Complete || !r.Regional.AuditPassed) {
		return fmt.Errorf("closed fourth regional qualification facts differ")
	}
	scenarios := []int{18, 13, 13, 13, 18, 18, 9, 9, 9, 8, 10, 11, 6, 8, 11, 3, 2, 9}
	if len(r.Scenarios) != scenarios[i] {
		return fmt.Errorf("closed fourth scenario count differs")
	}
	outcomes := map[string]int{}
	seen := map[string]bool{}
	for _, s := range r.Scenarios {
		if s.ID == "" || seen[s.ID] {
			return fmt.Errorf("closed fourth duplicate scenario")
		}
		seen[s.ID] = true
		outcomes[s.Outcome]++
		if i == 17 && (s.Outcome != "FAIL" || s.Calls != 1 || s.Attempts != 1) {
			return fmt.Errorf("closed fourth failed recorder9/9 changed")
		}
		if i != 17 && s.Outcome != "PASS" && !(i == 10 && s.Outcome == "UNSUPPORTED") {
			return fmt.Errorf("closed fourth successful outcome changed")
		}
	}
	want := map[string]int{"PASS": scenarios[i]}
	if i == 10 {
		want = map[string]int{"PASS": 9, "UNSUPPORTED": 1}
	}
	if i == 17 {
		want = map[string]int{"FAIL": 9}
	}
	if !reflect.DeepEqual(outcomes, want) || !reflect.DeepEqual(p.ScenarioOutcomes, want) {
		return fmt.Errorf("closed fourth scenario outcomes differ")
	}
	if i != 17 {
		return closedConsumer(root, stop.Files, fourthCollection, p.nativeProcess())
	}
	return nil
}

// Convert the closed fourth process projection at the shared native-report seam.
func (p fourthProcess) nativeProcess() thirdProcess {
	return thirdProcess{Name: p.Name, ExitCode: p.ExitCode, ParentBudget: p.ParentBudget, ReservedMicrousd: p.ReservedMicrousd, NativeBundle: p.NativeBundle, NativeReportSHA256: p.NativeReportSHA256, ActualCalls: p.ActualCalls, ActualHTTP: p.ActualHTTP, NetworkRoute: p.NetworkRoute, NativeOutcome: p.NativeOutcome, NativeComplete: p.NativeComplete, NativeAuditPassed: p.NativeAuditPassed, NativeConsumerReceiptSHA256: p.NativeConsumerReceiptSHA256, ScenarioOutcomes: p.ScenarioOutcomes, FailedScenarios: p.FailedScenarios}
}

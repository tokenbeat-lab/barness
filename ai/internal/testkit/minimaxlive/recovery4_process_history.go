package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
)

type thirdProcess struct {
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
}

func reportedParentBudget(raw []byte, calls, attempts int) bool {
	var b struct {
		State    string `json:"state"`
		Status   string `json:"status"`
		Calls    int    `json:"calls"`
		Attempts int    `json:"attempts"`
	}
	return StrictJSON(raw, &b) == nil && b.State == "REPORTED" && b.Status == "PASS" && b.Calls == calls && b.Attempts == attempts
}
func thirdConsumer(root *os.Root, stop thirdStop, p thirdProcess) error {
	name := thirdCollection + p.Name + "-consume-command.json"
	var c struct {
		State           string `json:"state"`
		ExitCode        int    `json:"exitCode"`
		CredentialsRead bool   `json:"credentialsRead"`
	}
	var envelope map[string]json.RawMessage
	raw, e := cohortJSON(root, name, &envelope)
	if e != nil || Hash(raw) != p.NativeConsumerReceiptSHA256 || stop.Files[name] != Hash(raw) || decodeRecoveryApproval(raw, &c) != nil || c.State != "COMPLETED" || c.ExitCode != 0 || c.CredentialsRead {
		return fmt.Errorf("closed third native consumer receipt differs")
	}
	return nil
}
func thirdMiniMaxProcess(root *os.Root, stop thirdStop, p thirdProcess, g RenewalGrant, pin string, claim renewalClaim, b Budget, rows []Receipt) error {
	if p.Name != "mm-global-formal-m3-"+claim.Identity.Route || p.ExitCode != 0 || p.ReservedMicrousd != 200000 || p.NativeOutcome != "PASS" || !p.NativeComplete || !p.NativeAuditPassed || p.ActualCalls != 13 || p.ActualHTTP != 12 || !reportedParentBudget(p.ParentBudget, 13, 12) || p.NetworkRoute != "http://127.0.0.1:7890" || p.NativeBundle != thirdCollection+p.Name+"/"+claim.Session {
		return fmt.Errorf("closed third native PASS13/12 differs")
	}
	var r Report
	raw, e := cohortJSON(root, p.NativeBundle+"/minimax-report.json", &r)
	source, _ := json.Marshal(r.ExecutionSource)
	if e != nil || Hash(raw) != p.NativeReportSHA256 || stop.Files[p.NativeBundle+"/minimax-report.json"] != Hash(raw) || r.Schema != 1 || r.Mode != "live" || r.Stage != "formal" || r.Kind != "minimax-formal" || r.Outcome != "PASS" || !r.Complete || !r.AuditPassed || r.Identity != claim.Identity || r.Session != claim.Session || r.GitCommit != g.SourceCommit || r.FormalRenewalID != RecoveryRenewalID || r.FormalRenewalGrantSHA256 != pin || r.Budget != b || r.Cumulative != b || Hash(source) != g.ExecutionSourceSHA256 {
		return fmt.Errorf("closed third original native report/source differs: %v", e)
	}
	plan := Plan(claim.Identity)
	if len(r.Calls) != len(plan) || !reflect.DeepEqual(r.Expected, plan) {
		return fmt.Errorf("closed third native complete13 calls required")
	}
	receipts := map[string]Receipt{}
	for _, row := range rows {
		receipts[row.Hash] = row
	}
	for i, c := range r.Calls {
		if c.ID != plan[i].ID || c.Outcome != "PASS" || c.Rerun != 0 || c.Reservation.Kind != "reserve" || c.Reservation.Identity != claim.Identity || c.Reservation.Case != c.ID || receipts[c.Reservation.Hash] != c.Reservation {
			return fmt.Errorf("closed third native own PASS receipt differs")
		}
		attempts := 1
		if c.ID == "preflight-error" {
			attempts = 0
		}
		if len(c.Attempts) != attempts {
			return fmt.Errorf("closed third native actual HTTP differs")
		}
		for _, a := range c.Attempts {
			if a.Kind != "attempt" || a.Identity != claim.Identity || a.Case != c.ID || receipts[a.Hash] != a {
				return fmt.Errorf("closed third native HTTP ledger receipt differs")
			}
		}
	}
	// The complete launcher and native source selectors intentionally differ:
	// each full map must match the same old commit, never assume one is a subset.
	if e = committedArchive(g.SourceCommit, r.ExecutionSource); e != nil {
		return e
	}
	return thirdConsumer(root, stop, p)
}
func thirdParent(root *os.Root, stop thirdStop, p thirdProcess, g RenewalGrant, pin string, i int) error {
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
	name := ".evidence/minimax-launches/" + RecoveryRenewalID + "-" + p.Name + ".json"
	var envelope map[string]json.RawMessage
	raw, e := cohortJSON(root, name, &envelope)
	if e != nil || stop.Files[name] != Hash(raw) || decodeRecoveryApproval(raw, &parent) != nil || parent.State != "COMPLETED" || parent.Route != RecoveryRenewalID+"-"+p.Name || parent.BaseName != p.Name || parent.ValidationRound != RecoveryRenewalID || parent.NativeRenewalID != RecoveryRenewalID || parent.SourceCommit != g.SourceCommit || parent.ExitCode != p.ExitCode || !parent.CredentialsRead || parent.ActualHumanApprovalSHA256 != stop.ActualHumanApprovalSHA256 || parent.ReservedMicrousd != p.ReservedMicrousd || parent.NetworkRoute != "http://127.0.0.1:7890" {
		return fmt.Errorf("closed third canonical parent authority differs: %s", p.Name)
	}
	if i < 3 && (parent.RenewalID != RecoveryRenewalID || parent.RenewalGrantSHA256 != pin) {
		return fmt.Errorf("closed third native parent grant differs")
	}
	if i == 4 {
		if !unknownParentBudget(parent.Budget) {
			return fmt.Errorf("closed third failed parent UNKNOWN changed")
		}
	} else if !reportedParentBudget(parent.Budget, p.ActualCalls, p.ActualHTTP) {
		return fmt.Errorf("closed third successful parent accounting changed")
	}
	return nil
}
func thirdRegionalProcesses(root *os.Root, stop thirdStop, g RenewalGrant) error {
	for i, route := range []string{"chat", "responses"} {
		p := stop.Processes[i+3]
		failed := i == 1
		if p.Name != "formal-zai-"+route || p.ExitCode != i || p.ActualCalls != 18 || p.ActualHTTP != 18 || p.ReservedMicrousd != 650000 || p.NetworkRoute != "http://127.0.0.1:7890" || p.ScenarioOutcomes["PASS"] != 18-i || p.ScenarioOutcomes["FAIL"] != i || p.ScenarioOutcomes["NOT_RUN"] != 0 {
			return fmt.Errorf("closed third regional18 actual differs")
		}
		if failed {
			if !unknownParentBudget(p.ParentBudget) || !slices.Equal(p.FailedScenarios, []string{"glm-5.3/native-history"}) || p.NativeConsumerReceiptSHA256 != "" {
				return fmt.Errorf("closed third failed regional/UNKNOWN differs")
			}
		} else if !reportedParentBudget(p.ParentBudget, 18, 18) || len(p.FailedScenarios) != 0 {
			return fmt.Errorf("closed third regional PASS differs")
		}
		if len(p.NativeBundle) <= len(thirdCollection+p.Name+"/") || p.NativeBundle[:len(thirdCollection+p.Name+"/")] != thirdCollection+p.Name+"/" {
			return fmt.Errorf("closed third regional bundle path differs")
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
		var envelope map[string]json.RawMessage
		raw, e := cohortJSON(root, p.NativeBundle+"/regional-report.json", &envelope)
		if e != nil || stop.Files[p.NativeBundle+"/regional-report.json"] != Hash(raw) || Hash(raw) != p.NativeReportSHA256 || decodeRecoveryApproval(raw, &r) != nil || r.Schema != 3 || r.Combo != "zai-"+route || r.Provider != "zai" || r.GitCommit != g.SourceCommit || r.Budget.Calls != 18 || r.Budget.Attempts != 18 || len(r.Scenarios) != 18 || r.Regional.Mode != "live" || r.Regional.Stage != "formal" || r.Regional.Complete == failed || !r.Regional.AuditPassed {
			return fmt.Errorf("closed third regional original report differs: %v", e)
		}
		failures := 0
		for _, s := range r.Scenarios {
			if s.Calls != 1 || s.Attempts != 1 {
				return fmt.Errorf("closed third regional scenario counters differ")
			}
			if s.Outcome == "FAIL" {
				failures++
				if !failed || s.ID != "glm-5.3/native-history" {
					return fmt.Errorf("closed third regional failure identity differs")
				}
			} else if s.Outcome != "PASS" {
				return fmt.Errorf("closed third regional outcome differs")
			}
		}
		if failures != i {
			return fmt.Errorf("closed third regional failure count differs")
		}
		if !failed {
			if e = thirdConsumer(root, stop, p); e != nil {
				return e
			}
		}
	}
	return nil
}

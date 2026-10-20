package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func (f *finalRenewalFixture) seedSecondCohort(t *testing.T) {
	t.Helper()
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", finalRenewalID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	l := minimaxlive.Ledger{Path: path, Stage: "formal", Session: "failed-chat", Identity: id}
	for _, p := range minimaxlive.Plan(id) {
		if _, e = l.Reserve(p.ID, p.Output, 2048); e != nil {
			t.Fatal(e)
		}
		if p.ID != "preflight-error" {
			if _, e = l.Attempt(p.ID); e != nil {
				t.Fatal(e)
			}
		}
	}
	os.Unsetenv("BARNESS_MINIMAX_FORMAL_RENEWAL")
	files := map[string]string{}
	write := func(name string, v any) []byte {
		b, _ := json.Marshal(v)
		p := filepath.Join(f.root, name)
		os.MkdirAll(filepath.Dir(p), 0700)
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
		files[name] = minimaxlive.Hash(b)
		return b
	}
	base := ".evidence/minimax-renewals/" + finalRenewalID + "/"
	write(base+"operator-authority.json", map[string]any{"validationRound": finalRenewalID})
	if e = os.WriteFile(filepath.Join(f.root, base, "launch.lock"), nil, 0600); e != nil {
		t.Fatal(e)
	}
	for _, name := range []string{"grant.json", "human-approval.json", "source-review.json", "chat.json", "formal.jsonl", "operator-authority.json", "launch.lock"} {
		b, _ := os.ReadFile(filepath.Join(f.root, base, name))
		files[base+name] = minimaxlive.Hash(b)
		f.history[strings.TrimPrefix(base+name, ".evidence/")] = b
	}
	source, _ := minimaxlive.ExecutionSource()
	rows, _ := minimaxlive.ReadLedger(path)
	budget := minimaxlive.Budget{}
	for _, r := range rows {
		if r.Kind == "reserve" {
			budget.Calls++
			budget.ReservedAttempts++
			budget.RequestedOutput += r.Output
			budget.ReservedInput += r.Input
		} else {
			budget.Attempts++
		}
	}
	bundle := ".evidence/provider-sites/final-acceptance-20261020/remaining/mm-global-formal-m3-chat/failed-chat"
	report := minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Mode: "live", Stage: "formal", Identity: id, Outcome: "FAIL", AuditPassed: true, Complete: false, Session: "failed-chat", GitCommit: f.commit, FormalRenewalID: finalRenewalID, FormalRenewalGrantSHA256: f.grantPin, Budget: budget, Cumulative: budget, ExecutionSource: source}
	reportRaw := write(bundle+"/minimax-report.json", report)
	unknown := map[string]any{"state": "UNKNOWN", "status": "UNKNOWN", "calls": nil, "attempts": nil}
	write(".evidence/minimax-launches/"+finalRenewalID+"-mm-global-formal-m3-chat.json", map[string]any{"state": "COMPLETED", "route": finalRenewalID + "-mm-global-formal-m3-chat", "baseName": "mm-global-formal-m3-chat", "validationRound": finalRenewalID, "renewalId": finalRenewalID, "renewalGrantSha256": f.grantPin, "sourceCommit": f.commit, "exitCode": 1, "credentialsRead": true, "budget": unknown})
	processes := []any{}
	for _, name := range strings.Fields("formal-zai-chat formal-zai-responses formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages") {
		processes = append(processes, map[string]any{"name": name, "state": "COMPLETED", "exitCode": 0})
	}
	processes = append(processes, map[string]any{"name": "mm-global-formal-m3-chat", "state": "COMPLETED", "exitCode": 1, "parentBudget": unknown, "wholeReservedMicrousd": 200000, "nativeBundle": bundle, "nativeReportSHA256": minimaxlive.Hash(reportRaw), "nativeOutcome": "FAIL", "nativeComplete": false, "nativeAuditPassed": true, "nativeBudget": budget, "nativeConsumerReceiptSHA256": nil})
	for len(files) < 234 {
		write(fmt.Sprintf(".evidence/provider-sites/final-acceptance-20261020/remaining/simulated-pin-%03d.json", len(files)), map[string]any{"fixture": true})
	}
	stop := map[string]any{"schema": 1, "kind": "ACTUAL_REMAINING_STOP_AFTER_REAL_GLOBAL_MINIMAX_CHAT_THINKING_FAILURE", "sourceCommit": f.commit, "regionalFormalProcesses": 6, "regionalFormalScenarios": 81, "regionalFormalCalls": 81, "regionalFormalHTTP": 81, "globalMiniMaxFailedProcesses": 1, "globalMiniMaxActualCalls": 13, "globalMiniMaxActualHTTP": 12, "campaignActualCalls": 303, "campaignActualHTTP": 302, "campaignClaimedCalls": 456, "campaignClaimedHTTP": 492, "wholeReservedMicrousd": 10700000, "refundMicrousd": 0, "unknownParentProcesses": 2, "originalFailedGLMUnchanged": true, "automaticReruns": false, "unstartedProcesses": strings.Fields("openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 mm-global-formal-m3-responses mm-global-formal-m3-messages text-jev text-luna visual-jev visual-decisions"), "processes": processes, "files": files, "releasePass": false, "race": "NOT_RUN", "newREADY": false, "proxy": "http://127.0.0.1:7890", "credentialsReadByThisAudit": false}
	raw := write(stoppedIndex, stop)
	write(secondClosure, map[string]any{"schema": 1, "kind": "CLOSED_SECOND_COHORT_HISTORY_SEAL", "roundId": finalRenewalID, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "stopIndex": stoppedIndex, "stopIndexSHA256": minimaxlive.Hash(raw), "actualCalls": 303, "actualHTTPAttempts": 302, "claimedCalls": 456, "claimedHTTPAttempts": 492, "wholeReservedMicrousd": 10700000, "refundMicrousd": 0, "noAutomaticReruns": true, "paidContinuationAuthorized": false})
}

func (f *finalRenewalFixture) installRecoveryGrant(t *testing.T, facts map[string]any) {
	t.Helper()
	f.base = filepath.Join(f.root, ".evidence/minimax-renewals", recoveryID)
	for k, v := range facts {
		if k != "closedHistory" && k != "recoveryHistorySha256" {
			f.grant[k] = v
		}
	}
	f.grant["schema"] = 4
	f.grant["decision"] = "APPROVED"
	recovery, _ := json.Marshal(facts["recovery"])
	r := facts["recovery"].(map[string]any)
	human := map[string]any{"decision": "APPROVED", "renewalId": recoveryID, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "spendingGuardPolicyId": finalRenewalID, "maxCalls": 506, "maxAttempts": 518, "campaignMaxCalls": 962, "campaignMaxAttempts": 1010, "newReservedMicrousd": 19200000, "retainedReservedMicrousd": 10700000, "campaignMaximumMicrousd": 30000000, "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "historySHA256": facts["closedHistorySha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery), "planSHA256": f.grant["planSha256"], "requestReference": "synthetic-private-no-paid", "requestText": "synthetic fixture only", "provider": "minimax", "model": minimaxlive.M3, "purpose": "standard", "routes": []string{"chat", "responses", "messages"}, "renewalMaxCalls": 60, "renewalMaxAttempts": 72}
	review := map[string]any{"decision": "APPROVED", "renewalId": recoveryID, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "historySHA256": facts["closedHistorySha256"], "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery)}
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(f.write(t, "human-approval.json", human))
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(f.write(t, "source-review.json", review))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
}

func recoveryMutation(t *testing.T, f *finalRenewalFixture, mode string) bool {
	t.Helper()
	mutate := func(path, key string, value any) {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var x map[string]any
		json.Unmarshal(b, &x)
		x[key] = value
		b, _ = json.Marshal(x)
		os.WriteFile(path, b, 0600)
	}
	human := filepath.Join(f.base, "human-approval.json")
	review := filepath.Join(f.base, "source-review.json")
	old := filepath.Join(f.root, ".evidence/minimax-renewals", finalRenewalID)
	switch mode {
	case "pending":
		f.grant["decision"] = "PENDING"
	case "missing", "preflight-no-grant":
		os.Remove(filepath.Join(f.base, "grant.json"))
		return true
	case "human-missing":
		os.Remove(human)
		return true
	case "human-pending":
		mutate(human, "decision", "PENDING")
	case "human-scope":
		mutate(human, "maxCalls", 507)
	case "human-old":
		b, _ := os.ReadFile(filepath.Join(old, "human-approval.json"))
		os.WriteFile(human, b, 0600)
	case "human-source":
		mutate(human, "nativeRenewalId", finalRenewalID)
	case "policy":
		mutate(human, "spendingGuardPolicyId", recoveryID)
	case "review-source":
		mutate(review, "sourceCommit", strings.Repeat("a", 40))
	case "source":
		f.grant["executionSourceSha256"] = strings.Repeat("0", 64)
	case "owner":
		f.grant["commonRootSha256"] = strings.Repeat("0", 64)
	case "unknown-grant":
		f.grant["nativeRenewalId"] = finalRenewalID
	case "seal":
		mutate(filepath.Join(f.root, secondClosure), "wholeReservedMicrousd", 0)
		return true
	case "stop-pin":
		os.WriteFile(filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/remaining/simulated-pin-009.json"), []byte("changed"), 0600)
		return true
	case "old-ledger":
		p := filepath.Join(old, "formal.jsonl")
		b, _ := os.ReadFile(p)
		os.WriteFile(p, append(b, '\n'), 0600)
		return true
	case "old-tail":
		f.grant["recovery"].(map[string]any)["secondLedgerTailSha256"] = strings.Repeat("0", 64)
	case "failed-report":
		mutate(filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/remaining/mm-global-formal-m3-chat/failed-chat/minimax-report.json"), "outcome", "PASS")
		return true
	case "parent-unknown":
		mutate(filepath.Join(f.root, ".evidence/minimax-launches/"+finalRenewalID+"-mm-global-formal-m3-chat.json"), "budget", map[string]any{"state": "REPORTED", "status": "PASS", "calls": 13, "attempts": 12})
		return true
	case "old2-inactive":
		os.WriteFile(filepath.Join(old, "responses.json"), []byte("{}"), 0600)
		return true
	case "caps":
		f.grant["perIdentityCalls"] = 21
	case "fifo":
		p := filepath.Join(f.root, secondClosure)
		os.Remove(p)
		if e := execFixtureFIFO(p); e != nil {
			t.Skip(e)
		}
		return true
	case "linked":
		p := filepath.Join(f.root, secondClosure)
		b, _ := os.ReadFile(p)
		other := filepath.Join(f.root, "copied.json")
		os.WriteFile(other, b, 0600)
		os.Remove(p)
		os.Symlink(other, p)
		return true
	default:
		return false
	}
	b, _ := os.ReadFile(human)
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(b)
	b, _ = os.ReadFile(review)
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(b)
	f.write(t, "grant.json", f.grant)
	return true
}

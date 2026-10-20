package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

const thirdClosure = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-3/seal.json"
const thirdStop = ".evidence/provider-sites/final-acceptance-20261020/recovery3-stopped/index.json"
const thirdOut = ".evidence/provider-sites/final-acceptance-20261020/recovery3/"
const thirdActualHuman = ".evidence/provider-sites/final-acceptance-20261020/recovery3-authorization.json"

func (f *finalRenewalFixture) seedThirdCohort(t *testing.T) {
	t.Helper()
	facts, e := minimaxlive.RecoveryFacts()
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(facts)
	var m map[string]any
	json.Unmarshal(raw, &m)
	f.installRecoveryGrant(t, m)
	os.Setenv("BARNESS_MINIMAX_FORMAL_RENEWAL", recoveryID)
	path, e := minimaxlive.StageLedgerPath("formal")
	if e != nil {
		t.Fatal(e)
	}
	files := map[string]string{}
	write := func(n string, v any) []byte {
		b, _ := json.Marshal(v)
		p := filepath.Join(f.root, n)
		os.MkdirAll(filepath.Dir(p), 0700)
		if e := os.WriteFile(p, b, 0600); e != nil {
			t.Fatal(e)
		}
		files[n] = minimaxlive.Hash(b)
		return b
	}
	actual := write(thirdActualHuman, map[string]any{"decision": "APPROVED", "maxUSD": 30, "round": recoveryID, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "sourceCommit": f.commit, "maxCalls": 506, "maxAttempts": 518, "campaignMaxCalls": 962, "campaignMaxAttempts": 1010, "newReservedMicrousd": 19200000, "retainedReservedMicrousd": 10700000, "requestReference": "synthetic-private-no-paid", "requestText": "synthetic fixture only"})
	source, _ := minimaxlive.ExecutionSource()
	archive := write(thirdOut+"source-at-start.json", map[string]any{"sourceCommit": f.commit, "files": source})
	f.grant["sourceManifestSha256"] = minimaxlive.Hash(archive)
	reviewRaw, _ := os.ReadFile(filepath.Join(f.base, "source-review.json"))
	var review map[string]any
	json.Unmarshal(reviewRaw, &review)
	review["sourceManifestSHA256"] = minimaxlive.Hash(archive)
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(f.write(t, "source-review.json", review))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
	actual = write(thirdActualHuman, map[string]any{"decision": "APPROVED", "maxUSD": 30, "round": recoveryID, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "maxCalls": 506, "maxAttempts": 518, "campaignMaxCalls": 962, "campaignMaxAttempts": 1010, "newReservedMicrousd": 19200000, "retainedReservedMicrousd": 10700000, "requestReference": "synthetic-private-no-paid", "requestText": "synthetic fixture only"})
	humanRaw, _ := os.ReadFile(filepath.Join(f.base, "human-approval.json"))
	var human map[string]any
	json.Unmarshal(humanRaw, &human)
	human["actualHumanApprovalSHA256"] = minimaxlive.Hash(actual)
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(f.write(t, "human-approval.json", human))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
	processes := []any{}
	// Closed ID3 is an independent historical fixture, just like the original
	// predecessor fixture. Do not exercise its active verifier75 times per case;
	// the public operations under test are facts/preflight and fresh ID4 admission.
	previous := ""
	ledger := []byte{}
	thirdRows := []minimaxlive.Receipt{}
	for _, row := range f.original {
		row.RenewalID = recoveryID
		row.RenewalGrantSHA256 = f.grantPin
		row.Session = "third-" + row.Identity.Route
		row.Previous = previous
		row.Hash = ""
		b, _ := json.Marshal(row)
		row.Hash = minimaxlive.Hash(b)
		previous = row.Hash
		thirdRows = append(thirdRows, row)
		b, _ = json.Marshal(row)
		ledger = append(ledger, append(b, '\n')...)
	}
	os.WriteFile(path, ledger, 0600)
	for _, route := range []string{"chat", "responses", "messages"} {
		id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
		session := "third-" + route
		claim := map[string]any{"schema": 2, "renewalId": recoveryID, "grantSha256": f.grantPin, "session": session, "identity": id, "processNonce": minimaxlive.Hash([]byte("synthetic-" + route))}
		write(".evidence/minimax-renewals/"+recoveryID+"/"+route+".json", claim)
		calls := []minimaxlive.Call{}
		for _, p := range minimaxlive.Plan(id) {
			call := minimaxlive.Call{ID: p.ID, Outcome: "PASS"}
			for _, row := range thirdRows {
				if row.Identity == id && row.Case == p.ID {
					if row.Kind == "reserve" {
						call.Reservation = row
					} else {
						call.Attempts = append(call.Attempts, row)
					}
				}
			}
			calls = append(calls, call)
		}
		rows, _ := minimaxlive.ReadLedger(path)
		budget := minimaxlive.Budget{}
		for _, r := range rows {
			if r.Identity == id {
				if r.Kind == "reserve" {
					budget.Calls++
					budget.ReservedAttempts++
					budget.RequestedOutput += r.Output
					budget.ReservedInput += r.Input
				} else {
					budget.Attempts++
				}
			}
		}
		name := "mm-global-formal-m3-" + route
		bundle := thirdOut + name + "/" + session
		report := minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Mode: "live", Stage: "formal", Identity: id, Outcome: "PASS", AuditPassed: true, Complete: true, Session: session, GitCommit: f.commit, FormalRenewalID: recoveryID, FormalRenewalGrantSHA256: f.grantPin, Budget: budget, Cumulative: budget, ExecutionSource: source, Calls: calls, Expected: minimaxlive.Plan(id)}
		rb := write(bundle+"/minimax-report.json", report)
		pb := map[string]any{"state": "REPORTED", "status": "PASS", "calls": 13, "attempts": 12}
		receipt := write(thirdOut+name+"-consume-command.json", map[string]any{"state": "COMPLETED", "exitCode": 0, "credentialsRead": false})
		processes = append(processes, map[string]any{"name": name, "exitCode": 0, "parentBudget": pb, "reservedMicrousd": 200000, "nativeBundle": strings.TrimSuffix(bundle, "/"), "nativeReportSHA256": minimaxlive.Hash(rb), "actualCalls": 13, "actualHTTP": 12, "networkRoute": "http://127.0.0.1:7890", "nativeOutcome": "PASS", "nativeComplete": true, "nativeAuditPassed": true, "nativeConsumerReceiptSHA256": minimaxlive.Hash(receipt)})
		write(".evidence/minimax-launches/"+recoveryID+"-"+name+".json", map[string]any{"state": "COMPLETED", "route": recoveryID + "-" + name, "baseName": name, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "renewalId": recoveryID, "renewalGrantSha256": f.grantPin, "sourceCommit": f.commit, "exitCode": 0, "credentialsRead": true, "budget": pb, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "reservedMicrousd": 200000, "networkRoute": "http://127.0.0.1:7890"})
	}
	os.Unsetenv("BARNESS_MINIMAX_FORMAL_RENEWAL")
	for i, route := range []string{"chat", "responses"} {
		name := "formal-zai-" + route
		bundle := thirdOut + name + "/third-zai-" + route
		exit := i
		pb := map[string]any{"state": "REPORTED", "status": "PASS", "calls": 18, "attempts": 18}
		if i == 1 {
			pb = map[string]any{"state": "UNKNOWN", "status": "UNKNOWN", "calls": nil, "attempts": nil}
		}
		scenarios := []any{}
		for n := 0; n < 18; n++ {
			outcome := "PASS"
			id := fmt.Sprint(n)
			if i == 1 && n == 0 {
				outcome = "FAIL"
				id = "glm-5.3/native-history"
			}
			scenarios = append(scenarios, map[string]any{"id": id, "outcome": outcome, "calls": 1, "attempts": 1})
		}
		rb := write(bundle+"/regional-report.json", map[string]any{"schema": 3, "combo": "zai-" + route, "provider": "zai", "gitCommit": f.commit, "budget": map[string]any{"callsUsed": 18, "httpAttempts": 18}, "scenarios": scenarios, "regional": map[string]any{"mode": "live", "stage": "formal", "complete": i == 0, "auditPassed": true}})
		p := map[string]any{"name": name, "exitCode": exit, "parentBudget": pb, "reservedMicrousd": 650000, "nativeBundle": bundle, "nativeReportSHA256": minimaxlive.Hash(rb), "actualCalls": 18, "actualHTTP": 18, "networkRoute": "http://127.0.0.1:7890", "scenarioOutcomes": map[string]int{"PASS": 18 - i, "FAIL": i, "NOT_RUN": 0}, "failedScenarios": []string{}}
		if i == 0 {
			receipt := write(thirdOut+name+"-consume-command.json", map[string]any{"state": "COMPLETED", "exitCode": 0, "credentialsRead": false})
			p["nativeConsumerReceiptSHA256"] = minimaxlive.Hash(receipt)
		} else {
			p["failedScenarios"] = []string{"glm-5.3/native-history"}
		}
		processes = append(processes, p)
		write(".evidence/minimax-launches/"+recoveryID+"-"+name+".json", map[string]any{"state": "COMPLETED", "route": recoveryID + "-" + name, "baseName": name, "validationRound": recoveryID, "nativeRenewalId": recoveryID, "sourceCommit": f.commit, "exitCode": exit, "credentialsRead": true, "budget": pb, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "reservedMicrousd": 650000, "networkRoute": "http://127.0.0.1:7890"})
	}
	base := ".evidence/minimax-renewals/" + recoveryID + "/"
	write(base+"operator-authority.json", map[string]any{"validationRound": recoveryID})
	os.WriteFile(filepath.Join(f.root, base, "launch.lock"), nil, 0600)
	for _, n := range strings.Fields("grant.json human-approval.json source-review.json operator-authority.json launch.lock chat.json responses.json messages.json formal.jsonl") {
		b, _ := os.ReadFile(filepath.Join(f.root, base, n))
		files[base+n] = minimaxlive.Hash(b)
	}
	for len(files) < 1012 {
		write(fmt.Sprintf(thirdOut+"simulated-%04d.json", len(files)), map[string]any{"fixture": true})
	}
	unused := strings.Fields("formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions")
	stop := write(thirdStop, map[string]any{"schema": 1, "kind": "ACTUAL_RECOVERY3_STOP_AFTER_ZAI_RESPONSES_NATIVE_SOURCE_FAILURE", "sourceCommit": f.commit, "createdAt": "synthetic", "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "processes": processes, "successfulProcesses": 4, "failedProcesses": 1, "newActualCalls": 75, "newActualHTTP": 72, "campaignActualCalls": 378, "campaignActualHTTP": 374, "campaignClaimedCalls": 564, "campaignClaimedHTTP": 612, "newWholeReservedMicrousd": 1900000, "wholeReservedMicrousd": 12600000, "refundMicrousd": 0, "unknownParentProcesses": 3, "originalFailedGLMUnchanged": true, "originalFailedM3Unchanged": true, "automaticReruns": false, "unstartedProcesses": unused, "files": files, "newREADY": false, "race": "NOT_RUN", "releasePass": false, "credentialsReadByThisAudit": false})
	write(thirdClosure, map[string]any{"schema": 1, "kind": "CLOSED_THIRD_COHORT_HISTORY_SEAL", "roundId": recoveryID, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "stopIndex": thirdStop, "stopIndexSHA256": minimaxlive.Hash(stop), "actualCalls": 378, "actualHTTPAttempts": 374, "claimedCalls": 564, "claimedHTTPAttempts": 612, "wholeReservedMicrousd": 12600000, "refundMicrousd": 0, "noAutomaticReruns": true, "paidContinuationAuthorized": false})
	// Change selected source after the closed capture: historical validation must
	// bind the old committed map, not compare it with this new current source.
	os.WriteFile(filepath.Join(f.root, "ai/runtime.go"), []byte("package ai\n// new selected source\n"), 0600)
	for _, args := range [][]string{{"add", "ai/runtime.go"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "current source"}} {
		c := exec.Command("git", args...)
		c.Dir = f.root
		if b, e := c.CombinedOutput(); e != nil {
			t.Fatalf("%v %s", e, b)
		}
	}
	c := exec.Command("git", "rev-parse", "HEAD")
	c.Dir = f.root
	b, _ := c.Output()
	f.commit = strings.TrimSpace(string(b))
}

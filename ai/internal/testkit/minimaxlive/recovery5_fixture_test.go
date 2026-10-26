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

const fourthClosure = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-4/seal.json"
const fourthStop = ".evidence/provider-sites/final-acceptance-20261020/recovery4-stopped/index.json"
const fourthOut = ".evidence/provider-sites/final-acceptance-20261020/recovery4/"
const fourthActualHuman = ".evidence/provider-sites/final-acceptance-20261020/recovery4-authorization.json"

// Every byte is synthetic. This independent owner simulates the exact closed
// fourth shape, never copies a paid capture or creates canonical real authority.
func (f *finalRenewalFixture) seedFourthCohort(t *testing.T) {
	t.Helper()
	facts, e := minimaxlive.Recovery4Facts()
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(facts)
	var m map[string]any
	json.Unmarshal(raw, &m)
	f.installRecovery4Grant(t, m)
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
	source, _ := minimaxlive.ExecutionSource()
	archive := write(fourthOut+"source-at-start.json", map[string]any{"sourceCommit": f.commit, "files": source})
	f.grant["sourceManifestSha256"] = minimaxlive.Hash(archive)
	rb, _ := os.ReadFile(filepath.Join(f.base, "source-review.json"))
	var review map[string]any
	json.Unmarshal(rb, &review)
	review["sourceManifestSHA256"] = minimaxlive.Hash(archive)
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(f.write(t, "source-review.json", review))
	actualPath := filepath.Join(f.root, fourthActualHuman)
	ab, _ := os.ReadFile(actualPath)
	var actual map[string]any
	json.Unmarshal(ab, &actual)
	actual["sourceManifestSHA256"] = minimaxlive.Hash(archive)
	actual["sourceCommit"] = f.commit
	ab = write(fourthActualHuman, actual)
	hb, _ := os.ReadFile(filepath.Join(f.base, "human-approval.json"))
	var human map[string]any
	json.Unmarshal(hb, &human)
	human["actualHumanApprovalSHA256"] = minimaxlive.Hash(ab)
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(f.write(t, "human-approval.json", human))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
	path := filepath.Join(f.base, "formal.jsonl")
	rows := []minimaxlive.Receipt{}
	ledger := []byte{}
	previous := ""
	for _, row := range f.original {
		row.RenewalID = recovery4ID
		row.RenewalGrantSHA256 = f.grantPin
		row.Session = "fourth-" + row.Identity.Route
		row.Previous = previous
		row.Hash = ""
		b, _ := json.Marshal(row)
		row.Hash = minimaxlive.Hash(b)
		previous = row.Hash
		rows = append(rows, row)
		b, _ = json.Marshal(row)
		ledger = append(ledger, append(b, '\n')...)
	}
	os.WriteFile(path, ledger, 0600)
	processes := []any{}
	names := strings.Fields("formal-zai-responses mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions")
	counts := []int{18, 13, 13, 13, 18, 18, 9, 9, 9, 14, 14, 20, 8, 10, 13, 3, 2, 9}
	scenarioCounts := []int{18, 13, 13, 13, 18, 18, 9, 9, 9, 8, 10, 11, 6, 8, 11, 3, 2, 9}
	usd := []int64{650000, 200000, 200000, 200000, 650000, 650000, 350000, 350000, 350000, 1500000, 1500000, 1800000, 500000, 500000, 500000, 3500000, 2600000, 200000}
	for i, name := range names {
		calls, attempts := counts[i], counts[i]
		native := strings.HasPrefix(name, "mm-")
		if native {
			attempts = 12
		}
		network := "http://127.0.0.1:7890"
		if i >= 6 && i <= 8 || i == 13 || i == 14 {
			network = "direct"
		}
		bundle := fourthOut + name + "/fourth-" + name
		reportName := "regional-report.json"
		if i >= 9 {
			bundle = fourthOut + "live/" + name + "/fourth-" + name
			reportName = "live-report.json"
		}
		pb := map[string]any{"state": "REPORTED", "status": "PASS", "calls": calls, "attempts": attempts}
		exit := 0
		outcomes := map[string]int{"PASS": scenarioCounts[i]}
		if i == 10 {
			outcomes = map[string]int{"PASS": 9, "UNSUPPORTED": 1}
		}
		if i == 17 {
			outcomes = map[string]int{"FAIL": 9}
			pb = map[string]any{"state": "UNKNOWN", "status": "UNKNOWN", "calls": nil, "attempts": nil}
			exit = 1
		}
		scenarios := []any{}
		for n := 0; n < scenarioCounts[i]; n++ {
			outcome := "PASS"
			if i == 17 {
				outcome = "FAIL"
			}
			if i == 10 && n == 9 {
				outcome = "UNSUPPORTED"
			}
			scenarios = append(scenarios, map[string]any{"id": fmt.Sprint(n), "outcome": outcome, "calls": 1, "attempts": 1})
		}
		combo := strings.TrimPrefix(name, "formal-")
		report := any(map[string]any{"schema": 3, "combo": combo, "gitCommit": f.commit, "budget": map[string]any{"callsUsed": calls, "httpAttempts": attempts}, "scenarios": scenarios, "regional": map[string]any{"mode": "live", "stage": "formal", "complete": true, "auditPassed": true}})
		if native {
			route := strings.TrimPrefix(name, "mm-global-formal-m3-")
			id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
			session := "fourth-" + route
			bundle = fourthOut + name + "/" + session
			reportName = "minimax-report.json"
			write(".evidence/minimax-renewals/"+recovery4ID+"/"+route+".json", map[string]any{"schema": 2, "renewalId": recovery4ID, "grantSha256": f.grantPin, "identity": id, "session": session, "processNonce": minimaxlive.Hash([]byte("fourth-" + route))})
			budget := minimaxlive.Budget{}
			cs := []minimaxlive.Call{}
			for _, row := range rows {
				if row.Identity != id {
					continue
				}
				if row.Kind == "reserve" {
					budget.Calls++
					budget.ReservedAttempts++
					budget.RequestedOutput += row.Output
					budget.ReservedInput += row.Input
				} else {
					budget.Attempts++
				}
			}
			for _, plan := range historicalM3Plan() {
				c := minimaxlive.Call{ID: plan.ID, Outcome: "PASS"}
				for _, row := range rows {
					if row.Identity == id && row.Case == plan.ID {
						if row.Kind == "reserve" {
							c.Reservation = row
						} else {
							c.Attempts = append(c.Attempts, row)
						}
					}
				}
				cs = append(cs, c)
			}
			report = minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Mode: "live", Stage: "formal", Identity: id, Outcome: "PASS", AuditPassed: true, Complete: true, Session: session, GitCommit: f.commit, FormalRenewalID: recovery4ID, FormalRenewalGrantSHA256: f.grantPin, Budget: budget, Cumulative: budget, ExecutionSource: source, Calls: cs, Expected: historicalM3Plan()}
		}
		reportRaw := write(bundle+"/"+reportName, report)
		usdAttempts := attempts
		usdReserved := int64(100)
		if i == 17 {
			usdAttempts = 0
			usdReserved = 0
		}
		write(fourthOut+name+"-usd.json", map[string]any{"schema": 1, "round": finalRenewalID, "maxMicrousd": usd[i], "reservedMicrousd": usdReserved, "attempts": usdAttempts, "basis": "synthetic fixture"})
		p := map[string]any{"name": name, "exitCode": exit, "parentBudget": pb, "reservedMicrousd": usd[i], "nativeBundle": bundle, "nativeReportSHA256": minimaxlive.Hash(reportRaw), "actualCalls": calls, "actualHTTP": attempts, "networkRoute": network, "usdLedgerAttempts": usdAttempts, "usdReservedMicrousd": usdReserved}
		if !native {
			p["scenarioOutcomes"] = outcomes
		}
		if i != 17 {
			receipt := write(fourthOut+name+"-consume-command.json", map[string]any{"state": "COMPLETED", "exitCode": 0, "credentialsRead": false})
			p["nativeConsumerReceiptSHA256"] = minimaxlive.Hash(receipt)
		}
		processes = append(processes, p)
		parent := map[string]any{"state": "COMPLETED", "route": recovery4ID + "-" + name, "baseName": name, "validationRound": recovery4ID, "nativeRenewalId": recovery4ID, "sourceCommit": f.commit, "exitCode": exit, "credentialsRead": true, "budget": pb, "actualHumanApprovalSHA256": minimaxlive.Hash(ab), "reservedMicrousd": usd[i], "networkRoute": network}
		if native {
			parent["renewalId"] = recovery4ID
			parent["renewalGrantSha256"] = f.grantPin
		}
		write(".evidence/minimax-launches/"+recovery4ID+"-"+name+".json", parent)
	}
	write(".evidence/minimax-renewals/"+recovery4ID+"/operator-authority.json", map[string]any{"validationRound": recovery4ID})
	os.WriteFile(filepath.Join(f.base, "launch.lock"), nil, 0600)
	for _, n := range strings.Fields("grant.json human-approval.json source-review.json operator-authority.json launch.lock chat.json responses.json messages.json formal.jsonl") {
		b, _ := os.ReadFile(filepath.Join(f.base, n))
		files[".evidence/minimax-renewals/"+recovery4ID+"/"+n] = minimaxlive.Hash(b)
	}
	for len(files) < 2032 {
		write(fmt.Sprintf(fourthOut+"simulated-%04d.json", len(files)), map[string]any{"fixture": true})
	}
	total := int64(0)
	for name := range files {
		v, _ := os.Stat(filepath.Join(f.root, name))
		total += v.Size()
	}
	stop := write(fourthStop, map[string]any{"schema": 1, "kind": "ACTUAL_RECOVERY4_STOP_AFTER_OPENAI_DECISIONS_PRESEND_BUDGET_FAILURE", "sourceCommit": f.commit, "createdAt": "synthetic", "actualHumanApprovalSHA256": minimaxlive.Hash(ab), "processes": processes, "successfulProcesses": 17, "failedProcesses": 1, "newActualCalls": 213, "newActualHTTP": 210, "campaignActualCalls": 591, "campaignActualHTTP": 584, "campaignClaimedCalls": 980, "campaignClaimedHTTP": 1040, "newWholeReservedMicrousd": 16200000, "wholeReservedMicrousd": 28800000, "refundMicrousd": 0, "unknownParentProcesses": 4, "originalFailedGLMUnchanged": true, "originalFailedM3Unchanged": true, "originalFailedZAIResponsesUnchanged": true, "automaticReruns": false, "unstartedProcesses": strings.Fields("typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions"), "files": files, "newREADY": false, "race": "NOT_RUN", "releasePass": false, "credentialsReadByThisAudit": false, "failedNativeRecordedCalls": 9, "failedNativeRecordedHTTP": 9, "failedUSDReservationMicrousd": 0, "failedUSDTransportReservations": 0, "fileCount": 2032, "fileBytes": total, "httpCountingNote": "synthetic recorder includes presend refusal", "priorStopIndex": thirdStop, "priorStopIndexSHA256": minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, thirdStop)))})
	write(fourthClosure, map[string]any{"schema": 1, "kind": "CLOSED_FOURTH_COHORT_HISTORY_SEAL", "roundId": recovery4ID, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "stopIndex": filepath.Join(f.root, fourthStop), "stopIndexSHA256": minimaxlive.Hash(stop), "actualCalls": 591, "actualHTTPAttempts": 584, "claimedCalls": 980, "claimedHTTPAttempts": 1040, "wholeReservedMicrousd": 28800000, "refundMicrousd": 0, "noAutomaticReruns": true, "paidContinuationAuthorized": false})
	os.WriteFile(filepath.Join(f.root, "ai/runtime.go"), []byte("package ai\n// fifth selected source\n"), 0600)
	for _, args := range [][]string{{"add", "ai/runtime.go"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "fifth source"}} {
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
func mustReadFixture(t *testing.T, p string) []byte {
	t.Helper()
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	return b
}

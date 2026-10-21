package minimaxlive_test

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const recovery7ID = "provider-sites-final-acceptance-7"
const sixthClosure = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-6/seal.json"
const sixthStopPath = ".evidence/provider-sites/final-acceptance-20261020/recovery6-stopped/index.json"
const seventhOut = ".evidence/provider-sites/final-acceptance-20261020/recovery7/"

// The interrupted sixth campaign is deliberately not a successful formal report.
// This separate owner synthesizes its partial hash chain before seventh code exists.
func (f *finalRenewalFixture) seedSixthCohort(t *testing.T) {
	t.Helper()
	facts, e := minimaxlive.Recovery6Facts()
	if e != nil {
		t.Fatal(e)
	}
	b, _ := json.Marshal(facts)
	var m map[string]any
	json.Unmarshal(b, &m)
	f.installRecovery6Grant(t, m)
	actualPath := ".evidence/provider-sites/final-acceptance-20261020/recovery6-authorization.json"
	actual := mustReadFixture(t, filepath.Join(f.root, actualPath))
	id, _ := minimaxlive.Select("chat", minimaxlive.M3, "standard")
	session := "sixth-interrupted-chat"
	bundle := sixthOut + "mm-global-formal-m3-chat/" + session
	failed := bundle + "/MM07-full-stream-rerun-1"
	previous := ""
	ledger := []byte{}
	for i := 0; i < 8; i++ {
		kind := "reserve"
		output, input := 1024, 2048
		if i%2 == 1 {
			kind = "attempt"
			output, input = 0, 0
		}
		caseID := "full-stream"
		if i >= 4 {
			caseID = "full-complete"
		}
		row := minimaxlive.Receipt{RenewalID: recovery6ID, RenewalGrantSHA256: f.grantPin, Session: session, Sequence: i + 1, Stage: "formal", Identity: id, Case: caseID, Kind: kind, Output: output, Input: input, At: time.Unix(int64(i+1), 0).UTC(), Previous: previous}
		raw, _ := json.Marshal(row)
		row.Hash = minimaxlive.Hash(raw)
		previous = row.Hash
		raw, _ = json.Marshal(row)
		ledger = append(ledger, append(raw, '\n')...)
	}
	os.WriteFile(filepath.Join(f.base, "formal.jsonl"), ledger, 0600)
	f.writeFifthFixture(t, ".evidence/minimax-renewals/"+recovery6ID+"/chat.json", map[string]any{"schema": 2, "renewalId": recovery6ID, "grantSha256": f.grantPin, "identity": id, "session": session, "processNonce": minimaxlive.Hash([]byte("sixth-chat-nonce"))})
	f.writeFifthFixture(t, ".evidence/minimax-renewals/"+recovery6ID+"/operator-authority.json", map[string]any{"validationRound": recovery6ID})
	os.WriteFile(filepath.Join(f.base, "launch.lock"), nil, 0600)
	f.writeFifthFixture(t, failed+"/result.json", map[string]any{"result": map[string]any{}, "errorCode": "deadline_exceeded", "errorPhase": "request"})
	f.writeFifthFixture(t, failed+"/assertions.json", map[string]any{"status": "FAIL", "case_id": "MM07-full-stream-rerun-1", "test": "TestMiniMaxFormal/full-stream-rerun-1"})
	f.writeFifthFixture(t, sixthOut+"mm-global-formal-m3-chat-command.json", map[string]any{"state": "COMPLETED", "name": "mm-global-formal-m3-chat", "exitCode": -15, "credentialsRead": true, "validationRound": recovery6ID, "nativeRenewalId": recovery6ID})
	f.writeFifthFixture(t, sixthOut+"mm-global-formal-m3-chat-usd.json", map[string]any{"schema": 1, "round": finalRenewalID, "maxMicrousd": 200000, "reservedMicrousd": 12884, "attempts": 4, "basis": "synthetic reservation, not invoice"})
	f.writeFifthFixture(t, sixthOut+"formal-zai-responses-usd.json", map[string]any{"schema": 1, "round": finalRenewalID, "maxMicrousd": 650000, "reservedMicrousd": 100, "attempts": 18, "basis": "synthetic reservation, not invoice"})
	parents := []any{}
	for i, name := range []string{"formal-zai-responses", "mm-global-formal-m3-chat"} {
		path := ".evidence/minimax-launches/" + recovery6ID + "-" + name + ".json"
		budget := map[string]any{"state": "UNKNOWN", "status": "UNKNOWN", "calls": nil, "attempts": nil}
		exit, reserved := -15, 200000
		var consumer any
		var parent map[string]any
		if i == 0 {
			json.Unmarshal(mustReadFixture(t, filepath.Join(f.root, path)), &parent)
			budget = map[string]any{"state": "REPORTED", "status": "PASS", "calls": 18, "attempts": 18}
			exit, reserved = 0, 650000
			consumer = parent["nativeConsumerReceiptSha256"]
		} else {
			parent = map[string]any{"state": "COMPLETED", "baseName": name, "route": recovery6ID + "-" + name, "validationRound": recovery6ID, "nativeRenewalId": recovery6ID, "sourceCommit": f.commit, "exitCode": exit, "credentialsRead": true, "budget": budget, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "renewalId": recovery6ID, "renewalGrantSha256": f.grantPin}
		}
		parent["reservedMicrousd"], parent["networkRoute"] = reserved, "http://127.0.0.1:7890"
		raw := f.writeFifthFixture(t, path, parent)
		if i == 0 {
			f.grant["firstRouteProof"].(map[string]string)[path] = minimaxlive.Hash(raw)
		}
		parents = append(parents, map[string]any{"name": name, "parent": path, "parentSHA256": minimaxlive.Hash(raw), "parentBudget": budget, "exitCode": exit, "reservedMicrousd": reserved, "nativeConsumerReceiptSHA256": consumer})
	}
	// Updating a first-parent pin changes this grant and therefore every receipt.
	oldPin := f.grantPin
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
	f.repinSixthLedger(t, oldPin, ledger, parents)
	rows, _ := minimaxlive.ReadLedger(filepath.Join(f.base, "formal.jsonl"))
	previous = rows[len(rows)-1].Hash
	link := sixthOut + "no-key-live/synthetic/formal-corrupt-symlink/formal-integrity.json"
	target := filepath.Join(f.root, strings.TrimSuffix(link, "formal-integrity.json")+"seal.json")
	os.MkdirAll(filepath.Dir(target), 0700)
	os.WriteFile(target, []byte("{}"), 0600)
	os.Symlink(target, filepath.Join(f.root, link))
	files := map[string]string{}
	for _, prefix := range []string{sixthOut, ".evidence/minimax-renewals/" + recovery6ID + "/"} {
		filepath.WalkDir(filepath.Join(f.root, prefix), func(path string, d os.DirEntry, e error) error {
			if e != nil {
				t.Fatal(e)
			}
			if !d.IsDir() && d.Type()&os.ModeSymlink == 0 {
				n, _ := filepath.Rel(f.root, path)
				files[n] = minimaxlive.Hash(mustReadFixture(t, path))
			}
			return nil
		})
	}
	files[actualPath] = minimaxlive.Hash(actual)
	for _, p := range parents {
		n := p.(map[string]any)["parent"].(string)
		files[n] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, n)))
	}
	for i := 0; i < 33; i++ {
		n := fmt.Sprintf(".evidence/provider-sites/final-acceptance-20261020/recovery6-preparation/fixture-%02d.json", i)
		files[n] = minimaxlive.Hash(f.writeFifthFixture(t, n, map[string]any{"synthetic": true}))
	}
	for len(files) < 30238 {
		n := fmt.Sprintf(sixthOut+"empty-%05d.json", len(files))
		os.WriteFile(filepath.Join(f.root, n), nil, 0600)
		files[n] = minimaxlive.Hash(nil)
	}
	total := int64(0)
	for n := range files {
		info, _ := os.Stat(filepath.Join(f.root, n))
		total += info.Size()
	}
	stop := f.writeFifthFixture(t, sixthStopPath, map[string]any{"schema": 1, "kind": "ACTUAL_FIXED6_STOP_AFTER_M3_CHAT_FINAL_REQUEST_DEADLINE_CASE_FAILURE", "state": "STOPPED_NO_FURTHER_PAID_LAUNCH", "sourceCommit": f.commit, "newPaidProcesses": 2, "passedProcesses": 1, "interruptedAfterFailedCaseProcesses": 1, "completedFullM3Processes": 0, "m3FinalReport": "ABSENT_INTERRUPTED_AFTER_IRREVERSIBLE_CASE_FAILURE", "m3Raw": filepath.Join(f.root, bundle), "failedFinalCase": filepath.Join(f.root, failed), "failure": map[string]any{"case": "full-stream-rerun-1", "status": "FAIL", "errorCode": "deadline_exceeded", "errorPhase": "request"}, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "newCompletedReportCalls": 18, "newCompletedReportHTTP": 18, "campaignCompletedReportKnownCalls": 720, "campaignCompletedReportKnownHTTP": 710, "interruptedM3NativeLedgerCalls": 4, "interruptedM3NativeLedgerHTTPAllocations": 4, "interruptedM3NativeLedgerRows": 8, "interruptedM3NativeLedgerTail": previous, "interruptedM3USDReservationMicrousd": 12884, "interruptedM3USDTransportReservations": 4, "interruptedM3AmountsAreNotVendorInvoiceOrConfirmedSends": true, "campaignClaimedCalls": 1204, "campaignClaimedHTTPAttempts": 1280, "wholeReservedMicrousd": 32900000, "refundMicrousd": 0, "unknownParentProcesses": 6, "failedParentRemainsUNKNOWN": true, "stopSignal": "SIGTERM", "childNumericExitCode": -15, "publicOperatorNumericExitCode": 1, "automaticReruns": false, "paidContinuationAuthorized": false, "formalREADY": "NOT_RUN", "race": "NOT_RUN", "releasePass": false, "processes": parents, "notRunProcesses": strings.Fields("mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages glm-chat glm-responses glm-messages openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions"), "fileCount": 30238, "fileBytes": total, "nonRegularRejectionFixtures": map[string]any{link: map[string]any{"kind": "symlink", "mode": 41453, "target": target, "followed": false}}, "files": files, "createdAt": "synthetic"})
	f.writeFifthFixture(t, sixthClosure, map[string]any{"schema": 1, "kind": "CLOSED_SIXTH_COHORT_HISTORY_SEAL", "roundId": recovery6ID, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "stopIndex": filepath.Join(f.root, sixthStopPath), "stopIndexSHA256": minimaxlive.Hash(stop), "knownRecordedCalls": 724, "knownRecordedHTTPAllocations": 714, "completedReportKnownCalls": 720, "completedReportKnownHTTP": 710, "interruptedNativeLedgerCalls": 4, "interruptedNativeLedgerHTTPAllocations": 4, "interruptedParentBudget": "UNKNOWN", "interruptedFinalReportAbsent": true, "recordedHTTPAllocationsAreNotConfirmedVendorSends": true, "claimedCalls": 1204, "claimedHTTPAttempts": 1280, "wholeReservedMicrousd": 32900000, "unknownParentProcesses": 6, "refundMicrousd": 0, "noAutomaticReruns": true, "paidContinuationAuthorized": false})
	os.WriteFile(filepath.Join(f.root, "ai/runtime.go"), []byte("package ai\n// seventh selected source\n"), 0600)
	for _, args := range [][]string{{"add", "ai/runtime.go"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "seventh source"}} {
		c := exec.Command("git", args...)
		if out, e := c.CombinedOutput(); e != nil {
			t.Fatalf("%v %s", e, out)
		}
	}
	c := exec.Command("git", "rev-parse", "HEAD")
	b, _ = c.Output()
	f.commit = strings.TrimSpace(string(b))
}

func (f *finalRenewalFixture) repinSixthLedger(t *testing.T, oldPin string, ledger []byte, parents []any) {
	previous := ""
	result := []byte{}
	for _, line := range strings.Split(strings.TrimSpace(string(ledger)), "\n") {
		var r minimaxlive.Receipt
		json.Unmarshal([]byte(line), &r)
		if r.RenewalGrantSHA256 != oldPin {
			t.Fatal("sixth fixture prior grant differs")
		}
		r.RenewalGrantSHA256, r.Previous, r.Hash = f.grantPin, previous, ""
		b, _ := json.Marshal(r)
		r.Hash = minimaxlive.Hash(b)
		previous = r.Hash
		b, _ = json.Marshal(r)
		result = append(result, append(b, '\n')...)
	}
	os.WriteFile(filepath.Join(f.base, "formal.jsonl"), result, 0600)
	for _, name := range []string{"chat.json"} {
		var x map[string]any
		json.Unmarshal(mustReadFixture(t, filepath.Join(f.base, name)), &x)
		x["grantSha256"] = f.grantPin
		f.write(t, name, x)
	}
	p := parents[1].(map[string]any)
	path := p["parent"].(string)
	var x map[string]any
	json.Unmarshal(mustReadFixture(t, filepath.Join(f.root, path)), &x)
	x["renewalGrantSha256"] = f.grantPin
	p["parentSHA256"] = minimaxlive.Hash(f.writeFifthFixture(t, path, x))
}

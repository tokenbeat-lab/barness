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

const recovery6ID = "provider-sites-final-acceptance-6"
const fifthClosure = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-5/seal.json"
const fifthStopPath = ".evidence/provider-sites/final-acceptance-20261020/recovery5-stopped/index.json"
const sixthOut = ".evidence/provider-sites/final-acceptance-20261020/recovery6/"

// Synthetic complete prior5 belongs to this separate owner. Earlier fixture
// builders remain independent, so changing sixth scope cannot rewrite them.
func (f *finalRenewalFixture) seedFifthCohort(t *testing.T) {
	t.Helper()
	facts, e := minimaxlive.Recovery5Facts()
	if e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(facts)
	var m map[string]any
	json.Unmarshal(raw, &m)
	f.installRecovery5Grant(t, m)
	actual := mustReadFixture(t, filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json"))
	// Extend the already valid first parent with the independently frozen5
	// accounting fields, and bind the new bytes before constructing its ledger.
	firstPath := ".evidence/minimax-launches/" + recovery5ID + "-formal-zai-responses.json"
	var first map[string]any
	json.Unmarshal(mustReadFixture(t, filepath.Join(f.root, firstPath)), &first)
	first["reservedMicrousd"], first["networkRoute"] = 650000, "http://127.0.0.1:7890"
	b := f.writeFifthFixture(t, firstPath, first)
	f.grant["firstRouteProof"].(map[string]string)[firstPath] = minimaxlive.Hash(b)
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
	source, _ := minimaxlive.ExecutionSource()
	rows := []minimaxlive.Receipt{}
	ledger := []byte{}
	previous := ""
	for _, row := range f.original {
		row.RenewalID, row.RenewalGrantSHA256, row.Session, row.Previous, row.Hash = recovery5ID, f.grantPin, "fifth-"+row.Identity.Route, previous, ""
		b, _ := json.Marshal(row)
		row.Hash = minimaxlive.Hash(b)
		previous = row.Hash
		rows = append(rows, row)
		b, _ = json.Marshal(row)
		ledger = append(ledger, append(b, '\n')...)
	}
	os.WriteFile(filepath.Join(f.base, "formal.jsonl"), ledger, 0600)
	processes := []any{}
	names := strings.Fields("formal-zai-responses mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages formal-zai-chat formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses")
	for i, name := range names {
		calls := []int{18, 13, 13, 13, 18, 18, 9, 9}[i]
		attempts := calls
		if i >= 1 && i <= 3 {
			attempts = 12
		}
		reserved := []int64{650000, 200000, 200000, 200000, 650000, 650000, 350000, 350000}[i]
		network := "http://127.0.0.1:7890"
		if i >= 6 {
			network = "direct"
		}
		bundle := fifthOut + name + "/fifth-" + name
		parentPath := ".evidence/minimax-launches/" + recovery5ID + "-" + name + ".json"
		parentBudget := map[string]any{"state": "REPORTED", "status": "PASS", "calls": calls, "attempts": attempts}
		exit := 0
		if i == 7 {
			parentBudget = map[string]any{"state": "UNKNOWN", "status": "UNKNOWN", "calls": nil, "attempts": nil}
			exit = 1
		}
		if i == 0 {
			bundle = fifthOut + "formal-zai-responses/fifth-first"
		} else if i >= 1 && i <= 3 {
			route := strings.TrimPrefix(name, "mm-global-formal-m3-")
			id, _ := minimaxlive.Select(route, minimaxlive.M3, "standard")
			session := "fifth-" + route
			bundle = fifthOut + name + "/" + session
			f.writeFifthFixture(t, ".evidence/minimax-renewals/"+recovery5ID+"/"+route+".json", map[string]any{"schema": 2, "renewalId": recovery5ID, "grantSha256": f.grantPin, "identity": id, "session": session, "processNonce": minimaxlive.Hash([]byte("fifth-" + route))})
			budget := minimaxlive.Summarize(rows, id)
			cs := []minimaxlive.Call{}
			for _, p := range minimaxlive.Plan(id) {
				c := minimaxlive.Call{ID: p.ID, Outcome: "PASS"}
				for _, r := range rows {
					if r.Identity == id && r.Case == p.ID {
						if r.Kind == "reserve" {
							c.Reservation = r
						} else {
							c.Attempts = append(c.Attempts, r)
						}
					}
				}
				cs = append(cs, c)
			}
			f.writeFifthFixture(t, bundle+"/minimax-report.json", minimaxlive.Report{Schema: 1, Kind: "minimax-formal", Mode: "live", Stage: "formal", Identity: id, Outcome: "PASS", AuditPassed: true, Complete: true, Session: session, GitCommit: f.commit, FormalRenewalID: recovery5ID, FormalRenewalGrantSHA256: f.grantPin, Budget: budget, Cumulative: budget, ExecutionSource: source, Calls: cs, Expected: minimaxlive.Plan(id)})
		} else {
			scenarios := []any{}
			for n := 0; n < calls; n++ {
				id := fmt.Sprint(n)
				outcome := "PASS"
				if i == 7 && n == 6 {
					id = "MiniMax-M3/native-history"
					outcome = "FAIL"
				}
				scenarios = append(scenarios, map[string]any{"id": id, "outcome": outcome, "calls": 1, "attempts": 1})
			}
			provider := "zai"
			if i >= 6 {
				provider = "minimax-cn"
			}
			f.writeFifthFixture(t, bundle+"/regional-report.json", map[string]any{"schema": 3, "combo": strings.TrimPrefix(name, "formal-"), "provider": provider, "gitCommit": f.commit, "budget": map[string]any{"callsUsed": calls, "httpAttempts": attempts}, "scenarios": scenarios, "regional": map[string]any{"mode": "live", "stage": "formal", "complete": i != 7, "auditPassed": true}})
		}
		if i != 0 {
			parent := map[string]any{"state": "COMPLETED", "route": recovery5ID + "-" + name, "baseName": name, "validationRound": recovery5ID, "nativeRenewalId": recovery5ID, "sourceCommit": f.commit, "exitCode": exit, "credentialsRead": true, "budget": parentBudget, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "reservedMicrousd": reserved, "networkRoute": network}
			if i >= 1 && i <= 3 {
				parent["renewalId"], parent["renewalGrantSha256"] = recovery5ID, f.grantPin
			}
			f.writeFifthFixture(t, parentPath, parent)
			if i != 7 {
				f.writeFifthFixture(t, fifthOut+name+"-consume-command.json", map[string]any{"state": "COMPLETED", "exitCode": 0, "credentialsRead": false})
			}
		}
		usd := int64(100)
		if i == 7 {
			usd = 54238
		}
		f.writeFifthFixture(t, fifthOut+name+"-usd.json", map[string]any{"schema": 1, "round": finalRenewalID, "maxMicrousd": reserved, "reservedMicrousd": usd, "attempts": attempts, "basis": "synthetic fixture"})
		p := map[string]any{"name": name, "canonicalClaim": parentPath, "state": "COMPLETED", "exitCode": exit, "parentBudget": parentBudget, "nativeRecordedCalls": calls, "nativeRecorderHTTP": attempts, "raw": bundle, "reservedMicrousd": reserved}
		if i == 7 {
			p["failedScenarios"] = []any{map[string]any{"id": "MiniMax-M3/native-history", "outcome": "FAIL", "errorCategory": "assertion"}}
		}
		processes = append(processes, p)
	}
	f.writeFifthFixture(t, ".evidence/minimax-renewals/"+recovery5ID+"/operator-authority.json", map[string]any{"validationRound": recovery5ID})
	os.WriteFile(filepath.Join(f.base, "launch.lock"), nil, 0600)
	link := fifthOut + "no-key-live/synthetic/formal-corrupt-symlink/formal-integrity.json"
	target := filepath.Join(f.root, strings.TrimSuffix(link, "formal-integrity.json")+"seal.json")
	os.MkdirAll(filepath.Dir(target), 0700)
	os.WriteFile(target, []byte("{}"), 0600)
	os.Symlink(target, filepath.Join(f.root, link))
	files := map[string]string{}
	filepath.WalkDir(filepath.Join(f.root, fifthOut), func(path string, d os.DirEntry, e error) error {
		if e != nil {
			t.Fatal(e)
		}
		if !d.IsDir() && d.Type()&os.ModeSymlink == 0 {
			n, _ := filepath.Rel(f.root, path)
			files[n] = minimaxlive.Hash(mustReadFixture(t, path))
		}
		return nil
	})
	for _, n := range strings.Fields("grant.json human-approval.json source-review.json operator-authority.json launch.lock chat.json responses.json messages.json formal.jsonl") {
		name := ".evidence/minimax-renewals/" + recovery5ID + "/" + n
		files[name] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, name)))
	}
	for _, name := range names {
		n := ".evidence/minimax-launches/" + recovery5ID + "-" + name + ".json"
		files[n] = minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, n)))
	}
	approval := ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json"
	files[approval] = minimaxlive.Hash(actual)
	for i := 0; i < 31; i++ {
		n := fmt.Sprintf(".evidence/provider-sites/final-acceptance-20261020/recovery5-preparation/actual-execution/fixture-%02d.json", i)
		files[n] = minimaxlive.Hash(f.writeFifthFixture(t, n, map[string]any{"synthetic": true, "credentialsRead": false, "exitCode": 0}))
	}
	for len(files) < 31152 {
		n := fmt.Sprintf(fifthOut+"empty-%05d.json", len(files))
		os.WriteFile(filepath.Join(f.root, n), nil, 0600)
		files[n] = minimaxlive.Hash(nil)
	}
	total := int64(0)
	for name := range files {
		info, _ := os.Stat(filepath.Join(f.root, name))
		total += info.Size()
	}
	stop := f.writeFifthFixture(t, fifthStopPath, map[string]any{"schema": 1, "kind": "ACTUAL_FIXED5_STOP_AFTER_MINIMAX_CN_RESPONSES_NATIVE_HISTORY_FAILURE", "state": "STOPPED_NO_FURTHER_PAID_LAUNCH", "createdAt": "synthetic", "sourceCommit": f.commit, "actualHumanApprovalSHA256": minimaxlive.Hash(actual), "newPaidProcesses": 8, "passedProcesses": 7, "failedProcesses": 1, "notRunProcesses": strings.Fields("formal-minimax-cn-messages glm-chat glm-responses glm-messages openai-decisions typesafe-classifier protocol114 text-jev text-luna visual-jev visual-decisions"), "processes": processes, "newNativeRecordedCalls": 111, "newNativeRecorderHTTP": 108, "campaignNativeRecordedCalls": 702, "campaignNativeRecorderHTTP": 692, "campaignClaimedCalls": 1160, "campaignClaimedHTTP": 1232, "wholeReservedMicrousd": 32050000, "maxMicrousd": 36000000, "refundMicrousd": 0, "unknownParentProcesses": 5, "failedParentRemainsUNKNOWN": true, "automaticReruns": false, "paidContinuationAuthorized": false, "credentialsReadByThisAudit": false, "vendorCallsByThisAudit": 0, "formalREADY": "NOT_RUN", "race": "NOT_RUN", "releasePass": false, "predecessorStopIndex": fourthStop, "predecessorStopIndexSHA256": minimaxlive.Hash(mustReadFixture(t, filepath.Join(f.root, fourthStop))), "fileCount": 31152, "fileBytes": total, "nonRegularRejectionFixtures": map[string]any{link: map[string]any{"kind": "symlink", "mode": 41453, "target": target, "followed": false}}, "files": files})
	f.writeFifthFixture(t, fifthClosure, map[string]any{"schema": 1, "kind": "CLOSED_FIFTH_COHORT_HISTORY_SEAL", "roundId": recovery5ID, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "stopIndex": filepath.Join(f.root, fifthStopPath), "stopIndexSHA256": minimaxlive.Hash(stop), "actualCalls": 702, "actualHTTPAttempts": 692, "claimedCalls": 1160, "claimedHTTPAttempts": 1232, "wholeReservedMicrousd": 32050000, "refundMicrousd": 0, "noAutomaticReruns": true, "paidContinuationAuthorized": false})
	os.WriteFile(filepath.Join(f.root, "ai/runtime.go"), []byte("package ai\n// sixth selected source\n"), 0600)
	for _, args := range [][]string{{"add", "ai/runtime.go"}, {"-c", "user.name=Fixture", "-c", "user.email=fixture@example.invalid", "commit", "-qm", "sixth source"}} {
		c := exec.Command("git", args...)
		if out, e := c.CombinedOutput(); e != nil {
			t.Fatalf("%v %s", e, out)
		}
	}
	c := exec.Command("git", "rev-parse", "HEAD")
	b, _ = c.Output()
	f.commit = strings.TrimSpace(string(b))
}

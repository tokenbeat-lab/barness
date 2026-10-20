package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func (f *finalRenewalFixture) installRecovery5Grant(t *testing.T, facts map[string]any) {
	t.Helper()
	f.base = filepath.Join(f.root, ".evidence/minimax-renewals", recovery5ID)
	for k, v := range facts {
		if k != "closedHistory" && k != "recoveryHistorySha256" && k != "thirdCohortHistorySha256" && k != "fourthCohortHistorySha256" {
			f.grant[k] = v
		}
	}
	f.prepareFifthArtifacts(t)
	f.grant["schema"] = 6
	f.grant["decision"] = "APPROVED"
	f.grant["requestReference"] = "synthetic-private-new5-no-paid"
	recovery, _ := json.Marshal(facts["recovery"])
	r := facts["recovery"].(map[string]any)
	human := map[string]any{"decision": "APPROVED", "renewalId": recovery5ID, "validationRound": recovery5ID, "nativeRenewalId": recovery5ID, "spendingGuardPolicyId": finalRenewalID, "maxCalls": 378, "maxAttempts": 414, "campaignMaxCalls": 1358, "campaignMaxAttempts": 1454, "newReservedMicrousd": 6900000, "retainedReservedMicrousd": 28800000, "campaignMaximumMicrousd": 36000000, "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "historySHA256": facts["closedHistorySha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery), "planSHA256": f.grant["planSha256"], "requestReference": "synthetic-private-new5-no-paid", "requestText": "synthetic fixture only", "provider": "minimax", "model": minimaxlive.M3, "purpose": "standard", "routes": []string{"chat", "responses", "messages"}, "renewalMaxCalls": 60, "renewalMaxAttempts": 72}
	review := map[string]any{"decision": "APPROVED", "renewalId": recovery5ID, "validationRound": recovery5ID, "nativeRenewalId": recovery5ID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "historySHA256": facts["closedHistorySha256"], "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery)}
	actual := map[string]any{"decision": "APPROVED", "maxUSD": 36, "round": recovery5ID, "validationRound": recovery5ID, "nativeRenewalId": recovery5ID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "maxCalls": 378, "maxAttempts": 414, "campaignMaxCalls": 1358, "campaignMaxAttempts": 1454, "newReservedMicrousd": 6900000, "retainedReservedMicrousd": 28800000, "requestReference": human["requestReference"], "requestText": human["requestText"]}
	actualRaw, _ := json.Marshal(actual)
	actualPath := filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json")
	os.WriteFile(actualPath, actualRaw, 0600)
	human["actualHumanApprovalSHA256"] = minimaxlive.Hash(actualRaw)
	human["thirdCohortSealSHA256"] = facts["thirdCohort"].(map[string]any)["cohortSealSha256"]
	human["thirdCohortHistorySHA256"] = facts["thirdCohortHistorySha256"]
	review["thirdCohortSealSHA256"] = human["thirdCohortSealSHA256"]
	review["thirdCohortHistorySHA256"] = human["thirdCohortHistorySHA256"]
	human["fourthCohortSealSHA256"] = facts["fourthCohort"].(map[string]any)["cohortSealSha256"]
	human["fourthCohortHistorySHA256"] = facts["fourthCohortHistorySha256"]
	review["requestReference"] = human["requestReference"]
	review["fourthCohortSealSHA256"] = human["fourthCohortSealSHA256"]
	review["fourthCohortHistorySHA256"] = human["fourthCohortHistorySHA256"]
	human["spendingGuardSourceSHA256"] = facts["spendingGuardSourceSha256"]
	review["spendingGuardSourceSHA256"] = facts["spendingGuardSourceSha256"]
	actual["spendingGuardSourceSHA256"] = facts["spendingGuardSourceSha256"]
	actualRaw, _ = json.Marshal(actual)
	os.WriteFile(actualPath, actualRaw, 0600)
	human["actualHumanApprovalSHA256"] = minimaxlive.Hash(actualRaw)
	human["operatorDescription"] = map[string]any{"note": "synthetic extension, no extra authority"}
	review["operatorDescription"] = map[string]any{"note": "synthetic extension, no extra authority"}
	f.grant["humanApprovalSha256"] = minimaxlive.Hash(f.write(t, "human-approval.json", human))
	f.grant["sourceReviewSha256"] = minimaxlive.Hash(f.write(t, "source-review.json", review))
	f.installFifthFirstRoute(t, human["actualHumanApprovalSHA256"].(string))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
}

func recovery5Mutation(t *testing.T, f *finalRenewalFixture, mode string) bool {
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
	old := filepath.Join(f.root, ".evidence/minimax-renewals", recovery4ID)
	switch mode {
	case "pending":
		f.grant["decision"] = "PENDING"
	case "missing", "preflight-no-grant":
		os.Remove(filepath.Join(f.base, "grant.json"))
		return true
	case "human-missing":
		os.Remove(human)
		return true
	case "human-case-alias", "review-case-alias":
		path := human
		if mode == "review-case-alias" {
			path = review
		}
		mutate(path, "decision", "PENDING")
		b, _ := os.ReadFile(path)
		b = append(b[:len(b)-1], []byte(`,"Decision":"APPROVED"}`)...)
		os.WriteFile(path, b, 0600)
	case "human-duplicate":
		b, _ := os.ReadFile(human)
		b = append([]byte(`{"decision":"PENDING",`), b[1:]...)
		os.WriteFile(human, b, 0600)
	case "review-duplicate":
		b, _ := os.ReadFile(review)
		b = append([]byte(`{"decision":"PENDING",`), b[1:]...)
		os.WriteFile(review, b, 0600)
	case "human-pending":
		mutate(human, "decision", "PENDING")
	case "human-scope":
		mutate(human, "maxCalls", 507)
	case "human-old":
		b, _ := os.ReadFile(filepath.Join(old, "human-approval.json"))
		os.WriteFile(human, b, 0600)
	case "human-reference":
		f.grant["requestReference"] = "synthetic-private-new4-no-paid"
		mutate(review, "requestReference", "synthetic-private-new4-no-paid")
		mutate(human, "requestReference", "synthetic-private-new4-no-paid")
		actual := filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery5-authorization.json")
		mutate(actual, "requestReference", "synthetic-private-new4-no-paid")
		b, _ := os.ReadFile(actual)
		mutate(human, "actualHumanApprovalSHA256", minimaxlive.Hash(b))
	case "human-source":
		mutate(human, "nativeRenewalId", finalRenewalID)
	case "policy":
		mutate(human, "spendingGuardPolicyId", recovery5ID)
	case "review-source":
		mutate(review, "sourceCommit", strings.Repeat("a", 40))
	case "source":
		f.grant["executionSourceSha256"] = strings.Repeat("0", 64)
	case "owner":
		f.grant["commonRootSha256"] = strings.Repeat("0", 64)
	case "unknown-grant":
		f.grant["nativeRenewalId"] = finalRenewalID
	case "seal":
		mutate(filepath.Join(f.root, fourthClosure), "wholeReservedMicrousd", 0)
		return true
	case "stop-pin":
		os.WriteFile(filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery4/simulated-0200.json"), []byte("changed"), 0600)
		return true
	case "old-ledger":
		p := filepath.Join(old, "formal.jsonl")
		b, _ := os.ReadFile(p)
		os.WriteFile(p, append(b, '\n'), 0600)
		return true
	case "old-tail":
		f.grant["fourthCohort"].(map[string]any)["nativeLedgerTailSha256"] = strings.Repeat("0", 64)
	case "passed-report":
		mutate(filepath.Join(f.root, fourthOut, "mm-global-formal-m3-chat/fourth-chat/minimax-report.json"), "outcome", "FAIL")
		return true
	case "failed-report":
		mutate(filepath.Join(f.root, fourthOut+"live/openai-decisions/fourth-openai-decisions/live-report.json"), "outcome", "PASS")
		return true
	case "parent-unknown":
		mutate(filepath.Join(f.root, ".evidence/minimax-launches/"+recovery4ID+"-openai-decisions.json"), "budget", map[string]any{"state": "REPORTED", "status": "PASS", "calls": 13, "attempts": 12})
		return true
	case "old4-inactive":
		os.WriteFile(filepath.Join(old, "responses.json"), []byte("{}"), 0600)
		return true
	case "caps":
		f.grant["perIdentityCalls"] = 21
	case "counters", "forged-pass":
		p := filepath.Join(f.root, fourthStop)
		b, _ := os.ReadFile(p)
		var stop map[string]any
		json.Unmarshal(b, &stop)
		if mode == "counters" {
			stop["campaignActualCalls"] = 592
		} else {
			stop["successfulProcesses"] = 18
			stop["failedProcesses"] = 0
		}
		b, _ = json.Marshal(stop)
		os.WriteFile(p, b, 0600)
		mutate(filepath.Join(f.root, fourthClosure), "stopIndexSHA256", minimaxlive.Hash(b))
		return true
	case "old1-fourth", "old2-fourth", "old3-fourth", "old4-fourth":
		id := renewalID
		if mode == "old2-fourth" {
			id = finalRenewalID
		}
		if mode == "old3-fourth" {
			id = recoveryID
		}
		if mode == "old4-fourth" {
			id = recovery4ID
		}
		mutate(filepath.Join(f.root, ".evidence/minimax-renewals", id, "grant.json"), "fourthCohort", map[string]any{})
		return true
	case "unstarted-claim":
		os.WriteFile(filepath.Join(f.root, ".evidence/minimax-launches/"+recovery4ID+"-typesafe-classifier.json"), []byte("{}"), 0600)
		return true
	case "missing-nonce":
		os.Remove(filepath.Join(old, "chat.json"))
		return true
	case "old-source-commit":
		mutate(filepath.Join(old, "grant.json"), "sourceCommit", strings.Repeat("a", 40))
		return true
	case "archive":
		mutate(filepath.Join(f.root, fourthOut, "source-at-start.json"), "sourceCommit", f.commit)
		return true
	case "human-record":
		mutate(filepath.Join(f.root, fourthActualHuman), "decision", "PENDING")
		return true
	case "human-nested-duplicate":
		b, _ := os.ReadFile(human)
		b = append(b[:len(b)-1], []byte(`,"nested":{"x":1,"x":2}}`)...)
		os.WriteFile(human, b, 0600)
	case "oversize":
		p := filepath.Join(f.root, fourthClosure)
		file, _ := os.Create(p)
		file.Truncate(3 << 20)
		file.Close()
		return true
	case "fifo":
		p := filepath.Join(f.root, fourthClosure)
		os.Remove(p)
		if e := execFixtureFIFO(p); e != nil {
			t.Skip(e)
		}
		return true
	case "linked":
		p := filepath.Join(f.root, fourthClosure)
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

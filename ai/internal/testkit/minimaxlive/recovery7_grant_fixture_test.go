package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func (f *finalRenewalFixture) installRecovery7Grant(t *testing.T, facts map[string]any) {
	t.Helper()
	f.base = filepath.Join(f.root, ".evidence/minimax-renewals", recovery7ID)
	for k, v := range facts {
		if k != "closedHistory" && k != "recoveryHistorySha256" && k != "thirdCohortHistorySha256" && k != "fourthCohortHistorySha256" && k != "fifthCohortHistorySha256" && k != "sixthCohortHistorySha256" {
			f.grant[k] = v
		}
	}
	f.prepareSeventhArtifacts(t)
	f.grant["schema"] = 8
	f.grant["decision"] = "APPROVED"
	f.grant["requestReference"] = "synthetic-private-new7-no-paid"
	recovery, _ := json.Marshal(facts["recovery"])
	r := facts["recovery"].(map[string]any)
	human := map[string]any{"decision": "APPROVED", "renewalId": recovery7ID, "validationRound": recovery7ID, "nativeRenewalId": recovery7ID, "spendingGuardPolicyId": finalRenewalID, "maxCalls": 378, "maxAttempts": 414, "campaignMaxCalls": 1582, "campaignMaxAttempts": 1694, "newReservedMicrousd": 6900000, "retainedReservedMicrousd": 32900000, "campaignMaximumMicrousd": 40000000, "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "historySHA256": facts["closedHistorySha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery), "planSHA256": f.grant["planSha256"], "requestReference": "synthetic-private-new7-no-paid", "requestText": "synthetic fixture only", "provider": "minimax", "model": minimaxlive.M3, "purpose": "standard", "routes": []string{"chat", "responses", "messages"}, "renewalMaxCalls": 60, "renewalMaxAttempts": 72}
	review := map[string]any{"decision": "APPROVED", "renewalId": recovery7ID, "validationRound": recovery7ID, "nativeRenewalId": recovery7ID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "historySHA256": facts["closedHistorySha256"], "firstCohortSealSHA256": facts["firstCohortSealSha256"], "secondCohortSealSHA256": r["secondCohortSealSha256"], "recoveryHistorySHA256": minimaxlive.Hash(recovery)}
	actual := map[string]any{"decision": "APPROVED", "maxUSD": 40, "round": recovery7ID, "validationRound": recovery7ID, "nativeRenewalId": recovery7ID, "spendingGuardPolicyId": finalRenewalID, "sourceCommit": f.commit, "sourceManifestSHA256": f.grant["sourceManifestSha256"], "binariesManifestSHA256": f.grant["binariesManifestSha256"], "planSHA256": f.grant["planSha256"], "maxCalls": 378, "maxAttempts": 414, "campaignMaxCalls": 1582, "campaignMaxAttempts": 1694, "newReservedMicrousd": 6900000, "retainedReservedMicrousd": 32900000, "requestReference": human["requestReference"], "requestText": human["requestText"]}
	actualRaw, _ := json.Marshal(actual)
	actualPath := filepath.Join(f.root, ".evidence/provider-sites/final-acceptance-20261020/recovery7-authorization.json")
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
	human["fifthCohortSealSHA256"] = facts["fifthCohort"].(map[string]any)["cohortSealSha256"]
	human["fifthCohortHistorySHA256"] = facts["fifthCohortHistorySha256"]
	review["fifthCohortSealSHA256"] = human["fifthCohortSealSHA256"]
	review["fifthCohortHistorySHA256"] = human["fifthCohortHistorySHA256"]
	human["sixthCohortSealSHA256"] = facts["sixthCohort"].(map[string]any)["cohortSealSha256"]
	sixthCanonical, _ := json.Marshal(facts["sixthCohort"])
	human["sixthCohortHistorySHA256"] = minimaxlive.Hash(sixthCanonical)
	review["sixthCohortSealSHA256"] = human["sixthCohortSealSHA256"]
	review["sixthCohortHistorySHA256"] = human["sixthCohortHistorySHA256"]
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
	f.installSeventhFirstRoute(t, human["actualHumanApprovalSHA256"].(string))
	f.grantPin = minimaxlive.Hash(f.write(t, "grant.json", f.grant))
}

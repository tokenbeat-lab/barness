package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// This independent owner has no live captures. Its closed cohort supplies the
// same fixed accounting/absence contract without modifying the15 prior facts.
func (f *finalRenewalFixture) seedFirstCohort(t *testing.T, history map[string]string) {
	t.Helper()
	const first = "provider-sites-final-acceptance-1"
	const closed = ".evidence/provider-sites/final-acceptance-20261020/closed-cohort-1/"
	files := map[string]string{}
	write := func(name string, value any) []byte {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(f.root, name)
		if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(path, raw, 0600); err != nil {
			t.Fatal(err)
		}
		files[name] = minimaxlive.Hash(raw)
		return raw
	}
	for name, pin := range history {
		files[".evidence/"+name] = pin
	}
	oldNative := ".evidence/minimax-renewals/" + first + "/"
	write(oldNative+"operator-authority.json", map[string]any{"round": first, "reservedMicrousd": 1500000})
	if err := os.WriteFile(filepath.Join(f.root, oldNative, "launch.lock"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	write(closed+"source/archived.json", map[string]any{"immutable": true, "largeBinary": strings.Repeat("a", 3<<20)})
	nativeAbsent := []string{}
	for _, name := range []string{"grant.json", "human-approval.json", "source-review.json", "formal.jsonl", "chat.json", "responses.json", "messages.json"} {
		nativeAbsent = append(nativeAbsent, oldNative+name)
	}
	unused := []string{}
	for _, job := range strings.Fields("openai-responses openai-chat anthropic-messages google-gemini deepseek-responses deepseek-chat openai-images google-interactions-image openai-decisions typesafe-classifier protocol114 glm-chat glm-responses glm-messages mm-global-formal-m3-chat mm-global-formal-m3-responses mm-global-formal-m3-messages text-jev text-luna visual-jev visual-decisions candidate-zai-responses candidate-minimax-cn-chat candidate-minimax-cn-responses candidate-minimax-cn-messages formal-zai-chat formal-zai-responses formal-zai-messages formal-minimax-cn-chat formal-minimax-cn-responses formal-minimax-cn-messages") {
		unused = append(unused, ".evidence/minimax-launches/"+first+"-"+job+".json")
	}
	facts := map[string]any{"renewalId": first, "originalCalls": 39, "originalAttempts": 36, "previousCalls": 39, "previousAttempts": 36, "closedHistory": history, "closedHistorySha256": f.grant["closedHistorySha256"]}
	factsRaw := write(closed+"native-minimax-history.log", facts)
	cmd := write(closed+"native-minimax-history-command.json", map[string]any{"exitCode": 0, "credentialsRead": false})
	index := map[string]any{"schema": 1, "kind": "CLOSED_FAILED_REGIONAL_CANDIDATE_COHORT", "roundId": first, "sourceCommit": f.commit, "state": "CLOSED_NO_FURTHER_PAID_LAUNCH", "noProcessRestarts": true, "paidContinuationAuthorized": false, "originalNativeM3HistoryUnchanged": true, "actualCalls": 40, "actualHTTPAttempts": 40, "claimedCalls": 52, "claimedHTTPAttempts": 52, "wholeReservedMicrousd": 1500000, "refundMicrousd": 0, "sourceArchive": "source", "files": files, "processes": []any{map[string]any{"qualification": "PASS"}, map[string]any{"qualification": "FAIL"}}, "unstartedCanonicalClaims": unused, "nativeMiniMaxRound1Absent": nativeAbsent}
	indexRaw, _ := json.Marshal(index)
	if err := os.WriteFile(filepath.Join(f.root, closed, "index.json"), indexRaw, 0600); err != nil {
		t.Fatal(err)
	}
	seal := map[string]any{"schema": 1, "kind": "CLOSED_FIRST_COHORT_HISTORY_SEAL", "cohortIndex": "index.json", "cohortIndexSHA256": minimaxlive.Hash(indexRaw), "nativeRound1NeverIssued": true, "remainingPaidScopeRequiresNewApproval": true, "campaignActualCalls": 40, "campaignActualHTTPAttempts": 40, "campaignClaimedCalls": 52, "campaignClaimedHTTPAttempts": 52, "campaignWholeReservedMicrousd": 1500000, "nativeHistory": map[string]any{"facts": "native-minimax-history.log", "factsSHA256": minimaxlive.Hash(factsRaw), "command": "native-minimax-history-command.json", "commandSHA256": minimaxlive.Hash(cmd), "closedHistory": history, "closedHistorySHA256": f.grant["closedHistorySha256"], "originalCalls": 39, "originalAttempts": 36, "previousCalls": 39, "previousAttempts": 36, "countsAreHistoricalNotAdditionalCampaignConsumption": true}}
	sealRaw := write(closed+"seal.json", seal)
	f.grant["firstCohortSealSha256"] = minimaxlive.Hash(sealRaw)
	f.grant["firstCohortActualCalls"], f.grant["firstCohortActualAttempts"] = 40, 40
	f.grant["firstCohortClaimedCalls"], f.grant["firstCohortClaimedAttempts"] = 52, 52
	f.grant["firstCohortReservedMicrousd"], f.grant["nativeRound1NeverIssued"] = 1500000, true
}

package minimaxlive_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

func (f *finalRenewalFixture) prepareSeventhArtifacts(t *testing.T) {
	source, e := minimaxlive.ExecutionSource()
	if e != nil {
		t.Fatal(e)
	}
	b := f.writeFifthFixture(t, seventhOut+"source-at-start.json", map[string]any{"sourceCommit": f.commit, "files": source})
	f.grant["sourceManifestSha256"] = minimaxlive.Hash(b)
	binary := []byte("synthetic nonexecutable consumer, no vendor capture")
	os.WriteFile(filepath.Join(f.root, seventhOut, "regionalverify"), binary, 0600)
	b = f.writeFifthFixture(t, seventhOut+"binaries.json", map[string]string{"regionalverify": minimaxlive.Hash(binary), "live.test": minimaxlive.Hash([]byte("synthetic live binary"))})
	f.grant["binariesManifestSha256"] = minimaxlive.Hash(b)
}
func (f *finalRenewalFixture) installSeventhFirstRoute(t *testing.T, humanPin string) {
	t.Helper()
	bundle := seventhOut + "formal-zai-responses/seventh-first"
	source, _ := minimaxlive.ExecutionSource()
	encoded, _ := json.Marshal(source)
	sourcePin := minimaxlive.Hash(encoded)
	f.writeFifthFixture(t, bundle+"/regional-execution-source.json", source)
	expected := []string{}
	scenarios := []any{}
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, name := range []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "tool-call", "tool-result", "native-history", "cancel", "recovery"} {
			id := model + "/" + name
			expected = append(expected, id)
			scenarios = append(scenarios, map[string]any{"id": id, "outcome": "PASS", "calls": 1, "attempts": 1})
		}
	}
	report := f.writeFifthFixture(t, bundle+"/regional-report.json", map[string]any{"schema": 3, "combo": "zai-responses", "provider": "zai", "gitCommit": f.commit, "expected": expected, "budget": map[string]any{"callsUsed": 18, "httpAttempts": 18}, "scenarios": scenarios, "regional": map[string]any{"schema": 1, "mode": "live", "stage": "formal", "purpose": "standard", "credentialProvider": "zai", "complete": true, "auditPassed": true, "sourceSha256": sourcePin, "buildSourceSha256": sourcePin, "binarySha256": minimaxlive.Hash([]byte("synthetic live binary")), "nativeSourceCases": map[string]string{"glm-5.2": "REGIONAL-zai-responses-glm-5.2-full-stream", "glm-5.3": "REGIONAL-zai-responses-glm-5.3-full-stream"}}})
	consumerPath := seventhOut + "formal-zai-responses-consume-command.json"
	consumer := f.writeFifthFixture(t, consumerPath, map[string]any{"state": "COMPLETED", "name": "formal-zai-responses-consume", "argv": []string{"$REPO/" + seventhOut + "regionalverify", "-formal", "$REPO/" + bundle}, "cwd": "$REPO", "exitCode": 0, "credentialsRead": false, "credentialsInherited": false, "launcherAudit": map[string]any{"passed": true, "findings": []string{}}})
	parentPath := ".evidence/minimax-launches/" + recovery7ID + "-formal-zai-responses.json"
	parent := f.writeFifthFixture(t, parentPath, map[string]any{"state": "COMPLETED", "baseName": "formal-zai-responses", "route": recovery7ID + "-formal-zai-responses", "validationRound": recovery7ID, "nativeRenewalId": recovery7ID, "sourceCommit": f.commit, "exitCode": 0, "credentialsRead": true, "budget": map[string]any{"state": "REPORTED", "status": "PASS", "calls": 18, "attempts": 18}, "sourceManifestSha256": f.grant["sourceManifestSha256"], "binariesManifestSha256": f.grant["binariesManifestSha256"], "planSha256": f.grant["planSha256"], "actualHumanApprovalSHA256": humanPin, "reportedFileSha256": minimaxlive.Hash(report), "nativeConsumerReceiptSha256": minimaxlive.Hash(consumer), "spendingGuardPolicyId": finalRenewalID, "spendingGuardSourceSHA256": f.grant["spendingGuardSourceSha256"]})
	f.grant["firstRouteProof"] = map[string]string{parentPath: minimaxlive.Hash(parent), consumerPath: minimaxlive.Hash(consumer), bundle + "/regional-report.json": minimaxlive.Hash(report)}
}

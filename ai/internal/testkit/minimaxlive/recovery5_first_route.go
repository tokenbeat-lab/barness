package minimaxlive

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"
)

const fifthCollection = ".evidence/provider-sites/final-acceptance-20261020/recovery5/"

// The new native grant is usable only after the independently consumed first
// regional process. These pins bind real receipts, not an invented PASS field;
// the Python launcher remains responsible for running regionalverify itself.
func recovery5FirstRoute(g RenewalGrant, humanPin string) error {
	owner, e := evidenceOwner()
	if e != nil {
		return e
	}
	root, e := os.OpenRoot(owner)
	if e != nil {
		return e
	}
	defer root.Close()
	parentPath := ".evidence/minimax-launches/" + Recovery5RenewalID + "-formal-zai-responses.json"
	consumerPath := fifthCollection + "formal-zai-responses-consume-command.json"
	reportPath := ""
	if len(g.FirstRouteProof) != 3 || !regexpHex64(g.FirstRouteProof[parentPath]) || !regexpHex64(g.FirstRouteProof[consumerPath]) {
		return fmt.Errorf("recovery5 first18 proof requires exact three pins")
	}
	for name, pin := range g.FirstRouteProof {
		if !regexpHex64(pin) {
			return fmt.Errorf("recovery5 first proof pin invalid")
		}
		if name != parentPath && name != consumerPath {
			reportPath = name
		}
	}
	prefix := fifthCollection + "formal-zai-responses/"
	if !strings.HasPrefix(reportPath, prefix) || !strings.HasSuffix(reportPath, "/regional-report.json") {
		return fmt.Errorf("recovery5 first report path differs")
	}
	session := strings.TrimSuffix(strings.TrimPrefix(reportPath, prefix), "/regional-report.json")
	if session == "" || strings.Contains(session, "/") {
		return fmt.Errorf("recovery5 first report session invalid")
	}
	for name, pin := range g.FirstRouteProof {
		got, _, e := cohortFileHash(root, name, maxCohortFile)
		if e != nil || got != pin {
			return fmt.Errorf("recovery5 first proof changed: %s", name)
		}
	}
	var parent struct {
		State                     string            `json:"state"`
		BaseName                  string            `json:"baseName"`
		Route                     string            `json:"route"`
		ValidationRound           string            `json:"validationRound"`
		NativeRenewalID           string            `json:"nativeRenewalId"`
		SourceCommit              string            `json:"sourceCommit"`
		ExitCode                  int               `json:"exitCode"`
		CredentialsRead           bool              `json:"credentialsRead"`
		Budget                    json.RawMessage   `json:"budget"`
		SourceManifestSHA256      string            `json:"sourceManifestSha256"`
		BinariesManifestSHA256    string            `json:"binariesManifestSha256"`
		PlanSHA256                string            `json:"planSha256"`
		ActualHumanApprovalSHA256 string            `json:"actualHumanApprovalSHA256"`
		ReportSHA256              string            `json:"reportedFileSha256"`
		ConsumerSHA256            string            `json:"nativeConsumerReceiptSha256"`
		SpendingGuardSourceSHA256 map[string]string `json:"spendingGuardSourceSHA256"`
		SpendingGuardPolicyID     string            `json:"spendingGuardPolicyId"`
	}
	_, e = cohortJSONProjection(root, parentPath, &parent, "exitCode", "credentialsRead")
	if e != nil || parent.State != "COMPLETED" || parent.BaseName != "formal-zai-responses" || parent.Route != Recovery5RenewalID+"-formal-zai-responses" || parent.ValidationRound != Recovery5RenewalID || parent.NativeRenewalID != Recovery5RenewalID || parent.SourceCommit != g.SourceCommit || parent.ExitCode != 0 || !parent.CredentialsRead || !reportedParentBudget(parent.Budget, 18, 18) || parent.SourceManifestSHA256 != g.SourceManifestSHA256 || parent.BinariesManifestSHA256 != g.BinariesManifestSHA256 || parent.PlanSHA256 != g.PlanSHA256 || parent.ActualHumanApprovalSHA256 != humanPin || parent.ReportSHA256 != g.FirstRouteProof[reportPath] || parent.ConsumerSHA256 != g.FirstRouteProof[consumerPath] || parent.SpendingGuardPolicyID != FinalAcceptanceRenewalID || !reflect.DeepEqual(parent.SpendingGuardSourceSHA256, g.SpendingGuardSourceSHA256) {
		return fmt.Errorf("recovery5 first canonical PASS18 source/authority differs: %v", e)
	}
	var consumer struct {
		State                string   `json:"state"`
		Name                 string   `json:"name"`
		Argv                 []string `json:"argv"`
		Cwd                  string   `json:"cwd"`
		ExitCode             int      `json:"exitCode"`
		CredentialsRead      bool     `json:"credentialsRead"`
		CredentialsInherited bool     `json:"credentialsInherited"`
		LauncherAudit        struct {
			Passed   bool     `json:"passed"`
			Findings []string `json:"findings"`
		} `json:"launcherAudit"`
	}
	_, e = cohortJSONProjection(root, consumerPath, &consumer, "state", "name", "argv", "cwd", "exitCode", "credentialsRead", "credentialsInherited", "launcherAudit")
	expectedArgv := []string{"$REPO/" + fifthCollection + "regionalverify", "-formal", "$REPO/" + strings.TrimSuffix(reportPath, "/regional-report.json")}
	if e != nil || consumer.State != "COMPLETED" || consumer.Name != "formal-zai-responses-consume" || consumer.ExitCode != 0 || consumer.CredentialsRead || consumer.CredentialsInherited || consumer.Cwd != "$REPO" || !reflect.DeepEqual(consumer.Argv, expectedArgv) || !consumer.LauncherAudit.Passed || len(consumer.LauncherAudit.Findings) != 0 {
		return fmt.Errorf("recovery5 first native consumer command differs")
	}
	var binaries map[string]string
	raw, e := cohortJSON(root, fifthCollection+"binaries.json", &binaries)
	if e != nil || Hash(raw) != g.BinariesManifestSHA256 || !regexpHex64(binaries["regionalverify"]) || !regexpHex64(binaries["live.test"]) {
		return fmt.Errorf("recovery5 first frozen binaries differ")
	}
	pin, _, e := cohortFileHash(root, fifthCollection+"regionalverify", maxCohortFile)
	if e != nil || pin != binaries["regionalverify"] {
		return fmt.Errorf("recovery5 first consumer binary changed")
	}
	return recovery5FirstReport(root, g, reportPath, binaries["live.test"])
}
func recovery5FirstReport(root *os.Root, g RenewalGrant, path, binaryPin string) error {
	var r struct {
		Schema    int      `json:"schema"`
		Combo     string   `json:"combo"`
		Provider  string   `json:"provider"`
		GitCommit string   `json:"gitCommit"`
		Expected  []string `json:"expected"`
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
			Schema             int               `json:"schema"`
			Mode               string            `json:"mode"`
			Stage              string            `json:"stage"`
			Purpose            string            `json:"purpose"`
			CredentialProvider string            `json:"credentialProvider"`
			SourceSHA256       string            `json:"sourceSha256"`
			BuildSourceSHA256  string            `json:"buildSourceSha256"`
			BinarySHA256       string            `json:"binarySha256"`
			Complete           bool              `json:"complete"`
			AuditPassed        bool              `json:"auditPassed"`
			NativeSourceCases  map[string]string `json:"nativeSourceCases"`
		} `json:"regional"`
	}
	_, e := cohortJSONProjection(root, path, &r)
	if e != nil || r.Schema != 3 || r.Combo != "zai-responses" || r.Provider != "zai" || r.GitCommit != g.SourceCommit || r.Budget.Calls != 18 || r.Budget.Attempts != 18 || len(r.Scenarios) != 18 || len(r.Expected) != 18 || r.Regional.Schema != 1 || r.Regional.Mode != "live" || r.Regional.Stage != "formal" || r.Regional.Purpose != "standard" || r.Regional.CredentialProvider != "zai" || !r.Regional.Complete || !r.Regional.AuditPassed || r.Regional.BinarySHA256 != binaryPin || r.Regional.SourceSHA256 != r.Regional.BuildSourceSHA256 {
		return fmt.Errorf("recovery5 first native formal18 report differs: %v", e)
	}
	wanted := []string{}
	for _, model := range []string{"glm-5.3", "glm-5.2"} {
		for _, name := range []string{"full-stream", "full-complete", "simple-stream", "simple-complete", "tool-call", "tool-result", "native-history", "cancel", "recovery"} {
			wanted = append(wanted, model+"/"+name)
		}
	}
	if !slices.Equal(r.Expected, wanted) {
		return fmt.Errorf("recovery5 first expected18 changed")
	}
	expected := map[string]bool{}
	for _, id := range r.Expected {
		if id == "" || expected[id] {
			return fmt.Errorf("recovery5 first duplicate expected case")
		}
		expected[id] = true
	}
	for _, s := range r.Scenarios {
		if !expected[s.ID] || s.Outcome != "PASS" || s.Calls != 1 || s.Attempts != 1 {
			return fmt.Errorf("recovery5 first scenario PASS18 differs")
		}
		delete(expected, s.ID)
	}
	want := map[string]string{"glm-5.2": "REGIONAL-zai-responses-glm-5.2-full-stream", "glm-5.3": "REGIONAL-zai-responses-glm-5.3-full-stream"}
	if !reflect.DeepEqual(r.Regional.NativeSourceCases, want) {
		return fmt.Errorf("recovery5 first native case selection differs")
	}
	var source map[string]string
	_, e = cohortJSON(root, strings.TrimSuffix(path, "regional-report.json")+"regional-execution-source.json", &source)
	encoded, _ := json.Marshal(source)
	if e != nil || len(source) == 0 || Hash(encoded) != r.Regional.SourceSHA256 {
		return fmt.Errorf("recovery5 first report source pin differs")
	}
	if e = committedArchive(g.SourceCommit, source); e != nil {
		return e
	}
	return closedHistoricalSource(root, map[string]string{fifthCollection + "source-at-start.json": g.SourceManifestSHA256}, fifthCollection, g)
}

package e2e

import (
	"encoding/json"
	"os/exec"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestRegionalProviderReleaseSourceCLI(t *testing.T) {
	ev := run.Case(t, "RP18-source-cli")
	cmd := exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/regionalverify", "-source")
	cmd.Dir = "../.."
	cmd.Env = regionalChildEnvironment()
	raw, err := cmd.CombinedOutput()
	if !ev.Check("keyless source consumer executes", err == nil, "%s", regionalSafeOutput(raw)) {
		return
	}
	var snapshot struct {
		SourceSHA256 string            `json:"sourceSha256"`
		Files        map[string]string `json:"files"`
	}
	if json.Unmarshal(raw, &snapshot) != nil {
		t.Fatal("invalid source consumer output")
	}
	ev.Record("execution-source", snapshot)
	for _, name := range []string{"ai/cache_key.go", "ai/anthropic_capabilities.go", "ai/chat_compat.go", "ai/responses_capabilities.go", "ai/live/regional_harness_test.go", "ai/internal/testkit/regional/zai-candidates.json", "ai/internal/testkit/regional/minimax-cn-candidates.json", "go.mod", "go.sum"} {
		ev.Check("full selected source retains "+name, len(snapshot.Files[name]) == 64, "missing source")
	}
	ev.Check("digest represents complete selected bytes", len(snapshot.SourceSHA256) == 64, "%s", snapshot.SourceSHA256)
	cmd = exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/regionalverify", "-grant-template", "zai-chat")
	cmd.Dir = "../.."
	cmd.Env = regionalChildEnvironment()
	raw, err = cmd.CombinedOutput()
	var grant supportmatrix.RegionalGrant
	ev.Check("template cannot authorize spend", err == nil && json.Unmarshal(raw, &grant) == nil && grant.Check("zai-chat", time.Now().UTC()) != nil, "template became a grant")
	ev.Record("incomplete-grant-template", grant)
	cmd = exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/regionalverify", "-source", "-grant-template", "zai-chat")
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err = cmd.CombinedOutput()
	ev.Check("ambiguous consumer mode rejected", err != nil, "%s", regionalSafeOutput(raw))
}

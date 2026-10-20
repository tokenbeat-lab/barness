package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// A scoped grant authorizes the reviewed source/catalog/SDK and endpoint,
// rather than letting the next binary silently spend against changed facts.
func TestRegionalProviderReleaseAuthorization(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "regional-live.test")
	build := exec.CommandContext(ctxFor(t), "go", "test", "-c", "-tags", "live", "-o", binary, "./ai/live")
	build.Dir, build.Env = "../..", regionalChildEnvironment()
	if raw, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, regionalSafeOutput(raw))
	}
	cat, _ := regional.CandidateCatalog()
	hash, _ := cat.Hash()
	source, _ := supportmatrix.RegionalExecutionSource()
	sdk, _ := supportmatrix.RegionalSDK("zai-chat")
	for _, field := range []string{"mode", "sourceSha256", "catalogHash", "catalogVersion", "sdk", "endpoint", "purpose", "provider"} {
		t.Run(field, func(t *testing.T) {
			ev := run.Case(t, "RP18-authorization-"+field)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:regional-controlled"})
			defer srv.Close()
			for range 18 {
				srv.Enqueue(regionalTextReply(t, ai.ProviderZAI, ai.APIOpenAICompletions, "glm-5.3"))
			}
			g := map[string]any{"schema": 1, "id": "synthetic-authorization", "contract": supportmatrix.RegionalAuthorization("zai-chat"), "combo": "zai-chat", "provider": "zai", "purpose": "standard", "accountAlias": "synthetic-account", "regionAlias": "synthetic-region", "models": supportmatrix.RegionalModels("zai-chat"), "expiresAt": time.Now().UTC().Add(time.Hour), "maxCalls": 24, "maxAttempts": 24, "maxRequestedOutputTokens": 98304, "maxReservedInputTokens": 120000, "mode": "controlled", "stage": "candidate", "endpoint": "https://api.z.ai/api/paas/v4", "sourceSha256": supportmatrix.RegionalSourceHash(source), "catalogHash": hash, "catalogVersion": cat.Version, "sdk": sdk}
			g[field] = "synthetic-wrong-scope"
			path := filepath.Join(t.TempDir(), "grant.json")
			raw, _ := json.Marshal(g)
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE=zai-chat", "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE=zai", "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(filepath.Dir(run.Dir()), "regional-authorization-native", filepath.Base(run.Dir()), field))
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"failed": err != nil, "output": regionalSafeOutput(raw), "requests": srv.Requests(), "vendorCalls": 0})
			ev.Check("wrong reviewed scope rejected before HTTP", err != nil && len(srv.Requests()) == 0, "%d requests", len(srv.Requests()))
		})
	}
}

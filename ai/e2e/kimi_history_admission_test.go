package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// At the process boundary, invalid scope must fail before reservation or HTTP.
// Every file below is controlled, and privateGrantProcess runs before writing it.
func TestKimiHistorySupplementAdmission(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary, digest := kimiQualificationBinary(t)
	kinds := []string{"full-grant", "old-nine", "full-entry", "timeout", "count", "source", "binary", "baseline", "scene", "provider", "model", "calls", "attempts", "input", "output", "spend", "expired", "purpose", "interval", "duplicate", "unknown", "null-scene"}
	for _, kind := range kinds {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "KHS-admission-"+kind)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			srv.Enqueue(kimiQualificationReply(t, "openai-completions", "xai-origin-history"))
			g, err := supportmatrix.KimiHistoryGrantSnapshot("kimi-cn-chat", strings.Repeat("a", 64))
			if err != nil {
				t.Fatal(err)
			}
			g.Mode, g.ID, g.AccountAlias, g.RegionAlias = "controlled", "bad-"+evidence.Digest([]byte(run.Dir() + t.Name()))[:20], "synthetic-kimi-account", "synthetic-kimi-region"
			g.ExpiresAt, g.BinarySHA256, g.MaxUSD, g.CallIntervalMs = time.Now().UTC().Add(time.Hour), digest, 0.10, 0
			args := []string{"-test.run=^TestKimiHistorySupplement$", "-test.timeout=240s"}
			switch kind {
			case "full-grant", "old-nine":
				full, e := supportmatrix.RegionalGrantSnapshot(g.Combo, "candidate")
				if e != nil {
					t.Fatal(e)
				}
				full.Mode, full.ID, full.AccountAlias, full.RegionAlias, full.ExpiresAt, full.BinarySHA256, full.MaxUSD = g.Mode, g.ID, g.AccountAlias, g.RegionAlias, g.ExpiresAt, digest, 10
				if kind == "old-nine" {
					full.MaxCalls, full.MaxAttempts, full.MaxRequestedOutputTokens = 9, 9, 36864
				}
				g = full
			case "full-entry":
				args[0] = "-test.run=^TestRegionalCandidate$"
			case "timeout":
				args[1] = "-test.timeout=1800s"
			case "count":
				args = append(args, "-test.count=2")
			case "source":
				g.SourceSHA256 = strings.Repeat("b", 64)
			case "binary":
				g.BinarySHA256 = strings.Repeat("b", 64)
			case "baseline":
				g.BaselineManifestSHA256 = "missing"
			case "scene":
				g.ScenarioIDs = []string{"kimi-k3/full-stream"}
			case "provider":
				g.Provider = "moonshot"
			case "model":
				g.Models = []string{"other"}
			case "calls":
				g.MaxCalls = 2
			case "attempts":
				g.MaxAttempts = 2
			case "input":
				g.MaxReservedInputTokens = 8192
			case "output":
				g.MaxRequestedOutputTokens = 8192
			case "spend":
				g.MaxUSD = 0.11
			case "expired":
				g.ExpiresAt = time.Now().UTC().Add(-time.Hour)
			case "purpose":
				g.Purpose = "standard"
			case "interval":
				g.CallIntervalMs = -1
			case "null-scene":
				g.ScenarioIDs = nil
			}
			raw := mustMarshal(t, g)
			if kind == "duplicate" {
				raw = append([]byte(`{"maxCalls":1,`), raw[1:]...)
			}
			if kind == "unknown" {
				raw = append([]byte(`{"allowAll":true,`), raw[1:]...)
			}
			file := filepath.Join(t.TempDir(), "controlled.json")
			cmd := exec.CommandContext(ctxFor(t), binary, args...)
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE=kimi-cn-chat", "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE=moonshot-cn", "BARNESS_REGIONAL_GRANT="+file, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(t.TempDir(), "native"))
			if !privateGrantProcess(t, ev, cmd) {
				return
			}
			if err := os.WriteFile(file, raw, 0600); err != nil {
				t.Fatal(err)
			}
			out, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(out), "requests": srv.Requests(), "vendorCalls": 0})
			ev.Check("reject before HTTP", err != nil && len(srv.Requests()) == 0, "err=%v requests=%d", err, len(srv.Requests()))
		})
	}
}

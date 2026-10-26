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

func TestKimiQualificationAuthorization(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	for _, mode := range []string{"missing-usd", "negative-usd", "tiny-usd", "binary", "source", "provider", "expiry", "mixed-route", "wrong-entry", "controlled-with-live", "old-nine-call"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-authorization-"+mode)
			combo := "kimi-intl-messages"
			route, _ := supportmatrix.RegionalRouteOf(combo)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, route.API, kind))
			}
			grant, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
			if err != nil {
				t.Fatal(err)
			}
			grant.Mode, grant.ID, grant.AccountAlias, grant.RegionAlias = "controlled", "auth-"+mode+"-"+evidence.Digest([]byte(run.Dir()))[:12], "synthetic-kimi-account", "synthetic-kimi-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			grant.MaxUSD = 10
			grant.BinarySHA256 = digest
			switch mode {
			case "old-nine-call":
				grant.MaxCalls, grant.MaxAttempts, grant.MaxRequestedOutputTokens = 9, 9, 36864
			case "missing-usd":
				grant.MaxUSD = 0
			case "negative-usd":
				grant.MaxUSD = -1
			case "tiny-usd":
				grant.MaxUSD = 1e-9
			case "binary":
				grant.BinarySHA256 = strings.Repeat("0", 64)
			case "source":
				grant.SourceSHA256 = strings.Repeat("0", 64)
			case "provider":
				grant.Provider = "moonshot-cn"
			case "expiry":
				grant.ExpiresAt = time.Time{}
			}
			path := filepath.Join(t.TempDir(), "grant.json")
			if err := os.WriteFile(path, mustMarshal(t, grant), 0600); err != nil {
				t.Fatal(err)
			}
			entry := "-test.run=^TestRegionalCandidate$"
			if mode == "wrong-entry" {
				entry = "-test.run=TestRegional"
			}
			cmd := exec.CommandContext(ctxFor(t), binary, entry, "-test.timeout=1800s")
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+filepath.Join(filepath.Dir(run.Dir()), "kimi-refused-native", filepath.Base(run.Dir()), mode))
			if mode == "mixed-route" {
				cmd.Env = append(cmd.Env, "BARNESS_GLM_ROUTE=chat")
			}
			if mode == "controlled-with-live" {
				cmd.Env = append(cmd.Env, "BARNESS_AI_LIVE=1")
			}
			raw, err := cmd.CombinedOutput()
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			ev.Record("requests", srv.Requests())
			ev.Check("unauthorized scope or spend cap rejected before any HTTP", err != nil && len(srv.Requests()) == 0, "%v %s", err, regionalSafeOutput(raw))
		})
	}
}

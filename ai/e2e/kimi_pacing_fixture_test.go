package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func kimiPacingGrant(t *testing.T, combo, digest, name string, interval any) map[string]any {
	t.Helper()
	g, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
	if err != nil {
		t.Fatal(err)
	}
	g.Mode, g.ID, g.AccountAlias, g.RegionAlias = "controlled", "pace-"+name+"-"+evidence.Digest([]byte(run.Dir()))[:12], "synthetic-kimi-account", "synthetic-kimi-region"
	g.ExpiresAt = time.Now().UTC().Add(time.Hour)
	g.MaxUSD = 10
	g.BinarySHA256 = digest
	var value map[string]any
	mustUnmarshal(t, mustMarshal(t, g), &value)
	value["callIntervalMs"] = interval
	return value
}
func kimiPacingCommand(t *testing.T, binary, combo, url string, grant map[string]any) (*exec.Cmd, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "grant.json")
	if err := os.WriteFile(path, mustMarshal(t, grant), 0600); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(filepath.Dir(run.Dir()), "kimi-paced-native", filepath.Base(run.Dir()), t.Name())
	route, _ := supportmatrix.RegionalRouteOf(combo)
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
	cmd.Dir = "../.."
	cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+url, "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+path, "BARNESS_AI_EVIDENCE_DIR="+base)
	return cmd, base
}
func kimiPacingBundle(t *testing.T, base string) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("native bundle count=%d err=%v", len(paths), err)
	}
	return filepath.Dir(paths[0])
}

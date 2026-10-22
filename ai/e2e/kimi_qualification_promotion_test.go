package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestKimiQualificationPromotionBoundary(t *testing.T) {
	ev := run.Case(t, "KIMI-K16-promotion-proof-boundary")
	binary := filepath.Join(t.TempDir(), "regionalpromote")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/live/cmd/regionalpromote")
	build.Dir, build.Env = "../..", regionalChildEnvironment()
	raw, err := build.CombinedOutput()
	if err != nil {
		t.Fatalf("promotion build: %v %s", err, raw)
	}
	missing := []string{"missing-cn-chat", "missing-cn-responses", "missing-cn-messages", "missing-intl-chat", "missing-intl-responses", "missing-intl-messages"}
	cmd := exec.CommandContext(ctxFor(t), binary, append([]string{"-kimi", "-controlled", "-proposal", filepath.Join(t.TempDir(), "proposal.json")}, missing...)...)
	cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
	raw, err = cmd.CombinedOutput()
	ev.Record("consumer", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
	ev.Check("explicit Kimi promotion requires its own actual proofs", err != nil && strings.Contains(string(raw), "Kimi") && !strings.Contains(string(raw), "flag provided but not defined"), "%v %s", err, raw)
	data, err := os.ReadFile("../catalog_kimi.json")
	if !ev.Check("independent inactive production payload exists", err == nil, "%v", err) {
		return
	}
	var payload struct {
		Active  bool
		Version string
		Models  []ai.Model
	}
	mustUnmarshal(t, data, &payload)
	ev.Check("no premature production identity", !payload.Active && payload.Version == "" && len(payload.Models) == 0, "%s", data)
	for _, m := range ai.BuiltinCatalog().Models {
		ev.Check("BuiltinCatalog remains Kimi-free", m.Provider != ai.ProviderMoonshot && m.Provider != ai.ProviderMoonshotCN, "%+v", m)
	}
	var decoded any
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("payload JSON")
	}
}

package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
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
	if !ev.Check("independent production payload exists", err == nil, "%v", err) {
		return
	}
	var payload struct {
		Active           bool
		Version          string
		ProvenanceSHA256 string
		Models           []ai.Model
	}
	mustUnmarshal(t, data, &payload)
	ev.Check("reviewed six-identity activation is embedded", payload.Active && payload.Version == "kimi-k3-formal-2026-10-22.1" && payload.ProvenanceSHA256 == kimi.SnapshotHash() && len(payload.Models) == 6, "%s", data)
	candidate := kimi.CandidateCatalog()
	ev.Check("activation retains exact reviewed model facts", reflect.DeepEqual(payload.Models, candidate.Models), "models=%d", len(payload.Models))
	formal := ai.BuiltinCatalog()
	for _, provider := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			want, exists := candidate.Lookup(provider, api, "kimi-k3")
			got, promoted := formal.Lookup(provider, api, "kimi-k3")
			ev.Check("BuiltinCatalog exposes each reviewed identity", exists && promoted && reflect.DeepEqual(want, got), "provider=%s api=%s promoted=%t", provider, api, promoted)
		}
	}
	kimiModels := 0
	for _, m := range ai.BuiltinCatalog().Models {
		if m.Provider == ai.ProviderMoonshot || m.Provider == ai.ProviderMoonshotCN {
			kimiModels++
		}
	}
	ev.Check("activation adds exactly six Kimi identities", kimiModels == 6, "models=%d", kimiModels)
	var decoded any
	if json.Unmarshal(data, &decoded) != nil {
		t.Fatal("payload JSON")
	}
}

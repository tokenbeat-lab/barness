package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLIRevisedTierPresence(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-revised-tier-presence")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	for _, identity := range []struct{ operation, api, model string }{{"chat", "openai-responses", "gpt-4"}, {"classifier", "openai-decisions", "gpt-6-luna"}} {
		for _, missing := range []string{"inputTokensAbove", "input", "output", "cacheRead", "cacheWrite", "none"} {
			dir := t.TempDir()
			tier := map[string]any{"inputTokensAbove": 100, "input": 1, "output": 2, "cacheRead": 0, "cacheWrite": 0}
			if missing != "none" {
				delete(tier, missing)
			}
			raw, _ := json.Marshal(map[string]any{"cost": map[string]any{"tiers": []any{tier}}})
			definition := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(definition, raw, 0600); err != nil {
				t.Fatal(err)
			}
			ev.Fixture(identity.operation+"-"+missing+".json", raw)
			store := filepath.Join(dir, "store")
			out, exit := invokeCatalogCLI(t, ev, binary, "revise", "--store", store, "--operation", identity.operation, "--provider", "openai", "--api", identity.api, "--model", identity.model, "--definition", definition, "--reason", "review complete tier presence", "--confirm")
			if missing == "none" {
				ev.Check("explicit zero is accepted on revision", exit == 0, "%s exit=%d output=%s", identity.operation, exit, out)
			} else {
				ev.Check("omitted revised tier field refuses before freeze", exit != 0, "%s missing=%s exit=%d output=%s", identity.operation, missing, exit, out)
				_, err := os.Stat(store)
				ev.Check("tier refusal leaves absent store unchanged", os.IsNotExist(err), "%s missing=%s err=%v", identity.operation, missing, err)
			}
		}
	}
}

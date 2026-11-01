package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLITierPresence(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-cli-all-target-tier-rate-presence")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	for _, kind := range []string{"models.dev", "official"} {
		source := filepath.Join(dir, "source-"+kind)
		raw := []byte(maintenanceSourceFixture)
		if kind == "official" {
			raw = []byte("Synthetic official excerpt with context-tier prices.")
		}
		os.WriteFile(source, raw, 0600)
		for _, missing := range []string{"cacheRead", "inputTokensAbove", "none"} {
			tier := map[string]any{"inputTokensAbove": 100, "input": 2, "output": 4, "cacheRead": 0, "cacheWrite": 0}
			if missing != "none" {
				delete(tier, missing)
			}
			definition := map[string]any{"cost": map[string]any{"tiers": []any{tier}}}
			model := "fixture-chat"
			if kind == "official" {
				model = "tier-official"
				definition = map[string]any{"provider": "openai", "api": "openai-completions", "id": model, "name": "Tier official", "reasoning": false, "input": []string{"text"}, "contextWindow": 4096, "maxTokens": 128, "cost": map[string]any{"input": 1, "output": 2, "cacheRead": 0, "cacheWrite": 0, "tiers": []any{tier}}}
			}
			input, _ := json.Marshal(definition)
			path := filepath.Join(dir, "definition.json")
			os.WriteFile(path, input, 0600)
			args := []string{"add", "--store", filepath.Join(dir, kind+"-"+missing), "--source-kind", kind, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--provider", "openai", "--api", "openai-completions", "--model", model, "--definition", path, "--reason", "explicit complete tier rates required", "--confirm"}
			if kind == "official" {
				args = append(args, "--source-url", "https://official.example.test/tier")
			}
			out, exit := invokeCatalogCLI(t, ev, binary, args...)
			if missing == "none" {
				ev.Check("explicit zero remains a supplied target tier rate", exit == 0, "kind=%s exit=%d output=%s", kind, exit, out)
			} else {
				ev.Check("every target tier path rejects omitted threshold or rate", exit != 0, "kind=%s missing=%s exit=%d output=%s", kind, missing, exit, out)
			}
		}
	}
}

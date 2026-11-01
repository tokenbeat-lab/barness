package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLIUnconfirmedPricing(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-cli-lossy-price-requires-definition")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "tier-source.json")
	var fixture map[string]any
	_ = json.Unmarshal([]byte(maintenanceSourceFixture), &fixture)
	models := fixture["providers"].(map[string]any)["openai"].(map[string]any)["models"].(map[string]any)
	cost := models["fixture-chat"].(map[string]any)["cost"].(map[string]any)
	cost["tiers"] = []any{map[string]any{"input": 2, "output": 4, "cache_read": 0, "cache_write": 0, "tier": map[string]any{"type": "context", "size": 272000}}}
	raw, _ := json.Marshal(fixture)
	os.WriteFile(source, raw, 0600)
	ev.Fixture("tier-source.json", raw)
	store := filepath.Join(dir, "store")
	args := []string{"add", "--store", store, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "review projected tier boundary", "--confirm"}
	out, exit := invokeCatalogCLI(t, ev, binary, args...)
	ev.Check("global confirmation cannot invent lost gt/gte boundary", exit != 0, "exit=%d output=%s", exit, out)
	_, err := os.Stat(store)
	ev.Check("lossy tier does not create a candidate", os.IsNotExist(err), "err=%v", err)
	manual := filepath.Join(dir, "tier-definition.json")
	definition := []byte(`{"cost":{"tiers":[{"inputTokensAbove":272000,"input":2,"output":4,"cacheRead":0,"cacheWrite":0}]}}`)
	os.WriteFile(manual, definition, 0600)
	ev.Fixture("tier-definition.json", definition)
	out, exit = invokeCatalogCLI(t, ev, binary, append(args, "--definition", manual)...)
	ev.Check("explicit complete target tier can be reviewed", exit == 0, "exit=%d output=%s", exit, out)
}

func TestCatalogMaintenanceCLISourceBoundaries(t *testing.T) {
	ev := run.Case(t, "CATALOG-T02-cli-source-boundaries")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	for _, alias := range []struct{ Source, Target, Model string }{{"moonshotai", "moonshot", "kimi-k3"}, {"moonshotai-cn", "moonshot-cn", "kimi-k3"}, {"zhipuai", "zhipu", "glm-5.3"}} {
		raw := []byte(`{"providers":{"` + alias.Source + `":{"id":"` + alias.Source + `","models":{"` + alias.Model + `":{"id":"` + alias.Model + `","type":"chat","name":"Serving"}}}},"models":{}}`)
		source := filepath.Join(dir, alias.Source+".json")
		os.WriteFile(source, raw, 0600)
		ev.Fixture(alias.Source+".json", raw)
		args := []string{"revise", "--store", filepath.Join(dir, alias.Source+"-store"), "--operation", "chat", "--provider", alias.Target, "--api", "openai-completions", "--model", alias.Model, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", alias.Source, "--reason", "confirm explicit regional association", "--confirm"}
		out, exit := invokeCatalogCLI(t, ev, binary, args...)
		ev.Check("reviewed alias preserves regional Provider", exit == 0, "%s exit=%d output=%s", alias.Source, exit, out)
		if alias.Target == "moonshot" {
			for i := range args {
				if i > 0 && args[i-1] == "--provider" {
					args[i] = "moonshot-cn"
				}
			}
			out, exit = invokeCatalogCLI(t, ev, binary, args...)
			ev.Check("international source cannot supply China Provider", exit != 0, "exit=%d output=%s", exit, out)
		}
	}
	var fixture map[string]any
	_ = json.Unmarshal([]byte(maintenanceSourceFixture), &fixture)
	models := fixture["providers"].(map[string]any)["openai"].(map[string]any)["models"].(map[string]any)
	chat := models["fixture-chat"].(map[string]any)
	delete(chat, "type")
	chat["status"] = "deprecated"
	decision := map[string]any{"id": "fixture-decision", "type": "decision", "name": "Decision", "modalities": map[string]any{"input": []string{"text"}, "output": []string{"text"}}, "limit": map[string]any{"context": 4096, "output": 0}, "cost": map[string]any{"input": 1, "output": 0, "cache_read": 0, "cache_write": 0}}
	models["fixture-decision"] = decision
	raw, _ := json.Marshal(fixture)
	source := filepath.Join(dir, "types.json")
	os.WriteFile(source, raw, 0600)
	ev.Fixture("types.json", raw)
	for _, api := range []string{"openai-completions", "openai-responses"} {
		out, exit := invokeCatalogCLI(t, ev, binary, "add", "--store", filepath.Join(dir, api), "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", api, "--model", "fixture-chat", "--reason", "confirm absent type and this API explicitly", "--confirm")
		ev.Check("same offering associates only with explicitly selected API", exit == 0, "%s exit=%d output=%s", api, exit, out)
	}
	definition := filepath.Join(dir, "decision.json")
	os.WriteFile(definition, []byte(`{"capabilities":{"kinds":["choice","score","bool"],"minChoices":2,"maxChoices":2,"descriptionFormats":["string"],"minScoreLevels":2,"maxScoreLevels":5}}`), 0600)
	out, exit := invokeCatalogCLI(t, ev, binary, "add", "--store", filepath.Join(dir, "decision"), "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "classifier", "--api", "openai-decisions", "--model", "fixture-decision", "--definition", definition, "--reason", "confirm decision operation and capabilities", "--confirm")
	ev.Check("decision maps to classifier after capability confirmation", exit == 0, "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "add", "--store", filepath.Join(dir, "missing-source"), "--source", filepath.Join(dir, "missing.json"), "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "missing", "--confirm")
	ev.Check("unavailable input reports a structured failure", exit != 0 && bytes.Contains(out, []byte(`"category":"source_unavailable"`)), "exit=%d output=%s", exit, out)
}

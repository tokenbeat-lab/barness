package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLIRawInputTypes(t *testing.T) {
	ev := run.Case(t, "CATALOG-T02-cli-null-and-raw-type-refusal")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	for _, field := range []string{"type", "reasoning"} {
		var fixture map[string]any
		_ = json.Unmarshal([]byte(maintenanceSourceFixture), &fixture)
		model := fixture["providers"].(map[string]any)["openai"].(map[string]any)["models"].(map[string]any)["fixture-chat"].(map[string]any)
		model[field] = nil
		raw, _ := json.Marshal(fixture)
		source := filepath.Join(dir, field+".json")
		os.WriteFile(source, raw, 0600)
		ev.Fixture(field+"-null.json", raw)
		store := filepath.Join(dir, field+"-store")
		out, exit := invokeCatalogCLI(t, ev, binary, "add", "--store", store, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "null is not a default", "--confirm")
		ev.Check("source null is refused rather than becoming a default", exit != 0, "field=%s exit=%d output=%s", field, exit, out)
	}
	for index, raw := range []string{`{"reasoning":null}`, `{"contextWindow":null}`, `{"cost":{"input":null}}`, `{"reasoning":false} {"futureConstraint":true}`, `{"reasoning":false,"reasoning":true}`} {
		definition := filepath.Join(dir, "manual.json")
		os.WriteFile(definition, []byte(raw), 0600)
		out, exit := invokeCatalogCLI(t, ev, binary, "revise", "--store", filepath.Join(dir, "manual-store"), "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--definition", definition, "--reason", "reject raw JSON type loss", "--confirm")
		ev.Check("manual raw invalid constraints survive no canonicalization", exit != 0, "case=%d exit=%d output=%s", index, exit, out)
	}
}

func TestCatalogMaintenanceCLILocalInputRetention(t *testing.T) {
	ev := run.Case(t, "CATALOG-T06-cli-local-input-retention")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	definition := filepath.Join(dir, "definition.json")
	decisions := filepath.Join(dir, "decisions.json")
	manual := []byte("{\n  \"name\": \"Confirmed name\"\n}\n")
	ownership := []byte("{\n \"/name\": {\"ownership\":\"manual\",\"reason\":\"fixed decision input\"}\n}\n")
	os.WriteFile(definition, manual, 0600)
	os.WriteFile(decisions, ownership, 0600)
	ev.Fixture("definition.json", manual)
	ev.Fixture("decisions.json", ownership)
	out, exit := invokeCatalogCLI(t, ev, binary, "revise", "--store", filepath.Join(dir, "store"), "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--definition", definition, "--decisions", decisions, "--reason", "retain exact manual input bytes", "--confirm")
	var result struct {
		Candidate struct {
			LocalInputs []struct {
				Role, Name, Digest string
				Bytes              []byte
			} `json:"localInputs"`
		} `json:"candidate"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("local input originals are frozen with role/name/digest", exit == 0 && len(result.Candidate.LocalInputs) == 2, "exit=%d output=%s", exit, out) {
		return
	}
	for _, input := range result.Candidate.LocalInputs {
		expected := manual
		if input.Role == "decisions" {
			expected = ownership
		}
		ev.Check("exact bytes retained without synthetic URL or fetch time", string(input.Bytes) == string(expected) && input.Digest != "" && input.Name != "", "input=%+v", input)
	}
}

func TestCatalogMaintenanceCLIDecisionPaths(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-cli-field-decision-paths")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source.json")
	decisions := filepath.Join(dir, "decisions.json")
	os.WriteFile(source, []byte(maintenanceSourceFixture), 0600)
	os.WriteFile(decisions, []byte(`{"/cost/ouput":{"ownership":"manual","reason":"intended output override must not be ignored"}}`), 0600)
	store := filepath.Join(dir, "store")
	out, exit := invokeCatalogCLI(t, ev, binary, "add", "--store", store, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--decisions", decisions, "--reason", "validate explicit field ownership paths", "--confirm")
	ev.Check("unresolvable ownership path is refused", exit != 0, "exit=%d output=%s", exit, out)
	_, err := os.Stat(store)
	ev.Check("invalid decision creates no candidate store", os.IsNotExist(err), "err=%v", err)
}

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLIRevise(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-cli-explicit-manual-revision")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	definition := filepath.Join(dir, "overlay.json")
	manual := []byte(`{"name":"Reviewed GPT four","compat":{"supportsStrictMode":false},"cost":{"input":7}}`)
	os.WriteFile(definition, manual, 0600)
	ev.Fixture("manual-overlay.json", manual)
	args := []string{"revise", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--definition", definition, "--reason", "explicitly revise earlier historical decisions", "--confirm"}
	out, exit := invokeCatalogCLI(t, ev, binary, args...)
	var result struct {
		Status    string `json:"status"`
		Candidate struct {
			State struct {
				Providers []struct {
					ID      string `json:"id"`
					Version uint64 `json:"version"`
					Entries []struct {
						Definition json.RawMessage                               `json:"definition"`
						Fields     map[string]struct{ Ownership, Reason string } `json:"fields"`
					} `json:"entries"`
				} `json:"providers"`
			} `json:"state"`
		} `json:"candidate"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("existing identity can be explicitly revised", exit == 0 && result.Status == "awaiting_check", "exit=%d output=%s", exit, out) {
		return
	}
	found := false
	for _, p := range result.Candidate.State.Providers {
		for _, e := range p.Entries {
			var m struct {
				ID, Name, API string
				ContextWindow int
				Compat        struct{ SupportsStrictMode bool }
				Cost          struct{ Input, Output, CacheRead, CacheWrite float64 }
			}
			_ = json.Unmarshal(e.Definition, &m)
			if m.ID != "gpt-4" || m.API != "openai-responses" {
				continue
			}
			found = true
			ev.Check("manual overlay is explicit and preserves other fields", m.Name == "Reviewed GPT four" && !m.Compat.SupportsStrictMode && m.Cost.Input == 7 && m.Cost.Output == 60 && m.ContextWindow == 8192, "model=%+v", m)
			ev.Check("manual field ownership retains reason", e.Fields["/cost/input"].Ownership == "manual" && e.Fields["/cost/input"].Reason != "", "fields=%+v", e.Fields)
		}
	}
	ev.Check("revised identity remains present", found, "identity missing")
	os.WriteFile(definition, []byte(`{"id":"another-model"}`), 0600)
	out, exit = invokeCatalogCLI(t, ev, binary, args...)
	ev.Check("revise cannot silently change identity", exit != 0, "exit=%d output=%s", exit, out)
	os.WriteFile(definition, []byte(`{"futureConstraint":true}`), 0600)
	out, exit = invokeCatalogCLI(t, ev, binary, args...)
	ev.Check("unknown manual constraints are not dropped", exit != 0, "exit=%d output=%s", exit, out)
}

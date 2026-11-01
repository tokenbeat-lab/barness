package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCatalogMaintenanceCLISourceOperation(t *testing.T) {
	ev := run.Case(t, "CATALOG-T02-source-operation-isolation")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "source.json")
	decisions := filepath.Join(dir, "decisions.json")
	ownership := []byte(`{"/cost/input":{"ownership":"source","reason":"confirmed classifier serving input price"}}`)
	os.WriteFile(decisions, ownership, 0600)
	ev.Fixture("ownership.json", ownership)
	writeSource := func(kind string) {
		offering := map[string]any{"id": "gpt-6-luna", "cost": map[string]any{"input": 3, "output": 4, "cache_read": 0, "cache_write": 0}}
		if kind != "missing" {
			offering["type"] = kind
		}
		raw, _ := json.Marshal(map[string]any{"providers": map[string]any{"openai": map[string]any{"id": "openai", "models": map[string]any{"gpt-6-luna": offering}}}, "models": map[string]any{}})
		os.WriteFile(source, raw, 0600)
		ev.Fixture("source-"+kind+".json", raw)
	}
	args := []string{"revise", "--operation", "classifier", "--provider", "openai", "--api", "openai-decisions", "--model", "gpt-6-luna", "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--decisions", decisions, "--reason", "confirm operation-specific source ownership"}
	for _, kind := range []string{"chat", "missing", "image"} {
		writeSource(kind)
		store := filepath.Join(dir, "bad-"+kind)
		command := append([]string{args[0], "--store", store, "--confirm"}, args[1:]...)
		out, exit := invokeCatalogCLI(t, ev, binary, command...)
		ev.Check("cross-operation association is structured conflict", exit != 0 && bytes.Contains(out, []byte(`"category":"mapping_conflict"`)), "kind=%s exit=%d output=%s", kind, exit, out)
		_, err := os.Stat(store)
		ev.Check("association conflict leaves current unchanged", os.IsNotExist(err), "kind=%s err=%v", kind, err)
	}
	writeSource("decision")
	store := filepath.Join(dir, "store")
	candidate := freezeCatalogAction(t, ev, binary, store, args...)
	checkApplyCatalogAction(t, ev, binary, store, candidate)
	before, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	if !ev.Check("approved classifier association exports", exit == 0, "exit=%d", exit) {
		return
	}
	for _, kind := range []string{"chat", "missing", "image"} {
		writeSource(kind)
		out, exit := invokeCatalogCLI(t, ev, binary, "cost-update", "--store", store, "--provider", "openai", "--source", source, "--source-time", "2026-11-01T01:00:00Z", "--source-provider", "openai", "--reason", "source operation changed", "--confirm")
		ev.Check("refresh cannot borrow another operation price", exit != 0 && bytes.Contains(out, []byte(`"category":"mapping_conflict"`)), "kind=%s exit=%d output=%s", kind, exit, out)
		after, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
		ev.Check("refresh conflict preserves approved bytes", exit == 0 && bytes.Equal(before, after), "kind=%s exit=%d", kind, exit)
	}
}

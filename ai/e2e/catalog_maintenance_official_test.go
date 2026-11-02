package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
)

func TestCatalogMaintenanceCLIOfficialSource(t *testing.T) {
	ev := run.Case(t, "CATALOG-T02-cli-official-source-gap")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "official-excerpt.txt")
	definition := filepath.Join(dir, "definition.json")
	capture := []byte("Synthetic official capture: official-new-chat; context4096; output128; USD/M input1 output2 cache-read0 cache-write0.\n")
	manual := []byte(`{"provider":"openai","api":"openai-completions","id":"official-new-chat","name":"Explicit official model","reasoning":false,"input":["text"],"contextWindow":4096,"maxTokens":128,"cost":{"input":1,"output":2,"cacheRead":0,"cacheWrite":0}}`)
	os.WriteFile(source, capture, 0600)
	os.WriteFile(definition, manual, 0600)
	ev.Fixture("official-excerpt.txt", capture)
	ev.Fixture("definition.json", manual)
	store := filepath.Join(dir, "store")
	args := []string{"add", "--operation", "chat", "--provider", "openai", "--api", "openai-completions", "--model", "official-new-chat", "--source-kind", "official", "--source", source, "--source-url", "https://official.example.test/models/new", "--source-time", "2026-11-01T00:00:00Z", "--definition", definition, "--reason", "explicit official supplement for absent serving offering"}
	candidate := freezeCatalogAction(t, ev, binary, store, args...)
	found := false
	for _, p := range candidate.State.Providers {
		for _, entry := range p.Entries {
			var id struct{ ID string }
			_ = json.Unmarshal(entry.Definition, &id)
			if id.ID == "official-new-chat" {
				found = true
				ev.Check("official source is recorded without invented serving association", entry.Source != nil && entry.Source.Kind == "official" && entry.Source.CanonicalID == "", "source=%+v", entry.Source)
				for _, field := range entry.Fields {
					ev.Check("official mapped fields require manual ownership", field.Ownership == "manual" && field.Reason != "", "field=%+v", field)
				}
			}
		}
	}
	ev.Check("full explicit definition retained", found && len(candidate.Inputs) == 1 && string(candidate.Inputs[0].Bytes) == string(capture), "found=%v inputs=%d", found, len(candidate.Inputs))
	revision := checkApplyCatalogAction(t, ev, binary, store, candidate)
	cat, _, err := catalog.Load(revision.Snapshot)
	_, found = cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAICompletions, "official-new-chat")
	ev.Check("official missing offering still requires actual code check", err == nil && found, "err=%v found=%v", err, found)
}

func TestCatalogMaintenanceCLIKnownDefinitionCodeGap(t *testing.T) {
	ev := run.Case(t, "CATALOG-T05-cli-known-definition-code-gap")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	source := filepath.Join(dir, "official.txt")
	definition := filepath.Join(dir, "future.json")
	model, ok := ai.BuiltinCatalog().Lookup(ai.ProviderMoonshot, ai.APIOpenAICompletions, "kimi-k3")
	if !ok {
		t.Fatal("baseline Kimi model missing")
	}
	model.ID, model.Name = "future-kimi", "Synthetic future Kimi"
	// Image is a typed modality outside the Kimi protocol subset.
	model.Input = append(model.Input, ai.ModalityImage)
	raw, _ := json.Marshal(model)
	os.WriteFile(definition, raw, 0600)
	os.WriteFile(source, []byte("Synthetic official source for a future typed model; does not prove vendor support."), 0600)
	candidate := freezeCatalogAction(t, ev, binary, store, "add", "--operation", "chat", "--provider", "moonshot", "--api", "openai-completions", "--model", "future-kimi", "--source-kind", "official", "--source", source, "--source-url", "https://official.example.test/models/future-kimi", "--source-time", "2026-11-01T00:00:00Z", "--definition", definition, "--reason", "retain candidate for an unsupported image capability")
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var result struct {
		Category string                             `json:"category"`
		Result   struct{ Status, Reference string } `json:"result"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("known typed unsupported definition becomes retained code gap", exit != 0 && result.Category == "code_gap" && result.Result.Reference != "", "exit=%d output=%s", exit, out) {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "candidate", "--store", store, "--id", candidate.ID, "--repo", root, "--check", result.Result.Reference)
	var queried struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(out, &queried)
	ev.Check("gap candidate remains queryable with actual failed check", exit == 0 && queried.Status == "code_gap", "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", result.Result.Reference, "--confirm")
	ev.Check("gap evidence cannot authorize apply", exit != 0, "exit=%d output=%s", exit, out)
	_, err := os.Stat(filepath.Join(store, "current.json"))
	ev.Check("effective catalog stays on original baseline", os.IsNotExist(err), "err=%v", err)
}

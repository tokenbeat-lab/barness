package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func TestCatalogMaintenanceCLILifecycle(t *testing.T) {
	ev := run.Case(t, "CATALOG-T07-cli-lifecycle-new-versions")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	store := filepath.Join(t.TempDir(), "store")
	seedVersions := catalogSeedVersions(t, ev, binary, store)
	identity := []string{"--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "explicit lifecycle fixture"}
	withdrawn := freezeCatalogAction(t, ev, binary, store, append([]string{"withdraw"}, identity...)...)
	first := checkApplyCatalogAction(t, ev, binary, store, withdrawn)
	restored := freezeCatalogAction(t, ev, binary, store, append([]string{"restore"}, identity...)...)
	second := checkApplyCatalogAction(t, ev, binary, store, restored)
	cat, _, err := catalog.Load(second.Snapshot)
	_, present := cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-6.1-sol")
	ev.Check("explicit restore is newly checked and available", err == nil && present, "err=%v present=%v", err, present)
	rollback := freezeCatalogAction(t, ev, binary, store, "rollback", "--provider", "openai", "--revision", first.ID, "--reason", "restore prior withdrawn content as new change")
	third := checkApplyCatalogAction(t, ev, binary, store, rollback)
	cat, _, err = catalog.Load(third.Snapshot)
	_, present = cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-6.1-sol")
	ev.Check("rollback restores selected content without builtin fallback", err == nil && !present, "err=%v present=%v", err, present)
	for _, p := range third.State.Providers {
		expected := seedVersions[string(p.ID)]
		if p.ID == ai.ProviderOpenAI {
			expected += 3
		}
		ev.Check("withdraw restore rollback each advance only changed Provider", p.Version == expected, "%s=%d", p.ID, p.Version)
	}
	out, exit := invokeCatalogCLI(t, ev, binary, "history", "--store", store)
	var history struct {
		History []maintenance.Revision `json:"history"`
	}
	_ = json.Unmarshal(out, &history)
	if !ev.Check("history appends all three changes and seed", exit == 0 && len(history.History) == 4, "exit=%d output=%s", exit, out) {
		return
	}
	ev.Check("original history identity and bytes remain unchanged", history.History[2].ID == first.ID && bytes.Equal(history.History[2].Snapshot, first.Snapshot) && history.History[0].ID != first.ID, "history changed")
	out, exit = invokeCatalogCLI(t, ev, binary, "restore", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-6.1-sol", "--reason", "preview only", "--dry-run")
	var preview struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(out, &preview)
	ev.Check("restore preview is independent pending change", exit == 0 && preview.Status == "unconfirmed", "exit=%d output=%s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "versions", "--store", store)
	var versions struct {
		Versions map[string]uint64 `json:"versions"`
	}
	_ = json.Unmarshal(out, &versions)
	ev.Check("preview does not advance current version", exit == 0 && versions.Versions["openai"] == seedVersions["openai"]+3, "exit=%d output=%s", exit, out)
}

func TestCatalogMaintenanceCLIAddedRollbackRecovery(t *testing.T) {
	ev := run.Case(t, "CATALOG-T07-cli-added-rollback-recovery")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	seedVersions := catalogSeedVersions(t, ev, binary, store)
	source := filepath.Join(dir, "source.json")
	if err := os.WriteFile(source, []byte(maintenanceSourceFixture), 0600); err != nil {
		t.Fatal(err)
	}
	ev.Fixture("source.json", []byte(maintenanceSourceFixture))
	added := freezeCatalogAction(t, ev, binary, store, "add", "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--operation", "chat", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "add recoverable serving definition")
	first := checkApplyCatalogAction(t, ev, binary, store, added)
	os.Remove(source)
	rollback := freezeCatalogAction(t, ev, binary, store, "rollback", "--provider", "openai", "--revision", first.Parent, "--reason", "restore earlier active content while retaining later definitions")
	second := checkApplyCatalogAction(t, ev, binary, store, rollback)
	cat, _, err := catalog.Load(second.Snapshot)
	_, present := cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAICompletions, "fixture-chat")
	ev.Check("rollback removes new identity from runtime", err == nil && !present, "err=%v present=%v", err, present)
	retained := false
	for _, p := range second.State.Providers {
		for _, e := range p.Entries {
			var identity struct{ ID string }
			_ = json.Unmarshal(e.Definition, &identity)
			retained = retained || identity.ID == "fixture-chat" && e.Lifecycle == maintenance.Withdrawn
		}
	}
	ev.Check("newer definition stays recoverable after older content restored", retained, "definition erased")
	restored := freezeCatalogAction(t, ev, binary, store, "restore", "--operation", "chat", "--provider", "openai", "--api", "openai-completions", "--model", "fixture-chat", "--reason", "restore retained definition with fresh code check")
	third := checkApplyCatalogAction(t, ev, binary, store, restored)
	cat, _, err = catalog.Load(third.Snapshot)
	_, present = cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAICompletions, "fixture-chat")
	ev.Check("restoration does not require refetching deleted source", err == nil && present, "err=%v present=%v", err, present)
	for _, p := range third.State.Providers {
		if p.ID == ai.ProviderOpenAI {
			ev.Check("add rollback restore advance monotonically", p.Version == seedVersions["openai"]+3, "version=%d", p.Version)
		}
	}
}

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Failure cases: a version below 4.8 remains discoverable or reaches HTTP,
// another provider/retained definition changes, history disappears, or the
// current formal process template still selects a withdrawn model.
func TestAnthropicRetiredBuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "CATALOG-ANTHROPIC-retirement")
	raw, err := os.ReadFile("testdata/catalog/pre-anthropic-retirement.json")
	if err != nil {
		t.Fatal(err)
	}
	var before struct {
		Hash    string
		Catalog ai.Catalog
	}
	mustUnmarshal(t, raw, &before)
	ev.Check("pre-withdrawal snapshot is valid", catalogHash(t, before.Catalog) == before.Hash, "historical snapshot invalid")
	current := ai.BuiltinCatalog()
	keep := []string{"claude-fable-5", "claude-fable-5-1", "claude-opus-4-8", "claude-opus-5", "claude-opus-5-5", "claude-sonnet-5-5", "claude-sonnet-5"}
	var ids []string
	for _, m := range current.ModelsOf(ai.ProviderAnthropic, ai.APIAnthropicMessages) {
		// This retirement proves only identities present in its original snapshot.
		if _, existed := before.Catalog.Lookup(m.Provider, m.API, m.ID); existed {
			ids = append(ids, m.ID)
		}
	}
	ev.Check("exact retained identities and order", slices.Equal(ids, keep), "got=%v want=%v", ids, keep)
	want := ai.Catalog{Version: before.Catalog.Version, ImageModels: before.Catalog.ImageModels, ClassifierModels: before.Catalog.ClassifierModels}
	var retired []ai.Model
	for _, m := range before.Catalog.Models {
		if m.Provider == ai.ProviderAnthropic && !slices.Contains(keep, m.ID) {
			retired = append(retired, m)
			continue
		}
		want.Models = append(want.Models, m)
	}
	ev.Check("original retained definitions survive in their approved order", reflect.DeepEqual(historicalCatalogInCurrentOrder(catalogWithHistoricalNames(current, before.Catalog), before.Catalog), want), "retained definitions, other providers or order changed")
	ev.Check("exact withdrawal count", len(retired) == 9, "count=%d", len(retired))
	state, err := maintenance.BuiltinState()
	if !ev.Check("reviewed state loads", err == nil, "%v", err) {
		return
	}
	for _, old := range retired {
		t.Run(old.ID, func(t *testing.T) {
			e := run.Case(t, "CATALOG-ANTHROPIC-retired-"+old.ID)
			_, found := current.Lookup(old.Provider, old.API, old.ID)
			e.Check("retired identity is absent from discovery", !found, "identity=%s", old.ID)
			preserved := false
			for _, p := range state.Providers {
				for _, entry := range p.Entries {
					var m ai.Model
					if entry.Operation == ai.OperationChat && json.Unmarshal(entry.Definition, &m) == nil && m.Provider == old.Provider && m.API == old.API && m.ID == old.ID {
						preserved = entry.Lifecycle == maintenance.Withdrawn && reflect.DeepEqual(m, old)
					}
				}
			}
			e.Check("withdrawn maintenance definition remains exact", preserved, "historical definition lost or changed")
			i := catalogContractIdentity{ai.OperationChat, old.Provider, old.API, old.ID}
			w, _ := newCatalogContractWorld(t, current, i)
			_, callErr := w.client.CompleteSimple(ctxFor(t), textScope("retired-"+old.ID), catalogContractTarget(i), ai.Request{Messages: []ai.Message{ai.UserText("Retired model must be refused.")}}, ai.SimpleOptions{})
			e.Check("allowlist cannot restore retired model or send HTTP", callErr != nil && len(w.provider.Requests()) == 0, "err=%v requests=%d", callErr, len(w.provider.Requests()))
		})
	}
	scope, err := supportmatrix.ProcessScope("anthropic-messages", "controlled", "formal", strings.Repeat("1", 64))
	ev.Check("current formal process template uses active models", err == nil, "%v", err)
	for _, id := range scope.Models {
		_, found := current.Lookup(ai.ProviderAnthropic, ai.APIAnthropicMessages, id)
		ev.Check("formal process model remains active "+id, found, "retired process target")
	}
	builtin, err := os.ReadFile("../catalog_builtin.json")
	if err != nil {
		t.Fatal(err)
	}
	baseline, err := os.ReadFile("../../catalog/data/baseline/snapshot.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("builtin and adopted review snapshot have identical bytes", bytes.Equal(builtin, baseline), "adoption differs")
	ev.Record("current-catalog", current)
}

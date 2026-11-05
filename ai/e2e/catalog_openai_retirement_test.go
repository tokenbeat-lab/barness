package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// Failure cases: wrong API set, surviving retired ID, accidental change to an
// unrelated definition, lost withdrawal history, or a retired call reaching HTTP.
func TestOpenAIRetiredBuiltinCatalog(t *testing.T) {
	ev := run.Case(t, "CATALOG-OPENAI-retirement")
	before := protocolFixtureCatalog()
	current := ai.BuiltinCatalog()
	expected := map[ai.API][]string{
		ai.APIOpenAIResponses:   {"gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra", "gpt-6-luna", "gpt-6.1-sol", "gpt-6-sol", "gpt-6.1-terra"},
		ai.APIOpenAICompletions: {"gpt-5.6-luna", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-6-astra", "gpt-6-luna", "gpt-6.1-sol", "gpt-6-sol"},
	}
	for api, want := range expected {
		var ids []string
		for _, m := range current.ModelsOf(ai.ProviderOpenAI, api) {
			ids = append(ids, m.ID)
		}
		ev.Check("exact retained API identities "+string(api), slices.Equal(ids, want), "got=%v want=%v", ids, want)
	}
	// Later provider withdrawals do not change this OpenAI retirement proof.
	want, got := []ai.Model{}, []ai.Model{}
	for _, m := range current.Models {
		if m.Provider == ai.ProviderOpenAI {
			got = append(got, m)
		}
	}
	retired := []ai.Model{}
	for _, m := range before.Models {
		if m.Provider == ai.ProviderOpenAI && !slices.Contains(expected[m.API], m.ID) {
			retired = append(retired, m)
			continue
		}
		if m.Provider == ai.ProviderOpenAI {
			want = append(want, m)
		}
	}
	ev.Check("OpenAI retained definitions and relative order remain exact", reflect.DeepEqual(got, want), "retained definitions or relative order changed")
	ev.Check("exact withdrawal count", len(retired) == 37, "count=%d", len(retired))
	state, err := maintenance.BuiltinState()
	if !ev.Check("reviewed state loads", err == nil, "%v", err) {
		return
	}
	for _, old := range retired {
		t.Run(string(old.API)+"/"+old.ID, func(t *testing.T) {
			e := run.Case(t, "CATALOG-OPENAI-retired-"+string(old.API)+"-"+old.ID)
			_, found := current.Lookup(old.Provider, old.API, old.ID)
			e.Check("retired identity is absent from discovery", !found, "identity=%s/%s", old.API, old.ID)
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

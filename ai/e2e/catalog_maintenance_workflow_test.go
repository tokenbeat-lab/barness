package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

func freezeCatalogAction(t *testing.T, ev *evidence.Case, binary, store string, args ...string) maintenance.Candidate {
	t.Helper()
	command := append([]string{args[0], "--store", store, "--confirm"}, args[1:]...)
	out, exit := invokeCatalogCLI(t, ev, binary, command...)
	var result struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("public action freezes a complete candidate", exit == 0 && result.Candidate.ID != "", "exit=%d output=%s", exit, out) {
		t.FailNow()
	}
	return result.Candidate
}

func checkApplyCatalogAction(t *testing.T, ev *evidence.Case, binary, store string, candidate maintenance.Candidate) maintenance.Revision {
	t.Helper()
	root := catalogSnapshotRoot(t)
	out, exit := invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	_ = json.Unmarshal(out, &checked)
	if !ev.Check("actual fixed public contract qualifies candidate", exit == 0 && checked.Status == "applicable", "exit=%d output=%s", exit, out) {
		t.FailNow()
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", candidate.ID, "--repo", root, "--check", checked.Reference, "--confirm")
	var result struct {
		Revision maintenance.Revision `json:"revision"`
	}
	_ = json.Unmarshal(out, &result)
	if !ev.Check("verified public apply commits candidate", exit == 0 && result.Revision.CandidateID == candidate.ID, "exit=%d output=%s", exit, out) {
		t.FailNow()
	}
	return result.Revision
}

func TestCatalogMaintenanceCLICostUpdate(t *testing.T) {
	ev := run.Case(t, "CATALOG-T03-cli-bulk-price-ownership")
	binary := buildCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	dir := t.TempDir()
	store := filepath.Join(dir, "store")
	source := filepath.Join(dir, "source.json")
	decisions := filepath.Join(dir, "decisions.json")
	raw := []byte(`{"providers":{"openai":{"id":"openai","models":{"gpt-4":{"id":"gpt-4","type":"chat","cost":{"input":1,"output":2,"cache_read":0,"cache_write":0}},"gpt-4.1":{"id":"gpt-4.1","type":"chat","cost":{"input":2,"output":8,"cache_read":0.5,"cache_write":0}}}}},"models":{}}`)
	os.WriteFile(source, raw, 0600)
	ev.Fixture("initial-source.json", raw)
	ownership := []byte(`{"/cost/input":{"ownership":"source","reason":"confirmed offering base USD per million token input"},"/cost/cacheRead":{"ownership":"source","reason":"confirmed base cache read unit"},"/cost/cacheWrite":{"ownership":"source","reason":"confirmed base cache write unit"}}`)
	os.WriteFile(decisions, ownership, 0600)
	ev.Fixture("source-ownership.json", ownership)
	manual := filepath.Join(dir, "manual.json")
	os.WriteFile(manual, []byte(`{"cost":{"output":70}}`), 0600)
	for _, id := range []string{"gpt-4", "gpt-4.1"} {
		args := []string{"revise", "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", id, "--source", source, "--source-time", "2026-11-01T00:00:00Z", "--source-provider", "openai", "--decisions", decisions, "--reason", "explicitly enroll selected price fields"}
		if id == "gpt-4" {
			args = append(args, "--definition", manual)
		}
		candidate := freezeCatalogAction(t, ev, binary, store, args...)
		checkApplyCatalogAction(t, ev, binary, store, candidate)
	}
	updated := []byte(`{"providers":{"openai":{"id":"openai","models":{"gpt-4":{"id":"gpt-4","type":"chat","name":"Must not overwrite","reasoning":true,"limit":{"context":999,"output":9},"cost":{"input":3,"output":4,"cache_write":0}},"gpt-4.1":{"id":"gpt-4.1","type":"chat","cost":{"input":5,"output":99,"cache_write":0}}}}},"models":{}}`)
	os.WriteFile(source, updated, 0600)
	ev.Fixture("updated-source.json", updated)
	candidate := freezeCatalogAction(t, ev, binary, store, "cost-update", "--provider", "openai", "--source", source, "--source-time", "2026-11-01T01:00:00Z", "--source-provider", "openai", "--reason", "refresh source-owned base rates")
	cat, report, err := catalog.Load(candidate.Snapshot)
	if !ev.Check("price candidate loads", err == nil && report.Failed == 0, "%v %+v", err, report) {
		return
	}
	for _, id := range []string{"gpt-4", "gpt-4.1"} {
		m, ok := cat.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, id)
		input, output, cache, context, name := float64(3), float64(70), float64(0), 8192, "GPT-4"
		if id == "gpt-4.1" {
			input, output, cache, context, name = 5, 8, 0.5, 1047576, "GPT-4.1"
		}
		ev.Check("source-only prices update without changing capability or manual values", ok && m.Cost.Input == input && m.Cost.Output == output && m.Cost.CacheRead == cache && m.Cost.CacheWrite == 0 && m.ContextWindow == context && m.Name == name && !m.Reasoning, "%s model=%+v", id, m)
	}
	manualHint, missingHint := false, false
	for _, d := range candidate.Diagnostics {
		manualHint = manualHint || d.Kind == "manual_override"
		missingHint = missingHint || d.Kind == "missing"
	}
	ev.Check("upstream manual differences and missing rates remain visible", manualHint && missingHint, "diagnostics=%+v", candidate.Diagnostics)
	revision := checkApplyCatalogAction(t, ev, binary, store, candidate)
	for _, p := range revision.State.Providers {
		expected := uint64(1)
		if p.ID == ai.ProviderOpenAI {
			expected = 3
		}
		ev.Check("only price-changing Provider advances content version", p.Version == expected, "%s=%d", p.ID, p.Version)
	}
	os.WriteFile(source, []byte(`{"providers":{"openai":{"id":"openai","models":{}}},"models":{}}`), 0600)
	out, exit := invokeCatalogCLI(t, ev, binary, "cost-update", "--store", store, "--provider", "openai", "--source", source, "--source-time", "2026-11-01T02:00:00Z", "--source-provider", "openai", "--reason", "missing is not withdrawal", "--confirm")
	var unchanged struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(out, &unchanged)
	ev.Check("missing offerings only produce review hints", exit == 0 && unchanged.Status == "unchanged", "exit=%d output=%s", exit, out)
}

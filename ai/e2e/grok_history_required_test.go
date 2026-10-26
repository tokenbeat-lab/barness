package e2e

import (
	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"testing"
	"time"
)

func TestGrokHistoryGenericMandatoryProjection(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-GENERIC-MANDATORY-PROJECTION")
	cat := ai.BuiltinCatalog()
	hash, _ := cat.Hash()
	now := time.Now().UTC()
	r := supportmatrix.Report{Schema: supportmatrix.SchemaVersion, Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses", Endpoint: "https://api.openai.com/v1", Model: "gpt-5-mini", SDK: "synthetic projection only", AccountAlias: "synthetic-account", CatalogVersion: cat.Version, CatalogHash: hash, GitCommit: "synthetic-no-live", StartedAt: now, FinishedAt: now, Expected: []string{grokhistory.Scenario}, Scenarios: []supportmatrix.ScenarioResult{{ID: grokhistory.Scenario, Model: "gpt-5-mini", Outcome: supportmatrix.Unsupported, Note: "synthetic protocol omission"}}}
	m := supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{{Combo: r.Combo, Operation: r.Operation, Provider: r.Provider, API: r.API}}}
	_, err := supportmatrix.Merge(m, r)
	ev.Record("synthetic-projection", r)
	ev.Check("mandatory foreign-history scene cannot be excused as optional unsupported", err != nil, "new scene contributed unsupported qualification")
	r.Expected = []string{"image"}
	r.Scenarios[0].ID = "image"
	_, err = supportmatrix.Merge(m, r)
	ev.Check("existing optional unsupported semantics remain", err == nil, "%v", err)
	r.Expected = []string{grokhistory.Scenario}
	r.Scenarios[0].ID = grokhistory.Scenario
	r.Scenarios[0].Outcome = supportmatrix.NotRun
	_, err = supportmatrix.Merge(m, r)
	ev.Check("not-run remains a partial reporting fact", err == nil, "%v", err)
}

func TestGrokHistoryOpenAIFullSelector(t *testing.T) {
	ev := run.Case(t, "GROK-HISTORY-OPENAI-FULL-SELECTOR")
	w, b, model := grokHistoryWorld(t, "openai-responses")
	reply, _ := regionalHostScript(t, b, model)
	enqueue(ev, w, reply)
	scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: t.Name()}
	req, err := grokhistory.Request(scope)
	if err != nil {
		t.Fatal(err)
	}
	o := drain(w.client.Stream(ctxFor(t), scope, ai.Target{BindingID: b.BindingID, ModelID: model}, req, ai.ResponsesOptions{MaxTokens: ai.Value(1024), ReasoningEffort: ai.ThinkingMinimal}))
	o.record(ev)
	ev.Record("requests", w.provider.Requests())
	if !ev.Check("actual full-options destination succeeds once", o.err == nil && len(w.provider.Requests()) == 1, "%v", o.err) {
		return
	}
	raw := w.provider.Requests()[0].Body
	ev.Check("own output selector is retained", string(bodyFields(t, raw)["include"]) == `["reasoning.encrypted_content"]`, "selector missing")
	ev.Check("own output selector is not foreign encrypted history", grokhistory.Verify(b.API, raw) == nil, "%v", grokhistory.Verify(b.API, raw))
	var fields map[string]any
	mustUnmarshal(t, raw, &fields)
	items := fields["input"].([]any)
	items[0].(map[string]any)["encrypted_content"] = "synthetic-unknown-foreign-cipher"
	ev.Check("an actual encrypted field in history remains forbidden", grokhistory.Verify(b.API, mustMarshal(t, fields)) != nil, "foreign field accepted")
}

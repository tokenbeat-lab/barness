package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func genericUnsupportedNativeFixture(t *testing.T, root string, ev *evidence.Case) (string, string, string) {
	t.Helper()
	parent := filepath.Join(root, ".evidence/barness-ai-grok/live/openai-chat")
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	attempt, err := os.MkdirTemp(parent, "TEST_ONLY-unsupported-")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(attempt, "bundles", "synthetic")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	r := genericUnsupportedReport(t)
	git := exec.CommandContext(ctxFor(t), "git", "rev-parse", "HEAD")
	git.Dir, git.Env = root, xaiProcessEnv(t)
	commit, err := git.Output()
	if err != nil {
		t.Fatal(err)
	}
	r.GitCommit = strings.TrimSpace(string(commit))
	r.Budget.InputTokensUsed, r.Budget.OutputTokensUsed = r.Budget.CallsUsed, r.Budget.CallsUsed
	var cases []map[string]any
	for _, s := range r.Scenarios {
		caseDir := filepath.Join(dir, s.CaseID)
		if err := os.Mkdir(caseDir, 0700); err != nil {
			t.Fatal(err)
		}
		assertions := map[string]any{"case_id": s.CaseID, "status": s.Outcome, "assertions": []any{}}
		if s.Outcome == supportmatrix.Unsupported {
			assertions["unsupported"] = s.Note
		} else {
			assertions["assertions"] = []map[string]any{{"pass": true}}
			meta := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: tenantA.tenant, RequestID: s.CaseID, BindingID: r.Combo, AccountScopeID: "live-" + r.Combo, Operation: ai.OperationChat, Resolved: true, ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAICompletions, ModelID: s.Model}, CatalogVersion: r.CatalogVersion, CatalogHash: r.CatalogHash, Attempts: []ai.Attempt{{AttemptID: s.CaseID + "#1", HTTPStatus: 200}}}
			wire := string(mustMarshal(t, map[string]string{"model": s.Model}))
			if s.ID == grokhistory.Scenario {
				wire = genericUnsupportedNativeHistory(t, caseDir, ev)
				meta.NativeStateDowngrades = ai.NativeStateDowngrades{CrossModel: 2}
			}
			result := ai.Result{Message: ai.AssistantMessage{Content: []ai.AssistantContent{ai.Text{Text: "TEST_ONLY synthetic response"}}, API: meta.API, Provider: meta.ProviderID, Model: s.Model, Usage: ai.Usage{Input: 1, Output: 1}, StopReason: ai.StopReasonStop}, Metadata: meta}
			writeReleaseJSON(t, caseDir, "attempt-1-result.json", map[string]any{"result": result, "error": ""})
			writeReleaseJSON(t, caseDir, "attempt-1-exchanges.json", []map[string]any{{"method": "POST", "url": r.Endpoint + "/chat/completions", "requestBody": wire, "status": 200, "truncated": false}})
			writeReleaseJSON(t, caseDir, "observations-attempt-1.json", []ai.Observation{{Kind: ai.ObservationCallFinished, Call: meta}})
		}
		writeReleaseJSON(t, caseDir, "assertions.json", assertions)
		cases = append(cases, map[string]any{"id": s.CaseID, "dir": s.CaseID, "status": s.Outcome, "artifacts": map[string]string{}})
	}
	// Match the native read/verify contract byte for byte without issuing a
	// real approval. Only the releasegate child sees this synthetic Git store.
	gitDir := filepath.Join(attempt, "TEST_ONLY.git")
	init := exec.CommandContext(ctxFor(t), "git", "init", "--bare", "--initial-branch=synthetic", gitDir)
	init.Dir, init.Env = root, xaiProcessEnv(t)
	if raw, err := init.CombinedOutput(); err != nil {
		t.Fatalf("isolated Git namespace: %v %s", err, raw)
	}
	scope, err := supportmatrix.ProcessScope(r.Combo, "live", "formal", strings.Repeat("0", 64))
	if err != nil {
		t.Fatal(err)
	}
	grant := livegrant.Grant{Schema: 1, ID: "TEST_ONLY_" + evidence.Digest([]byte(attempt))[:16], AccountAlias: r.AccountAlias, RegionAlias: "TEST_ONLY", ExpiresAt: r.FinishedAt.Add(time.Hour), Scope: scope}
	rawGrant, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	mustWriteXAINative(t, filepath.Join(dir, "process-grant.json"), rawGrant)
	id := livegrant.Hash([]byte(grant.ID))
	scopeGrant := grant
	scopeGrant.ID = ""
	rawScope, err := json.Marshal(scopeGrant)
	if err != nil {
		t.Fatal(err)
	}
	claim := livegrant.Receipt{Schema: 1, GrantID: grant.ID, GrantSHA256: livegrant.Hash(rawGrant), ScopeSHA256: livegrant.Hash(rawScope), Session: "TEST_ONLY_not_actual_live", StartedAt: r.StartedAt}
	writeReleaseJSON(t, dir, "process-authorization-receipt.json", claim)
	store := filepath.Join(gitDir, "barness-live-process-grants")
	if err := os.Mkdir(store, 0700); err != nil {
		t.Fatal(err)
	}
	// Verify compares exact compact receipt bytes in both authoritative keys.
	for _, name := range []string{"id-" + id + ".json", "scope-" + claim.ScopeSHA256 + ".json"} {
		mustWriteXAINative(t, filepath.Join(store, name), mustMarshal(t, claim))
	}
	writeReleaseJSON(t, dir, "live-report.json", r)
	writeReleaseJSON(t, dir, "audit.json", map[string]any{"findings": []any{}})
	writeReleaseJSON(t, dir, "manifest.json", map[string]any{"package": "-tags live ./ai/live", "git_commit": r.GitCommit, "started_at": r.StartedAt, "finished_at": r.FinishedAt, "versions": map[string]string{"live_combo": r.Combo, "live_sdk": r.SDK, "model_catalog_version": r.CatalogVersion, "model_catalog_hash": r.CatalogHash, "process_build_source": scope.SourceSHA256}, "cases": cases})
	resealXAINativeManifest(t, dir)
	return dir, filepath.Join(attempt, "synthetic-record.json"), gitDir
}

// The history wire must use Chat encoding. It independently evolves from the
// existing Responses-only decoder fixture; neither helper grants live support.
func genericUnsupportedNativeHistory(t *testing.T, dir string, ev *evidence.Case) string {
	t.Helper()
	w, binding, model := grokHistoryWorld(t, "openai-chat")
	reply, _ := regionalHostScript(t, binding, model)
	enqueue(ev, w, reply)
	scope := ai.CallScope{TenantID: tenantA.tenant, RequestID: t.Name()}
	req, err := grokhistory.Request(scope)
	if err != nil {
		t.Fatal(err)
	}
	provenance := grokhistory.Provenance{Before: grokhistory.Envelopes(req)}
	writeReleaseJSON(t, dir, "public-request.json", req)
	got := drain(w.client.StreamSimple(ctxFor(t), scope, ai.Target{BindingID: binding.BindingID, ModelID: model}, req, ai.SimpleOptions{MaxTokens: ai.Value(64)}))
	requests := w.provider.Requests()
	if got.err != nil || got.result.Message.StopReason != ai.StopReasonStop || len(requests) != 1 {
		t.Fatal("controlled Chat history fixture did not complete")
	}
	provenance.After = grokhistory.Envelopes(req)
	writeReleaseJSON(t, dir, "immutable-input.json", req)
	writeReleaseJSON(t, dir, "history-provenance.json", provenance)
	if grokhistory.Verify(binding.API, requests[0].Body) != nil || got.result.Metadata.NativeStateDowngrades != (ai.NativeStateDowngrades{CrossModel: 2}) {
		t.Fatal("controlled Chat history fixture has invalid wire or downgrade")
	}
	return string(requests[0].Body)
}

func genericUnsupportedNativeRecord(t *testing.T, dir, record string) {
	t.Helper()
	genericCaptureDecoderReceipt(t, dir, record)
	attempt := filepath.Dir(record)
	receiptPath := filepath.Join(attempt, "synthetic-receipt.json")
	var receipt release.StageReceipt
	raw, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &receipt)
	route := "openai-chat"
	receipt.Route = &route
	writeReleaseJSON(t, attempt, filepath.Base(receiptPath), receipt)
	var index release.Record
	raw, err = os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	mustUnmarshal(t, raw, &index)
	stage := index.Live.Routes["synthetic-chat"]
	stage.Receipt.SHA256 = evidence.Digest(mustMarshal(t, receipt))
	stage.RequiredCases = nil
	for _, id := range supportmatrix.GenericExpected(route) {
		stage.RequiredCases = append(stage.RequiredCases, "synthetic-"+id)
	}
	index.Live.Routes = map[string]release.StageRecord{route: stage}
	index.Impact.LiveRoutes = []string{route}
	index.Impact.Reason = "TEST_ONLY unsupported qualification replay; isolated authority and empty source cannot establish live support."
	writeReleaseJSON(t, attempt, filepath.Base(record), index)
}

package e2e

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/catalog"
	"github.com/tokenbeat-lab/barness/catalog/maintenance"
)

// The empty final snapshot still has a mandatory native child contract. It
// proves the absence of models through public loading, queries and each
// operation's Client boundary, rather than qualifying a zero-match process.
func catalogEmptyContract(t *testing.T, ev *evidence.Case, raw []byte) {
	cat, report, err := catalog.Load(raw)
	ev.Record("load-report", report)
	if !ev.Check("empty snapshot loads with an honest empty report", err == nil && report.Loaded == 0 && report.Failed == 0 && len(report.Models) == 0, "err=%v report=%+v", err, report) {
		return
	}
	ev.Check("all three effective collections are empty", len(cat.Models) == 0 && len(cat.ImageModels) == 0 && len(cat.ClassifierModels) == 0, "%+v", cat)
	for _, identity := range []catalogContractIdentity{
		{ai.OperationChat, ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4"},
		{ai.OperationImage, ai.ProviderOpenAI, ai.APIOpenAIImages, "gpt-image-1"},
		{ai.OperationClassifier, ai.ProviderOpenAI, ai.APIOpenAIDecisions, "gpt-6-luna"},
	} {
		_, chatFound := cat.Lookup(identity.Provider, identity.API, identity.ID)
		_, imageFound := cat.LookupImage(identity.Provider, identity.API, identity.ID)
		_, classifierFound := cat.LookupClassifier(identity.Provider, identity.API, identity.ID)
		ev.Check("empty queries do not restore builtins", !chatFound && !imageFound && !classifierFound && len(cat.ModelsOf(identity.Provider, identity.API)) == 0 && len(cat.ImageModelsOf(identity.Provider, identity.API)) == 0 && len(cat.ClassifierModelsOf(identity.Provider, identity.API)) == 0, "identity=%+v", identity)
		w, _ := newCatalogContractWorld(t, cat, identity)
		var callErr error
		switch identity.Operation {
		case ai.OperationChat:
			_, callErr = w.client.CompleteSimple(ctxFor(t), textScope("empty-chat"), catalogContractTarget(identity), ai.Request{Messages: []ai.Message{ai.UserText("Empty catalog contract.")}}, ai.SimpleOptions{})
		case ai.OperationImage:
			_, callErr = w.client.GenerateImages(ctxFor(t), textScope("empty-image"), catalogContractTarget(identity), ai.ImagesRequest{Prompt: "Empty catalog contract."}, nil)
		case ai.OperationClassifier:
			req := ai.ClassifierRequest{State: json.RawMessage(`"Empty catalog contract"`), Questions: map[string]ai.ClassifierQuestion{"q": ai.BoolQuestion{Instructions: json.RawMessage(`"Is this true?"`)}}}
			_, callErr = w.client.Classify(ctxFor(t), textScope("empty-classifier"), catalogContractTarget(identity), req, nil)
		}
		ev.Record(string(identity.Operation)+"-error", errString(callErr))
		ev.Record(string(identity.Operation)+"-requests", w.provider.Requests())
		ev.Check("unknown target refused at catalog capability resolution", errors.Is(callErr, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseCapability}), "operation=%s err=%v", identity.Operation, callErr)
		ev.Check("empty catalog sends no provider request", len(w.provider.Requests()) == 0, "operation=%s requests=%d", identity.Operation, len(w.provider.Requests()))
	}
}

// A separately compiled synthetic reviewed default has one active identity.
// All changes to its store still go through the public CLI; no unchecked
// current revision is installed and the checked-in default remains untouched.
func TestCatalogCandidateLastWithdrawalToEmpty(t *testing.T) {
	ev := run.Case(t, "CATALOG-CHECK-last-withdrawal-empty")
	binary := buildSingleModelCatalogCLI(t, ev)
	if binary == "" {
		return
	}
	directory := t.TempDir()
	store := filepath.Join(directory, "store")
	out, exit := invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	initial, report, err := catalog.Load(out)
	if !ev.Check("fresh default has exactly one active model", exit == 0 && err == nil && report.Failed == 0 && len(initial.Models) == 1 && len(initial.ImageModels) == 0 && len(initial.ClassifierModels) == 0, "exit=%d err=%v report=%+v", exit, err, report) {
		return
	}
	out, exit = invokeCatalogCLI(t, ev, binary, "withdraw", "--store", store, "--operation", "chat", "--provider", "openai", "--api", "openai-responses", "--model", "gpt-4", "--reason", "withdraw the last active model", "--confirm")
	var frozen struct {
		Candidate maintenance.Candidate `json:"candidate"`
	}
	mustUnmarshal(t, out, &frozen)
	if !ev.Check("last model withdrawal freezes a candidate", exit == 0 && frozen.Candidate.ID != "", "exit=%d %s", exit, out) {
		return
	}
	empty, report, err := catalog.Load(frozen.Candidate.Snapshot)
	ev.Check("empty candidate is valid without builtin fallback", err == nil && report.Failed == 0 && len(empty.Models) == 0 && len(empty.ImageModels) == 0 && len(empty.ClassifierModels) == 0, "err=%v report=%+v", err, report)
	root := catalogSnapshotRoot(t)
	out, exit = invokeCatalogCLI(t, ev, binary, "check", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--evidence-feature", "barness-ai-dynamic-catalog", "--confirm")
	var checked maintenance.CheckResult
	_ = json.Unmarshal(out, &checked)
	if !ev.Check("empty candidate passes a real global contract", exit == 0 && checked.Status == "applicable" && slices.Equal(checked.RequiredCases, []string{"CATALOG-CANDIDATE-PLAN", "CC-empty-catalog"}), "exit=%d %s", exit, out) {
		return
	}
	catalogEmptyEvidenceRefusals(t, ev, checked)
	_, err = os.Stat(filepath.Join(store, "current.json"))
	ev.Check("check does not apply last withdrawal", os.IsNotExist(err), "%v", err)
	output := filepath.Join(directory, "empty-runtime.json")
	out, exit = invokeCatalogCLI(t, ev, binary, "apply", "--store", store, "--id", frozen.Candidate.ID, "--repo", root, "--check", checked.Reference, "--out", output, "--confirm")
	if !ev.Check("checked last withdrawal applies", exit == 0, "exit=%d %s", exit, out) {
		return
	}
	raw, err := os.ReadFile(output)
	ev.Check("empty runtime is exactly the tested snapshot", err == nil && bytes.Equal(raw, frozen.Candidate.Snapshot), "%v", err)
	loaded, report, err := catalog.Load(raw)
	_, found := loaded.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4")
	ev.Check("applied runtime stays empty and withdrawn model stays absent", err == nil && report.Failed == 0 && len(loaded.Models) == 0 && len(loaded.ImageModels) == 0 && len(loaded.ClassifierModels) == 0 && !found, "err=%v report=%+v", err, report)
	out, exit = invokeCatalogCLI(t, ev, binary, "export", "--store", store)
	ev.Check("current export retains exact empty runtime", exit == 0 && bytes.Equal(out, raw), "exit=%d %s", exit, out)
	out, exit = invokeCatalogCLI(t, ev, binary, "history", "--store", store)
	var history struct {
		History []maintenance.Revision `json:"history"`
	}
	mustUnmarshal(t, out, &history)
	ev.Check("empty application appends rather than erases history", exit == 0 && len(history.History) == 2, "exit=%d %s", exit, out)
}

func catalogEmptyEvidenceRefusals(t *testing.T, ev *evidence.Case, check maintenance.CheckResult) {
	t.Helper()
	root := catalogSnapshotRoot(t)
	original := filepath.Dir(filepath.Join(root, check.Receipt))
	for _, mutation := range []string{"intact", "parent-only", "shortened-plan", "missing-empty-scene"} {
		directory, err := os.MkdirTemp(filepath.Join(root, ".evidence", "barness-ai-dynamic-catalog"), "consumer-empty-fixture-"+mutation+"-")
		if err != nil {
			t.Fatal(err)
		}
		copied := filepath.Join(directory, "process")
		catalogCopyTree(t, original, copied)
		catalogWriteFixtureJSON(t, filepath.Join(directory, "fixture-origin.json"), map[string]any{"purpose": "derived empty-catalog consumer fixture; not an actual stage", "mutation": mutation, "sourceReceipt": check.Receipt, "sourceSHA256": check.ReceiptSHA256})
		var receipt map[string]any
		catalogReadFixtureJSON(t, filepath.Join(copied, "receipt.json"), &receipt)
		bundleRef := receipt["bundles"].([]any)[0].(map[string]any)
		bundle := filepath.Join(copied, bundleRef["path"].(string))
		var manifest map[string]any
		catalogReadFixtureJSON(t, filepath.Join(bundle, "manifest.json"), &manifest)
		switch mutation {
		case "parent-only":
			catalogWriteFixtureJSON(t, filepath.Join(copied, "stdout.jsonl"), map[string]any{"Action": "pass", "Package": "github.com/tokenbeat-lab/barness/ai/e2e", "Test": "TestCatalogCandidateContracts"})
			receipt["files"].(map[string]any)["stdout.jsonl"] = catalogFileSHA(t, filepath.Join(copied, "stdout.jsonl"))
		case "shortened-plan":
			for _, item := range manifest["cases"].([]any) {
				c := item.(map[string]any)
				if c["id"] != "CATALOG-CANDIDATE-PLAN" {
					continue
				}
				path := filepath.Join(bundle, c["dir"].(string), "plan.json")
				var plan map[string]any
				catalogReadFixtureJSON(t, path, &plan)
				plan["cases"] = []any{}
				catalogWriteFixtureJSON(t, path, plan)
				c["artifacts"].(map[string]any)["plan.json"] = catalogFileSHA(t, path)
			}
		case "missing-empty-scene":
			var remaining []any
			for _, item := range manifest["cases"].([]any) {
				if item.(map[string]any)["id"] != "CC-empty-catalog" {
					remaining = append(remaining, item)
				}
			}
			manifest["cases"] = remaining
		}
		catalogWriteFixtureJSON(t, filepath.Join(bundle, "manifest.json"), manifest)
		bundleRef["manifestSHA256"] = catalogFileSHA(t, filepath.Join(bundle, "manifest.json"))
		catalogWriteFixtureJSON(t, filepath.Join(copied, "receipt.json"), receipt)
		relative, _ := filepath.Rel(root, filepath.Join(copied, "receipt.json"))
		cmd := exec.CommandContext(ctxFor(t), "python3", "tools/verification/catalog_check.py", "inspect", "--prepared", check.Prepared, "--receipt", filepath.ToSlash(relative))
		cmd.Dir = root
		out, err := cmd.CombinedOutput()
		exit := commandExit(err)
		ev.Record("consumer-"+mutation, map[string]any{"argv": cmd.Args, "exitCode": exit, "output": string(out)})
		ev.Check("empty contract evidence is exact "+mutation, (mutation == "intact") == (exit == 0), "exit=%d %s", exit, out)
	}
}

func buildSingleModelCatalogCLI(t *testing.T, ev *evidence.Case) string {
	t.Helper()
	fresh := copyCatalogBootstrapSource(t)
	state, err := maintenance.BuiltinState()
	if err != nil {
		t.Fatal(err)
	}
	// This synthetic one-model seed intentionally starts from historical entries.
	// The adopted baseline binds the real default and cannot describe this fixture.
	baseline := filepath.Join(fresh, "catalog", "data", "baseline")
	if err := os.RemoveAll(baseline); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(baseline, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(baseline, "README.md"), []byte("Synthetic historical-only seed.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for pi := range state.Providers {
		p := &state.Providers[pi]
		historical := p.Entries[:0]
		for _, entry := range p.Entries {
			if entry.Source != nil && entry.Source.Kind == "historical" {
				historical = append(historical, entry)
			}
		}
		p.Entries = historical
		for ei := range p.Entries {
			e := &p.Entries[ei]
			var identity struct{ Provider, API, ID string }
			mustUnmarshal(t, e.Definition, &identity)
			if e.Operation != ai.OperationChat || identity.Provider != "openai" || identity.API != "openai-responses" || identity.ID != "gpt-4" {
				e.Lifecycle = maintenance.Withdrawn
			} else {
				e.Lifecycle = maintenance.Active
			}
		}
		raw := mustMarshal(t, p)
		if err := os.WriteFile(filepath.Join(fresh, "catalog", "data", "providers", string(p.ID)+".json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := maintenance.Generate(state)
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("synthetic-reviewed-seed", state)
	ev.Fixture("synthetic-reviewed-default.json", snapshot)
	if err := os.WriteFile(filepath.Join(fresh, "ai", "catalog_builtin.json"), snapshot, 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "catalogctl-single-model")
	cmd := exec.CommandContext(ctxFor(t), "go", catalogBuildArgs(binary, "./cmd/catalogctl")...)
	cmd.Dir = fresh
	out, err := cmd.CombinedOutput()
	ev.Record("fresh-build", map[string]any{"argv": cmd.Args, "exitCode": commandExit(err), "output": string(out)})
	if !ev.Check("single-model reviewed source compiles", err == nil, "%v %s", err, out) {
		return ""
	}
	raw, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	ev.Record("binary", map[string]any{"sha256": hex.EncodeToString(sum[:]), "bytes": len(raw)})
	return binary
}

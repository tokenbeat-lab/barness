package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure inventory: nonempty domain text/thinking/tool content cannot be
// unmarshaled into AssistantContent interfaces. Reading the capture must still
// reject attribution, terminal, usage and inventory tampering. These synthetic
// bundles deliberately name an unknown route: consuming their captures never
// produces real live qualification, a grant, a claim or a vendor request.
func TestGenericChatCaptureDecoder(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build public consumer: %v %s", err, out)
	}
	for _, mutation := range []string{"nonempty-content", "attribution", "terminal", "negative-usage", "usage-budget", "inventory"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "GROK-GENERIC-CAPTURE-"+mutation)
			dir, record := genericCaptureDecoderFixture(t, root)
			result := filepath.Join(dir, "synthetic-text", "attempt-1-result.json")
			switch mutation {
			case "attribution":
				editXAINativeJSON(t, dir, "synthetic-text/attempt-1-result.json", func(v map[string]any) {
					v["result"].(map[string]any)["metadata"].(map[string]any)["bindingId"] = "synthetic-other-binding"
				})
			case "terminal":
				editXAINativeArray(t, dir, "synthetic-text/observations-attempt-1.json", func(v []any) {
					v[len(v)-1].(map[string]any)["kind"] = "attempt_finished"
				})
			case "negative-usage":
				editXAINativeJSON(t, dir, "synthetic-text/attempt-1-result.json", func(v map[string]any) {
					v["result"].(map[string]any)["message"].(map[string]any)["usage"].(map[string]any)["input"] = -1
				})
			case "usage-budget":
				editXAINativeJSON(t, dir, "live-report.json", func(v map[string]any) { v["budget"].(map[string]any)["inputTokensUsed"] = 4 })
			case "inventory":
				raw, readErr := os.ReadFile(result)
				if readErr != nil {
					t.Fatal(readErr)
				}
				mustWriteXAINative(t, result, append(raw, ' '))
			}
			if mutation != "inventory" {
				resealXAINativeManifest(t, dir)
			}
			audited := release.AuditBundle(dir, nil)
			ev.Record("synthetic-audit", map[string]any{"kind": "synthetic-capture-regression", "notActualLive": true, "error": audited.Err, "findings": audited.Findings, "vendorCalls": 0})
			accepted := mutation == "nonempty-content"
			ev.Check("capture consumption accepts only intact nonempty result", (audited.Err == "" && len(audited.Findings) == 0) == accepted, "accepted=%t error=%s", accepted, audited.Err)
			// Rebind only this synthetic receipt's inventory after its mutation.
			genericCaptureDecoderReceipt(t, dir, record)
			out := filepath.Join(filepath.Dir(record), "consumer")
			cmd := exec.CommandContext(ctxFor(t), binary, "-record", record, "-out", out)
			cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
			raw, processErr := cmd.CombinedOutput()
			exit := -1
			if cmd.ProcessState != nil {
				exit = cmd.ProcessState.ExitCode()
			}
			ev.Record("public-consumer-process", map[string]any{"argv": cmd.Args, "exitCode": exit, "output": string(raw), "notActualLive": true})
			data, readErr := os.ReadFile(filepath.Join(out, "release-report.json"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			var report release.RecordReport
			mustUnmarshal(t, data, &report)
			ev.Fixture("public-consumer-report.json", data)
			problem := ""
			for _, check := range report.Checks {
				if check.ID == "live" {
					problem = check.Problem
				}
			}
			ev.Check("synthetic incomplete qualification always remains false", processErr != nil && exit == 1 && !report.Pass, "exit=%d pass=%t", exit, report.Pass)
			if accepted {
				ev.Check("public CLI reaches qualification after valid capture audit", strings.HasPrefix(problem, "native live qualification rejected:"), "problem=%s", problem)
			} else {
				ev.Check("public CLI refuses tampered capture at audit", problem == "native evidence audit failed", "problem=%s", problem)
			}
		})
	}
}

func genericCaptureDecoderFixture(t *testing.T, root string) (string, string) {
	t.Helper()
	parent := filepath.Join(root, ".evidence/barness-ai-grok/live/synthetic-chat")
	if err := os.MkdirAll(parent, 0700); err != nil {
		t.Fatal(err)
	}
	attempt, err := os.MkdirTemp(parent, "capture-decoder-")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(attempt, "bundles", "synthetic")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	git := exec.Command("git", "rev-parse", "HEAD")
	git.Dir, git.Env = root, xaiProcessEnv(t)
	commitRaw, err := git.Output()
	if err != nil {
		t.Fatal(err)
	}
	commit := strings.TrimSpace(string(commitRaw))
	cat := ai.BuiltinCatalog()
	hash, err := cat.Hash()
	if err != nil {
		t.Fatal(err)
	}
	r := supportmatrix.Report{Schema: 3, Combo: "synthetic-chat", Operation: "chat", Provider: "openai", API: "openai-responses", Endpoint: "https://api.openai.com/v1", Model: "gpt-5-mini", SDK: "synthetic capture only", AccountAlias: "synthetic", GitCommit: commit, CatalogVersion: cat.Version, CatalogHash: hash, StartedAt: now, FinishedAt: now.Add(time.Second), Budget: supportmatrix.Budget{MaxCalls: 32, MaxRetries: 1, MaxOutputTokens: 4096, CallsUsed: 3, HTTPAttempts: 3, InputTokensUsed: 3, OutputTokensUsed: 3}}
	var cases []map[string]any
	blocks := []ai.AssistantContent{ai.Text{Text: "synthetic text"}, ai.Thinking{Thinking: "synthetic thinking"}, ai.ToolCall{ID: "synthetic-call", Name: "synthetic_tool", Arguments: ai.UncheckedArguments(`{"value":1}`)}}
	for i, kind := range []string{"text", "thinking", "tool"} {
		id := "synthetic-" + kind
		caseDir := filepath.Join(dir, id)
		if err := os.Mkdir(caseDir, 0700); err != nil {
			t.Fatal(err)
		}
		meta := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: "synthetic-tenant", RequestID: id, BindingID: r.Combo, Operation: ai.OperationChat, Resolved: true, ProviderID: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ModelID: r.Model}, CatalogVersion: r.CatalogVersion, CatalogHash: hash, Attempts: []ai.Attempt{{AttemptID: id + "#1", HTTPStatus: 200}}}
		result := ai.Result{Message: ai.AssistantMessage{Content: []ai.AssistantContent{blocks[i]}, API: ai.APIOpenAIResponses, Provider: ai.ProviderOpenAI, Model: r.Model, Usage: ai.Usage{Input: 1, Output: 1}, StopReason: ai.StopReasonStop}, Metadata: meta}
		writeReleaseJSON(t, caseDir, "attempt-1-result.json", map[string]any{"result": result, "error": ""})
		writeReleaseJSON(t, caseDir, "attempt-1-exchanges.json", []map[string]any{{"method": "POST", "url": r.Endpoint + "/responses", "requestBody": `{"model":"gpt-5-mini"}`, "status": 200, "truncated": false}})
		writeReleaseJSON(t, caseDir, "observations-attempt-1.json", []ai.Observation{{Kind: ai.ObservationCallFinished, Call: meta}})
		writeReleaseJSON(t, caseDir, "assertions.json", map[string]any{"case_id": id, "status": "PASS", "assertions": []map[string]any{{"pass": true}}})
		r.Expected = append(r.Expected, kind)
		r.Scenarios = append(r.Scenarios, supportmatrix.ScenarioResult{ID: kind, CaseID: id, Model: r.Model, Outcome: supportmatrix.Pass, Calls: 1, Attempts: 1})
		cases = append(cases, map[string]any{"id": id, "dir": id, "status": "PASS", "artifacts": map[string]string{}})
	}
	writeReleaseJSON(t, dir, "live-report.json", r)
	writeReleaseJSON(t, dir, "audit.json", map[string]any{"findings": []any{}})
	writeReleaseJSON(t, dir, "manifest.json", map[string]any{"package": "-tags live ./ai/live", "git_commit": commit, "started_at": now, "finished_at": now.Add(2 * time.Second), "versions": map[string]string{"live_combo": r.Combo, "live_sdk": r.SDK, "model_catalog_version": r.CatalogVersion, "model_catalog_hash": hash}, "cases": cases})
	resealXAINativeManifest(t, dir)
	recordPath := filepath.Join(attempt, "synthetic-record.json")
	return dir, recordPath
}

func genericCaptureDecoderReceipt(t *testing.T, dir, recordPath string) {
	t.Helper()
	attempt := filepath.Dir(recordPath)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	rel := func(path string) string {
		p, e := filepath.Rel(root, path)
		if e != nil {
			t.Fatal(e)
		}
		return filepath.ToSlash(p)
	}
	ref := func(path string) release.FileRef {
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		return release.FileRef{Path: rel(path), SHA256: evidence.Digest(raw)}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "live-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var live supportmatrix.Report
	mustUnmarshal(t, raw, &live)
	toolchain := map[string]string{"goEnv": "off", "goFlags": "", "executableSHA256": strings.Repeat("0", 64), "collectorPython": "synthetic-only"}
	argv := []string{"synthetic-only-not-an-execution"}
	writeReleaseJSON(t, attempt, "command.json", map[string]any{"schema": 1, "stage": "live", "argv": argv, "toolchain": toolchain})
	for name, value := range map[string]string{"stdout.jsonl": "", "stderr.log": "", "exit-code": "0\n", "synthetic-state.md": "Synthetic capture decoder regression; never live qualification.\n"} {
		mustWriteXAINative(t, filepath.Join(attempt, name), []byte(value))
	}
	files := map[string]string{}
	for _, name := range []string{"command.json", "stdout.jsonl", "stderr.log", "exit-code"} {
		files[name] = ref(filepath.Join(attempt, name)).SHA256
	}
	source := release.ExecutionSource{Commit: live.GitCommit, Files: map[string]string{}}
	bundle := release.NativeBundleRef{Path: "bundles/synthetic", Package: "-tags live ./ai/live", ManifestSHA256: ref(filepath.Join(dir, "manifest.json")).SHA256, AuditSHA256: ref(filepath.Join(dir, "audit.json")).SHA256}
	route := "synthetic-chat"
	receipt := release.StageReceipt{Schema: 1, Feature: "barness-ai-grok", Stage: "live", Route: &route, Status: "PASS", ExitCode: new(int), Source: source, SelectedInputs: map[string]string{}, StartedAt: live.StartedAt.Format(time.RFC3339Nano), FinishedAt: live.FinishedAt.Format(time.RFC3339Nano), Argv: argv, Bundles: []release.NativeBundleRef{bundle}, CommandSHA256: files["command.json"], Files: files, Toolchain: toolchain}
	writeReleaseJSON(t, attempt, "synthetic-receipt.json", receipt)
	pressure := true
	record := release.Record{Schema: 1, Feature: "barness-ai-grok", CurrentState: ref(filepath.Join(attempt, "synthetic-state.md")), Source: source, SelectedInputs: map[string]string{}, Impact: release.Impact{LiveRoutes: []string{route}, PressureRequired: &pressure, Reason: "Synthetic capture decoding only; empty source is deliberately ineligible."}, Live: release.StageRecord{Status: "PASS", Routes: map[string]release.StageRecord{route: {Status: "PASS", Receipt: ref(filepath.Join(attempt, "synthetic-receipt.json")), RequiredCases: []string{"synthetic-text"}, Selection: &release.Selection{By: "human", Argv: argv}}}}}
	writeReleaseJSON(t, attempt, filepath.Base(recordPath), record)
}

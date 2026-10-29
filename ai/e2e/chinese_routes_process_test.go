package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func evaluationCommand(t *testing.T, args ...string) ([]byte, error) {
	t.Helper()
	cmd := exec.CommandContext(ctxFor(t), "go", append([]string{"run", "./ai/examples/chineseeval/cmd/chineseeval"}, args...)...)
	cmd.Dir = "../.."
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	return cmd.CombinedOutput()
}

func evaluationSources(t *testing.T, route string) (string, string) {
	t.Helper()
	dir := t.TempDir()
	catalog := ai.BuiltinCatalog()
	catalog.Version, catalog.Models, catalog.ImageModels = "host-zh-eval-v1", nil, nil
	hash, err := catalog.Hash()
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(catalog)
	catalogPath := filepath.Join(dir, "catalog.json")
	if err := os.WriteFile(catalogPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	raw, err = os.ReadFile("../examples/chineseeval/testdata/jev-1.13.0-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]any
	json.Unmarshal(raw, &config)
	provider, api, model := "typesafe", "typesafe-system-one", "jev-1.13.0"
	if route == "luna" {
		provider, api, model = "openai", "openai-decisions", "gpt-6-luna"
	}
	config["version"], config["model"] = "zh-eval-"+route+"-host-v1", model
	config["catalogVersion"], config["catalogHash"] = catalog.Version, hash
	config["route"] = map[string]any{"operation": "classifier", "provider": provider, "api": api, "model": model}
	config["timeoutSeconds"] = 600
	raw, _ = json.Marshal(config)
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return configPath, catalogPath
}

func evaluationBundle(t *testing.T, out string, ev *evidence.Case) (string, map[string]any) {
	t.Helper()
	dirs, err := os.ReadDir(out)
	if err != nil || len(dirs) != 1 {
		t.Fatalf("expected one bundle: %v", err)
	}
	dir := filepath.Join(out, dirs[0].Name())
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() {
			raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			ev.Fixture("bundle-"+e.Name(), raw)
		}
	}
	raw, err := os.ReadFile(filepath.Join(dir, "evaluation-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report map[string]any
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	return dir, report
}

// CLI/process and its serialized artifacts are the agreed OD55 seam. Expected
// counts come from the fixed 24-sample task, not from metric implementation.
func TestChineseEvaluationRoutesProcess(t *testing.T) {
	for _, tc := range []struct {
		route, fixture, status string
		scored, refused, calls int
	}{
		{"jev", "normal", "COMPLETE", 24, 0, 24}, {"jev", "dual-request-id", "COMPLETE", 24, 0, 24}, {"luna", "mixed", "COMPLETE", 12, 12, 24}, {"luna", "response-alias", "COMPLETE", 24, 0, 24},
		{"luna", "all-refusal", "COMPLETE", 0, 24, 24},
		{"luna", "all-refusal-usage-absent", "FAIL", 0, 24, 24},
		{"luna", "all-refusal-usage-partial", "FAIL", 0, 24, 24}, {"luna", "usage-absent", "FAIL", 24, 0, 24},
		{"luna", "usage-partial", "FAIL", 24, 0, 24}, {"luna", "cancel", "FAIL", 0, 0, 1}, {"luna", "pre-attempt-cancel", "FAIL", 0, 0, 1}, {"luna", "http-failure", "FAIL", 0, 0, 24},
		{"luna", "budget", "FAIL", 1, 0, 1},
	} {
		t.Run(tc.route+"-"+tc.fixture, func(t *testing.T) {
			ev := run.Case(t, "OD55-process-"+tc.route+"-"+tc.fixture)
			cfg, cat := evaluationSources(t, tc.route)
			if tc.fixture == "budget" {
				raw, _ := os.ReadFile(cfg)
				var c map[string]any
				json.Unmarshal(raw, &c)
				c["maxCalls"], c["maxQuestions"], c["maxAttempts"] = 1, 1, 1
				raw, _ = json.Marshal(c)
				os.WriteFile(cfg, raw, 0600)
			}
			expectedAttempts := tc.calls
			if tc.fixture == "pre-attempt-cancel" {
				expectedAttempts = 0
			}
			out := t.TempDir()
			output, err := evaluationCommand(t, "-route", tc.route, "-config", cfg, "-catalog", cat, "-fixture", tc.fixture, "-out", out)
			ev.Record("process", map[string]any{"failed": err != nil, "has_status": strings.Contains(string(output), tc.status)})
			if !ev.Check("expected process status", (err == nil) == (tc.status == "COMPLETE"), "exit=%v", err) && !strings.Contains(string(output), "stage=sample_results") {
				return
			}
			dir, r := evaluationBundle(t, out, ev)
			capturesRaw, captureErr := os.ReadFile(filepath.Join(dir, "captures.json"))
			var captures []map[string]any
			json.Unmarshal(capturesRaw, &captures)
			ev.Check("actual bounded transport captures", captureErr == nil && len(captures) == expectedAttempts, "missing actual captures")
			m := r["metrics"].(map[string]any)
			b := r["budget"].(map[string]any)
			ev.Check("task and evidence identity", r["schema"] == float64(2) && r["kind"] == "fixture" && r["status"] == tc.status, "unexpected identity")
			refused, _ := m["refused"].(float64)
			ev.Check("full denominator and refusal partition", m["total"] == float64(24) && m["scored"] == float64(tc.scored) && refused == float64(tc.refused) && m["scored"].(float64)+refused+m["failed"].(float64)+m["unexecuted"].(float64) == 24, "invalid partition")
			ev.Check("serial zero retry budget", b["calls"] == float64(tc.calls) && b["attempts"] == float64(expectedAttempts) && b["retries"] == float64(0), "invalid accounting")
			if tc.fixture == "all-refusal" {
				ev.Check("all-refusal null conditional metrics", m["accuracy"] == float64(0) && m["scoredAccuracy"] == nil && m["confidenceECE"] == nil && m["topProbabilityECE"] == nil && m["brier"] == nil && m["refusalRate"] == float64(1) && b["knownCostUSD"].(float64) > 0, "invalid refusal metrics")
			}
			if tc.fixture == "usage-absent" || tc.fixture == "all-refusal-usage-absent" {
				ev.Check("unknown cost preserved", b["unreported"] == float64(24), "missing reporting")
			}
			if tc.fixture == "usage-partial" || tc.fixture == "all-refusal-usage-partial" {
				ev.Check("partial usage preserved", b["partial"] == float64(24) && b["inputTokens"] == float64(2400), "missing reporting")
			}
			if tc.fixture == "cancel" || tc.fixture == "pre-attempt-cancel" {
				ev.Check("cancel sends no later requests", m["unexecuted"] == float64(23), "later calls sent")
			}
			_, verifyErr := evaluationCommand(t, "-verify", dir)
			ev.Check("recorded snapshot verifies independently", verifyErr == nil, "verify=%v", verifyErr)
		})
	}
}

func TestChineseEvaluationRoutesInputProcess(t *testing.T) {
	for _, name := range []string{"route", "provider", "api", "operation", "model", "hash", "budget", "old-default"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "OD55-input-process-"+name)
			out := t.TempDir()
			if name == "old-default" {
				_, err := evaluationCommand(t, "-out", out)
				_, r := evaluationBundle(t, out, ev)
				c := r["config"].(map[string]any)
				ev.Check("Jev default and schema preserved", err == nil && r["schema"] == float64(1) && r["status"] == "NOT_RUN" && c["model"] == "jev-1.13.0", "default drift")
				return
			}
			cfg, cat := evaluationSources(t, "luna")
			raw, _ := os.ReadFile(cfg)
			var c map[string]any
			json.Unmarshal(raw, &c)
			route := c["route"].(map[string]any)
			switch name {
			case "route":
				route["provider"] = "typesafe"
			case "provider":
				route["provider"] = "unknown"
			case "api":
				route["api"] = "openai-responses"
			case "operation":
				route["operation"] = "chat"
			case "model":
				route["model"] = "jev-1.13.0"
			case "hash":
				c["catalogHash"] = "sha256:changed"
			case "budget":
				c["maxAttempts"] = 25
			}
			raw, _ = json.Marshal(c)
			os.WriteFile(cfg, raw, 0600)
			output, err := evaluationCommand(t, "-route", "luna", "-config", cfg, "-catalog", cat, "-out", out)
			dirs, _ := os.ReadDir(out)
			ev.Record("process", map[string]any{"rejected": err != nil, "task_pins": strings.Contains(string(output), "stage=task_pins")})
			ev.Check("reject before key and bundle creation", err != nil && strings.Contains(string(output), "stage=task_pins") && len(dirs) == 0, "wrong guard stage")
		})
	}
}

func TestChineseEvaluationHistoricalProcess(t *testing.T) {
	ev := run.Case(t, "OD55-historical-recorded-snapshot")
	output, err := evaluationCommand(t, "-verify", ".scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live")
	ev.Record("process", map[string]any{"verified": err == nil, "historical": strings.Contains(string(output), "recorded snapshot")})
	ev.Check("old Jev bundle uses recorded pins", err == nil, "history verify=%v", err)
	// Current plan must still reject old evidence for current release applicability.
	raw, _ := os.ReadFile("../../.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live/evaluation-report.json")
	var r chineseeval.Report
	json.Unmarshal(raw, &r)
	ev.Check("historical verification grants no current applicability", chinesePlan(t).Verify(r) != nil, "old hash accepted as current")
}

func TestChineseEvaluationRoutesVerifierProcess(t *testing.T) {
	for _, name := range []string{"route", "hash", "budget", "response-alias", "identity", "missing-observer", "null-metric", "capture-answer", "capture-question", "coherent-live", "missing-captures"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "OD55-verifier-process-"+name)
			cfg, cat := evaluationSources(t, "luna")
			out := t.TempDir()
			output, err := evaluationCommand(t, "-route", "luna", "-config", cfg, "-catalog", cat, "-fixture", "all-refusal", "-out", out)
			if !ev.Check("fixture executes", err == nil, "process=%v", err) {
				ev.Record("process", map[string]any{"fixture_failed": true, "has_stage": strings.Contains(string(output), "stage=")})
				return
			}
			dir, r := evaluationBundle(t, out, ev)
			switch name {
			case "route":
				r["route"].(map[string]any)["api"] = "openai-responses"
			case "hash":
				r["codeHash"] = "sha256:changed"
			case "budget":
				r["budget"].(map[string]any)["calls"] = 25
			case "response-alias": // Aliases are recorded separately and never rewrite the authorized ModelID.
				r["samples"].([]any)[0].(map[string]any)["metadata"].(map[string]any)["modelId"] = "gpt-6-luna-alias"
			case "identity":
				r["kind"] = "live"
			case "coherent-live":
				r["kind"] = "live"
				os.Remove(filepath.Join(dir, "fixture-captures.json"))
				path := filepath.Join(dir, "manifest.json")
				raw, _ := os.ReadFile(path)
				var m map[string]any
				json.Unmarshal(raw, &m)
				m["versions"].(map[string]any)["evidence_kind"] = "live"
				raw, _ = json.Marshal(m)
				os.WriteFile(path, raw, 0600)
			case "missing-captures":
				os.Remove(filepath.Join(dir, "captures.json"))
			case "missing-observer":
				os.Remove(filepath.Join(dir, "observations-evaluation.json"))
			case "null-metric":
				r["metrics"].(map[string]any)["confidenceECE"] = 0
			case "capture-answer", "capture-question":
				path := filepath.Join(dir, "fixture-captures.json")
				raw, _ := os.ReadFile(path)
				var captures []map[string]any
				json.Unmarshal(raw, &captures)
				if name == "capture-answer" {
					captures[0]["response"].(map[string]any)["answers"] = []any{map[string]any{"name": "intent", "type": "bool", "probability": 1}}
				} else {
					captures[0]["request"].(map[string]any)["questions"] = []any{}
				}
				raw, _ = json.Marshal(captures)
				os.WriteFile(path, raw, 0600)
			}
			raw, _ := json.Marshal(r)
			os.WriteFile(filepath.Join(dir, "evaluation-report.json"), raw, 0600)
			// Rehash the edited file, so this tests semantic import validation too.
			invPath := filepath.Join(dir, "files.json")
			invRaw, _ := os.ReadFile(invPath)
			var inv map[string]string
			if json.Unmarshal(invRaw, &inv) == nil {
				inv["evaluation-report.json"] = chineseeval.Hash(raw)
				captureRaw, _ := os.ReadFile(filepath.Join(dir, "fixture-captures.json"))
				if len(captureRaw) > 0 {
					inv["fixture-captures.json"] = chineseeval.Hash(captureRaw)
				}
				if name == "coherent-live" {
					delete(inv, "fixture-captures.json")
					manifestRaw, _ := os.ReadFile(filepath.Join(dir, "manifest.json"))
					inv["manifest.json"] = chineseeval.Hash(manifestRaw)
				}
				if name == "missing-captures" {
					delete(inv, "captures.json")
				}
				invRaw, _ = json.Marshal(inv)
				os.WriteFile(invPath, invRaw, 0600)
			}
			_, err = evaluationCommand(t, "-verify", dir)
			ev.Check("tampered artifact rejected", err != nil, "accepted %s", name)
		})
	}
}

func TestChineseEvaluationDecisionsUsageImport(t *testing.T) {
	ev := run.Case(t, "OD55-usage-import-boundary")
	cfg, cat := evaluationSources(t, "luna")
	out := t.TempDir()
	_, err := evaluationCommand(t, "-route", "luna", "-config", cfg, "-catalog", cat, "-fixture", "normal", "-out", out)
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := evaluationBundle(t, out, ev)
	data, _ := os.ReadFile(filepath.Join(dir, "dataset.json"))
	config, _ := os.ReadFile(filepath.Join(dir, "config.json"))
	snapshot, _ := os.ReadFile(filepath.Join(dir, "catalog.json"))
	p, err := chineseeval.LoadRecorded(data, config, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "evaluation-report.json"))
	report, err := p.ReadReport(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"missing-reasoning", "null-reasoning", "negative-reasoning", "excess-reasoning", "chat-cache"} {
		var changed chineseeval.Report
		json.Unmarshal(raw, &changed)
		usage := &changed.Samples[0].Metadata.Attempts[0].Usage
		switch name {
		case "missing-reasoning":
			usage.Reasoning = ai.Nullable[int64]{}
		case "null-reasoning":
			usage.Reasoning = ai.Null[int64]()
		case "negative-reasoning":
			usage.Reasoning = ai.Value[int64](-1)
		case "excess-reasoning":
			usage.Reasoning = ai.Value[int64](11)
		case "chat-cache":
			usage.CacheWrite1h = ai.Value[int64](0)
		}
		corrupted, _ := json.Marshal(changed)
		_, err := p.ReadReport(corrupted)
		ev.Check("artifact import rejects "+name, err != nil, "accepted %s", name)
	}
	ev.Check("complete explicit zero reasoning remains valid", p.Verify(report) == nil, "valid report rejected")
}

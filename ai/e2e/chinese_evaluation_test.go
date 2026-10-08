package e2e

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/chineseeval"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func chinesePlan(t *testing.T) *chineseeval.Plan {
	t.Helper()
	data, err := os.ReadFile("../examples/chineseeval/testdata/zh-support-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile("../examples/chineseeval/testdata/jev-1.13.0-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := chineseeval.Load(data, config, ai.BuiltinCatalog())
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func chineseWorld(t *testing.T) *world {
	t.Helper()
	w := newWorldWith(t, func(c *ai.Config) {
		c.Policy.Classifier = &ai.ClassifierPolicy{MaxQuestions: 1, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
	}, tenantA)
	b := primaryBinding(tenantA, w.provider.URL())
	b.BindingID, b.ProviderID, b.API, b.Operation = "zh-evaluation", ai.ProviderTypeSafe, ai.APITypeSafeSystemOne, ai.OperationClassifier
	b.AllowedModels = []string{"jev-1.13.0"}
	w.host.PutBinding(b)
	return w
}

func chineseReply(label string) provider.Reply {
	body := `{"model":"jev-1.13.0","answers":{"intent":{"type":"choice","choice":"` + label + `","probabilities":{"billing":0.1,"technical":0.1,"delivery":0.1,"account":0.1},"confidence":0.8}},"usage":{"input_tokens":100,"output_tokens":10}}`
	var fields map[string]any
	json.Unmarshal([]byte(body), &fields)
	fields["answers"].(map[string]any)["intent"].(map[string]any)["probabilities"].(map[string]any)[label] = 0.7
	raw, _ := json.Marshal(fields)
	return provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-typesafe-request-id": "synthetic-zh-evaluation"}, Chunks: [][]byte{raw}, Script: string(raw)}
}

func TestChineseEvaluationFixedWorkflow(t *testing.T) {
	ev := run.Case(t, "P07-zh-evaluation-fixed-workflow")
	p, w := chinesePlan(t), chineseWorld(t)
	// Independently specified labels, deliberately one error in each group.
	predictions := []string{"billing", "billing", "billing", "billing", "billing", "technical", "technical", "technical", "technical", "technical", "technical", "delivery", "delivery", "delivery", "delivery", "delivery", "delivery", "account", "account", "account", "account", "account", "account", "billing"}
	for _, label := range predictions {
		w.provider.Enqueue(chineseReply(label))
	}
	report := p.Run(context.Background(), w.client, textScope("zh-fixture"), "zh-evaluation", "fixture", "synthetic@loopback")
	ev.Record("evaluation", report)
	ev.Record("requests", w.provider.Requests())
	ev.Check("complete fixture never live", report.Status == "COMPLETE" && report.Kind == "fixture", "got %s", report.Status)
	ev.Check("all labels stay in denominator", report.Metrics.Total == 24 && report.Metrics.Scored == 24 && report.Metrics.Correct == 20 && report.Metrics.Accuracy != nil && *report.Metrics.Accuracy == 20.0/24, "got %+v", report.Metrics)
	ev.Check("per-label errors", report.Metrics.Confusion["billing"]["technical"] == 1 && report.Metrics.Confusion["account"]["billing"] == 1, "got %+v", report.Metrics.Confusion)
	ev.Check("calibration uses actual confidence", report.Metrics.ConfidenceBins[8].Count == 24 && report.Metrics.ConfidenceBins[8].Correct == 20, "got %+v", report.Metrics.ConfidenceBins)
	ev.Check("attempts priced independently", report.Budget.Calls == 24 && report.Budget.Attempts == 24 && report.Budget.InputTokens == 2400 && report.Budget.OutputTokens == 240 && report.Budget.Unreported == 0, "got %+v", report.Budget)
	report.Audit = "PASS"
	ev.Check("complete artifact verifies", p.Verify(report) == nil, "invalid report")
	raw, _ := json.Marshal(report)
	_, readErr := p.ReadReport(raw)
	ev.Check("serialized complete artifact replays", readErr == nil, "report roundtrip failed")
	report.Samples = report.Samples[:23]
	ev.Check("incomplete artifact rejected", p.Verify(report) != nil, "missing sample accepted")
}

func TestChineseEvaluationFailures(t *testing.T) {
	for _, name := range []string{"missing_answer", "wrong_key", "protocol", "usage_absent", "usage_partial", "response_drift", "http_failure", "budget", "cancel"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-"+name)
			data, _ := os.ReadFile("../examples/chineseeval/testdata/zh-support-v1.json")
			config, _ := os.ReadFile("../examples/chineseeval/testdata/jev-1.13.0-v1.json")
			var c map[string]any
			json.Unmarshal(config, &c)
			c["maxCalls"], c["maxQuestions"], c["maxAttempts"] = 1, 1, 1
			config, _ = json.Marshal(c)
			p, err := chineseeval.Load(data, config, ai.BuiltinCatalog())
			if err != nil {
				t.Fatal(err)
			}
			w := chineseWorld(t)
			reply := chineseReply("billing")
			switch name {
			case "missing_answer":
				reply.Chunks = [][]byte{[]byte(`{"model":"jev-1.13.0","answers":{},"usage":{"input_tokens":100,"output_tokens":10}}`)}
			case "wrong_key":
				reply.Chunks = [][]byte{[]byte(strings.ReplaceAll(string(reply.Chunks[0]), `"intent"`, `"other"`))}
			case "protocol":
				reply.Chunks = [][]byte{[]byte(`{"model":"jev-1.13.0","answers":`)}
			case "usage_absent", "usage_partial":
				var response map[string]any
				json.Unmarshal(reply.Chunks[0], &response)
				delete(response, "usage")
				if name == "usage_partial" {
					response["usage"] = map[string]int{"input_tokens": 100}
				}
				raw, _ := json.Marshal(response)
				reply.Chunks = [][]byte{raw}
			case "response_drift":
				reply.Chunks = [][]byte{[]byte(strings.ReplaceAll(string(reply.Chunks[0]), "jev-1.13.0", "jev-1.14.0"))}
			case "http_failure":
				reply.Status = 529
				reply.Chunks = [][]byte{[]byte(`{"detail":"synthetic failure"}`)}
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if name == "cancel" {
				reply.OnHold = cancel
				reply.End = provider.EndHold
			}
			reply.Script = string(reply.Chunks[0])
			w.provider.Enqueue(reply)
			report := p.Run(ctx, w.client, textScope("zh-failure"), "zh-evaluation", "fixture", "synthetic@loopback")
			ev.Record("evaluation", report)
			ev.Record("response-script", reply)
			ev.Record("requests", w.provider.Requests())
			want := "failed"
			switch name {
			case "budget":
				want = "answered"
			case "usage_absent", "usage_partial":
				want = "usage_incomplete"
			case "response_drift":
				want = "version_drift"
			case "cancel":
				want = "cancelled"
			}
			ev.Check("failure status with full denominator", report.Status == "FAIL" && report.Samples[0].Status == want && report.Metrics.Total == 24 && report.Budget.Calls == 1 && report.Budget.Attempts == 1, "got %+v", report)
			ev.Check("remaining samples preserved", len(report.Samples) == 24 && report.Metrics.Unexecuted == 23, "got %+v", report.Metrics)
			if name == "missing_answer" || name == "wrong_key" || name == "response_drift" || name == "usage_partial" {
				ev.Check("sent usage retained", report.Budget.InputTokens == 100, "got %+v", report.Budget)
			}
			if name == "usage_absent" || name == "http_failure" {
				ev.Check("unknown usage not free", report.Budget.Unreported == 1, "got %+v", report.Budget)
			}
			report.Audit = "PASS"
			ev.Check("honest failure artifact verifies", p.Verify(report) == nil, "got invalid artifact")
			report.Budget.Attempts = 0
			ev.Check("accounting tamper rejected", p.Verify(report) != nil, "accepted omitted attempts")
		})
	}
}

func TestChineseEvaluationInputGuards(t *testing.T) {
	for _, name := range []string{"missing_label", "unknown_label", "duplicate_sample", "no_basis", "wrong_distribution", "dataset_version", "dataset_hash", "question_key", "model_version", "catalog_version", "catalog_hash", "retry_budget", "duplicate_key"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-input-"+name)
			raw, _ := os.ReadFile("../examples/chineseeval/testdata/zh-support-v1.json")
			cfg, _ := os.ReadFile("../examples/chineseeval/testdata/jev-1.13.0-v1.json")
			var d chineseeval.Dataset
			json.Unmarshal(raw, &d)
			var c chineseeval.Config
			json.Unmarshal(cfg, &c)
			switch name {
			case "missing_label":
				d.Samples[0].Label = ""
			case "unknown_label":
				d.Samples[0].Label = "other"
			case "duplicate_sample":
				d.Samples[1].ID = d.Samples[0].ID
			case "no_basis":
				d.Samples[0].Rationale = ""
			case "wrong_distribution":
				d.Distribution["billing"] = 5
			case "dataset_version":
				c.DatasetVersion = "zh-support-v2"
			case "dataset_hash":
				c.DatasetHash = "sha256:changed"
			case "question_key":
				c.QuestionKey = "wrong"
			case "model_version":
				c.Model = "jev-latest"
			case "catalog_version":
				c.CatalogVersion = "different"
			case "catalog_hash":
				c.CatalogHash = "sha256:changed"
			case "retry_budget":
				c.MaxRetries = 1
			}
			raw, _ = json.Marshal(d)
			if name != "dataset_hash" {
				c.DatasetHash = chineseeval.Hash(raw)
			}
			cfg, _ = json.Marshal(c)
			if name == "duplicate_key" {
				cfg = []byte(strings.Replace(string(cfg), `"model":"jev-1.13.0"`, `"model":"jev-latest","model":"jev-1.13.0"`, 1))
			}
			w := chineseWorld(t)
			_, err := chineseeval.Load(raw, cfg, ai.BuiltinCatalog())
			ev.Check("invalid input never sends", err != nil && len(w.provider.Requests()) == 0, "guard accepted %s", name)
		})
	}
}

func TestChineseEvaluationNotRunAndArtifactGuards(t *testing.T) {
	ev := run.Case(t, "P07-zh-evaluation-not-run-and-artifact")
	p := chinesePlan(t)
	report := p.NotRun("disabled")
	ev.Record("evaluation", report)
	ev.Check("no invented scores", report.Status == "NOT_RUN" && report.Metrics.Total == 24 && report.Metrics.Unexecuted == 24 && report.Metrics.Accuracy == nil && report.Metrics.ScoredAccuracy == nil && report.Budget.Calls == 0, "got %+v", report)
	ev.Check("audit required", p.Verify(report) != nil, "unaudited report accepted")
	report.Audit = "PASS"
	ev.Check("NOT_RUN verifies without completion", p.Verify(report) == nil, "invalid NOT_RUN")
	// Reports are reviewable copies. Mutating one must never rewrite the plan's
	// truth set or make a second run silently use a modified annotation.
	report.Dataset.Samples[0].Label = "account"
	ev.Check("artifact cannot mutate fixed plan", p.NotRun("disabled").Dataset.Samples[0].Label == "billing", "plan was mutated by artifact")
	ev.Check("truth tamper rejected", p.Verify(report) != nil, "accepted modified truth")
}

func TestChineseEvaluationCommand(t *testing.T) {
	for _, name := range []string{"disabled", "missing_key", "foreign_credential", "verify_incomplete", "verify_observer"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-command-"+name)
			outputDir := t.TempDir()
			args := []string{"run", "./ai/examples/chineseeval/cmd/chineseeval", "-out", outputDir}
			cmd := exec.CommandContext(ctxFor(t), "go", args...)
			cmd.Dir = "../.."
			for _, key := range []string{"PATH", "HOME", "TMPDIR", "GOCACHE", "GOPATH"} {
				if value, ok := os.LookupEnv(key); ok {
					cmd.Env = append(cmd.Env, key+"="+value)
				}
			}
			if name != "disabled" {
				cmd.Args = append(cmd.Args, "-live")
				cmd.Env = append(cmd.Env, "BARNESS_AI_CHINESE_EVAL=1")
			}
			if name == "foreign_credential" {
				cmd.Env = append(cmd.Env, "BARNESS_AI_CHINESE_EVAL_TYPESAFE_KEY=key:isolated-evaluation", "OPENAI_API_KEY=key:foreign-credential")
			}
			output, err := cmd.CombinedOutput()
			ev.Record("command", map[string]any{"refused": err != nil, "notRun": strings.Contains(string(output), "NOT_RUN")})
			if name == "foreign_credential" {
				ev.Check("mixed credentials refused", err != nil, "accepted credentials")
				return
			}
			dirs, readErr := os.ReadDir(outputDir)
			if readErr != nil || len(dirs) != 1 {
				t.Fatalf("expected fresh evidence bundle")
			}
			bundle := filepath.Join(outputDir, dirs[0].Name())
			raw, readErr := os.ReadFile(filepath.Join(bundle, "evaluation-report.json"))
			if readErr != nil {
				t.Fatal(readErr)
			}
			report, parseErr := chinesePlan(t).ReadReport(raw)
			ev.Check("disabled or absent credential NOT_RUN", err == nil && parseErr == nil && report.Status == "NOT_RUN" && report.Metrics.Accuracy == nil, "command=%v parse=%v", err, parseErr)
			if name == "verify_incomplete" || name == "verify_observer" {
				if name == "verify_observer" {
					os.WriteFile(filepath.Join(bundle, "observations-invalid.json"), []byte(`[{"kind":"call_finished","state":"synthetic data"}]`), 0600)
				}
				if name == "verify_incomplete" {
					report.Samples = report.Samples[:23]
				}
				raw, _ = json.Marshal(report)
				os.WriteFile(filepath.Join(bundle, "evaluation-report.json"), raw, 0600)
				verify := exec.CommandContext(ctxFor(t), "go", "run", "./ai/examples/chineseeval/cmd/chineseeval", "-verify", bundle)
				verify.Dir = "../.."
				verify.Env = cmd.Env
				_, verifyErr := verify.CombinedOutput()
				ev.Check("file boundary rejects incomplete report", verifyErr != nil, "accepted missing sample")
			}
		})
	}
}

func TestChineseEvaluationPreCancelled(t *testing.T) {
	ev := run.Case(t, "P07-zh-evaluation-pre-cancelled")
	p, w := chinesePlan(t), chineseWorld(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := p.Run(ctx, w.client, textScope("zh-pre-cancelled"), "zh-evaluation", "fixture", "synthetic@loopback")
	ev.Record("evaluation", r)
	ev.Check("cancel before execution has no score", r.Status == "NOT_RUN" && r.Metrics.Accuracy == nil && r.Budget.Calls == 0 && len(w.provider.Requests()) == 0 && r.Reason != "", "got %+v", r.Metrics)
}

func TestChineseEvaluationImportedEvidence(t *testing.T) {
	for _, name := range []string{"negative_tokens", "incorrect_cost", "attempt_id", "http_status", "missing_observer", "missing_manifest", "incomplete_observer"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P07-zh-evaluation-import-"+name)
			source := "../../.scratch/barness-ai-pi-1.0/chinese-evaluation-evidence/live"
			dir := t.TempDir()
			entries, err := os.ReadDir(source)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				raw, err := os.ReadFile(filepath.Join(source, entry.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, entry.Name()), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(dir, "evaluation-report.json")
			raw, _ := os.ReadFile(path)
			var report chineseeval.Report
			json.Unmarshal(raw, &report)
			switch name {
			case "negative_tokens":
				a := &report.Samples[0].Metadata.Attempts[0]
				report.Budget.InputTokens -= 2 * a.Usage.Input
				a.Usage.Input = -a.Usage.Input
			case "incorrect_cost":
				a := &report.Samples[0].Metadata.Attempts[0]
				report.Budget.KnownCostUSD += 1
				a.Usage.Cost.Input += 1
				a.Usage.Cost.Total += 1
			case "attempt_id":
				report.Samples[0].Metadata.Attempts[0].AttemptID = "another-call#1"
			case "http_status":
				report.Samples[0].Metadata.Attempts[0].HTTPStatus = 529
			case "missing_observer":
				os.Remove(filepath.Join(dir, "observations-evaluation.json"))
			case "missing_manifest":
				os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{}`), 0600)
			case "incomplete_observer":
				os.WriteFile(filepath.Join(dir, "observations-evaluation.json"), []byte(`[]`), 0600)
			}
			raw, _ = json.Marshal(report)
			os.WriteFile(path, raw, 0600)
			if name == "negative_tokens" || name == "incorrect_cost" || name == "attempt_id" || name == "http_status" {
				ev.Check("coherent report tamper rejected", chinesePlan(t).Verify(report) != nil, "accepted %s", name)
			}
			cmd := exec.CommandContext(ctxFor(t), "go", "run", "./ai/examples/chineseeval/cmd/chineseeval", "-verify", dir)
			cmd.Dir = "../.."
			output, err := cmd.CombinedOutput()
			ev.Record("command", map[string]any{"refused": err != nil, "diagnostic": strings.Contains(string(output), "stage=")})
			ev.Check("file import refuses incomplete evidence", err != nil, "accepted %s", name)
			ev.Check("safe diagnostic retained", strings.Contains(string(output), "stage="), "lost stage context")
		})
	}
}

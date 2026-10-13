package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// OD53 public merge and command seam. Synthetic records below independently
// specify the required route, capability set, reservations and exact counters.
func formalDecisionsFixture(t *testing.T) map[string]any {
	t.Helper()
	at := time.Date(2026, 10, 13, 1, 0, 0, 0, time.UTC)
	ids := []string{"text-choice", "text-score", "text-bool", "mixed-questions", "image-png", "image-jpeg", "image-webp", "cancel-in-flight", "recovery"}
	scenarios := []any{}
	for _, id := range ids {
		q, images, state, imageBytes := 1, 0, 40, 0
		if id == "mixed-questions" || id == "recovery" {
			q = 3
			state = 100
		}
		if id == "image-png" || id == "image-jpeg" || id == "image-webp" {
			images = 1
			imageBytes = 100
		}
		scenarios = append(scenarios, map[string]any{"id": id, "caseId": "synthetic-" + id, "model": "gpt-6-luna", "outcome": "PASS", "calls": 1, "attempts": 1, "questions": q, "images": images, "stateBytes": state, "decodedImageBytes": imageBytes, "providerRequestIds": []string{"req-synthetic"}})
	}
	cat := ai.BuiltinCatalog()
	hash := catalogHash(t, cat)
	return map[string]any{"schema": 3, "combo": "openai-decisions", "operation": "classifier", "provider": "openai", "api": "openai-decisions", "endpoint": "https://api.openai.com/v1", "model": "gpt-6-luna", "sdk": "direct HTTP (OpenAI Decisions; ADR-0026)", "accountAlias": "unknown", "gitCommit": "synthetic-fixture", "catalogVersion": cat.Version, "catalogHash": hash, "startedAt": at.Add(-time.Minute), "finishedAt": at, "expected": ids, "scenarios": scenarios,
		"budget":    map[string]any{"maxCalls": 12, "maxAttempts": 12, "attemptsReserved": 9, "maxQuestions": 24, "maxStateBytesTotal": 4 << 20, "maxImages": 8, "maxDecodedImageBytesTotal": 4 << 20, "processSeconds": 180, "callSeconds": 30, "finalizationReserveSeconds": 10, "callsUsed": 9, "httpAttempts": 9, "questionsUsed": 13, "stateBytesSubmitted": 480, "imagesUsed": 3, "decodedImageBytesSubmitted": 300},
		"decisions": map[string]any{"mode": "live", "complete": true, "auditPassed": true, "sourceManifestHash": "61be0c194e7231aa5c3b986c849c2b23935ae6a1830038413530eb0d3618b8fe", "accountVerificationStatus": "unknown", "regionVerificationStatus": "unknown", "permissionVerificationStatus": "unknown", "accountVerificationPolicy": "allow_unknown"}}
}
func formalFixtureReport(t *testing.T, r map[string]any) supportmatrix.Report {
	var out supportmatrix.Report
	mustUnmarshal(t, mustMarshal(t, r), &out)
	return out
}
func TestDecisionsFormalMatrix(t *testing.T) {
	names := []string{"pass", "not-run", "fail", "schema2", "candidate-hash", "candidate-version", "wrong-model", "wrong-route", "both-wrong", "expected-missing", "expected-duplicate", "scenario-missing", "scenario-duplicate", "unsupported-image", "calls-over", "attempts-over", "questions-over", "state-over", "images-over", "bytes-over", "retry", "counter-disagreement", "missing-request-id", "wrong-source", "controlled", "audit-failed", "incomplete-pass", "timing-over", "future-start", "duplicate-row", "negative-input", "negative-output"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD53-formal-merge-"+name)
			raw := formalDecisionsFixture(t)
			scenarios := raw["scenarios"].([]any)
			b := raw["budget"].(map[string]any)
			proof := raw["decisions"].(map[string]any)
			old := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
			row := supportmatrix.Row{Combo: "openai-decisions", Operation: "classifier", Provider: "openai", API: "openai-decisions", Model: "gpt-6-luna", LastRunAt: &old, AllPassedAt: &old}
			for _, id := range raw["expected"].([]string) {
				row.Capabilities = append(row.Capabilities, supportmatrix.Capability{ID: id, Model: row.Model, Outcome: supportmatrix.Pass, LastRunAt: &old, LastPassedAt: &old})
			}
			matrix := supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "typesafe-classifier", Operation: "classifier", Provider: "typesafe", API: "typesafe-system-one"}, row}}
			switch name {
			case "schema2":
				raw["schema"] = 2
			case "candidate-hash":
				raw["catalogHash"] = "sha256:b00222918033826f65c545519c4b6ad57ec1e619f427472c6c53a32ae41b1503"
			case "candidate-version":
				raw["catalogVersion"] = "decisions-candidate-2026-10-13.1"
			case "wrong-model":
				raw["model"] = "jev-1.13.0"
			case "wrong-route":
				raw["api"] = "openai-responses"
			case "both-wrong":
				raw["api"] = "openai-responses"
				matrix.Rows[1].API = "openai-responses"
			case "expected-missing":
				raw["expected"] = raw["expected"].([]string)[:8]
			case "expected-duplicate":
				raw["expected"].([]string)[8] = "text-choice"
			case "scenario-missing":
				raw["scenarios"] = scenarios[:8]
			case "scenario-duplicate":
				raw["scenarios"] = append(scenarios, scenarios[0])
			case "unsupported-image":
				scenarios[4].(map[string]any)["outcome"] = "UNSUPPORTED"
				scenarios[4].(map[string]any)["note"] = "synthetic missing image"
			case "negative-input":
				b["inputTokensUsed"] = -1
			case "negative-output":
				b["outputTokensUsed"] = -1
			case "calls-over":
				b["callsUsed"] = 13
			case "attempts-over":
				b["httpAttempts"] = 13
			case "questions-over":
				b["questionsUsed"] = 25
			case "state-over":
				b["stateBytesSubmitted"] = (4 << 20) + 1
			case "images-over":
				b["imagesUsed"] = 9
			case "bytes-over":
				b["decodedImageBytesSubmitted"] = (4 << 20) + 1
			case "retry":
				b["environmentRetries"] = 1
			case "counter-disagreement":
				b["callsUsed"] = 8
			case "missing-request-id":
				scenarios[0].(map[string]any)["providerRequestIds"] = []string{}
			case "wrong-source":
				proof["sourceManifestHash"] = "rewritten"
			case "controlled":
				proof["mode"] = "controlled"
			case "audit-failed":
				proof["auditPassed"] = false
			case "incomplete-pass":
				proof["complete"] = false
			case "timing-over":
				raw["startedAt"] = raw["finishedAt"].(time.Time).Add(-181 * time.Second)
			case "future-start":
				raw["startedAt"] = raw["finishedAt"].(time.Time).Add(time.Second)
			case "duplicate-row":
				matrix.Rows = append(matrix.Rows, row)
			case "fail":
				scenarios[4].(map[string]any)["outcome"] = "FAIL"
				scenarios[4].(map[string]any)["errorCategory"] = "upstream_error"
				proof["complete"] = false
			case "not-run":
				proof["complete"] = false
				for _, s := range scenarios {
					x := s.(map[string]any)
					x["outcome"] = "NOT_RUN"
					for _, k := range []string{"calls", "attempts", "questions", "images", "stateBytes", "decodedImageBytes"} {
						x[k] = 0
					}
					x["providerRequestIds"] = []string{}
				}
				for _, k := range []string{"callsUsed", "attemptsReserved", "httpAttempts", "questionsUsed", "stateBytesSubmitted", "imagesUsed", "decodedImageBytesSubmitted"} {
					b[k] = 0
				}
			}
			r := formalFixtureReport(t, raw)
			before := mustMarshal(t, matrix)
			out, err := supportmatrix.Merge(matrix, r)
			ev.Record("synthetic-report", raw)
			ev.Record("merge-error", errString(err))
			allowed := name == "pass" || name == "fail" || name == "not-run"
			ev.Check("fixed formal contract", (err == nil) == allowed, "err=%v", err)
			ev.Check("caller unchanged", bytes.Equal(before, mustMarshal(t, matrix)), "input mutation")
			if err == nil {
				ev.Record("merged", out)
				ev.Check("Jev independent", reflect.DeepEqual(out.Rows[0], matrix.Rows[0]), "Jev changed")
				if name == "not-run" {
					ev.Check("NOT_RUN retains old PASS", reflect.DeepEqual(out, matrix), "history updated")
				}
				if name == "fail" {
					ev.Check("FAIL blocks current with old pass time retained", out.Rows[1].Capabilities[4].Outcome == supportmatrix.Fail && out.Rows[1].AllPassedAt.Equal(old) && out.Rows[1].Capabilities[4].LastPassedAt.Equal(old), "history lost")
				}
			}
		})
	}
}
func TestDecisionsFormalBatchAtomic(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "supportmatrix")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", bin, "./ai/live/cmd/supportmatrix")
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, out)
	}
	for _, kind := range []string{"schema2", "wrong-identity", "candidate", "narrowed"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "P10-OD53-formal-batch-"+kind)
			dir := t.TempDir()
			matrix := filepath.Join(dir, "matrix.json")
			m := supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses"}, {Combo: "openai-decisions", Operation: "classifier", Provider: "openai", API: "openai-decisions"}}}
			before := mustMarshal(t, m)
			os.WriteFile(matrix, before, 0600)
			good := supportmatrix.Report{Schema: 3, Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses", FinishedAt: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), Expected: []string{"text-stream"}, Scenarios: []supportmatrix.ScenarioResult{{ID: "text-stream", Outcome: supportmatrix.Pass}}}
			os.WriteFile(filepath.Join(dir, "good.json"), mustMarshal(t, good), 0600)
			bad := formalDecisionsFixture(t)
			switch kind {
			case "schema2":
				bad["schema"] = 2
			case "wrong-identity":
				bad["api"] = "openai-responses"
			case "candidate":
				bad["catalogHash"] = "candidate"
			case "narrowed":
				bad["expected"] = []string{"text-choice"}
			}
			os.WriteFile(filepath.Join(dir, "bad.json"), mustMarshal(t, bad), 0600)
			cmd := exec.CommandContext(ctxFor(t), bin, "-matrix", matrix, filepath.Join(dir, "good.json"), filepath.Join(dir, "bad.json"))
			output, err := cmd.CombinedOutput()
			after, _ := os.ReadFile(matrix)
			ev.Record("command", map[string]any{"output": string(output), "refused": err != nil})
			ev.Check("rejected batch never partially writes", err != nil && bytes.Equal(before, after), "exit=%v", err)
		})
	}
}

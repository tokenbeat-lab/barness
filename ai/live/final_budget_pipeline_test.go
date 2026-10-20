//go:build live

package live

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Exercise full unchanged classifier plans with allowance enabled, not merely
// their guard-disabled controlled regressions. No real credential or service.
func TestFinalBudgetClassifierHomeArtifacts(t *testing.T) {
	cs := run.Case(t, "REGIONAL-budget-home-artifacts")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "../.."))
	// The real checkout and child cwd are inside this synthetic HOME. Only
	// this probe's evidence root is aliased in the outer process.
	home := filepath.Dir(root)
	base, err := os.MkdirTemp(filepath.Join(root, ".evidence"), "budget-home-artifacts-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(base)
	run.RedactSecret(base, "$BUDGET_HOME_PROBE")
	cmd := exec.Command(os.Args[0], "-test.run=^TestFinalBudgetClassifierPipelines$", "-test.timeout=90s", "-test.v")
	cmd.Dir = cwd
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "TMPDIR=" + os.TempDir(), "HOME=" + home, "BARNESS_AI_EVIDENCE_DIR=" + base}
	output, err := cmd.CombinedOutput()
	cs.Record("process", map[string]any{"argv": cmd.Args, "exitSuccess": err == nil, "output": string(output)})
	if !cs.Check("full public pipelines and unchanged TestMain audit pass inside HOME", err == nil, "%v %s", err, output) {
		return
	}
	for _, route := range []string{"decisions", "protocol", "typesafe"} {
		matches, _ := filepath.Glob(filepath.Join(base, "*", "REGIONAL-budget-full-"+route, "process.json"))
		if !cs.Check("retained child diagnostic exists "+route, len(matches) == 1, "count=%d", len(matches)) {
			continue
		}
		data, e := os.ReadFile(matches[0])
		cs.Fixture(route+"-process.json", data)
		var process struct {
			Argv    []string `json:"argv"`
			Success bool     `json:"exitSuccess"`
			Output  string   `json:"output"`
		}
		decodeErr := json.Unmarshal(data, &process)
		cs.Check("redacted path preserves full child diagnostics "+route, e == nil && decodeErr == nil && process.Success && len(process.Argv) == 4 && strings.Contains(process.Output, "PASS\n") && strings.Contains(process.Output, "$BUDGET_PIPELINE_ROOT/budget-pipeline-"+route+"/") && !strings.Contains(string(data), home), "read=%v decode=%v argv=%d success=%v", e, decodeErr, len(process.Argv), process.Success)
	}
	audits, _ := filepath.Glob(filepath.Join(base, "*", "audit.json"))
	if cs.Check("unchanged environment audit retained", len(audits) == 1, "count=%d", len(audits)) {
		data, e := os.ReadFile(audits[0])
		cs.Fixture("child-audit.json", data)
		var audit struct {
			Findings []json.RawMessage `json:"findings"`
		}
		decodeErr := json.Unmarshal(data, &audit)
		cs.Check("zero forbidden environment findings", e == nil && decodeErr == nil && len(audit.Findings) == 0, "read=%v decode=%v findings=%d", e, decodeErr, len(audit.Findings))
	}
}

func TestFinalBudgetClassifierPipelines(t *testing.T) {
	for _, route := range []string{"decisions", "protocol", "typesafe"} {
		t.Run(route, func(t *testing.T) {
			cs := run.Case(t, "REGIONAL-budget-full-"+route)
			base := filepath.Join(run.Dir(), "budget-pipeline-"+route)
			// Child Finish reports its evidence location. Hide only this owned
			// pipeline root; retain route/session suffixes and the full output.
			// HOME stays forbidden so unrelated environment leaks still fail.
			run.RedactSecret(base, "$BUDGET_PIPELINE_ROOT/"+filepath.Base(base))
			if err := os.MkdirAll(base, 0700); err != nil {
				t.Fatal(err)
			}
			budget := filepath.Join(base, "usd.json")
			maximum := int64(200000)
			if route == "typesafe" {
				maximum = 100000
			}
			initial, _ := json.Marshal(map[string]any{"schema": 1, "round": "provider-sites-final-acceptance-2", "maxMicrousd": maximum, "reservedMicrousd": 0, "attempts": 0})
			if err := os.WriteFile(budget, initial, 0600); err != nil {
				t.Fatal(err)
			}
			env := []string{"PATH=" + os.Getenv("PATH"), "TMPDIR=" + os.TempDir(), "BARNESS_FINAL_ROUND=provider-sites-final-acceptance-2", "BARNESS_FINAL_BUDGET=" + budget, "BARNESS_AI_EVIDENCE_DIR=" + base}
			entry, want := "^TestDecisionsFormalControlledChild$", 9
			switch route {
			case "decisions":
				env = append(env, "BARNESS_DECISIONS_FORMAL_CONTROLLED=success", "BARNESS_AI_LIVE=1", "BARNESS_AI_LIVE_COMBO=openai-decisions", "BARNESS_AI_LIVE_KEY_OPENAI_DECISIONS=key:controlled-decisions", "BARNESS_AI_LIVE_ALLOW_UNKNOWN_ACCOUNT=1")
			case "protocol":
				entry, want = "^TestMultimodalFormal$", 12
				proof := filepath.Join(base, "controlled-proof.json")
				raw, _ := json.Marshal(multimodalAuthorization{Route: "jev", Model: "jev-1.14.0", Service: "https://api.typesafe.ai/v1/systemone", AccountAlias: "controlled", RegionAlias: "offline", AuthorizationReference: "issue18-controlled", DedicatedCredential: true, MaxCalls: 12, MaxAttempts: 12})
				if err := os.WriteFile(proof, raw, 0600); err != nil {
					t.Fatal(err)
				}
				env = append(env, "BARNESS_MM_ROUTE=jev", "BARNESS_MM_FORMAL=1", "BARNESS_MM_CONTROLLED=success", "BARNESS_MM_LIVE=1", "BARNESS_MM_KEY=controlled-mm-credential", "BARNESS_MM_AUTHORIZATION="+proof)
			case "typesafe":
				entry, want = "^TestFinalBudgetTypeSafeChild$", 3
				env = append(env, "BARNESS_BUDGET_TYPESAFE_CONTROLLED=1")
			}
			cmd := exec.Command(os.Args[0], "-test.run="+entry, "-test.timeout=60s", "-test.v")
			cmd.Env = env
			output, err := cmd.CombinedOutput()
			cs.Record("process", map[string]any{"argv": cmd.Args, "exitSuccess": err == nil, "output": string(output)})
			raw, _ := os.ReadFile(budget)
			var ledger struct {
				Attempts int   `json:"attempts"`
				Reserved int64 `json:"reservedMicrousd"`
			}
			json.Unmarshal(raw, &ledger)
			cs.Fixture("usd.json", raw)
			cs.Check("complete public classifier pipeline with final allowance", err == nil && ledger.Attempts == want && ledger.Reserved > 0 && ledger.Reserved <= maximum, "err=%v ledger=%+v %s", err, ledger, output)
		})
	}
}

func TestFinalBudgetTypeSafeChild(t *testing.T) {
	if os.Getenv("BARNESS_BUDGET_TYPESAFE_CONTROLLED") != "1" {
		t.Skip("controlled child only")
	}
	c := &combo{name: "typesafe-classifier", operation: ai.OperationClassifier, provider: ai.ProviderTypeSafe, api: ai.APITypeSafeSystemOne, endpoint: "https://api.typesafe.ai/v1", model: "jev-1.13.0"}
	rec := newRecorder()
	rec.next.(*http.Transport).CloseIdleConnections()
	count := 0
	rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		count++
		raw, _ := io.ReadAll(req.Body)
		req.Body.Close()
		status := 200
		body := `{"model":"jev-1.13.0","answers":{"intent":{"type":"choice","choice":"billing","probabilities":{"billing":0.9,"technical":0.1},"confidence":0.8}`
		if strings.Contains(string(raw), `"severity"`) {
			body += `,"severity":{"type":"score","score":0,"confidence":1,"probabilities":{"0":1,"1":0,"2":0},"legend":{"0":"Low","1":"Medium","2":"High"}},"urgent":{"type":"noul","noul":0.9}`
		}
		body += `},"usage":{"input_tokens":100,"output_tokens":0}}`
		if len(raw) > 80000 {
			status = 422
			body = `{"detail":"context token limit exceeded"}`
		}
		header := http.Header{}
		header.Set("Content-Type", "application/json")
		header.Set("x-typesafe-request-id", "synthetic-budget-request")
		return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})
	client, err := newClient(c, "synthetic-budget-credential", rec, ai.BuiltinCatalog())
	if err != nil {
		t.Fatal(err)
	}
	before := budget
	defer func() { budget = before }()
	budget = classifierBudget()
	env := &liveEnv{combo: c, client: client, rec: rec}
	for _, sc := range classifierScenarios() {
		t.Run(sc.id, func(t *testing.T) {
			s := &session{t: t, ctx: context.Background(), cs: run.Case(t, "REGIONAL-budget-typesafe-"+sc.id), env: env, model: c.model, attempt: 1}
			sc.run(s)
		})
	}
	if count != 3 {
		t.Fatalf("full generic classifier inventory calls=%d", count)
	}
}

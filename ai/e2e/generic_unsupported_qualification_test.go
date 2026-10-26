package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Failure inventory: only the fixed native OpenAI Chat reasoning-history
// declaration may retain UNSUPPORTED without a request. Different identities,
// reasons, scenarios, nonzero activity and missing/failed capabilities cannot
// borrow that exception. Report validation alone never grants live support.
func TestGenericDeclaredUnsupported(t *testing.T) {
	for _, mutation := range []string{"declared", "pretended-pass", "route", "provider", "api", "endpoint", "default-model", "scene-model", "reason", "scene", "calls", "attempts", "retries", "request-id", "images", "questions", "state-bytes", "decoded-bytes", "image-call", "fault", "environment", "budget", "mandatory-history", "other-unsupported", "other-fail", "other-not-run", "missing", "duplicate"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "GROK-GENERIC-UNSUPPORTED-"+mutation)
			r := genericUnsupportedReport(t)
			s := &r.Scenarios[4]
			switch mutation {
			case "pretended-pass":
				s.Outcome, s.Note, s.Calls, s.Attempts = supportmatrix.Pass, "", 1, 1
				r.Budget.CallsUsed++
				r.Budget.HTTPAttempts++
			case "route":
				r.Combo = "openai-responses"
			case "provider":
				r.Provider = "deepseek"
			case "api":
				r.API = "openai-responses"
			case "endpoint":
				r.Endpoint = "https://wrong.example/v1"
			case "default-model":
				r.Model = "gpt-5.4"
			case "scene-model":
				s.Model = "gpt-5.4"
			case "reason":
				s.Note = "synthetic arbitrary optional capability"
			case "scene":
				s.ID = "optional-history"
			case "calls":
				s.Calls = 1
				r.Budget.CallsUsed++
			case "attempts":
				s.Attempts = 1
				r.Budget.HTTPAttempts++
			case "retries":
				s.Retries = 1
				r.Budget.EnvironmentRetries++
			case "request-id":
				s.ProviderRequestIDs = []string{"synthetic-request"}
			case "images":
				s.Images = 1
			case "questions":
				s.Questions = 1
			case "state-bytes":
				s.StateBytes = 1
			case "decoded-bytes":
				s.DecodedImageBytes = 1
			case "image-call":
				s.ImageCalls = []supportmatrix.ImageCall{{}}
			case "fault":
				s.ErrorCategory = "upstream_error"
			case "environment":
				s.Environment = true
			case "budget":
				r.Budget.HTTPAttempts++
			case "mandatory-history", "other-unsupported":
				i := 0
				if mutation == "mandatory-history" {
					i = len(r.Scenarios) - 1
				}
				r.Scenarios[i].Outcome, r.Scenarios[i].Note = supportmatrix.Unsupported, s.Note
				r.Scenarios[i].Calls, r.Scenarios[i].Attempts = 0, 0
				r.Budget.CallsUsed--
				r.Budget.HTTPAttempts--
			case "other-fail":
				r.Scenarios[0].Outcome, r.Scenarios[0].ErrorCategory = supportmatrix.Fail, "upstream_error"
			case "other-not-run":
				r.Scenarios[0].Outcome = supportmatrix.NotRun
			case "missing":
				r.Scenarios = r.Scenarios[1:]
				r.Budget.CallsUsed--
				r.Budget.HTTPAttempts--
			case "duplicate":
				r.Scenarios[0] = r.Scenarios[1]
			}
			err := supportmatrix.VerifyGenericReport(r)
			ev.Record("synthetic-contract", map[string]any{"report": r, "error": err, "notActualLive": true, "vendorCalls": 0})
			ev.Check("only declared zero-request native exception qualifies report", (err == nil) == (mutation == "declared"), "mutation=%s err=%v", mutation, err)
			if mutation == "declared" {
				m := supportmatrix.Matrix{Schema: 3, Rows: []supportmatrix.Row{{Combo: r.Combo, Operation: r.Operation, Provider: r.Provider, API: r.API}}}
				got, mergeErr := supportmatrix.Merge(m, r)
				if mergeErr != nil {
					t.Fatal(mergeErr)
				}
				for _, capability := range got.Rows[0].Capabilities {
					if capability.ID == "reasoning-history" {
						ev.Check("matrix retains unsupported without last-passed", capability.Outcome == supportmatrix.Unsupported && capability.LastPassedAt == nil && capability.Note == s.Note, "%+v", capability)
					}
				}
				genericUnsupportedCLI(t, r)
			}
		})
	}
}

func genericUnsupportedReport(t *testing.T) supportmatrix.Report {
	t.Helper()
	cat := ai.BuiltinCatalog()
	hash, err := cat.Hash()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	r := supportmatrix.Report{Schema: 3, Combo: "openai-chat", Operation: "chat", Provider: "openai", API: "openai-completions", Endpoint: "https://api.openai.com/v1", Model: "gpt-5-mini", SDK: "github.com/openai/openai-go/v3 v3.66.0", AccountAlias: "synthetic", GitCommit: "synthetic-not-actual-live", CatalogVersion: cat.Version, CatalogHash: hash, StartedAt: now, FinishedAt: now.Add(time.Second), Expected: supportmatrix.GenericExpected("openai-chat"), Budget: supportmatrix.Budget{MaxCalls: 32, MaxRetries: 1, MaxOutputTokens: 4096}}
	for _, id := range r.Expected {
		model := r.Model
		if strings.HasPrefix(id, "tool-changes-") {
			model = strings.TrimPrefix(id, "tool-changes-")
		}
		if id == "pro-model-unavailable" {
			model = "gpt-5.4-pro"
		}
		s := supportmatrix.ScenarioResult{ID: id, CaseID: "synthetic-" + id, Model: model, Outcome: supportmatrix.Pass, Calls: 1, Attempts: 1}
		if id == "reasoning-history" {
			s.Outcome, s.Calls, s.Attempts = supportmatrix.Unsupported, 0, 0
			s.Note = "OpenAI Chat Completions returns no reasoning content or signature to replay; reasoning is only counted in usage"
		}
		r.Scenarios = append(r.Scenarios, s)
		r.Budget.CallsUsed += s.Calls
		r.Budget.HTTPAttempts += s.Attempts
	}
	return r
}

func genericUnsupportedCLI(t *testing.T, r supportmatrix.Report) {
	t.Helper()
	ev := run.Case(t, "GROK-GENERIC-UNSUPPORTED-CLI-INELIGIBLE")
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "releasegate")
	build := exec.CommandContext(ctxFor(t), "go", "build", "-o", binary, "./ai/release/cmd/releasegate")
	build.Dir, build.Env = root, xaiProcessEnv(t)
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("consumer build: %v %s", err, output)
	}
	bundle, err := os.MkdirTemp(run.Dir(), "synthetic-unsupported-")
	if err != nil {
		t.Fatal(err)
	}
	writeReleaseJSON(t, bundle, "live-report.json", r)
	rel, err := filepath.Rel(root, bundle)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(run.Dir(), "synthetic-unsupported-consumer")
	cmd := exec.CommandContext(ctxFor(t), binary, "-live-bundle", filepath.ToSlash(rel), "-out", out)
	cmd.Dir, cmd.Env = root, xaiProcessEnv(t)
	output, processErr := cmd.CombinedOutput()
	exit := -1
	if cmd.ProcessState != nil {
		exit = cmd.ProcessState.ExitCode()
	}
	var proof struct {
		Pass  bool   `json:"pass"`
		Error string `json:"error"`
	}
	raw, err := os.ReadFile(filepath.Join(out, "live-consumer.json"))
	if err != nil || json.Unmarshal(raw, &proof) != nil {
		t.Fatal("synthetic rejection proof unavailable")
	}
	ev.Record("synthetic-process", map[string]any{"argv": cmd.Args, "exit": exit, "output": string(output), "notActualLive": true})
	ev.Check("report-only acceptance grants no live qualification", processErr != nil && exit == 1 && !proof.Pass && proof.Error == "native live consumer rejected evidence", "exit=%d pass=%t", exit, proof.Pass)
}

// This boundary consumes the already authenticated native case inventory and
// assertion shape, rather than a report's self-declared zero counters alone.
func TestGenericUnsupportedCaseInventory(t *testing.T) {
	for _, mutation := range []string{"assertions-only", "http", "result", "observer", "unrecognized-capture", "missing-assertions", "reason", "assertions", "calls"} {
		t.Run(mutation, func(t *testing.T) {
			ev := run.Case(t, "GROK-GENERIC-UNSUPPORTED-ARTIFACT-"+mutation)
			r := genericUnsupportedReport(t)
			s := r.Scenarios[4]
			artifacts := map[string]string{"assertions.json": strings.Repeat("0", 64)}
			reason, assertions := s.Note, 0
			switch mutation {
			case "http":
				artifacts["attempt-1-exchanges.json"] = strings.Repeat("0", 64)
			case "result":
				artifacts["attempt-1-result.json"] = strings.Repeat("0", 64)
			case "observer":
				artifacts["observations-attempt-1.json"] = strings.Repeat("0", 64)
			case "unrecognized-capture":
				artifacts["hidden-http.json"] = strings.Repeat("0", 64)
			case "missing-assertions":
				delete(artifacts, "assertions.json")
			case "reason":
				reason = "synthetic arbitrary reason"
			case "assertions":
				assertions = 1
			case "calls":
				s.Calls = 1
			}
			err := supportmatrix.VerifyGenericUnsupportedCase(r, s, artifacts, reason, assertions)
			ev.Record("synthetic-artifact-contract", map[string]any{"artifacts": artifacts, "reason": reason, "assertions": assertions, "error": err, "notActualLive": true})
			ev.Check("unsupported case contains only matching native declaration", (err == nil) == (mutation == "assertions-only"), "mutation=%s err=%v", mutation, err)
		})
	}
}

// All dynamic unsupported branches are checked against the native smoke's
// fixed current models, not against a model a report may substitute.
func TestGenericUnsupportedFixedCatalog(t *testing.T) {
	cat := ai.BuiltinCatalog()
	for _, target := range []struct {
		route    string
		provider ai.ProviderID
		api      ai.API
		model    string
	}{
		{"openai-responses", ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-5-mini"},
		{"openai-chat", ai.ProviderOpenAI, ai.APIOpenAICompletions, "gpt-5-mini"},
		{"deepseek-responses", ai.ProviderDeepSeek, ai.APIOpenAIResponses, "deepseek-flash"},
		{"deepseek-chat", ai.ProviderDeepSeek, ai.APIOpenAICompletions, "deepseek-flash"},
		{"anthropic-messages", ai.ProviderAnthropic, ai.APIAnthropicMessages, "claude-haiku-4-5"},
		{"google-gemini", ai.ProviderGoogle, ai.APIGoogleGenerativeAI, "gemini-3.8-flash"},
	} {
		t.Run(target.route, func(t *testing.T) {
			ev := run.Case(t, "GROK-GENERIC-UNSUPPORTED-CATALOG-"+target.route)
			model, ok := cat.Lookup(target.provider, target.api, target.model)
			ev.Record("fixed-model", model)
			ev.Check("fixed default cannot take no-reasoning or text-only skip", ok && model.Reasoning && slices.Contains(model.Input, ai.ModalityImage), "model=%s found=%t", target.model, ok)
			if target.provider == ai.ProviderDeepSeek {
				levelModel := model
				if target.api == ai.APIOpenAICompletions {
					levelModel, ok = cat.Lookup(target.provider, target.api, "deepseek-v4-pro")
				}
				ev.Check("fixed extra high/max levels cannot take null-map skip", ok && !levelModel.ThinkingLevelMap.High.IsNull() && !levelModel.ThinkingLevelMap.Max.IsNull(), "model=%s", levelModel.ID)
			}
		})
	}
}

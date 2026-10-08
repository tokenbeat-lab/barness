//go:build live

package live

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// These are the existing harness's isolated public Client / controlled TLS
// Provider seams. Failure modes precede implementation in issue 08 FAILURES.md.
func TestClassifierHarnessBudget(t *testing.T) {
	before := budget
	defer func() { budget = before }()
	budget = classifierBudget()
	s := &session{t: t, ctx: context.Background(), cs: run.Case(t, "P07-live-harness-budget")}
	for range 3 {
		if !s.spendQuestions(3, 100) {
			t.Fatal("budget refused early")
		}
	}
	if budget.QuestionsUsed != 9 || budget.CallsUsed != 3 || budget.StateBytesSubmitted != 300 {
		t.Fatal("missing actual budget counters")
	}
	// The next reservation is observed without marking this passing verifier
	// failed; the scenario itself exercises FAIL/budget in controlled subruns.
	if classifierBudgetAllows(2, 100) || classifierBudgetAllows(1, 131073) {
		t.Fatal("overspend allowed")
	}
	if !classifierBudgetAllows(1, 100) {
		t.Fatal("last available question refused")
	}
}

func TestClassifierHarnessWire(t *testing.T) {
	for _, name := range []string{"success", "usage-absent", "422", "context-400", "other-400", "invalid-answers", "empty-questions"} {
		t.Run(name, func(t *testing.T) {
			before := budget
			defer func() { budget = before }()
			budget = classifierBudget()
			cs := run.Case(t, "P07-live-harness-"+name)
			count := 0
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count++
				var body map[string]any
				json.NewDecoder(r.Body).Decode(&body)
				cs.Check("fixed model and native question", body["model"] == "jev-1.13.0" && strings.HasSuffix(r.URL.Path, "/systemone"), "wrong request route")
				w.Header().Set("x-typesafe-request-id", "synthetic-request-id")
				w.Header().Set("Content-Type", "application/json")
				if name == "context-400" || name == "other-400" {
					w.WriteHeader(400)
					if name == "context-400" {
						w.Write([]byte(`{"detail":{"error_type":"max_tokens_exceeded"}}`))
					} else {
						w.Write([]byte(`{"detail":"invalid account"}`))
					}
					return
				}
				if name == "422" {
					w.WriteHeader(422)
					w.Write([]byte(`{"detail":"context token limit exceeded"}`))
					return
				}
				response := `{"model":"jev-1.13.0","answers":{"intent":{"type":"choice","choice":"billing","probabilities":{"billing":0.9,"technical":0.1},"confidence":0.8}}`
				if name == "invalid-answers" {
					response = `{"model":"jev-1.13.0","answers":{}`
				}
				if name != "usage-absent" {
					response += `,"usage":{"input_tokens":100,"output_tokens":10}`
				}
				w.Write([]byte(response + `}`))
			}))
			defer server.Close()
			c := &combos[len(combos)-1]
			rec := newRecorder()
			defer rec.next.(*http.Transport).CloseIdleConnections()
			transport := server.Client().Transport
			rec.next = roundTripFunc(func(req *http.Request) (*http.Response, error) {
				copy := req.Clone(req.Context())
				u := *req.URL
				copy.URL = &u
				target := strings.TrimPrefix(server.URL, "https://")
				copy.URL.Host = target
				copy.Host = target
				return transport.RoundTrip(copy)
			})
			cat := ai.BuiltinCatalog()
			var facts struct{ Model ai.ClassifierModel }
			readOfficial(t, &facts)
			cat.ClassifierModels = []ai.ClassifierModel{facts.Model}
			client, err := newClient(c, "key:typesafe-test", rec, cat)
			if err != nil {
				t.Fatal(err)
			}
			s := &session{t: t, ctx: context.Background(), cs: cs, env: &liveEnv{combo: c, client: client, rec: rec}, model: c.model, attempt: 1}
			req := singleChoiceRequest()
			if name == "empty-questions" {
				req.Questions = nil
			}
			res, err := s.classify("controlled", req)
			cs.Record("classification", res)
			if name == "empty-questions" {
				cs.Check("local refusal", err != nil && count == 0 && len(res.Metadata.Attempts) == 0, "sent=%d", count)
				return
			}
			cs.Check("one actual attempt", count == 1 && budget.HTTPAttempts == 1 && s.httpAttempts == 1, "count=%d", count)
			if name == "context-400" || name == "other-400" {
				cs.Check("only explicit context rejection accepted", context400(res, err, s.lastExchanges) == (name == "context-400"), "err=%v", err)
				return
			}
			if name == "422" {
				cs.Check("422 classification", isContext422(res, err), "err=%v", err)
				return
			}
			if name == "invalid-answers" {
				cs.Check("bad answers retain usage", err != nil && len(res.Answers) == 0 && res.Usage.Input == 100, "err=%v", err)
				return
			}
			cs.Check("success", err == nil && res.Metadata.Operation == ai.OperationClassifier, "err=%v", err)
			want := ai.UsageComplete
			if name == "usage-absent" {
				want = ai.UsageUnreported
			}
			cs.Check("absence differs from zero", res.Metadata.Attempts[0].UsageReporting == want, "got %s", res.Metadata.Attempts[0].UsageReporting)
			cs.Check("input only price", res.Usage.Cost.Output == 0 && (name == "usage-absent" || math.Abs(res.Usage.Cost.Input-0.0000042) < 1e-12), "cost %+v", res.Usage.Cost)
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestClassifierHarnessNotRun(t *testing.T) {
	for _, c := range combos {
		t.Setenv(keyVar(c.name), "")
	}
	t.Setenv(envCombo, "typesafe-classifier")
	t.Setenv(envLive, "1")
	cfg := loadConfig()
	if cfg.combo == nil || cfg.notRun == "" || cfg.refused != "" {
		t.Fatal("missing key must be NOT_RUN")
	}
	t.Setenv(keyVar("typesafe-classifier"), "key:typesafe-test")
	t.Setenv(envAlias, "local-test@unknown")
	cfg = loadConfig()
	if cfg.notRun != "" || cfg.refused != "" {
		t.Fatal("isolated key refused")
	}
	t.Setenv(keyVar("openai-chat"), "key:foreign-test")
	if loadConfig().refused == "" {
		t.Fatal("another combination key allowed")
	}
}

package supportmatrix

import (
	"errors"
	"reflect"
	"testing"
)

// Failure modes are recorded before implementation in issue 08 FAILURES.md.
// The merger is the existing public seam of the isolated report mechanism.
func classifierReport() (Matrix, Report) {
	m := baseMatrix()
	m.Rows = append(m.Rows, Row{Combo: "typesafe-classifier", Operation: "classifier", Provider: "typesafe", API: "typesafe-system-one"})
	r := report("typesafe-classifier", "typesafe", "typesafe-system-one", t1,
		pass("mixed-questions"), pass("single-choice"), pass("context-422"))
	r.Operation, r.Model, r.Endpoint = "classifier", "jev-1.13.0", "https://api.typesafe.ai/v1"
	r.Budget = Budget{MaxCalls: 6, MaxRetries: 1, CallsUsed: 3, MaxQuestions: 10, QuestionsUsed: 5, HTTPAttempts: 3, MaxStateBytes: 131072, StateBytesSubmitted: 80100}
	for i := range r.Scenarios {
		s := &r.Scenarios[i]
		s.Model, s.Calls, s.Attempts, s.Questions = r.Model, 1, 1, 1
		s.ProviderRequestIDs = []string{"test-request-id"}
	}
	r.Scenarios[0].Questions = 3
	return m, r
}
func TestClassifierReportRefusals(t *testing.T) {
	for name, change := range map[string]func(*Report){
		"wrong-operation":    func(r *Report) { r.Operation = "chat" },
		"alias":              func(r *Report) { r.Model = "jev-latest" },
		"wrong-target":       func(r *Report) { r.Endpoint = "https://proxy.invalid" },
		"missing-expected":   func(r *Report) { r.Expected = r.Expected[:2] },
		"duplicate-expected": func(r *Report) { r.Expected = append(r.Expected, r.Expected[0]) },
		"call-budget":        func(r *Report) { r.Budget.CallsUsed = 7 },
		"question-budget":    func(r *Report) { r.Budget.QuestionsUsed = 11 },
		"missing-budget":     func(r *Report) { r.Budget.MaxQuestions = 0 },
		"actual-counts":      func(r *Report) { r.Budget.HTTPAttempts = 0 },
		"no-request":         func(r *Report) { r.Scenarios[0].Calls = 0 },
		"no-vendor-id":       func(r *Report) { r.Scenarios[0].ProviderRequestIDs = nil },
		"substitute-model":   func(r *Report) { r.Scenarios[0].Model = "jev-preview" },
		"missing-capability": func(r *Report) { r.Scenarios[0].Outcome = Unsupported; r.Scenarios[0].Note = "not supported" },
	} {
		t.Run(name, func(t *testing.T) {
			m, r := classifierReport()
			before, _ := Encode(m)
			change(&r)
			_, err := Merge(m, r)
			if !errors.Is(err, ErrInvalidReport) {
				t.Fatalf("accepted invalid classifier report: %v", err)
			}
			after, _ := Encode(m)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("refusal changed input")
			}
		})
	}
}
func TestClassifierReportOwnEvidenceOnly(t *testing.T) {
	m, r := classifierReport()
	got, err := Merge(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if got.Rows[2].AllPassedAt == nil || !reflect.DeepEqual(got.Rows[:2], m.Rows[:2]) {
		t.Fatal("pass crossed route or did not record complete evidence")
	}
	r.FinishedAt = t2
	r.Scenarios = r.Scenarios[:1]
	r.Budget.CallsUsed = 1
	r.Budget.HTTPAttempts = 1
	r.Budget.QuestionsUsed = 3
	partial, err := Merge(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if partial.Rows[2].AllPassedAt != nil {
		t.Fatal("partial run fabricated complete pass")
	}
}

func TestReportAndMatrixCannotAgreeOnWrongRoute(t *testing.T) {
	for _, name := range []string{"wrong-operation", "wrong-provider", "wrong-api"} {
		t.Run(name, func(t *testing.T) {
			m, r := classifierReport()
			switch name {
			case "wrong-operation":
				m.Rows[0].Operation = "classifier"
			case "wrong-provider":
				m.Rows[0].Provider = "deepseek"
			case "wrong-api":
				m.Rows[0].API = "openai-completions"
			}
			r = report(m.Rows[0].Combo, m.Rows[0].Provider, m.Rows[0].API, t1, pass("text-stream"))
			r.Operation = m.Rows[0].Operation
			if _, err := Merge(m, r); !errors.Is(err, ErrInvalidReport) {
				t.Fatal("shared incorrect identity accepted")
			}
		})
	}
}

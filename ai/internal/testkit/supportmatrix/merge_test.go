package supportmatrix

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

// The merge is the one place a live report changes the support matrix, so
// it is tested in isolation. Ways it could go wrong, written before the code:
//
//  1. A report names a combination the matrix does not list: a row is
//     invented instead of refused.
//  2. A report's provider or API disagrees with its combination's row: a
//     shared adapter's pass is written onto another combination.
//  3. The report's schema is not the one this code reads.
//  4. A scenario outcome outside PASS/FAIL/NOT_RUN/UNSUPPORTED is accepted.
//  5. One scenario appears twice and the second silently wins.
//  6. A FAIL without an error category, or an UNSUPPORTED without a reason,
//     is accepted, so the matrix cannot say why.
//  7. A report without a finish time, or one older than the row's last run,
//     overwrites newer results.
//  8. A report where nothing ran (no key: all NOT_RUN) changes the row.
//  9. A NOT_RUN scenario in an otherwise executed report erases what the
//     capability last showed.
// 10. A FAIL erases the capability's last pass time.
// 11. A PASS of one combination touches another combination's row.
// 12. The row is marked all-passed while a scenario failed or did not run.
// 13. A refused merge leaves a partly changed matrix, or a merge mutates the
//     caller's matrix.
// 14. A run narrowed to some scenarios marks the row all-passed: the report
//     does not say which scenarios the combination has.
// 16. A report with a missing or wrong operation crosses the route boundary.
// 17. An old report is accepted after schema migration.
// 15. A misconfigured process (FAIL with category config) overwrites real
//     results and moves the row's last run forward.

var (
	t0 = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	t1 = time.Date(2026, 10, 2, 9, 0, 0, 0, time.UTC)
	t2 = time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)
)

func baseMatrix() Matrix {
	return Matrix{Schema: SchemaVersion, Rows: []Row{
		{Combo: "openai-responses", Operation: "chat", Provider: "openai", API: "openai-responses"},
		{Combo: "deepseek-responses", Operation: "chat", Provider: "deepseek", API: "openai-responses",
			LastRunAt: &t0, Model: "deepseek-flash", SDK: "openai-go v0", AccountAlias: "old",
			Capabilities: []Capability{{ID: "text-stream", Model: "deepseek-flash", Outcome: Pass, LastRunAt: &t0, LastPassedAt: &t0}}},
	}}
}

func report(combo, provider, api string, at time.Time, scenarios ...ScenarioResult) Report {
	expected := make([]string, len(scenarios))
	for i, s := range scenarios {
		expected[i] = s.ID
	}
	return Report{Schema: SchemaVersion, Combo: combo, Operation: "chat", Provider: provider, API: api, Model: "m", SDK: "sdk v1",
		AccountAlias: "ci@region", StartedAt: at.Add(-time.Minute), FinishedAt: at, Expected: expected, Scenarios: scenarios}
}

func pass(id string) ScenarioResult { return ScenarioResult{ID: id, Model: "m", Outcome: Pass} }

func TestMergeRefusals(t *testing.T) {
	cases := []struct {
		name   string
		report Report
	}{
		{"1 unknown combination", report("mistral-chat", "mistral", "openai-completions", t1, pass("text-stream"))},
		{"2 provider differs from the row", report("deepseek-responses", "openai", "openai-responses", t1, pass("text-stream"))},
		{"2 API differs from the row", report("deepseek-responses", "deepseek", "openai-completions", t1, pass("text-stream"))},
		{"16 wrong operation", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Operation = "classifier"
			return r
		}()},
		{"16 missing operation", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Operation = ""
			return r
		}()},
		{"17 old report", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Schema = 1
			return r
		}()},
		{"3 schema", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Schema = SchemaVersion + 1
			return r
		}()},
		{"4 unknown outcome", report("openai-responses", "openai", "openai-responses", t1,
			ScenarioResult{ID: "text-stream", Outcome: "SKIP"})},
		{"5 duplicate scenario", report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"), pass("text-stream"))},
		{"6 FAIL without category", report("openai-responses", "openai", "openai-responses", t1,
			ScenarioResult{ID: "text-stream", Outcome: Fail})},
		{"6 UNSUPPORTED without reason", report("openai-responses", "openai", "openai-responses", t1,
			ScenarioResult{ID: "image", Outcome: Unsupported})},
		{"6 scenario without id", report("openai-responses", "openai", "openai-responses", t1, pass(""))},
		{"7 no finish time", report("openai-responses", "openai", "openai-responses", time.Time{}, pass("text-stream"))},
		{"7 older than the row's last run", report("deepseek-responses", "deepseek", "openai-responses", t0, pass("text-stream"))},
		{"14 no expected scenarios", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Expected = nil
			return r
		}()},
		{"14 a scenario the combination does not expect", func() Report {
			r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
			r.Expected = []string{"cancel"}
			return r
		}()},
		{"15 misconfigured process", report("openai-responses", "openai", "openai-responses", t1,
			ScenarioResult{ID: "text-stream", Outcome: Fail, ErrorCategory: CategoryConfig})},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := baseMatrix()
			got, err := Merge(m, c.report)
			if !errors.Is(err, ErrInvalidReport) {
				t.Fatalf("Merge error = %v, want ErrInvalidReport", err)
			}
			if !reflect.DeepEqual(got, Matrix{}) {
				t.Errorf("refused merge returned a matrix: %+v", got)
			}
			if !reflect.DeepEqual(m, baseMatrix()) {
				t.Errorf("13 refused merge changed the caller's matrix")
			}
		})
	}
}

func TestMergeNothingRanLeavesRowAlone(t *testing.T) {
	m := baseMatrix()
	r := report("deepseek-responses", "deepseek", "openai-responses", t1,
		ScenarioResult{ID: "text-stream", Model: "deepseek-flash", Outcome: NotRun, Note: "no key"})
	r.AccountAlias, r.SDK = "new", "openai-go v1"
	got, err := Merge(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, baseMatrix()) {
		t.Errorf("8 an all NOT_RUN report changed the matrix:\n got %+v", got.Rows[1])
	}
}

func TestMergeExecutedReport(t *testing.T) {
	m := baseMatrix()
	r := report("deepseek-responses", "deepseek", "openai-responses", t1,
		ScenarioResult{ID: "text-stream", Model: "deepseek-flash", Outcome: Fail, ErrorCategory: "upstream_error", Environment: true},
		ScenarioResult{ID: "image", Model: "deepseek-flash", Outcome: NotRun, Note: "budget"},
		ScenarioResult{ID: "cancel", Model: "deepseek-flash", Outcome: Pass},
		ScenarioResult{ID: "reasoning-history", Model: "deepseek-flash", Outcome: Unsupported, Note: "no reasoning"})
	r.Model, r.AccountAlias, r.SDK = "deepseek-flash", "ci-ds@cn", "openai-go v3.66.0"
	got, err := Merge(m, r)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.Rows[0], baseMatrix().Rows[0]) {
		t.Errorf("11 the deepseek report changed the openai-responses row: %+v", got.Rows[0])
	}
	row := got.Rows[1]
	if row.LastRunAt == nil || !row.LastRunAt.Equal(t1) || row.AccountAlias != "ci-ds@cn" || row.SDK != "openai-go v3.66.0" || row.Model != "deepseek-flash" {
		t.Errorf("row identity not updated from an executed report: %+v", row)
	}
	if row.AllPassedAt != nil {
		t.Errorf("12 row marked all-passed with a FAIL and a NOT_RUN")
	}
	caps := map[string]Capability{}
	for _, c := range row.Capabilities {
		caps[c.ID] = c
	}
	if c := caps["text-stream"]; c.Outcome != Fail || c.LastPassedAt == nil || !c.LastPassedAt.Equal(t0) || !c.LastRunAt.Equal(t1) || c.ErrorCategory != "upstream_error" {
		t.Errorf("10 FAIL: %+v", c)
	}
	if _, ok := caps["image"]; ok {
		t.Errorf("9 a NOT_RUN scenario without history created a capability: %+v", caps["image"])
	}
	if c := caps["cancel"]; c.Outcome != Pass || c.LastPassedAt == nil || !c.LastPassedAt.Equal(t1) {
		t.Errorf("PASS: %+v", c)
	}
	if c := caps["reasoning-history"]; c.Outcome != Unsupported || c.Note != "no reasoning" || c.LastPassedAt != nil {
		t.Errorf("UNSUPPORTED: %+v", c)
	}
	if !reflect.DeepEqual(m, baseMatrix()) {
		t.Errorf("13 merge mutated the caller's matrix")
	}

	// 9: a later run where text-stream did not run keeps its FAIL.
	r2 := report("deepseek-responses", "deepseek", "openai-responses", t2,
		ScenarioResult{ID: "text-stream", Model: "deepseek-flash", Outcome: NotRun, Note: "budget"},
		pass("cancel"))
	got2, err := Merge(got, r2)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range got2.Rows[1].Capabilities {
		if c.ID == "text-stream" && (c.Outcome != Fail || !c.LastRunAt.Equal(t1)) {
			t.Errorf("9 NOT_RUN erased the last result: %+v", c)
		}
	}
}

func TestMergeAllPassed(t *testing.T) {
	r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"), pass("cancel"),
		ScenarioResult{ID: "image", Model: "m", Outcome: Unsupported, Note: "text-only model"})
	got, err := Merge(baseMatrix(), r)
	if err != nil {
		t.Fatal(err)
	}
	row := got.Rows[0]
	if row.AllPassedAt == nil || !row.AllPassedAt.Equal(t1) {
		t.Errorf("all PASS/UNSUPPORTED did not mark the row all-passed: %+v", row)
	}
	if ids := []string{row.Capabilities[0].ID, row.Capabilities[1].ID, row.Capabilities[2].ID}; !reflect.DeepEqual(ids, []string{"text-stream", "cancel", "image"}) {
		t.Errorf("capabilities not in report order: %v", ids)
	}
	if !reflect.DeepEqual(got.Rows[1], baseMatrix().Rows[1]) {
		t.Errorf("11 the openai report changed the deepseek-responses row, which shares its adapter")
	}

	// 12: a later run with a failure keeps the earlier all-passed time but
	// a run that did not execute everything never sets it.
	r2 := report("openai-responses", "openai", "openai-responses", t2, pass("text-stream"),
		ScenarioResult{ID: "cancel", Model: "m", Outcome: Fail, ErrorCategory: "assertion"})
	got2, err := Merge(got, r2)
	if err != nil {
		t.Fatal(err)
	}
	if a := got2.Rows[0].AllPassedAt; a == nil || !a.Equal(t1) {
		t.Errorf("12 all-passed time after a failing run = %v, want the earlier %v", a, t1)
	}
}

func TestMergeNarrowedRunIsNotAllPassed(t *testing.T) {
	r := report("openai-responses", "openai", "openai-responses", t1, pass("text-stream"))
	r.Expected = []string{"text-stream", "cancel"}
	got, err := Merge(baseMatrix(), r)
	if err != nil {
		t.Fatal(err)
	}
	if got.Rows[0].AllPassedAt != nil {
		t.Errorf("14 a run without the expected cancel scenario marked the row all-passed")
	}
	if c := got.Rows[0].Capabilities; len(c) != 1 || c[0].Outcome != Pass {
		t.Errorf("the scenario that ran was not merged: %+v", c)
	}
}

package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestJevFormalMatrixHistory(t *testing.T) {
	ev := run.Case(t, "UC13-model-specific-matrix-history")
	before, err := supportmatrix.LoadMatrix("testdata/typesafe/formal-matrix-before-1.14.json")
	if err != nil {
		t.Fatal(err)
	}
	current, err := supportmatrix.LoadMatrix("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, identity := range []struct{ combo, model string }{
		{"typesafe-classifier", "jev-1.13.0"}, {"typesafe-classifier-jev-1.14.0", "jev-1.14.0"},
	} {
		matches := 0
		for _, row := range current.Rows {
			if row.Combo == identity.combo && row.Model == identity.model && row.ValidIdentity() {
				matches++
			}
		}
		if !ev.Check("independent fixed model identity exists", matches == 1, "wrong identity for %s", identity.combo) {
			return
		}
	}
	// This snapshot comes from the initial admission commit 527f633. Migration
	// preserves the old rows; later smoke runs legitimately advance their own row.
	migrated, err := supportmatrix.LoadMatrix("testdata/typesafe/formal-matrix-after-1.14.json")
	if err != nil {
		t.Fatal(err)
	}
	i := slices.IndexFunc(migrated.Rows, func(r supportmatrix.Row) bool { return r.Combo == "typesafe-classifier-jev-1.14.0" })
	if !ev.Check("migration adds independent 1.14 identity", i >= 0 && migrated.Rows[i].ValidIdentity(), "new model row absent") {
		return
	}
	migrated.Rows = slices.Delete(slices.Clone(migrated.Rows), i, i+1)
	ev.Check("all migration historical rows states and timestamps preserved", reflect.DeepEqual(migrated, before), "historical support rewritten")
	raw, err := os.ReadFile("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	ev.Record("matrix", v)
	for _, name := range []string{"before", "after"} {
		data, err := os.ReadFile("testdata/typesafe/formal-matrix-" + name + "-1.14.json")
		if err != nil {
			t.Fatal(err)
		}
		ev.Fixture("migration-"+name+".json", data)
	}

	// Exercise a later 1.13 run through the public file command. The retained
	// 1.14 row and every unrelated field must survive, including timestamps.
	dir := t.TempDir()
	matrixPath, reportPath := filepath.Join(dir, "matrix.json"), filepath.Join(dir, "report.json")
	if err := os.WriteFile(matrixPath, raw, 0600); err != nil {
		t.Fatal(err)
	}
	i = slices.IndexFunc(current.Rows, func(r supportmatrix.Row) bool { return r.Combo == "typesafe-classifier" })
	at := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	if current.Rows[i].LastRunAt != nil {
		at = current.Rows[i].LastRunAt.Add(time.Second)
	}
	report := supportmatrix.Report{Schema: 3, Combo: "typesafe-classifier", Operation: "classifier", Provider: "typesafe",
		API: "typesafe-system-one", Model: "jev-1.13.0", Endpoint: "https://api.typesafe.ai/v1", SDK: "synthetic",
		AccountAlias: "matrix-fixture@offline", FinishedAt: at, Expected: []string{"mixed-questions", "single-choice", "context-422"},
		Budget: supportmatrix.Budget{MaxCalls: 6, MaxQuestions: 10, MaxStateBytes: 131072, MaxRetries: 1,
			CallsUsed: 3, HTTPAttempts: 3, QuestionsUsed: 3, StateBytesSubmitted: 120}}
	for _, id := range report.Expected {
		report.Scenarios = append(report.Scenarios, supportmatrix.ScenarioResult{ID: id, Model: report.Model,
			Outcome: supportmatrix.Pass, Calls: 1, Attempts: 1, Questions: 1, ProviderRequestIDs: []string{"req-synthetic"}})
	}
	if err := os.WriteFile(reportPath, mustMarshal(t, report), 0600); err != nil {
		t.Fatal(err)
	}
	ev.Record("synthetic-report", report)
	invoke := func() ([]byte, error) {
		cmd := exec.CommandContext(ctxFor(t), "go", "run", "./ai/live/cmd/supportmatrix", "-matrix", matrixPath, reportPath)
		cmd.Dir = "../.."
		return cmd.CombinedOutput()
	}
	output, err := invoke()
	ev.Record("merge-command", map[string]any{"output": string(output), "refused": err != nil})
	if !ev.Check("later own-model report accepted", err == nil, "exit=%v", err) {
		return
	}
	after, err := supportmatrix.LoadMatrix(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("matrix-after-own-run", after)
	ev.Check("same row set retained", len(after.Rows) == len(current.Rows), "row count changed")
	for _, row := range current.Rows {
		j := slices.IndexFunc(after.Rows, func(r supportmatrix.Row) bool { return r.Combo == row.Combo })
		if !ev.Check("route retained", j >= 0, "missing %s", row.Combo) {
			continue
		}
		got := after.Rows[j]
		if row.Combo == report.Combo {
			ev.Check("own timestamps advance", got.LastRunAt != nil && got.AllPassedAt != nil &&
				got.LastRunAt.Equal(at) && got.AllPassedAt.Equal(at), "own report not consumed")
		} else {
			ev.Check("other model history preserved exactly", reflect.DeepEqual(got, row), "%s rewritten", row.Combo)
		}
	}
	accepted, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	output, refused := invoke()
	ev.Record("stale-command", map[string]any{"output": string(output), "refused": refused != nil})
	stale, err := os.ReadFile(matrixPath)
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("stale repeat refuses without write", refused != nil && bytes.Equal(accepted, stale), "stale report changed matrix")
	unchanged, err := os.ReadFile("../live/support-matrix.json")
	if err != nil {
		t.Fatal(err)
	}
	ev.Check("published matrix untouched by offline regression", bytes.Equal(raw, unchanged), "published history changed")
}

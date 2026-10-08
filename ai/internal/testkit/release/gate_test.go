package release

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// The release gate is tested in isolation (AGENTS.md 测试原则): these are the
// ways it could wrongly let a release through, or misreport one, enumerated
// before the implementation.
//
//  1. A failed go test, -race or go vet run is not a failed gate.
//  2. A required command that was never run counts as passed.
//  3. An empty offline bundle passes the P0 gate.
//  4. A FAIL or NOT_RUN offline case passes the P0 gate.
//  5. A P0 scenario without any case, or with only UNSUPPORTED cases,
//     passes.
//  6. A bundle without differential cases passes the differential gate
//     ("differential not run" looks like "no differences").
//  7. A differential case that did not run, or reports pending findings
//     while its test passed, passes.
//  8. A pending ledger decision passes.
//  9. A protocol required in the differential has no differential case and
//     passes.
// 10. An extension route (DeepSeek × Responses) is not registered in the
//     ledger and passes, i.e. its absence from the differential is
//     unexplained.
// 11. A support matrix missing a combination, a combination never fully
//     passed, or one fully passed earlier but failing now, passes the live
//     gate.
// 12. UNSUPPORTED live capabilities are hidden instead of listed.
// 13. An audit finding, an audit error, or no audit at all passes.
// 14. A traceability item that is only mapped (no evidence) is reported
//     PASS; one with a failing or not-run case is PASS.
// 15. A traceability map naming an unknown scenario, a bad pattern or a
//     duplicate item is accepted.
// 16. A stale or missing snapshot check passes.
// 17. Offline, differential and live results are mixed in one count.
// 18. The rendered report hides a failed gate or merges the three result
//     sections.
// 19. A combination's complete live pass was made with an SDK version other
//     than the one go.mod pins now, and still passes.
// 20. A combination passes live without any audited live evidence bundle
//     of its own.
// 21. A differential record lacking the spec's required fields (case_id,
//     pi_commit, sdk_versions, model_catalog_hash, request and frame
//     hashes) passes.

const testTrace = `{
  "schema": 1,
  "scenarios": {
    "E01": "(^|-)E01-",
    "E07": "(^|-)E07-",
    "P01": "^(PIDIFF-)?P01-",
    "P05": "^P05-",
    "D2": "(^|-)E07-snapshot-",
    "PIDIFF": "^PIDIFF-"
  },
  "p0": ["E01", "E07", "P01", "P05", "D2"],
  "differential": {
    "required": ["P01"],
    "routes": [{"scenario": "P05", "provider": "deepseek", "api": "openai-responses"}]
  },
  "liveCombos": ["openai-responses", "deepseek-responses"],
  "items": [
    {"id": "T01", "source": "S T01", "requirement": "explicit tenant", "scenarios": ["E07"], "cases": "missing-tenant"},
    {"id": "C01", "source": "V C01", "requirement": "lifecycle", "scenarios": ["E01"]},
    {"id": "D2", "source": "ADR-0003", "requirement": "snapshot", "scenarios": ["D2"]},
    {"id": "H2", "source": "D H2", "requirement": "pi differential", "scenarios": ["PIDIFF"]},
    {"id": "V6", "source": "V§6", "requirement": "live smoke", "scenarios": ["LIVE"]}
  ]
}`

func mustTrace(t *testing.T, data string) TraceMap {
	t.Helper()
	m, err := ParseTraceMap([]byte(data))
	if err != nil {
		t.Fatalf("ParseTraceMap: %v", err)
	}
	return m
}

func passedAt() *time.Time {
	at := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	return &at
}

func goodInputs(t *testing.T) Inputs {
	t.Helper()
	row := func(combo, provider, api string) supportmatrix.Row {
		return supportmatrix.Row{Combo: combo, Operation: "chat", Provider: provider, API: api, Model: "m", SDK: "sdk", AccountAlias: "ci@x",
			LastRunAt: passedAt(), AllPassedAt: passedAt(), Capabilities: []supportmatrix.Capability{
				{ID: "text-stream", Model: "m", Outcome: supportmatrix.Pass, LastRunAt: passedAt(), LastPassedAt: passedAt()},
				{ID: "image", Model: "m", Outcome: supportmatrix.Unsupported, Note: "model takes no images", LastRunAt: passedAt()},
			}}
	}
	return Inputs{
		Commands: []Command{
			{ID: CmdTest, Line: "go test ./...", Passed: true},
			{ID: CmdRace, Line: "go test -race ./...", Passed: true},
			{ID: CmdVet, Line: "go vet ./...", Passed: true},
			{ID: CmdVetLive, Line: "go vet -tags live ./...", Passed: true},
		},
		Offline: Bundle{Dir: "/b", Cases: []Case{
			{ID: "E01-text-stream", Status: Pass},
			{ID: "P01-E01-text-complete", Status: Pass},
			{ID: "P01-E07-reject-missing-tenant-id-stream", Status: Pass},
			{ID: "E07-snapshot-rotation-between-reads-stream", Status: Pass},
			{ID: "P05-E01-text-stream", Status: Pass},
			{ID: "PIDIFF-P01-E01-text-stream", Status: Pass},
		}, Diffs: []DiffCase{{ID: "PIDIFF-P01-E01-text-stream", Status: Pass, Pending: 0}}},
		Ledger: Ledger{Extension: 3, Fixed: 1, Routes: []Route{{Provider: "deepseek", API: "openai-responses"}}},
		Matrix: supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{
			row("openai-responses", "openai", "openai-responses"), row("deepseek-responses", "deepseek", "openai-responses"),
		}},
		Trace:     mustTrace(t, testTrace),
		Audits:    []AuditResult{{Bundle: "/b"}, {Bundle: "/live-1", Combo: "openai-responses"}, {Bundle: "/live-2", Combo: "deepseek-responses"}},
		Modules:   map[string]string{"github.com/openai/openai-go/v3": "v3.66.0"},
		Snapshots: []Snapshot{{Name: "catalog", Current: true}},
	}
}

func gate(t *testing.T, r Report, id GateID) Gate {
	t.Helper()
	i := slices.IndexFunc(r.Gates, func(g Gate) bool { return g.ID == id })
	if i < 0 {
		t.Fatalf("gate %s missing from report", id)
	}
	return r.Gates[i]
}

func expectFail(t *testing.T, r Report, id GateID, problem string) {
	t.Helper()
	g := gate(t, r, id)
	if g.Pass || r.Pass {
		t.Fatalf("gate %s passed (overall %v): %+v", id, r.Pass, g)
	}
	if !slices.ContainsFunc(g.Problems, func(p string) bool { return strings.Contains(p, problem) }) {
		t.Fatalf("gate %s problems %q do not mention %q", id, g.Problems, problem)
	}
}

func TestAllGreen(t *testing.T) {
	r := Evaluate(goodInputs(t))
	for _, g := range r.Gates {
		if !g.Pass {
			t.Errorf("gate %s failed: %v", g.ID, g.Problems)
		}
	}
	if !r.Pass {
		t.Fatal("overall gate failed on good inputs")
	}
}

func TestCommands(t *testing.T) { // 1, 2
	for _, c := range []struct {
		id   string
		gate GateID
	}{{CmdTest, GateTest}, {CmdRace, GateRace}, {CmdVet, GateVet}, {CmdVetLive, GateVet}} {
		in := goodInputs(t)
		in.Commands[slices.IndexFunc(in.Commands, func(x Command) bool { return x.ID == c.id })].Passed = false
		expectFail(t, Evaluate(in), c.gate, c.id)

		in = goodInputs(t)
		in.Commands = slices.DeleteFunc(in.Commands, func(x Command) bool { return x.ID == c.id })
		expectFail(t, Evaluate(in), c.gate, "not run")
	}
}

func TestOfflineP0(t *testing.T) { // 3, 4, 5
	in := goodInputs(t)
	in.Offline.Cases = nil
	expectFail(t, Evaluate(in), GateP0, "no offline cases")

	in = goodInputs(t)
	in.Offline.Cases[0].Status = Fail
	expectFail(t, Evaluate(in), GateP0, "E01-text-stream")

	in = goodInputs(t)
	in.Offline.Cases[0].Status = NotRun
	expectFail(t, Evaluate(in), GateP0, "E01-text-stream")

	in = goodInputs(t)
	in.Offline.Cases = slices.DeleteFunc(in.Offline.Cases, func(c Case) bool { return strings.HasPrefix(c.ID, "P05-") })
	expectFail(t, Evaluate(in), GateP0, "P05")

	in = goodInputs(t)
	in.Offline.Cases[4].Status = Unsupported // the only P05 case
	expectFail(t, Evaluate(in), GateP0, "P05")
}

func TestDifferential(t *testing.T) { // 6, 7, 8, 9, 10
	in := goodInputs(t)
	in.Offline.Cases = slices.DeleteFunc(in.Offline.Cases, func(c Case) bool { return strings.HasPrefix(c.ID, "PIDIFF-") })
	in.Offline.Diffs = nil
	expectFail(t, Evaluate(in), GateDifferential, "not run")

	in = goodInputs(t)
	in.Offline.Cases[5].Status = NotRun
	in.Offline.Diffs = nil
	expectFail(t, Evaluate(in), GateDifferential, "PIDIFF-P01-E01-text-stream")

	in = goodInputs(t)
	in.Offline.Diffs[0].Pending = 2
	expectFail(t, Evaluate(in), GateDifferential, "pending")

	in = goodInputs(t)
	in.Offline.Diffs = nil // the case ran, but its verdict is missing
	expectFail(t, Evaluate(in), GateDifferential, "verdict")

	in = goodInputs(t)
	in.Ledger.Pending = 1
	expectFail(t, Evaluate(in), GateDifferential, "ledger")

	in = goodInputs(t)
	in.Trace.Differential.Required = append(in.Trace.Differential.Required, "E07")
	expectFail(t, Evaluate(in), GateDifferential, "E07")

	in = goodInputs(t)
	in.Ledger.Routes = nil
	expectFail(t, Evaluate(in), GateDifferential, "deepseek")
}

func TestLive(t *testing.T) { // 11, 12
	in := goodInputs(t)
	in.Matrix.Rows = in.Matrix.Rows[:1]
	expectFail(t, Evaluate(in), GateLive, "deepseek-responses")

	in = goodInputs(t)
	in.Matrix.Rows[1].AllPassedAt, in.Matrix.Rows[1].LastRunAt, in.Matrix.Rows[1].Capabilities = nil, nil, nil
	r := Evaluate(in)
	expectFail(t, r, GateLive, "deepseek-responses")
	if row := r.Live.Rows[1]; row.Status != NotRun {
		t.Fatalf("never-run row status %s, want NOT_RUN", row.Status)
	}

	in = goodInputs(t)
	in.Matrix.Rows[0].Capabilities[0].Outcome = supportmatrix.Fail
	in.Matrix.Rows[0].Capabilities[0].ErrorCategory = "assertion"
	r = Evaluate(in)
	expectFail(t, r, GateLive, "text-stream")
	if r.Live.Rows[0].Status != Fail {
		t.Fatalf("row failing now has status %s", r.Live.Rows[0].Status)
	}

	r = Evaluate(goodInputs(t))
	if got := r.Live.Rows[0].Unsupported; len(got) != 1 || !strings.Contains(got[0], "image") || !strings.Contains(got[0], "model takes no images") {
		t.Fatalf("UNSUPPORTED capability not listed with its note: %q", got)
	}
}

func TestLiveStaleOrUnaudited(t *testing.T) { // 19, 20
	in := goodInputs(t)
	in.Matrix.Rows[0].SDK = "github.com/openai/openai-go/v3 v3.65.0"
	r := Evaluate(in)
	expectFail(t, r, GateLive, "v3.66.0")
	if r.Live.Rows[0].Status != NotRun {
		t.Fatalf("stale pass has status %s, want NOT_RUN", r.Live.Rows[0].Status)
	}

	in = goodInputs(t)
	in.Matrix.Rows[0].SDK = "github.com/openai/openai-go/v3 v3.66.0 (note)"
	if r := Evaluate(in); !gate(t, r, GateLive).Pass {
		t.Fatalf("current SDK refused: %v", gate(t, r, GateLive).Problems)
	}

	in = goodInputs(t)
	in.Audits = in.Audits[:2] // deepseek-responses' live bundle not audited
	expectFail(t, Evaluate(in), GateLive, "deepseek-responses")
}

func TestDifferentialRecordFields(t *testing.T) { // 21
	in := goodInputs(t)
	in.Offline.Diffs[0].Missing = []string{"pi_commit"}
	expectFail(t, Evaluate(in), GateDifferential, "pi_commit")
}

func TestAudit(t *testing.T) { // 13
	in := goodInputs(t)
	in.Audits[0].Findings = []audit.Finding{{File: "c/x.json", Rule: audit.RuleKeyShaped, Detail: "sk-… (20 chars)"}}
	expectFail(t, Evaluate(in), GateAudit, "c/x.json")

	in = goodInputs(t)
	in.Audits[0].Err = "not an evidence bundle"
	expectFail(t, Evaluate(in), GateAudit, "not an evidence bundle")

	in = goodInputs(t)
	in.Audits = in.Audits[1:] // only live bundles: the offline one was not audited
	expectFail(t, Evaluate(in), GateAudit, "offline")
}

func TestTraceability(t *testing.T) { // 14
	in := goodInputs(t)
	in.Offline.Cases = slices.DeleteFunc(in.Offline.Cases, func(c Case) bool { return strings.Contains(c.ID, "missing-tenant") })
	r := Evaluate(in)
	expectFail(t, r, GateTrace, "T01")
	if s := traceStatus(t, r, "T01"); s != NoEvidence {
		t.Fatalf("mapped-only item has status %s, want NO_EVIDENCE", s)
	}

	in = goodInputs(t)
	in.Offline.Cases[0].Status = Fail
	if s := traceStatus(t, Evaluate(in), "C01"); s != Fail {
		t.Fatalf("item with a failing case has status %s", s)
	}

	in = goodInputs(t)
	in.Matrix.Rows[1].AllPassedAt, in.Matrix.Rows[1].LastRunAt, in.Matrix.Rows[1].Capabilities = nil, nil, nil
	r = Evaluate(in)
	if s := traceStatus(t, r, "V6"); s != NotRun {
		t.Fatalf("live item with a not-run combination has status %s", s)
	}
	expectFail(t, r, GateTrace, "V6")

	r = Evaluate(goodInputs(t))
	for _, item := range r.Trace {
		if item.Status != Pass {
			t.Errorf("item %s status %s on good inputs", item.ID, item.Status)
		}
	}
	if tr := r.Trace[slices.IndexFunc(r.Trace, func(x TraceResult) bool { return x.ID == "C01" })]; tr.Offline.Pass != 3 || tr.Differential.Pass != 1 || len(tr.Evidence) != 4 {
		t.Fatalf("C01 evidence not traced to its cases: %+v", tr)
	}
}

func traceStatus(t *testing.T, r Report, id string) Status {
	t.Helper()
	i := slices.IndexFunc(r.Trace, func(x TraceResult) bool { return x.ID == id })
	if i < 0 {
		t.Fatalf("trace item %s missing", id)
	}
	return r.Trace[i].Status
}

func TestTraceMapValidation(t *testing.T) { // 15
	for name, data := range map[string]string{
		"unknown scenario": strings.Replace(testTrace, `"scenarios": ["E01"]}`, `"scenarios": ["E99"]}`, 1),
		"bad pattern":      strings.Replace(testTrace, `"(^|-)E01-"`, `"(E01"`, 1),
		"duplicate item":   strings.Replace(testTrace, `"id": "C01"`, `"id": "T01"`, 1),
		"bad item pattern": strings.Replace(testTrace, `"cases": "missing-tenant"`, `"cases": "(x"`, 1),
		"no scenarios":     strings.Replace(testTrace, `"scenarios": ["E01"]}`, `"scenarios": []}`, 1),
		"unknown p0":       strings.Replace(testTrace, `"p0": ["E01"`, `"p0": ["E42"`, 1),
		"unknown field":    strings.Replace(testTrace, `"schema": 1,`, `"schema": 1, "extra": true,`, 1),
	} {
		if _, err := ParseTraceMap([]byte(data)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

func TestSnapshots(t *testing.T) { // 16
	in := goodInputs(t)
	in.Snapshots[0] = Snapshot{Name: "catalog", Current: false, Detail: "version 2026-10-02.6 differs"}
	expectFail(t, Evaluate(in), GateSnapshots, "catalog")

	in = goodInputs(t)
	in.Snapshots = nil
	expectFail(t, Evaluate(in), GateSnapshots, "no snapshot")
}

func TestSeparateSections(t *testing.T) { // 17
	in := goodInputs(t)
	in.Offline.Cases = append(in.Offline.Cases, Case{ID: "LIVE-P01-text-stream", Status: NotRun})
	r := Evaluate(in)
	if r.Offline.Total() != 5 || r.Differential.Total() != 1 {
		t.Fatalf("offline %+v differential %+v: sections mixed", r.Offline.Counts, r.Differential.Counts)
	}
	if !r.Pass {
		t.Fatalf("a live case in an offline bundle failed the offline gate: %+v", r.Gates)
	}
}

func TestMarkdown(t *testing.T) { // 18
	in := goodInputs(t)
	in.Matrix.Rows[1].AllPassedAt, in.Matrix.Rows[1].LastRunAt, in.Matrix.Rows[1].Capabilities = nil, nil, nil
	md := Evaluate(in).Markdown()
	for _, want := range []string{"FAIL", "## Offline", "## Differential", "## Live", "## Traceability", "deepseek-responses", "NOT_RUN"} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q", want)
		}
	}
	if strings.Index(md, "## Offline") > strings.Index(md, "## Differential") || strings.Index(md, "## Differential") > strings.Index(md, "## Live") {
		t.Error("sections out of order")
	}
}

// Missing/mismatched operation or an old matrix must never reuse a prior pass.
func TestLiveIdentity(t *testing.T) {
	for _, name := range []string{"wrong-operation", "missing-operation", "old-schema"} {
		t.Run(name, func(t *testing.T) {
			in := goodInputs(t)
			switch name {
			case "wrong-operation":
				in.Matrix.Rows[0].Operation = "classifier"
			case "missing-operation":
				in.Matrix.Rows[0].Operation = ""
			case "old-schema":
				in.Matrix.Schema = 1
			}
			expectFail(t, Evaluate(in), GateLive, "identity")
		})
	}
}

// Package release is barness-ai's release gate (spec Testing Decisions §6,
// issue 24): from the results of go test, -race and go vet, the offline
// evidence bundle with its frozen pi differential, the difference ledger,
// the support matrix, the redaction audit and the snapshot checks, it
// decides whether a release may ship and reports offline, differential and
// live results separately.
//
// Evaluate is a pure function of its Inputs, so every way it could wrongly
// pass is tested in isolation (gate_test.go); the command
// ai/release/cmd/releasegate gathers the inputs.
package release

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// Status is an evidence or item outcome.
type Status string

const (
	Pass        Status = "PASS"
	Fail        Status = "FAIL"
	NotRun      Status = "NOT_RUN"
	Unsupported Status = "UNSUPPORTED"
	// NoEvidence is a traceability item that is mapped but has no passing
	// evidence: never a pass.
	NoEvidence Status = "NO_EVIDENCE"
)

// Command IDs the gate requires.
const (
	CmdTest    = "go-test"
	CmdRace    = "go-test-race"
	CmdVet     = "go-vet"
	CmdVetLive = "go-vet-live"
)

// GateID names one gate.
type GateID string

const (
	GateTest         GateID = "go-test"
	GateRace         GateID = "go-test-race"
	GateVet          GateID = "go-vet"
	GateP0           GateID = "p0-offline"
	GateDifferential GateID = "differential"
	GateLive         GateID = "live"
	GateAudit        GateID = "redaction-audit"
	GateTrace        GateID = "traceability"
	GateSnapshots    GateID = "snapshots"
)

// Case is one evidence case as the bundle manifest records it.
type Case struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
}

// DiffCase is one differential case's verdict (pidiff.json).
type DiffCase struct {
	ID      string `json:"id"`
	Status  Status `json:"status"`
	Pending int    `json:"pending"`
	// Missing names the record fields spec Testing Decisions §6 requires
	// that pidiff.json lacks.
	Missing []string `json:"missing,omitempty"`
}

// Bundle is an offline evidence bundle.
type Bundle struct {
	Dir    string     `json:"dir"`
	Commit string     `json:"commit"`
	Cases  []Case     `json:"-"`
	Diffs  []DiffCase `json:"-"`
}

// Command is one gate command's result.
type Command struct {
	ID       string        `json:"id"`
	Line     string        `json:"line"`
	Passed   bool          `json:"passed"`
	Duration time.Duration `json:"duration"`
	// Tail is the end of the command's output, for a failed run.
	Tail string `json:"tail,omitempty"`
}

// Ledger summarizes the difference ledger.
type Ledger struct {
	Fixed     int     `json:"fixed"`
	Extension int     `json:"extension"`
	Pending   int     `json:"pending"`
	Routes    []Route `json:"routes"`
}

// Route is a provider × api registered in the ledger as an extension route.
type Route struct {
	Provider string `json:"provider"`
	API      string `json:"api"`
}

// AuditResult is the redaction audit of one bundle.
type AuditResult struct {
	Bundle string `json:"bundle"`
	// Combo is the live combination of a live bundle (its
	// live-report.json); empty for offline and race bundles.
	Combo    string          `json:"combo,omitempty"`
	Findings []audit.Finding `json:"findings"`
	Err      string          `json:"error,omitempty"`
}

// Snapshot is whether a checked-in snapshot matches the code.
type Snapshot struct {
	Name    string `json:"name"`
	Current bool   `json:"current"`
	Detail  string `json:"detail,omitempty"`
}

// Inputs is everything the gate decides on.
type Inputs struct {
	Commands  []Command
	Offline   Bundle
	Ledger    Ledger
	Matrix    supportmatrix.Matrix
	Trace     TraceMap
	Audits    []AuditResult
	Snapshots []Snapshot
	// Modules are the module versions go.mod pins now; a live pass made
	// with another SDK version is stale.
	Modules map[string]string
}

// Gate is one gate's outcome.
type Gate struct {
	ID       GateID   `json:"id"`
	Name     string   `json:"name"`
	Pass     bool     `json:"pass"`
	Problems []string `json:"problems,omitempty"`
}

// Counts tallies outcomes.
type Counts struct {
	Pass        int `json:"pass"`
	Fail        int `json:"fail"`
	NotRun      int `json:"notRun"`
	Unsupported int `json:"unsupported"`
}

func (c *Counts) add(s Status) {
	switch s {
	case Pass:
		c.Pass++
	case Unsupported:
		c.Unsupported++
	case NotRun:
		c.NotRun++
	default:
		c.Fail++
	}
}

func (c Counts) plus(o Counts) Counts {
	return Counts{c.Pass + o.Pass, c.Fail + o.Fail, c.NotRun + o.NotRun, c.Unsupported + o.Unsupported}
}

// Total is the number of outcomes counted.
func (c Counts) Total() int { return c.Pass + c.Fail + c.NotRun + c.Unsupported }

// OfflineSection is the offline E2E result, differential cases excluded.
type OfflineSection struct {
	Counts
	Scenarios map[string]Counts `json:"scenarios"`
	Failing   []string          `json:"failing,omitempty"`
}

// DifferentialSection is the frozen pi differential result.
type DifferentialSection struct {
	Counts
	Pending    int               `json:"pending"`
	ByScenario map[string]Counts `json:"byScenario"`
	Ledger     Ledger            `json:"ledger"`
	Failing    []string          `json:"failing,omitempty"`
}

// LiveSection is the real-API result, from the support matrix.
type LiveSection struct {
	Rows []LiveRow `json:"rows"`
}

// LiveRow is one required combination.
type LiveRow struct {
	Combo        string     `json:"combo"`
	Operation    string     `json:"operation"`
	Provider     string     `json:"provider,omitempty"`
	API          string     `json:"api,omitempty"`
	Model        string     `json:"model,omitempty"`
	SDK          string     `json:"sdk,omitempty"`
	AccountAlias string     `json:"accountAlias,omitempty"`
	Status       Status     `json:"status"`
	LastRunAt    *time.Time `json:"lastRunAt,omitempty"`
	AllPassedAt  *time.Time `json:"allPassedAt,omitempty"`
	// Unsupported lists "capability: note" for each UNSUPPORTED capability.
	Unsupported []string `json:"unsupported,omitempty"`
	Failing     []string `json:"failing,omitempty"`
}

// Report is the gate's decision.
type Report struct {
	Pass         bool                `json:"pass"`
	Commit       string              `json:"commit"`
	Bundle       string              `json:"bundle"`
	Gates        []Gate              `json:"gates"`
	Commands     []Command           `json:"commands"`
	Offline      OfflineSection      `json:"offline"`
	Differential DifferentialSection `json:"differential"`
	Live         LiveSection         `json:"live"`
	Trace        []TraceResult       `json:"traceability"`
	Audits       []AuditResult       `json:"audits"`
	Snapshots    []Snapshot          `json:"snapshots"`
}

type caseKind int

const (
	kindOffline caseKind = iota
	kindDifferential
	kindLive
)

func kindOf(id string) caseKind {
	switch {
	case strings.HasPrefix(id, "PIDIFF-"):
		return kindDifferential
	case strings.HasPrefix(id, "LIVE-"):
		return kindLive
	}
	return kindOffline
}

// Evaluate decides the release.
func Evaluate(in Inputs) Report {
	r := Report{Commit: in.Offline.Commit, Bundle: in.Offline.Dir, Commands: in.Commands, Audits: in.Audits, Snapshots: in.Snapshots}
	r.Gates = append(r.Gates,
		commandGate(in, GateTest, "go test ./... (offline, differential on)", CmdTest),
		commandGate(in, GateRace, "go test -race ./...", CmdRace),
		commandGate(in, GateVet, "go vet ./... (with and without the live tag)", CmdVet, CmdVetLive),
	)
	var g Gate
	r.Offline, g = offline(in)
	r.Gates = append(r.Gates, g)
	r.Differential, g = differential(in)
	r.Gates = append(r.Gates, g)
	r.Live, g = live(in)
	r.Gates = append(r.Gates, g)
	r.Gates = append(r.Gates, auditGate(in))
	r.Trace = traceItems(in.Trace, in.Offline.Cases, r.Live.Rows)
	g = Gate{ID: GateTrace, Name: "traceability: research item → scenario → evidence"}
	for _, t := range r.Trace {
		if t.Status != Pass {
			g.Problems = append(g.Problems, fmt.Sprintf("%s (%s) is %s", t.ID, t.Source, t.Status))
		}
	}
	r.Gates = append(r.Gates, settle(g))
	r.Gates = append(r.Gates, snapshotGate(in))
	r.Pass = !slices.ContainsFunc(r.Gates, func(g Gate) bool { return !g.Pass })
	return r
}

func settle(g Gate) Gate {
	g.Pass = len(g.Problems) == 0
	return g
}

func commandGate(in Inputs, id GateID, name string, cmds ...string) Gate {
	g := Gate{ID: id, Name: name}
	for _, want := range cmds {
		i := slices.IndexFunc(in.Commands, func(c Command) bool { return c.ID == want })
		switch {
		case i < 0:
			g.Problems = append(g.Problems, want+": not run")
		case !in.Commands[i].Passed:
			g.Problems = append(g.Problems, want+": failed ("+in.Commands[i].Line+")")
		}
	}
	return settle(g)
}

func offline(in Inputs) (OfflineSection, Gate) {
	s := OfflineSection{Scenarios: map[string]Counts{}}
	g := Gate{ID: GateP0, Name: "every offline P0 case passes"}
	for _, c := range in.Offline.Cases {
		if kindOf(c.ID) != kindOffline {
			continue
		}
		s.add(c.Status)
		if c.Status != Pass && c.Status != Unsupported {
			s.Failing = append(s.Failing, c.ID+"="+string(c.Status))
		}
		for _, sc := range in.Trace.P0 {
			if in.Trace.matches(sc, c.ID) {
				n := s.Scenarios[sc]
				n.add(c.Status)
				s.Scenarios[sc] = n
			}
		}
	}
	if s.Total() == 0 {
		g.Problems = append(g.Problems, "no offline cases in the bundle")
	}
	for _, f := range s.Failing {
		g.Problems = append(g.Problems, "case "+f)
	}
	for _, sc := range in.Trace.P0 {
		if s.Scenarios[sc].Pass == 0 {
			g.Problems = append(g.Problems, "scenario "+sc+" has no passing case")
		}
	}
	return s, settle(g)
}

func differential(in Inputs) (DifferentialSection, Gate) {
	s := DifferentialSection{ByScenario: map[string]Counts{}, Ledger: in.Ledger}
	g := Gate{ID: GateDifferential, Name: "frozen pi differential: no pending difference"}
	for _, c := range in.Offline.Cases {
		if kindOf(c.ID) != kindDifferential {
			continue
		}
		s.add(c.Status)
		if c.Status != Pass {
			s.Failing = append(s.Failing, c.ID+"="+string(c.Status))
		}
		i := slices.IndexFunc(in.Offline.Diffs, func(d DiffCase) bool { return d.ID == c.ID })
		switch {
		case c.Status == NotRun:
		case i < 0:
			g.Problems = append(g.Problems, "case "+c.ID+": no differential verdict (pidiff.json)")
		default:
			d := in.Offline.Diffs[i]
			if d.Pending > 0 {
				s.Pending += d.Pending
				g.Problems = append(g.Problems, fmt.Sprintf("case %s: %d pending finding(s)", c.ID, d.Pending))
			}
			if len(d.Missing) > 0 {
				g.Problems = append(g.Problems, fmt.Sprintf("case %s: differential record lacks %s", c.ID, strings.Join(d.Missing, ", ")))
			}
		}
		for _, sc := range in.Trace.Differential.Required {
			if in.Trace.matches(sc, c.ID) {
				n := s.ByScenario[sc]
				n.add(c.Status)
				s.ByScenario[sc] = n
			}
		}
	}
	if s.Total() == 0 {
		g.Problems = append(g.Problems, "differential not run: no PIDIFF case in the bundle (set BARNESS_AI_PIDIFF=1)")
	}
	for _, f := range s.Failing {
		g.Problems = append(g.Problems, "case "+f)
	}
	if in.Ledger.Pending > 0 {
		g.Problems = append(g.Problems, fmt.Sprintf("ledger holds %d pending decision(s)", in.Ledger.Pending))
	}
	for _, sc := range in.Trace.Differential.Required {
		if s.ByScenario[sc].Pass == 0 {
			g.Problems = append(g.Problems, "scenario "+sc+" has no passing differential case")
		}
	}
	for _, rt := range in.Trace.Differential.Routes {
		if !slices.Contains(in.Ledger.Routes, Route{Provider: rt.Provider, API: rt.API}) {
			g.Problems = append(g.Problems, fmt.Sprintf("route %s × %s (%s) is outside the differential but not registered in the ledger", rt.Provider, rt.API, rt.Scenario))
		}
	}
	return s, settle(g)
}

func live(in Inputs) (LiveSection, Gate) {
	var s LiveSection
	g := Gate{ID: GateLive, Name: "every combination fully passed its live smoke (PASS or explicit UNSUPPORTED)"}
	for _, combo := range in.Trace.LiveCombos {
		i := slices.IndexFunc(in.Matrix.Rows, func(r supportmatrix.Row) bool { return r.Combo == combo })
		if i < 0 {
			s.Rows = append(s.Rows, LiveRow{Combo: combo, Status: NotRun})
			g.Problems = append(g.Problems, combo+": missing from the support matrix")
			continue
		}
		m := in.Matrix.Rows[i]
		row := LiveRow{Combo: combo, Operation: m.Operation, Provider: m.Provider, API: m.API, Model: m.Model, SDK: m.SDK, AccountAlias: m.AccountAlias,
			LastRunAt: m.LastRunAt, AllPassedAt: m.AllPassedAt}
		for _, c := range m.Capabilities {
			switch c.Outcome {
			case supportmatrix.Unsupported:
				row.Unsupported = append(row.Unsupported, c.ID+": "+c.Note)
			case supportmatrix.Pass:
			default:
				row.Failing = append(row.Failing, fmt.Sprintf("%s=%s %s", c.ID, c.Outcome, c.ErrorCategory))
			}
		}
		switch {
		case in.Matrix.Schema != supportmatrix.SchemaVersion || !m.ValidIdentity():
			row.Status = Fail
			g.Problems = append(g.Problems, combo+": invalid schema or route identity")
		case len(row.Failing) > 0:
			row.Status = Fail
			g.Problems = append(g.Problems, combo+": failing now: "+strings.Join(row.Failing, ", "))
		case m.AllPassedAt == nil:
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": never fully passed (no complete live run merged)")
		case staleSDK(m.SDK, in.Modules) != "":
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": "+staleSDK(m.SDK, in.Modules))
		case !slices.ContainsFunc(in.Audits, func(a AuditResult) bool { return a.Combo == combo }):
			row.Status = Pass
			g.Problems = append(g.Problems, combo+": no live evidence bundle of this combination was audited (pass it with -live)")
		default:
			row.Status = Pass
		}
		s.Rows = append(s.Rows, row)
	}
	return s, settle(g)
}

// staleSDK explains why a live pass made with sdk ("<module> <version>
// ...", as the live suite records it) no longer covers the code, or returns
// "". A pass through direct HTTP names no module and is never stale here;
// adapter changes are not detectable from the matrix and need a rerun by
// the maintainer (spec Testing Decisions §6).
func staleSDK(sdk string, modules map[string]string) string {
	fields := strings.Fields(sdk)
	if len(fields) < 2 || !strings.Contains(fields[0], "/") {
		return ""
	}
	if now, ok := modules[fields[0]]; ok && now != fields[1] {
		return fmt.Sprintf("last complete pass used %s %s, go.mod now pins %s; rerun the live smoke", fields[0], fields[1], now)
	}
	return ""
}

func auditGate(in Inputs) Gate {
	g := Gate{ID: GateAudit, Name: "redaction audit: no key, Authorization or real content in any bundle"}
	if !slices.ContainsFunc(in.Audits, func(a AuditResult) bool { return a.Bundle == in.Offline.Dir }) {
		g.Problems = append(g.Problems, "the offline bundle was not audited")
	}
	for _, a := range in.Audits {
		if a.Err != "" {
			g.Problems = append(g.Problems, a.Bundle+": "+a.Err)
		}
		for _, f := range a.Findings {
			g.Problems = append(g.Problems, fmt.Sprintf("%s: %s [%s] %s", a.Bundle, f.File, f.Rule, f.Detail))
		}
	}
	return settle(g)
}

func snapshotGate(in Inputs) Gate {
	g := Gate{ID: GateSnapshots, Name: "checked-in snapshots match the code"}
	if len(in.Snapshots) == 0 {
		g.Problems = append(g.Problems, "no snapshot was checked")
	}
	for _, s := range in.Snapshots {
		if !s.Current {
			g.Problems = append(g.Problems, s.Name+": stale: "+s.Detail)
		}
	}
	return settle(g)
}

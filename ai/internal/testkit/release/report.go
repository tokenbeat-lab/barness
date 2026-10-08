package release

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Markdown renders the report for maintainers: the verdict and each gate
// first, then the offline, differential and live results as separate
// sections (spec Testing Decisions §6), then traceability, audit and
// snapshots.
func (r Report) Markdown() string {
	var b strings.Builder
	verdict := "FAIL"
	if r.Pass {
		verdict = "PASS"
	}
	fmt.Fprintf(&b, "# barness-ai release gate: %s\n\n", verdict)
	fmt.Fprintf(&b, "Commit `%s`; offline bundle `%s`.\n\n", r.Commit, r.Bundle)
	b.WriteString("| Gate | Result | Problems |\n| --- | --- | --- |\n")
	for _, g := range r.Gates {
		fmt.Fprintf(&b, "| %s — %s | %s | %s |\n", g.ID, g.Name, passFail(g.Pass), cell(g.Problems, 8))
	}
	b.WriteString("\n### Commands\n\n| Command | Result | Duration |\n| --- | --- | --- |\n")
	for _, c := range r.Commands {
		fmt.Fprintf(&b, "| `%s` | %s | %s |\n", c.Line, passFail(c.Passed), c.Duration.Round(time.Second))
	}

	b.WriteString("\n## Offline\n\n")
	fmt.Fprintf(&b, "%s, differential cases excluded.\n\n", counts(r.Offline.Counts))
	b.WriteString("| P0 scenario | Pass | Fail | Not run | Unsupported |\n| --- | --- | --- | --- | --- |\n")
	for _, k := range keys(r.Offline.Scenarios) {
		c := r.Offline.Scenarios[k]
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d |\n", k, c.Pass, c.Fail, c.NotRun, c.Unsupported)
	}
	listing(&b, "Not passing", r.Offline.Failing)

	b.WriteString("\n## Differential\n\n")
	fmt.Fprintf(&b, "%s against frozen pi; %d pending finding(s). Ledger: %d extension, %d fixed, %d pending decision(s); extension routes outside pi:",
		counts(r.Differential.Counts), r.Differential.Pending, r.Differential.Ledger.Extension, r.Differential.Ledger.Fixed, r.Differential.Ledger.Pending)
	for _, rt := range r.Differential.Ledger.Routes {
		fmt.Fprintf(&b, " %s × %s", rt.Provider, rt.API)
	}
	b.WriteString(".\n\n| Protocol | Pass | Fail | Not run |\n| --- | --- | --- | --- |\n")
	for _, k := range keys(r.Differential.ByScenario) {
		c := r.Differential.ByScenario[k]
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", k, c.Pass, c.Fail, c.NotRun)
	}
	listing(&b, "Not passing", r.Differential.Failing)

	b.WriteString("\n## Live\n\nFrom the support matrix (ai/live/support-matrix.json).\n\n")
	b.WriteString("| Combination | Operation | Status | Model | SDK | Account | Last complete pass | Unsupported | Failing |\n| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, row := range r.Live.Rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", row.Combo, row.Operation, row.Status, dash(row.Model), dash(row.SDK),
			dash(row.AccountAlias), when(row.AllPassedAt), cell(row.Unsupported, 4), cell(row.Failing, 4))
	}

	b.WriteString("\n## Traceability\n\nStatus comes from evidence only; an item that is only mapped is NO_EVIDENCE, never PASS.\n\n")
	b.WriteString("| Item | Source | Requirement | Scenarios | Status | Offline | Differential | Live | Evidence |\n| --- | --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, t := range r.Trace {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s | %s | %s |\n", t.ID, t.Source, t.Requirement, strings.Join(t.Scenarios, " "),
			t.Status, short(t.Offline), short(t.Differential), short(t.Live), cell(t.Evidence, 3))
	}

	b.WriteString("\n## Redaction audit\n\n")
	b.WriteString("\n## Design load and business effect\n\nProtocol smoke and business accuracy are separate evidence. COMPLETE verifies the fixed evaluation, with no accuracy threshold.\n\n")
	for _, e := range r.Evidence {
		fmt.Fprintf(&b, "- %s: %s; `%s`; %s\n", e.Name, e.Status, e.Bundle, e.Detail)
	}
	for _, a := range r.Audits {
		state := fmt.Sprintf("%d finding(s)", len(a.Findings))
		if a.Err != "" {
			state = "error: " + a.Err
		}
		fmt.Fprintf(&b, "- `%s`: %s\n", a.Bundle, state)
	}
	b.WriteString("\n## Snapshots\n\n")
	for _, s := range r.Snapshots {
		state := "current"
		if !s.Current {
			state = "STALE: " + s.Detail
		}
		fmt.Fprintf(&b, "- %s: %s\n", s.Name, state)
	}
	return b.String()
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

func counts(c Counts) string {
	return fmt.Sprintf("%d cases: %d PASS, %d FAIL, %d NOT_RUN, %d UNSUPPORTED", c.Total(), c.Pass, c.Fail, c.NotRun, c.Unsupported)
}

func short(c Counts) string {
	if c.Total() == 0 {
		return "—"
	}
	s := fmt.Sprintf("%d/%d", c.Pass, c.Total())
	if c.Fail > 0 {
		s += fmt.Sprintf(" (%d FAIL)", c.Fail)
	}
	if c.NotRun > 0 {
		s += fmt.Sprintf(" (%d NOT_RUN)", c.NotRun)
	}
	return s
}

// cell joins up to n entries for a table cell, naming how many were left out.
func cell(items []string, n int) string {
	if len(items) == 0 {
		return "—"
	}
	shown := items[:min(n, len(items))]
	s := strings.ReplaceAll(strings.Join(shown, "; "), "|", "\\|")
	if len(items) > n {
		s += fmt.Sprintf("; … %d more", len(items)-n)
	}
	return s
}

func listing(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s:\n\n", title)
	for _, it := range items {
		fmt.Fprintf(b, "- %s\n", it)
	}
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func when(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.UTC().Format(time.RFC3339)
}

func keys(m map[string]Counts) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

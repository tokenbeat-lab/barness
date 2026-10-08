package release

import (
	"fmt"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// expectedCapabilities is the release contract, independent of a supplied
// report's Expected list. Keep extras in sync with the live scenario registry;
// a missing scenario must fail closed rather than shrink the release.
func expectedCapabilities(combo string) []string {
	switch combo {
	case "typesafe-classifier":
		return supportmatrix.ClassifierExpected()
	case "openai-images":
		return supportmatrix.ImagesExpected()
	case "google-interactions-image":
		return supportmatrix.GoogleImagesExpected()
	}
	base := []string{"text-stream", "text-complete", "tool-round-trip", "cancel-after-first-frame", "reasoning-history", "image"}
	extras := map[string][]string{
		"openai-responses":   {"tool-changes-gpt-5.4", "tool-changes-gpt-6-sol"},
		"anthropic-messages": {"tool-changes-claude-opus-4-8", "tool-changes-claude-fable-5", "effort-changes-claude-fable-5-1", "effort-changes-claude-opus-5", "effort-changes-claude-opus-5-5"},
		"google-gemini":      {},
		"openai-chat":        {"usage-position", "tool-changes-gpt-5.4", "tool-changes-gpt-6-sol", "pro-model-unavailable"},
		"deepseek-responses": {"reasoning-level-high", "reasoning-level-max"},
		"deepseek-chat":      {"usage-position", "forced-tool-with-thinking", "prompt-cache-long-retention", "reasoning-level-high", "reasoning-level-max"},
	}
	if extra, ok := extras[combo]; ok {
		return append(base, extra...)
	}
	return nil
}

func capabilityProblems(row supportmatrix.Row) []string {
	var problems []string
	expected := expectedCapabilities(row.Combo)
	if len(expected) == 0 || len(row.Capabilities) != len(expected) {
		problems = append(problems, "incomplete capability set")
	}
	for _, id := range expected {
		matches := 0
		for _, c := range row.Capabilities {
			if c.ID != id {
				continue
			}
			matches++
			if c.Model == "" || (row.Operation != "chat" && c.Model != row.Model) {
				problems = append(problems, id+": missing or substituted capability model")
			}
			if c.Outcome == supportmatrix.Unsupported && (c.Note == "" || (row.Operation != "chat" && !(row.Combo == "openai-images" && id == "mask-edit"))) {
				problems = append(problems, id+": required capability cannot be UNSUPPORTED")
			}
		}
		if matches != 1 {
			problems = append(problems, fmt.Sprintf("capability %s occurs %d times", id, matches))
		}
	}
	for _, c := range row.Capabilities {
		if !slices.Contains(expected, c.ID) {
			problems = append(problems, "unexpected capability "+c.ID)
		}
	}
	return problems
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
		row.Failing = append(row.Failing, capabilityProblems(m)...)
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
		case m.AllPassedAt == nil && !slices.ContainsFunc(m.Capabilities, func(c supportmatrix.Capability) bool { return c.Outcome == supportmatrix.Fail }):
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": never fully passed (no complete live run merged)")
		case len(row.Failing) > 0:
			row.Status = Fail
			g.Problems = append(g.Problems, combo+": failing now: "+strings.Join(row.Failing, ", "))
		case m.AllPassedAt == nil:
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": never fully passed (no complete live run merged)")
		case staleSDK(m.SDK, in.Modules) != "":
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": "+staleSDK(m.SDK, in.Modules))
		case liveEvidenceProblem(in, m) != "":
			row.Status = NotRun
			g.Problems = append(g.Problems, combo+": "+liveEvidenceProblem(in, m))
		default:
			row.Status = Pass
		}
		s.Rows = append(s.Rows, row)
	}
	return s, settle(g)
}

func liveEvidenceProblem(in Inputs, row supportmatrix.Row) string {
	var reasons []string
	for _, a := range in.Audits {
		if a.Combo != row.Combo || a.Live == nil {
			continue
		}
		r := *a.Live
		if a.Err != "" || len(a.Findings) > 0 {
			reasons = append(reasons, "live bundle failed audit")
			continue
		}
		if r.CatalogVersion != in.CatalogVersion || r.CatalogHash != in.CatalogHash || r.GitCommit == "" || r.StartedAt.IsZero() || r.FinishedAt.Before(r.StartedAt) {
			reasons = append(reasons, "live provenance or catalog/price snapshot is stale")
			continue
		}
		if row.LastRunAt == nil || row.AllPassedAt == nil || !r.FinishedAt.Equal(*row.LastRunAt) || !r.FinishedAt.Equal(*row.AllPassedAt) || r.Model != row.Model || r.SDK != row.SDK || r.AccountAlias != row.AccountAlias {
			reasons = append(reasons, "live report does not match the current complete matrix run")
			continue
		}
		// Reuse the matrix's report validation without mutating its history.
		fresh := row
		fresh.LastRunAt, fresh.AllPassedAt, fresh.Capabilities = nil, nil, nil
		if _, err := supportmatrix.Merge(supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion, Rows: []supportmatrix.Row{fresh}}, r); err != nil {
			reasons = append(reasons, "invalid live route or report: "+err.Error())
			continue
		}
		if len(r.Scenarios) != len(row.Capabilities) || len(r.Expected) != len(expectedCapabilities(row.Combo)) {
			reasons = append(reasons, "live report has an incomplete capability set")
			continue
		}
		valid := true
		for _, c := range row.Capabilities {
			i := slices.IndexFunc(r.Scenarios, func(s supportmatrix.ScenarioResult) bool { return s.ID == c.ID })
			if i < 0 || !slices.Contains(r.Expected, c.ID) || c.LastRunAt == nil || !c.LastRunAt.Equal(r.FinishedAt) {
				valid = false
				continue
			}
			s := r.Scenarios[i]
			if s.Outcome != c.Outcome || s.Model != c.Model || s.Note != c.Note || s.CaseID == "" || (s.Outcome != supportmatrix.Pass && s.Outcome != supportmatrix.Unsupported) {
				valid = false
			}
		}
		if valid {
			return ""
		}
		reasons = append(reasons, "live capability evidence disagrees with the matrix")
	}
	if len(reasons) == 0 {
		return "no live evidence bundle of this combination was audited (pass its current report with -live)"
	}
	return strings.Join(reasons, "; ")
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

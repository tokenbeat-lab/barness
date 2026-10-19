package release

import "slices"

// PreflightReport checks only external inputs available before collection.
// Ready does not certify a release or assert any unexecuted ordinary/race case.
type PreflightReport struct {
	Scope  string `json:"scope"`
	Ready  bool   `json:"ready"`
	Checks []Gate `json:"checks"`
}

func Preflight(in Inputs) PreflightReport {
	_, lg := live(in)
	out := PreflightReport{Scope: "preflight", Checks: []Gate{snapshotGate(in), lg}}
	ledger := Gate{ID: "difference-ledger", Name: "no pending differences and every extension route registered"}
	if in.Ledger.Pending != 0 {
		ledger.Problems = append(ledger.Problems, "pending differences block collection")
	}
	for _, route := range in.Trace.Differential.Routes {
		if !slices.Contains(in.Ledger.Routes, Route{Provider: route.Provider, API: route.API}) {
			ledger.Problems = append(ledger.Problems, "required extension route missing from difference ledger")
		}
	}
	out.Checks = append(out.Checks, settle(ledger))
	effects := Gate{ID: "external-effects", Name: "current independent Chinese and visual effects"}
	for _, name := range []string{"chinese-effect-jev", "chinese-effect-luna", "visual-effect-jev", "visual-effect-decisions"} {
		i := slices.IndexFunc(in.Evidence, func(e EvidenceCheck) bool { return e.Name == name })
		if i < 0 || in.Evidence[i].Status != Pass {
			effects.Problems = append(effects.Problems, name+": missing, stale or not PASS")
			continue
		}
		dir := in.Evidence[i].Bundle
		if dir == "" || !slices.ContainsFunc(in.Audits, func(a AuditResult) bool { return a.Bundle == dir && a.Err == "" && len(a.Findings) == 0 }) {
			effects.Problems = append(effects.Problems, name+": clean audit missing")
		}
	}
	out.Checks = append(out.Checks, settle(effects))
	audits := Gate{ID: "external-audit", Name: "all selected external bundles audited"}
	for _, dir := range in.RequiredAudits {
		if !slices.ContainsFunc(in.Audits, func(a AuditResult) bool { return a.Bundle == dir && a.Err == "" && len(a.Findings) == 0 }) {
			audits.Problems = append(audits.Problems, "selected external bundle missing or failed audit")
		}
	}
	out.Checks = append(out.Checks, settle(audits))
	out.Ready = !slices.ContainsFunc(out.Checks, func(g Gate) bool { return !g.Pass })
	return out
}

// OrdinaryChecks applies the original consumers without requiring the
// deliberately uncollected independent race evidence.
func OrdinaryChecks(in Inputs) []Gate {
	_, off := offline(in)
	_, diff := differential(in)
	evidence := append(DesignLoadEvidence(in.Offline), RevocationEvidence(in.Offline))
	evidence = append(evidence, MultimodalImageEvidence(in.Offline)...)
	evidence = append(evidence, MiniMaxDesignLoadEvidence(in.Offline)...)
	g := Gate{ID: "ordinary-artifacts", Name: "ordinary native artifacts and audits"}
	for _, e := range evidence {
		if e.Status != Pass {
			g.Problems = append(g.Problems, e.Name+": "+string(e.Status))
		}
	}
	return []Gate{off, diff, auditGate(in), settle(g)}
}

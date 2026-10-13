package release

import "strings"

// RaceEvidence applies the same artifact consumers to the independent race run.
// A successful process or a clean redaction scan alone cannot prove its P0,
// frozen differential, design loads or authorization revocation.
func RaceEvidence(dir string, trace TraceMap, ledger Ledger) []EvidenceCheck {
	names := []string{"race-offline", "race-differential", "race-mixed-pressure-local", "race-mixed-pressure-cloud", "race-authorization-revocation"}
	b, err := LoadBundle(dir)
	if dir == "" || err != nil {
		var out []EvidenceCheck
		for _, name := range names {
			out = append(out, EvidenceCheck{Name: name, Bundle: dir, Status: NotRun, Detail: "independent race bundle missing or invalid"})
		}
		return out
	}
	in := Inputs{Offline: b, Trace: trace, Ledger: ledger}
	_, off := offline(in)
	_, diff := differential(in)
	var out []EvidenceCheck
	for i, g := range []Gate{off, diff} {
		status := Pass
		if !g.Pass {
			status = Fail
		}
		out = append(out, EvidenceCheck{Name: names[i], Bundle: dir, Status: status, Detail: strings.Join(g.Problems, "; ")})
	}
	for _, e := range append(DesignLoadEvidence(b), RevocationEvidence(b)) {
		e.Name = "race-" + e.Name
		out = append(out, e)
	}
	return out
}

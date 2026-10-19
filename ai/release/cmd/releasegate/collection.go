package main

import (
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"os"
	"path/filepath"
)

func externalInputs(o options) (release.Inputs, error) {
	in := release.Inputs{}
	snapshot, err := release.CatalogSnapshot()
	if err != nil {
		return in, err
	}
	var catalog struct {
		Hash    string
		Catalog struct{ Version string }
	}
	if err = json.Unmarshal(snapshot, &catalog); err != nil {
		return in, err
	}
	in.CatalogVersion, in.CatalogHash = catalog.Catalog.Version, catalog.Hash
	if in.Trace, err = release.LoadTraceMap(o.Trace); err != nil {
		return in, fmt.Errorf("trace input unavailable or invalid")
	}
	if in.Ledger, err = release.LoadLedger(o.Ledger); err != nil {
		return in, fmt.Errorf("ledger input unavailable or invalid")
	}
	if in.Matrix, err = supportmatrix.LoadMatrix(o.Matrix); err != nil {
		return in, fmt.Errorf("matrix input unavailable or invalid")
	}
	if in.Modules, err = release.GoModules("."); err != nil {
		return in, fmt.Errorf("module inputs unavailable")
	}
	extension, _ := release.CheckClassifierExtensions(o.Extensions)
	in.Snapshots = []release.Snapshot{release.CheckSnapshot("catalog", o.Snapshot, snapshot), release.CheckReleasePlan(in.Trace), release.HistoricalPressureBlocker(), extension}
	for _, dir := range append(append([]string{}, o.Live...), o.Jev, o.Luna, o.VisualJev, o.VisualDecisions) {
		if dir != "" {
			in.RequiredAudits = append(in.RequiredAudits, dir)
			in.Audits = append(in.Audits, release.AuditBundle(dir, audit.Environment()))
		}
	}
	in.Evidence = append(release.ChineseEffectsEvidence(o.Jev, o.Luna), release.VisualEffectsEvidence(o.VisualJev, o.VisualDecisions)...)
	return in, nil
}
func loadOrdinary(in *release.Inputs, bundle string) error {
	b, err := release.LoadBundle(bundle)
	if err != nil {
		return fmt.Errorf("ordinary bundle missing or invalid")
	}
	in.Offline = b
	in.RequiredAudits = append(in.RequiredAudits, bundle)
	in.Audits = append(in.Audits, release.AuditBundle(bundle, audit.Environment()))
	in.Evidence = append(in.Evidence, append(append(release.DesignLoadEvidence(b), release.RevocationEvidence(b)), release.MultimodalImageEvidence(b)...)...)
	return nil
}
func writeJSON(dir, name string, v any) error {
	raw, err := release.Encode(v)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), raw, 0600)
}
func writeVerdict(o options, in release.Inputs) (bool, error) {
	r := release.Evaluate(in)
	if err := writeJSON(o.Out, "release-report.json", r); err != nil {
		return false, err
	}
	if err := os.WriteFile(filepath.Join(o.Out, "release-report.md"), []byte(r.Markdown()), 0600); err != nil {
		return false, err
	}
	return r.Pass, nil
}

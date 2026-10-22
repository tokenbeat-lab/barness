package release

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// qualifiedLive reads the same native contracts as the published support
// consumer, without updating the support matrix or starting any process.
func qualifiedLive(root, dir string) (supportmatrix.Report, error) {
	report, err := supportmatrix.LoadReport(filepath.Join(dir, supportmatrix.ReportFile))
	if err != nil {
		return report, fmt.Errorf("native live report missing")
	}
	route := report.Combo
	switch {
	case strings.HasPrefix(route, "zai-") || strings.HasPrefix(route, "minimax-cn-"):
		report, err = supportmatrix.NewRegionalReport(dir)
	case strings.HasPrefix(route, "minimax-"):
		report, err = supportmatrix.NewMiniMaxReport(dir)
	case route == supportmatrix.JevFormalCombo:
		report, err = supportmatrix.NewJevReport(dir)
	case len(supportmatrix.GLMExpected(route)) > 0:
		report, err = supportmatrix.NewGLMReport(dir)
	}
	if err != nil {
		return report, fmt.Errorf("native live qualification unavailable")
	}
	if err := verifyLiveManifest(dir, report); err != nil {
		return report, err
	}
	audited := AuditBundle(dir, audit.Environment())
	if audited.Err != "" || len(audited.Findings) > 0 {
		return report, fmt.Errorf("native live audit failed")
	}
	// A fresh row retains every operation-specific native check and avoids
	// coupling acceptance to mutable historical support timestamps.
	_, err = supportmatrix.Merge(supportmatrix.Matrix{Schema: supportmatrix.SchemaVersion,
		Rows: []supportmatrix.Row{{Combo: report.Combo, Operation: report.Operation, Provider: report.Provider, API: report.API}}}, report)
	if err != nil {
		return report, fmt.Errorf("native live qualification rejected: %w", err)
	}
	expected := expectedCapabilities(route)
	a, b := slices.Clone(expected), slices.Clone(report.Expected)
	slices.Sort(a)
	slices.Sort(b)
	if len(expected) == 0 || !slices.Equal(a, b) || len(report.Scenarios) != len(expected) {
		return report, fmt.Errorf("incomplete native capability set")
	}
	catalog := ai.BuiltinCatalog()
	hash, err := catalog.Hash()
	if err != nil || report.CatalogVersion != catalog.Version || report.CatalogHash != hash {
		return report, fmt.Errorf("live catalog drift")
	}
	modules, err := GoModules(root)
	if err != nil || staleSDK(report.SDK, modules) != "" {
		return report, fmt.Errorf("live SDK drift")
	}
	seen := map[string]bool{}
	for _, s := range report.Scenarios {
		if seen[s.ID] || !slices.Contains(expected, s.ID) || s.Outcome != supportmatrix.Pass {
			return report, fmt.Errorf("duplicate, substituted or unpassed native scenario")
		}
		seen[s.ID] = true
	}
	return report, nil
}

func verifySelectedLive(root, dir, route string, stage StageRecord, source ExecutionSource) error {
	report, err := qualifiedLive(root, dir)
	if err != nil {
		return err
	}
	if report.Combo != route || report.GitCommit != source.Commit || len(stage.Consumers) != 1 {
		return fmt.Errorf("live qualification stale or consumer receipt missing")
	}
	seen := map[string]bool{}
	for _, id := range stage.RequiredCases {
		if id == "" || seen[id] || !slices.ContainsFunc(report.Scenarios, func(s supportmatrix.ScenarioResult) bool { return s.CaseID == id }) {
			return fmt.Errorf("live required case missing or duplicated")
		}
		seen[id] = true
	}
	return verifyLiveConsumer(root, dir, stage.Consumers[0], source, report)
}

package release

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/audit"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// QualifyNativeLive reads the same native contracts as the published support
// consumer, without updating the support matrix or starting any process.
func QualifyNativeLive(root, dir string, supplement ...string) (supportmatrix.Report, error) {
	report, err := supportmatrix.LoadReport(filepath.Join(dir, supportmatrix.ReportFile))
	projectedMiniMax := err != nil
	original := report
	native := true
	if err != nil {
		if _, e := os.Stat(filepath.Join(dir, minimaxlive.ReportFile)); e != nil {
			return report, fmt.Errorf("native live report missing")
		}
		report, err = supportmatrix.NewMiniMaxReport(dir)
		if err != nil {
			return report, err
		}
	}
	route := report.Combo
	switch {
	case route == "xai-responses":
		report, err = supportmatrix.NewXAIReport(dir)
	case isRegionalLiveRoute(route):
		report, err = supportmatrix.NewRegionalReport(dir)
	case strings.HasPrefix(route, "minimax-"):
		if !projectedMiniMax {
			report, err = supportmatrix.NewMiniMaxReport(dir)
		}
	case route == supportmatrix.JevFormalCombo:
		report, err = supportmatrix.NewJevReport(dir)
	case len(supportmatrix.GLMExpected(route)) > 0:
		report, err = supportmatrix.NewGLMReport(dir)
	default:
		native = false
	}
	if err != nil {
		return report, fmt.Errorf("native live qualification unavailable")
	}
	if native && !projectedMiniMax {
		a, _ := json.Marshal(original)
		b, _ := json.Marshal(report)
		if !bytes.Equal(a, b) {
			return report, fmt.Errorf("native evidence audit failed")
		}
	}
	generic := len(supportmatrix.GenericExpected(route)) > 0
	if len(supplement) > 1 || !generic && len(supplement) == 1 && supplement[0] != "" {
		return report, fmt.Errorf("source supplement requires one generic route")
	}
	if generic {
		if err := supportmatrix.VerifyGenericReport(report); err != nil {
			return report, err
		}
	}
	audited := auditBundle(dir, audit.Environment(), !native, supplement...)
	if audited.Err != "" || len(audited.Findings) > 0 {
		return report, fmt.Errorf("native evidence audit failed")
	}
	// A fresh row retains every operation-specific native check and avoids
	// coupling acceptance to mutable historical support timestamps.
	err = supportmatrix.VerifyReportShape(report)
	if err != nil {
		return report, fmt.Errorf("native live qualification rejected: %w", err)
	}
	if !generic {
		expected := expectedCapabilities(route)
		a, b := slices.Clone(expected), slices.Clone(report.Expected)
		slices.Sort(a)
		slices.Sort(b)
		if len(expected) == 0 || !slices.Equal(a, b) || len(report.Scenarios) != len(expected) {
			return report, fmt.Errorf("incomplete native capability set")
		}
	}
	if _, err := supportmatrix.RouteCatalog(dir, report); err != nil {
		return report, err
	}
	modules, err := GoModules(root)
	if err != nil || staleSDK(report.SDK, modules) != "" {
		return report, fmt.Errorf("live SDK drift")
	}
	if !generic {
		expected := expectedCapabilities(route)
		seen := map[string]bool{}
		for _, s := range report.Scenarios {
			if seen[s.ID] || !slices.Contains(expected, s.ID) || s.Outcome != supportmatrix.Pass {
				return report, fmt.Errorf("duplicate, substituted or unpassed native scenario")
			}
			seen[s.ID] = true
		}
	}
	return report, nil
}

func verifySelectedLive(root, dir, route string, stage StageRecord, source, measured ExecutionSource) error {
	supplement := ""
	if len(stage.Consumers) == 1 {
		_, args, err := readLiveConsumer(root, stage.Consumers[0])
		if err != nil {
			return err
		}
		composition, compositionErr := recordPath(root, args.bundle)
		if compositionErr == nil && supportmatrix.IsMiniMaxThinkingComposition(composition) {
			if args.supplement != "" {
				return fmt.Errorf("MiniMax composition source override forbidden")
			}
			return verifySelectedMiniMaxThinking(root, dir, composition, route, stage, source, measured)
		}
		if compositionErr == nil && supportmatrix.IsKimiComposition(composition) {
			if args.supplement != "" {
				return fmt.Errorf("Kimi composition source override forbidden")
			}
			return verifySelectedKimiComposition(root, dir, composition, route, stage, source, measured)
		}
		if args.supplement != "" {
			supplement, err = recordPath(root, args.supplement)
			if err != nil {
				return err
			}
		}
	}
	report, err := QualifyNativeLive(root, dir, supplement)
	if err != nil {
		return err
	}
	if report.Combo != route || report.GitCommit != measured.Commit || len(stage.Consumers) != 1 {
		return fmt.Errorf("live qualification stale or consumer receipt missing")
	}
	seen := map[string]bool{}
	for _, id := range stage.RequiredCases {
		if id == "" || seen[id] || !slices.ContainsFunc(report.Scenarios, func(s supportmatrix.ScenarioResult) bool { return s.CaseID == id }) {
			return fmt.Errorf("live required case missing or duplicated")
		}
		seen[id] = true
	}
	return verifyLiveConsumer(root, dir, stage.Consumers[0], source, liveReportHash(report))
}

func isRegionalLiveRoute(route string) bool { _, ok := supportmatrix.RegionalRouteOf(route); return ok }

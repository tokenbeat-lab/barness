package supportmatrix

import (
	"fmt"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/grokhistory"
	"regexp"
	"slices"
	"time"
)

const KimiHistoryPurpose = "xai-history-supplement"

type regionalSelection uint8

const (
	regionalFull regionalSelection = iota
	kimiBaselineNine
	kimiHistoryOne
)

func selectedRegionalBudget(combo string, selection regionalSelection) Budget {
	b := RegionalBudgetFor(combo)
	switch selection {
	case kimiBaselineNine:
		b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens = 9, 9, 36864
	case kimiHistoryOne:
		b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens, b.MaxReservedInputTokens, b.ProcessSeconds = 1, 1, 4096, 4096, 240
	}
	return b
}
func selectedRegionalExpected(combo string, selection regionalSelection) []string {
	all := RegionalExpected(combo)
	switch selection {
	case kimiBaselineNine:
		return slices.Clone(all[:len(all)-1])
	case kimiHistoryOne:
		return []string{"kimi-k3/" + grokhistory.Scenario}
	}
	return all
}
func selectedRegionalContract(combo string, selection regionalSelection) string {
	if selection == kimiHistoryOne {
		return "kimi/" + combo + "/xai-origin-history/v1"
	}
	return RegionalAuthorization(combo)
}
func validManifestHash(h string) bool { return regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(h) }

// KimiHistoryGrantSnapshot is incomplete machine scope, never execution
// authority. Human identity/expiry/spend and the frozen binary remain absent.
func KimiHistoryGrantSnapshot(combo, baseline string) (RegionalGrant, error) {
	if !RegionalIsKimi(combo) || !validManifestHash(baseline) {
		return RegionalGrant{}, fmt.Errorf("Kimi baseline manifest hash required")
	}
	g, err := RegionalGrantSnapshot(combo, "candidate")
	if err != nil {
		return g, err
	}
	g.Schema, g.Purpose, g.Contract = 2, KimiHistoryPurpose, selectedRegionalContract(combo, kimiHistoryOne)
	g.ScenarioIDs = selectedRegionalExpected(combo, kimiHistoryOne)
	g.BaselineManifestSHA256 = baseline
	b := selectedRegionalBudget(combo, kimiHistoryOne)
	g.MaxCalls, g.MaxAttempts, g.MaxRequestedOutputTokens, g.MaxReservedInputTokens = b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens, b.MaxReservedInputTokens
	g.CallIntervalMs = 60000
	return g, nil
}

func (g RegionalGrant) checkSelectedMeasurement(r Report, at time.Time, selection regionalSelection) error {
	var want RegionalGrant
	var err error
	if selection == kimiHistoryOne {
		want, err = KimiHistoryGrantSnapshot(r.Combo, g.BaselineManifestSHA256)
	} else {
		want, err = RegionalGrantSnapshot(r.Combo, g.Stage)
	}
	if selection != regionalFull && !RegionalIsKimi(r.Combo) {
		return refuse("Kimi-only evidence selection")
	}
	if selection == kimiBaselineNine {
		if g.Stage != "candidate" {
			return refuse("historical nine is candidate evidence")
		}
		b := selectedRegionalBudget(r.Combo, selection)
		want.MaxCalls, want.MaxAttempts, want.MaxRequestedOutputTokens = b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens
	}
	want.SourceSHA256, want.CatalogVersion, want.CatalogHash = r.Regional.SourceSHA256, r.CatalogVersion, r.CatalogHash
	return g.check(r.Combo, at, want, err)
}

// CheckKimiBaselineFixture is only the controlled old-nine fixture admission.
// No live grant can enter it, even in a specially tagged test binary.
func (g RegionalGrant) CheckKimiBaselineFixture(combo string, at time.Time) error {
	if g.Mode != "controlled" || !RegionalIsKimi(combo) || g.Stage != "candidate" {
		return refuse("baseline fixture must be controlled Kimi candidate")
	}
	want, err := RegionalGrantSnapshot(combo, g.Stage)
	b := selectedRegionalBudget(combo, kimiBaselineNine)
	want.MaxCalls, want.MaxAttempts, want.MaxRequestedOutputTokens = b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens
	return g.check(combo, at, want, err)
}

// KimiHistoryBudget is the fixed single-scene process profile, not a free-form selector.
func KimiHistoryBudget() Budget { return selectedRegionalBudget("kimi-cn-chat", kimiHistoryOne) }

// VerifyKimiHistoryCapture checks selected evidence only. It never grants full support.
func VerifyKimiHistoryCapture(dir, mode string) (Report, error) {
	if mode != "live" && mode != "controlled" {
		return Report{}, refuse("invalid selected capture mode")
	}
	return regionalSelectedReport(dir, false, mode, kimiHistoryOne)
}

func VerifyKimiBaselineCapture(dir, mode string) (Report, error) {
	if mode != "live" && mode != "controlled" {
		return Report{}, refuse("invalid baseline mode")
	}
	return regionalSelectedReport(dir, false, mode, kimiBaselineNine)
}

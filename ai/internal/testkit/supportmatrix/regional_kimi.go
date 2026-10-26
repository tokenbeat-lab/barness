package supportmatrix

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

// Kimi has an independent contract atop the existing native capture lifecycle.
// Old regional routes deliberately keep their exact identities and budgets.
func kimiRouteOf(combo string) (RegionalRoute, bool) {
	var provider, base, kind string
	switch {
	case strings.HasPrefix(combo, "kimi-cn-"):
		provider, base, kind = "moonshot-cn", "https://api.moonshot.cn", strings.TrimPrefix(combo, "kimi-cn-")
	case strings.HasPrefix(combo, "kimi-intl-"):
		provider, base, kind = "moonshot", "https://api.moonshot.ai", strings.TrimPrefix(combo, "kimi-intl-")
	default:
		return RegionalRoute{}, false
	}
	r := RegionalRoute{Combo: combo, Provider: provider, Endpoint: base + "/v1"}
	switch kind {
	case "chat":
		r.API = "openai-completions"
	case "responses":
		r.API = "openai-responses"
	case "messages":
		r.API = "anthropic-messages"
		r.Endpoint = base + "/anthropic"
	default:
		return RegionalRoute{}, false
	}
	return r, true
}

func RegionalIsKimi(combo string) bool { _, ok := kimiRouteOf(combo); return ok }

func RegionalCandidateCatalog(combo string) (ai.Catalog, error) {
	if RegionalIsKimi(combo) {
		return kimi.CandidateCatalog(), nil
	}
	return regional.CandidateCatalog()
}

func RegionalCandidateFacts(combo string) any {
	if RegionalIsKimi(combo) {
		return struct {
			Snapshot string       `json:"snapshot"`
			Prices   []kimi.Price `json:"prices"`
		}{kimi.SnapshotHash(), kimi.Prices()}
	}
	return regional.Candidates()
}

func RegionalBudgetFor(combo string) Budget {
	b := Budget{MaxCalls: 24, MaxAttempts: 24, MaxRetries: 0, MaxOutputTokens: 4096, MaxRequestedOutputTokens: 98304, MaxReservedInputTokens: 120000, CallSeconds: 120, ProcessSeconds: 1800}
	if RegionalIsKimi(combo) {
		b.MaxCalls, b.MaxAttempts, b.MaxRequestedOutputTokens, b.MaxReservedInputTokens = 10, 10, 40960, 131072
	}
	return b
}

// Reserve the encoded input conservatively as the most expensive input-side
// class, including one-hour writes. Never release this reservation based on
// incomplete or canceled response counters, or on the Usage.Cost lower point.
func RegionalReservedUSD(combo, model string, input, output int) (float64, error) {
	if !RegionalIsKimi(combo) {
		return 0, nil
	}
	route, _ := RegionalRouteOf(combo)
	cat, err := RegionalCandidateCatalog(combo)
	if err != nil {
		return 0, err
	}
	m, ok := cat.Lookup(ai.ProviderID(route.Provider), ai.API(route.API), model)
	if !ok || input < 0 || output < 0 {
		return 0, refuse("Kimi reservation identity or counts invalid")
	}
	rate := max(m.Cost.Input, m.Cost.CacheRead, m.Cost.CacheWrite, m.Cost.Input*2)
	return float64(input)*rate/1e6 + float64(output)*m.Cost.Output/1e6, nil
}

func regionalValidUSD(n float64) bool { return n > 0 && !math.IsNaN(n) && !math.IsInf(n, 0) }

func regionalSameFacts(got any, combo string) bool {
	raw, _ := json.Marshal(RegionalCandidateFacts(combo))
	var want any
	return json.Unmarshal(raw, &want) == nil && string(mustRegionalJSON(got)) == string(mustRegionalJSON(want))
}

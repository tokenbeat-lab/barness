package minimaxlive

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

const ThinkingProfile = "minimax-messages-thinking-on/v1"
const ThinkingKind = "minimax-thinking-on-supplement"

// ThinkingOnScope is deliberately one concrete contract, not a selectable plan.
// Account purpose stays standard; the baseline and USD are execution authority.
func ThinkingOnScope(id Identity, mode, binary, baseline string) (livegrant.Scope, error) {
	want, e := Select("messages", M3, "standard")
	if e != nil || id != want || (mode != "live" && mode != "controlled") || !regexp.MustCompile(`^[a-f0-9]{64}$`).MatchString(baseline) {
		return livegrant.Scope{}, fmt.Errorf("invalid MiniMax thinking-on identity or baseline pin")
	}
	s, e := ProcessScope(id, mode, "candidate", binary)
	if e != nil {
		return s, e
	}
	s.Scenes = []string{"thinking-on"}
	s.MaxCalls = 1
	s.MaxAttempts = 1
	s.MaxRequestReruns = 0
	s.MaxRetries = 0
	s.MaxOutputTokens = 2048
	s.MaxRequestedOutputTokens = 2048
	s.MaxReservedInputTokens = 2048
	s.ProcessSeconds = 180
	s.BaselineManifestSHA256 = baseline
	s.MaxUSD = 0.01
	return s, nil
}
func ThinkingPlan(id Identity) []Case {
	for _, p := range Plan(id) {
		if p.ID == "thinking-on" {
			return []Case{p}
		}
	}
	return nil
}

// The known highest price tier reserves both input and output, without taking
// a cache discount. Positive unpriced cache writes remain a qualification error.
type ThinkingBudget struct {
	Profile     string  `json:"profile"`
	WireSHA256  string  `json:"wireSha256"`
	WireBytes   int     `json:"wireBytes"`
	Input       int     `json:"reservedInput"`
	Output      int     `json:"requestedOutput"`
	ReservedUSD float64 `json:"reservedUSD"`
	MaxUSD      float64 `json:"maxUSD"`
}

func ThinkingWireBudget(raw []byte, s livegrant.Scope) (ThinkingBudget, error) {
	var b struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		Stream    bool   `json:"stream"`
		Thinking  struct {
			Type string `json:"type"`
		} `json:"thinking"`
	}
	if e := minimax.VerifyPriceSources(); e != nil {
		return ThinkingBudget{}, e
	}
	id, _ := Select("messages", M3, "standard")
	m, ok := CandidateCatalog(id).Lookup(ai.ProviderMiniMax, id.API, id.Model)
	if !ok {
		return ThinkingBudget{}, fmt.Errorf("MiniMax fixed price model missing")
	}
	input, output := m.Cost.Input, m.Cost.Output
	for _, tier := range m.Cost.Tiers {
		input = max(input, tier.Input)
		output = max(output, tier.Output)
	}
	p := ThinkingBudget{ThinkingProfile, Hash(raw), len(raw), max(2048, len(raw)+1024), 2048, float64(2048) * (input + output) / 1e6, s.MaxUSD}
	if json.Unmarshal(raw, &b) != nil || b.Model != M3 || b.MaxTokens != 2048 || !b.Stream || b.Thinking.Type != "adaptive" || p.Input > s.MaxReservedInputTokens || p.Output > s.MaxRequestedOutputTokens || p.ReservedUSD > s.MaxUSD || s.MaxUSD != 0.01 {
		return p, fmt.Errorf("MiniMax thinking-on final wire or known-rate reservation exceeds scope")
	}
	return p, nil
}

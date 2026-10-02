package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestAnthropicText is P02's text path and stream parsing through the public
// Client (E01 success terminal, interleaved blocks): every scenario of
// testdata/anthropic/text.json is streamed; the text scenarios also run
// through Result without ever calling Next and through Complete, on the full
// and the simple entry, and every SSE framing decodes to the same result.
func TestAnthropicText(t *testing.T) {
	f, raw := loadFixture(t, anthropicProtocol, "text.json")
	for _, sc := range f.Scenarios {
		modes := []callMode{modeStream}
		if sc.ID == "text" || sc.ID == "text-simple" {
			modes = append(modes, modeResult, modeComplete)
		}
		for _, mode := range modes {
			t.Run(sc.ID+"/"+string(mode), func(t *testing.T) {
				ev := run.Case(t, "P02-E01-"+sc.ID+"-"+string(mode))
				runScenario(t, ev, sc, raw, mode)
			})
		}
	}
	for _, framing := range provider.Framings {
		t.Run("text/framing-"+string(framing), func(t *testing.T) {
			ev := run.Case(t, "P02-E01-text-framing-"+string(framing))
			sc := scenarioByID(t, f, "text")
			sc.framing = framing
			runScenario(t, ev, sc, raw, modeStream)
		})
	}
}

func scenarioByID(t *testing.T, f scenarioFixture, id string) fixtureScenario {
	t.Helper()
	for _, sc := range f.Scenarios {
		if sc.ID == id {
			return sc
		}
	}
	t.Fatalf("no scenario %q", id)
	return fixtureScenario{}
}

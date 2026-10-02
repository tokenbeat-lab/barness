package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestGeminiText is P03's text path and stream parsing through the public
// Client (E01 success terminal, thought, text and function call parts):
// every scenario of testdata/gemini/text.json is streamed; the text
// scenarios also run through Result without ever calling Next and through
// Complete, on the full and the simple entry, and every data-event framing
// decodes to the same result.
func TestGeminiText(t *testing.T) {
	f, raw := loadFixture(t, geminiProtocol, "text.json")
	for _, sc := range f.Scenarios {
		modes := []callMode{modeStream}
		if sc.ID == "text" || sc.ID == "text-simple" {
			modes = append(modes, modeResult, modeComplete)
		}
		for _, mode := range modes {
			t.Run(sc.ID+"/"+string(mode), func(t *testing.T) {
				ev := run.Case(t, "P03-E01-"+sc.ID+"-"+string(mode))
				runScenario(t, ev, sc, raw, mode)
			})
		}
	}
	for _, framing := range provider.DataFramings {
		t.Run("text/framing-"+string(framing), func(t *testing.T) {
			ev := run.Case(t, "P03-E01-text-framing-"+string(framing))
			sc := scenarioByID(t, f, "text")
			sc.framing = framing
			runScenario(t, ev, sc, raw, modeStream)
		})
	}
}

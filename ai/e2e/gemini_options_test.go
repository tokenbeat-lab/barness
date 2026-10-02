package e2e

import "testing"

// TestGeminiOptions is P03/E04: full and simple options, field presence,
// the tool calling mode and the Google level/budget reasoning mapping, with
// samplingParams and the options pi's Google path ignores sending nothing
// (testdata/gemini/options.json).
func TestGeminiOptions(t *testing.T) {
	f, raw := loadFixture(t, geminiProtocol, "options.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P03-E04-"+sc.ID)
			runScenario(t, ev, sc, raw, modeStream)
		})
	}
}

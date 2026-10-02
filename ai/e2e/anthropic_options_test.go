package e2e

import "testing"

// TestAnthropicOptions is P02/E04: full and simple options, the Anthropic
// reasoning mapping (adaptive effort, token budgets), field presence, cache
// retention and samplingParams being ignored (testdata/anthropic/options.json).
func TestAnthropicOptions(t *testing.T) {
	f, raw := loadAnthropicFixture(t, "options.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P02-E04-"+sc.ID)
			runAnthropic(t, ev, sc, raw, modeStream)
		})
	}
}

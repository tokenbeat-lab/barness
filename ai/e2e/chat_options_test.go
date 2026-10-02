package e2e

import "testing"

// TestChatOptions is P04/E04: full and simple options, field presence, the
// tool choice forms, the reasoning effort mapping, the prompt cache fields
// and samplingParams applied over the named fields, as OpenAI-compatible
// protocols do (testdata/chat/options.json).
func TestChatOptions(t *testing.T) {
	f, raw := loadFixture(t, chatProtocol, "options.json")
	for _, sc := range f.Scenarios {
		t.Run(sc.ID, func(t *testing.T) {
			ev := run.Case(t, "P04-E04-"+sc.ID)
			runScenario(t, ev, sc, raw, modeStream)
		})
	}
}

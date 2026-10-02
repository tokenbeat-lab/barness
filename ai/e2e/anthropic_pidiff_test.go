package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// anthropicFixtureFiles are the P02 fixtures under testdata/anthropic; the
// differential runs every scenario in them that does not opt out.
var anthropicFixtureFiles = []string{"text.json", "failures.json", "history.json", "options.json", "usage.json", "retry.json"}

// anthropicPiDifferential is P02's differential gate: the offline Anthropic
// scenarios, same input and response script, through frozen pi-ai.
func anthropicPiDifferential(t *testing.T, ledger pioracle.Ledger) {
	for _, file := range anthropicFixtureFiles {
		f, raw := loadAnthropicFixture(t, file)
		for _, sc := range f.Scenarios {
			if sc.PidiffSkip != "" {
				continue
			}
			t.Run(sc.ID+"/"+sc.entry(), func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-P02-"+sc.ID+"-"+sc.entry())
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, sc.entry(), sc.pidiff(t, file, raw))
			})
		}
	}
}

// pidiff is the scenario for the differential. pi's options are the
// scenario's, plus the binding's retry count as pi's maxRetries.
func (sc anthropicScenario) pidiff(t *testing.T, file string, raw []byte) pidiffScenario {
	t.Helper()
	options := map[string]json.RawMessage{}
	if sc.Options != nil {
		mustUnmarshal(t, sc.Options, &options)
	}
	if sc.MaxRetries > 0 {
		options["maxRetries"] = mustMarshal(t, sc.MaxRetries)
	}
	p := pidiffScenario{api: ai.APIAnthropicMessages, fixture: file, raw: raw, model: sc.Model, req: sc.request(t),
		replies: sc.replies(t), abortAfter: sc.CancelAfterEvents, requests: max(sc.Expect.Requests, 1),
		piOptionsRaw: mustMarshal(t, options), modelPatch: sc.ModelPatch, retry: ai.RetryPolicy{MaxRetries: sc.MaxRetries}}
	if sc.entry() == "streamSimple" {
		p.simple = sc.simpleOptions(t)
	} else {
		p.full = sc.fullOptions(t)
	}
	return p
}

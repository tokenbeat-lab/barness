package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// scenarioPiDifferential is a protocol's differential gate: its offline
// scenarios, same input and response script, through frozen pi-ai. Case ids
// are "PIDIFF-<p>-<scenario>-<entry>".
func scenarioPiDifferential(t *testing.T, ledger pioracle.Ledger, proto *fixtureProtocol, p string, files []string) {
	for _, file := range files {
		f, raw := loadFixture(t, proto, file)
		for _, sc := range f.Scenarios {
			if sc.PidiffSkip != "" {
				continue
			}
			t.Run(sc.ID+"/"+sc.entry(), func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-"+p+"-"+sc.ID+"-"+sc.entry())
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				differential(t, ev, o, ledger, sc.entry(), sc.pidiff(t, file, raw))
			})
		}
	}
}

// pidiff is the scenario for the differential. pi's options are the
// scenario's, plus the binding's retry count as pi's maxRetries.
func (sc fixtureScenario) pidiff(t *testing.T, file string, raw []byte) pidiffScenario {
	t.Helper()
	options := map[string]json.RawMessage{}
	if sc.Options != nil {
		mustUnmarshal(t, sc.Options, &options)
	}
	if sc.MaxRetries > 0 {
		options["maxRetries"] = mustMarshal(t, sc.MaxRetries)
	}
	p := pidiffScenario{api: sc.proto.api, fixture: file, raw: raw, model: sc.Model, req: sc.request(t),
		replies: sc.replies(t), abortAfter: sc.CancelAfterEvents, requests: max(sc.Expect.Requests, 1),
		piOptionsRaw: mustMarshal(t, options), modelPatch: sc.ModelPatch, retry: ai.RetryPolicy{MaxRetries: sc.MaxRetries}}
	if sc.entry() == "streamSimple" {
		p.simple = sc.simpleOptions(t)
	} else {
		p.full = sc.fullOptions(t)
	}
	return p
}

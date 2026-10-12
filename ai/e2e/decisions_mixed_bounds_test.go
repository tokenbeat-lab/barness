package e2e

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsScoreBounds(t *testing.T) {
	for _, levels := range []int{2, 10} {
		for _, entry := range []string{"plain", "hooked"} {
			t.Run(fmt.Sprint(levels)+"/"+entry, func(t *testing.T) {
				ev := run.Case(t, fmt.Sprintf("P10-OD13-score-bound-%d-%s", levels, entry))
				u := newDecisionsMixedWorld(t, nil)
				q := ai.ScoreQuestion{Instructions: json.RawMessage(`"Rate"`)}
				entries := []string{}
				for i := range levels {
					q.Criteria = append(q.Criteria, json.RawMessage(`"same"`))
					p := 0
					if i == levels-1 {
						p = 1
					}
					entries = append(entries, fmt.Sprintf(`{"value":%d,"label":"same","probability":%d}`, i, p))
				}
				req := ai.ClassifierRequest{State: json.RawMessage(`{}`), Questions: map[string]ai.ClassifierQuestion{"q": q}}
				body := fmt.Sprintf(`{"answers":[{"type":"score","name":"q","score":%d,"confidence":1,"probabilities":[%s]}],%s}`, levels-1, strings.Join(entries, ","), decisionsUsage)
				enqueue(ev, u.world, fixtureReply{Status: 200, Body: body}.script(t, decisionsProtocol, ""))
				res, err := decisionsPublicCall(t, u, entry, fmt.Sprintf("req-score-bound-%d-%s", levels, entry), req, decisionsNoopHooks())
				ev.Record("result", res)
				ev.Record("error", errString(err))
				a, ok := res.Answers["q"].(ai.ScoreAnswer)
				ev.Check("public bounds and repeated labels retained", err == nil && ok && a.Score == float64(levels-1) && len(a.Probabilities) == levels && len(a.Legend) == levels, "got %+v / %v", a, err)
				u.records(ev, res, err)
				u.released(ev)
			})
		}
	}
}

func TestDecisionsMetadataBound(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "P10-OD39-metadata-bound-"+entry)
			u := newDecisionsMixedWorld(t, func(c *ai.Config) { c.Policy.MaxOutputBytes = 1024; c.Policy.MaxToolJSONBytes = 512 })
			body := `{"answers":[` + decisionsAnswer + `,` + decisionsScore + `,` + decisionsPredicate + `],"vendor_metadata":"` + strings.Repeat("x", 2048) + `"}`
			enqueue(ev, u.world, fixtureReply{Status: 200, Body: body}.script(t, decisionsProtocol, ""))
			res, err := decisionsPublicCall(t, u, entry, "req-metadata-bound-"+entry, decisionsMixedInput(), decisionsNoopHooks())
			checkUnaryFailure(ev, res, err, ai.CodeResourceLimit, ai.PhaseResponse, ai.StopReasonError)
			ev.Check("oversized 2xx not replayed", len(res.Metadata.Attempts) == 1 && len(u.provider.Requests()) == 1, "got %+v", res.Metadata)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

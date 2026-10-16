package e2e

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestCurrentCatalogRelease(t *testing.T) {
	for _, entry := range []string{"plain", "hooked"} {
		t.Run(entry, func(t *testing.T) {
			ev := run.Case(t, "UC14-current-release-"+entry)
			u := newJevBoxWorld(t, ev, func(c *ai.Config) { *c.Catalog = ai.BuiltinCatalog() })
			req := jevBoxInput(t)
			enqueueJevBox(t, ev, u, jevBoxAnswer, "0.9")
			var result ai.ClassifierResult
			var err error
			if entry == "plain" {
				result, err = u.client.Classify(ctxFor(t), textScope("req-release-"+entry), jevContextTarget(), req, jevContextOptions())
			} else {
				result, err = u.client.WithHooks(ai.Hooks{}).Classify(ctxFor(t), textScope("req-release-"+entry), jevContextTarget(), req, jevContextOptions())
			}
			ev.Check("current formal catalog executes image Box and existence", err == nil && len(result.Answers) == 2 && len(result.ImageInputs) == 1 && result.Metadata.CatalogHash == builtinCatalogHash, "public Classify failed: %v", err)
			recordJevContext(ev, u, req, result, err)
			u.released(ev)
			report := releaseProcess(t, ev)
			ev.Check("all eleven independent identities required", len(report.Live.Rows) == 11, "got %d rows", len(report.Live.Rows))
			for _, identity := range []struct{ combo, model string }{{"typesafe-classifier", "jev-1.13.0"}, {supportmatrix.JevFormalCombo, "jev-1.14.0"}, {"openai-decisions", "gpt-6-luna"}} {
				i := slices.IndexFunc(report.Live.Rows, func(r release.LiveRow) bool { return r.Combo == identity.combo })
				ev.Check("independent model "+identity.combo, i >= 0 && report.Live.Rows[i].Model == identity.model, "model missing or substituted")
			}
		})
	}
}

func TestCurrentReleaseTraceAndBlocker(t *testing.T) {
	ev := run.Case(t, "UC14-trace-and-external-blocker")
	report := releaseProcess(t, ev)
	for _, group := range []struct {
		prefix  string
		numbers []int
	}{
		{"R", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}},
		{"T", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}},
		{"US", []int{1, 3, 17, 23, 25, 32, 33, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50}},
	} {
		for _, number := range group.numbers {
			id := fmt.Sprintf("UC-%s%02d", group.prefix, number)
			ev.Check("required trace "+id, slices.ContainsFunc(report.Trace, func(r release.TraceResult) bool { return r.ID == id }), "missing trace")
		}
	}
	i := slices.IndexFunc(report.Snapshots, func(s release.Snapshot) bool { return s.Name == "pressure-stream-integrity" })
	ev.Check("historical causal-evidence blocker retained", i >= 0 && !report.Snapshots[i].Current && strings.Contains(report.Snapshots[i].Detail, "07-pressure-stream-integrity"), "unresolved historical blocker passed")
	for _, name := range []string{"multimodal-images-plain", "multimodal-images-hooked", "race-multimodal-images-plain", "race-multimodal-images-hooked"} {
		e := releaseEvidence(t, report, name)
		ev.Check("missing public-entry load never passes "+name, e.Status != release.Pass, "missing load passed")
	}
}

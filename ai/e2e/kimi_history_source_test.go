package e2e

import (
	"encoding/json"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Failure inventory is in kimi-history-supplement/spec.md. Exact reviewed
// transitions are bounded to Kimi; later product edits cannot inherit them.
func TestKimiHistoryReviewedSources(t *testing.T) {
	// Reuse the independent Git inventory of the pre-Imagine reviewed current
	// source. New checkout bytes must not redefine an old positive transition.
	historical := reviewedHistoricalSource(t)
	for _, route := range kimiQualificationRoutes {
		for _, mode := range []string{"reviewed", "current-unreviewed", "later-history", "changed-origin", "new-runtime", "dependency", "other-provider"} {
			t.Run(route+"/"+mode, func(t *testing.T) {
				ev := run.Case(t, "KHS-source-"+route+"-"+mode)
				var f struct {
					Provider, API string
					Source        map[string]string
				}
				raw, e := os.ReadFile(filepath.Join("testdata/kimi-history-source", route+".json"))
				if e != nil || json.Unmarshal(raw, &f) != nil {
					t.Fatal(e)
				}
				cur := maps.Clone(historical.Files)
				if mode == "current-unreviewed" {
					cur, e = supportmatrix.RegionalExecutionSource()
					if e != nil {
						t.Fatal(e)
					}
					for n := range cur {
						if strings.HasPrefix(n, "ai/internal/testkit/") || strings.HasPrefix(n, "ai/live/") {
							delete(cur, n)
						}
					}
				}
				switch mode {
				case "later-history":
					cur["ai/history.go"] = strings.Repeat("1", 64)
				case "changed-origin":
					f.Source["ai/history.go"] = strings.Repeat("2", 64)
				case "new-runtime":
					cur["ai/unreviewed/history.go"] = strings.Repeat("3", 64)
				case "dependency":
					cur["go.mod"] = strings.Repeat("4", 64)
				case "other-provider":
					f.Provider = "deepseek"
				}
				e = evidence.VerifyRouteSource(f.Source, cur, f.Provider, f.API)
				ev.Check("only reviewed original route bytes apply", (e == nil) == (mode == "reviewed"), "%v", e)
				ev.Record("source-check", map[string]any{"route": route, "mode": mode, "error": errString(e), "vendorCalls": 0, "historicalCurrentCommit": historical.SourceCommit})
			})
		}
	}
}

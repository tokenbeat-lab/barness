package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/release"
)

func TestGLMReleaseTraceContract(t *testing.T) {
	for _, mode := range []string{"complete", "missing-requirement", "missing-acceptance", "missing-story", "missing-required-case"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "P0-GLM13-release-trace-"+mode)
			raw, err := os.ReadFile("../release/traceability.json")
			if err != nil {
				t.Fatal(err)
			}
			var trace map[string]any
			mustUnmarshal(t, raw, &trace)
			wanted := map[string]string{"missing-requirement": "GLM-R01", "missing-acceptance": "GLM-A15", "missing-story": "GLM-US62", "missing-required-case": "GLM-R04"}[mode]
			items := trace["items"].([]any)
			out := []any{}
			for _, item := range items {
				m := item.(map[string]any)
				if m["id"] == wanted {
					if mode == "missing-required-case" {
						m["requiredCases"] = []any{}
					} else {
						continue
					}
				}
				out = append(out, m)
			}
			trace["items"] = out
			dir := t.TempDir()
			writeReleaseJSON(t, dir, "trace.json", trace)
			report := releaseProcess(t, ev, "-trace", filepath.Join(dir, "trace.json"))
			plan := slices.IndexFunc(report.Snapshots, func(s release.Snapshot) bool { return s.Name == "release-plan" })
			ev.Check("complete additive GLM contract is fixed", plan >= 0 && report.Snapshots[plan].Current == (mode == "complete"), "mode=%s snapshot=%+v", mode, report.Snapshots)
			if mode == "complete" {
				for prefix, n := range map[string]int{"GLM-R": 22, "GLM-A": 15, "GLM-US": 62} {
					for i := 1; i <= n; i++ {
						id := fmt.Sprintf("%s%02d", prefix, i)
						idx := slices.IndexFunc(report.Trace, func(it release.TraceResult) bool { return it.ID == id })
						ev.Check("required actual trace "+id, idx >= 0 && report.Trace[idx].Status != release.Pass, "missing or false accepting %s", id)
					}
				}
			}
		})
	}
}

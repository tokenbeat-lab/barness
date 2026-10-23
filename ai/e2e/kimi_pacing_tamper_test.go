package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func kimiPacingTamper(t *testing.T, original, combo, consumer string) {
	t.Helper()
	for _, mode := range []string{"report-interval", "receipt-interval", "initial-gap", "later-gap"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+combo+"-K16-pacing-forgery-"+mode)
			dir := filepath.Join(t.TempDir(), filepath.Base(original))
			if err := os.CopyFS(dir, os.DirFS(original)); err != nil {
				t.Fatal(err)
			}
			file := supportmatrix.RegionalReportFile
			caseID := ""
			if mode == "receipt-interval" {
				file = "regional-authorization-receipt.json"
			}
			if mode == "initial-gap" || mode == "later-gap" {
				kind := "full-stream"
				previous := time.Time{}
				if mode == "later-gap" {
					kind = "full-complete"
					raw, _ := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalCaseID(combo, "kimi-k3/full-stream"), "observations.json"))
					var before []ai.Observation
					mustUnmarshal(t, raw, &before)
					previous = before[3].Time
				} else {
					raw, _ := os.ReadFile(filepath.Join(dir, "regional-authorization-receipt.json"))
					var origin supportmatrix.RegionalReservation
					mustUnmarshal(t, raw, &origin)
					previous = origin.StartedAt
				}
				caseID = supportmatrix.RegionalCaseID(combo, "kimi-k3/"+kind)
				file = caseID + "/observations.json"
				raw, _ := os.ReadFile(filepath.Join(dir, file))
				var obs []ai.Observation
				mustUnmarshal(t, raw, &obs)
				obs[0].Time = previous.Add(time.Millisecond)
				updated := mustMarshal(t, obs)
				if err := os.WriteFile(filepath.Join(dir, file), updated, 0600); err != nil {
					t.Fatal(err)
				}
				kimiQualificationManifest(t, dir, caseID, "observations.json", updated)
			} else {
				raw, _ := os.ReadFile(filepath.Join(dir, file))
				var v map[string]any
				mustUnmarshal(t, raw, &v)
				if mode == "report-interval" {
					v["regional"].(map[string]any)["callIntervalMs"] = 0
				} else {
					v["callIntervalMs"] = 0
				}
				if err := os.WriteFile(filepath.Join(dir, file), mustMarshal(t, v), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := supportmatrix.SealRegionalBundle(dir); err != nil {
				t.Fatal(err)
			}
			cmd := exec.CommandContext(ctxFor(t), consumer, "-controlled", dir)
			cmd.Dir, cmd.Env = "../..", regionalChildEnvironment()
			raw, err := cmd.CombinedOutput()
			ev.Record("consumer", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw)})
			ev.Check("resealed timing or interval contradiction refuses", err != nil, "%s", regionalSafeOutput(raw))
		})
	}
}

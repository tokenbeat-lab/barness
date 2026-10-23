package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestKimiQualificationPacing(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	consumer := kimiQualificationConsumer(t)
	for _, sc := range []struct {
		combo    string
		interval int
	}{{"kimi-intl-chat", 0}, {"kimi-intl-chat", 40}, {"kimi-cn-messages", 40}} {
		t.Run(sc.combo+"/"+(time.Duration(sc.interval)*time.Millisecond).String(), func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+sc.combo+"-K16-pacing-"+(time.Duration(sc.interval)*time.Millisecond).String())
			route, _ := supportmatrix.RegionalRouteOf(sc.combo)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, route.API, kind))
			}
			grant := kimiPacingGrant(t, sc.combo, digest, sc.combo+"-"+(time.Duration(sc.interval)*time.Millisecond).String(), sc.interval)
			cmd, base := kimiPacingCommand(t, binary, sc.combo, srv.URL(), grant)
			raw, err := cmd.CombinedOutput()
			ev.Record("native-process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			ev.Record("requests", srv.Requests())
			if !ev.Check("same nine logical calls and HTTP attempts", err == nil && len(srv.Requests()) == 9, "%v %s", err, regionalSafeOutput(raw)) {
				return
			}
			dir := kimiPacingBundle(t, base)
			var report supportmatrix.Report
			b, _ := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile))
			mustUnmarshal(t, b, &report)
			ev.Record("native-report", json.RawMessage(b))
			var fields map[string]any
			mustUnmarshal(t, b, &fields)
			value := fields["regional"].(map[string]any)["callIntervalMs"]
			if value == nil {
				value = float64(0)
			}
			ev.Check("report binds the selected interval", value == float64(sc.interval), "got=%v want=%d", value, sc.interval)
			b, _ = os.ReadFile(filepath.Join(dir, "regional-authorization-receipt.json"))
			var origin supportmatrix.RegionalReservation
			mustUnmarshal(t, b, &origin)
			previous := origin.StartedAt
			for _, row := range report.Scenarios {
				b, _ = os.ReadFile(filepath.Join(dir, row.CaseID, "observations.json"))
				var obs []ai.Observation
				mustUnmarshal(t, b, &obs)
				if !ev.Check("four native observations", len(obs) == 4, "case=%s", row.ID) {
					continue
				}
				ev.Check("initial and inter-call pause is real", !obs[0].Time.Before(previous.Add(time.Duration(sc.interval)*time.Millisecond)), "case=%s gap=%s interval=%dms", row.ID, obs[0].Time.Sub(previous), sc.interval)
				previous = obs[3].Time
				ev.Check("per-call network duration excludes pacing", row.DurationMS <= 120000 && row.Calls == 1 && row.Attempts == 1 && row.Retries == 0, "%+v", row)
			}
			verify := exec.CommandContext(ctxFor(t), consumer, "-controlled", dir)
			verify.Dir, verify.Env = "../..", regionalChildEnvironment()
			vr, ve := verify.CombinedOutput()
			ev.Record("consumer", map[string]any{"argv": verify.Args, "error": errString(ve), "output": regionalSafeOutput(vr)})
			ev.Check("native consumer accepts actual paced capture", ve == nil, "%v %s", ve, regionalSafeOutput(vr))
			if sc.interval > 0 {
				kimiPacingTamper(t, dir, sc.combo, consumer)
			}
		})
	}
}

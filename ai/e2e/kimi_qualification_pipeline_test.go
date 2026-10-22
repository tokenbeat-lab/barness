package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestKimiQualificationPipeline(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	consumer := kimiQualificationConsumer(t)
	for _, combo := range kimiQualificationRoutes {
		t.Run(combo, func(t *testing.T) {
			ev := run.Case(t, "KIMI-"+combo+"-K16-controlled-pipeline")
			route, _ := supportmatrix.RegionalRouteOf(combo)
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, route.API, kind))
			}
			grant, err := supportmatrix.RegionalGrantSnapshot(combo, "candidate")
			if err != nil {
				t.Fatal(err)
			}
			grant.Mode, grant.ID, grant.AccountAlias, grant.RegionAlias = "controlled", "kimi-"+combo+"-"+evidence.Digest([]byte(run.Dir()))[:12], "synthetic-kimi-account", "synthetic-kimi-region"
			grant.ExpiresAt = time.Now().UTC().Add(time.Hour)
			grant.MaxUSD = 10
			grant.BinarySHA256 = digest
			grantPath := filepath.Join(t.TempDir(), "grant.json")
			if err := os.WriteFile(grantPath, mustMarshal(t, grant), 0600); err != nil {
				t.Fatal(err)
			}
			base := filepath.Join(filepath.Dir(run.Dir()), "kimi-native", filepath.Base(run.Dir()), combo)
			cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestRegionalCandidate$", "-test.timeout=1800s")
			cmd.Dir = "../.."
			cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+combo, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+route.Provider, "BARNESS_REGIONAL_GRANT="+grantPath, "BARNESS_AI_EVIDENCE_DIR="+base)
			raw, err := cmd.CombinedOutput()
			ev.Record("native-process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
			ev.Record("requests", srv.Requests())
			if !ev.Check("nine actual own native scenarios execute", err == nil && len(srv.Requests()) == 9, "%v %s", err, regionalSafeOutput(raw)) {
				return
			}
			paths, err := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
			if !ev.Check("one original native capture", err == nil && len(paths) == 1, "%v %d", err, len(paths)) {
				return
			}
			dir := filepath.Dir(paths[0])
			verify := exec.CommandContext(ctxFor(t), consumer, "-controlled", dir)
			verify.Dir, verify.Env = "../..", regionalChildEnvironment()
			vraw, verr := verify.CombinedOutput()
			ev.Record("consumer-process", map[string]any{"argv": verify.Args, "error": errString(verr), "output": regionalSafeOutput(vraw)})
			ev.Check("read-only consumer validates counters native prefix and bounded costs", verr == nil, "%v %s", verr, regionalSafeOutput(vraw))
			var report supportmatrix.Report
			data, _ := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile))
			mustUnmarshal(t, data, &report)
			ev.Check("finite contract retains worst-one-hour spend reserve", report.Budget.MaxCalls == 9 && report.Budget.CallsUsed == 9 && report.Regional.MaxUSD == 10 && report.Regional.ReservedUSD > 0 && report.Regional.ReservedUSD <= 10, "%+v", report)
			_, err = supportmatrix.VerifyRegionalQualification(dir)
			ev.Check("controlled never grants actual live support", err != nil, "controlled promoted")
			_, err = supportmatrix.NewRegionalReport(dir)
			ev.Check("candidate never grants formal support", err != nil, "candidate promoted")
			again := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			again.Dir, again.Env = cmd.Dir, cmd.Env
			araw, aerr := again.CombinedOutput()
			ev.Check("one-use authorization prevents another request", aerr != nil && len(srv.Requests()) == 9, "%v %s", aerr, regionalSafeOutput(araw))
			kimiQualificationTamper(t, dir, combo, consumer)
			grant.ID += "-relabeled"
			if err := os.WriteFile(grantPath, mustMarshal(t, grant), 0600); err != nil {
				t.Fatal(err)
			}
			relabeled := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			relabeled.Dir, relabeled.Env = cmd.Dir, cmd.Env
			rr, rer := relabeled.CombinedOutput()
			ev.Check("same scope cannot regain authority by changing ID", rer != nil && len(srv.Requests()) == 9, "%s", regionalSafeOutput(rr))
			ev.Record("native-pointer", map[string]string{"manifest": paths[0], "sha256": evidence.Digest(data)})
		})
	}
}

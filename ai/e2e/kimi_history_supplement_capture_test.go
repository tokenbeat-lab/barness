package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

// All authorities in these process fixtures are controlled and private. Failure
// inventory: exact entry/timeout, one request, complete raw proof, no standalone
// ten-case support, no grant reuse or reinterpretation as a full run.
func TestKimiHistorySupplementCapture(t *testing.T) {
	kimiHistoryPrivateStore(t)
	binary, digest := kimiQualificationBinary(t)
	for _, route := range kimiQualificationRoutes {
		t.Run(route, func(t *testing.T) {
			ev := run.Case(t, "KHS-capture-"+route)
			dir, srv, cmd := kimiHistoryCapture(t, ev, binary, digest, route, strings.Repeat("a", 64), false)
			defer srv.Close()
			if dir == "" {
				return
			}
			var r supportmatrix.Report
			raw, err := os.ReadFile(filepath.Join(dir, supportmatrix.RegionalReportFile))
			if err != nil {
				t.Fatal(err)
			}
			mustUnmarshal(t, raw, &r)
			ev.Check("only fixed selected scene measured", len(r.Scenarios) == 1 && r.Scenarios[0].ID == "kimi-k3/xai-origin-history" && r.Budget.CallsUsed == 1 && r.Budget.HTTPAttempts == 1 && r.Budget.MaxReservedInputTokens == 4096 && r.Budget.MaxRequestedOutputTokens == 4096 && r.Budget.ProcessSeconds == 240, "%+v", r.Budget)
			ev.Check("single scene never claims full qualification", r.Regional.Purpose == supportmatrix.KimiHistoryPurpose && !r.Regional.Complete, "%+v", r.Regional)
			_, err = supportmatrix.VerifyRegionalControlledCapture(dir)
			ev.Check("full consumer rejects one scene", err != nil, "%v", err)
			again := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
			again.Dir, again.Env = cmd.Dir, cmd.Env
			out, err := again.CombinedOutput()
			ev.Record("repeat", map[string]any{"error": errString(err), "output": regionalSafeOutput(out)})
			ev.Check("consumed synthetic approval cannot send again", err != nil && len(srv.Requests()) == 1, "%v", err)
		})
	}
}

func kimiHistoryCapture(t *testing.T, ev *evidence.Case, binary, digest, route, baseline string, expectFailure bool, replies ...provider.Reply) (string, *provider.Server, *exec.Cmd) {
	t.Helper()
	info, _ := supportmatrix.RegionalRouteOf(route)
	srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
	if len(replies) == 0 {
		srv.Enqueue(kimiQualificationReply(t, info.API, "xai-origin-history"))
	} else {
		for _, reply := range replies {
			srv.Enqueue(reply)
		}
	}
	g, err := supportmatrix.KimiHistoryGrantSnapshot(route, baseline)
	if err != nil {
		t.Fatal(err)
	}
	g.Mode, g.ID, g.AccountAlias, g.RegionAlias = "controlled", "history-"+route+"-"+evidence.Digest([]byte(t.Name() + run.Dir()))[:12], "synthetic-kimi-account", "synthetic-kimi-region"
	g.ExpiresAt = time.Now().UTC().Add(time.Hour)
	g.MaxUSD = 0.10
	g.BinarySHA256 = digest
	g.CallIntervalMs = 0
	file := filepath.Join(t.TempDir(), "grant.json")
	base := filepath.Join(filepath.Dir(run.Dir()), "kimi-history-native", filepath.Base(run.Dir()), g.ID)
	binary = kimiRetainedBinary(t, binary, "single")
	cmd := exec.CommandContext(ctxFor(t), binary, "-test.run=^TestKimiHistorySupplement$", "-test.timeout=240s")
	cmd.Dir = "../.."
	cmd.Env = regionalChildEnvironment("BARNESS_REGIONAL_ROUTE="+route, "BARNESS_REGIONAL_CONTROLLED="+srv.URL(), "BARNESS_REGIONAL_SERVICE="+info.Provider, "BARNESS_REGIONAL_GRANT="+file, "BARNESS_AI_EVIDENCE_DIR="+base)
	if !privateGrantProcess(t, ev, cmd) {
		return "", srv, cmd
	}
	if err := os.WriteFile(file, mustMarshal(t, g), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	started := time.Now().UTC()
	err = cmd.Run()
	finished := time.Now().UTC()
	raw := append(append([]byte{}, stdout.Bytes()...), stderr.Bytes()...)
	ev.Record("native", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(raw), "vendorCalls": 0})
	ev.Record("requests", srv.Requests())
	if expectFailure {
		ev.Check("failed single scene consumes one request without retry", err != nil && len(srv.Requests()) == 1, "%s requests=%d", regionalSafeOutput(raw), len(srv.Requests()))
		return "", srv, cmd
	}
	if !ev.Check("one controlled actual request", err == nil && len(srv.Requests()) == 1, "%v %s", err, regionalSafeOutput(raw)) {
		return "", srv, cmd
	}
	paths, e := filepath.Glob(filepath.Join(base, "*", "manifest.json"))
	if e != nil || len(paths) != 1 {
		t.Fatal("native bundle missing", e)
	}
	dir := filepath.Dir(paths[0])
	kimiProducerReceipt(t, cmd, dir, stdout.Bytes(), stderr.Bytes(), started, finished)
	ev.Record("native-pointer", map[string]string{"bundle": dir})
	return dir, srv, cmd
}

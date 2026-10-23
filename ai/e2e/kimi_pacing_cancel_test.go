package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
)

func TestKimiPacingInterruptedWait(t *testing.T) {
	binary, digest := kimiQualificationBinary(t)
	for _, mode := range []string{"expiry", "signal"} {
		t.Run(mode, func(t *testing.T) {
			ev := run.Case(t, "KIMI-K16-pacing-initial-"+mode)
			combo := "kimi-intl-chat"
			srv := provider.New(map[string]string{"synthetic-regional-controlled": "key:kimi-controlled"})
			defer srv.Close()
			for _, kind := range kimiQualificationKinds() {
				srv.Enqueue(kimiQualificationReply(t, "openai-completions", kind))
			}
			grant := kimiPacingGrant(t, combo, digest, mode, 60000)
			if mode == "expiry" {
				grant["expiresAt"] = time.Now().UTC().Add(3 * time.Second)
			}
			cmd, base := kimiPacingCommand(t, binary, combo, srv.URL(), grant)
			var buf bytes.Buffer
			cmd.Stdout, cmd.Stderr = &buf, &buf
			began := time.Now()
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			if mode == "signal" {
				deadline := time.Now().Add(5 * time.Second)
				ready := false
				for time.Now().Before(deadline) {
					paths, _ := filepath.Glob(filepath.Join(base, "*", "regional-authorization-receipt.json"))
					if len(paths) == 1 {
						ready = true
						break
					}
					time.Sleep(time.Millisecond)
				}
				ev.Check("grant consumed before cancellation", ready, "base=%s", base)
				_ = cmd.Process.Signal(os.Interrupt)
			}
			err := <-done
			ev.Record("process", map[string]any{"argv": cmd.Args, "error": errString(err), "output": regionalSafeOutput(buf.Bytes()), "elapsed": time.Since(began).String()})
			ev.Record("requests", srv.Requests())
			ev.Check("interrupted wait starts no HTTP", err != nil && len(srv.Requests()) == 0 && time.Since(began) < 6*time.Second, "%v requests=%d", err, len(srv.Requests()))
			origins, _ := filepath.Glob(filepath.Join(base, "*", "regional-authorization-receipt.json"))
			ev.Check("valid grant was durably consumed before wait ended", len(origins) == 1, "origins=%d", len(origins))
			paths, _ := filepath.Glob(filepath.Join(base, "*", supportmatrix.RegionalReportFile))
			if !ev.Check("graceful wait interruption retains report", len(paths) == 1, "reports=%d", len(paths)) {
				return
			}
			raw, _ := os.ReadFile(paths[0])
			var report supportmatrix.Report
			mustUnmarshal(t, raw, &report)
			ev.Record("native-report", report)
			ev.Check("wait spends no call token or USD reservation", report.Budget.CallsUsed == 0 && report.Budget.HTTPAttempts == 0 && report.Budget.AttemptsReserved == 0 && report.Budget.RequestedOutputTokens == 0 && report.Budget.ReservedInputTokens == 0 && report.Regional.ReservedUSD == 0, "%+v", report.Budget)
			if mode == "signal" {
				again := exec.CommandContext(ctxFor(t), cmd.Path, cmd.Args[1:]...)
				again.Dir, again.Env = cmd.Dir, cmd.Env
				raw, e := again.CombinedOutput()
				ev.Record("replay", map[string]any{"error": errString(e), "output": regionalSafeOutput(raw)})
				ev.Check("canceled wait permanently consumes still-valid grant", e != nil && len(srv.Requests()) == 0, "%v", e)
			}
		})
	}
}

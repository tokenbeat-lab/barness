package minimaxlive_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimaxlive"
)

// The real closed owner is read only. This seam never creates a current grant,
// reads a credential, acquires a nonce or contacts a provider.
func TestRecovery7CurrentKeylessCLI(t *testing.T) {
	if os.Getenv("BARNESS_RECOVERY7_REAL_HISTORY") != "1" {
		t.Skip("explicit read-only retained history verification")
	}
	root, e := minimaxlive.Root()
	if e != nil {
		t.Fatal(e)
	}
	cli := filepath.Join(t.TempDir(), "minimaxverify")
	c := exec.Command("go", "build", "-o", cli, "./ai/live/cmd/minimaxverify")
	c.Dir = root
	if raw, e := c.CombinedOutput(); e != nil {
		t.Fatalf("build %v %s", e, raw)
	}
	c = exec.Command(cli, "-recovery7-source")
	c.Env = []string{"PATH=" + os.Getenv("PATH")}
	raw, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("current keyless recovery7 facts: %v %s", e, raw)
	}
	var f map[string]any
	if json.Unmarshal(raw, &f) != nil || f["renewalId"] != "provider-sites-final-acceptance-7" {
		t.Fatalf("fixed7 facts identity differs: %s", raw)
	}
	h, ok := f["sixthCohort"].(map[string]any)
	if !ok || h["retainedCompletedReportKnownCalls"] != float64(720) || h["retainedCompletedReportKnownHTTP"] != float64(710) || h["retainedKnownRecordedCalls"] != float64(724) || h["retainedKnownRecordedHTTPAllocations"] != float64(714) || h["retainedClaimedCalls"] != float64(1204) || h["retainedClaimedAttempts"] != float64(1280) || h["retainedReservedMicrousd"] != float64(32900000) || h["retainedUnknownParentProcesses"] != float64(6) || h["interruptedNativeCalls"] != float64(4) || h["interruptedNativeHTTPAllocations"] != float64(4) || h["interruptedUSDReservationMicrousd"] != float64(12884) || h["interruptedUSDTransportReservations"] != float64(4) {
		t.Fatalf("actual retained sixth counters differ: %s", raw)
	}
	t.Logf("PASS keyless current7 facts; closed6 history=%s", f["sixthCohortHistorySha256"])
	for _, route := range []string{"chat", "responses", "messages"} {
		c = exec.Command(cli, "-recovery7-preflight="+route)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		raw, e = c.CombinedOutput()
		if e == nil || !strings.Contains(string(raw), "grant") && !strings.Contains(string(raw), "authority") {
			t.Fatalf("missing independent7 authorization was not rejected: %v %s", e, raw)
		}
	}
	if _, _, e = minimaxlive.RenewalAuthority("provider-sites-final-acceptance-6", true); e == nil {
		t.Fatal("closed6 US39 authorization reopened execution")
	}
}

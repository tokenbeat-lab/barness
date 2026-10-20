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
func TestRecovery6CurrentKeylessCLI(t *testing.T) {
	if os.Getenv("BARNESS_RECOVERY6_REAL_HISTORY") != "1" {
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
	c = exec.Command(cli, "-recovery6-source")
	c.Env = []string{"PATH=" + os.Getenv("PATH")}
	raw, e := c.CombinedOutput()
	if e != nil {
		t.Fatalf("current keyless recovery6 facts: %v %s", e, raw)
	}
	var f map[string]any
	if json.Unmarshal(raw, &f) != nil || f["renewalId"] != "provider-sites-final-acceptance-6" {
		t.Fatalf("fixed6 facts identity differs: %s", raw)
	}
	h, ok := f["fifthCohort"].(map[string]any)
	if !ok || h["retainedActualCalls"] != float64(702) || h["retainedActualAttempts"] != float64(692) || h["retainedClaimedCalls"] != float64(1160) || h["retainedClaimedAttempts"] != float64(1232) || h["retainedReservedMicrousd"] != float64(32050000) || h["retainedUnknownParentProcesses"] != float64(5) || h["nativeCalls"] != float64(39) || h["nativeAttempts"] != float64(36) || h["failedUSDReservationMicrousd"] != float64(54238) || h["failedUSDTransportReservations"] != float64(9) {
		t.Fatalf("actual retained fifth counters differ: %s", raw)
	}
	t.Logf("PASS keyless current6 facts; closed5 history=%s", f["fifthCohortHistorySha256"])
	for _, route := range []string{"chat", "responses", "messages"} {
		c = exec.Command(cli, "-recovery6-preflight="+route)
		c.Env = []string{"PATH=" + os.Getenv("PATH")}
		raw, e = c.CombinedOutput()
		if e == nil || !strings.Contains(string(raw), "grant") && !strings.Contains(string(raw), "authority") {
			t.Fatalf("missing independent6 authorization was not rejected: %v %s", e, raw)
		}
	}
	if _, _, e = minimaxlive.RenewalAuthority("provider-sites-final-acceptance-5", true); e == nil {
		t.Fatal("closed5 US36 authorization reopened execution")
	}
}

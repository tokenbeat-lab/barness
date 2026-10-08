package e2e

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
)

// TestPressureLoopback repeats the exact release design load with a fresh
// fixture and Client per round, inside one process. Each round has its own
// evidence directory so a later success cannot overwrite an earlier failure.
// Run standalone to exclude preceding-suite state, or alongside the suite to
// include it. Separate invocations distinguish process residue from loopback.
func TestPressureLoopback(t *testing.T) {
	const env = "BARNESS_AI_PRESSURE_LOOPS"
	value := os.Getenv(env)
	if value == "" {
		t.Skip("diagnostic opt-in: " + env + "=100")
	}
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 1000 {
		t.Fatal(env + " must be an integer from 1 to 1000")
	}
	var baseline uint64
	for i := range n {
		t.Run(fmt.Sprintf("round-%03d", i), func(t *testing.T) {
			ev := pressureCase(t, fmt.Sprintf("E08-pressure-loopback-%03d", i))
			ev.ReplayEnv(pressureEnv + "=1 " + env + "=" + value)
			// Later-round failures depend on earlier fixtures and heap state.
			ev.ReplayRun("^TestPressureLoopback$")
			r := runDesignLoad(t, ev, *hostintegration.CloudInteractivePolicy(), cloudInteractiveDesignLoad(), 2<<30)
			ev.Record("policy-pressure", r)
			if i == 0 {
				baseline = r.Heap.BaselineLive
			}
			growth := r.Heap.BaselineLive - min(r.Heap.BaselineLive, baseline)
			ev.Check("completed fixtures do not accumulate beyond the unchanged heap budget", growth <= 2<<30, "baseline grew %d MiB", growth>>20)
		})
		if t.Failed() {
			// Keep the first failing round and fail the process; no retries
			// or successful tail may replace the diagnostic signal.
			break
		}
	}
}

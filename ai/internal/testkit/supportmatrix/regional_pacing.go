package supportmatrix

import (
	"github.com/tokenbeat-lab/barness/ai"
	"os"
	"time"
)

// The grant and durable authorization receipt bind configuration; public
// Observer timestamps independently prove that the captured invocations do
// not contradict it. The first interval starts no earlier than consumption.
func verifyRegionalPacing(root *os.Root, r Report, g RegionalGrant, origin RegionalReservation) error {
	if r.Regional.CallIntervalMs != g.CallIntervalMs || origin.CallIntervalMs != g.CallIntervalMs {
		return refuse("regional captured call interval differs from original grant")
	}
	if g.CallIntervalMs == 0 {
		return nil
	}
	interval := time.Duration(g.CallIntervalMs) * time.Millisecond
	previous := origin.StartedAt
	for _, s := range r.Scenarios {
		raw, err := regionalRead(root, s.CaseID+"/observations.json")
		if err != nil {
			return err
		}
		var obs []ai.Observation
		if regionalDecode(raw, &obs) != nil || len(obs) != 4 || obs[0].Kind != ai.ObservationCallStarted || obs[3].Kind != ai.ObservationCallFinished {
			return refuse("regional paced call timestamps unavailable")
		}
		start, finish := obs[0].Time, obs[3].Time
		if start.IsZero() || start.Before(previous.Add(interval)) || finish.Before(start) || finish.After(r.FinishedAt) {
			return refuse("regional captured invocation contradicts granted call interval")
		}
		previous = finish
	}
	return nil
}

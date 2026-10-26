package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type xaiRuntimeWorld struct {
	*world
	gauges *probe.Probe
	bodies *provider.BodyTracker
	adm    *host.Admission
}

func newXAIRuntimeWorld(t *testing.T, mutate func(*ai.Catalog), configure ...func(*ai.Config)) xaiRuntimeWorld {
	t.Helper()
	r := xaiRuntimeWorld{gauges: probe.New(), adm: host.NewAdmission()}
	r.world = newXAIWorld(t, mutate, func(c *ai.Config) {
		r.bodies = provider.TrackBodies(c.Transport)
		c.Transport, c.Probe, c.Admission = r.bodies, r.gauges, r.adm
		for _, f := range configure {
			f(c)
		}
	})
	return r
}

func (r xaiRuntimeWorld) released(ev *evidence.Case, label ...string) {
	eventually(ev, "producer has settled", func() bool { return r.gauges.ActiveCalls() == 0 })
	state := map[string]int64{"calls": r.gauges.ActiveCalls(), "permits": r.gauges.Permits(), "queued-events": r.gauges.QueuedEvents(), "waiters": r.gauges.AdmissionWaiters(), "bodies": r.bodies.Open(), "host-permits": int64(r.adm.Held())}
	name := "resources"
	if len(label) > 0 {
		name += "-" + label[0]
	}
	ev.Record(name, state)
	for name, n := range state {
		ev.Check("released "+name, n == 0, "%v", state)
	}
}

func xaiHTTPFailure(status int) provider.Reply {
	body := `{"error":{"code":"private-business-code","message":"private-prompt-and-ciphertext"}}`
	return provider.Reply{Status: status, Header: map[string]string{"Content-Type": "application/json", "Retry-After-Ms": "1", "X-Request-Id": "unverified-xai-id", "Request-Id": "unverified-other-id"}, Script: body, Chunks: [][]byte{[]byte(body)}}
}

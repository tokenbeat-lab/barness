package e2e

import (
	"context"
	"net/http"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// GLM recovery must not widen existing providers' synchronous Host contract.
// The original unary panic/cleanup tests remain unchanged and run alongside it.
func TestGLMLegacyCallbackPanic(t *testing.T) {
	for _, p := range []*fixtureProtocol{chatProtocol, deepseekChatProtocol} {
		for _, stage := range []string{"headers", "payload", "response"} {
			t.Run(p.dir+"/"+stage, func(t *testing.T) {
				ev := run.Case(t, "P0-GLM13-legacy-panic-"+p.dir+"-"+stage)
				f, raw := loadFixture(t, p, "text.json")
				sc := f.Scenarios[0]
				ev.Fixture("own-wire.json", raw)
				g := probe.New()
				bodies := provider.TrackBodies(provider.LoopbackTransport())
				w := newWorldWith(t, func(c *ai.Config) { c.Probe = g; c.Transport = bodies }, tenantA)
				enqueue(ev, w, sc.replies(t)...)
				hooks := ai.Hooks{}
				switch stage {
				case "headers":
					hooks.TransformHeaders = func(context.Context, ai.CallScope, http.Header) error { panic("synthetic Host panic") }
				case "payload":
					hooks.OnPayload = func(context.Context, ai.CallScope, *ai.Payload) (ai.PayloadDecision, error) {
						panic("synthetic Host panic")
					}
				case "response":
					hooks.OnResponse = func(context.Context, ai.CallScope, ai.ResponseInfo) error { panic("synthetic Host panic") }
				}
				var recovered any
				func() {
					defer func() { recovered = recover() }()
					_, _ = w.client.WithHooks(hooks).Complete(ctxFor(t), textScope("glm13-legacy-panic"), sc.target(), sc.request(t), nil)
				}()
				ev.Record("panic-recovered", recovered == "synthetic Host panic")
				ev.Record("requests", w.provider.Requests())
				ev.Record("resources", map[string]int64{"bodies": bodies.Open(), "calls": g.ActiveCalls(), "permits": g.Permits(), "events": g.QueuedEvents(), "waiters": g.AdmissionWaiters()})
				ev.Check("old synchronous panic reaches its Host", recovered == "synthetic Host panic", "got %v", recovered)
				count := 0
				if stage == "response" {
					count = 1
				}
				ev.Check("no hidden request or retained resources", len(w.provider.Requests()) == count && bodies.Open() == 0 && g.ActiveCalls() == 0 && g.Permits() == 0 && g.QueuedEvents() == 0 && g.AdmissionWaiters() == 0, "requests=%d", len(w.provider.Requests()))
			})
		}
	}
}

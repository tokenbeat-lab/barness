package e2e

import (
	"context"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// TestEventEnvelope covers E10's per-event attribution (spec I2): every event
// a Stream publishes comes with the immutable attribution of its call, so a
// host merging streams never infers ownership from arrival order. A resolved
// call's events name the tenant, request, binding and what actually serves
// it; a call refused before resolution reports Resolved=false and echoes no
// requested model as authorized. Merging streams of several tenants is the
// host example's (TestHostIntegrationExample).
func TestEventEnvelope(t *testing.T) {
	f, raw := loadIsolationFixture(t)
	streams := []struct {
		name  string
		start func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope) *ai.Stream
	}{
		{"stream-full", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope) *ai.Stream {
			return w.client.Stream(ctx, scope, p.target(), p.request(), nil)
		}},
		{"stream-simple", func(ctx context.Context, w *world, p isolationProtocol, scope ai.CallScope) *ai.Stream {
			return w.client.StreamSimple(ctx, scope, p.target(), p.request(), ai.SimpleOptions{})
		}},
	}
	for _, p := range f.Protocols {
		for _, e := range streams {
			t.Run(p.ID+"/success/"+e.name, func(t *testing.T) {
				ev := run.Case(t, protocolCase(p)+"-E10-envelope-success-"+e.name)
				ev.Fixture("interleave.json", raw)
				w := newWorld(t, tenantA)
				enqueue(ev, w, p.success(t, tenantA))
				scope := textScope("req-e10-envelope-" + p.ID + "-" + e.name)
				envs, res, err := drainEnvelopes(ev, e.start(ctxFor(t), w, p, scope))
				ev.Check("call succeeds", err == nil, "err=%v", err)
				want := ai.CallAttribution{TenantID: tenantA.tenant, RequestID: scope.RequestID, ActorID: scope.ActorID, JobID: scope.JobID,
					BindingID: p.BindingID, Resolved: true, ProviderID: p.provider(), API: p.api(), ModelID: p.Model, AccountScopeID: p.account(tenantA)}
				checkEnvelopes(ev, envs, res, want)
			})

			t.Run(p.ID+"/failure-after-resolution/"+e.name, func(t *testing.T) {
				ev := run.Case(t, protocolCase(p)+"-E10-envelope-upstream-401-"+e.name)
				ev.Fixture("interleave.json", raw)
				w := newWorld(t, tenantA)
				enqueue(ev, w, p.script(tenantA).Unauthorized.reply())
				scope := textScope("req-e10-envelope-401-" + p.ID + "-" + e.name)
				envs, res, err := drainEnvelopes(ev, e.start(ctxFor(t), w, p, scope))
				ev.Check("call fails upstream", err != nil, "err=%v", err)
				want := ai.CallAttribution{TenantID: tenantA.tenant, RequestID: scope.RequestID, ActorID: scope.ActorID, JobID: scope.JobID,
					BindingID: p.BindingID, Resolved: true, ProviderID: p.provider(), API: p.api(), ModelID: p.Model, AccountScopeID: p.account(tenantA)}
				checkEnvelopes(ev, envs, res, want)
			})

			t.Run(p.ID+"/refused-before-resolution/"+e.name, func(t *testing.T) {
				// Tenant B has no bindings in this world: the requested
				// binding and model must not be reported as what serves it.
				ev := run.Case(t, protocolCase(p)+"-E10-envelope-unresolved-"+e.name)
				ev.Fixture("interleave.json", raw)
				w := newWorld(t, tenantA)
				scope := ai.CallScope{TenantID: tenantB.tenant, RequestID: "req-e10-envelope-unresolved-" + p.ID + "-" + e.name}
				envs, res, err := drainEnvelopes(ev, e.start(ctxFor(t), w, p, scope))
				ev.Check("call is refused", err != nil, "err=%v", err)
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				want := ai.CallAttribution{TenantID: tenantB.tenant, RequestID: scope.RequestID, BindingID: p.BindingID}
				checkEnvelopes(ev, envs, res, want)
			})
		}
	}
}

// drainEnvelopes reads every envelope of s and its Result.
func drainEnvelopes(ev *evidence.Case, s *ai.Stream) ([]ai.EventEnvelope, ai.Result, error) {
	defer s.Close()
	var envs []ai.EventEnvelope
	for s.Next() {
		env := s.Envelope()
		ev.Check("Envelope carries the event Event returns", reflect.DeepEqual(env.Event, s.Event()), "envelope %T event %T", env.Event, s.Event())
		envs = append(envs, env)
	}
	r := resultWithin(ev, "Result returns", s)
	return envs, r.res, r.err
}

// checkEnvelopes asserts every envelope names want and the terminal's equals
// the Result's attribution.
func checkEnvelopes(ev *evidence.Case, envs []ai.EventEnvelope, res ai.Result, want ai.CallAttribution) {
	ev.Record("envelopes", envelopesJSON(envs))
	ev.Record("result", res)
	if !ev.Check("stream published events", len(envs) > 0, "none") {
		return
	}
	for i, env := range envs {
		ev.Check("event attribution", env.Call == want, "event %d (%s): got %+v want %+v", i, env.Event.Type(), env.Call, want)
	}
	ev.Check("Result attribution equals the terminal's", res.Metadata.CallAttribution == envs[len(envs)-1].Call,
		"result %+v terminal %+v", res.Metadata.CallAttribution, envs[len(envs)-1].Call)
}

// envelopesJSON keeps each event's type and attribution for the evidence bundle.
func envelopesJSON(envs []ai.EventEnvelope) []map[string]any {
	out := make([]map[string]any, 0, len(envs))
	for _, e := range envs {
		out = append(out, map[string]any{"call": e.Call, "type": e.Event.Type(), "event": e.Event})
	}
	return out
}

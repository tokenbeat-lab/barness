package e2e

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// unaryWorld observes only public host/HTTP boundaries and the existing gauges.
type unaryWorld struct {
	*world
	sc     fixtureScenario
	bodies *provider.BodyTracker
	gauges *probe.Probe
	adm    *host.Admission
	rec    *recorder
}

func newUnaryWorld(t *testing.T, configure func(*ai.Config)) *unaryWorld {
	t.Helper()
	f, _ := loadFixture(t, classifierProtocol, "choice.json")
	u := &unaryWorld{sc: f.Scenarios[0], bodies: provider.TrackBodies(provider.LoopbackTransport()), gauges: probe.New(), adm: host.NewAdmission(), rec: newRecorder()}
	u.world = newWorldWith(t, func(c *ai.Config) {
		configureClassifier(c)
		c.Transport, c.Probe, c.Admission, c.Observer = u.bodies, u.gauges, u.adm, u.rec
		c.Policy.MaxQueuedObservations = 64
		if configure != nil {
			configure(c)
		}
	}, tenantA, tenantB)
	installClassifierBinding(u.world, tenantA)
	installClassifierBinding(u.world, tenantB)
	return u
}

func (u *unaryWorld) call(t *testing.T, ctx context.Context, id string, hooks ai.Hooks) (ai.ClassifierResult, error) {
	t.Helper()
	return u.client.WithHooks(hooks).Classify(ctx, textScope(id), u.sc.target(), classifierInput(t, u.sc), nil)
}

func (u *unaryWorld) released(ev *evidence.Case) {
	state := map[string]int64{"bodies": u.bodies.Open(), "calls": u.gauges.ActiveCalls(), "permits": u.gauges.Permits(), "waiters": u.gauges.AdmissionWaiters(), "events": u.gauges.QueuedEvents(), "hostPermits": int64(u.adm.Held())}
	ev.Record("resources", state)
	for name, n := range state {
		ev.Check("released "+name, n == 0, "got %d", n)
	}
}

func checkUnaryFailure(ev *evidence.Case, res ai.ClassifierResult, err error, code ai.Code, phase ai.Phase, stop ai.StopReason) {
	ev.Record("result", res)
	ev.Record("error", errString(err))
	var ae *ai.Error
	ev.Check("typed error", errors.As(err, &ae), "got %v", err)
	ev.Check("code-only match", errors.Is(err, &ai.Error{Code: code}), "got %v", err)
	ev.Check("phase match", errors.Is(err, &ai.Error{Code: code, Phase: phase}), "got %v", err)
	ev.Check("other phase does not match", !errors.Is(err, &ai.Error{Code: code, Phase: ai.PhaseEventQueue}), "got %v", err)
	ev.Check("terminal and empty answers", res.StopReason == stop && len(res.Answers) == 0, "got %+v", res)
	if ae != nil {
		ev.Check("fixed result error text", ae.Message == res.ErrorMessage, "got %v", err)
	}
}

func (u *unaryWorld) records(ev *evidence.Case, res ai.ClassifierResult, err error) {
	obs := u.rec.awaitCall(ev, res.Metadata.RequestID)
	base := res.Metadata
	base.Attempts = nil
	starts, finishes := map[string]int{}, map[string]ai.Attempt{}
	callStarts, callFinishes := 0, 0
	for _, o := range obs {
		switch o.Kind {
		case ai.ObservationCallStarted:
			callStarts++
			want := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: res.Metadata.TenantID, RequestID: res.Metadata.RequestID, ActorID: res.Metadata.ActorID, JobID: res.Metadata.JobID, BindingID: "classifier", Operation: ai.OperationClassifier}}
			ev.Check("unresolved trusted entry", reflect.DeepEqual(o.Call, want), "got %+v", o.Call)
		case ai.ObservationCallFinished:
			callFinishes++
			ev.Check("terminal attribution and usage", reflect.DeepEqual(o.Call, res.Metadata) && o.Usage == res.Usage && o.StopReason == res.StopReason, "got %+v", o)
			var ae *ai.Error
			if errors.As(err, &ae) {
				ev.Check("observed classification", o.Error != nil && o.Error.Code == ae.Code && o.Error.Phase == ae.Phase && o.Error.HTTPStatus == ae.HTTPStatus, "got %+v", o.Error)
			} else {
				ev.Check("observed success", o.Error == nil, "got %+v", o.Error)
			}
		default:
			ev.Check("attempt attribution", reflect.DeepEqual(o.Call, base) && o.Attempt != nil, "got %+v", o)
			if o.Attempt == nil {
				continue
			}
			if o.Kind == ai.ObservationAttemptStarted {
				ev.Check("started attempt is unreported", o.Attempt.UsageReporting == ai.UsageUnreported, "got %+v", o.Attempt)
				starts[o.Attempt.AttemptID]++
			} else {
				finishes[o.Attempt.AttemptID] = *o.Attempt
			}
		}
	}
	ev.Check("one logical call", callStarts == 1 && callFinishes == 1, "got %d/%d", callStarts, callFinishes)
	ev.Check("no invented attempts", len(starts) == len(res.Metadata.Attempts) && len(finishes) == len(res.Metadata.Attempts), "got %d/%d", len(starts), len(finishes))
	for i, a := range res.Metadata.Attempts {
		ev.Check("unique attempt id", a.AttemptID == fmt.Sprintf("%s#%d", res.Metadata.RequestID, i+1) && starts[a.AttemptID] == 1, "got %+v", a)
		ev.Check("attempt usage axis agrees", reflect.DeepEqual(finishes[a.AttemptID], a), "got %+v", finishes[a.AttemptID])
	}
	ev.Record("requests", u.provider.Requests())
}

type bindingResolverFunc func(context.Context, ai.CallScope, string) (ai.Binding, error)

func (f bindingResolverFunc) ResolveBinding(ctx context.Context, s ai.CallScope, id string) (ai.Binding, error) {
	return f(ctx, s, id)
}

type credentialResolverFunc func(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error)

func (f credentialResolverFunc) ResolveCredential(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
	return f(ctx, s, b)
}

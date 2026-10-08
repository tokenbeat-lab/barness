package e2e

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// recorder is the host's Observer double. It keeps every record it is
// handed; it can be made to block until released, to fail, or to panic.
type recorder struct {
	mu   sync.Mutex
	obs  []ai.Observation
	gate chan struct{} // when set, Observe waits until it is closed
	// entered is signaled (capacity 1) each time Observe starts.
	entered chan struct{}
	fail    error
	panics  bool
}

func newRecorder() *recorder { return &recorder{entered: make(chan struct{}, 1)} }

func (r *recorder) Observe(o ai.Observation) error {
	select {
	case r.entered <- struct{}{}:
	default:
	}
	if r.gate != nil {
		<-r.gate
	}
	r.mu.Lock()
	r.obs = append(r.obs, o)
	r.mu.Unlock()
	if r.panics {
		panic("observer bug: " + string(o.Kind))
	}
	return r.fail
}

func (r *recorder) all() []ai.Observation {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]ai.Observation(nil), r.obs...)
}

// forRequest is the records of one logical call, in delivery order.
func (r *recorder) forRequest(requestID string) []ai.Observation {
	var out []ai.Observation
	for _, o := range r.all() {
		if o.Call.RequestID == requestID {
			out = append(out, o)
		}
	}
	return out
}

// awaitCall waits until the call's CallFinished was delivered and returns
// its records.
func (r *recorder) awaitCall(ev *evidence.Case, requestID string) []ai.Observation {
	eventually(ev, "observer received call_finished of "+requestID, func() bool {
		obs := r.forRequest(requestID)
		return len(obs) > 0 && obs[len(obs)-1].Kind == ai.ObservationCallFinished
	})
	obs := r.forRequest(requestID)
	ev.Record("observations-"+requestID, obs)
	return obs
}

// observedWorld is a world for tenants whose Client reports to rec, with a
// queue of 64 records unless configure changes it.
func observedWorld(t *testing.T, rec *recorder, configure func(*ai.Config), tenants ...tenantKey) *world {
	t.Helper()
	return newWorldWith(t, func(c *ai.Config) {
		c.Observer = rec
		c.Policy.MaxQueuedObservations = 64
		if configure != nil {
			configure(c)
		}
	}, tenants...)
}

// kinds summarizes records as kind[:attemptID].
func kinds(obs []ai.Observation) []string {
	out := make([]string, len(obs))
	for i, o := range obs {
		out[i] = string(o.Kind)
		if o.Attempt != nil {
			out[i] += ":" + o.Attempt.AttemptID
		}
	}
	return out
}

// lifecycle is the expected kind summary of a call whose attempts are
// numbered 1..attempts.
func lifecycle(requestID string, attempts int) []string {
	out := []string{string(ai.ObservationCallStarted)}
	for i := 1; i <= attempts; i++ {
		id := fmt.Sprintf("%s#%d", requestID, i)
		out = append(out, string(ai.ObservationAttemptStarted)+":"+id, string(ai.ObservationAttemptFinished)+":"+id)
	}
	return append(out, string(ai.ObservationCallFinished))
}

// checkCallRecords asserts the relation of one call's records to its
// outcome: the lifecycle, CallStarted holding only the scope and requested
// binding, attempt records attributed to the resolved snapshot and matching
// the Result's attempts, and CallFinished mirroring the Result without
// error text.
func checkCallRecords(ev *evidence.Case, obs []ai.Observation, scope ai.CallScope, bindingID string, o outcome, attempts int) {
	ev.Check("call/attempt lifecycle", reflect.DeepEqual(kinds(obs), lifecycle(scope.RequestID, attempts)),
		"got %q want %q", kinds(obs), lifecycle(scope.RequestID, attempts))
	if len(obs) < 2 {
		return
	}
	for i := 1; i < len(obs); i++ {
		ev.Check("records are in time order", !obs[i].Time.Before(obs[i-1].Time), "record %d at %s before %s", i, obs[i].Time, obs[i-1].Time)
	}
	started := obs[0].Call
	want := ai.CallMetadata{CallAttribution: ai.CallAttribution{TenantID: scope.TenantID, RequestID: scope.RequestID, ActorID: scope.ActorID, JobID: scope.JobID, BindingID: bindingID, Operation: ai.OperationChat}}
	ev.Check("call_started holds the scope, requested binding and entry operation", reflect.DeepEqual(started, want), "got %+v want %+v", started, want)

	res := o.result
	resolved := res.Metadata
	resolved.Attempts = nil
	var recorded []ai.Attempt
	for _, x := range obs[1 : len(obs)-1] {
		ev.Check("attempt record attributed to the resolved call", reflect.DeepEqual(x.Call, resolved), "got %+v want %+v", x.Call, resolved)
		if x.Kind == ai.ObservationAttemptFinished && x.Attempt != nil {
			recorded = append(recorded, *x.Attempt)
			ev.Check("attempt duration is set", x.Duration > 0, "got %s", x.Duration)
		}
	}
	ev.Check("attempt_finished records equal the Result's attempts", reflect.DeepEqual(recorded, res.Metadata.Attempts),
		"got %+v want %+v", recorded, res.Metadata.Attempts)

	fin := obs[len(obs)-1]
	ev.Check("call_finished carries the Result's metadata", reflect.DeepEqual(fin.Call, res.Metadata), "got %+v want %+v", fin.Call, res.Metadata)
	ev.Check("call_finished stop reason", fin.StopReason == res.Message.StopReason, "got %q want %q", fin.StopReason, res.Message.StopReason)
	ev.Check("call_finished usage is the message's", fin.Usage == res.Message.Usage, "got %+v want %+v", fin.Usage, res.Message.Usage)
	ev.Check("call_finished duration is set", fin.Duration > 0, "got %s", fin.Duration)
	var ae *ai.Error
	if errors.As(o.err, &ae) {
		want := &ai.ObservedError{Code: ae.Code, Phase: ae.Phase, HTTPStatus: ae.HTTPStatus, ProviderRequestID: ae.ProviderRequestID, RetryAfter: ae.RetryAfter}
		ev.Check("call_finished classifies the error, without its text", reflect.DeepEqual(fin.Error, want), "got %+v want %+v", fin.Error, want)
	} else {
		ev.Check("call_finished has no error on success", fin.Error == nil && o.err == nil, "got %+v (err=%v)", fin.Error, o.err)
	}
}

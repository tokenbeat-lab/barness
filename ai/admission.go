package ai

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

// Admission is a trusted host's admission for the attempts of every call,
// consulted in addition to the Client's built-in per-tenant and process
// concurrency limits, after they granted a permit (spec I9). It is where a
// host aggregates concurrency per vendor account (AccountScopeID) or across
// processes; the built-in limits only bound this process. Cross-instance
// quotas, hard cost budgets and their reconciliation are the host's.
//
// Admit is called once before each attempt, retries included, and never for
// a call that failed its preflight. It must honor ctx and bound its own
// waiting: ResourcePolicy.AdmissionWait bounds only the built-in wait, and
// the call's time limits bound the rest. An error refuses the attempt with
// CodeAdmissionDenied; its text is kept for the host's diagnosis
// (errors.Is/As) but never published. On success Admit returns release,
// which the Client calls exactly once when the attempt ended: when it failed,
// or when the stream it opened was closed. A nil release releases nothing.
type Admission interface {
	Admit(ctx context.Context, req AdmissionRequest) (release func(), err error)
}

// AdmissionRequest is what an attempt asks admission for. Every value comes
// from the trusted scope and the resolved binding snapshot, never from the
// request.
type AdmissionRequest struct {
	// TenantID owns the call's resources.
	TenantID string
	// AccountScopeID is the vendor account the attempt is billed to; several
	// tenants and keys may share one. It is never an API key.
	AccountScopeID string
	// AttemptID identifies the attempt (see Attempt.AttemptID). It is
	// globally unique, so a distributed admission can key, correlate or
	// deduplicate its leases on it alone.
	AttemptID string
}

// admitter grants each attempt its permits: the built-in limits' first,
// then the host's.
type admitter struct {
	local *limiter
	host  Admission // nil when the host injects none
}

// admit obtains the attempt's permits, or classifies why it got none. It
// holds nothing when it fails, and the returned release is safe to call
// more than once.
func (a admitter) admit(ctx context.Context, req AdmissionRequest) (func(), *Error) {
	if ctx.Err() != nil {
		return nil, admissionInterrupted(ctx)
	}
	releaseLocal, failure := a.local.acquire(ctx, req.TenantID)
	if failure != nil {
		return nil, failure
	}
	if a.host == nil {
		return releaseLocal, nil
	}
	releaseHost, err := a.host.Admit(ctx, req)
	if releaseHost == nil {
		releaseHost = func() {}
	}
	switch {
	case ctx.Err() != nil:
		// The call ended while the host decided: whatever it granted goes
		// back at once.
		if err == nil {
			releaseHost()
		}
		releaseLocal()
		return nil, admissionInterrupted(ctx)
	case err != nil:
		releaseLocal()
		failure := newError(CodeAdmissionDenied, PhaseAdmission, "admission denied by the host")
		failure.cause = err
		return nil, failure
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			releaseHost()
			releaseLocal()
		})
	}, nil
}

// admissionInterrupted classifies an ended context while an attempt waited
// for admission. It is inside pi's retry of the initial request, so the
// texts are the request phase's.
func admissionInterrupted(ctx context.Context) *Error {
	failure := requestInterrupted(ctx)
	failure.Phase = PhaseAdmission
	return failure
}

// limiter is the built-in admission: at most perTenant permits per tenant
// and process permits in all, held by attempts of this Client's process
// only. At capacity an attempt waits up to wait for a permit, or is refused
// at once when wait is 0 or maxWaiters attempts already wait. Waiters are
// served first come first served among those a freed permit fits, so a
// tenant at its own limit never holds back another tenant's waiters.
type limiter struct {
	perTenant, process, maxWaiters int
	wait                           time.Duration
	probe                          *probe.Probe

	mu      sync.Mutex
	held    map[string]int // permits per tenant; a tenant holding none is absent
	total   int
	waiters []*waiter
}

// waiter is one attempt waiting for a permit; granted is closed once it
// holds one.
type waiter struct {
	tenant  string
	granted chan struct{}
}

func newLimiter(p *ResourcePolicy, pr *probe.Probe) *limiter {
	return &limiter{perTenant: p.MaxConcurrentPerTenant, process: p.MaxConcurrentProcess, maxWaiters: p.MaxAdmissionWaiters, wait: p.AdmissionWait,
		probe: pr, held: map[string]int{}}
}

// acquire obtains a permit for tenant, waiting at most l.wait.
func (l *limiter) acquire(ctx context.Context, tenant string) (func(), *Error) {
	l.mu.Lock()
	if l.fits(tenant) {
		l.grantLocked(tenant)
		l.mu.Unlock()
		return l.releaser(tenant), nil
	}
	if l.wait == 0 {
		failure := l.deniedLocked(tenant)
		l.mu.Unlock()
		return nil, failure
	}
	if len(l.waiters) >= l.maxWaiters {
		l.mu.Unlock()
		return nil, newError(CodeAdmissionDenied, PhaseAdmission,
			"admission queue full: the resource policy's MaxAdmissionWaiters ("+strconv.Itoa(l.maxWaiters)+")")
	}
	w := &waiter{tenant: tenant, granted: make(chan struct{})}
	l.waiters = append(l.waiters, w)
	l.mu.Unlock()

	l.probe.WaitStarted()
	defer l.probe.WaitEnded()
	timer := time.NewTimer(l.wait)
	defer timer.Stop()
	select {
	case <-w.granted:
		return l.releaser(tenant), nil
	case <-timer.C:
	case <-ctx.Done():
	}

	l.mu.Lock()
	if !l.removeLocked(w) {
		// A release granted the permit as the wait ended: keep it unless the
		// call is over.
		l.mu.Unlock()
		if ctx.Err() == nil {
			return l.releaser(tenant), nil
		}
		l.release(tenant)
		return nil, admissionInterrupted(ctx)
	}
	failure := l.deniedLocked(tenant)
	l.mu.Unlock()
	if ctx.Err() != nil {
		return nil, admissionInterrupted(ctx)
	}
	return nil, failure
}

func (l *limiter) fits(tenant string) bool {
	return l.held[tenant] < l.perTenant && l.total < l.process
}

func (l *limiter) grantLocked(tenant string) {
	l.held[tenant]++
	l.total++
	l.probe.Admitted()
}

// deniedLocked names the limit that keeps tenant out.
func (l *limiter) deniedLocked(tenant string) *Error {
	field, limit := "MaxConcurrentProcess", l.process
	if l.held[tenant] >= l.perTenant {
		field, limit = "MaxConcurrentPerTenant", l.perTenant
	}
	return newError(CodeAdmissionDenied, PhaseAdmission,
		"concurrency limit reached: the resource policy's "+field+" ("+strconv.Itoa(limit)+")")
}

// removeLocked takes w off the waiters; false means it was already granted.
func (l *limiter) removeLocked(w *waiter) bool {
	for i, x := range l.waiters {
		if x == w {
			l.waiters = append(l.waiters[:i], l.waiters[i+1:]...)
			return true
		}
	}
	return false
}

// releaser returns tenant's permit once, however often it is called.
func (l *limiter) releaser(tenant string) func() {
	var once sync.Once
	return func() { once.Do(func() { l.release(tenant) }) }
}

// release returns a permit and hands every permit that is now free to the
// earliest waiters it fits.
func (l *limiter) release(tenant string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.held[tenant]--; l.held[tenant] == 0 {
		delete(l.held, tenant)
	}
	l.total--
	l.probe.Released()
	for i := 0; i < len(l.waiters); {
		w := l.waiters[i]
		if !l.fits(w.tenant) {
			i++
			continue
		}
		l.grantLocked(w.tenant)
		close(w.granted)
		l.waiters = append(l.waiters[:i], l.waiters[i+1:]...)
	}
}

package host

import (
	"context"
	"sync"

	"github.com/tokenbeat-lab/barness/ai"
)

// Admission is the trusted host's injected admission double, standing in for
// an account-wide or distributed admission service (spec I9). It records
// every request it receives, counts the permits it granted and has not got
// back, and can be told to deny. It is safe for concurrent use.
type Admission struct {
	mu       sync.Mutex
	requests []ai.AdmissionRequest
	held     int
	peak     int
	deny     error
}

// NewAdmission returns an admission that grants everything.
func NewAdmission() *Admission { return &Admission{} }

// Admit implements ai.Admission.
func (a *Admission) Admit(_ context.Context, req ai.AdmissionRequest) (func(), error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.requests = append(a.requests, req)
	if a.deny != nil {
		return nil, a.deny
	}
	a.held++
	a.peak = max(a.peak, a.held)
	var once sync.Once
	return func() {
		once.Do(func() {
			a.mu.Lock()
			a.held--
			a.mu.Unlock()
		})
	}, nil
}

// Deny makes every later Admit fail with err; nil grants again.
func (a *Admission) Deny(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.deny = err
}

// Requests are the requests Admit received, in order.
func (a *Admission) Requests() []ai.AdmissionRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]ai.AdmissionRequest(nil), a.requests...)
}

// Held is the number of granted permits not yet released.
func (a *Admission) Held() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.held
}

// PeakHeld is the highest Held seen so far.
func (a *Admission) PeakHeld() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.peak
}

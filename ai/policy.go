package ai

import "time"

// ResourcePolicy is the explicit, finite resource policy every Client is built
// with (D1, ADR-0002). There is no "unlimited" mode and no implicit default:
// zero or negative capacities and time limits are configuration errors.
//
// The policy is not a protocol option and cannot be changed by a Request.
type ResourcePolicy struct {
	// Byte limits, checked while reading or decoding rather than after.
	MaxRequestBytes   int64 // whole encoded request body
	MaxImageBytes     int64 // a single image; must not exceed MaxRequestBytes
	MaxFrameBytes     int64 // a single SSE frame
	MaxToolJSONBytes  int64 // one tool call's argument JSON; must not exceed MaxOutputBytes
	MaxErrorBodyBytes int64 // a provider error body read for diagnosis
	MaxOutputBytes    int64 // all assistant output of one call

	// Event queue bounds for Stream consumers. Overflow ends the call with a
	// resource_limit error; Complete never queues events.
	MaxQueuedEvents     int
	MaxQueuedEventBytes int64

	// In-process concurrency; MaxConcurrentPerTenant must not exceed MaxConcurrentProcess.
	MaxConcurrentPerTenant int
	MaxConcurrentProcess   int

	// AdmissionWait bounds how long an attempt waits for a concurrency permit.
	// Zero is valid and means "reject immediately when at capacity".
	AdmissionWait time.Duration

	// CallTimeout bounds a whole logical call; every other duration must fit within it.
	CallTimeout           time.Duration
	ConnectTimeout        time.Duration
	ResponseHeaderTimeout time.Duration
	// ReadIdleTimeout bounds the wait for the next upstream byte only.
	ReadIdleTimeout time.Duration
}

// validate reports the first violated rule, in declaration order, so the error
// is deterministic.
func (p *ResourcePolicy) validate() error {
	if p == nil {
		return &ConfigError{Field: "Policy", Problem: "a resource policy is required"}
	}
	positive := []struct {
		field string
		value int64
	}{
		{"MaxRequestBytes", p.MaxRequestBytes},
		{"MaxImageBytes", p.MaxImageBytes},
		{"MaxFrameBytes", p.MaxFrameBytes},
		{"MaxToolJSONBytes", p.MaxToolJSONBytes},
		{"MaxErrorBodyBytes", p.MaxErrorBodyBytes},
		{"MaxOutputBytes", p.MaxOutputBytes},
		{"MaxQueuedEvents", int64(p.MaxQueuedEvents)},
		{"MaxQueuedEventBytes", p.MaxQueuedEventBytes},
		{"MaxConcurrentPerTenant", int64(p.MaxConcurrentPerTenant)},
		{"MaxConcurrentProcess", int64(p.MaxConcurrentProcess)},
		{"CallTimeout", int64(p.CallTimeout)},
		{"ConnectTimeout", int64(p.ConnectTimeout)},
		{"ResponseHeaderTimeout", int64(p.ResponseHeaderTimeout)},
		{"ReadIdleTimeout", int64(p.ReadIdleTimeout)},
	}
	for _, f := range positive {
		if f.value <= 0 {
			return policyError(f.field, "must be positive; unlimited is not supported")
		}
	}
	if p.AdmissionWait < 0 {
		return policyError("AdmissionWait", "must not be negative")
	}
	switch {
	case p.MaxImageBytes > p.MaxRequestBytes:
		return policyError("MaxImageBytes", "must not exceed MaxRequestBytes")
	case p.MaxToolJSONBytes > p.MaxOutputBytes:
		return policyError("MaxToolJSONBytes", "must not exceed MaxOutputBytes")
	case p.MaxConcurrentPerTenant > p.MaxConcurrentProcess:
		return policyError("MaxConcurrentPerTenant", "must not exceed MaxConcurrentProcess")
	}
	for _, d := range []struct {
		field string
		value time.Duration
	}{
		{"AdmissionWait", p.AdmissionWait},
		{"ConnectTimeout", p.ConnectTimeout},
		{"ResponseHeaderTimeout", p.ResponseHeaderTimeout},
		{"ReadIdleTimeout", p.ReadIdleTimeout},
	} {
		if d.value > p.CallTimeout {
			return policyError(d.field, "must not exceed CallTimeout")
		}
	}
	return nil
}

func policyError(field, problem string) error {
	return &ConfigError{Field: "Policy." + field, Problem: problem}
}

package ai

import "time"

// ResourcePolicy is the explicit, finite resource policy every Client is built
// with (D1, ADR-0002). There is no "unlimited" mode and no implicit default:
// zero or negative capacities and time limits are configuration errors.
//
// The policy is not a protocol option and cannot be changed by a Request.
type ResourcePolicy struct {
	// Byte limits, enforced while encoding, reading or merging rather than
	// after. Reaching one ends the call with StopReason error and a
	// CodeResourceLimit *Error; a limit is never retried around.
	//
	// MaxRequestBytes bounds the final encoded request body (after the
	// payload callback), refused in PhaseRequest before it is sent.
	MaxRequestBytes int64
	// MaxImageBytes bounds a single image of the request history, counted as
	// the image's own bytes rather than its base64 text, refused in
	// PhaseScope before anything resolves. It must not exceed MaxRequestBytes.
	MaxImageBytes int64
	// MaxFrameBytes bounds a single SSE frame: an event's lines through the
	// blank line that ends it. Some protocols repeat the whole response in
	// their terminal frame (OpenAI Responses does), so it must then hold the
	// largest response the host accepts.
	MaxFrameBytes int64
	// MaxToolJSONBytes bounds one tool call's argument JSON as it grows. It
	// must not exceed MaxOutputBytes.
	MaxToolJSONBytes int64
	// MaxErrorBodyBytes bounds a provider's non-2xx body read for diagnosis.
	MaxErrorBodyBytes int64
	// MaxOutputBytes bounds the provider's streamed response body of one
	// call, as read after transport decoding. Every frame counts, so it bounds
	// everything the call reads and may keep, not only the message's text
	// (ADR-0007).
	MaxOutputBytes int64

	// Event queue bounds for Stream consumers: how many events wait unread,
	// and how many bytes of text they carry (deltas, block contents, tool
	// call fields). The terminal event's place is reserved and never counts.
	// Overflow ends the call with a resource_limit error in PhaseEventQueue;
	// the producer is never blocked. Complete never queues events.
	MaxQueuedEvents     int
	MaxQueuedEventBytes int64

	// Built-in admission: how many attempts may run at once per tenant and
	// in this process, each attempt (retries included) holding a permit
	// until it failed or the stream it opened was closed. It only bounds
	// this process; per-account or distributed admission is the host's
	// (Config.Admission). MaxConcurrentPerTenant must not exceed
	// MaxConcurrentProcess.
	MaxConcurrentPerTenant int
	MaxConcurrentProcess   int

	// AdmissionWait bounds how long an attempt waits for a built-in permit
	// before it is refused with CodeAdmissionDenied. Zero is valid and means
	// "reject immediately when at capacity". There is no admission queue
	// beyond the attempts waiting, each for at most this long.
	AdmissionWait time.Duration

	// CallTimeout bounds a whole logical call, resolution included, as its
	// context's deadline; the host's own deadline ends it earlier. Every
	// other duration must fit within it.
	CallTimeout time.Duration
	// ConnectTimeout bounds each attempt until the transport has a
	// connection (see Config.Transport).
	ConnectTimeout time.Duration
	// ResponseHeaderTimeout bounds each attempt until its response headers
	// arrived, connecting included; a protocol request timeout (timeoutMs)
	// applies instead when it is earlier.
	ResponseHeaderTimeout time.Duration
	// ReadIdleTimeout bounds each wait for the next upstream bytes only;
	// time the call spends between reads does not count. A downstream
	// consumer's deadline is the host's.
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

package ai

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"
)

// Time limits of one call (spec I9). The call as a whole is bounded by its
// context: the host's deadline and the policy's CallTimeout, whichever ends
// first (Client.run). Each attempt, one HTTP round trip, is bounded on top
// by the policy's ConnectTimeout, ResponseHeaderTimeout and ReadIdleTimeout
// and by the protocol's request timeout (timeoutMs), which watchedTransport
// enforces whatever SDK the adapter uses.

// timeLimitExpired is the error a watched round trip or response body
// returns once one of the attempt's time limits ran out; the attempt's I/O
// was canceled there.
type timeLimitExpired struct {
	field string // the ResourcePolicy field, or "timeoutMs" for the protocol's
	limit time.Duration
}

func (e *timeLimitExpired) Error() string { return e.failure(PhaseRequest).Message }

// failure classifies the expiry in the phase it happened. The protocol's
// timeoutMs keeps openai-node's text, as pi reports it; the policy's limits
// are barness's own and name the field.
func (e *timeLimitExpired) failure(phase Phase) *Error {
	msg := msgRequestTimedOut
	if e.field != fieldProtocolTimeout {
		msg = "exceeded the resource policy's " + e.field + " (" + e.limit.String() + ")"
	}
	failure := newError(CodeDeadlineExceeded, phase, msg)
	failure.attemptTimeout = true
	return failure
}

// The time limits a timeLimitExpired names.
const (
	fieldConnectTimeout        = "ConnectTimeout"
	fieldResponseHeaderTimeout = "ResponseHeaderTimeout"
	fieldReadIdleTimeout       = "ReadIdleTimeout"
	fieldProtocolTimeout       = "timeoutMs"
)

// protocolTimeoutKey carries the protocol's request timeout of a call to its
// round trips.
type protocolTimeoutKey struct{}

// withProtocolTimeout makes d the protocol's request timeout of the round
// trips made with ctx: like openai-node's timeout, it bounds each attempt
// until its response headers arrived. The earlier of it and the policy's
// ResponseHeaderTimeout applies.
func withProtocolTimeout(ctx context.Context, d time.Duration) context.Context {
	return context.WithValue(ctx, protocolTimeoutKey{}, d)
}

// watchedTransport bounds every round trip of a Client by the attempt time
// limits. It wraps the configured transport, so the limits hold for every
// adapter and SDK, and they act by canceling the attempt's own context:
// nothing waits on a stalled upstream once a limit ran out.
//
// ConnectTimeout runs from the start of the round trip until the transport
// reports a connection through net/http/httptrace (GotConn), as
// *http.Transport does; ResponseHeaderTimeout and the protocol timeout run
// until the response headers arrived; ReadIdleTimeout bounds each wait of a
// body Read for upstream data, so time the call spends elsewhere between
// reads never counts.
type watchedTransport struct {
	next                      http.RoundTripper
	connect, header, readIdle time.Duration
}

func (t *watchedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithCancelCause(req.Context())
	header, field := t.header, fieldResponseHeaderTimeout
	if d, ok := req.Context().Value(protocolTimeoutKey{}).(time.Duration); ok && d < header {
		header, field = d, fieldProtocolTimeout
	}
	// The connect and header timers settle under mu, so a timer that fires
	// as the round trip returns either canceled it before (and the round
	// trip reports the expiry) or does nothing: it never cancels the body
	// that is handed on.
	var mu sync.Mutex
	settled := false
	expire := func(field string, limit time.Duration) func() {
		return func() {
			mu.Lock()
			defer mu.Unlock()
			if !settled {
				cancel(&timeLimitExpired{field: field, limit: limit})
			}
		}
	}
	connectTimer := armAttemptTimer(ctx, t.connect, expire(fieldConnectTimeout, t.connect), nil)
	headerTimer := armAttemptTimer(ctx, header, expire(field, header), nil)
	trace := &httptrace.ClientTrace{GotConn: func(httptrace.GotConnInfo) { stopAttemptTimer(connectTimer) }}
	res, err := t.next.RoundTrip(req.WithContext(httptrace.WithClientTrace(ctx, trace)))
	stopAttemptTimer(connectTimer)
	stopAttemptTimer(headerTimer)
	mu.Lock()
	settled = true
	mu.Unlock()
	if expired := expiredLimit(ctx); expired != nil {
		// A limit ran out as the round trip ended; the response, if any, is
		// dropped with its body.
		if res != nil && res.Body != nil {
			_ = res.Body.Close()
		}
		cancel(nil)
		return nil, expired
	}
	if err != nil {
		cancel(nil)
		return nil, err
	}
	body := &idleBody{ReadCloser: res.Body, ctx: ctx, cancel: cancel, limit: t.readIdle}
	res.Body = body
	return res, nil
}

// expiredLimit is the time limit that canceled ctx, or nil.
func expiredLimit(ctx context.Context) *timeLimitExpired {
	var expired *timeLimitExpired
	if errors.As(context.Cause(ctx), &expired) {
		return expired
	}
	return nil
}

// A later attempt timer must not win a race with the earlier call deadline's
// cancellation callback. Context propagation and timer dispatch are asynchronous:
// both can be overdue while ctx.Err() is still nil (release-verification #01).
// Keep ownership with the earlier deadline, as ADR-0008 requires. Read timers
// are reused, but stopped once the remaining call time becomes the smaller bound.
func armAttemptTimer(ctx context.Context, limit time.Duration, expire func(), timer *time.Timer) *time.Timer {
	if deadline, ok := ctx.Deadline(); ok && !time.Now().Add(limit).Before(deadline) {
		stopAttemptTimer(timer)
		return nil
	}
	if timer == nil {
		return time.AfterFunc(limit, expire)
	}
	timer.Reset(limit)
	return timer
}

func stopAttemptTimer(timer *time.Timer) {
	if timer != nil {
		timer.Stop()
	}
}

// idleBody bounds each Read's wait for upstream data by the read idle limit
// and releases the attempt's context when closed. The timer only cancels
// while a Read waits, so one firing as a Read returns data does nothing.
type idleBody struct {
	io.ReadCloser
	ctx    context.Context
	cancel context.CancelCauseFunc
	idle   *time.Timer
	limit  time.Duration

	mu      sync.Mutex
	reading bool
}

func (b *idleBody) expire() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.reading && b.idle != nil {
		b.cancel(&timeLimitExpired{field: fieldReadIdleTimeout, limit: b.limit})
	}
}

func (b *idleBody) Read(p []byte) (int, error) {
	b.mu.Lock()
	b.reading = true
	b.idle = armAttemptTimer(b.ctx, b.limit, b.expire, b.idle)
	b.mu.Unlock()
	n, err := b.ReadCloser.Read(p)
	b.mu.Lock()
	b.reading = false
	stopAttemptTimer(b.idle)
	b.mu.Unlock()
	if err != nil {
		if expired := expiredLimit(b.ctx); expired != nil {
			return n, expired
		}
	}
	return n, err
}

func (b *idleBody) Close() error {
	b.mu.Lock()
	stopAttemptTimer(b.idle)
	b.mu.Unlock()
	err := b.ReadCloser.Close()
	b.cancel(nil)
	return err
}

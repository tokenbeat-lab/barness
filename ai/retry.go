package ai

import (
	"context"
	"errors"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tokenbeat-lab/barness/ai/internal/clock"
)

// RetryPolicy is an operator's explicit retry configuration for the calls
// made through one binding (spec I8). The zero value retries nothing.
//
// It is binding configuration, not a request option, so neither a Request
// nor Options can raise it, and every attempt of a call uses the policy of
// the binding snapshot the call resolved. Only the initial request is
// retried, by pi-ai's provider-retry rules: x-should-retry decides first,
// then a connection failure without a response and HTTP 408, 409, 429 and
// 5xx are retried, and nothing else. A stream that has started is never
// replayed. A request that reached the provider but whose response was lost
// may be billed twice.
type RetryPolicy struct {
	// MaxRetries is how often a retryable failure of the initial request is
	// retried; 0 sends it once. It must not be negative.
	MaxRetries int
	// MaxRetryDelay caps the delay a provider asks for through retry-after-ms
	// or retry-after: a longer one fails the call at once instead of
	// waiting. Nil means 60 seconds; 0 lifts the cap, so the wait is bounded
	// only by the call's context. It must not be negative. Backoff without a
	// requested delay starts at 500ms, doubles per retry up to 8s and is
	// lowered by up to 25% jitter.
	MaxRetryDelay *time.Duration
}

// defaultMaxRetryDelay is pi's DEFAULT_MAX_RETRY_DELAY_MS.
const defaultMaxRetryDelay = 60 * time.Second

func (p RetryPolicy) validate() string {
	if p.MaxRetries < 0 {
		return "binding retry policy: MaxRetries must not be negative"
	}
	if p.MaxRetryDelay != nil && *p.MaxRetryDelay < 0 {
		return "binding retry policy: MaxRetryDelay must not be negative"
	}
	return ""
}

// pinned copies the policy so a resolver that reuses its delay value cannot
// change a call's pinned snapshot.
func (p RetryPolicy) pinned() RetryPolicy {
	if p.MaxRetryDelay != nil {
		d := *p.MaxRetryDelay
		p.MaxRetryDelay = &d
	}
	return p
}

// maxDelay is the effective cap; 0 means none.
func (p RetryPolicy) maxDelay() time.Duration {
	if p.MaxRetryDelay == nil {
		return defaultMaxRetryDelay
	}
	return *p.MaxRetryDelay
}

// Attempt is one HTTP request a logical call sent to the provider. A call
// that failed before sending anything has none.
type Attempt struct {
	// AttemptID identifies the attempt; it is the call's RequestID with the
	// attempt's ordinal, so it is unique wherever the RequestID is.
	AttemptID string `json:"attemptId"`
	// HTTPStatus is the response status; 0 when no response arrived.
	HTTPStatus int `json:"httpStatus,omitempty"`
	// ProviderRequestID is the vendor's request id from the response.
	ProviderRequestID string `json:"providerRequestId,omitempty"`
	// Code classifies why the attempt failed to obtain the initial response;
	// empty when it obtained it. A failure later in the stream is the
	// call's error, not the attempt's.
	Code Code `json:"code,omitempty"`
	// RetryDelay is the wait planned before the next attempt; 0 when no
	// retry followed. It is encoded in nanoseconds.
	RetryDelay time.Duration `json:"retryDelay,omitempty"`
}

// Texts for a call interrupted while its initial request is in flight or
// waiting to be retried, whatever the protocol.
const (
	// pi's provider-retry wrapper turns any failure of a canceled request
	// into its own AbortError, with or without retries configured.
	msgRequestAborted = "Request aborted"
	// pi has no deadline of its own here; this is openai-node's timeout text.
	msgRequestTimedOut = "Request timed out."
)

// requestInterrupted classifies an ended context in the request phase.
func requestInterrupted(ctx context.Context) *Error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return interrupted(ctx, PhaseRequest, msgRequestTimedOut)
	}
	return interrupted(ctx, PhaseRequest, msgRequestAborted)
}

// initialRequest sends one call's initial request under the retry policy
// pinned with its snapshot, and records every attempt.
type initialRequest struct {
	policy    RetryPolicy
	clock     *clock.Clock // nil is the system clock
	requestID string
	attempts  []Attempt
}

// attemptOutcome is how one attempt at the initial request ended, as the
// adapter observed it.
type attemptOutcome struct {
	// status and header are the provider's response; 0 and nil when none
	// arrived.
	status            int
	header            http.Header
	providerRequestID string
	// failure is nil when the attempt obtained the initial response.
	failure *Error
	// connection marks a failure to get any response from the provider: the
	// baseline's network-error shape (openai-node's APIConnectionError), the
	// only failure without a status that is retried.
	connection bool
	// sdkMessage is the protocol SDK's own text for an HTTP failure, which
	// pi quotes when it refuses a requested delay. Already redacted.
	sdkMessage string
}

// retryable ports pi's isRetryableProviderError.
func (o attemptOutcome) retryable() bool {
	if o.status == 0 {
		return o.connection
	}
	switch o.header.Get("x-should-retry") {
	case "true":
		return true
	case "false":
		return false
	}
	s := o.status
	return s == http.StatusRequestTimeout || s == http.StatusConflict || s == http.StatusTooManyRequests || s >= 500
}

// send runs attempt until one obtains the initial response, the policy
// allows no further retry, or ctx ends; it ports pi's retryProviderRequest.
// The adapter keeps what a successful attempt opened.
func (r *initialRequest) send(ctx context.Context, attempt func(context.Context) attemptOutcome) *Error {
	for retryIndex := 0; ; retryIndex++ {
		out := attempt(ctx)
		rec := Attempt{
			AttemptID:         r.requestID + "#" + strconv.Itoa(retryIndex+1),
			HTTPStatus:        out.status,
			ProviderRequestID: out.providerRequestID,
		}
		if out.failure == nil {
			r.attempts = append(r.attempts, rec)
			return nil
		}
		// As pi, an ended call is an interruption whatever the attempt got,
		// even an HTTP error that arrived as it ended.
		if ctx.Err() != nil {
			out.failure = requestInterrupted(ctx)
		}
		rec.Code = out.failure.Code
		if ctx.Err() != nil || retryIndex >= r.policy.MaxRetries || !out.retryable() {
			r.attempts = append(r.attempts, rec)
			return out.failure
		}
		delay, refused := r.delay(out, retryIndex)
		rec.RetryDelay = delay
		r.attempts = append(r.attempts, rec)
		if refused != nil {
			return refused
		}
		if r.clock.Sleep(ctx, delay) != nil {
			return requestInterrupted(ctx)
		}
	}
}

// delay ports pi's getRetryDelayMs: the provider's requested delay unless it
// exceeds the cap, which fails the call with the attempt's classification,
// else exponential backoff with jitter.
func (r *initialRequest) delay(out attemptOutcome, retryIndex int) (time.Duration, *Error) {
	if ms, ok := serverDelayMs(out.header, r.clock.Now()); ok {
		if capMs := durationMs(r.policy.maxDelay()); capMs > 0 && ms > capMs {
			refused := *out.failure
			refused.Message = "Server requested " + jsNumberString(math.Ceil(ms/1000)) + "s retry delay (max: " +
				jsNumberString(math.Ceil(capMs/1000)) + "s). " + out.sdkMessage
			return 0, &refused
		}
		return msDuration(ms), nil
	}
	backoff := math.Min(0.5*math.Pow(2, float64(retryIndex)), 8) * 1000
	return msDuration(backoff * (1 - r.clock.Jitter()*0.25)), nil
}

// serverDelayMs is the delay a provider asked for, in milliseconds:
// retry-after-ms, else retry-after as seconds or an HTTP date. ok is false
// when it asked for none. As in pi, a retry-after that is neither yields NaN,
// which waits no time; a number is read as JavaScript's parseFloat reads it.
func serverDelayMs(h http.Header, now time.Time) (ms float64, ok bool) {
	if v := h.Get("retry-after-ms"); v != "" {
		if ms := jsParseFloat(v); !math.IsNaN(ms) {
			return ms, true
		}
	}
	v := h.Get("retry-after")
	if v == "" {
		return 0, false
	}
	if s := jsParseFloat(v); !math.IsNaN(s) {
		return s * 1000, true
	}
	if t, ok := parseHTTPDate(v); ok {
		return durationMs(t.Sub(now)), true
	}
	return math.NaN(), true
}

// parseHTTPDate reads the date forms retry-after uses (RFC 9110 and its
// obsolete forms). JavaScript's Date.parse accepts more free-form dates,
// which pi then waits for; such text waits no time here (see serverDelayMs
// and ADR-0006). Forms starting with digits, such as ISO 8601, never reach
// this: parseFloat reads their year as seconds, in pi as here.
func parseHTTPDate(v string) (time.Time, bool) {
	t, err := http.ParseTime(v)
	return t, err == nil
}

// jsFloatPrefix is the part of a string JavaScript's parseFloat reads.
var jsFloatPrefix = regexp.MustCompile(`^[+-]?(?:Infinity|(?:\d+\.?\d*|\.\d+)(?:[eE][+-]?\d+)?)`)

// jsParseFloat is JavaScript's Number.parseFloat: the longest decimal prefix
// after leading whitespace, NaN when there is none.
func jsParseFloat(s string) float64 {
	lit := jsFloatPrefix.FindString(strings.TrimLeftFunc(s, isJSSpace))
	if lit == "" {
		return math.NaN()
	}
	if strings.HasSuffix(lit, "Infinity") {
		if lit[0] == '-' {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	f, err := jsNumberLiteral(lit)
	if err != nil {
		return math.NaN()
	}
	return f
}

func durationMs(d time.Duration) float64 { return float64(d) / float64(time.Millisecond) }

// msDuration converts a delay in milliseconds as setTimeout treats it: NaN
// and negative wait no time. Unlike setTimeout, which fires at once for
// more than 2^31-1 ms, a long delay is waited for (up to the longest
// Duration): spec I8 bounds an uncapped wait by the call's deadline instead
// (ADR-0006).
func msDuration(ms float64) time.Duration {
	switch ns := ms * float64(time.Millisecond); {
	case math.IsNaN(ns) || ns <= 0:
		return 0
	case ns >= math.MaxInt64:
		return math.MaxInt64
	default:
		return time.Duration(ns)
	}
}

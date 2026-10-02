package ai

import (
	"net/http"
	"strconv"
	"time"
)

// connectionFailure marks an error of the HTTP client itself, so the request
// got no response at all: the only error without a status the retry rules
// retry. Errors an SDK raises before sending are never marked.
type connectionFailure struct{ err error }

func (e *connectionFailure) Error() string { return e.err.Error() }
func (e *connectionFailure) Unwrap() error { return e.err }

// roundTrip is the next step of an SDK middleware chain. Each SDK names its
// own type for it, so adapters wrap the middleware below in their SDK's
// option.Middleware.
type roundTrip = func(*http.Request) (*http.Response, error)

// markConnectionFailures is SDK middleware around the HTTP client's Do.
func markConnectionFailures(req *http.Request, next roundTrip) (*http.Response, error) {
	res, err := next(req)
	if err != nil && res == nil {
		return nil, &connectionFailure{err: err}
	}
	return res, err
}

// announceTimeout is SDK middleware that announces the protocol's request
// timeout as the Stainless TypeScript SDKs do, in whole seconds truncated
// (so "0" below one second). The Go SDKs only announce their own request
// timeout, which also bounds the body; the timeout itself is enforced by
// watchedTransport.
func announceTimeout(d time.Duration) func(*http.Request, roundTrip) (*http.Response, error) {
	seconds := strconv.FormatInt(int64(d/time.Second), 10)
	return func(req *http.Request, next roundTrip) (*http.Response, error) {
		req.Header.Set("X-Stainless-Timeout", seconds)
		return next(req)
	}
}

// limitBodies is SDK middleware that puts each response body behind the
// call's byte limits before the SDK or the adapter reads it.
func (l byteLimits) limitBodies(req *http.Request, next roundTrip) (*http.Response, error) {
	res, err := next(req)
	l.limitBody(res)
	return res, err
}

// closeBody closes a response's body, if there is one.
func closeBody(res *http.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

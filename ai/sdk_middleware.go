package ai

import (
	"io"
	"net/http"
	"strconv"
	"sync"
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

// deliveredBody is SDK middleware that remembers the body of the response the
// HTTP client delivered to an attempt, so the attempt can close it when the
// SDK fails without handing that response back. The Stainless SDKs do so when
// the caller's context ends as the response arrives: their request loop
// returns the context's error before giving the response to the caller or
// closing it (internal/requestconfig in anthropic-sdk-go v1.75.0 and openai-go
// v3.66.0), which would leave the body, and whatever the host's transport
// holds for it, open. The body closes once, whoever closes it first. Adapters
// register it last, so it sits next to the HTTP client, under limitBodies.
// E06's a-canceled-on-arrival scenarios catch the leak; once both SDKs
// close the response themselves there, this can go (issue 29).
type deliveredBody struct{ body io.Closer }

func (d *deliveredBody) keep(req *http.Request, next roundTrip) (*http.Response, error) {
	d.body = nil
	res, err := next(req)
	if res != nil && res.Body != nil {
		b := &closeOnce{ReadCloser: res.Body}
		res.Body, d.body = b, b
	}
	return res, err
}

// close closes the body delivered to a failed attempt, if any. The SDK or
// the adapter may already have closed it, in either order: closeOnce makes
// that harmless.
func (d *deliveredBody) close() {
	if d.body != nil {
		_ = d.body.Close()
		d.body = nil
	}
}

// closeOnce is a body whose Close closes the body under it the first time
// only (io.Closer leaves a second Close undefined).
type closeOnce struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func (b *closeOnce) Close() error {
	b.once.Do(func() { b.err = b.ReadCloser.Close() })
	return b.err
}

// closeBody closes a response's body, if there is one.
func closeBody(res *http.Response) {
	if res != nil && res.Body != nil {
		_ = res.Body.Close()
	}
}

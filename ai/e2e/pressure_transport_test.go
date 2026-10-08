package e2e

import (
	"context"
	"io"
	"net/http"
	"slices"
	"sync"
	"sync/atomic"

	"github.com/tokenbeat-lab/barness/ai"
)

// A synthetic call marker ties a concurrent Client's failure to the Provider
// handler that received it. Enqueue/arrival order is not call identity.
const pressureCallHeader = "X-Barness-Test-Call"

func pressureDiagnosticHooks() ai.Hooks {
	return ai.Hooks{TransformHeaders: func(_ context.Context, s ai.CallScope, h http.Header) error {
		h.Set(pressureCallHeader, s.RequestID)
		return nil
	}}
}

type pressureNetworkError struct {
	RequestID    string `json:"requestId,omitempty"`
	Phase        string `json:"phase"`
	ReadBytes    int64  `json:"readBytes"`
	Error        string `json:"error"`
	ContextError string `json:"contextError,omitempty"`
	BodyClosed   bool   `json:"bodyClosed"`
}

// Raw network causes are kept only in audited synthetic E2E evidence. Product
// terminals and Observer records remain sanitized (issue 18).
type pressureNetworkErrors struct {
	next http.RoundTripper
	mu   sync.Mutex
	errs []pressureNetworkError
}

func (f *pressureNetworkErrors) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := f.next.RoundTrip(req)
	if err != nil {
		f.record(req.Context(), pressureNetworkError{RequestID: req.Header.Get(pressureCallHeader), Phase: "round-trip", Error: err.Error()})
	} else if res.Body != nil {
		res.Body = &pressureReadBody{ReadCloser: res.Body, failures: f, ctx: req.Context(), requestID: req.Header.Get(pressureCallHeader)}
	}
	return res, err
}

func (f *pressureNetworkErrors) record(ctx context.Context, e pressureNetworkError) {
	if err := ctx.Err(); err != nil {
		e.ContextError = err.Error()
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errs = append(f.errs, e)
}

func (f *pressureNetworkErrors) list() []pressureNetworkError {
	f.mu.Lock()
	defer f.mu.Unlock()
	return slices.Clone(f.errs)
}

type pressureReadBody struct {
	io.ReadCloser
	failures  *pressureNetworkErrors
	ctx       context.Context
	requestID string
	bytes     int64
	closed    atomic.Bool
}

func (b *pressureReadBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.bytes += int64(n)
	if err != nil && err != io.EOF {
		b.failures.record(b.ctx, pressureNetworkError{RequestID: b.requestID, Phase: "body-read", ReadBytes: b.bytes, Error: err.Error(), BodyClosed: b.closed.Load()})
	}
	return n, err
}

func (b *pressureReadBody) Close() error {
	b.closed.Store(true)
	return b.ReadCloser.Close()
}

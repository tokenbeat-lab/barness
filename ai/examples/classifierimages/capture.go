package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type responseMetadata struct {
	Status    int               `json:"status"`
	Header    map[string]string `json:"header"`
	Truncated bool              `json:"truncated"`
}

// Live capture retains no credential/header dump. The response wrapper records
// only bytes the Client actually reads; it preserves timeout/cancel/HTTP behavior.
type captureTransport struct {
	inner                       http.RoundTripper
	requestLimit, responseLimit int64
	mu                          sync.Mutex
	requests                    []provider.Request
	response                    []byte
	meta                        responseMetadata
}

func (t *captureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	raw, err := io.ReadAll(io.LimitReader(r.Body, t.requestLimit+1))
	closeErr := r.Body.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	if int64(len(raw)) > t.requestLimit {
		return nil, errors.New("request capture exceeds Host limit")
	}
	r.Body = io.NopCloser(bytes.NewReader(raw))
	t.mu.Lock()
	t.requests = append(t.requests, provider.Request{Method: r.Method, Path: r.URL.Path, KeyAlias: "[REDACTED]", Header: map[string]string{"Content-Type": r.Header.Get("Content-Type")}, Body: raw})
	t.mu.Unlock()
	res, err := t.inner.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{}
	for _, name := range []string{"Content-Type", "X-Request-ID", "X-Typesafe-Request-ID"} {
		if value := res.Header.Get(name); value != "" {
			headers[name] = value
		}
	}
	t.mu.Lock()
	t.meta = responseMetadata{Status: res.StatusCode, Header: headers}
	t.mu.Unlock()
	res.Body = &captureBody{ReadCloser: res.Body, capture: t}
	return res, nil
}

type captureBody struct {
	io.ReadCloser
	capture *captureTransport
}

func (b *captureBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	t := b.capture
	t.mu.Lock()
	remaining := int(t.responseLimit) - len(t.response)
	keep := n
	if keep > remaining {
		keep = remaining
		t.meta.Truncated = true
	}
	t.response = append(t.response, p[:keep]...)
	t.mu.Unlock()
	return n, err
}
func (t *captureTransport) snapshot() ([]provider.Request, []byte, responseMetadata) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]provider.Request{}, t.requests...), append([]byte(nil), t.response...), t.meta
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/finalbudget"
	"io"
	"net/http"
	"sync"
)

// The collector records bodies and only the public request-id header. No
// credentials, authentication headers, URL parameters or error strings enter
// captures. The fixed task is synthetic; the bundle writer redacts the selected
// secret and the final audit rejects unexpected sensitive content.
type wireCapture struct {
	Source            string          `json:"source"`
	Path              string          `json:"path"`
	Request           json.RawMessage `json:"request"`
	Response          string          `json:"response,omitempty"`
	Status            int             `json:"status"`
	ProviderRequestID string          `json:"providerRequestId,omitempty"`
	NetworkFailure    bool            `json:"networkFailure,omitempty"`
	ReadFailure       bool            `json:"readFailure,omitempty"`
}
type captureTransport struct {
	inner           http.RoundTripper
	kind            string
	requestIDHeader string
	mu              sync.Mutex
	captures        []wireCapture
}

func (t *captureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	raw, err := io.ReadAll(io.LimitReader(req.Body, 16385))
	req.Body.Close()
	if err != nil || len(raw) > 16384 {
		return nil, errors.New("capture request budget exceeded")
	}
	req.Body = io.NopCloser(bytes.NewReader(raw))
	t.mu.Lock()
	index := len(t.captures)
	t.captures = append(t.captures, wireCapture{Source: t.kind, Path: req.URL.Path, Request: raw})
	t.mu.Unlock()
	if err := finalbudget.Reserve(req); err != nil {
		return nil, err
	}
	res, err := t.inner.RoundTrip(req)
	t.mu.Lock()
	c := &t.captures[index]
	if err != nil {
		c.NetworkFailure = true
		t.mu.Unlock()
		return res, err
	}
	c.Status = res.StatusCode
	c.ProviderRequestID = res.Header.Get(t.requestIDHeader)
	t.mu.Unlock()
	limit := 16385
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		limit = 4097
	}
	res.Body = &capturedBody{ReadCloser: res.Body, transport: t, index: index, limit: limit}
	return res, nil
}

type capturedBody struct {
	io.ReadCloser
	transport    *captureTransport
	index, limit int
}

func (b *capturedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	t := b.transport
	t.mu.Lock()
	c := &t.captures[b.index]
	remaining := b.limit - len(c.Response)
	if remaining > 0 {
		c.Response += string(p[:min(n, remaining)])
	}
	if err != nil && err != io.EOF {
		c.ReadFailure = true
	}
	t.mu.Unlock()
	return n, err
}
func (t *captureTransport) snapshot() []wireCapture {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]wireCapture, len(t.captures))
	for i, c := range t.captures {
		c.Request = bytes.Clone(c.Request)
		out[i] = c
	}
	return out
}

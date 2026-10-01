package provider

import (
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

// BodyTracker wraps the Client's transport and counts response bodies that
// were handed to the client but not yet closed. It observes resource release
// at the HTTP boundary (spec Testing Decisions §3), without looking inside
// barness-ai.
type BodyTracker struct {
	next http.RoundTripper
	open atomic.Int64
}

// TrackBodies wraps next.
func TrackBodies(next http.RoundTripper) *BodyTracker { return &BodyTracker{next: next} }

// RoundTrip implements http.RoundTripper.
func (t *BodyTracker) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(req)
	if res != nil && res.Body != nil {
		t.open.Add(1)
		res.Body = &trackedBody{ReadCloser: res.Body, tracker: t}
	}
	return res, err
}

// Open is the number of response bodies not yet closed.
func (t *BodyTracker) Open() int64 { return t.open.Load() }

type trackedBody struct {
	io.ReadCloser
	tracker *BodyTracker
	once    sync.Once
}

func (b *trackedBody) Close() error {
	b.once.Do(func() { b.tracker.open.Add(-1) })
	return b.ReadCloser.Close()
}

package e2e

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
)

// mixedEOFTransport observes the public HTTP body boundary. Unlike a server
// flush, reaching EOF here proves every byte was handed back to the Client.
// No body copy is kept. Chat is held by its public OnResponse callback instead.
type mixedEOFTransport struct {
	next    http.RoundTripper
	ended   func()
	release <-chan struct{}
	read    atomic.Int64
}

func (t *mixedEOFTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.next.RoundTrip(req)
	if err == nil && res.Body != nil && !strings.Contains(req.URL.Path, "/chat/") {
		res.Body = &mixedEOFBody{ReadCloser: res.Body, ctx: req.Context(), transport: t}
	}
	return res, err
}

type mixedEOFBody struct {
	io.ReadCloser
	ctx        context.Context
	transport  *mixedEOFTransport
	pendingEOF bool
	once       sync.Once
}

func (b *mixedEOFBody) Read(p []byte) (int, error) {
	if !b.pendingEOF {
		n, err := b.ReadCloser.Read(p)
		b.transport.read.Add(int64(n))
		if err != io.EOF {
			return n, err
		}
		b.pendingEOF = true
		// A Reader may return data and EOF together. Hand the data back
		// first so io.ReadAll has appended it before the measurement barrier.
		if n > 0 {
			return n, nil
		}
	}
	b.once.Do(b.transport.ended)
	select {
	case <-b.transport.release:
		return 0, io.EOF
	case <-b.ctx.Done():
		return 0, b.ctx.Err()
	}
}

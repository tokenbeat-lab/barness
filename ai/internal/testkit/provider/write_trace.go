package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sync"
	"sync/atomic"
)

// SocketWrite records failed writes: what net.Conn reported accepting. It does
// not claim delivery to TCP or that the unreported suffix is safe to resend.
// Digests cover wire bytes (including HTTP framing), never expose credentials.
type SocketWrite struct {
	Sequence int    `json:"sequence"`
	Offset   int64  `json:"offset"`
	Bytes    int    `json:"bytes"`
	Written  int    `json:"written"`
	SHA256   string `json:"sha256"`
	Error    string `json:"error,omitempty"`
}

type traceListener struct {
	net.Listener
	next atomic.Int64
}

func (l *traceListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	// Bound the controlled fixture's kernel queue before any HTTP writes.
	// Autotuned send buffers across 32 design-load streams can exhaust Darwin
	// mbufs; unlike ordinary TCP backpressure, ENOBUFS may hide consumed bytes
	// (#07). Request a small send window so writers wait for readers instead.
	// Custom socket-fault listeners retain their own connection semantics.
	if tcp, ok := c.(*net.TCPConn); ok {
		if err := tcp.SetWriteBuffer(64 << 10); err != nil {
			_ = c.Close()
			return nil, err
		}
	}
	return &traceConn{Conn: c, id: l.next.Add(1)}, nil
}

type traceConn struct {
	net.Conn
	mu       sync.Mutex
	id       int64
	trace    bool
	writes   []SocketWrite
	offset   int64
	sequence int
}

// Never resume ENOBUFS here. Darwin can append to the TCP send buffer before
// returning ENOBUFS, while the syscall discards that write's consumed count.
// Replaying p[n:] can therefore duplicate bytes already enqueued, corrupting
// HTTP/SSE or silently accumulating duplicate tool deltas (thinking-sampling
// issue 07). Preserve the native failure; only the OS knows what is retryable.
func (c *traceConn) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sequence++
	n, err := c.Conn.Write(p)
	if c.trace && err != nil {
		digest := sha256.Sum256(p)
		c.writes = append(c.writes, SocketWrite{Sequence: c.sequence, Offset: c.offset, Bytes: len(p), Written: n,
			SHA256: hex.EncodeToString(digest[:]), Error: err.Error()})
	}
	c.offset += int64(n)
	return n, err
}

type traceContextKey struct{}

func traceContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, traceContextKey{}, c.(*traceConn))
}

func startWriteTrace(ctx context.Context) int64 {
	c := ctx.Value(traceContextKey{}).(*traceConn)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trace, c.writes = true, nil
	return c.id
}

func finishWriteTrace(ctx context.Context) []SocketWrite {
	c := ctx.Value(traceContextKey{}).(*traceConn)
	c.mu.Lock()
	defer c.mu.Unlock()
	writes := c.writes
	c.trace, c.writes = false, nil
	return writes
}

package provider

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Issue 18 captured Darwin loopback ENOBUFS in the Provider's TCP writes,
// followed by the Client's unexpected EOF. Go retries EAGAIN but returns
// ENOBUFS immediately. Keep socket backpressure below net/http's sticky write
// error: resume only the unwritten suffix, on this connection, for at most one
// second per Write. No HTTP request or generated frame is replayed. Persistent
// ENOBUFS, other errors, deadlines and close remain failures. Remove this
// test-infrastructure boundary if Go gains equivalent bounded handling.
const writeBufferBudget = time.Second

type writeBufferListener struct{ net.Listener }

func (l writeBufferListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &writeBufferConn{Conn: c, closed: make(chan struct{})}, nil
}

type writeBufferConn struct {
	net.Conn
	writeMu   sync.Mutex
	closed    chan struct{}
	closeOnce sync.Once
	waits     atomic.Int64
}

func (c *writeBufferConn) Write(p []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	until := time.Now().Add(writeBufferBudget)
	written := 0
	for {
		n, err := c.Conn.Write(p[written:])
		written += n
		if !errors.Is(err, syscall.ENOBUFS) || written == len(p) || !time.Now().Before(until) {
			return written, err
		}
		c.waits.Add(1)
		timer := time.NewTimer(min(time.Millisecond, time.Until(until)))
		select {
		case <-timer.C:
		case <-c.closed:
			timer.Stop()
			return written, net.ErrClosed
		}
	}
}

func (c *writeBufferConn) Close() error {
	c.closeOnce.Do(func() { close(c.closed) })
	return c.Conn.Close()
}

type writeBufferContextKey struct{}

func writeBufferContext(ctx context.Context, c net.Conn) context.Context {
	return context.WithValue(ctx, writeBufferContextKey{}, c.(*writeBufferConn))
}

func writeBufferWaits(ctx context.Context) int64 {
	c, _ := ctx.Value(writeBufferContextKey{}).(*writeBufferConn)
	if c == nil {
		return 0
	}
	return c.waits.Load()
}

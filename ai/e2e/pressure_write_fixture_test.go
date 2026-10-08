package e2e

import (
	"net"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// Inject at the controlled Provider's socket, after HTTP headers and the first
// stream frame. Client request building, SDK, framing and transport stay real.
type pressureFaultListener struct {
	net.Listener
	mode   string
	hit    chan struct{}
	once   sync.Once
	faults atomic.Int64
}

func (l *pressureFaultListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &pressureFaultConn{Conn: c, listener: l}, nil
}

type pressureFaultConn struct {
	net.Conn
	listener *pressureFaultListener
	writes   int
}

func (c *pressureFaultConn) Write(p []byte) (int, error) {
	// Honor underlying close/deadline even while injecting an OS error.
	if _, err := c.Conn.Write(nil); err != nil {
		return 0, err
	}
	c.writes++
	l := c.listener
	if c.writes < 2 || ((l.mode == "transient" || l.mode == "partial") && c.writes > 2) {
		return c.Conn.Write(p)
	}
	l.faults.Add(1)
	l.once.Do(func() { close(l.hit) })
	switch l.mode {
	case "other":
		return 0, syscall.EIO
	case "deadline":
		if c.writes == 2 {
			if err := c.Conn.SetWriteDeadline(time.Now().Add(20 * time.Millisecond)); err != nil {
				return 0, err
			}
		}
	case "closed":
		_ = c.Conn.Close()
	case "partial":
		n, err := c.Conn.Write(p[:len(p)/2])
		if err != nil {
			return n, err
		}
		return n, syscall.ENOBUFS
	}
	return 0, syscall.ENOBUFS
}

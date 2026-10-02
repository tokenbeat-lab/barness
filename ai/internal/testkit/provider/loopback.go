package provider

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// LoopbackTransport is the test transport injected at Client construction. It
// refuses to dial anything but loopback addresses, so an offline E2E run can
// never reach the internet even if configuration leaks a real endpoint.
func LoopbackTransport() *http.Transport {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ip := net.ParseIP(host)
			if ip == nil || !ip.IsLoopback() {
				return nil, fmt.Errorf("loopback transport: refusing non-loopback address %q", addr)
			}
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConnsPerHost: 16,
	}
}

// StalledDialTransport is a transport whose connection attempts never
// complete: every dial waits until the request's context ends, as a dial to
// an upstream that drops SYNs does. Nothing is ever sent.
func StalledDialTransport() *http.Transport {
	return &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			<-ctx.Done()
			return nil, ctx.Err()
		},
	}
}

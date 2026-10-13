//go:build live

package live

import (
	"errors"
	"net"
	"net/url"
	"strconv"
)

// Explicit Host transport configuration, never ambient HTTP(S)_PROXY. Keep
// endpoint identity and normal TLS verification in the existing smoke client.
func smokeProxy(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid explicit loopback HTTP proxy")
	}
	port, err := strconv.Atoi(u.Port())
	ip := net.ParseIP(u.Hostname())
	if err != nil || port < 1 || port > 65535 || ip == nil || !ip.IsLoopback() || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("invalid explicit loopback HTTP proxy")
	}
	return u, nil
}

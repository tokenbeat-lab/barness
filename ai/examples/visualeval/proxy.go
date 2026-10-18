package main

import (
	"errors"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Keep standalone Host configuration independent from the Chinese evaluator
// and smoke runner: their evidence contracts and lifecycles evolve separately.
// This only selects a transport; official endpoints and TLS checks stay fixed.
func visualProxy(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return nil, errors.New("invalid proxy")
	}
	ip := net.ParseIP(u.Hostname())
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 || ip == nil || !ip.IsLoopback() || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(raw, "#") || (u.Path != "" && u.Path != "/") {
		return nil, errors.New("explicit loopback HTTP proxy required")
	}
	return u, nil
}

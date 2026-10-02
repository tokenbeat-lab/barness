package ai

import (
	"net/http"
)

// userAgent is what barness-ai sends as User-Agent on every Provider request
// instead of an SDK default. It deliberately carries no host OS, release or
// architecture details: in multi-tenant cloud use they would describe the
// shared host to every vendor account (maintainer decision 2026-10-01; pi's
// own "pi (<os> <release>; <arch>)" is an approved difference in the pi
// differential ledger).
const userAgent = "barness-ai"

// newHTTPClient builds the one HTTP client shared by all calls of a Client.
// It carries no tenant state: no cookie jar, no default credentials, and it
// never follows redirects, so a credential-bearing request cannot be replayed
// to another origin. Authentication is attached per attempt by the adapter.
// Every round trip is bounded by the policy's attempt time limits.
func newHTTPClient(rt http.RoundTripper, p *ResourcePolicy) *http.Client {
	if rt == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		// The default transport reads HTTP(S)_PROXY from the environment; the
		// core must not take network identity from ambient configuration.
		t.Proxy = nil
		rt = t
	}
	return &http.Client{
		Transport: &watchedTransport{next: rt, connect: p.ConnectTimeout, header: p.ResponseHeaderTimeout, readIdle: p.ReadIdleTimeout},
		Jar:       nil,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

package ai

import (
	"net/http"
)

// newHTTPClient builds the one HTTP client shared by all calls of a Client.
// It carries no tenant state: no cookie jar, no default credentials, and it
// never follows redirects, so a credential-bearing request cannot be replayed
// to another origin. Authentication is attached per attempt by the adapter.
func newHTTPClient(rt http.RoundTripper) *http.Client {
	if rt == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		// The default transport reads HTTP(S)_PROXY from the environment; the
		// core must not take network identity from ambient configuration.
		t.Proxy = nil
		rt = t
	}
	return &http.Client{
		Transport: rt,
		Jar:       nil,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

//go:build live

package live

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Controlled transport retains the selected service request at the public
// Client/SDK seam. Its immutable mode stays controlled throughout all native
// origin receipts and can never qualify a service or catalog promotion.
func installRegionalControlled(rec *recorder, raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || net.ParseIP(u.Hostname()) == nil || !net.ParseIP(u.Hostname()).IsLoopback() || u.Port() == "" {
		return fmt.Errorf("regional controlled target must be explicit loopback HTTP")
	}
	rec.next = regionalControlledTransport{target: u, next: provider.LoopbackTransport()}
	return nil
}

type regionalControlledTransport struct {
	target *url.URL
	next   http.RoundTripper
}

func (t regionalControlledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	u := *req.URL
	u.Scheme, u.Host = t.target.Scheme, t.target.Host
	clone.URL = &u
	return t.next.RoundTrip(clone)
}

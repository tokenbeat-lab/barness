//go:build live

package live

import (
	"fmt"
	"io"
	"net/http"
)

// Enforce final native resources before the recorder delegates to any network.
// This second boundary accounts for Hook changes and all failed HTTP attempts.
type multimodalGate struct {
	route  multimodalRoute
	next   http.RoundTripper
	report *multimodalExecution
}

func (g *multimodalGate) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() != g.route.endpoint+g.route.path || req.Method != "POST" || req.GetBody == nil {
		return nil, fmt.Errorf("unapproved candidate service")
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer body.Close()
	raw, err := io.ReadAll(io.LimitReader(body, (4<<20)+1))
	if err != nil || len(raw) > 4<<20 {
		return nil, fmt.Errorf("native body budget exceeded")
	}
	actual, err := multimodalNativeResources(g.route, string(raw))
	if err != nil {
		return nil, err
	}
	actual.Attempts = 1
	prospective := g.report.Budget.Actual.add(actual)
	if !prospective.fits(g.report.Budget.Reserved) || !prospective.fits(g.report.Budget.Limits) {
		return nil, fmt.Errorf("final native cumulative budget exceeded")
	}
	g.report.Budget.Actual = prospective
	return g.next.RoundTrip(req)
}

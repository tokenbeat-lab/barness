package supportmatrix

import (
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/livegrant"
	"time"
)

func VerifyProcessAuthorization(dir, route, account, region string, started, finished time.Time) error {
	g, raw, r, e := livegrant.ReadEvidence(dir)
	if e != nil {
		return e
	}
	want, e := ProcessScope(route, "live", "formal", g.Scope.BinarySHA256)
	if e != nil {
		return e
	}
	if region == "" {
		region = g.RegionAlias
	}
	return livegrant.VerifyEvidence(g, raw, r, want, account, region, started, finished)
}

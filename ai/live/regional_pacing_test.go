//go:build live

package live

import (
	"context"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/supportmatrix"
	"time"
)

// Pacing belongs to this one native qualification process. It precedes both
// reservation and the per-call network timeout, including the first call.
// No global account state, retry, or public model payload is involved.
func (p *regionalProcess) waitInterval(ctx context.Context) error {
	if !supportmatrix.RegionalIsKimi(p.report.Combo) {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.grant.CallIntervalMs == 0 {
		return nil
	}
	timer := time.NewTimer(time.Duration(p.grant.CallIntervalMs) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ctx.Err()
	}
}

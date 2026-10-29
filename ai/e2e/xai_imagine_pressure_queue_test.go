package e2e

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Fixed six-call schedule: two active owners, two queued calls, one refused
// overflow, and one recovery call. A canceled waiter never reaches HTTP.
func imaginePressureQueue(t *testing.T, ev *evidence.Case) {
	p := newImaginePressureWorld(t, func(c *ai.Config) {
		c.Policy.MaxConcurrentPerTenant = 1
		c.Policy.MaxConcurrentProcess = 2
		c.Policy.MaxAdmissionWaiters = 2
	})
	ctx, cancel := context.WithTimeout(context.Background(), pressureDeadline)
	held, release := make(chan string, 2), make(chan struct{})
	var once sync.Once
	unblock := func() { once.Do(func() { close(release) }) }
	var wg sync.WaitGroup
	defer wg.Wait()
	defer cancel()
	defer unblock()
	type terminal struct {
		res ai.ImagesResult
		err error
	}
	owners := map[string]chan terminal{}
	waiters := map[string]chan terminal{}
	ownerCancel := map[string]context.CancelFunc{}
	waiterCancel := map[string]context.CancelFunc{}
	images := make([]ai.Image, 10)
	for i := range images {
		images[i] = imaginePressureImage(t, 512<<10, 80+i)
	}
	req := ai.ImagesRequest{Prompt: "pressure prompt queue"}
	for i := range 5 {
		req.ReferenceImages = append(req.ReferenceImages, imaginePressureImage(t, 256<<10, 96+i))
	}
	for _, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		p.provider.EnqueueAt(imaginePressurePrefix(tenant, true), imaginePressureReply(images[:1], 900000000))
		ownCtx, ownCancel := context.WithCancel(ctx)
		ownerCancel[tenant] = ownCancel
		result := make(chan terminal, 1)
		owners[tenant] = result
		wg.Go(func() {
			res, err := p.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
				held <- tenant
				select {
				case <-release:
				case <-ctx.Done():
				}
				return nil
			}}).GenerateImages(ownCtx, imaginePressureScope("pressure-owner-"+tenant, tenant), imagineTarget(), req, imaginePressureOptions(1))
			result <- terminal{res, err}
		})
	}
	for range 2 {
		waitValue(ev, "owner holds body", held)
	}
	ev.Check("both tenants own one body and permit", p.gauges.Permits() == 2 && p.bodies.Open() == 2, "permits=%d bodies=%d", p.gauges.Permits(), p.bodies.Open())
	// Only B's waiter and A's eventual recovery have a response script.
	p.provider.EnqueueAt(imaginePressurePrefix(tenantB.tenant, false), imaginePressureReply(images, 4000000000))
	p.provider.EnqueueAt(imaginePressurePrefix(tenantA.tenant, false), imaginePressureReply(images, 4000000000))
	for _, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		waitCtx, waitCancel := context.WithCancel(ctx)
		waiterCancel[tenant] = waitCancel
		result := make(chan terminal, 1)
		waiters[tenant] = result
		wg.Go(func() {
			res, err := p.client.GenerateImages(waitCtx, imaginePressureScope("pressure-waiter-"+tenant, tenant), imagineTarget(), ai.ImagesRequest{Prompt: "pressure prompt queued"}, imaginePressureOptions(10))
			result <- terminal{res, err}
		})
	}
	eventually(ev, "finite waiter queue fills exactly", func() bool { return p.gauges.AdmissionWaiters() == 2 })
	overflow := within(ev, "overflow is refused without waiting", func() terminal {
		r, e := p.client.GenerateImages(ctx, imaginePressureScope("pressure-overflow", tenantA.tenant), imagineTarget(), ai.ImagesRequest{Prompt: "pressure prompt overflow"}, imaginePressureOptions(10))
		return terminal{r, e}
	})
	ev.Check("overflow refuses before any HTTP attempt", errors.Is(overflow.err, &ai.Error{Code: ai.CodeResourceLimit}) && len(overflow.res.Metadata.Attempts) == 0 && len(overflow.res.Content) == 0, "err=%v", overflow.err)
	imaginePressureSummary(ev, "overflow", overflow.res, overflow.err)
	p.records.awaitCall(ev, overflow.res.Metadata.RequestID)
	waiterCancel[tenantA.tenant]()
	waiting := waitValue(ev, "A waiter cancels", waiters[tenantA.tenant])
	ev.Check("waiting cancellation sends nothing", errors.Is(waiting.err, &ai.Error{Code: ai.CodeCanceled}) && len(waiting.res.Metadata.Attempts) == 0, "err=%v", waiting.err)
	ownerCancel[tenantA.tenant]()
	first := waitValue(ev, "A owner cancellation releases its body", owners[tenantA.tenant])
	ev.Check("active cancellation retains one accepted attempt", errors.Is(first.err, &ai.Error{Code: ai.CodeCanceled}) && len(first.res.Metadata.Attempts) == 1 && len(first.res.Content) == 0, "err=%v", first.err)
	ev.Check("other tenant keeps its independent active call", p.gauges.Permits() == 1 && p.bodies.Open() == 1 && p.gauges.AdmissionWaiters() == 1, "permits=%d bodies=%d waiters=%d", p.gauges.Permits(), p.bodies.Open(), p.gauges.AdmissionWaiters())
	unblock()
	second := waitValue(ev, "B owner succeeds", owners[tenantB.tenant])
	last := waitValue(ev, "B waiter proceeds after release", waiters[tenantB.tenant])
	ev.Check("B survives A cancellation", second.err == nil && len(second.res.Content) == 1 && last.err == nil && len(last.res.Content) == 10, "owner=%v waiter=%v", second.err, last.err)
	recovery, recoveryErr := p.client.GenerateImages(ctx, imaginePressureScope("pressure-recovery", tenantA.tenant), imagineTarget(), ai.ImagesRequest{Prompt: "pressure prompt recovered"}, imaginePressureOptions(10))
	ev.Check("released capacity serves next A call", recoveryErr == nil && len(recovery.Content) == 10, "err=%v", recoveryErr)
	for name, outcome := range map[string]terminal{"owner-a": first, "owner-b": second, "waiter-a": waiting, "waiter-b": last, "recovery": {recovery, recoveryErr}} {
		imaginePressureSummary(ev, name, outcome.res, outcome.err)
		observations := p.records.awaitCall(ev, outcome.res.Metadata.RequestID)
		for _, o := range observations {
			ev.Check("queue Observer stays with its tenant", o.Call.TenantID == outcome.res.Metadata.TenantID, "call=%s", name)
		}
	}
	requests := p.provider.Requests()
	ev.Check("six logical calls make exactly four HTTP attempts", len(requests) == 4, "requests=%d", len(requests))
	for _, r := range requests {
		tenant := tenantA.tenant
		alias := xaiKey.alias
		if r.KeyAlias == tenantB.alias {
			tenant = tenantB.tenant
			alias = tenantB.alias
		}
		ev.Check("queued HTTP remains in its credential scope", r.KeyAlias == alias && (r.Path == imaginePressurePrefix(tenant, true) || r.Path == imaginePressurePrefix(tenant, false)), "path=%s", r.Path)
	}
	ev.Record("queue-load", map[string]any{"logicalCalls": 6, "httpAttempts": len(requests), "tenants": 2, "maxConcurrentProcess": 2, "maxConcurrentPerTenant": 1, "maxWaiters": 2, "canceledActive": 1, "canceledWaiting": 1, "overflow": 1, "realService": "NOT_RUN"})
	p.released(ev)
}

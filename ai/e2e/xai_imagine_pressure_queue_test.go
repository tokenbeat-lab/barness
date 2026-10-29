package e2e

import (
	"context"
	"errors"
	"math"
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
	owners, waiters := map[string]chan terminal{}, map[string]chan terminal{}
	ownerCancel, waiterCancel := map[string]context.CancelFunc{}, map[string]context.CancelFunc{}
	calls := map[string]*imaginePressureCall{}
	for n, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		// Each tenant has distinct response bytes. Expected hashes are fixed
		// before routing, independently of captured credentials and results.
		images := make([]ai.Image, 10)
		for i := range images {
			images[i] = imaginePressureImage(t, 512<<10, 80+n*16+i)
		}
		for _, role := range []string{"owner", "waiter"} {
			id := "pressure-" + role + "-" + tenant
			c := &imaginePressureCall{id: id, tenant: tenant, req: ai.ImagesRequest{Prompt: "pressure prompt " + id}, opts: imaginePressureOptions(10), amount: .40}
			output, ticks := images, int64(4000000000)
			if role == "owner" {
				c.opts, c.amount = imaginePressureOptions(1), .09
				output, ticks = images[:1], 900000000
				for i := range 5 {
					c.req.ReferenceImages = append(c.req.ReferenceImages, imaginePressureImage(t, 256<<10, 128+n*16+i))
				}
			}
			for _, img := range output {
				c.want = append(c.want, imaginePressureHash(img.Data))
			}
			c.reply = imaginePressureReply(output, ticks)
			calls[id] = c
		}
		if tenant == tenantA.tenant {
			queued := calls["pressure-waiter-"+tenant]
			recovery := *queued
			recovery.id, recovery.req.Prompt = "pressure-recovery", "pressure prompt pressure-recovery"
			calls[recovery.id] = &recovery
			overflow := *queued
			overflow.id, overflow.req.Prompt = "pressure-overflow", "pressure prompt pressure-overflow"
			calls[overflow.id] = &overflow
		}
	}
	for _, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		c := calls["pressure-owner-"+tenant]
		p.provider.EnqueueAt(imaginePressurePrefix(tenant, true), c.reply)
		ownCtx, ownCancel := context.WithCancel(ctx)
		ownerCancel[tenant] = ownCancel
		result := make(chan terminal, 1)
		owners[tenant] = result
		wg.Go(func() {
			res, err := p.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, scope ai.CallScope, _ ai.ResponseInfo) error {
				ev.Check("held response hook retains planned scope", scope.RequestID == c.id && scope.TenantID == c.tenant, "call=%s", c.id)
				held <- tenant
				select {
				case <-release:
				case <-ctx.Done():
				}
				return nil
			}}).GenerateImages(ownCtx, imaginePressureScope(c.id, c.tenant), imagineTarget(), c.req, c.opts)
			result <- terminal{res, err}
		})
	}
	for range 2 {
		waitValue(ev, "owner holds body", held)
	}
	ev.Check("both tenants own one body and permit", p.gauges.Permits() == 2 && p.bodies.Open() == 2, "permits=%d bodies=%d", p.gauges.Permits(), p.bodies.Open())
	// Only B's waiter and A's eventual recovery have a response script.
	p.provider.EnqueueAt(imaginePressurePrefix(tenantB.tenant, false), calls["pressure-waiter-"+tenantB.tenant].reply)
	p.provider.EnqueueAt(imaginePressurePrefix(tenantA.tenant, false), calls["pressure-recovery"].reply)
	for _, tenant := range []string{tenantA.tenant, tenantB.tenant} {
		c := calls["pressure-waiter-"+tenant]
		waitCtx, waitCancel := context.WithCancel(ctx)
		waiterCancel[tenant] = waitCancel
		result := make(chan terminal, 1)
		waiters[tenant] = result
		wg.Go(func() {
			res, err := p.client.GenerateImages(waitCtx, imaginePressureScope(c.id, c.tenant), imagineTarget(), c.req, c.opts)
			result <- terminal{res, err}
		})
	}
	eventually(ev, "finite waiter queue fills exactly", func() bool { return p.gauges.AdmissionWaiters() == 2 })
	overflow := within(ev, "overflow is refused without waiting", func() terminal {
		c := calls["pressure-overflow"]
		r, e := p.client.GenerateImages(ctx, imaginePressureScope(c.id, c.tenant), imagineTarget(), c.req, c.opts)
		return terminal{r, e}
	})
	ev.Check("overflow refuses before any HTTP attempt", errors.Is(overflow.err, &ai.Error{Code: ai.CodeResourceLimit}) && len(overflow.res.Metadata.Attempts) == 0 && len(overflow.res.Content) == 0, "err=%v", overflow.err)
	waiterCancel[tenantA.tenant]()
	waiting := waitValue(ev, "A waiter cancels", waiters[tenantA.tenant])
	ev.Check("waiting cancellation sends nothing", errors.Is(waiting.err, &ai.Error{Code: ai.CodeCanceled}) && len(waiting.res.Metadata.Attempts) == 0 && len(waiting.res.Content) == 0, "err=%v", waiting.err)
	ownerCancel[tenantA.tenant]()
	first := waitValue(ev, "A owner cancellation releases its body", owners[tenantA.tenant])
	ev.Check("active cancellation retains one accepted attempt", errors.Is(first.err, &ai.Error{Code: ai.CodeCanceled}) && len(first.res.Metadata.Attempts) == 1 && len(first.res.Content) == 0, "err=%v", first.err)
	ev.Check("other tenant keeps its independent active call", p.gauges.Permits() == 1 && p.bodies.Open() == 1 && p.gauges.AdmissionWaiters() == 1, "permits=%d bodies=%d waiters=%d", p.gauges.Permits(), p.bodies.Open(), p.gauges.AdmissionWaiters())
	unblock()
	second := waitValue(ev, "B owner succeeds", owners[tenantB.tenant])
	last := waitValue(ev, "B waiter proceeds after release", waiters[tenantB.tenant])
	ev.Check("B survives A cancellation", second.err == nil && len(second.res.Content) == 1 && last.err == nil && len(last.res.Content) == 10, "owner=%v waiter=%v", second.err, last.err)
	c := calls["pressure-recovery"]
	recovery, recoveryErr := p.client.GenerateImages(ctx, imaginePressureScope(c.id, c.tenant), imagineTarget(), c.req, c.opts)
	ev.Check("released capacity serves next A call", recoveryErr == nil && len(recovery.Content) == 10, "err=%v", recoveryErr)
	outcomes := map[string]terminal{"pressure-owner-" + tenantA.tenant: first, "pressure-owner-" + tenantB.tenant: second, "pressure-waiter-" + tenantA.tenant: waiting, "pressure-waiter-" + tenantB.tenant: last, "pressure-overflow": overflow, "pressure-recovery": {recovery, recoveryErr}}
	attempts := map[string]int{"pressure-owner-" + tenantA.tenant: 1, "pressure-owner-" + tenantB.tenant: 1, "pressure-waiter-" + tenantA.tenant: 0, "pressure-waiter-" + tenantB.tenant: 1, "pressure-overflow": 0, "pressure-recovery": 1}
	for id, outcome := range outcomes {
		c := calls[id]
		c.result, c.err = outcome.res, outcome.err
		imaginePressureSummary(ev, id, c.result, c.err)
		ev.Check("queue attempt count follows independent schedule", len(c.result.Metadata.Attempts) == attempts[id], "call=%s attempts=%d", id, len(c.result.Metadata.Attempts))
		for _, a := range c.result.Metadata.Attempts {
			ev.Check("queue retains one accepted HTTP response and missing tokens", a.HTTPStatus == 200 && a.UsageReporting == ai.UsageUnreported && a.Usage == (ai.Usage{}) && c.result.Usage == (ai.Usage{}), "call=%s", id)
			if c.err == nil {
				amount, known := a.ReportedCostUSD.Get()
				lo, lok := a.CostEstimate.Minimum.Get()
				hi, hok := a.CostEstimate.Maximum.Get()
				ev.Check("successful queue charge and estimate match planned amount", known && math.Abs(amount-c.amount) < 1e-12 && lok && hok && math.Abs(lo.Total-c.amount) < 1e-12 && math.Abs(hi.Total-c.amount) < 1e-12 && a.CostEstimate.Status == ai.CostResolved, "call=%s", id)
			} else {
				ev.Check("cancel before decoding invents no charge or estimate", a.ReportedCostUSD == (ai.Nullable[float64]{}) && a.CostEstimate.Status == ai.CostUnreported, "call=%s", id)
			}
		}
		if c.err == nil {
			ev.Check("queue output count matches planned tenant", len(c.result.Content) == len(c.want), "call=%s", id)
			for i, block := range c.result.Content {
				img, ok := block.(ai.ImageOutputImage)
				ev.Check("queue preserves tenant-specific ordered output hash", ok && i < len(c.want) && img.MimeType == "image/png" && imaginePressureHash(img.Data) == c.want[i], "call=%s image=%d", id, i)
			}
		}
		p.checkObserver(ev, *c, true)
	}
	requests, seen := p.provider.Requests(), map[string]int{}
	ev.Check("six logical calls make exactly four HTTP attempts", len(requests) == 4, "requests=%d", len(requests))
	for _, r := range requests {
		var wire struct{ Prompt string }
		mustUnmarshal(t, r.Body, &wire)
		matched := false
		for id, c := range calls {
			if wire.Prompt != c.req.Prompt {
				continue
			}
			matched = true
			seen[id]++
			ev.Check("planned prompt pins HTTP tenant credential and endpoint", r.KeyAlias == imaginePressureAlias(c.tenant) && r.Path == imaginePressurePrefix(c.tenant, len(c.req.ReferenceImages) > 0) && r.Method == "POST" && r.Query == "", "call=%s alias=%s path=%s", id, r.KeyAlias, r.Path)
			ev.Record(id+"-wire", map[string]any{"request": imaginePressureDigest(r.Body), "scriptedResponse": imaginePressureDigest([]byte(c.reply.Script)), "path": r.Path, "keyAlias": r.KeyAlias})
		}
		ev.Check("HTTP prompt belongs to one planned call", matched, "unknown prompt")
	}
	for id := range calls {
		ev.Check("only independently scheduled calls reach HTTP once", seen[id] == attempts[id], "call=%s requests=%d", id, seen[id])
	}
	ev.Record("queue-load", map[string]any{"logicalCalls": 6, "httpAttempts": len(requests), "tenants": 2, "maxConcurrentProcess": 2, "maxConcurrentPerTenant": 1, "maxWaiters": 2, "canceledActive": 1, "canceledWaiting": 1, "overflow": 1, "realService": "NOT_RUN"})
	p.released(ev)
}

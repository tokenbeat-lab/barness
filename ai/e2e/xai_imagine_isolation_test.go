package e2e

import (
	"context"
	"errors"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
)

func TestXAIImagineTenantAdmissionIsolation(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-tenant-admission-isolation")
	g := probe.New()
	records := newRecorder()
	w := newImagineWorld(t, func(c *ai.Config) {
		c.Policy.MaxConcurrentPerTenant = 1
		c.Policy.MaxConcurrentProcess = 2
		c.Probe = g
		c.Observer = records
		c.Policy.MaxQueuedObservations = 64
	})
	b := imagineBinding()
	b.TenantID = tenantB.tenant
	b.AccountScopeID = "acct-imagine-b"
	w.host.PutBinding(b)
	w.host.PutCredential(ai.Credential{OwnerTenantID: tenantB.tenant, CredentialID: b.CredentialRef, Version: "b-v1", AccountScopeID: b.AccountScopeID, Active: true, APIKey: ai.NewSecret(tenantB.secret)})
	reply := imagineReply(`{"data":[{"b64_json":"` + imaginePNG + `"}],"usage":{"cost_in_usd_ticks":400000000,"input_tokens":10,"output_tokens":8,"total_tokens":18,"input_tokens_details":{"text_tokens":6,"image_tokens":4,"cached_tokens":2},"output_tokens_details":{"text_tokens":2,"image_tokens":3,"reasoning_tokens":3}}}`)
	enqueue(ev, w, reply, reply, reply)
	held, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	type terminal struct {
		result ai.ImagesResult
		err    error
	}
	owner := make(chan terminal, 1)
	go func() {
		r, e := w.client.WithHooks(ai.Hooks{OnResponse: func(ctx context.Context, _ ai.CallScope, _ ai.ResponseInfo) error {
			close(held)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}}).GenerateImages(ctxFor(t), textScope("imagine-owner"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw tenant A."}, nil)
		owner <- terminal{r, e}
	}()
	waitFor(ev, "owner holds response permit", held)
	ctx, cancel := context.WithCancel(ctxFor(t))
	defer cancel()
	waiter := make(chan terminal, 1)
	go func() {
		r, e := w.client.GenerateImages(ctx, textScope("imagine-waiter"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw waiting A."}, nil)
		waiter <- terminal{r, e}
	}()
	eventually(ev, "same-tenant call enters bounded queue", func() bool { return g.AdmissionWaiters() == 1 })
	cancel()
	waiting := waitValue(ev, "waiter cancels", waiter)
	var failure *ai.Error
	ev.Check("waiter has no vendor attempt", errors.As(waiting.err, &failure) && failure.Code == ai.CodeCanceled && len(waiting.result.Metadata.Attempts) == 0, "err=%v metadata=%+v", waiting.err, waiting.result.Metadata)
	scope := textScope("imagine-other-tenant")
	scope.TenantID = tenantB.tenant
	other, err := w.client.GenerateImages(ctxFor(t), scope, imagineTarget(), ai.ImagesRequest{Prompt: "Draw tenant B."}, nil)
	ev.Check("other tenant receives independent result", err == nil && other.Metadata.TenantID == tenantB.tenant && other.Metadata.AccountScopeID == "acct-imagine-b", "err=%v metadata=%+v", err, other.Metadata)
	close(release)
	first := waitValue(ev, "owner finishes", owner)
	ev.Check("owner succeeds", first.err == nil, "err=%v", first.err)
	again, err := w.client.GenerateImages(ctxFor(t), textScope("imagine-followup"), imagineTarget(), ai.ImagesRequest{Prompt: "Draw after release."}, nil)
	ev.Check("canceled waiter releases capacity", err == nil && len(again.Content) == 1, "err=%v", err)
	ev.Record("requests", w.provider.Requests())
	ev.Record("owner-result", first.result)
	ev.Record("other-result", other)
	ev.Record("waiter-result", waiting.result)
	requests := w.provider.Requests()
	ev.Check("credentials do not cross tenants", len(requests) == 3 && requests[0].KeyAlias == xaiKey.alias && requests[1].KeyAlias == tenantB.alias && requests[2].KeyAlias == xaiKey.alias, "requests=%+v", requests)
	observations := records.awaitCall(ev, "imagine-owner")
	first.result.Metadata.Attempts[0].Usage.Modalities.InputText = ai.Value(int64(999))
	first.result.Usage.Modalities.InputText = ai.Value(int64(998))
	for _, o := range observations {
		if o.Attempt != nil && o.Kind == ai.ObservationAttemptFinished {
			n, _ := o.Attempt.Usage.Modalities.InputText.Get()
			ev.Check("Observer owns token facts", n == 6, "attempt=%+v", o.Attempt)
		}
	}
	ev.Record("observations", observations)
	eventually(ev, "resources return to zero", func() bool { return g.ActiveCalls() == 0 && g.Permits() == 0 && g.AdmissionWaiters() == 0 })
}

package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestOperationCatalogChatIntersection(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*ai.Catalog, *ai.Binding)
		code   ai.Code
	}{
		{"outside-whitelist", func(_ *ai.Catalog, b *ai.Binding) { b.AllowedModels = []string{"other"} }, ai.CodeTenantDenied},
		{"image-and-classifier-only", func(c *ai.Catalog, _ *ai.Binding) { c.Models = nil }, ai.CodeInvalidRequest},
	} {
		for _, entry := range outcomeEntries {
			t.Run(tc.name+"/"+entry.name, func(t *testing.T) {
				ev := run.Case(t, "E03-operation-intersection-"+tc.name+"-"+entry.name)
				catalog := operationCatalog()
				binding := primaryBinding(tenantA, "https://example.invalid")
				tc.change(&catalog, &binding)
				w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &catalog }, tenantA)
				binding.Endpoint = w.provider.URL() + "/v1"
				w.host.PutBinding(binding)
				f, _ := loadTextFixture(t, "text-basic.json")
				scope := textScope("req-operation-intersection-" + tc.name + "-" + entry.name)
				o := entry.invoke(ctxFor(t), w, scope, textTarget(f), textRequest(f))
				o.record(ev)
				checkRejected(ev, o, tc.code, ai.PhaseCapability)
				checkUnresolved(ev, o.result, scope)
				reads := w.host.ReadsFor(scope.RequestID)
				ev.Record("host_reads", reads)
				ev.Check("intersection refusal reads no credential and sends nothing", reads.Credentials == 0 && len(w.provider.Requests()) == 0, "reads=%+v", reads)
			})
		}
	}
}

func TestOperationSameNamedBindings(t *testing.T) {
	for _, op := range []ai.Operation{ai.OperationImage, ai.OperationClassifier} {
		t.Run(string(op), func(t *testing.T) {
			ev := run.Case(t, "E03-operation-same-named-bindings-"+string(op))
			rec := newRecorder()
			w := observedWorld(t, rec, nil, tenantA, tenantB)
			w.updateBinding(tenantB, func(b *ai.Binding) { b.Operation = op })
			f, raw := loadTextFixture(t, "text-basic.json")
			ev.Fixture("text-basic.json", raw)
			enqueue(ev, w, sseReply(t, f, provider.FramingLF))
			scopeA := textScope("req-operation-a-" + string(op))
			scopeB := textScope("req-operation-b-" + string(op))
			scopeB.TenantID = tenantB.tenant
			// A shared Client receives both tenants concurrently; their same
			// BindingID must retain tenant-local operation authorization.
			results := make(chan outcome, 1)
			go func() {
				r, err := w.client.Complete(ctxFor(t), scopeA, textTarget(f), textRequest(f), nil)
				results <- outcome{result: r, err: err}
			}()
			b, err := w.client.Complete(ctxFor(t), scopeB, textTarget(f), textRequest(f), nil)
			a := <-results
			a.record(ev)
			ev.Record("other-tenant", b)
			checkSucceeded(ev, f, a)
			checkRejected(ev, outcome{result: b, err: err}, ai.CodeTenantDenied, ai.PhaseCapability)
			checkUnresolved(ev, b, scopeB)
			checkKeysSent(ev, w, tenantA)
			ev.Check("other tenant's same-named binding reads no credentials", w.host.ReadsFor(scopeB.RequestID).Credentials == 0, "reads=%+v", w.host.ReadsFor(scopeB.RequestID))
			checkCallRecords(ev, rec.awaitCall(ev, scopeA.RequestID), scopeA, "primary", a, 1)
			checkCallRecords(ev, rec.awaitCall(ev, scopeB.RequestID), scopeB, "primary", outcome{result: b, err: err}, 0)
		})
	}
}

func TestOperationRotationConflict(t *testing.T) {
	ev := run.Case(t, "E03-operation-rotation-conflict")
	w := newWorld(t, tenantA)
	f, _ := loadTextFixture(t, "text-basic.json")
	w.host.BeforeCredential(func() {
		w.updateBinding(tenantA, func(b *ai.Binding) { b.Operation, b.Version = ai.OperationImage, "b2" })
		w.host.PutCredential(primaryCredential(tenantA2, "v2"))
	})
	scope := textScope("req-operation-conflict")
	res, err := w.client.Complete(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
	ev.Record("conflicted-result", res)
	checkRejected(ev, outcome{result: res, err: err}, ai.CodeCredentialUnavailable, ai.PhaseConsistency)
	checkUnresolved(ev, res, scope)
	ev.Check("conflict has no attempts or Provider request", len(res.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "metadata=%+v", res.Metadata)
	reads := w.host.ReadsFor(scope.RequestID)
	ev.Record("conflicted-reads", reads)
	ev.Check("no internal re-resolution on rotation conflict", reads.Bindings == 1 && reads.Credentials == 1, "reads=%+v", reads)
	w.host.BeforeCredential(nil)
	scope.RequestID = "req-operation-after-conflict"
	next, err := w.client.Complete(ctxFor(t), scope, textTarget(f), textRequest(f), nil)
	ev.Record("next-result", next)
	checkRejected(ev, outcome{result: next, err: err}, ai.CodeTenantDenied, ai.PhaseCapability)
	ev.Check("new call sees new operation before credentials", w.host.ReadsFor(scope.RequestID).Credentials == 0 && len(w.provider.Requests()) == 0, "reads=%+v", w.host.ReadsFor(scope.RequestID))
}

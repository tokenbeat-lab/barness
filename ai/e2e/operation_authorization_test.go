package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// Issue 04's agreed seams are NewClient/catalog discovery and the four chat
// entries. Failures to cover before implementation: unknown operation,
// operation mismatch, whitelist/catalog intersection, duplicate full keys,
// contradictory modalities/capabilities, invalid classifier ranges, missing
// or invalid rates, mutable configuration/query leakage, same-named bindings
// across tenants, and binding/credential rotation conflicts.
// GenerateImages/Classify cross-entry tests belong to their delivery slices.
func TestChatOperationAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name  string
		op    ai.Operation
		code  ai.Code
		phase ai.Phase
	}{
		{"unknown", "images", ai.CodeInvalidRequest, ai.PhaseBinding},
		{"image", ai.OperationImage, ai.CodeTenantDenied, ai.PhaseCapability},
		{"classifier", ai.OperationClassifier, ai.CodeTenantDenied, ai.PhaseCapability},
	} {
		for _, entry := range outcomeEntries {
			t.Run(tc.name+"/"+entry.name, func(t *testing.T) {
				ev := run.Case(t, "E03-operation-refuse-"+tc.name+"-"+entry.name)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				f, raw := loadTextFixture(t, "text-basic.json")
				ev.Fixture("text-basic.json", raw)
				w.updateBinding(tenantA, func(b *ai.Binding) {
					b.Operation = tc.op
					// A wrong API/model must not obscure the operation refusal.
					b.API, b.AllowedModels = "unknown", nil
				})
				scope := textScope("req-operation-" + tc.name + "-" + entry.name)
				o := entry.invoke(ctxFor(t), w, scope, textTarget(f), textRequest(f))
				o.record(ev)
				checkRejected(ev, o, tc.code, tc.phase)
				checkUnresolved(ev, o.result, scope)
				checkNoSecrets(ev, o)
				reads := w.host.ReadsFor(scope.RequestID)
				ev.Record("host_reads", reads)
				ev.Check("operation refused without reading credentials", reads.Bindings == 1 && reads.Credentials == 0, "got %+v", reads)
				ev.Check("operation refused without Provider requests or attempts", len(w.provider.Requests()) == 0 && len(o.result.Metadata.Attempts) == 0, "metadata=%+v", o.result.Metadata)
				obs := rec.awaitCall(ev, scope.RequestID)
				checkCallRecords(ev, obs, scope, "primary", o, 0)
				ev.Check("failed chat entry retains chat attribution", o.result.Metadata.Operation == ai.OperationChat, "got %q", o.result.Metadata.Operation)
			})
		}
	}

	for _, op := range []ai.Operation{"", ai.OperationChat} {
		for _, entry := range outcomeEntries {
			t.Run("allowed-"+string(op)+"/"+entry.name, func(t *testing.T) {
				ev := run.Case(t, "E03-operation-allowed-"+string(op)+"-"+entry.name)
				rec := newRecorder()
				w := observedWorld(t, rec, nil, tenantA)
				f, raw := loadTextFixture(t, "text-basic.json")
				ev.Fixture("text-basic.json", raw)
				w.updateBinding(tenantA, func(b *ai.Binding) { b.Operation = op })
				enqueue(ev, w, sseReply(t, f, provider.FramingLF))
				scope := textScope("req-operation-allowed-" + string(op) + "-" + entry.name)
				o := entry.invoke(ctxFor(t), w, scope, textTarget(f), textRequest(f))
				o.record(ev)
				checkSucceeded(ev, f, o)
				checkCallRecords(ev, rec.awaitCall(ev, scope.RequestID), scope, "primary", o, 1)
				ev.Check("legacy and explicit chat bindings attribute chat", o.result.Metadata.Operation == ai.OperationChat, "got %q", o.result.Metadata.Operation)
			})
		}
	}
}

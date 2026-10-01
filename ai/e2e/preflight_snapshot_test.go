package e2e

import (
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestConfigurationSnapshot covers when credential updates and revocations
// take effect (spec I3, ADR-0003, E07): every new logical call resolves
// afresh, a call in flight keeps the snapshot it was built with, and a call
// that cannot obtain a consistent snapshot fails without re-resolving.
func TestConfigurationSnapshot(t *testing.T) {
	decoy := polluteEnvironment(t)

	t.Run("rotation-between-reads-proceeds-with-new-key", func(t *testing.T) {
		// A consistent new credential version is not a conflict: same tenant,
		// account and reference, only the key moved on.
		for _, e := range outcomeEntries {
			t.Run(e.name, func(t *testing.T) {
				ev, w, f := startHardeningCase(t, "E07-snapshot-rotation-between-reads-"+e.name)
				enqueue(ev, w, sseReply(t, f, provider.FramingLF))
				w.host.BeforeCredential(func() { w.host.PutCredential(primaryCredential(tenantA2, "v2")) })
				o := e.invoke(ctxFor(t), w, textScope("req-rotate-"+e.name), textTarget(f), textRequest(f))
				o.record(ev)
				checkSucceeded(ev, f, o)
				checkKeysSent(ev, w, tenantA2)
			})
		}
	})

	t.Run("new-calls-see-k1-then-k2", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "E07-snapshot-new-calls-see-k1-then-k2")
		enqueue(ev, w, sseReply(t, f, provider.FramingLF), sseReply(t, f, provider.FramingLF))
		first, err1 := w.client.Complete(ctxFor(t), textScope("req-k1"), textTarget(f), textRequest(f), nil)
		w.host.PutCredential(primaryCredential(tenantA2, "v2"))
		second, err2 := w.client.Complete(ctxFor(t), textScope("req-k2"), textTarget(f), textRequest(f), nil)
		ev.Record("results", []ai.Result{first, second})
		ev.Check("both calls succeed", err1 == nil && err2 == nil, "err1=%v err2=%v", err1, err2)
		checkKeysSent(ev, w, tenantA, tenantA2)
	})

	t.Run("in-flight-call-keeps-snapshot", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "E07-snapshot-in-flight-call-keeps-snapshot")
		reply := sseReply(t, f, provider.FramingBytewise)
		// Once the first request has reached the Provider, rotate to K2 and then
		// revoke the credential outright.
		reply.OnReceive = func() {
			w.host.PutCredential(primaryCredential(tenantA2, "v2"))
			revoked := primaryCredential(tenantA2, "v3")
			revoked.Active = false
			w.host.PutCredential(revoked)
		}
		enqueue(ev, w, reply)
		s := w.client.Stream(ctxFor(t), textScope("req-in-flight"), textTarget(f), textRequest(f), nil)
		inFlight := checkStreamedText(ev, f, s)
		ev.Check("in-flight call completed with its original identity", inFlight.Metadata.Resolved &&
			inFlight.Metadata.AccountScopeID == "acct-"+tenantA.tenant, "metadata=%+v", inFlight.Metadata)

		res, err := w.client.Complete(ctxFor(t), textScope("req-after-revoke"), textTarget(f), textRequest(f), nil)
		ev.Record("after_revoke", res)
		checkFailed(ev, res, err)
		checkRejected(ev, outcome{result: res, err: err}, ai.CodeCredentialUnavailable, ai.PhaseCredential)
		ev.Check("revocation blocked the new call", len(w.provider.Requests()) == 1, "got %d", len(w.provider.Requests()))
		checkKeysSent(ev, w, tenantA)
	})

	t.Run("conflict-fails-then-new-request-id-sees-new-snapshot", func(t *testing.T) {
		ev, w, f := startHardeningCase(t, "E07-snapshot-conflict-then-new-request-id")
		enqueue(ev, w, sseReply(t, f, provider.FramingLF))
		// Between the first call's two reads the tenant moves the binding to a
		// new account with a new credential version: b1 can no longer be paired
		// with a credential consistently.
		var fired bool
		w.host.BeforeCredential(func() {
			if fired {
				return
			}
			fired = true
			w.updateBinding(tenantA, func(b *ai.Binding) { b.Version, b.AccountScopeID = "b2", "acct-tenant-a-2" })
			c := primaryCredential(tenantA2, "v2")
			c.AccountScopeID = "acct-tenant-a-2"
			w.host.PutCredential(c)
		})

		failed, err := w.client.Complete(ctxFor(t), textScope("req-conflict"), textTarget(f), textRequest(f), nil)
		ev.Record("conflicted", failed)
		ev.Record("conflicted_error", errString(err))
		checkFailed(ev, failed, err)
		checkRejected(ev, outcome{result: failed, err: err}, ai.CodeCredentialUnavailable, ai.PhaseConsistency)
		checkUnresolved(ev, failed, textScope("req-conflict"))
		checkNoSecrets(ev, outcome{result: failed, err: err})
		ev.Check("no inference request for the conflicted call", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
		reads := w.host.ReadsFor("req-conflict")
		ev.Record("conflicted_reads", reads)
		ev.Check("no internal re-resolution", reads.Bindings == 1 && reads.Credentials == 1, "got %+v", reads)

		retried, err := w.client.Complete(ctxFor(t), textScope("req-conflict-retry"), textTarget(f), textRequest(f), nil)
		ev.Record("retried", retried)
		ev.Check("host's new logical call succeeds", err == nil, "err=%v", err)
		ev.Check("new call carries the new consistent snapshot", retried.Metadata.AccountScopeID == "acct-tenant-a-2",
			"metadata=%+v", retried.Metadata)
		checkKeysSent(ev, w, tenantA2)
	})

	ev := run.Case(t, "E07-snapshot-no-environment-fallback")
	ev.Record("decoy_requests", decoy.Requests())
	ev.Check("environment endpoints and proxies received nothing", len(decoy.Requests()) == 0, "got %d", len(decoy.Requests()))
}

// TestRequestSurface proves at the type level that nothing a caller submits
// per call can carry a key, credential reference, endpoint, Authorization,
// header, proxy, transport or executable function (spec I3).
func TestRequestSurface(t *testing.T) {
	ev := run.Case(t, "E07-request-surface-carries-no-authority")
	// Every value a call accepts, including each concrete message kind behind
	// the Message and content interfaces.
	inputs := []any{
		ai.CallScope{}, ai.Target{}, ai.Request{}, ai.ResponsesOptions{}, ai.SimpleOptions{},
		ai.UserMessage{}, ai.Text{},
	}
	forbidden := []string{"key", "secret", "credential", "endpoint", "url", "auth", "header", "proxy", "transport", "dial", "cookie", "project"}
	var fields []string
	for _, in := range inputs {
		walkFields(reflect.TypeOf(in), func(path string, f reflect.StructField) {
			fields = append(fields, path)
			name := strings.ToLower(f.Name)
			for _, word := range forbidden {
				ev.Check("no authority-bearing field", !strings.Contains(name, word), "%s matches %q", path, word)
			}
			switch f.Type.Kind() {
			case reflect.Func, reflect.Chan, reflect.UnsafePointer:
				ev.Check("no executable or channel field", false, "%s is %s", path, f.Type)
			}
		})
	}
	ev.Record("fields", fields)
}

// walkFields visits every struct field reachable from t, recursing through
// pointers, slices, arrays and maps.
func walkFields(t reflect.Type, visit func(path string, f reflect.StructField)) {
	seen := map[reflect.Type]bool{}
	var walk func(t reflect.Type, prefix string)
	walk = func(t reflect.Type, prefix string) {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct || seen[t] {
			return
		}
		seen[t] = true
		for i := range t.NumField() {
			f := t.Field(i)
			path := prefix + t.Name() + "." + f.Name
			visit(path, f)
			walk(f.Type, path+">")
		}
	}
	walk(t, "")
}

func checkSucceeded(ev *evidence.Case, f textFixture, o outcome) {
	ev.Check("call succeeds", o.err == nil, "err=%v", o.err)
	if o.streamed {
		ev.Check("stream ends in done", len(o.events) > 0 && o.events[len(o.events)-1].Type() == ai.EventDone,
			"events=%q", eventSummary(o.events))
		ev.Check("Err is nil", o.streamErr == nil, "got %v", o.streamErr)
	}
	checkFinalMessage(ev, f, o.result.Message)
}

// checkKeysSent asserts the Provider received exactly one request per key, in
// order.
func checkKeysSent(ev *evidence.Case, w *world, keys ...tenantKey) {
	reqs := w.provider.Requests()
	ev.Record("requests", reqs)
	got := make([]string, len(reqs))
	for i, r := range reqs {
		got[i] = r.KeyAlias
	}
	want := make([]string, len(keys))
	for i, k := range keys {
		want[i] = k.alias
	}
	ev.Check("keys sent in order", reflect.DeepEqual(got, want), "got %q want %q", got, want)
}

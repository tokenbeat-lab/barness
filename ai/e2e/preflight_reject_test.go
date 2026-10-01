package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// backendLeak is what a failing secret backend puts in its error text. It
// stands for backend internals (paths, tokens) that must never surface.
const backendLeak = "vault://kv/tenant-a/openai token=hvs.BACKEND-INTERNAL"

// rejection is one way preflight must refuse a call before any inference
// request (spec I3, E07). Every rejection is enumerated here before the
// implementation and runs through all four public entry points.
type rejection struct {
	id    string
	code  ai.Code
	phase ai.Phase
	// scope and target adjust the default tenant-A call; arrange adjusts the
	// host before the call.
	scope   func(*ai.CallScope)
	target  func(*ai.Target)
	arrange func(w *world)
}

var rejections = []rejection{
	// Scope.
	{id: "missing-tenant-id", code: ai.CodeInvalidRequest, phase: ai.PhaseScope,
		scope: func(s *ai.CallScope) { s.TenantID = "" }},
	{id: "missing-request-id", code: ai.CodeInvalidRequest, phase: ai.PhaseScope,
		scope: func(s *ai.CallScope) { s.RequestID = "" }},

	// Binding.
	{id: "binding-not-found", code: ai.CodeBindingNotFound, phase: ai.PhaseBinding,
		target: func(t *ai.Target) { t.BindingID = "does-not-exist" }},
	{id: "binding-of-another-tenant", code: ai.CodeBindingNotFound, phase: ai.PhaseBinding,
		target: func(t *ai.Target) { t.BindingID = "b-only" },
		arrange: func(w *world) {
			b := primaryBinding(tenantB, w.provider.URL())
			b.BindingID = "b-only"
			w.host.PutBinding(b)
		}},
	{id: "host-returns-another-tenants-binding", code: ai.CodeBindingNotFound, phase: ai.PhaseBinding,
		arrange: func(w *world) {
			w.host.PutBindingAt(tenantA.tenant, "primary", primaryBinding(tenantB, w.provider.URL()))
		}},
	{id: "host-returns-another-binding-id", code: ai.CodeBindingNotFound, phase: ai.PhaseBinding,
		arrange: func(w *world) {
			b := primaryBinding(tenantA, w.provider.URL())
			b.BindingID = "secondary"
			w.host.PutBindingAt(tenantA.tenant, "primary", b)
		}},
	{id: "actor-not-permitted", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
		arrange: func(w *world) { w.host.RestrictActors(tenantA.tenant, "primary", "actor-admin") }},
	{id: "binding-disabled", code: ai.CodeTenantDenied, phase: ai.PhaseBinding,
		arrange: func(w *world) { w.updateBinding(tenantA, func(b *ai.Binding) { b.Enabled = false }) }},
	{id: "binding-auth-kind-unsupported", code: ai.CodeInvalidRequest, phase: ai.PhaseBinding,
		arrange: func(w *world) { w.updateBinding(tenantA, func(b *ai.Binding) { b.AuthKind = "oauth" }) }},

	// Capability.
	{id: "model-not-allowed-by-binding", code: ai.CodeTenantDenied, phase: ai.PhaseCapability,
		target: func(t *ai.Target) { t.ModelID = "gpt-4.1" }},
	{id: "model-not-in-catalog", code: ai.CodeInvalidRequest, phase: ai.PhaseCapability,
		target: func(t *ai.Target) { t.ModelID = "gpt-legacy-x" }},

	// Credential.
	{id: "credential-missing", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
		arrange: func(w *world) { w.host.DeleteCredential(tenantA.tenant, "cred-"+tenantA.tenant) }},
	{id: "credential-revoked", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
		arrange: func(w *world) { w.replaceCredential(tenantA, func(c *ai.Credential) { c.Active = false }) }},
	{id: "credential-empty-key", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
		arrange: func(w *world) { w.replaceCredential(tenantA, func(c *ai.Credential) { c.APIKey = ai.NewSecret("") }) }},
	{id: "secret-backend-failure", code: ai.CodeCredentialUnavailable, phase: ai.PhaseCredential,
		arrange: func(w *world) { w.host.FailCredentials(errors.New("backend down: " + backendLeak)) }},
	{id: "credential-access-denied", code: ai.CodeTenantDenied, phase: ai.PhaseCredential,
		arrange: func(w *world) { w.host.FailCredentials(ai.ErrAccessDenied) }},

	// Consistency between the binding and credential snapshots (D2).
	{id: "binding-updated-between-reads", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) {
			w.host.BeforeCredential(func() {
				w.updateBinding(tenantA, func(b *ai.Binding) { b.Version = "b2" })
			})
		}},
	{id: "binding-disabled-between-reads", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) {
			w.host.BeforeCredential(func() {
				w.updateBinding(tenantA, func(b *ai.Binding) { b.Version, b.Enabled = "b2", false })
			})
		}},
	{id: "credential-rehomed-between-reads", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) {
			w.host.BeforeCredential(func() {
				w.replaceCredential(tenantA, func(c *ai.Credential) { c.Version, c.AccountScopeID = "v2", "acct-elsewhere" })
			})
		}},
	{id: "credential-of-another-tenant", code: ai.CodeTenantDenied, phase: ai.PhaseConsistency,
		arrange: func(w *world) {
			c := primaryCredential(tenantB, "v1")
			c.CredentialID = "cred-" + tenantA.tenant
			w.host.PutCredentialAt(tenantA.tenant, "cred-"+tenantA.tenant, c)
		}},
	{id: "credential-not-the-referenced-one", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) {
			c := primaryCredential(tenantA, "v1")
			c.CredentialID = "cred-other"
			w.host.PutCredentialAt(tenantA.tenant, "cred-"+tenantA.tenant, c)
		}},
	{id: "credential-unversioned", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) { w.replaceCredential(tenantA, func(c *ai.Credential) { c.Version = "" }) }},
	{id: "binding-unversioned", code: ai.CodeCredentialUnavailable, phase: ai.PhaseConsistency,
		arrange: func(w *world) { w.updateBinding(tenantA, func(b *ai.Binding) { b.Version = "" }) }},
}

// TestPreflightRejections proves each rejection ends the call with an error
// assistant message, a distinguishable Code/Phase, unresolved identity, no
// leaked secret, no internal re-resolution and zero inference requests — on
// every entry point, with a polluted environment that offers a fallback.
func TestPreflightRejections(t *testing.T) {
	decoy := polluteEnvironment(t)
	for _, r := range rejections {
		for _, e := range outcomeEntries {
			t.Run(r.id+"/"+e.name, func(t *testing.T) {
				ev, w, f := startHardeningCase(t, "E07-reject-"+r.id+"-"+e.name)
				enqueue(ev, w, sseReply(t, f, provider.FramingLF))
				if r.arrange != nil {
					r.arrange(w)
				}
				scope := textScope("req-" + r.id + "-" + e.name)
				if r.scope != nil {
					r.scope(&scope)
				}
				target := textTarget(f)
				if r.target != nil {
					r.target(&target)
				}

				o := e.invoke(ctxFor(t), w, scope, target, textRequest(f))
				o.record(ev)
				ev.Record("host_reads", w.host.ReadsFor(scope.RequestID))
				ev.Check("no inference request sent", len(w.provider.Requests()) == 0, "got %d", len(w.provider.Requests()))
				checkFailed(ev, o.result, o.err)
				checkRejected(ev, o, r.code, r.phase)
				checkUnresolved(ev, o.result, scope)
				checkNoSecrets(ev, o)
				reads := w.host.ReadsFor(scope.RequestID)
				ev.Check("each resolver read at most once (no internal re-resolution)",
					reads.Bindings <= 1 && reads.Credentials <= 1, "got %+v", reads)
			})
		}
	}
	ev := run.Case(t, "E07-reject-no-environment-fallback")
	ev.Record("decoy_requests", decoy.Requests())
	ev.Check("environment endpoints and proxies received nothing", len(decoy.Requests()) == 0, "got %d", len(decoy.Requests()))
}

// outcome is what a caller observes from one call on any entry point.
type outcome struct {
	result ai.Result
	err    error
	// Stream entry points only.
	streamed  bool
	events    []ai.Event
	streamErr error
}

func (o outcome) record(ev *evidence.Case) {
	ev.Record("result", o.result)
	ev.Record("error", errString(o.err))
	if o.streamed {
		ev.Record("events", eventsJSON(o.events))
	}
}

type outcomeEntry struct {
	name   string
	invoke func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome
}

// outcomeEntries are the four public entry points, reporting whatever
// happens instead of asserting success.
var outcomeEntries = []outcomeEntry{
	{"stream-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
		return drain(w.client.Stream(ctx, scope, target, req, nil))
	}},
	{"stream-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
		return drain(w.client.StreamSimple(ctx, scope, target, req, ai.SimpleOptions{}))
	}},
	{"complete-full", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
		res, err := w.client.Complete(ctx, scope, target, req, ai.ResponsesOptions{})
		return outcome{result: res, err: err}
	}},
	{"complete-simple", func(ctx context.Context, w *world, scope ai.CallScope, target ai.Target, req ai.Request) outcome {
		res, err := w.client.CompleteSimple(ctx, scope, target, req, ai.SimpleOptions{})
		return outcome{result: res, err: err}
	}},
}

func drain(s *ai.Stream) outcome {
	defer s.Close()
	o := outcome{streamed: true}
	for s.Next() {
		o.events = append(o.events, s.Event())
	}
	o.streamErr = s.Err()
	o.result, o.err = s.Result()
	return o
}

// checkRejected asserts the classification and, on streams, that the only
// event is the error terminal carrying the same error.
func checkRejected(ev *evidence.Case, o outcome, code ai.Code, phase ai.Phase) {
	ev.Check("errors.Is matches code and phase", errors.Is(o.err, &ai.Error{Code: code, Phase: phase}),
		"want %s/%s, got %v", code, phase, o.err)
	var aiErr *ai.Error
	if errors.As(o.err, &aiErr) {
		ev.Check("errors.As exposes code and phase", aiErr.Code == code && aiErr.Phase == phase,
			"got %s/%s", aiErr.Code, aiErr.Phase)
		ev.Check("message errorMessage is the error's message", o.result.Message.ErrorMessage == aiErr.Message,
			"message=%q error=%q", o.result.Message.ErrorMessage, aiErr.Message)
	}
	if !o.streamed {
		return
	}
	got := eventSummary(o.events)
	ev.Check("stream ends with only an error terminal", reflect.DeepEqual(got, []string{"error:error"}), "got %q", got)
	ev.Check("Err equals the Result error", o.streamErr == o.err, "Err=%v Result=%v", o.streamErr, o.err)
	if len(o.events) == 1 {
		e, ok := o.events[0].(ai.ErrorEvent)
		ev.Check("error event carries the Result's message and error",
			ok && reflect.DeepEqual(e.Message, o.result.Message) && error(e.Err) == o.err, "event=%+v", o.events[0])
	}
}

// checkUnresolved asserts a preflight failure reports the trusted scope but
// no provider, API, model or account, rather than echoing request values.
func checkUnresolved(ev *evidence.Case, res ai.Result, scope ai.CallScope) {
	m := res.Metadata
	ev.Check("scope attribution kept", m.TenantID == scope.TenantID && m.RequestID == scope.RequestID &&
		m.ActorID == scope.ActorID && m.JobID == scope.JobID, "metadata=%+v", m)
	ev.Check("identity reported unresolved",
		!m.Resolved && m.ProviderID == "" && m.API == "" && m.ModelID == "" && m.AccountScopeID == "", "metadata=%+v", m)
	msg := res.Message
	ev.Check("message names no provider, API or model", msg.API == "" && msg.Provider == "" && msg.Model == "",
		"api=%q provider=%q model=%q", msg.API, msg.Provider, msg.Model)
}

// checkNoSecrets asserts no key or secret-backend internal appears in the
// message, events, Result or error text.
func checkNoSecrets(ev *evidence.Case, o outcome) {
	surface, err := json.Marshal(map[string]any{"result": o.result, "events": eventsJSON(o.events)})
	if err != nil {
		ev.Check("outcome encodes", false, "%v", err)
		return
	}
	text := string(surface) + errString(o.err) + errString(o.streamErr)
	secrets := []string{backendLeak, "hvs.BACKEND-INTERNAL", "sk-env-leak"}
	for _, k := range knownKeys {
		secrets = append(secrets, k.secret)
	}
	for _, s := range secrets {
		ev.Check("no secret in outcome", !strings.Contains(text, s), "found %q", s)
	}
}

// polluteEnvironment sets every key, endpoint and proxy variable a vendor SDK
// or Go's HTTP stack could fall back to, all pointing at a decoy Provider.
func polluteEnvironment(t *testing.T) *provider.Server {
	t.Helper()
	decoy := provider.New(map[string]string{"sk-env-leak": "key:environment"})
	t.Cleanup(decoy.Close)
	run.RedactSecret("sk-env-leak", "key:environment")
	for _, k := range []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "GOOGLE_API_KEY", "GEMINI_API_KEY"} {
		t.Setenv(k, "sk-env-leak")
	}
	for _, k := range []string{"OPENAI_BASE_URL", "ANTHROPIC_BASE_URL", "GOOGLE_GEMINI_BASE_URL"} {
		t.Setenv(k, decoy.URL()+"/v1")
	}
	for _, k := range []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "http_proxy", "https_proxy", "all_proxy"} {
		t.Setenv(k, decoy.URL())
	}
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	return decoy
}

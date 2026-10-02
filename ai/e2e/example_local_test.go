package e2e

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestLocalAssemblyExample runs the local assembly example (E10 本地装配,
// User Stories 2, 49) against the local controlled Provider: the program
// names where its key is, reads it itself and assembles one tenant and
// binding; barness-ai never reads the environment. The process environment
// is polluted with decoy keys, endpoints and proxies throughout.
func TestLocalAssemblyExample(t *testing.T) {
	f, raw := loadIsolationFixture(t)
	for _, p := range f.Protocols {
		for _, source := range []string{"env", "file"} {
			t.Run(p.ID+"/"+source, func(t *testing.T) {
				ev := run.Case(t, protocolCase(p)+"-E10-example-local-"+source)
				ev.Fixture("interleave.json", raw)
				decoy := polluteEnvironment(t)
				srv := localProvider(t)
				key := p.key(tenantA)
				cfg := localConfig(p, srv.URL())
				if source == "env" {
					t.Setenv("BARNESS_EXAMPLE_KEY", key.secret)
					cfg.Key = localassembly.KeySource{EnvVar: "BARNESS_EXAMPLE_KEY"}
				} else {
					path := filepath.Join(t.TempDir(), "provider.key")
					if err := os.WriteFile(path, []byte(key.secret+"\n"), 0o600); err != nil {
						t.Fatal(err)
					}
					cfg.Key = localassembly.KeySource{File: path}
				}
				ev.Record("assembly", map[string]any{"key_source": source, "binding": cfg.Binding, "policy": "localassembly.LocalPolicy"})
				local, err := localassembly.Open(cfg)
				if !ev.Check("the program assembles its Client", err == nil, "err=%v", err) {
					return
				}

				// The same script through the library's own test assembly is
				// the reference: the example must not change the outcome.
				solo := soloOutcome(t, p, tenantA)
				srv.Enqueue(p.success(t, tenantA), p.success(t, tenantA))
				ev.Record("response-script", []provider.Reply{p.success(t, tenantA)})

				scope1 := local.Scope()
				res, err := local.Client.Complete(ctxFor(t), scope1, local.Target(p.Model), p.request(), nil)
				ev.Record("complete", res)
				checkLocalTurn(ev, p, solo, res, err, scope1)

				scope2 := local.Scope()
				o := drain(local.Client.Stream(ctxFor(t), scope2, local.Target(p.Model), p.request(), nil))
				o.record(ev)
				checkLocalTurn(ev, p, solo, o.result, o.err, scope2)
				ev.Check("stream events equal the reference", reflect.DeepEqual(eventSummary(o.events), eventSummary(solo.events)),
					"got %q want %q", eventSummary(o.events), eventSummary(solo.events))
				ev.Check("every call has its own RequestID", scope1.RequestID != scope2.RequestID && scope1.RequestID != "", "got %q %q", scope1.RequestID, scope2.RequestID)

				reqs := srv.Requests()
				ev.Record("requests", reqs)
				ev.Check("two inference requests", len(reqs) == 2, "got %d", len(reqs))
				for _, r := range reqs {
					ev.Check("the named key was sent", r.KeyAlias == key.alias, "got %q want %q", r.KeyAlias, key.alias)
					ev.Check("the binding's endpoint was used", r.Path == localPath(p), "got %q", r.Path)
				}
				ev.Record("decoy_requests", decoy.Requests())
				ev.Check("environment keys, endpoints and proxies received nothing", len(decoy.Requests()) == 0, "got %d", len(decoy.Requests()))
			})
		}
	}

	// A source that is not set up fails assembly; there is no fallback to
	// the provider's conventional variable or another source.
	refusals := []struct {
		name  string
		setup func(t *testing.T) localassembly.KeySource
	}{
		{"env-unset", func(t *testing.T) localassembly.KeySource {
			_ = os.Unsetenv("BARNESS_EXAMPLE_KEY")
			return localassembly.KeySource{EnvVar: "BARNESS_EXAMPLE_KEY"}
		}},
		{"env-empty", func(t *testing.T) localassembly.KeySource {
			t.Setenv("BARNESS_EXAMPLE_KEY", " ")
			return localassembly.KeySource{EnvVar: "BARNESS_EXAMPLE_KEY"}
		}},
		{"file-missing", func(t *testing.T) localassembly.KeySource {
			return localassembly.KeySource{File: filepath.Join(t.TempDir(), "absent.key")}
		}},
		{"file-empty", func(t *testing.T) localassembly.KeySource {
			path := filepath.Join(t.TempDir(), "empty.key")
			if err := os.WriteFile(path, []byte("\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			return localassembly.KeySource{File: path}
		}},
		{"no-source", func(*testing.T) localassembly.KeySource { return localassembly.KeySource{} }},
		{"two-sources", func(t *testing.T) localassembly.KeySource {
			t.Setenv("BARNESS_EXAMPLE_KEY", tenantA.secret)
			return localassembly.KeySource{EnvVar: "BARNESS_EXAMPLE_KEY", File: "provider.key"}
		}},
	}
	for _, c := range refusals {
		t.Run("refused/"+c.name, func(t *testing.T) {
			ev := run.Case(t, "E10-example-local-refused-"+c.name)
			decoy := polluteEnvironment(t)
			t.Setenv("BARNESS_EXAMPLE_KEY", "")
			srv := localProvider(t)
			cfg := localConfig(f.Protocols[0], srv.URL())
			cfg.Key = c.setup(t)
			local, err := localassembly.Open(cfg)
			ev.Record("error", errString(err))
			ev.Check("assembly fails on its key source", local == nil && errors.Is(err, localassembly.ErrKeySource), "err=%v", err)
			ev.Check("nothing was sent", len(srv.Requests()) == 0 && len(decoy.Requests()) == 0, "provider %d decoy %d", len(srv.Requests()), len(decoy.Requests()))
		})
	}

	// The scope is the program's one tenant: another tenant's name or
	// another binding is not found by the library, and nothing is sent.
	t.Run("only-the-local-tenant", func(t *testing.T) {
		ev := run.Case(t, "E10-example-local-only-local-tenant")
		p := f.Protocols[0]
		srv := localProvider(t)
		t.Setenv("BARNESS_EXAMPLE_KEY", tenantA.secret)
		cfg := localConfig(p, srv.URL())
		cfg.Key = localassembly.KeySource{EnvVar: "BARNESS_EXAMPLE_KEY"}
		local, err := localassembly.Open(cfg)
		if !ev.Check("assembled", err == nil, "err=%v", err) {
			return
		}
		other := local.Scope()
		other.TenantID = tenantB.tenant
		res, err := local.Client.Complete(ctxFor(t), other, local.Target(p.Model), p.request(), nil)
		ev.Record("result", res)
		ev.Check("another tenant finds no binding", errors.Is(err, &ai.Error{Code: ai.CodeBindingNotFound}), "err=%v", err)
		res, err = local.Client.Complete(ctxFor(t), local.Scope(), ai.Target{BindingID: "other", ModelID: p.Model}, p.request(), nil)
		ev.Check("another binding is not found", errors.Is(err, &ai.Error{Code: ai.CodeBindingNotFound}) && !res.Metadata.Resolved, "err=%v", err)
		ev.Check("nothing was sent", len(srv.Requests()) == 0, "got %d", len(srv.Requests()))
	})
}

// localProvider is a local controlled Provider for the example.
func localProvider(t *testing.T) *provider.Server {
	t.Helper()
	aliases := map[string]string{}
	for _, k := range knownKeys {
		aliases[k.secret] = k.alias
		run.RedactSecret(k.secret, k.alias)
	}
	srv := provider.New(aliases)
	t.Cleanup(srv.Close)
	return srv
}

// localConfig is the example's configuration for p's binding at providerURL,
// without a key source. The test assembly allows the plain-http loopback
// Provider; a program sets neither field.
func localConfig(p isolationProtocol, providerURL string) localassembly.Config {
	b := ai.Binding{BindingID: p.BindingID, ProviderID: p.provider(), API: p.api(), Endpoint: providerURL + "/local",
		AccountScopeID: "acct-local", AllowedModels: []string{p.Model}}
	switch {
	case p.gemini():
		b.Endpoint += "/v1beta"
	case !p.anthropic() && !p.deepseek():
		b.Endpoint += "/v1"
	}
	return localassembly.Config{TenantID: "local", Binding: b,
		Transport: provider.LoopbackTransport(), AllowLoopbackHTTP: true}
}

func localPath(p isolationProtocol) string {
	switch {
	case p.anthropic():
		return "/local/v1/messages"
	case p.gemini():
		return "/local/v1beta/models/" + p.Model + ":streamGenerateContent"
	case p.chat():
		return "/local/v1/chat/completions"
	case p.deepseek():
		return "/local/responses"
	}
	return "/local/v1/responses"
}

// soloOutcome is k's success script for p streamed through the library's
// test assembly: the reference outcome.
func soloOutcome(t *testing.T, p isolationProtocol, k tenantKey) outcome {
	t.Helper()
	w := newWorld(t, k)
	w.provider.Enqueue(p.success(t, k))
	return drain(w.client.Stream(ctxFor(t), ai.CallScope{TenantID: k.tenant, RequestID: "req-reference"}, p.target(), p.request(), nil))
}

// checkLocalTurn asserts a local turn produced the reference message and is
// attributed to the local tenant, binding and account.
func checkLocalTurn(ev *evidence.Case, p isolationProtocol, solo outcome, res ai.Result, err error, scope ai.CallScope) {
	ev.Check("call succeeds", err == nil, "err=%v", err)
	got, want := res.Message, solo.result.Message
	env, trusted := got.NativeState.Envelope()
	ev.Check("native state vouched for the local tenant and account", trusted && env == ai.NativeStateEnvelope{TenantID: "local",
		AccountScopeID: "acct-local", ProviderID: p.provider(), API: p.api(), ModelID: p.Model}, "got %+v", env)
	// Provenance aside (the reference ran for another tenant), the message
	// is the library's own.
	got.Timestamp, want.Timestamp = 0, 0
	got.NativeState, want.NativeState = ai.TrustedNativeState{}, ai.TrustedNativeState{}
	ev.Check("message equals the reference", reflect.DeepEqual(got, want), "got %+v\nwant %+v", got, want)
	m := res.Metadata
	ev.Check("attributed to the local tenant and account", m.TenantID == "local" && m.RequestID == scope.RequestID &&
		m.BindingID == p.BindingID && m.Resolved && m.ProviderID == p.provider() && m.API == p.api() && m.ModelID == p.Model &&
		m.AccountScopeID == "acct-local", "metadata %+v", m)
}

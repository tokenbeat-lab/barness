package e2e

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/glmhost"
	"github.com/tokenbeat-lab/barness/ai/examples/toolloop"
)

// The example is exercised with the formal built-in catalog, not09's
// deliberately synthetic candidate assembly.
func TestGLMHostExample(t *testing.T) {
	for _, r := range glmSharedRoutes {
		for _, model := range []string{"glm-5.3", "glm-5.2"} {
			for _, entry := range []glmhost.Entry{glmhost.FullStream, glmhost.FullComplete, glmhost.SimpleStream, glmhost.SimpleComplete} {
				for _, hooked := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/hooks-%t", r.name, model, entry, hooked), func(t *testing.T) {
						ev := run.Case(t, fmt.Sprintf("P0-GLM13-host-%s-%s-%s-hooks-%t", r.name, model, entry, hooked))
						w := newGLMSharedWorld(t, func(c *ai.Config) { c.Catalog = nil })
						bindings, err := glmhost.Bindings(glmhost.Authorization{ProviderID: ai.ProviderZhipu, TenantID: glmA.tenant, AccountScopeID: "acct-" + glmA.tenant, CredentialRef: "same-credential", Version: "b1", Purpose: glmhost.Standard})
						if err != nil {
							t.Fatal(err)
						}
						for _, b := range bindings {
							if b.API == r.api {
								b.Endpoint = w.provider.URL() + r.prefix(glmA)
								w.host.PutBinding(b)
							}
						}
						req, _, answer, raw := r.script(t, model)
						req.Tools = nil
						ev.Fixture("own-wire.json", raw)
						w.provider.EnqueueAt(r.prefix(glmA), answer)
						hooks := ai.Hooks{}
						if hooked {
							hooks.TransformHeaders = func(_ context.Context, _ ai.CallScope, h http.Header) error {
								h.Set("X-Host-Trace", "safe-trace")
								return nil
							}
						}
						scope := ai.CallScope{TenantID: glmA.tenant, RequestID: "glm13-" + r.name + "-" + model + "-" + string(entry)}
						got, err := glmhost.Turn(ctxFor(t), w.client, scope, r.target(model), req, r.api, entry, hooks)
						outcome{result: got.Result, events: got.Events, streamed: len(got.Events) > 0, err: err}.record(ev)
						ev.Record("event-envelopes", got.Envelopes)
						w.record(ev, mixedOutcome{metadata: got.Result.Metadata, usage: got.Result.Message.Usage, chat: got.Result, err: err})
						ev.Record("bindings", bindings)
						ev.Record("requests", w.provider.Requests())
						ev.Check("formal catalog and explicit binding snapshot", err == nil && got.Result.Metadata.BindingVersion == "b1" && got.Result.Metadata.CredentialVersion == "v1" && got.Result.Metadata.ProviderID == ai.ProviderZhipu && got.Result.Metadata.API == r.api && got.Result.Metadata.ModelID == model, "err=%v metadata=%+v", err, got.Result.Metadata)
						m, ok := ai.BuiltinCatalog().Lookup(ai.ProviderZhipu, r.api, model)
						ev.Record("formal-model", m)
						ev.Record("resource-policy", validPolicy())
						ev.Check("formal independent model", ok && m.Cost.Output > 0, "model=%+v", m)
						requests := w.provider.Requests()
						suffix := "/chat/completions"
						if r.name == "responses" {
							suffix = "/responses"
						}
						if r.name == "messages" {
							suffix = "/v1/messages"
						}
						ev.Check("one own endpoint and key", len(requests) == 1 && requests[0].Path == r.prefix(glmA)+suffix && requests[0].KeyAlias == glmA.alias, "requests=%+v", requests)
						if hooked && len(requests) == 1 {
							ev.Check("trusted hook passed", requests[0].Header["X-Host-Trace"] == "safe-trace", "header missing")
						}
						w.released(ev)
					})
				}
			}
			t.Run(r.name+"/"+model+"/host-tools", func(t *testing.T) {
				ev := run.Case(t, "P0-GLM13-host-tools-"+r.name+"-"+model)
				w := newGLMSharedWorld(t, func(c *ai.Config) { c.Catalog = nil })
				req, first, last, raw := r.script(t, model)
				ev.Fixture("own-wire.json", raw)
				w.provider.EnqueueAt(r.prefix(glmA), first, last)
				called := 0
				ids := []string{}
				tools := []toolloop.Tool{}
				for _, decl := range req.Tools {
					tools = append(tools, toolloop.Tool{Declaration: decl, Run: func(context.Context, ai.ValidToolCall) (string, error) { called++; return "BRN42", nil }})
				}
				loop := toolloop.Loop{Client: w.client, Target: r.target(model), MaxTurns: 2, Scope: func() ai.CallScope {
					id := fmt.Sprintf("glm13-tool-%s-%s-%d", r.name, model, len(ids))
					ids = append(ids, id)
					return ai.CallScope{TenantID: glmA.tenant, RequestID: id}
				}}
				turns, err := loop.Run(ctxFor(t), req, tools)
				ev.Record("turns", turns)
				ev.Record("requests", w.provider.Requests())
				ev.Check("Host validated and ran tools only after complete turn", err == nil && len(turns) == 2 && called > 0 && len(ids) == 2 && ids[0] != ids[1], "err=%v turns=%d calls=%d", err, len(turns), called)
				w.released(ev)
			})
		}
	}
}

func TestGLMHostBindings(t *testing.T) {
	ev := run.Case(t, "P0-GLM13-host-bindings")
	got, assemblyErr := glmhost.Bindings(glmhost.Authorization{ProviderID: ai.ProviderZhipu, TenantID: "tenant", AccountScopeID: "account", CredentialRef: "standard-key", Version: "v7", Purpose: glmhost.Standard})
	if assemblyErr != nil {
		t.Fatal(assemblyErr)
	}
	want := map[ai.API]string{ai.APIOpenAIResponses: "https://open.bigmodel.cn/api/v1", ai.APIOpenAICompletions: "https://open.bigmodel.cn/api/paas/v4", ai.APIAnthropicMessages: "https://open.bigmodel.cn/api/anthropic"}
	ev.Record("bindings", got)
	for _, b := range got {
		ev.Check("official independent standard route", b.Operation == ai.OperationChat && b.ProviderID == ai.ProviderZhipu && b.Endpoint == want[b.API] && b.Version == "v7" && b.Enabled && b.TenantID == "tenant" && b.AccountScopeID == "account" && b.CredentialRef == "standard-key" && reflect.DeepEqual(b.AllowedModels, []string{"glm-5.3", "glm-5.2"}), "binding=%+v", b)
	}
	ev.Check("three explicit bindings", len(got) == 3, "got %d", len(got))
	got[0].AllowedModels[0] = "mutation"
	ev.Check("model allowlists independent", got[1].AllowedModels[0] == "glm-5.3" && got[2].AllowedModels[0] == "glm-5.3", "allowlist aliased")
	_, err := glmhost.Turn(ctxFor(t), nil, ai.CallScope{}, ai.Target{}, ai.Request{}, ai.APIOpenAIResponses, glmhost.Entry("wrong"), ai.Hooks{})
	ev.Check("invalid example entry rejected before Client", errors.Is(err, glmhost.ErrEntry), "err=%v", err)
}

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/glmhost"
	"github.com/tokenbeat-lab/barness/ai/examples/minimaxhost"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

type regionalHostWorld struct {
	*world
	bindings    map[ai.ProviderID][]ai.Binding
	secretReads map[ai.ProviderID]int
	gauges      *probe.Probe
	bodies      *provider.BodyTracker
}

func newRegionalHostWorld(t *testing.T) *regionalHostWorld {
	t.Helper()
	candidate, err := regional.CandidateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	cat := ai.BuiltinCatalog()
	cat.Version = "regional-host-qualified-candidate-2026-10-20.1"
	cat.Models = append(cat.Models, candidate.Models...)
	resolvers := map[ai.ProviderID]ai.CredentialResolver{}
	reads := map[ai.ProviderID]int{}
	gauges := probe.New()
	bodies := provider.TrackBodies(provider.LoopbackTransport())
	w := newWorldWith(t, func(c *ai.Config) {
		c.Catalog = &cat
		c.Probe = gauges
		c.Transport = bodies
		c.Credentials = credentialResolverFunc(func(ctx context.Context, s ai.CallScope, b ai.Binding) (ai.Credential, error) {
			r, ok := resolvers[b.ProviderID]
			if !ok {
				return ai.Credential{}, ai.ErrAccessDenied
			}
			return r.ResolveCredential(ctx, s, b)
		})
	}, tenantA)
	out := &regionalHostWorld{world: w, bindings: map[ai.ProviderID][]ai.Binding{}, secretReads: reads, gauges: gauges, bodies: bodies}
	for _, service := range []ai.ProviderID{"zhipu", "zai", "minimax", "minimax-cn"} {
		endpoints := map[ai.API]string{}
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			endpoints[api] = w.provider.URL() + "/regional/" + string(service) + "/" + string(api)
		}
		read := func(context.Context, string) (ai.Secret, error) {
			reads[service]++
			return ai.NewSecret(tenantA.secret), nil
		}
		var bindings []ai.Binding
		if service == "zhipu" || service == "zai" {
			a := glmhost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "same-account", CredentialRef: string(service) + "-record", Version: "v1", Purpose: glmhost.Standard, Endpoints: endpoints, TrustedGateway: true}
			bindings, err = glmhost.Bindings(a)
			resolvers[service] = glmhost.Credentials{Authorization: a, Record: glmhost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: a.Purpose, Version: "key-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}, ReadSecret: read}
		} else {
			a := minimaxhost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "same-account", CredentialRef: string(service) + "-record", Version: "v1", Purpose: minimaxhost.Standard, Endpoints: endpoints, TrustedGateway: true}
			bindings, err = minimaxhost.Bindings(a)
			resolvers[service] = minimaxhost.Credentials{Authorization: a, Record: minimaxhost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: a.Purpose, Version: "key-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}, ReadSecret: read}
		}
		if err != nil {
			t.Fatal(err)
		}
		out.bindings[service] = bindings
		for _, b := range bindings {
			w.host.PutBinding(b)
		}
	}
	// Caller mutations after construction cannot change the fixed Client catalog.
	for i := range cat.Models {
		cat.Models[i].Cost.Input = 999
		cat.Models[i].Input[0] = "mutated"
	}
	return out
}

// All scripts are synthetic controlled protocol fixtures, never relabelled
// captures from another market and never qualification evidence.
func regionalHostScript(t *testing.T, b ai.Binding, model string) (provider.Reply, []byte) {
	t.Helper()
	var frames []json.RawMessage
	switch b.API {
	case ai.APIOpenAICompletions:
		frames = []json.RawMessage{json.RawMessage(`{"id":"synthetic-own-chat","model":"` + model + `","choices":[{"index":0,"delta":{"role":"assistant","content":"Regional reply."},"finish_reason":null}]}`), json.RawMessage(`{"id":"synthetic-own-chat","model":"` + model + `","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":4,"total_tokens":12}}`)}
		return chatEvents(t, frames, ""), mustMarshal(t, frames)
	case ai.APIOpenAIResponses:
		frames = []json.RawMessage{
			json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"synthetic-own-message","type":"message","role":"assistant","content":[]}}`),
			json.RawMessage(`{"type":"response.content_part.added","item_id":"synthetic-own-message","output_index":0,"content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}`),
			json.RawMessage(`{"type":"response.output_text.delta","item_id":"synthetic-own-message","output_index":0,"content_index":0,"delta":"Regional reply."}`),
			json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":{"id":"synthetic-own-message","type":"message","role":"assistant","content":[{"type":"output_text","text":"Regional reply."}]}}`),
			json.RawMessage(`{"type":"response.completed","response":{"id":"synthetic-own-response","model":"` + model + `","status":"completed","output":[],"usage":{"input_tokens":8,"output_tokens":4,"total_tokens":12}}}`),
		}
	default:
		frames = []json.RawMessage{
			json.RawMessage(`{"type":"message_start","message":{"id":"synthetic-own-message","model":"` + model + `","type":"message","role":"assistant","content":[],"usage":{"input_tokens":8,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`),
			json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`),
			json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Regional reply."}}`),
			json.RawMessage(`{"type":"content_block_stop","index":0}`),
			json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"input_tokens":8,"output_tokens":4,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}`),
			json.RawMessage(`{"type":"message_stop"}`),
		}
	}
	return sseEvents(t, frames, provider.FramingLF), mustMarshal(t, frames)
}

type regionalCaller interface {
	Complete(context.Context, ai.CallScope, ai.Target, ai.Request, ai.Options) (ai.Result, error)
	Stream(context.Context, ai.CallScope, ai.Target, ai.Request, ai.Options) *ai.Stream
	CompleteSimple(context.Context, ai.CallScope, ai.Target, ai.Request, ai.SimpleOptions) (ai.Result, error)
	StreamSimple(context.Context, ai.CallScope, ai.Target, ai.Request, ai.SimpleOptions) *ai.Stream
}

func regionalHostCall(t *testing.T, c regionalCaller, b ai.Binding, model, entry, id string) outcome {
	t.Helper()
	scope := ai.CallScope{TenantID: b.TenantID, RequestID: id}
	target := ai.Target{BindingID: b.BindingID, ModelID: model}
	req := ai.Request{Messages: []ai.Message{ai.UserText("Synthetic regional Host input.")}}
	var opts ai.Options
	switch b.API {
	case ai.APIOpenAICompletions:
		opts = ai.ChatOptions{MaxTokens: ai.Value(64)}
	case ai.APIOpenAIResponses:
		opts = ai.ResponsesOptions{MaxTokens: ai.Value(64)}
	default:
		opts = ai.AnthropicOptions{MaxTokens: ai.Value(64)}
	}
	switch entry {
	case "stream":
		return drain(c.Stream(ctxFor(t), scope, target, req, opts))
	case "simple-stream":
		return drain(c.StreamSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(64)}))
	case "simple-complete":
		r, e := c.CompleteSimple(ctxFor(t), scope, target, req, ai.SimpleOptions{MaxTokens: ai.Value(64)})
		return outcome{result: r, err: e}
	default:
		r, e := c.Complete(ctxFor(t), scope, target, req, opts)
		return outcome{result: r, err: e}
	}
}

func regionalHostCaseID(b ai.Binding, model, entry string, hooked bool) string {
	return fmt.Sprintf("regional-host-%s-%s-%s-%s-hooks-%t", b.ProviderID, b.API, model, entry, hooked)
}

package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/kimihost"
	"github.com/tokenbeat-lab/barness/ai/internal/probe"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/kimi"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

type kimiHostWorld struct {
	*world
	bindings map[ai.ProviderID][]ai.Binding
	reads    map[ai.ProviderID]int
	gauges   *probe.Probe
	bodies   *provider.BodyTracker
}

func newKimiHostWorld(t *testing.T) *kimiHostWorld {
	t.Helper()
	cat := kimi.CandidateCatalog()
	resolvers := map[ai.ProviderID]ai.CredentialResolver{}
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
	out := &kimiHostWorld{world: w, bindings: map[ai.ProviderID][]ai.Binding{}, reads: map[ai.ProviderID]int{}, gauges: gauges, bodies: bodies}
	for _, service := range []ai.ProviderID{ai.ProviderMoonshot, ai.ProviderMoonshotCN} {
		endpoints := map[ai.API]string{}
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			endpoints[api] = w.provider.URL() + "/kimi-host/" + string(service) + "/" + string(api)
		}
		a := kimihost.Authorization{ProviderID: service, TenantID: tenantA.tenant, AccountScopeID: "same-account-label", CredentialRef: string(service) + "-record", Version: "binding-v1", Purpose: kimihost.Standard, Endpoints: endpoints, TrustedGateway: true}
		bindings, err := kimihost.Bindings(a)
		if err != nil {
			t.Fatal(err)
		}
		out.bindings[service] = bindings
		for _, b := range bindings {
			w.host.PutBinding(b)
		}
		resolvers[service] = kimihost.Credentials{Authorization: a, Record: kimihost.CredentialRecord{ProviderID: service, TenantID: a.TenantID, AccountScopeID: a.AccountScopeID, CredentialRef: a.CredentialRef, Purpose: a.Purpose, Version: "key-v1", BindingVersion: a.Version, Active: true, AllowedAPIs: []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages}}, ReadSecret: func(_ context.Context, ref string) (ai.Secret, error) {
			out.reads[service]++
			return ai.NewSecret(tenantA.secret), nil
		}}
	}
	return out
}

// Independent synthetic Kimi Host scripts, not relabelled vendor captures.
func kimiHostScript(t *testing.T, api ai.API) (provider.Reply, []byte) {
	t.Helper()
	var frames []json.RawMessage
	switch api {
	case ai.APIOpenAICompletions:
		frames = []json.RawMessage{json.RawMessage(`{"id":"kimi-host-chat","model":"kimi-k3","choices":[{"index":0,"delta":{"role":"assistant","content":"Ready."},"finish_reason":null}]}`), json.RawMessage(`{"choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":8,"completion_tokens":4,"total_tokens":12,"prompt_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"completion_tokens_details":{"reasoning_tokens":0}}}`)}
		return chatEvents(t, frames, ""), mustMarshal(t, frames)
	case ai.APIOpenAIResponses:
		frames = []json.RawMessage{
			json.RawMessage(`{"type":"response.output_item.added","output_index":0,"item":{"id":"kimi-host-message","type":"message","role":"assistant","content":[]}}`),
			json.RawMessage(`{"type":"response.content_part.added","item_id":"kimi-host-message","output_index":0,"content_index":0,"part":{"type":"output_text","text":"","annotations":[]}}`),
			json.RawMessage(`{"type":"response.output_text.delta","item_id":"kimi-host-message","output_index":0,"content_index":0,"delta":"Ready."}`),
			json.RawMessage(`{"type":"response.output_item.done","output_index":0,"item":{"id":"kimi-host-message","type":"message","role":"assistant","content":[{"type":"output_text","text":"Ready.","annotations":[]}]}}`),
			json.RawMessage(`{"type":"response.completed","response":{"id":"kimi-host-response","model":"kimi-k3","status":"completed","output":[{"id":"kimi-host-message","type":"message","role":"assistant","content":[{"type":"output_text","text":"Ready.","annotations":[]}]}],"usage":{"input_tokens":8,"output_tokens":4,"total_tokens":12,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}}}`),
		}
	default:
		frames = []json.RawMessage{
			json.RawMessage(`{"type":"message_start","message":{"id":"kimi-host-message","model":"kimi-k3","type":"message","role":"assistant","content":[],"usage":{"input_tokens":8,"output_tokens":0,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`),
			json.RawMessage(`{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`),
			json.RawMessage(`{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Ready."}}`),
			json.RawMessage(`{"type":"content_block_stop","index":0}`),
			json.RawMessage(`{"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":4}}`),
			json.RawMessage(`{"type":"message_stop"}`),
		}
	}
	return sseEvents(t, frames, provider.FramingLF), mustMarshal(t, frames)
}

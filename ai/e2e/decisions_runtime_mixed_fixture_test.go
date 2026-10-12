package e2e

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

var runtimeMixedRoutes = []mixedRoute{mixedRoutes[0], mixedRoutes[1], mixedRoutes[3], {"luna", ai.OperationClassifier, ai.ProviderOpenAI, ai.APIOpenAIDecisions}}

func newRuntimeMixedWorld(t *testing.T, shared bool, configure func(*ai.Config)) *mixedWorld {
	t.Helper()
	w := newMixedWorld(t, func(c *ai.Config) {
		c.Catalog.Version = "runtime-06-synthetic"
		m := c.Catalog.ClassifierModels[0]
		m.Provider, m.API, m.Name = ai.ProviderOpenAI, ai.APIOpenAIDecisions, "Synthetic Luna runtime candidate"
		m.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
		m.Capabilities.ImageMIMETypes = []string{"image/png", "image/jpeg", "image/webp"}
		m.Capabilities.MaxInputImages = 128
		m.Capabilities.MinChoices = 2
		m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
		m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: .1}}
		c.Catalog.ClassifierModels = append(c.Catalog.ClassifierModels, m)
		c.Policy.Classifier.Images = &ai.ClassifierImagePolicy{MaxImages: 128, MaxTotalImageBytes: 1 << 19}
		c.Policy.MaxQueuedObservations = 128
		if configure != nil {
			configure(c)
		}
	})
	for _, k := range []tenantKey{tenantA, tenantB} {
		for _, r := range runtimeMixedRoutes {
			b := r.binding(k, w.provider.URL())
			b.AllowClassifierImages = r.api == ai.APIOpenAIDecisions
			secret := k.secret
			if shared {
				b.AccountScopeID = "shared-" + string(r.provider)
				secret = tenantA.secret
			}
			w.host.PutBinding(b)
			w.host.PutCredential(ai.Credential{OwnerTenantID: k.tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "v1", Active: true, APIKey: ai.NewSecret(secret)})
		}
	}
	return w
}

func runtimeMixedInput(t *testing.T, k tenantKey, images bool) ai.ClassifierRequest {
	t.Helper()
	req := decisionsInput()
	req.State = json.RawMessage(`"synthetic private state ` + k.tenant + `"`)
	if images {
		assets := classifierAssets(t)
		if k.tenant == tenantB.tenant {
			req.Images = assets[1:2]
		} else {
			req.Images = assets[:1]
		}
	}
	return req
}

func runtimeMixedInvoke(t *testing.T, ctx context.Context, w *mixedWorld, r mixedRoute, k tenantKey, id string, hooks ai.Hooks, images bool) mixedOutcome {
	t.Helper()
	scope := ai.CallScope{TenantID: k.tenant, ActorID: "actor-" + k.tenant, JobID: "job-" + k.tenant, RequestID: id}
	target := ai.Target{BindingID: r.id, ModelID: mixedModel}
	if r.api != ai.APIOpenAIDecisions {
		return r.invokeHooked(ctx, w.client, scope, target, hooks)
	}
	res, err := w.client.WithHooks(hooks).Classify(ctx, scope, target, runtimeMixedInput(t, k, images), nil)
	return mixedOutcome{metadata: res.Metadata, usage: res.Usage, classifier: res, err: err}
}

func runtimeMixedReply(t *testing.T, r mixedRoute, k tenantKey) provider.Reply {
	t.Helper()
	reply := r.tenantReply(t, k)
	if r.api == ai.APIOpenAIDecisions {
		input, choice, other := 31, "track", "other"
		if k.tenant == tenantB.tenant {
			input, choice, other = 47, "other", "track"
		}
		body := `{"answers":[{"type":"choice","name":"intent","choice":"` + choice + `","probabilities":[{"value":"` + choice + `","probability":0.9},{"value":"` + other + `","probability":0.1}],"confidence":0.82}],"usage":{"input_tokens":` + strconv.Itoa(input) + `,"output_tokens":20,"total_tokens":` + strconv.Itoa(input+20) + `,"input_tokens_details":{"cached_tokens":0,"cache_write_tokens":0},"output_tokens_details":{"reasoning_tokens":0}}}`
		reply = provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json"}, Chunks: [][]byte{[]byte(body)}}
	}
	if reply.Header == nil {
		reply.Header = map[string]string{}
	}
	header := "x-request-id"
	if r.provider == ai.ProviderTypeSafe {
		header = "x-typesafe-request-id"
	}
	reply.Header[header] = "up-" + k.tenant + "-" + r.id
	reply.Script = joinChunks(reply.Chunks)
	return reply
}

func runtimeMixedEnqueue(ev *evidence.Case, w *mixedWorld, r mixedRoute, k tenantKey, replies ...provider.Reply) {
	w.provider.EnqueueAt(r.prefix(k), replies...)
	for i, reply := range replies {
		ev.Record("script-"+k.tenant+"-"+r.id+"-"+strconv.Itoa(len(w.provider.Requests()))+"-"+strconv.Itoa(i), map[string]any{"status": reply.Status, "headers": reply.Header, "body": reply.Script, "chunks": reply.Chunks, "end": reply.End, "endpointPrefix": r.prefix(k)})
	}
}

func recordRuntimeMixedConfig(ev *evidence.Case, w *mixedWorld) {
	ev.Record("catalog", w.client.Catalog())
	for _, k := range []tenantKey{tenantA, tenantB} {
		for _, r := range runtimeMixedRoutes {
			ev.Record("binding-"+k.tenant+"-"+r.id, w.host.Binding(k.tenant, r.id))
		}
	}
}

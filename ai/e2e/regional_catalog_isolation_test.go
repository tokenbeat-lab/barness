package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

// Mechanism-only metadata: these intentionally different Host restrictions,
// budgets, sampling values and prices never describe actual vendor products.
// Wire names stay unchanged so a sibling lookup would visibly change behavior.
func regionalMetadataCatalog() ai.Catalog {
	cat := ai.Catalog{Version: "regional-metadata-synthetic-v1"}
	for pi, p := range regionalRuntimeProviders {
		for ri, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			rate := float64((pi+1)*10 + ri + 1)
			model := ai.Model{Provider: p, API: api, ID: regionalMetadataModel(p), Name: "Synthetic independent Host metadata", Reasoning: true, Input: []ai.Modality{ai.ModalityText}, ContextWindow: 32768, MaxTokens: 170 + pi*10 + ri,
				ThinkingLevelMap: ai.ThinkingLevelMap{Off: ai.Value("none"), Minimal: ai.Null[string](), Low: ai.Null[string](), Medium: ai.Null[string](), High: ai.Value("high"), XHigh: ai.Null[string](), Max: ai.Null[string]()}, DefaultThinkingLevel: ai.ThinkingHigh,
				SamplingParams: map[string]json.RawMessage{"temperature": json.RawMessage("0.1")}, SamplingParamsByThinkingLevel: map[ai.ThinkingLevel]map[string]json.RawMessage{ai.ThinkingHigh: {"temperature": json.RawMessage(fmt.Sprintf("%.2f", regionalMetadataTemperature(pi, ri)))}},
				Compat: ai.ModelCompat{SupportsLongCacheRetention: ai.Value(false)}, Cost: ai.ModelCost{CostRates: ai.CostRates{Input: rate, Output: rate * 2, CacheRead: rate / 2, CacheWrite: rate}, Tiers: []ai.CostTier{{InputTokensAbove: 512000, CostRates: ai.CostRates{Input: rate * 2, Output: rate * 4, CacheRead: rate, CacheWrite: rate * 2}}}},
			}
			if p == ai.ProviderZAI {
				model.ThinkingLevelMap.High = ai.Value("max")
			}
			if p == ai.ProviderMiniMax || p == ai.ProviderMiniMaxCN {
				model.Compat.ThinkingSwitchOnly = true
				model.ThinkingLevelMap.Off = ai.Value("disabled")
			}
			if api == ai.APIAnthropicMessages && (p == ai.ProviderZhipu || p == ai.ProviderZAI) {
				// GLM Messages has no canonical effort/sampling mapping. Keeping these
				// disabled demonstrates that model-owned metadata cannot invent one.
				model.DefaultThinkingLevel = ""
				model.ThinkingLevelMap.High = ai.Null[string]()
				model.ThinkingLevelMap.Off = ai.Null[string]()
			}
			cat.Models = append(cat.Models, model)
		}
	}
	return cat
}
func regionalMetadataTemperature(pi, ri int) float64 { return float64(20+pi*10+ri) / 100 }
func regionalMetadataModel(p ai.ProviderID) string {
	if p == ai.ProviderZhipu || p == ai.ProviderZAI {
		return "glm-5.2"
	}
	return "MiniMax-M3"
}
func regionalMetadataWorld(t *testing.T, cat *ai.Catalog) *world {
	t.Helper()
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = cat }, tenantA)
	for _, p := range regionalRuntimeProviders {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			b := ai.Binding{TenantID: tenantA.tenant, BindingID: string(p) + "-" + string(api), Operation: ai.OperationChat, Version: "metadata-v1", Enabled: true, ProviderID: p, API: api, Endpoint: w.provider.URL() + "/metadata/" + string(p) + "/" + string(api), AuthKind: ai.AuthAPIKey, AccountScopeID: "same-account", CredentialRef: string(p) + "-record", AllowedModels: []string{regionalMetadataModel(p)}}
			w.host.PutBinding(b)
			w.host.PutCredential(ai.Credential{OwnerTenantID: tenantA.tenant, CredentialID: b.CredentialRef, AccountScopeID: b.AccountScopeID, Version: "key-v1", Active: true, APIKey: ai.NewSecret(tenantA.secret)})
		}
	}
	return w
}
func regionalMetadataBinding(w *world, p ai.ProviderID, api ai.API) ai.Binding {
	return w.host.Binding(tenantA.tenant, string(p)+"-"+string(api))
}
func regionalMetadataMutate(model *ai.Model) {
	model.Input[0] = "mutated"
	model.SamplingParams["temperature"][0] = '9'
	model.SamplingParamsByThinkingLevel[ai.ThinkingHigh]["temperature"][0] = '9'
	model.Cost.Tiers[0].Input = 999
}
func regionalMetadataReject(ev *evidence.Case, w *world, b ai.Binding, opts ai.Options, id string, code ai.Code) {
	n := len(w.provider.Requests())
	r, err := w.client.Complete(ctxFor(ev.T()), ai.CallScope{TenantID: tenantA.tenant, RequestID: id}, ai.Target{BindingID: b.BindingID, ModelID: regionalMetadataModel(b.ProviderID)}, ai.Request{}, opts)
	ev.Record(id+"-result", r)
	ev.Record(id+"-resolver-reads", w.host.ReadsFor(id))
	e, ok := err.(*ai.Error)
	ev.Check("invalid selection stops before credential/attempt/send", ok && e.Code == code && e.Phase == ai.PhaseCapability && w.host.ReadsFor(id).Credentials == 0 && len(r.Metadata.Attempts) == 0 && len(w.provider.Requests()) == n, "err=%v reads=%+v", err, w.host.ReadsFor(id))
}

func TestRegionalProviderCatalogIsolation(t *testing.T) {
	ev := run.Case(t, "RP02-RP03-RP04-catalog-one-client-independent-metadata")
	cat := regionalMetadataCatalog()
	hash, err := cat.Hash()
	if err != nil {
		t.Fatal(err)
	}
	ev.Record("synthetic-input-catalog", cat)
	w := regionalMetadataWorld(t, &cat)
	// All three public views must own nested slices, RawMessage bytes, profile
	// maps and price tiers. None can mutate another view or the active Client.
	view := w.client.Catalog()
	viewHash, _ := view.Hash()
	for _, p := range regionalRuntimeProviders {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			one, _ := view.Lookup(p, api, regionalMetadataModel(p))
			regionalMetadataMutate(&one)
			group := view.ModelsOf(p, api)
			regionalMetadataMutate(&group[0])
		}
	}
	after, _ := view.Hash()
	ev.Check("Lookup and ModelsOf own all nested values", after == viewHash, "before=%s after=%s", viewHash, after)
	for i := range cat.Models {
		regionalMetadataMutate(&cat.Models[i])
	}
	mutated := w.client.Catalog()
	for i := range mutated.Models {
		regionalMetadataMutate(&mutated.Models[i])
	}
	activeHash, _ := w.client.Catalog().Hash()
	ev.Check("caller and Catalog-view mutations leave active Client unchanged", activeHash == hash, "got=%s want=%s", activeHash, hash)
	for pi, p := range regionalRuntimeProviders {
		for ri, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			b := regionalMetadataBinding(w, p, api)
			model := regionalMetadataModel(p)
			id := fmt.Sprintf("metadata-%s-%s", p, api)
			reply, raw := regionalHostScript(t, b, model)
			ev.Fixture(id+"-script.json", raw)
			w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
			r, e := w.client.CompleteSimple(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: id}, ai.Target{BindingID: b.BindingID, ModelID: model}, ai.Request{Messages: []ai.Message{ai.UserText("Synthetic metadata input.")}}, ai.SimpleOptions{})
			ev.Record(id+"-result", r)
			if !ev.Check("same-named sibling uses actual full identity", e == nil && r.Metadata.ProviderID == p && r.Metadata.API == api && r.Metadata.ModelID == model && r.Metadata.CatalogHash == hash, "err=%v metadata=%+v", e, r.Metadata) {
				continue
			}
			captures := w.provider.Requests()
			fields := bodyFields(t, captures[len(captures)-1].Body)
			budget := "max_output_tokens"
			if api == ai.APIAnthropicMessages {
				budget = "max_tokens"
			}
			if api == ai.APIOpenAICompletions {
				budget = "max_completion_tokens"
				if p == ai.ProviderZhipu || p == ai.ProviderZAI {
					budget = "max_tokens"
				}
			}
			ev.Check("each full identity keeps own default output ceiling", string(fields[budget]) == fmt.Sprint(170+pi*10+ri), "body=%s", captures[len(captures)-1].Body)
			if api == ai.APIAnthropicMessages {
				ev.Check("Messages cannot inherit OpenAI sampling profiles", fields["temperature"] == nil && fields["reasoning_effort"] == nil, "body=%s", captures[len(captures)-1].Body)
			} else {
				ev.Check("model-specific default thinking profile owns sampling", string(fields["temperature"]) == fmt.Sprintf("%.2f", regionalMetadataTemperature(pi, ri)) || jsonEqual(fields["temperature"], mustMarshal(t, regionalMetadataTemperature(pi, ri))), "temperature=%s", fields["temperature"])
			}
			// 8 input and4 output at synthetic rates11..43: exactly16*rate /1M.
			want := 16 * float64((pi+1)*10+ri+1) / 1e6
			ev.Check("usage estimate stays with this Provider/API price", math.Abs(r.Message.Usage.Cost.Total-want) < 1e-12, "usage=%+v want=%g", r.Message.Usage, want)
			if api != ai.APIAnthropicMessages {
				var opts ai.Options = ai.ChatOptions{ReasoningEffort: ai.ThinkingLow}
				if api == ai.APIOpenAIResponses {
					opts = ai.ResponsesOptions{ReasoningEffort: ai.ThinkingLow}
				}
				regionalMetadataReject(ev, w, b, opts, id+"-unsupported", ai.CodeInvalidRequest)
				// Explicit high selects each model's independently restricted wire map.
				reply, _ = regionalHostScript(t, b, model)
				w.provider.EnqueueAt(strings.TrimPrefix(b.Endpoint, w.provider.URL()), reply)
				_, e = w.client.CompleteSimple(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: id + "-high"}, ai.Target{BindingID: b.BindingID, ModelID: model}, ai.Request{}, ai.SimpleOptions{Reasoning: ai.ThinkingHigh})
				captures = w.provider.Requests()
				fields = bodyFields(t, captures[len(captures)-1].Body)
				ev.Check("supported explicit effort request succeeds", e == nil, "%v", e)
				if p == ai.ProviderZhipu || p == ai.ProviderZAI {
					effort := fields["reasoning_effort"]
					if api == ai.APIOpenAIResponses {
						effort = bodyFields(t, fields["reasoning"])["effort"]
					}
					wantEffort := `"high"`
					if p == ai.ProviderZAI {
						wantEffort = `"max"`
					}
					ev.Check("sibling effort map is not inherited", string(effort) == wantEffort, "got=%s want=%s", effort, wantEffort)
				}
			} else if p == ai.ProviderZAI || p == ai.ProviderMiniMaxCN {
				regionalMetadataReject(ev, w, b, ai.AnthropicOptions{Effort: ai.AnthropicEffortHigh}, id+"-unsupported", ai.CodeInvalidRequest)
			}
		}
	}
	ev.Record("requests", w.provider.Requests())
	changed := w.client.Catalog()
	changed.Version = "regional-metadata-synthetic-v2"
	changed.Models[3].Cost.Input++
	nextHash, e := changed.Hash()
	ev.Check("content revision uses new version and hash", e == nil && nextHash != hash && changed.Version != w.client.Catalog().Version, "old=%s new=%s", hash, nextHash)
	ev.Record("next-catalog", changed)
	ev.Record("next-catalog-hash", nextHash)
}

func TestRegionalProviderCatalogMissingSibling(t *testing.T) {
	for _, p := range []ai.ProviderID{ai.ProviderZAI, ai.ProviderMiniMaxCN} {
		for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
			t.Run(string(p)+"/"+string(api), func(t *testing.T) {
				ev := run.Case(t, "RP02-RP03-missing-sibling-"+string(p)+"-"+string(api))
				cat := regionalMetadataCatalog()
				cat.Models = slices.DeleteFunc(cat.Models, func(m ai.Model) bool { return m.Provider == p && m.API == api })
				w := regionalMetadataWorld(t, &cat)
				b := regionalMetadataBinding(w, p, api)
				regionalMetadataReject(ev, w, b, nil, "missing-sibling", ai.CodeInvalidRequest)
				sibling := ai.ProviderZhipu
				if p == ai.ProviderMiniMaxCN {
					sibling = ai.ProviderMiniMax
				}
				_, exists := w.client.Catalog().Lookup(sibling, api, regionalMetadataModel(p))
				ev.Check("same wire model exists at sibling but grants no catalog fallback", exists, "sibling=%s", sibling)
			})
		}
	}
}

func TestRegionalProviderCatalogInvalidMetadata(t *testing.T) {
	for _, kind := range []string{"duplicate", "negative-price", "infinite-price", "nan-price", "invalid-modality", "invalid-sampling", "invalid-profile", "invalid-default", "negative-tier"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "RP04-invalid-regional-catalog-"+kind)
			cat := regionalMetadataCatalog()
			switch kind {
			case "duplicate":
				cat.Models = append(cat.Models, cat.Models[0])
			case "negative-price":
				cat.Models[3].Cost.Input = -1
			case "infinite-price":
				cat.Models[3].Cost.Output = math.Inf(1)
			case "nan-price":
				cat.Models[3].Cost.CacheRead = math.NaN()
			case "invalid-modality":
				cat.Models[3].Input = []ai.Modality{"video"}
			case "invalid-sampling":
				cat.Models[3].SamplingParams["temperature"] = json.RawMessage("{")
			case "invalid-profile":
				cat.Models[3].SamplingParamsByThinkingLevel["invented"] = map[string]json.RawMessage{}
			case "invalid-default":
				cat.Models[3].DefaultThinkingLevel = ai.ThinkingLow
			case "negative-tier":
				cat.Models[3].Cost.Tiers[0].InputTokensAbove = -1
			}
			_, hashErr := cat.Hash()
			w := newWorld(t, tenantA)
			_, buildErr := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Check("Hash and public Client reject same invalid regional metadata", hashErr != nil && buildErr != nil, "hash=%v client=%v", hashErr, buildErr)
		})
	}
}

// No retained Z.AI source/own wire qualification grants China's exact
// Messages budget1024/max4096 pair. Host metadata and hooks cannot grant it.
func TestRegionalProviderCatalogCannotGrantChinaBudget(t *testing.T) {
	for _, hooked := range []bool{false, true} {
		t.Run(fmt.Sprint(hooked), func(t *testing.T) {
			ev := run.Case(t, fmt.Sprintf("RP04-RP13-zai-cannot-grant-China-budget-hooks-%t", hooked))
			cat := regionalMetadataCatalog()
			for i := range cat.Models {
				m := &cat.Models[i]
				if m.Provider == ai.ProviderZAI && m.API == ai.APIAnthropicMessages {
					m.MaxTokens = 4096
					m.Compat.MessagesThinkingBudgetTokens = 1024
				}
			}
			w := regionalMetadataWorld(t, &cat)
			b := regionalMetadataBinding(w, ai.ProviderZAI, ai.APIAnthropicMessages)
			target := ai.Target{BindingID: b.BindingID, ModelID: "glm-5.2"}
			id := fmt.Sprintf("zai-budget-%t", hooked)
			opts := ai.AnthropicOptions{MaxTokens: ai.Value(4096), ThinkingEnabled: ai.Value(true), ThinkingBudgetTokens: 1024}
			var caller regionalCaller = w.client
			phase := ai.PhaseCapability
			if hooked {
				opts = ai.AnthropicOptions{MaxTokens: ai.Value(4096)}
				phase = ai.PhaseRequest
				caller = w.client.WithHooks(ai.Hooks{OnPayload: func(_ context.Context, _ ai.CallScope, payload *ai.Payload) (ai.PayloadDecision, error) {
					payload.Body["thinking"] = map[string]any{"type": "enabled", "budget_tokens": 1024}
					return ai.KeepPayload(), nil
				}})
			}
			r, err := caller.Complete(ctxFor(t), ai.CallScope{TenantID: tenantA.tenant, RequestID: id}, target, ai.Request{}, opts)
			ev.Record("result", r)
			ev.Record("resolver-reads", w.host.ReadsFor(id))
			ev.Record("requests", w.provider.Requests())
			failure, ok := err.(*ai.Error)
			ev.Check("ZAI has no independently qualified exact-budget subset", ok && failure.Code == ai.CodeInvalidRequest && failure.Phase == phase && len(r.Metadata.Attempts) == 0 && len(w.provider.Requests()) == 0, "err=%v attempts=%+v", err, r.Metadata.Attempts)
			if !hooked {
				ev.Check("unsupported full budget fails before credential", w.host.ReadsFor(id).Credentials == 0, "reads=%+v", w.host.ReadsFor(id))
			}
		})
	}
}

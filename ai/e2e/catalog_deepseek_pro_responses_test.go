package e2e

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// Failure inventory: missing Pro Responses, accidental change to Chat/Flash or
// another provider, wrong effort mapping/prices, stateful OpenAI fields leaking
// to DeepSeek, or a foreign protocol/service-tier request reaching HTTP.
func TestDeepSeekProResponsesBuiltin(t *testing.T) {
	ev := run.Case(t, "CATALOG-DEEPSEEK-PRO-RESPONSES-builtin")
	cat := ai.BuiltinCatalog()
	i := catalogContractIdentity{Operation: ai.OperationChat, Provider: ai.ProviderDeepSeek, API: ai.APIOpenAIResponses, ID: "deepseek-v4-pro"}
	m, ok := cat.Lookup(i.Provider, i.API, i.ID)
	if !ev.Check("priority Pro Responses identity is discoverable", ok, "identity=%+v", i) {
		return
	}
	ev.Check("official text model and limits", m.Name == "DeepSeek V4 Pro" && m.Reasoning && m.ContextWindow == 1000000 && m.MaxTokens == 384000 && slices.Equal(m.Input, []ai.Modality{ai.ModalityText}) && m.DefaultThinkingLevel == ai.ThinkingHigh, "model=%+v", m)
	ev.Check("fixed peak-rate estimate is explicit", m.Cost.CostRates == (ai.CostRates{Input: 1.32, Output: 3.96, CacheRead: .044, CacheWrite: 0}) && len(m.Cost.Tiers) == 0, "cost=%+v", m.Cost)
	for _, entry := range []struct {
		name, wire string
		value      ai.Nullable[string]
	}{{"off", "none", m.ThinkingLevelMap.Off}, {"minimal", "low", m.ThinkingLevelMap.Minimal}, {"low", "low", m.ThinkingLevelMap.Low}, {"medium", "high", m.ThinkingLevelMap.Medium}, {"high", "high", m.ThinkingLevelMap.High}, {"xhigh", "high", m.ThinkingLevelMap.XHigh}, {"max", "max", m.ThinkingLevelMap.Max}} {
		value, set := entry.value.Get()
		ev.Check("official Responses effort "+entry.name, set && value == entry.wire, "got=%s want=%s", value, entry.wire)
	}
	raw, err := os.ReadFile("testdata/catalog/pre-deepseek-pro-responses.json")
	if err != nil {
		t.Fatal(err)
	}
	var prior struct {
		Hash    string
		Catalog ai.Catalog
	}
	mustUnmarshal(t, raw, &prior)
	ev.Check("pre-add snapshot remains valid", catalogHash(t, prior.Catalog) == prior.Hash, "snapshot invalid")
	without := cat
	without.Version = prior.Catalog.Version
	without.Models = slices.DeleteFunc(slices.Clone(cat.Models), func(model ai.Model) bool {
		return model.Provider == i.Provider && model.API == i.API && model.ID == i.ID
	})
	without = catalogWithHistoricalNames(without, prior.Catalog)
	ev.Check("only the new protocol identity changes the catalog", len(cat.Models) == len(prior.Catalog.Models)+1 && reflect.DeepEqual(without, prior.Catalog), "prior definitions or order changed")
	chat, exists := cat.Lookup(i.Provider, ai.APIOpenAICompletions, i.ID)
	ev.Check("original Chat identity remains independent", exists && chat.ThinkingLevelMap.Low.IsNull() && !m.ThinkingLevelMap.Low.IsNull(), "chat=%+v responses=%+v", chat, m)
	for _, scene := range []string{"load-query", "basic", "cache-usage", "history-tools", "input-modalities", "parameters", "reasoning", "authorization"} {
		t.Run(scene, func(t *testing.T) {
			catalogChatContract(t, run.Case(t, "CATALOG-DEEPSEEK-PRO-RESPONSES-"+scene), cat, i, scene)
		})
	}
	t.Run("provider-boundary", func(t *testing.T) {
		e := run.Case(t, "CATALOG-DEEPSEEK-PRO-RESPONSES-provider-boundary")
		w, _ := newCatalogContractWorld(t, cat, i)
		enqueue(e, w, catalogChatReply(t, m, 8, false))
		req := ai.Request{Messages: []ai.Message{ai.UserText("Controlled Pro Responses input.")}}
		result, err := w.client.CompleteSimple(ctxFor(t), textScope("pro-responses-priority"), catalogContractTarget(i), req, ai.SimpleOptions{Reasoning: ai.ThinkingHigh, MaxTokens: ai.Value(64), CacheRetention: ai.CacheRetentionLong, SessionID: "host-session"})
		e.Check("recommended high-effort Responses call completes", err == nil, "%v", err)
		if err != nil {
			return
		}
		catalogContractMetadata(e, cat, i, result.Metadata)
		raw := catalogChatWire(e, w, i)
		var body map[string]json.RawMessage
		mustUnmarshal(t, raw, &body)
		for _, field := range []string{"store", "include", "previous_response_id", "conversation", "prompt_cache_key", "prompt_cache_retention", "service_tier"} {
			_, sent := body[field]
			e.Check("unsupported stateful/provider field omitted "+field, !sent, "%s", raw)
		}
		var reasoning struct{ Effort string }
		mustUnmarshal(t, body["reasoning"], &reasoning)
		e.Check("explicit priority effort is high", reasoning.Effort == "high", "%s", raw)
		before := len(w.provider.Requests())
		_, err = w.client.Complete(ctxFor(t), textScope("pro-responses-tier-denied"), catalogContractTarget(i), req, ai.ResponsesOptions{ServiceTier: ai.Value("priority")})
		e.Check("unsupported service tier refused before HTTP", err != nil && len(w.provider.Requests()) == before, "err=%v", err)
		_, err = w.client.Complete(ctxFor(t), textScope("pro-responses-protocol-denied"), catalogContractTarget(i), req, ai.ChatOptions{})
		e.Check("Chat options cannot select an implicit fallback", err != nil && len(w.provider.Requests()) == before, "err=%v", err)
		e.Record("requests", w.provider.Requests())
	})
}

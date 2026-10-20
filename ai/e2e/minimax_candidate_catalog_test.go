package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/minimax"
)

func TestMiniMaxCandidateCatalog(t *testing.T) {
	ev := run.Case(t, "MM01-six-candidate-identities")
	cat := minimax.CandidateCatalog()
	h := host.New()
	client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: h, Credentials: h, Catalog: &cat})
	if !ev.Check("candidate accepted at configuration boundary", err == nil, "%v", err) {
		return
	}
	hash := catalogHash(t, cat)
	ev.Record("candidate-catalog", client.Catalog())
	ev.Check("synthetic rates cannot look formal", strings.Contains(cat.Version, "candidate") && strings.Contains(cat.Version, "synthetic-prices"), "%s", cat.Version)
	ev.Check("six independent identities", len(cat.Models) == 6, "%d", len(cat.Models))
	ev.Check("regional provider identity", ai.ProviderMiniMax == ai.ProviderID("minimax"), "%s", ai.ProviderMiniMax)
	for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
		for _, id := range []string{"MiniMax-M3", "MiniMax-M3.1-Flash-Preview"} {
			m, ok := client.Catalog().Lookup(ai.ProviderMiniMax, api, id)
			if !ev.Check("own candidate identity "+string(api)+"/"+id, ok, "missing") {
				continue
			}
			ev.Check("text subset and conservative limits", m.Reasoning && len(m.Input) == 1 && m.Input[0] == ai.ModalityText && m.ContextWindow == 1000000 && m.MaxTokens == 131072, "%+v", m)
			fields := bodyFields(t, mustMarshal(t, m))
			wantDefault := `"max"`
			if id == "MiniMax-M3" {
				wantDefault = `"off"`
				if api == ai.APIOpenAICompletions {
					wantDefault = `"high"`
				}
				ev.Check("binary thinking exposes high only", !m.Compat.RequiresThinking && m.ThinkingLevelMap.Minimal.IsNull() && m.ThinkingLevelMap.Low.IsNull() && m.ThinkingLevelMap.Medium.IsNull() && m.ThinkingLevelMap.XHigh.IsNull() && m.ThinkingLevelMap.Max.IsNull(), "%+v", m)
			} else {
				ev.Check("mandatory thinking excludes off and minimal", m.Compat.RequiresThinking && m.ThinkingLevelMap.Off.IsNull() && m.ThinkingLevelMap.Minimal.IsNull(), "%+v", m)
				for _, pair := range []struct {
					key   string
					entry ai.Nullable[string]
				}{{"low", m.ThinkingLevelMap.Low}, {"medium", m.ThinkingLevelMap.Medium}, {"high", m.ThinkingLevelMap.High}, {"xhigh", m.ThinkingLevelMap.XHigh}, {"max", m.ThinkingLevelMap.Max}} {
					value, supported := pair.entry.Get()
					ev.Check("five independent depths "+pair.key, supported && value == pair.key, "%s", value)
				}
			}
			ev.Check("per-protocol omitted label", string(fields["defaultThinkingLevel"]) == wantDefault, "%s", mustMarshal(t, m))
			_, formal := ai.BuiltinCatalog().Lookup(ai.ProviderMiniMax, api, id)
			if id == "MiniMax-M3.1-Flash-Preview" {
				ev.Check("unpriced M3.1 stays outside formal catalog", !formal, "%s/%s", api, id)
			} else if formal {
				sourced := minimax.PricedM3Catalog()
				want, _ := sourced.Lookup(ai.ProviderMiniMax, api, id)
				got, _ := ai.BuiltinCatalog().Lookup(ai.ProviderMiniMax, api, id)
				ev.Check("promoted M3 keeps own sourced constraints", got.MaxTokens == want.MaxTokens && jsonEqual(mustMarshal(t, got.Cost), mustMarshal(t, want.Cost)), "%+v", got)
			}
		}
	}
	cat.Models[0].Input[0] = ai.ModalityImage
	cat.Models[0].SamplingParams = map[string]json.RawMessage{"temperature": json.RawMessage(`2`)}
	ev.Check("fresh assembly and Client own candidate data", catalogHash(t, minimax.CandidateCatalog()) == hash && catalogHash(t, client.Catalog()) == hash, "mutation escaped")
}

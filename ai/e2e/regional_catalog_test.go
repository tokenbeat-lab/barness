package e2e

import (
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/regional"
)

func TestRegionalProviderCatalog(t *testing.T) {
	ev := run.Case(t, "regional-catalog-independent-candidates")
	cat, err := regional.CandidateCatalog()
	if !ev.Check("all nine source-backed targets usable for qualification", err == nil && len(cat.Models) == 9, "err=%v count=%d", err, len(cat.Models)) {
		return
	}
	hash, err := cat.Hash()
	ev.Record("candidate-catalog", cat)
	ev.Record("candidate-catalog-hash", hash)
	ev.Check("valid complete identity catalog", err == nil && hash != "", "%v", err)
	ev.Check("official source hashes match retained bytes", regional.VerifySources() == nil, "source verification failed")
	promotedCount := 0
	for _, api := range []ai.API{ai.APIOpenAICompletions, ai.APIOpenAIResponses, ai.APIAnthropicMessages} {
		for _, id := range []string{"glm-5.3", "glm-5.2"} {
			model, ok := cat.Lookup("zai", api, id)
			ev.Check("production GLM display name is stable", ok && model.Name == "Z.AI GLM-"+id[4:], "name=%s", model.Name)
			ev.Check("international GLM own direct USD rates", ok && model.Cost.Input == 1.4 && model.Cost.Output == 4.4 && model.Cost.CacheRead == 0.26 && model.Cost.CacheWrite == 0, "id=%s api=%s", id, api)
			formal, promoted := ai.BuiltinCatalog().Lookup("zai", api, id)
			if promoted {
				promotedCount++
			}
			ev.Check("promoted identity retains independent own facts", !promoted || reflect.DeepEqual(formal, model), "id=%s api=%s", id, api)
		}
		model, ok := cat.Lookup("minimax-cn", api, "MiniMax-M3")
		ev.Check("production China display name is stable", ok && model.Name == "MiniMax-M3 中国站", "name=%s", model.Name)
		ev.Check("China MiniMax uses fixed own CNY-to-USD estimate", ok && model.Cost.Input == 0.3124070217 && model.Cost.Output == 1.2496280869 && model.Cost.CacheRead == 0.0624814043 && len(model.Cost.Tiers) == 1 && model.Cost.Tiers[0].InputTokensAbove == 512000, "model=%+v", model)
		formal, promoted := ai.BuiltinCatalog().Lookup("minimax-cn", api, "MiniMax-M3")
		if promoted {
			promotedCount++
		}
		ev.Check("China MiniMax promotion retains independent own facts", !promoted || reflect.DeepEqual(formal, model), "promoted=%t", promoted)
	}
	ev.Check("formal activation is all nine identities or none", promotedCount == 0 || promotedCount == 9, "promoted=%d", promotedCount)
	candidates := regional.Candidates()
	ev.Record("candidate-status", candidates)
	ev.Check("nine own source/status/price records", len(candidates) == 9, "count=%d", len(candidates))
	for _, candidate := range candidates {
		ev.Check("qualification and missing token-write rate are explicit", candidate.Status == "INCOMPLETE" && candidate.OwnLive == "NOT_RUN" && len(candidate.Sources) > 0 && candidate.Price.CacheWrite == nil && candidate.Price.Unit == "per-million-tokens", "%+v", candidate)
	}
	// Public Catalog lookup/ModelsOf and Client construction own their snapshots.
	models := cat.ModelsOf("zai", ai.APIOpenAICompletions)
	models[0].Input[0] = "mutated"
	fresh, _ := cat.Lookup("zai", ai.APIOpenAICompletions, "glm-5.3")
	ev.Check("lookup result does not share nested slices", fresh.Input[0] == ai.ModalityText, "%v", fresh.Input)
	bad := cat
	bad.Models = append(append([]ai.Model{}, cat.Models...), cat.Models[0])
	w := newWorld(t, tenantA)
	_, err = ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &bad})
	ev.Check("duplicate full identity fails at public construction", err != nil, "err=%v", err)
}

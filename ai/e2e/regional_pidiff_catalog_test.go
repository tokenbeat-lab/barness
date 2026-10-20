package e2e

import (
	"runtime"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

// The frozen data already contains these complete identities. An incomplete
// wrapper inventory must never reclassify them as route extensions.
func TestRegionalProviderFrozenCatalog(t *testing.T) {
	ev := run.Case(t, "RP06-frozen-regional-inventory")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	o := openOracle(t)
	keys := []pioracle.ModelKey{{Provider: "zai", API: "openai-completions", ID: "glm-5.3"}, {Provider: "zai", API: "openai-completions", ID: "glm-5.2"}, {Provider: "minimax-cn", API: "anthropic-messages", ID: "MiniMax-M3"}}
	models, err := o.Models(ctxFor(t), keys)
	ev.Record("frozen-models", models)
	if !ev.Check("pinned lookup executes", err == nil && len(models) == 3, "%v", err) {
		return
	}
	for i, m := range models {
		ev.Check("actual frozen full identity available", string(m) != "null", "%+v missing", keys[i])
		if string(m) != "null" {
			v := jsonObject(t, m)
			ev.Check("Provider/model/API remain exact", v["provider"] == keys[i].Provider && v["api"] == keys[i].API && v["id"] == keys[i].ID, "%s", m)
		}
	}
}

func TestRegionalProviderFrozenCatalogDifferences(t *testing.T) {
	for _, r := range regionalFrozenRoutes {
		for _, id := range r.models {
			t.Run(r.name+"/"+id, func(t *testing.T) {
				ev := run.Case(t, "PIDIFF-RP06-"+r.name+"-"+id+"-catalog")
				ev.ReplayEnv(pioracle.EnableEnv + "=1")
				o := openOracle(t)
				models, err := o.Models(ctxFor(t), []pioracle.ModelKey{{Provider: string(r.provider), API: string(r.api), ID: id}})
				if !ev.Check("complete frozen model exists", err == nil && len(models) == 1 && string(models[0]) != "null", "%v", err) {
					return
				}
				w, _ := regionalFrozenWorld(t, r)
				m, ok := w.client.Catalog().Lookup(r.provider, r.api, id)
				if !ok {
					t.Fatal("missing own model")
				}
				own := jsonObject(t, m)
				own["type"] = "chat"
				base := "https://api.z.ai/api/paas/v4"
				frozenBase := "https://api.z.ai/api/coding/paas/v4"
				if r.provider == ai.ProviderMiniMaxCN {
					base, frozenBase = "https://api.minimax.cn/anthropic", "https://api.minimaxi.com/anthropic"
				}
				own["baseUrl"] = base
				frozen := jsonObject(t, models[0])
				ev.Check("own and frozen full identities remain exact", frozen["provider"] == string(r.provider) && frozen["api"] == string(r.api) && frozen["id"] == id && frozen["baseUrl"] == frozenBase && m.Provider == r.provider && m.API == r.api && m.ID == id, "%s", models[0])
				ev.Check("own tested text scope and context are explicit", jsonEqual(mustMarshal(t, m.Input), []byte(`["text"]`)) && m.ContextWindow == 1000000, "%s", mustMarshal(t, m))
				if r.provider == ai.ProviderZAI {
					ev.Check("own ZAI budget and source-backed USD prices", m.MaxTokens == 131072 && jsonEqual(mustMarshal(t, m.Cost), []byte(`{"input":1.4,"output":4.4,"cacheRead":0.26,"cacheWrite":0}`)) && m.DefaultThinkingLevel == "max" && m.Name == "Z.AI "+id+" openai-completions own candidate", "%s", mustMarshal(t, m))
					compat := []byte(`{"supportsLongCacheRetention":false}`)
					levels := []byte(`{"off":"none","minimal":"minimal","low":"high","medium":"high","high":"high","xhigh":"max","max":"max"}`)
					if id == "glm-5.3" {
						compat = []byte(`{"supportsLongCacheRetention":false,"requiresThinking":true}`)
						levels = []byte(`{"off":null,"minimal":null,"low":"low","medium":null,"high":"high","xhigh":null,"max":"max"}`)
					}
					ev.Check("own complete hard constraints stay bounded", jsonEqual(mustMarshal(t, m.Compat), compat) && jsonEqual(mustMarshal(t, m.ThinkingLevelMap), levels), "%s", mustMarshal(t, m))
					ev.Check("frozen ZAI compat remains unmodified", jsonEqual(mustMarshal(t, frozen["compat"]), []byte(`{"supportsStore":false,"supportsDeveloperRole":false,"supportsReasoningEffort":true,"maxTokensField":"max_tokens","thinkingFormat":"zai","supportsStrictMode":true,"zaiToolStream":true}`)), "%s", models[0])
				} else {
					ev.Check("own China ceiling and binary thinking stay explicit", m.MaxTokens == 2048 && m.Name == "MiniMax China M3 anthropic-messages own candidate" && m.DefaultThinkingLevel == "off" && jsonEqual(mustMarshal(t, m.Compat), []byte(`{"supportsLongCacheRetention":false,"thinkingSwitchOnly":true}`)) && jsonEqual(mustMarshal(t, m.ThinkingLevelMap), []byte(`{"off":"disabled","minimal":null,"low":null,"medium":null,"high":"high","xhigh":null,"max":null}`)), "%s", mustMarshal(t, m))
					ev.Check("complete China converted prices and tiers", jsonEqual(mustMarshal(t, m.Cost), []byte(`{"input":0.3124070217,"output":1.2496280869,"cacheRead":0.0624814043,"cacheWrite":0,"tiers":[{"inputTokensAbove":512000,"input":0.6248140434,"output":2.4992561738,"cacheRead":0.1249628087,"cacheWrite":0}]}`)), "%s", mustMarshal(t, m))
					ev.Check("frozen CN broad metadata is preserved", frozen["maxTokens"] == float64(512000) && jsonEqual(mustMarshal(t, frozen["input"]), []byte(`["text","image"]`)) && jsonEqual(mustMarshal(t, frozen["inputLimits"]), []byte(`{"images":{"resize":{"maxWidth":2000,"maxHeight":2000,"maxBytes":4718592,"jpegQuality":80}}}`)), "%s", models[0])
				}
				pi, b := pioracle.Observation{Result: models[0]}, pioracle.Observation{Result: mustMarshal(t, own)}
				ev.Record("pi-observation", pi)
				ev.Record("barness-observation", b)
				diffs, e := pioracle.Compare(pi, b)
				if !ev.Check("complete metadata comparison", e == nil, "%v", e) {
					return
				}
				v := loadLedger(t).Classify(string(r.api), ev.ID(), diffs)
				ev.Record("pidiff", o.Record(ev.ID(), string(r.api), v, pioracle.Sides{BarnessVersions: map[string]string{"go": runtime.Version(), "openai-go": evidence.ModuleVersion("github.com/openai/openai-go/v3"), "anthropic-sdk-go": evidence.ModuleVersion("github.com/anthropics/anthropic-sdk-go")}, ModelCatalogHash: catalogHash(t, w.client.Catalog()), PiRequest: mustMarshal(t, pi.Request), BarnessRequest: mustMarshal(t, b.Request), PiResult: pi.Result, BarnessResult: b.Result, Frames: mustMarshal(t, models)}))
				ev.Check("all exact metadata differences are reviewed", v.Pass, "%s", mustMarshal(t, v))
			})
		}
	}
}

package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai/internal/testkit/pioracle"
)

func TestMiniMaxPiCatalog(t *testing.T) {
	ev := run.Case(t, "PIDIFF-MM06-frozen-catalog")
	ev.ReplayEnv(pioracle.EnableEnv + "=1")
	o := openOracle(t)
	models, err := o.Models(ctxFor(t), []pioracle.ModelKey{
		{Provider: "minimax", API: "anthropic-messages", ID: "MiniMax-M3"},
		{Provider: "minimax", API: "anthropic-messages", ID: "MiniMax-M3.1-Flash-Preview"},
		{Provider: "minimax", API: "openai-completions", ID: "MiniMax-M3"},
		{Provider: "minimax", API: "openai-responses", ID: "MiniMax-M3"},
	})
	if !ev.Check("frozen lookup completed", err == nil && len(models) == 4, "%v", err) {
		return
	}
	ev.Record("frozen-models", models)
	if !ev.Check("M3 Messages uses existing pinned model", string(models[0]) != "null", "%s", models[0]) {
		return
	}
	m := jsonObject(t, models[0])
	ev.Check("pinned endpoint and model facts stay honest", m["baseUrl"] == "https://api.minimax.io/anthropic" && m["contextWindow"] == float64(1000000) && m["maxTokens"] == float64(512000), "%s", models[0])
	ev.Check("frozen complete historical rates are retained", jsonEqual(mustMarshal(t, m["cost"]), []byte(`{"input":0.3,"output":1.2,"cacheRead":0.06,"cacheWrite":0,"tiers":[{"inputTokensAbove":512000,"input":0.6,"output":2.4,"cacheRead":0.12,"cacheWrite":0}]}`)), "%v", m["cost"])
	ev.Check("frozen image policy is recorded, not copied as support", jsonEqual(mustMarshal(t, m["input"]), []byte(`["text","image"]`)) && jsonEqual(mustMarshal(t, m["inputLimits"]), []byte(`{"images":{"resize":{"maxWidth":2000,"maxHeight":2000,"maxBytes":4718592,"jpegQuality":80}}}`)), "%s", models[0])
	ev.Check("new model and routes are not fabricated", string(models[1]) == "null" && string(models[2]) == "null" && string(models[3]) == "null", "%s", mustMarshal(t, models))
}

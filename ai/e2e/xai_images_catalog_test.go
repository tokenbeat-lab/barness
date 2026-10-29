package e2e

import (
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

// The spec's dated price snapshot is an independent public-JSON contract.
// This helper uses JSON so the red scenario exercises the configuration
// boundary even before Go adds the new price branch.
func xaiImagesCatalogJSON() []byte {
	return []byte(`{"version":"xai-images-candidate-2026-10-29","models":null,"imageModels":[{
 "provider":"xai","api":"xai-images","id":"grok-imagine-image-2.0","name":"Grok Imagine Image 2.0 candidate",
 "input":["text","image"],"output":["image"],
 "capabilities":{"maxReferenceImages":5,"maxOutputImages":10,"imageSizes":["1k","1.5k","2k"],"qualities":["low","medium","auto"],"aspectRatios":["1:1","3:4","4:3","9:16","16:9","2:3","3:2","9:19.5","19.5:9","9:20","20:9","1:2","2:1","21:9","5:2","auto"]},
 "pricing":{"perImage":{"inputImage":0.01,"output":[
 {"quality":"low","imageSize":"1k","price":0.04},
 {"quality":"low","imageSize":"1.5k","price":0.05},
 {"quality":"low","imageSize":"2k","price":0.06},
 {"quality":"medium","imageSize":"1k","price":0.06},
 {"quality":"medium","imageSize":"1.5k","price":0.07},
 {"quality":"medium","imageSize":"2k","price":0.08}]},"source":"https://docs.x.ai/developers/pricing; retrieved 2026-10-29"}}]}`)
}

func TestXAIImagesCatalog(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-catalog-candidate")
	var cat ai.Catalog
	mustUnmarshal(t, xaiImagesCatalogJSON(), &cat)
	w := newWorld(t, tenantA)
	client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
	ev.Record("catalog", cat)
	ev.Check("candidate is accepted without credentials", err == nil, "NewClient=%v", err)
	if err != nil {
		return
	}
	got := client.Catalog()
	ev.Check("per-image public JSON retains the complete dated price snapshot", jsonEqual(mustMarshal(t, got), xaiImagesCatalogJSON()), "got=%s", mustMarshal(t, got))
	_, found := ai.BuiltinCatalog().LookupImage(ai.ProviderXAI, ai.API("xai-images"), "grok-imagine-image-2.0")
	ev.Check("candidate availability does not assert formal support", !found, "candidate entered BuiltinCatalog")
	ev.Check("discovery never resolves credentials", w.host.ReadsFor("catalog-discovery").Credentials == 0, "unexpected credential access")
}

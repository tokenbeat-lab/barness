package e2e

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/xai"
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
	reads := 0
	credentials := credentialResolverFunc(func(context.Context, ai.CallScope, ai.Binding) (ai.Credential, error) {
		reads++
		return ai.Credential{}, ai.ErrAccessDenied
	})
	client, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: credentials, Catalog: &cat})
	ev.Record("catalog", cat)
	ev.Check("candidate is accepted without credentials", err == nil, "NewClient=%v", err)
	if err != nil {
		return
	}
	got := client.Catalog()
	ev.Check("per-image public JSON retains the complete dated price snapshot", jsonEqual(mustMarshal(t, got), xaiImagesCatalogJSON()), "got=%s", mustMarshal(t, got))
	_, found := ai.BuiltinCatalog().LookupImage(ai.ProviderXAI, ai.API("xai-images"), "grok-imagine-image-2.0")
	ev.Check("candidate availability does not assert formal support", !found, "candidate entered BuiltinCatalog")
	ev.Check("discovery never resolves credentials", reads == 0, "unexpected credential access")
}

func TestXAIImagesCatalogValidation(t *testing.T) {
	cases := map[string]func(*ai.ImageModel){
		"missing-per-image": func(m *ai.ImageModel) {
			m.Pricing.PerImage = nil
			m.Pricing.InputText = ai.Value(1.0)
			m.Pricing.InputImage = ai.Value(1.0)
			m.Pricing.OutputImage = ai.Value(1.0)
		},
		"token-text":         func(m *ai.ImageModel) { m.Pricing.InputText = ai.Value(0.0) },
		"token-image":        func(m *ai.ImageModel) { m.Pricing.InputImage = ai.Value(0.0) },
		"token-output-text":  func(m *ai.ImageModel) { m.Pricing.OutputText = ai.Null[float64]() },
		"token-output-image": func(m *ai.ImageModel) { m.Pricing.OutputImage = ai.Value(0.0) },
		"token-cache-text":   func(m *ai.ImageModel) { m.Pricing.CacheReadText = ai.Value(0.0) },
		"token-cache-image":  func(m *ai.ImageModel) { m.Pricing.CacheReadImage = ai.Null[float64]() },
		"input-missing":      func(m *ai.ImageModel) { m.Pricing.PerImage.InputImage = ai.Nullable[float64]{} },
		"input-null":         func(m *ai.ImageModel) { m.Pricing.PerImage.InputImage = ai.Null[float64]() },
		"input-negative":     func(m *ai.ImageModel) { m.Pricing.PerImage.InputImage = ai.Value(-0.01) },
		"input-infinity":     func(m *ai.ImageModel) { m.Pricing.PerImage.InputImage = ai.Value(math.Inf(1)) },
		"output-missing":     func(m *ai.ImageModel) { m.Pricing.PerImage.Output = nil },
		"output-hole":        func(m *ai.ImageModel) { m.Pricing.PerImage.Output = m.Pricing.PerImage.Output[1:] },
		"output-duplicate": func(m *ai.ImageModel) {
			m.Pricing.PerImage.Output = append(m.Pricing.PerImage.Output, m.Pricing.PerImage.Output[0])
		},
		"output-auto":          func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Quality = "auto" },
		"output-quality":       func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Quality = "high" },
		"output-size":          func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].ImageSize = "4k" },
		"output-null":          func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Price = ai.Null[float64]() },
		"output-unset":         func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Price = ai.Nullable[float64]{} },
		"output-negative":      func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Price = ai.Value(-0.04) },
		"output-nan":           func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Price = ai.Value(math.NaN()) },
		"output-infinity":      func(m *ai.ImageModel) { m.Pricing.PerImage.Output[0].Price = ai.Value(math.Inf(1)) },
		"source-missing":       func(m *ai.ImageModel) { m.Pricing.Source = " " },
		"reference-cap":        func(m *ai.ImageModel) { m.Capabilities.MaxReferenceImages = 6 },
		"output-cap":           func(m *ai.ImageModel) { m.Capabilities.MaxOutputImages = 11 },
		"unknown-resolution":   func(m *ai.ImageModel) { m.Capabilities.ImageSizes = append(m.Capabilities.ImageSizes, "4k") },
		"duplicate-resolution": func(m *ai.ImageModel) { m.Capabilities.ImageSizes = append(m.Capabilities.ImageSizes, "1k") },
		"unknown-quality":      func(m *ai.ImageModel) { m.Capabilities.Qualities = append(m.Capabilities.Qualities, "high") },
		"duplicate-quality":    func(m *ai.ImageModel) { m.Capabilities.Qualities = append(m.Capabilities.Qualities, "low") },
		"unknown-ratio":        func(m *ai.ImageModel) { m.Capabilities.AspectRatios = append(m.Capabilities.AspectRatios, "17:7") },
		"duplicate-ratio":      func(m *ai.ImageModel) { m.Capabilities.AspectRatios = append(m.Capabilities.AspectRatios, "1:1") },
		"numeric-size":         func(m *ai.ImageModel) { m.Capabilities.Sizes = []string{"1024x1024"} },
		"mask":                 func(m *ai.ImageModel) { m.Capabilities.Mask = true },
		"transparent":          func(m *ai.ImageModel) { m.Capabilities.TransparentBackground = true },
		"fidelity":             func(m *ai.ImageModel) { m.Capabilities.InputFidelity = []string{"high"} },
		"no-text":              func(m *ai.ImageModel) { m.Input = []ai.Modality{ai.ModalityImage} },
		"text-output":          func(m *ai.ImageModel) { m.Output = []ai.Modality{ai.ModalityText, ai.ModalityImage} },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-catalog-reject-"+name)
			var cat ai.Catalog
			mustUnmarshal(t, xaiImagesCatalogJSON(), &cat)
			change(&cat.ImageModels[0])
			_, hashErr := cat.Hash()
			w := newWorld(t, tenantA)
			_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
			ev.Check("both catalog and client reject invalid image metadata", hashErr != nil && err != nil, "Hash=%v NewClient=%v", hashErr, err)
		})
	}
}

func TestXAIImagesCatalogSnapshots(t *testing.T) {
	for _, kind := range []string{"construction", "catalog", "lookup", "list"} {
		t.Run(kind, func(t *testing.T) {
			ev := run.Case(t, "XAI-IMG-catalog-snapshot-"+kind)
			var source ai.Catalog
			mustUnmarshal(t, xaiImagesCatalogJSON(), &source)
			w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &source }, tenantA)
			before := catalogHash(t, w.client.Catalog())
			owner := w.client.Catalog()
			var model ai.ImageModel
			switch kind {
			case "construction":
				model = source.ImageModels[0]
			case "catalog":
				model = owner.ImageModels[0]
			case "lookup":
				model, _ = owner.LookupImage(ai.ProviderXAI, ai.APIXAIImages, "grok-imagine-image-2.0")
			case "list":
				model = owner.ImageModelsOf(ai.ProviderXAI, ai.APIXAIImages)[0]
			}
			model.Pricing.PerImage.InputImage = ai.Value(9.0)
			model.Pricing.PerImage.Output[0].Price = ai.Value(99.0)
			ev.Check("price pointer and rows are independently owned", jsonEqual(mustMarshal(t, w.client.Catalog()), xaiImagesCatalogJSON()), "catalog changed through %s", kind)
			if kind == "lookup" || kind == "list" {
				ev.Check("typed queries preserve their source catalog", jsonEqual(mustMarshal(t, owner), xaiImagesCatalogJSON()), "query mutated its source catalog")
			}
			after := catalogHash(t, w.client.Catalog())
			ev.Record("snapshot-hashes", []string{before, after})
			ev.Check("caller mutation leaves the client price hash stable", before == after, "hash changed")
		})
	}
}

func TestXAIImagesPriceHashAndZero(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-catalog-zero-and-hash")
	var cat ai.Catalog
	mustUnmarshal(t, xaiImagesCatalogJSON(), &cat)
	before := catalogHash(t, cat)
	cat.ImageModels[0].Pricing.PerImage.InputImage = ai.Value(0.0)
	cat.ImageModels[0].Pricing.PerImage.Output[0].Price = ai.Value(0.0)
	after := catalogHash(t, cat)
	ev.Check("explicit zero is valid and changes the price content hash", before != after, "hash unchanged")
	ev.Record("zero-price-catalog", cat)
	w := newWorld(t, tenantA)
	_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &cat})
	ev.Check("NewClient preserves explicit free rates", err == nil, "NewClient=%v", err)
	legacy := ai.BuiltinCatalog().ImageModels[0]
	raw := mustMarshal(t, legacy.Pricing)
	var fields map[string]json.RawMessage
	mustUnmarshal(t, raw, &fields)
	_, present := fields["perImage"]
	ev.Check("token catalogs retain their old encoding", !present, "unexpected perImage branch")
}

func TestXAIImagesCandidateCatalog(t *testing.T) {
	ev := run.Case(t, "XAI-IMG-catalog-candidate-source")
	cat := xai.ImagesCandidateCatalog()
	ev.Check("candidate snapshot matches the dated spec", jsonEqual(mustMarshal(t, cat), xaiImagesCatalogJSON()), "candidate=%s", mustMarshal(t, cat))
	cat.ImageModels[0].Pricing.PerImage.Output[0].Price = ai.Value(99.0)
	ev.Check("candidate calls return independent price tables", jsonEqual(mustMarshal(t, xai.ImagesCandidateCatalog()), xaiImagesCatalogJSON()), "candidate source changed")
	w := newWorld(t, tenantA)
	source := xai.ImagesCandidateCatalog()
	_, err := ai.NewClient(ai.Config{Policy: validPolicy(), Bindings: w.host, Credentials: w.host, Catalog: &source})
	ev.Check("candidate needs no credential for catalog construction", err == nil, "NewClient=%v", err)
}

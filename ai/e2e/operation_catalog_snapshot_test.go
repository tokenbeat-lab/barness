package e2e

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

func TestOperationCatalogDiscovery(t *testing.T) {
	ev := run.Case(t, "E03-operation-catalog-discovery-and-snapshot")
	source := operationCatalog()
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &source }, tenantA)
	want := operationCatalog()
	hash := catalogHash(t, want)
	ev.Record("catalog-at-construction", want)

	mutateOperationCatalog(&source)
	got := w.client.Catalog()
	ev.Check("client owns the complete construction snapshot", reflect.DeepEqual(got, want), "got %+v", got)
	chat, chatOK := got.Lookup(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
	image, imageOK := got.LookupImage(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
	classifier, classifierOK := got.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIResponses, "gpt-4.1-mini")
	ev.Check("same ID resolves separately in all three operations", chatOK && imageOK && classifierOK &&
		chat.Name == "Host chat" && image.Name == "Host image" && classifier.Name == "Host classifier", "found=%v/%v/%v", chatOK, imageOK, classifierOK)
	// A named string kind must retain the public JSON and catalog hash. The
	// expected encoding is the existing contract, independent of constants.
	wantCaps := []byte(`{"kinds":["choice","score","bool"],"maxChoices":255,"minScoreLevels":2,"maxScoreLevels":10}`)
	ev.Check("typed question kinds preserve the capability JSON", jsonEqual(mustMarshal(t, classifier.Capabilities), wantCaps), "got %+v", classifier.Capabilities)
	var decoded ai.Catalog
	mustUnmarshal(t, mustMarshal(t, want), &decoded)
	ev.Check("catalog JSON round trip preserves typed question kinds and hash", reflect.DeepEqual(decoded, want) && catalogHash(t, decoded) == hash, "decoded=%+v", decoded)
	queries := ai.Catalog{Version: got.Version,
		Models:           got.ModelsOf(ai.ProviderOpenAI, ai.APIOpenAIResponses),
		ImageModels:      got.ImageModelsOf(ai.ProviderOpenAI, ai.APIOpenAIResponses),
		ClassifierModels: got.ClassifierModelsOf(ai.ProviderOpenAI, ai.APIOpenAIResponses)}
	ev.Check("typed lists contain the matching models", reflect.DeepEqual(queries, want), "got %+v", queries)
	mutateOperationCatalog(&queries)
	mutateOperationCatalog(&ai.Catalog{Models: []ai.Model{chat}, ImageModels: []ai.ImageModel{image}, ClassifierModels: []ai.ClassifierModel{classifier}})
	ev.Check("lookup and list results own their nested values", reflect.DeepEqual(got, want), "got %+v", got)
	mutateOperationCatalog(&got)
	ev.Check("modifying the returned client catalog cannot alter the client", reflect.DeepEqual(w.client.Catalog(), want), "snapshot changed")

	for _, tc := range []struct {
		provider ai.ProviderID
		api      ai.API
		id       string
	}{
		{ai.ProviderGoogle, ai.APIOpenAIResponses, "gpt-4.1-mini"},
		{ai.ProviderOpenAI, ai.APIOpenAICompletions, "gpt-4.1-mini"},
		{ai.ProviderOpenAI, ai.APIOpenAIResponses, "missing"},
	} {
		c := w.client.Catalog()
		_, a := c.Lookup(tc.provider, tc.api, tc.id)
		_, b := c.LookupImage(tc.provider, tc.api, tc.id)
		_, d := c.LookupClassifier(tc.provider, tc.api, tc.id)
		ev.Check("typed lookup keeps provider, API and ID boundaries", !a && !b && !d, "query=%+v", tc)
		if tc.id == "missing" {
			continue
		}
		ev.Check("typed lists filter provider and API", len(c.ModelsOf(tc.provider, tc.api)) == 0 && len(c.ImageModelsOf(tc.provider, tc.api)) == 0 && len(c.ClassifierModelsOf(tc.provider, tc.api)) == 0, "query=%+v", tc)
	}

	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("text-basic.json", raw)
	enqueue(ev, w, sseReply(t, f, provider.FramingLF))
	res, err := w.client.CompleteSimple(ctxFor(t), textScope("req-operation-catalog-snapshot"), textTarget(f), textRequest(f), ai.SimpleOptions{})
	ev.Record("result", res)
	ev.Check("chat selects the chat entry despite identical image/classifier IDs", err == nil, "err=%v", err)
	checkFinalMessage(ev, f, res.Message)
	checkPricingSnapshot(ev, res.Metadata, want.Version, hash)
	f.Expect.Request.Body = json.RawMessage(`{"model":"gpt-4.1-mini","stream":true,"store":false,"max_output_tokens":32768,"top_p":0.5,"input":[{"role":"system","content":"You are terse."},{"role":"user","content":[{"type":"input_text","text":"Say hello."}]}]}`)
	checkTextRequest(ev, w, f, tenantA, true)
}

func mutateOperationCatalog(c *ai.Catalog) {
	c.Version = "mutated"
	chat := &c.Models[0]
	chat.Input[0], chat.ID = ai.ModalityImage, "mutated-chat"
	chat.SamplingParams["top_p"][2] = '9'
	chat.SamplingParams["temperature"] = json.RawMessage(`0.9`)
	chat.Cost.Tiers[0].Input = 99
	image := &c.ImageModels[0]
	image.Input[0], image.Output[0], image.ID = ai.ModalityImage, ai.ModalityImage, "mutated-image"
	image.Capabilities.Sizes[0], image.Capabilities.ImageSizes[0] = "mutated", "mutated"
	image.Capabilities.AspectRatios[0], image.Capabilities.InputFidelity[0] = "mutated", "mutated"
	image.Pricing.InputText = ai.Value(99.0)
	classifier := &c.ClassifierModels[0]
	classifier.ID, classifier.Capabilities.Kinds[0] = "mutated-classifier", ai.ClassifierQuestionBool
	classifier.Cost.Tiers[0].Input = 99
}

func TestOperationCatalogHash(t *testing.T) {
	for name, change := range map[string]func(*ai.Catalog){
		"chat-price":            func(c *ai.Catalog) { c.Models[0].Cost.Input = 5 },
		"image-price":           func(c *ai.Catalog) { c.ImageModels[0].Pricing.InputText = ai.Value(5.0) },
		"image-capability":      func(c *ai.Catalog) { c.ImageModels[0].Capabilities.MaxOutputImages = 2 },
		"image-identity":        func(c *ai.Catalog) { c.ImageModels[0].ID = "another-image" },
		"classifier-price":      func(c *ai.Catalog) { c.ClassifierModels[0].Cost.Tiers[0].Input = 5 },
		"classifier-capability": func(c *ai.Catalog) { c.ClassifierModels[0].Capabilities.MaxChoices = 10 },
		"classifier-identity":   func(c *ai.Catalog) { c.ClassifierModels[0].ID = "another-classifier" },
	} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "E03-operation-catalog-hash-"+name)
			catalog := operationCatalog()
			before := catalogHash(t, catalog)
			change(&catalog)
			// Compare content separately from Version: a host reusing a
			// version must still get a different hash for changed metadata.
			catalog.Version = "host-operations-04"
			after := catalogHash(t, catalog)
			ev.Record("hashes", []string{before, after})
			ev.Check("the hash includes all operation metadata", before != after, "unchanged hash")
		})
	}
}

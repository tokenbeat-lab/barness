package e2e

import (
	"encoding/json"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

// TestCatalogSnapshot observes every mutable part of a host's catalog through
// the public Client: image handling, model sampling defaults, tiered cost and
// the version/hash carried by the result must keep their construction values.
// Regressions include sharing Input, the SamplingParams map or its raw bytes,
// Cost.Tiers, or Models with the caller. The keyed literals also keep the host's
// existing Model/Catalog construction surface exercised (issue 01).
func TestCatalogSnapshot(t *testing.T) {
	ev := run.Case(t, "P01-E11-catalog-deep-copy-after-construction")
	f, raw := loadTextFixture(t, "text-basic.json")
	ev.Fixture("text-basic.json", raw)
	catalog := ai.Catalog{
		Version: "host-prefactor-01",
		Models: []ai.Model{{
			Provider: ai.ProviderOpenAI, API: ai.APIOpenAIResponses, ID: "gpt-4.1-mini", Name: "Host model",
			Input: []ai.Modality{ai.ModalityText, ai.ModalityImage}, ContextWindow: 1047576, MaxTokens: 32768,
			SamplingParams: map[string]json.RawMessage{"top_p": json.RawMessage(`0.5`)},
			Cost: ai.ModelCost{
				CostRates: ai.CostRates{Input: 30, Output: 60},
				// The response's 20 input-side tokens select this tier. Its
				// rates match the frozen text-basic fixture's expected cost.
				Tiers: []ai.CostTier{{InputTokensAbove: 1, CostRates: ai.CostRates{Input: 0.4, Output: 1.6, CacheRead: 0.1}}},
			},
		}},
	}
	version, hash := catalog.Version, catalogHash(t, catalog)
	ev.Record("catalog-at-construction", catalog)
	w := newWorldWith(t, func(c *ai.Config) { c.Catalog = &catalog }, tenantA)

	m := &catalog.Models[0]
	m.Input[1] = ai.ModalityText
	m.SamplingParams["top_p"][2] = '9'
	m.SamplingParams["temperature"] = json.RawMessage(`0.7`)
	m.Cost.Tiers[0].Output = 100
	m.ID = "mutated"
	catalog.Version, catalog.Models = "mutated", nil

	enqueue(ev, w, sseReply(t, f, provider.FramingLF))
	req := textRequest(f)
	req.Messages = []ai.Message{ai.UserMessage{Content: []ai.UserContent{
		ai.Text{Text: "Say hello."}, ai.Image{Data: "AA==", MimeType: "image/png"},
	}}}
	res, err := w.client.CompleteSimple(ctxFor(t), textScope("req-catalog-snapshot"), textTarget(f), req, ai.SimpleOptions{})
	ev.Record("result", res)
	ev.Check("call resolves the original host model", err == nil, "err=%v", err)
	checkFinalMessage(ev, f, res.Message)
	checkPricingSnapshot(ev, res.Metadata, version, hash)
	// Independent expected request, including the original image input and
	// sampling default. A shared slice, map or raw bytes changes this body.
	f.Expect.Request.Body = json.RawMessage(`{
		"model":"gpt-4.1-mini", "stream":true, "store":false,
		"max_output_tokens":32768, "top_p":0.5,
		"input":[
			{"role":"system","content":"You are terse."},
			{"role":"user","content":[
				{"type":"input_text","text":"Say hello."},
				{"type":"input_image","detail":"auto","image_url":"data:image/png;base64,AA=="}
			]}
		]
	}`)
	checkTextRequest(ev, w, f, tenantA, true)
}

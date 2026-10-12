package e2e

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
	"github.com/tokenbeat-lab/barness/ai/examples/localassembly"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/provider"
)

const loadClassifierModel = "gpt-6-luna"

var loadClassifierRoute = mixedRoute{"luna", ai.OperationClassifier, ai.ProviderOpenAI, ai.APIOpenAIDecisions}

// The original mixed batch remains TestMixedPressure. Each additional batch
// fills the same 4/8 permits; sequential batches reuse the original budget.
func TestClassifierImagesPressure(t *testing.T) {
	for _, p := range []struct {
		name    string
		policy  *ai.ResourcePolicy
		tenants int
		budget  uint64
	}{
		{"local", localassembly.MixedPolicy(), 1, 256 << 20}, {"cloud", hostintegration.CloudMixedPolicy(), 2, 512 << 20},
	} {
		t.Run(p.name, func(t *testing.T) {
			ev := pressureCase(t, "E08-mixed-pressure-"+p.name+"-classifier-images-design-load")
			if !ev.Check("explicit finite classifier images", p.policy.Classifier.Images != nil, "images nil") {
				return
			}
			l := mixedDesign(p.policy.MaxConcurrentProcess, p.tenants)
			transport := &mixedBatchTransport{}
			w := newMixedWorld(t, func(c *ai.Config) {
				c.Policy = p.policy
				transport.next = c.Transport
				c.Transport = transport
				c.Catalog.Version = "classifier-load-08-synthetic"
				m := c.Catalog.ClassifierModels[0]
				m.ID, m.Name, m.Provider, m.API = loadClassifierModel, "Synthetic Luna design load", ai.ProviderOpenAI, ai.APIOpenAIDecisions
				m.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
				m.Capabilities.MinChoices = 2
				m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionString}
				m.Capabilities.MaxInputImages = 128
				m.Capabilities.ImageMIMETypes = []string{"image/png", "image/jpeg", "image/webp"}
				m.Cost = ai.ModelCost{CostRates: ai.CostRates{Input: .1}}
				c.Catalog.ClassifierModels = append(c.Catalog.ClassifierModels, m)
			})
			for _, k := range []tenantKey{tenantA, tenantB}[:p.tenants] {
				w.install(k)
				b := loadClassifierRoute.binding(k, w.provider.URL())
				b.AllowClassifierImages = true
				b.AllowedModels = []string{loadClassifierModel}
				w.host.PutBinding(b)
				ev.Record("binding-"+k.tenant, w.host.Binding(k.tenant, "luna"))
			}
			req := mixedLoadClassifier(l)
			img := mixedPNG(t, l.InputImageBytes)
			req.Images = []ai.Image{img, img}
			answers := make([]any, 0, l.Questions)
			for i := range l.Questions {
				answers = append(answers, map[string]any{"name": fmt.Sprintf("q%d", i), "type": "predicate", "probability": .8})
			}
			reply := provider.Reply{Status: 200, Header: map[string]string{"Content-Type": "application/json", "x-request-id": "up-design-load"}, Chunks: [][]byte{mustMarshal(t, map[string]any{"answers": answers, "usage": map[string]any{"input_tokens": 296, "output_tokens": 20, "total_tokens": 316, "input_tokens_details": map[string]int{"cached_tokens": 0, "cache_write_tokens": 0}, "output_tokens_details": map[string]int{"reasoning_tokens": 0}}})}}
			ev.Record("response-script", map[string]any{"status": reply.Status, "header": reply.Header, "chunks": reply.Chunks})
			ev.Record("policy", p.policy)
			ev.Record("catalog", w.client.Catalog())
			ev.Record("input-generation", map[string]any{"load": l, "image": "mixedPNG: 1x1 PNG plus zero padding", "imageSHA256": fmt.Sprintf("%x", sha256.Sum256([]byte(img.Data))), "state": "JSON string of s", "instructions": "JSON string of q"})
			batches := make([]map[string]any, 0, 3)
			var requestBytes int64
			for _, kind := range []string{"design", "cancel", "recovery"} {
				batch, largest := runClassifierImageBatch(t, ev, w, p.name, kind, req, reply, l, p.budget, transport)
				batches = append(batches, batch)
				requestBytes = max(requestBytes, largest)
			}
			rooms := map[string]headroom{
				"Classifier.Images.MaxImages":          room(2, int64(p.policy.Classifier.Images.MaxImages)),
				"MaxImageBytes":                        room(int64(l.InputImageBytes), p.policy.MaxImageBytes),
				"Classifier.Images.MaxTotalImageBytes": room(int64(2*l.InputImageBytes), p.policy.Classifier.Images.MaxTotalImageBytes),
				"MaxRequestBytes":                      room(requestBytes, p.policy.MaxRequestBytes),
				"Classifier.MaxStateBytes":             room(int64(l.StateBytes), p.policy.Classifier.MaxStateBytes),
				"Classifier.MaxQuestions":              room(int64(l.Questions), int64(p.policy.Classifier.MaxQuestions)),
				"Classifier.MaxQuestionBytes":          room(int64(len(mustMarshal(t, req.Questions["q0"]))), p.policy.Classifier.MaxQuestionBytes),
				"MaxOutputBytes":                       room(int64(len(reply.Chunks[0])), p.policy.MaxOutputBytes),
			}
			for name, h := range rooms {
				ev.Check("limit use <=75%: "+name, h.Ratio > 0 && h.Ratio <= .75, "%+v", h)
			}
			cat := w.client.Catalog()
			hash, err := cat.Hash()
			if err != nil {
				t.Fatal(err)
			}
			ev.Record("classifier-images-pressure", map[string]any{
				"schema": 1, "identity": map[string]any{"operation": ai.OperationClassifier, "provider": ai.ProviderOpenAI, "api": ai.APIOpenAIDecisions, "model": loadClassifierModel, "catalogVersion": cat.Version, "catalogHash": hash},
				"load":   map[string]any{"calls": l.Calls, "tenants": l.Tenants, "callsPerTenant": 4, "stateBytes": l.StateBytes, "questions": l.Questions, "instructionBytes": l.InstructionBytes, "imageCount": 2, "imageBytes": []int{l.InputImageBytes, l.InputImageBytes}, "totalImageBytes": 2 * l.InputImageBytes},
				"policy": p.policy, "headroom": rooms, "heapBudget": p.budget, "batches": batches,
				"imageBase64Bytes":  []int{base64.StdEncoding.EncodedLen(l.InputImageBytes), base64.StdEncoding.EncodedLen(l.InputImageBytes)},
				"imageDataURLBytes": []int{len(img.Data) + len("data:image/png;base64,"), len(img.Data) + len("data:image/png;base64,")},
				"requestJSONBytes":  requestBytes, "providerSharedResponseScriptBytes": len(reply.Chunks[0]),
			})
			w.released(ev)
		})
	}
}

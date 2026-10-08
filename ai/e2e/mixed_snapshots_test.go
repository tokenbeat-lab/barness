package e2e

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/examples/hostintegration"
)

func TestMixedClientSnapshot(t *testing.T) {
	ev := run.Case(t, "E07-mixed-snapshot-policy-catalog")
	cat := mixedCatalog()
	p := hostintegration.CloudMixedPolicy()
	hash, err := cat.Hash()
	if err != nil {
		t.Fatal(err)
	}
	w := newMixedWorld(t, func(c *ai.Config) { c.Catalog = &cat; c.Policy = p }, tenantA)
	p.Image.MaxInputImages = 1
	p.Image.MaxOutputImages = 1
	p.Image = nil
	p.Classifier.MaxQuestions = 1
	p.Classifier.MaxStateBytes = 1
	p.Classifier = nil
	p.MaxRequestBytes = 1
	cat.Models[0].ID = "mutated"
	cat.ImageModels[0].Capabilities.Sizes[0] = "mutated"
	cat.ClassifierModels[0].Capabilities.Kinds[0] = "unknown"
	actual := w.client.Catalog()
	got, _ := actual.Hash()
	ev.Check("constructor owns nested catalog", got == hash, "got %s want %s", got, hash)
	for _, r := range mixedRoutes {
		w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
		o := r.invoke(ctxFor(t), w.client, textScope("req-frozen-policy-"+r.id), ai.Target{BindingID: r.id, ModelID: mixedModel})
		w.record(ev, o)
		ev.Check("original policies cannot disable or alter calls", o.err == nil && o.metadata.CatalogHash == hash, "got %v", o.err)
	}
	ev.Record("catalog-hash", hash)
	w.released(ev)
}

func TestMixedMutableInputSnapshots(t *testing.T) {
	for _, r := range mixedRoutes[1:] {
		t.Run(r.id, func(t *testing.T) {
			ev := run.Case(t, "E07-mixed-snapshot-input-options-"+r.id)
			w := newMixedWorld(t, nil, tenantA)
			ref := mixedPNG(t, 128)
			imageReq := ai.ImagesRequest{Prompt: "original prompt", ReferenceImages: []ai.Image{ref}}
			mask := ref
			opts := &ai.OpenAIImagesOptions{Mask: &mask, N: ai.Value(1)}
			classifierReq := mixedClassifierInput()
			original := mustMarshal(t, classifierReq)
			hooks := ai.Hooks{TransformHeaders: func(context.Context, ai.CallScope, http.Header) error {
				imageReq.ReferenceImages[0].Data = "mutated"
				imageReq.Prompt = "mutated"
				opts.Mask.Data = "mutated"
				opts.N = ai.Value(4)
				classifierReq.State[2] = 'X'
				q := classifierReq.Questions["intent"].(ai.ChoiceQuestion)
				q.Instructions[1] = 'X'
				q.Criteria["track"] = json.RawMessage(`"mutated"`)
				delete(classifierReq.Questions, "intent")
				return nil
			}}
			w.provider.EnqueueAt(r.prefix(tenantA), r.reply(t))
			o := mixedOutcome{}
			target := ai.Target{BindingID: r.id, ModelID: mixedModel}
			if r.op == ai.OperationImage {
				var options ai.ImageOptions
				if r.provider == ai.ProviderOpenAI {
					options = opts
				}
				res, err := w.client.WithHooks(hooks).GenerateImages(ctxFor(t), textScope("req-input-copy-"+r.id), target, imageReq, options)
				o = mixedOutcome{metadata: res.Metadata, images: res, usage: res.Usage, err: err}
			} else {
				res, err := w.client.WithHooks(hooks).Classify(ctxFor(t), textScope("req-input-copy-"+r.id), target, classifierReq, nil)
				o = mixedOutcome{metadata: res.Metadata, classifier: res, usage: res.Usage, err: err}
			}
			ev.Check("in-flight input isolated from host mutation", o.err == nil, "got %v", o.err)
			w.record(ev, o)
			reqs := w.provider.Requests()
			ev.Record("requests", reqs)
			if len(reqs) == 1 {
				var body map[string]json.RawMessage
				mustUnmarshal(t, reqs[0].Body, &body)
				if r.op == ai.OperationClassifier {
					var expected map[string]json.RawMessage
					mustUnmarshal(t, original, &expected)
					ev.Check("deep raw/map snapshot", jsonEqual(body["state"], expected["state"]) && jsonEqual(body["questions"], expected["questions"]), "state/questions changed")
				} else if r.provider == ai.ProviderOpenAI {
					ev.Check("mask, refs and pointed options frozen", jsonEqual(body["n"], json.RawMessage(`1`)) && reflect.DeepEqual(body["prompt"], json.RawMessage(`"original prompt"`)), "got %s", reqs[0].Body)
				} else {
					var input []map[string]any
					mustUnmarshal(t, body["input"], &input)
					ev.Check("Google ordered refs frozen", len(input) == 2 && input[0]["text"] == "original prompt" && input[1]["data"] == ref.Data, "got %+v", input)
				}
			}
			w.released(ev)
		})
	}
}

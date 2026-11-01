package e2e

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/evidence"
)

func catalogClassifierImages(t *testing.T, ev *evidence.Case, w *world, i catalogContractIdentity, m ai.ClassifierModel, call func(ai.ClassifierRequest, int64)) {
	if !slices.Contains(m.Input, ai.ModalityImage) {
		ev.Check("text-only classifier has no image declaration", m.Capabilities.MaxInputImages == 0, "%+v", m.Capabilities)
		return
	}
	assets := classifierAssets(t)
	if i.Provider == ai.ProviderTypeSafe {
		assets = jevContextImages(t)
	}
	locations := m.Capabilities.ImageLocations
	if len(locations) == 0 {
		locations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
	}
	for _, location := range locations {
		for _, mime := range m.Capabilities.ImageMIMETypes {
			var image ai.Image
			for _, asset := range assets {
				if asset.MimeType == mime {
					image = asset
				}
			}
			ev.Check("declared MIME has controlled fixture", image.Data != "", "%s", mime)
			if image.Data == "" {
				continue
			}
			r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog image state"`), Images: []ai.Image{image}, Questions: map[string]ai.ClassifierQuestion{"q": ai.ChoiceQuestion{Instructions: json.RawMessage(`"Choose"`), Criteria: map[string]json.RawMessage{"a": json.RawMessage(`"A"`), "b": json.RawMessage(`"B"`)}}}}
			switch location {
			case ai.ClassifierImageLocationContext:
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementContext}}
			case ai.ClassifierImageLocationState:
				r.State = json.RawMessage(`null`)
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/state"}}
			case ai.ClassifierImageLocationInstructions:
				q := r.Questions["q"].(ai.ChoiceQuestion)
				q.Instructions = json.RawMessage(`null`)
				r.Questions["q"] = q
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/instructions"}}
			case ai.ClassifierImageLocationCriteria:
				q := r.Questions["q"].(ai.ChoiceQuestion)
				q.Criteria["a"] = json.RawMessage(`null`)
				r.Questions["q"] = q
				r.ImageReferences = []ai.ImageReference{{ImageIndex: 0, Placement: ai.ImagePlacementValue, Pointer: "/questions/q/criteria/a"}}
			}
			call(r, 8)
		}
	}
	r := ai.ClassifierRequest{State: json.RawMessage(`"Catalog image state"`), Questions: map[string]ai.ClassifierQuestion{"q": ai.BoolQuestion{Instructions: json.RawMessage(`"Is visible?"`)}}, Images: make([]ai.Image, m.Capabilities.MaxInputImages)}
	for n := range r.Images {
		r.Images[n] = assets[0]
	}
	call(r, 8)
	r.Images = append(r.Images, assets[0])
	before := len(w.provider.Requests())
	var opts ai.Options
	if i.Provider == ai.ProviderTypeSafe {
		opts = ai.TypeSafeOptions{ExperimentalImages: true}
	}
	_, err := w.client.Classify(ctxFor(t), textScope("catalog-classifier-image-over"), catalogContractTarget(i), r, opts)
	ev.Check("model image bound refuses before send", err != nil && len(w.provider.Requests()) == before, "%v", err)
}

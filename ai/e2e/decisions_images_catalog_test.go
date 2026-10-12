package e2e

import (
	"errors"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsImagesCatalog(t *testing.T) {
	u := newDecisionsWorld(t, nil)
	for _, name := range []string{"valid", "negative", "zero", "mime-empty", "mime-unknown", "mime-duplicate", "text-with-count", "text-with-mime", "omitted-with-count", "129", "jev-image"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD05-images-catalog-"+name)
			cat := u.client.Catalog()
			m := &cat.ClassifierModels[len(cat.ClassifierModels)-1]
			m.Input = []ai.Modality{ai.ModalityText, ai.ModalityImage}
			m.Capabilities.MaxInputImages = 128
			m.Capabilities.ImageMIMETypes = []string{"image/png"}
			switch name {
			case "negative":
				m.Capabilities.MaxInputImages = -1
			case "zero":
				m.Capabilities.MaxInputImages = 0
			case "mime-empty":
				m.Capabilities.ImageMIMETypes = nil
			case "mime-unknown":
				m.Capabilities.ImageMIMETypes = []string{"image/gif"}
			case "mime-duplicate":
				m.Capabilities.ImageMIMETypes = []string{"image/png", "image/png"}
			case "text-with-count":
				m.Input = []ai.Modality{ai.ModalityText}
				m.Capabilities.ImageMIMETypes = nil
			case "text-with-mime":
				m.Input = []ai.Modality{ai.ModalityText}
				m.Capabilities.MaxInputImages = 0
			case "omitted-with-count":
				m.Input = nil
			case "129":
				m.Capabilities.MaxInputImages = 129
			case "jev-image":
				m.Provider = ai.ProviderTypeSafe
				m.API = ai.APITypeSafeSystemOne
			}
			hash, err := cat.Hash()
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
			_, construct := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: u.host, Credentials: u.host})
			ev.Record("catalog", cat)
			ev.Record("hash", hash)
			ev.Check("shared catalog rejection", (err == nil) == (name == "valid") && (construct == nil) == (name == "valid"), "hash=%v construct=%v", err, construct)
		})
	}
}

func TestDecisionsImagesPolicy(t *testing.T) {
	u := newDecisionsWorld(t, nil)
	for _, name := range []string{"nil", "valid", "count-zero", "count-negative", "total-zero", "total-negative", "total-unbounded"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD07-images-policy-"+name)
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 256, MaxTotalImageBytes: 1024}}
			switch name {
			case "nil":
				p.Classifier.Images = nil
			case "count-zero":
				p.Classifier.Images.MaxImages = 0
			case "count-negative":
				p.Classifier.Images.MaxImages = -1
			case "total-zero":
				p.Classifier.Images.MaxTotalImageBytes = 0
			case "total-negative":
				p.Classifier.Images.MaxTotalImageBytes = -1
			case "total-unbounded":
				p.Classifier.Images.MaxTotalImageBytes = p.MaxRequestBytes + 1
			}
			cat := u.client.Catalog()
			_, err := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: u.host, Credentials: u.host})
			ev.Record("policy", p)
			ev.Record("error", errString(err))
			ev.Check("positive bounded subpolicy", (err == nil) == (name == "nil" || name == "valid"), "got %v", err)
		})
	}
}

func TestDecisionsImagesCopies(t *testing.T) {
	ev := run.Case(t, "P10-OD06-images-copies")
	var source *ai.Catalog
	var policy *ai.ResourcePolicy
	u := newClassifierImagesWorld(t, func(c *ai.Config) { source = c.Catalog; policy = c.Policy })
	want := u.client.Catalog()
	index := len(want.ClassifierModels) - 1
	hash := catalogHash(t, want)
	source.ClassifierModels[index].Capabilities.ImageMIMETypes[0] = "image/gif"
	policy.Classifier.Images.MaxImages = 1
	got := u.client.Catalog()
	lookup, _ := got.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "host-decisions")
	list := got.ClassifierModelsOf(ai.ProviderOpenAI, ai.APIOpenAIDecisions)
	lookup.Capabilities.ImageMIMETypes[0] = "image/gif"
	list[0].Capabilities.ImageMIMETypes[0] = "image/gif"
	ev.Check("Lookup and List own their MIME slices", reflect.DeepEqual(got, want), "catalog query mutated its source")
	got.ClassifierModels[index].Capabilities.ImageMIMETypes[0] = "image/gif"
	ev.Check("all copies independent", reflect.DeepEqual(want, u.client.Catalog()), "catalog changed")
	for _, field := range []string{"count", "mime", "input"} {
		cat := u.client.Catalog()
		m := &cat.ClassifierModels[index]
		switch field {
		case "count":
			m.Capabilities.MaxInputImages = 127
		case "mime":
			m.Capabilities.ImageMIMETypes = []string{"image/png"}
		case "input":
			m.Input = []ai.Modality{ai.ModalityText}
			m.Capabilities.MaxInputImages = 0
			m.Capabilities.ImageMIMETypes = nil
		}
		changed, err := cat.Hash()
		ev.Check("hash includes "+field, err == nil && changed != hash, "got %v", err)
	}
	enqueueClassifierImages(t, ev, u)
	req := decisionsInput()
	req.Images = classifierAssets(t)
	res, err := u.client.Classify(ctxFor(t), textScope("req-image-copies"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Check("snapshot still authorizes original assets", err == nil && res.Metadata.CatalogHash == hash, "got %v", err)
	u.records(ev, res, err)
	u.released(ev)
}

func TestDecisionsImagesWrongBinding(t *testing.T) {
	for _, op := range []ai.Operation{"", ai.OperationChat, ai.OperationImage} {
		t.Run(string(op), func(t *testing.T) {
			ev := run.Case(t, "P10-OD08-images-binding-"+string(op))
			u := newDecisionsWorld(t, nil)
			u.updateBindingOf(tenantA, "classifier", func(b *ai.Binding) { b.Operation = op; b.AllowClassifierImages = true })
			res, err := u.client.Classify(ctxFor(t), textScope("req-image-binding-"+string(op)), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
			ev.Check("invalid flag fails binding phase", errors.Is(err, &ai.Error{Code: ai.CodeInvalidRequest, Phase: ai.PhaseBinding}) && u.host.ReadsFor(res.Metadata.RequestID).Credentials == 0, "got %v", err)
			u.records(ev, res, err)
			u.released(ev)
		})
	}
}

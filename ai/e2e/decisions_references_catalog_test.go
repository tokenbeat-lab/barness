package e2e

import (
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
)

func TestDecisionsReferencesCatalog(t *testing.T) {
	u := newClassifierImagesWorld(t, nil)
	for _, name := range []string{"omitted", "context", "unknown", "duplicate", "state", "instructions", "criteria", "text", "jev"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "UC02-catalog-"+name)
			cat := u.client.Catalog()
			m := &cat.ClassifierModels[len(cat.ClassifierModels)-1]
			legacy := catalogHash(t, cat)
			if name != "omitted" {
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
			}
			switch name {
			case "unknown":
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{"other"}
			case "duplicate":
				m.Capabilities.ImageLocations = append(m.Capabilities.ImageLocations, ai.ClassifierImageLocationContext)
			case "state":
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationState}
			case "instructions":
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationInstructions}
			case "criteria":
				m.Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationCriteria}
			case "text":
				m.Input = []ai.Modality{ai.ModalityText}
				m.Capabilities.MaxInputImages = 0
				m.Capabilities.ImageMIMETypes = nil
			case "jev":
				m.Provider = ai.ProviderTypeSafe
				m.API = ai.APITypeSafeSystemOne
			}
			hash, err := cat.Hash()
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
			_, construct := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: u.host, Credentials: u.host})
			valid := name == "omitted" || name == "context"
			ev.Check("locations match implemented protocol at both catalog boundaries", (err == nil) == valid && (construct == nil) == valid, "hash=%v construct=%v", err, construct)
			if valid {
				ev.Check("omitted default leaves hash unchanged; explicit context affects hash", (hash == legacy) == (name == "omitted"), "unexpected hash")
			}
			ev.Record("catalog", cat)
			ev.Record("hash", hash)
		})
	}
}

func TestDecisionsReferencesSnapshots(t *testing.T) {
	ev := run.Case(t, "UC02-snapshots")
	var source *ai.Catalog
	var policy *ai.ResourcePolicy
	u := newClassifierImagesWorld(t, func(c *ai.Config) {
		source, policy = c.Catalog, c.Policy
		c.Catalog.ClassifierModels[len(c.Catalog.ClassifierModels)-1].Capabilities.ImageLocations = []ai.ClassifierImageLocation{ai.ClassifierImageLocationContext}
		setReferencePolicy(t, c, 3, 1<<19)
	})
	want := u.client.Catalog()
	hash := catalogHash(t, want)
	index := len(want.ClassifierModels) - 1
	source.ClassifierModels[index].Capabilities.ImageLocations[0] = ai.ClassifierImageLocationState
	policy.Classifier.Images.References.MaxReferences = 1
	got := u.client.Catalog()
	lookup, _ := got.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "host-decisions")
	list := got.ClassifierModelsOf(ai.ProviderOpenAI, ai.APIOpenAIDecisions)
	lookup.Capabilities.ImageLocations[0] = ai.ClassifierImageLocationState
	list[0].Capabilities.ImageLocations[0] = ai.ClassifierImageLocationCriteria
	ev.Check("queries own locations", reflect.DeepEqual(got, want), "query mutated catalog")
	got.ClassifierModels[index].Capabilities.ImageLocations[0] = ai.ClassifierImageLocationInstructions
	ev.Check("config and catalog snapshots own locations", reflect.DeepEqual(u.client.Catalog(), want), "snapshot mutated")
	req := decisionsInput()
	req.Images = classifierAssets(t)[:1]
	req = decisionsReferenceRequest(t, req, `[{"imageIndex":0,"placement":"context"},{"imageIndex":0,"placement":"context"}]`)
	ev.Fixture("public-request.json", mustMarshal(t, req))
	enqueueClassifierImages(t, ev, u)
	res, err := u.client.Classify(ctxFor(t), textScope("req-reference-snapshot"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, req, nil)
	ev.Check("snapshot capability and reference subpolicy survive caller mutation", err == nil && res.Metadata.CatalogHash == hash, "got %v", err)
	u.records(ev, res, err)
	u.released(ev)
}

func TestDecisionsReferencesPolicy(t *testing.T) {
	u := newClassifierImagesWorld(t, nil)
	for _, sc := range []struct {
		name  string
		count int
		bytes int64
		valid bool
	}{
		{"valid", 2, 256, true}, {"expanded-over-body", 2, 1 << 22, true},
		{"zero-count", 0, 256, false}, {"negative-count", -1, 256, false}, {"zero-bytes", 2, 0, false}, {"negative-bytes", 2, -1, false},
	} {
		t.Run(sc.name, func(t *testing.T) {
			ev := run.Case(t, "UC02-policy-"+sc.name)
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096, Images: &ai.ClassifierImagePolicy{MaxImages: 2, MaxTotalImageBytes: 256, References: &ai.ClassifierImageReferencePolicy{MaxReferences: sc.count, MaxExpandedImageBytes: sc.bytes}}}
			cat := u.client.Catalog()
			_, err := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: u.host, Credentials: u.host})
			ev.Check("explicit expansion limits must be positive; encoded body remains independent", (err == nil) == sc.valid, "got %v", err)
			ev.Record("policy", p)
			ev.Record("error", errString(err))
		})
	}
}

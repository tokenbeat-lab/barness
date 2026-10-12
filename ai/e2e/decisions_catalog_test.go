package e2e

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/tokenbeat-lab/barness/ai"
	"github.com/tokenbeat-lab/barness/ai/internal/testkit/host"
)

func TestDecisionsCatalogValidation(t *testing.T) {
	u := newDecisionsWorld(t, nil)
	base := u.client.Catalog()
	for _, name := range []string{"valid", "input-empty", "input-unknown", "input-duplicate", "input-no-text", "min-negative", "min-above-max", "min-omitted", "formats-omitted", "formats-unknown", "formats-duplicate", "formats-object", "legacy-wide"} {
		t.Run(name, func(t *testing.T) {
			ev := run.Case(t, "P10-OD05-catalog-"+name)
			cat := u.client.Catalog()
			m := &cat.ClassifierModels[len(cat.ClassifierModels)-1]
			valid := false
			switch name {
			case "valid":
				valid = true
			case "input-empty":
				m.Input = nil
				valid = true
			case "input-unknown":
				m.Input = []ai.Modality{"audio"}
			case "input-duplicate":
				m.Input = []ai.Modality{ai.ModalityText, ai.ModalityText}
			case "input-no-text":
				m.Input = []ai.Modality{ai.ModalityImage}
			case "min-negative":
				m.Capabilities.MinChoices = -1
			case "min-above-max":
				m.Capabilities.MinChoices = 256
			case "min-omitted":
				m.Capabilities.MinChoices = 0
			case "formats-omitted":
				m.Capabilities.DescriptionFormats = nil
			case "formats-unknown":
				m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{"number"}
			case "formats-duplicate":
				m.Capabilities.DescriptionFormats = append(m.Capabilities.DescriptionFormats, ai.ClassifierDescriptionString)
			case "formats-object":
				m.Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionObject}
			case "legacy-wide":
				cat.ClassifierModels = cat.ClassifierModels[:1]
				m = &cat.ClassifierModels[0]
				m.Capabilities.MaxChoices = 300
				m.Capabilities.MaxScoreLevels = 20
				valid = true
			}
			p := validPolicy()
			p.Classifier = &ai.ClassifierPolicy{MaxQuestions: 8, MaxStateBytes: 4096, MaxQuestionBytes: 4096}
			h := host.New()
			_, err := ai.NewClient(ai.Config{Policy: p, Catalog: &cat, Bindings: h, Credentials: h})
			hash, hashErr := cat.Hash()
			ev.Record("catalog", cat)
			ev.Record("hash", hash)
			ev.Record("construction-error", errString(err))
			ev.Record("hash-error", errString(hashErr))
			ev.Check("Hash and NewClient share validation", (err == nil) == valid && (hashErr == nil) == valid, "construct=%v hash=%v", err, hashErr)
		})
	}
	legacy := base.ClassifierModels[0]
	raw := mustMarshal(t, legacy)
	var fields map[string]json.RawMessage
	mustUnmarshal(t, raw, &fields)
	ev := run.Case(t, "P10-OD01-legacy-omission")
	ev.Record("model", legacy)
	_, hasInput := fields["input"]
	var caps map[string]json.RawMessage
	mustUnmarshal(t, fields["capabilities"], &caps)
	_, hasMin := caps["minChoices"]
	_, hasFormats := caps["descriptionFormats"]
	ev.Check("legacy JSON remains omitted", !hasInput && !hasMin && !hasFormats, "got %s", raw)
}

func TestDecisionsCatalogCopies(t *testing.T) {
	ev := run.Case(t, "P10-OD06-catalog-independent-copies")
	var source *ai.Catalog
	u := newDecisionsWorld(t, func(c *ai.Config) {
		source = c.Catalog
		m := &source.ClassifierModels[len(source.ClassifierModels)-1]
		m.Cost.Tiers = []ai.CostTier{{InputTokensAbove: 1000, CostRates: ai.CostRates{Input: 2}}}
	})
	want := u.client.Catalog()
	hash := catalogHash(t, want)
	index := len(want.ClassifierModels) - 1
	mutate := func(m *ai.ClassifierModel) {
		m.Input[0] = ai.ModalityImage
		m.Capabilities.DescriptionFormats[0] = ai.ClassifierDescriptionObject
		m.Capabilities.Kinds[0] = ai.ClassifierQuestionScore
		m.Cost.Tiers[0].Input = 99
	}
	mutate(&source.ClassifierModels[index])
	got := u.client.Catalog()
	ev.Check("construction snapshot owns slices", reflect.DeepEqual(got, want), "catalog changed")
	lookup, ok := got.LookupClassifier(ai.ProviderOpenAI, ai.APIOpenAIDecisions, "host-decisions")
	ev.Check("lookup exact identity", ok, "not found")
	list := got.ClassifierModelsOf(ai.ProviderOpenAI, ai.APIOpenAIDecisions)
	ev.Check("filtered list", len(list) == 1, "got %d", len(list))
	mutate(&lookup)
	mutate(&list[0])
	mutate(&got.ClassifierModels[index])
	ev.Check("three return boundaries independent", reflect.DeepEqual(u.client.Catalog(), want), "client changed")
	for _, field := range []string{"input", "formats", "min", "tiers"} {
		cat := u.client.Catalog()
		m := &cat.ClassifierModels[index]
		switch field {
		case "input":
			m.Input = nil
		case "formats":
			cat.ClassifierModels[0].Capabilities.DescriptionFormats = []ai.ClassifierDescriptionFormat{ai.ClassifierDescriptionObject}
		case "min":
			m.Capabilities.MinChoices = 3
		case "tiers":
			m.Cost.Tiers[0].Input = 3
		}
		changed, err := cat.Hash()
		ev.Check("Hash includes "+field, err == nil && changed != hash, "got %v", err)
	}
	enqueue(ev, u.world, fixtureReply{Status: 200, Body: `{"answers":[` + decisionsAnswer + `],` + decisionsUsage + `}`}.script(t, decisionsProtocol, ""))
	res, err := u.client.Classify(ctxFor(t), textScope("req-catalog-copies"), ai.Target{BindingID: "classifier", ModelID: "host-decisions"}, decisionsInput(), nil)
	ev.Check("mutation never changes actual call snapshot", err == nil && res.Metadata.CatalogHash == hash && res.Usage.Cost.Total == .000296, "got %v", err)
	u.records(ev, res, err)
	u.released(ev)
}
